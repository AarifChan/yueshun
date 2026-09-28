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

// ProductSettingHandler 商品设置处理器（单位/规格/标签/销售范围）
type ProductSettingHandler struct {
	db *gorm.DB
}

func NewProductSettingHandler(db *gorm.DB) *ProductSettingHandler {
	return &ProductSettingHandler{db: db}
}

func (h *ProductSettingHandler) RegisterRoutes(r *gin.RouterGroup) {
	ps := r.Group("/product-settings")
	{
		ps.GET("/units", h.ListUnits)
		ps.POST("/units", h.CreateUnit)
		ps.PUT("/units/:id", h.UpdateUnit)
		ps.DELETE("/units/:id", h.DeleteUnit)

		ps.GET("/specs", h.ListSpecs)
		ps.POST("/specs", h.CreateSpec)
		ps.PUT("/specs/:id", h.UpdateSpec)
		ps.DELETE("/specs/:id", h.DeleteSpec)

		ps.GET("/tags", h.ListTags)
		ps.POST("/tags", h.CreateTag)
		ps.PUT("/tags/:id", h.UpdateTag)
		ps.DELETE("/tags/:id", h.DeleteTag)

		ps.GET("/scopes", h.ListScopes)
		ps.POST("/scopes", h.CreateScope)
		ps.PUT("/scopes/:id", h.UpdateScope)
		ps.DELETE("/scopes/:id", h.DeleteScope)
	}
}

type settingListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

func (h *ProductSettingHandler) companyID(c *gin.Context) uint {
	return middleware.GetCompanyID(c)
}

// ==================== 商品单位 ====================

func (h *ProductSettingHandler) ListUnits(c *gin.Context) {
	var req settingListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	query := h.db.Model(&model.GoodsUnit{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	var total int64
	query.Count(&total)
	var list []model.GoodsUnit
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductSettingHandler) CreateUnit(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required,max=32"`
		Code   string `json:"code" binding:"omitempty,max=32"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	unit := model.GoodsUnit{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Status:               req.Status,
	}
	if err := h.db.Create(&unit).Error; err != nil {
		log.Error().Err(err).Msg("create goods unit failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, unit)
}

func (h *ProductSettingHandler) UpdateUnit(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name   string `json:"name" binding:"required,max=32"`
		Code   string `json:"code" binding:"omitempty,max=32"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	var unit model.GoodsUnit
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&unit).Error; err != nil {
		response.NotFound(c, "单位不存在")
		return
	}
	unit.Name = req.Name
	unit.Code = req.Code
	unit.Status = req.Status
	if err := h.db.Save(&unit).Error; err != nil {
		log.Error().Err(err).Msg("update goods unit failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, unit)
}

func (h *ProductSettingHandler) DeleteUnit(c *gin.Context) {
	h.deleteByID(c, &model.GoodsUnit{}, "单位")
}

// ==================== 商品规格 ====================

func (h *ProductSettingHandler) ListSpecs(c *gin.Context) {
	var req settingListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	query := h.db.Model(&model.ProductSpec{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	var total int64
	query.Count(&total)
	var list []model.ProductSpec
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductSettingHandler) CreateSpec(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Values string `json:"values" binding:"omitempty,max=255"`
		Remark string `json:"remark" binding:"omitempty,max=255"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	spec := model.ProductSpec{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Values:               req.Values,
		Remark:               req.Remark,
		Status:               req.Status,
	}
	if err := h.db.Create(&spec).Error; err != nil {
		log.Error().Err(err).Msg("create product spec failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, spec)
}

func (h *ProductSettingHandler) UpdateSpec(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Values string `json:"values" binding:"omitempty,max=255"`
		Remark string `json:"remark" binding:"omitempty,max=255"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	var spec model.ProductSpec
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&spec).Error; err != nil {
		response.NotFound(c, "规格不存在")
		return
	}
	spec.Name = req.Name
	spec.Values = req.Values
	spec.Remark = req.Remark
	spec.Status = req.Status
	if err := h.db.Save(&spec).Error; err != nil {
		log.Error().Err(err).Msg("update product spec failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, spec)
}

func (h *ProductSettingHandler) DeleteSpec(c *gin.Context) {
	h.deleteByID(c, &model.ProductSpec{}, "规格")
}

// ==================== 商品标签 ====================

func (h *ProductSettingHandler) ListTags(c *gin.Context) {
	var req settingListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	query := h.db.Model(&model.ProductTag{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	var total int64
	query.Count(&total)
	var list []model.ProductTag
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductSettingHandler) CreateTag(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Color  string `json:"color" binding:"omitempty,max=16"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	if req.Color == "" {
		req.Color = "#409EFF"
	}
	tag := model.ProductTag{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Color:                req.Color,
		Status:               req.Status,
	}
	if err := h.db.Create(&tag).Error; err != nil {
		log.Error().Err(err).Msg("create product tag failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, tag)
}

func (h *ProductSettingHandler) UpdateTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Color  string `json:"color" binding:"omitempty,max=16"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	var tag model.ProductTag
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&tag).Error; err != nil {
		response.NotFound(c, "标签不存在")
		return
	}
	tag.Name = req.Name
	if req.Color != "" {
		tag.Color = req.Color
	}
	tag.Status = req.Status
	if err := h.db.Save(&tag).Error; err != nil {
		log.Error().Err(err).Msg("update product tag failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, tag)
}

func (h *ProductSettingHandler) DeleteTag(c *gin.Context) {
	h.deleteByID(c, &model.ProductTag{}, "标签")
}

// ==================== 商品销售范围 ====================

func (h *ProductSettingHandler) ListScopes(c *gin.Context) {
	var req settingListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	query := h.db.Model(&model.ProductSalesScope{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	var total int64
	query.Count(&total)
	var list []model.ProductSalesScope
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductSettingHandler) CreateScope(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Code   string `json:"code" binding:"omitempty,max=32"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	scope := model.ProductSalesScope{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Status:               req.Status,
	}
	if err := h.db.Create(&scope).Error; err != nil {
		log.Error().Err(err).Msg("create sales scope failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, scope)
}

func (h *ProductSettingHandler) UpdateScope(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Code   string `json:"code" binding:"omitempty,max=32"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	var scope model.ProductSalesScope
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&scope).Error; err != nil {
		response.NotFound(c, "销售范围不存在")
		return
	}
	scope.Name = req.Name
	scope.Code = req.Code
	scope.Status = req.Status
	if err := h.db.Save(&scope).Error; err != nil {
		log.Error().Err(err).Msg("update sales scope failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, scope)
}

func (h *ProductSettingHandler) DeleteScope(c *gin.Context) {
	h.deleteByID(c, &model.ProductSalesScope{}, "销售范围")
}

// deleteByID 通用删除
func (h *ProductSettingHandler) deleteByID(c *gin.Context, modelRef interface{}, name string) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := h.companyID(c)
	result := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(modelRef)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("delete setting failed")
		response.ServerError(c, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, name+"不存在")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
