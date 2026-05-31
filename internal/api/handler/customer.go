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

// CustomerHandler 客户/供应商处理器
type CustomerHandler struct {
	db *gorm.DB
}

func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

func (h *CustomerHandler) RegisterRoutes(r *gin.RouterGroup) {
	customer := r.Group("/customers")
	{
		// 客户分类
		customer.GET("/categories", h.ListCategories)
		customer.POST("/categories", h.CreateCategory)
		customer.PUT("/categories/:id", h.UpdateCategory)
		customer.DELETE("/categories/:id", h.DeleteCategory)

		// 区域
		customer.GET("/regions", h.ListRegions)
		customer.POST("/regions", h.CreateRegion)
		customer.PUT("/regions/:id", h.UpdateRegion)
		customer.DELETE("/regions/:id", h.DeleteRegion)

		// 客户等级
		customer.GET("/levels", h.ListLevels)
		customer.POST("/levels", h.CreateLevel)
		customer.PUT("/levels/:id", h.UpdateLevel)
		customer.DELETE("/levels/:id", h.DeleteLevel)

		// 客户
		customer.GET("", h.ListCustomers)
		customer.POST("", h.CreateCustomer)
		customer.GET("/:id", h.GetCustomer)
		customer.PUT("/:id", h.UpdateCustomer)
		customer.DELETE("/:id", h.DeleteCustomer)
		customer.PUT("/:id/status", h.UpdateCustomerStatus)
	}
}

// ==================== 客户分类 ====================

type CustomerCategoryListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

