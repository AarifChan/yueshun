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
		ps.POST("/units/sort", h.SortUnits)
		ps.PUT("/units/:id", h.UpdateUnit)
		ps.DELETE("/units/:id", h.DeleteUnit)

		ps.GET("/specs", h.ListSpecs)
		ps.POST("/specs", h.CreateSpec)
		ps.POST("/specs/sort", h.SortSpecs)
		ps.PUT("/specs/:id", h.UpdateSpec)
		ps.DELETE("/specs/:id", h.DeleteSpec)

		ps.GET("/tags", h.ListTags)
		ps.POST("/tags", h.CreateTag)
		ps.POST("/tags/sort", h.SortTags)
		ps.PUT("/tags/:id", h.UpdateTag)
		ps.DELETE("/tags/:id", h.DeleteTag)

		ps.GET("/scopes", h.ListScopes)
		ps.POST("/scopes", h.CreateScope)
		ps.PUT("/scopes/:id", h.UpdateScope)
		ps.DELETE("/scopes/:id", h.DeleteScope)
	}
}

type settingListReq struct {
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=20"`
	Keyword     string `form:"keyword"`
	StorageType int8   `form:"storageType"`
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
	if req.StorageType == 1 || req.StorageType == 2 {
		query = query.Where("storage_type = ?", req.StorageType)
	}
	var total int64
	query.Count(&total)
	var list []model.GoodsUnit
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductSettingHandler) CreateUnit(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,max=4"`
		Code        string `json:"code" binding:"omitempty,max=32"`
		Sort        int    `json:"sort"`
		StorageType int8   `json:"storageType" binding:"omitempty,oneof=1 2"`
		Status      int8   `json:"status"`
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
	if req.StorageType == 0 {
		req.StorageType = 1
	}
	unit := model.GoodsUnit{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Sort:                 req.Sort,
		StorageType:          req.StorageType,
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
		Name        string `json:"name" binding:"required,max=4"`
		Code        string `json:"code" binding:"omitempty,max=32"`
		Sort        int    `json:"sort"`
		StorageType int8   `json:"storageType" binding:"omitempty,oneof=1 2"`
		Status      int8   `json:"status"`
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
	if req.StorageType == 0 {
		req.StorageType = 1
	}
	unit.Name = req.Name
	unit.Code = req.Code
	unit.Sort = req.Sort
	unit.StorageType = req.StorageType
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

type SortUnitsReq struct {
	Items []struct {
		ID   uint `json:"id" binding:"required"`
		Sort int  `json:"sort"`
	} `json:"items" binding:"required"`
}

func (h *ProductSettingHandler) SortUnits(c *gin.Context) {
	var req SortUnitsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			if err := tx.Model(&model.GoodsUnit{}).
				Where("id = ? AND company_id = ?", item.ID, companyID).
				Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("sort units failed")
		response.ServerError(c, "排序失败")
		return
	}
	response.OkWithMessage(c, "排序成功", nil)
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
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductSettingHandler) CreateSpec(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required,max=10"`
		Values string `json:"values" binding:"omitempty,max=255"`
		Remark string `json:"remark" binding:"omitempty,max=255"`
		Sort   int    `json:"sort"`
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
		Sort:                 req.Sort,
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
		Name   string `json:"name" binding:"required,max=10"`
		Values string `json:"values" binding:"omitempty,max=255"`
		Remark string `json:"remark" binding:"omitempty,max=255"`
		Sort   int    `json:"sort"`
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
	spec.Sort = req.Sort
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

func (h *ProductSettingHandler) SortSpecs(c *gin.Context) {
	var req SortUnitsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			if err := tx.Model(&model.ProductSpec{}).
				Where("id = ? AND company_id = ?", item.ID, companyID).
				Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("sort specs failed")
		response.ServerError(c, "排序失败")
		return
	}
	response.OkWithMessage(c, "排序成功", nil)
}

// ==================== 商品标签 ====================

type TagResp struct {
	model.ProductTag
	ProductCount int64 `json:"productCount"`
}

type tagReq struct {
	Name       string `json:"name" binding:"required,max=20"`
	Color      string `json:"color" binding:"omitempty,max=16"`
	Type       int8   `json:"type" binding:"omitempty,oneof=1 2"`
	Rule       string `json:"rule" binding:"omitempty,max=4000"`
	Filterable int8   `json:"filterable" binding:"omitempty,oneof=0 1"`
	ShowInList int8   `json:"showInList" binding:"omitempty,oneof=0 1"`
	ShowBadge  int8   `json:"showBadge" binding:"omitempty,oneof=0 1"`
	Sort       int    `json:"sort"`
	Status     int8   `json:"status" binding:"omitempty,oneof=0 1"`
}

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
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	type tagCount struct {
		TagID uint
		Cnt   int64
	}
	var counts []tagCount
	if len(list) > 0 {
		ids := make([]uint, 0, len(list))
		for _, t := range list {
			ids = append(ids, t.ID)
		}
		h.db.Model(&model.ProductTagRelation{}).
			Select("tag_id, COUNT(*) as cnt").
			Where("company_id = ? AND tag_id IN ?", companyID, ids).
			Group("tag_id").Scan(&counts)
	}
	countMap := make(map[uint]int64, len(counts))
	for _, cnt := range counts {
		countMap[cnt.TagID] = cnt.Cnt
	}
	resp := make([]TagResp, 0, len(list))
	for _, t := range list {
		var pc int64
		if t.Type == 1 {
			pc = countMap[t.ID]
		}
		resp = append(resp, TagResp{ProductTag: t, ProductCount: pc})
	}
	response.OkWithPage(c, resp, req.Page, req.PageSize, int(total))
}

func (h *ProductSettingHandler) CreateTag(c *gin.Context) {
	var req tagReq
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
	if req.Type == 0 {
		req.Type = 1
	}
	tag := model.ProductTag{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Color:                req.Color,
		Type:                 req.Type,
		Rule:                 req.Rule,
		Filterable:           req.Filterable,
		ShowInList:           req.ShowInList,
		ShowBadge:            req.ShowBadge,
		Sort:                 req.Sort,
		Status:               req.Status,
	}
	if err := h.db.Select("CompanyID", "Name", "Color", "Type", "Rule", "Filterable", "ShowInList", "ShowBadge", "Sort", "Status").Create(&tag).Error; err != nil {
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
	var req tagReq
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
	if req.Type == 1 || req.Type == 2 {
		tag.Type = req.Type
	}
	tag.Rule = req.Rule
	tag.Filterable = req.Filterable
	tag.ShowInList = req.ShowInList
	tag.ShowBadge = req.ShowBadge
	tag.Sort = req.Sort
	tag.Status = req.Status
	if err := h.db.Save(&tag).Error; err != nil {
		log.Error().Err(err).Msg("update product tag failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, tag)
}

func (h *ProductSettingHandler) DeleteTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := h.companyID(c)
	result := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.ProductTag{})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("delete product tag failed")
		response.ServerError(c, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, "标签不存在")
		return
	}
	if err := h.db.Where("company_id = ? AND tag_id = ?", companyID, id).Delete(&model.ProductTagRelation{}).Error; err != nil {
		log.Error().Err(err).Msg("delete tag relations failed")
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *ProductSettingHandler) SortTags(c *gin.Context) {
	var req SortUnitsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			if err := tx.Model(&model.ProductTag{}).
				Where("id = ? AND company_id = ?", item.ID, companyID).
				Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("sort tags failed")
		response.ServerError(c, "排序失败")
		return
	}
	response.OkWithMessage(c, "排序成功", nil)
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
