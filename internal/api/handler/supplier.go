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

// SupplierHandler 供应商处理器
type SupplierHandler struct {
	db *gorm.DB
}

func NewSupplierHandler(db *gorm.DB) *SupplierHandler {
	return &SupplierHandler{db: db}
}

func (h *SupplierHandler) RegisterRoutes(r *gin.RouterGroup) {
	supplier := r.Group("/suppliers")
	{
		supplier.GET("", h.ListSuppliers)
		supplier.POST("", h.CreateSupplier)
		supplier.PUT("/:id", h.UpdateSupplier)
		supplier.DELETE("/:id", h.DeleteSupplier)
	}
}

type SupplierListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	var req SupplierListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Supplier{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.Supplier
	if err := query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list).Error; err != nil {
		log.Error().Err(err).Msg("list suppliers failed")
		response.ServerError(c, "查询失败")
		return
	}
	if list == nil {
		list = []model.Supplier{}
	}
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required,max=64"`
		Code    string `json:"code" binding:"max=64"`
		Contact string `json:"contact" binding:"max=64"`
		Phone   string `json:"phone" binding:"max=32"`
		Status  *int8  `json:"status" binding:"omitempty,oneof=0 1"`
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

	supplier := model.Supplier{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Contact:              req.Contact,
		Phone:                req.Phone,
		Status:               1,
	}
	if req.Status != nil {
		supplier.Status = *req.Status
	}
	if err := h.db.Select("CompanyID", "Name", "Code", "Contact", "Phone", "Status").Create(&supplier).Error; err != nil {
		log.Error().Err(err).Msg("create supplier failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, supplier)
}

func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name    string `json:"name" binding:"max=64"`
		Code    string `json:"code" binding:"max=64"`
		Contact string `json:"contact" binding:"max=64"`
		Phone   string `json:"phone" binding:"max=32"`
		Status  int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var supplier model.Supplier
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&supplier).Error; err != nil {
		response.NotFound(c, "供应商不存在")
		return
	}

	if req.Name != "" {
		supplier.Name = req.Name
	}
	supplier.Code = req.Code
	supplier.Contact = req.Contact
	supplier.Phone = req.Phone
	supplier.Status = req.Status

	if err := h.db.Save(&supplier).Error; err != nil {
		log.Error().Err(err).Msg("update supplier failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, supplier)
}

func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Supplier{}).Error; err != nil {
		log.Error().Err(err).Msg("delete supplier failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