func (h *CustomerHandler) ListCategories(c *gin.Context) {
	var req CustomerCategoryListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.CustomerCategory{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.CustomerCategory
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *CustomerHandler) CreateCategory(c *gin.Context) {
	var req struct {
		ParentID uint   `json:"parentId"`
		Name     string `json:"name" binding:"required,max=64"`
		Code     string `json:"code" binding:"max=64"`
		Sort     int    `json:"sort"`
		Status   int8   `json:"status" binding:"oneof=0 1"`
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

	if req.ParentID > 0 {
		var parent model.CustomerCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父分类不存在")
			return
		}
	}

	cat := model.CustomerCategory{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		ParentID:             req.ParentID,
		Name:                 req.Name,
		Code:                 req.Code,
		Sort:                 req.Sort,
		Status:               req.Status,
	}
	if err := h.db.Create(&cat).Error; err != nil {
		log.Error().Err(err).Msg("create customer category failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, cat)
}

func (h *CustomerHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		ParentID uint   `json:"parentId"`
		Name     string `json:"name" binding:"max=64"`
		Code     string `json:"code" binding:"max=64"`
		Sort     int    `json:"sort"`
		Status   int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var cat model.CustomerCategory
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&cat).Error; err != nil {
		response.NotFound(c, "分类不存在")
		return
	}

	if req.ParentID > 0 {
		if uint(id) == req.ParentID {
			response.BadRequest(c, "不能将自身设为父分类")
			return
		}
		var parent model.CustomerCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父分类不存在")
			return
		}
	}

	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Code != "" {
		cat.Code = req.Code
	}
	cat.ParentID = req.ParentID
	cat.Sort = req.Sort
	cat.Status = req.Status

	if err := h.db.Save(&cat).Error; err != nil {
		log.Error().Err(err).Msg("update customer category failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, cat)
}

func (h *CustomerHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var childCount int64
	h.db.Model(&model.CustomerCategory{}).Where("parent_id = ? AND company_id = ?", id, companyID).Count(&childCount)
	if childCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该分类下存在子分类，无法删除")
		return
	}

	var customerCount int64
	h.db.Model(&model.Customer{}).Where("category_id = ? AND company_id = ?", id, companyID).Count(&customerCount)
	if customerCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该分类下存在客户，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.CustomerCategory{}).Error; err != nil {
		log.Error().Err(err).Msg("delete customer category failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 区域 ====================

type RegionListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

func (h *CustomerHandler) ListRegions(c *gin.Context) {
	var req RegionListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Region{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.Region
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *CustomerHandler) CreateRegion(c *gin.Context) {
	var req struct {
		ParentID uint   `json:"parentId"`
		Name     string `json:"name" binding:"required,max=64"`
		Code     string `json:"code" binding:"max=64"`
		Sort     int    `json:"sort"`
		Status   int8   `json:"status" binding:"oneof=0 1"`
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

	region := model.Region{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		ParentID:             req.ParentID,
		Name:                 req.Name,
		Code:                 req.Code,
		Sort:                 req.Sort,
		Status:               req.Status,
	}
	if err := h.db.Create(&region).Error; err != nil {
		log.Error().Err(err).Msg("create region failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, region)
}

func (h *CustomerHandler) UpdateRegion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		ParentID uint   `json:"parentId"`
		Name     string `json:"name" binding:"max=64"`
		Code     string `json:"code" binding:"max=64"`
		Sort     int    `json:"sort"`
		Status   int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var region model.Region
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&region).Error; err != nil {
		response.NotFound(c, "区域不存在")
		return
	}

	if req.Name != "" {
		region.Name = req.Name
	}
	if req.Code != "" {
		region.Code = req.Code
	}
	region.ParentID = req.ParentID
	region.Sort = req.Sort
	region.Status = req.Status

	if err := h.db.Save(&region).Error; err != nil {
		log.Error().Err(err).Msg("update region failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, region)
}

func (h *CustomerHandler) DeleteRegion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var customerCount int64
	h.db.Model(&model.Customer{}).Where("region_id = ? AND company_id = ?", id, companyID).Count(&customerCount)
	if customerCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该区域下存在客户，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Region{}).Error; err != nil {
		log.Error().Err(err).Msg("delete region failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 客户等级 ====================

type LevelListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

func (h *CustomerHandler) ListLevels(c *gin.Context) {
	var req LevelListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.CustomerLevel{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.CustomerLevel
	query.Order("min_amount ASC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *CustomerHandler) CreateLevel(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required,max=64"`
		Code        string  `json:"code" binding:"max=64"`
		MinAmount   float64 `json:"minAmount"`
		MaxAmount   float64 `json:"maxAmount"`
		Discount    float64 `json:"discount"`
		Description string  `json:"description" binding:"max=255"`
		Status      int8    `json:"status" binding:"oneof=0 1"`
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

	level := model.CustomerLevel{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		MinAmount:            req.MinAmount,
		MaxAmount:            req.MaxAmount,
		Discount:             req.Discount,
		Description:          req.Description,
		Status:               req.Status,
	}
	if err := h.db.Create(&level).Error; err != nil {
		log.Error().Err(err).Msg("create customer level failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, level)
}

func (h *CustomerHandler) UpdateLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"max=64"`
		Code        string  `json:"code" binding:"max=64"`
		MinAmount   float64 `json:"minAmount"`
		MaxAmount   float64 `json:"maxAmount"`
		Discount    float64 `json:"discount"`
		Description string  `json:"description" binding:"max=255"`
		Status      int8    `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var level model.CustomerLevel
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&level).Error; err != nil {
		response.NotFound(c, "等级不存在")
		return
	}

	if req.Name != "" {
		level.Name = req.Name
	}
	if req.Code != "" {
		level.Code = req.Code
	}
	level.MinAmount = req.MinAmount
	level.MaxAmount = req.MaxAmount
	level.Discount = req.Discount
	level.Description = req.Description
	level.Status = req.Status

	if err := h.db.Save(&level).Error; err != nil {
		log.Error().Err(err).Msg("update customer level failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, level)
}

func (h *CustomerHandler) DeleteLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var customerCount int64
	h.db.Model(&model.Customer{}).Where("level_id = ? AND company_id = ?", id, companyID).Count(&customerCount)
	if customerCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该等级下存在客户，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.CustomerLevel{}).Error; err != nil {
		log.Error().Err(err).Msg("delete customer level failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 客户/供应商 ====================

type CustomerListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	Type       string `form:"type"`
	CategoryID uint   `form:"categoryId"`
	RegionID   uint   `form:"regionId"`
	LevelID    uint   `form:"levelId"`
	Status     int8   `form:"status"`
}

type CustomerResp struct {
	model.Customer
	CategoryName string `json:"categoryName"`
	RegionName   string `json:"regionName"`
	LevelName    string `json:"levelName"`
}

func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	var req CustomerListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Customer{}).Where("customers.company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("customers.name LIKE ? OR customers.code LIKE ? OR customers.phone LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		query = query.Where("customers.type = ?", req.Type)
	}
	if req.CategoryID > 0 {
		query = query.Where("customers.category_id = ?", req.CategoryID)
	}
	if req.RegionID > 0 {
		query = query.Where("customers.region_id = ?", req.RegionID)
	}
	if req.LevelID > 0 {
		query = query.Where("customers.level_id = ?", req.LevelID)
	}
	if req.Status == 0 || req.Status == 1 {
		query = query.Where("customers.status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []CustomerResp
	query.Select("customers.*, customer_categories.name as category_name, regions.name as region_name, customer_levels.name as level_name").
		Joins("LEFT JOIN customer_categories ON customer_categories.id = customers.category_id").
		Joins("LEFT JOIN regions ON regions.id = customers.region_id").
		Joins("LEFT JOIN customer_levels ON customer_levels.id = customers.level_id").
		Order("customers.created_at DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var customer model.Customer
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&customer).Error; err != nil {
		response.NotFound(c, "客户不存在")
		return
	}
	response.Ok(c, customer)
}

func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req struct {
		CategoryID  uint    `json:"categoryId"`
		RegionID    uint    `json:"regionId"`
		LevelID     uint    `json:"levelId"`
		Name        string  `json:"name" binding:"required,max=128"`
		Code        string  `json:"code" binding:"max=64"`
		Type        string  `json:"type" binding:"required,oneof=customer supplier both"`
		Contact     string  `json:"contact" binding:"max=64"`
		Phone       string  `json:"phone" binding:"max=20"`
		Email       string  `json:"email" binding:"max=128"`
		Address     string  `json:"address" binding:"max=255"`
		CreditLimit float64 `json:"creditLimit"`
		CreditDays  int     `json:"creditDays"`
		TaxNo       string  `json:"taxNo" binding:"max=64"`
		BankName    string  `json:"bankName" binding:"max=128"`
		BankAccount string  `json:"bankAccount" binding:"max=64"`
		Remark      string  `json:"remark" binding:"max=500"`
		Status      int8    `json:"status" binding:"oneof=0 1"`
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

	// 验证分类
	if req.CategoryID > 0 {
		var cat model.CustomerCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.CategoryID, companyID).First(&cat).Error; err != nil {
			response.BadRequest(c, "客户分类不存在")
			return
		}
	}
	// 验证区域
	if req.RegionID > 0 {
		var region model.Region
		if err := h.db.Where("id = ? AND company_id = ?", req.RegionID, companyID).First(&region).Error; err != nil {
			response.BadRequest(c, "区域不存在")
			return
		}
	}
	// 验证等级
	if req.LevelID > 0 {
		var level model.CustomerLevel
		if err := h.db.Where("id = ? AND company_id = ?", req.LevelID, companyID).First(&level).Error; err != nil {
			response.BadRequest(c, "客户等级不存在")
			return
		}
	}

	customer := model.Customer{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CategoryID:           req.CategoryID,
		RegionID:             req.RegionID,
		LevelID:              req.LevelID,
		Name:                 req.Name,
		Code:                 req.Code,
		Type:                 req.Type,
		Contact:              req.Contact,
		Phone:                req.Phone,
		Email:                req.Email,
		Address:              req.Address,
		CreditLimit:          req.CreditLimit,
		CreditDays:           req.CreditDays,
		TaxNo:                req.TaxNo,
		BankName:             req.BankName,
		BankAccount:          req.BankAccount,
		Remark:               req.Remark,
		Status:               req.Status,
	}
	if err := h.db.Create(&customer).Error; err != nil {
		log.Error().Err(err).Msg("create customer failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, customer)
}

func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		CategoryID  uint    `json:"categoryId"`
		RegionID    uint    `json:"regionId"`
		LevelID     uint    `json:"levelId"`
		Name        string  `json:"name" binding:"max=128"`
		Code        string  `json:"code" binding:"max=64"`
		Type        string  `json:"type" binding:"oneof=customer supplier both"`
		Contact     string  `json:"contact" binding:"max=64"`
		Phone       string  `json:"phone" binding:"max=20"`
		Email       string  `json:"email" binding:"max=128"`
		Address     string  `json:"address" binding:"max=255"`
		CreditLimit float64 `json:"creditLimit"`
		CreditDays  int     `json:"creditDays"`
		TaxNo       string  `json:"taxNo" binding:"max=64"`
		BankName    string  `json:"bankName" binding:"max=128"`
		BankAccount string  `json:"bankAccount" binding:"max=64"`
		Remark      string  `json:"remark" binding:"max=500"`
		Status      int8    `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var customer model.Customer
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&customer).Error; err != nil {
		response.NotFound(c, "客户不存在")
		return
	}

	if req.CategoryID > 0 {
		var cat model.CustomerCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.CategoryID, companyID).First(&cat).Error; err != nil {
			response.BadRequest(c, "客户分类不存在")
			return
		}
		customer.CategoryID = req.CategoryID
	}
	if req.RegionID > 0 {
		var region model.Region
		if err := h.db.Where("id = ? AND company_id = ?", req.RegionID, companyID).First(&region).Error; err != nil {
			response.BadRequest(c, "区域不存在")
			return
		}
		customer.RegionID = req.RegionID
	}
	if req.LevelID > 0 {
		var level model.CustomerLevel
		if err := h.db.Where("id = ? AND company_id = ?", req.LevelID, companyID).First(&level).Error; err != nil {
			response.BadRequest(c, "客户等级不存在")
			return
		}
		customer.LevelID = req.LevelID
	}
	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Code != "" {
		customer.Code = req.Code
	}
	if req.Type != "" {
		customer.Type = req.Type
	}
	if req.Contact != "" {
		customer.Contact = req.Contact
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.Address != "" {
		customer.Address = req.Address
	}
	customer.CreditLimit = req.CreditLimit
	customer.CreditDays = req.CreditDays
	if req.TaxNo != "" {
		customer.TaxNo = req.TaxNo
	}
	if req.BankName != "" {
		customer.BankName = req.BankName
	}
	if req.BankAccount != "" {
		customer.BankAccount = req.BankAccount
	}
	if req.Remark != "" {
		customer.Remark = req.Remark
	}
	customer.Status = req.Status

	if err := h.db.Save(&customer).Error; err != nil {
		log.Error().Err(err).Msg("update customer failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, customer)
}

func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 简化：暂不检查业务单据关联
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Customer{}).Error; err != nil {
		log.Error().Err(err).Msg("delete customer failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *CustomerHandler) UpdateCustomerStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Status int8 `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Model(&model.Customer{}).Where("id = ? AND company_id = ?", id, companyID).Update("status", req.Status).Error; err != nil {
		log.Error().Err(err).Msg("update customer status failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}
