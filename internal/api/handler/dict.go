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

// DictHandler 字典管理处理器
type DictHandler struct {
	db *gorm.DB
}

func NewDictHandler(db *gorm.DB) *DictHandler {
	return &DictHandler{db: db}
}

func (h *DictHandler) RegisterRoutes(r *gin.RouterGroup) {
	dict := r.Group("/dicts")
	{
		dict.GET("/types", h.ListDictTypes)
		dict.POST("/types", h.CreateDictType)
		dict.PUT("/types/:id", h.UpdateDictType)
		dict.DELETE("/types/:id", h.DeleteDictType)
		dict.GET("/items", h.ListDictItems)
		dict.POST("/items", h.CreateDictItem)
		dict.PUT("/items/:id", h.UpdateDictItem)
		dict.DELETE("/items/:id", h.DeleteDictItem)
	}
}

// --- DictType ---

type DictTypeListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

type CreateDictTypeReq struct {
	Name   string `json:"name" binding:"required,max=64"`
	Code   string `json:"code" binding:"required,max=64"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateDictTypeReq struct {
	Name   string `json:"name" binding:"max=64"`
	Code   string `json:"code" binding:"max=64"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

func (h *DictHandler) ListDictTypes(c *gin.Context) {
	var req DictTypeListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.DictType{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.DictType
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *DictHandler) CreateDictType(c *gin.Context) {
	var req CreateDictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 检查 code 是否重复
	var count int64
	h.db.Model(&model.DictType{}).Where("company_id = ? AND code = ?", companyID, req.Code).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "字典编码已存在")
		return
	}

	item := model.DictType{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:               req.Name,
		Code:               req.Code,
		Status:             req.Status,
	}
	if err := h.db.Create(&item).Error; err != nil {
		log.Error().Err(err).Msg("create dict type failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, item)
}

func (h *DictHandler) UpdateDictType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateDictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var item model.DictType
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&item).Error; err != nil {
		response.NotFound(c, "字典类型不存在")
		return
	}

	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Code != "" {
		item.Code = req.Code
	}
	item.Status = req.Status

	if err := h.db.Save(&item).Error; err != nil {
		log.Error().Err(err).Msg("update dict type failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, item)
}

func (h *DictHandler) DeleteDictType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 检查是否有关联的字典项
	var count int64
	h.db.Model(&model.DictItem{}).Where("dict_type_id = ?", id).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "该字典类型下存在字典项，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.DictType{}).Error; err != nil {
		log.Error().Err(err).Msg("delete dict type failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// --- DictItem ---

type DictItemListReq struct {
	DictTypeID uint   `form:"dictTypeId" binding:"required"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
}

type CreateDictItemReq struct {
	DictTypeID uint   `json:"dictTypeId" binding:"required"`
	Label      string `json:"label" binding:"required,max=64"`
	Value      string `json:"value" binding:"required,max=64"`
	Sort       int    `json:"sort"`
	Status     int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateDictItemReq struct {
	Label  string `json:"label" binding:"max=64"`
	Value  string `json:"value" binding:"max=64"`
	Sort   int    `json:"sort"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

func (h *DictHandler) ListDictItems(c *gin.Context) {
	var req DictItemListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	query := h.db.Model(&model.DictItem{}).Where("dict_type_id = ?", req.DictTypeID)
	if req.Keyword != "" {
		query = query.Where("label LIKE ? OR value LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.DictItem
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *DictHandler) CreateDictItem(c *gin.Context) {
	var req CreateDictItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 验证 DictType 存在且属于当前公司
	var dictType model.DictType
	if err := h.db.Where("id = ? AND company_id = ?", req.DictTypeID, companyID).First(&dictType).Error; err != nil {
		response.NotFound(c, "字典类型不存在")
		return
	}

	item := model.DictItem{
		DictTypeID: req.DictTypeID,
		Label:      req.Label,
		Value:      req.Value,
		Sort:       req.Sort,
		Status:     req.Status,
	}
	if err := h.db.Create(&item).Error; err != nil {
		log.Error().Err(err).Msg("create dict item failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, item)
}

func (h *DictHandler) UpdateDictItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateDictItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var item model.DictItem
	if err := h.db.First(&item, id).Error; err != nil {
		response.NotFound(c, "字典项不存在")
		return
	}
	// 验证所属 DictType 的 company_id
	var dictType model.DictType
	if err := h.db.Where("id = ? AND company_id = ?", item.DictTypeID, companyID).First(&dictType).Error; err != nil {
		response.Unauthorized(c, "无权操作")
		return
	}

	if req.Label != "" {
		item.Label = req.Label
	}
	if req.Value != "" {
		item.Value = req.Value
	}
	item.Sort = req.Sort
	item.Status = req.Status

	if err := h.db.Save(&item).Error; err != nil {
		log.Error().Err(err).Msg("update dict item failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, item)
}

func (h *DictHandler) DeleteDictItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var item model.DictItem
	if err := h.db.First(&item, id).Error; err != nil {
		response.NotFound(c, "字典项不存在")
		return
	}
	var dictType model.DictType
	if err := h.db.Where("id = ? AND company_id = ?", item.DictTypeID, companyID).First(&dictType).Error; err != nil {
		response.Unauthorized(c, "无权操作")
		return
	}

	if err := h.db.Delete(&item).Error; err != nil {
		log.Error().Err(err).Msg("delete dict item failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
