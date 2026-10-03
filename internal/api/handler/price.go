package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// PriceHandler 价格体系处理器
type PriceHandler struct {
	db *gorm.DB
}

func NewPriceHandler(db *gorm.DB) *PriceHandler {
	return &PriceHandler{db: db}
}

func (h *PriceHandler) RegisterRoutes(r *gin.RouterGroup) {
	price := r.Group("/prices")
	{
		// 价格体系
		price.GET("/levels", h.ListPriceLevels)
		price.POST("/levels", h.CreatePriceLevel)
		price.PUT("/levels/:id", h.UpdatePriceLevel)
		price.DELETE("/levels/:id", h.DeletePriceLevel)

		// 商品价格
		price.GET("/products/:productId", h.GetProductPrices)
		price.POST("/products/:productId", h.SetProductPrices)
		price.PUT("/products/:productId", h.UpdateProductPrices)
	}
}

// ==================== 价格体系 ====================

type PriceLevelListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

func (h *PriceHandler) ListPriceLevels(c *gin.Context) {
	var req PriceLevelListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.PriceLevel{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.PriceLevel
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *PriceHandler) CreatePriceLevel(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required,max=64"`
		Code          string `json:"code" binding:"max=64"`
		Description   string `json:"description" binding:"max=255"`
		IsDefault     bool   `json:"isDefault"`
		Required      bool   `json:"required"`
		Rule          string `json:"rule" binding:"max=255"`
		CustomerScope string `json:"customerScope" binding:"max=255"`
		Status        int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	level := model.PriceLevel{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Description:          req.Description,
		IsDefault:            req.IsDefault,
		Required:             req.Required,
		Rule:                 req.Rule,
		CustomerScope:        req.CustomerScope,
		Status:               req.Status,
	}
	if err := h.db.Create(&level).Error; err != nil {
		log.Error().Err(err).Msg("create price level failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, level)
}

func (h *PriceHandler) UpdatePriceLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name          string `json:"name" binding:"max=64"`
		Code          string `json:"code" binding:"max=64"`
		Description   string `json:"description" binding:"max=255"`
		IsDefault     bool   `json:"isDefault"`
		Required      bool   `json:"required"`
		Rule          string `json:"rule" binding:"max=255"`
		CustomerScope string `json:"customerScope" binding:"max=255"`
		Status        int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var level model.PriceLevel
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&level).Error; err != nil {
		response.NotFound(c, "价格体系不存在")
		return
	}

	if req.Name != "" {
		level.Name = req.Name
	}
	if req.Code != "" {
		level.Code = req.Code
	}
	level.Description = req.Description
	level.IsDefault = req.IsDefault
	level.Required = req.Required
	level.Rule = req.Rule
	level.CustomerScope = req.CustomerScope
	level.Status = req.Status

	if err := h.db.Save(&level).Error; err != nil {
		log.Error().Err(err).Msg("update price level failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, level)
}

func (h *PriceHandler) DeletePriceLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var count int64
	h.db.Model(&model.ProductPrice{}).Where("level_id = ?", id).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "该价格体系已设置价格，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.PriceLevel{}).Error; err != nil {
		log.Error().Err(err).Msg("delete price level failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 商品价格 ====================

type ProductPriceItem struct {
	UnitID   uint    `json:"unitId" binding:"required"`
	LevelID  uint    `json:"levelId"`
	Price    float64 `json:"price" binding:"required,gte=0"`
}

func (h *PriceHandler) GetProductPrices(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 验证商品属于当前公司
	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", productID, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	var prices []model.ProductPrice
	h.db.Where("product_id = ? AND customer_id = 0", productID).Find(&prices)
	response.Ok(c, prices)
}

func (h *PriceHandler) SetProductPrices(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	var req struct {
		Prices []ProductPriceItem `json:"prices" binding:"required,dive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 验证商品
	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", productID, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	// 批量设置价格
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧价格（非客户专属）
		if err := tx.Where("product_id = ? AND customer_id = 0", productID).Delete(&model.ProductPrice{}).Error; err != nil {
			return err
		}
		// 创建新价格
		for _, p := range req.Prices {
			pp := model.ProductPrice{
				ProductID:  uint(productID),
				UnitID:     p.UnitID,
				LevelID:    p.LevelID,
				CustomerID: 0,
				Price:      p.Price,
				Status:     1,
			}
			if err := tx.Create(&pp).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("set product prices failed")
		response.ServerError(c, "设置价格失败")
		return
	}
	response.OkWithMessage(c, "价格设置成功", nil)
}

func (h *PriceHandler) UpdateProductPrices(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	var req struct {
		Prices []ProductPriceItem `json:"prices" binding:"required,dive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 验证商品
	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", productID, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	// 更新价格（删除旧 + 创建新）
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ? AND customer_id = 0", productID).Delete(&model.ProductPrice{}).Error; err != nil {
			return err
		}
		for _, p := range req.Prices {
			pp := model.ProductPrice{
				ProductID:  uint(productID),
				UnitID:     p.UnitID,
				LevelID:    p.LevelID,
				CustomerID: 0,
				Price:      p.Price,
				Status:     1,
			}
			if err := tx.Create(&pp).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update product prices failed")
		response.ServerError(c, "更新价格失败")
		return
	}
	response.OkWithMessage(c, "价格更新成功", nil)
}
