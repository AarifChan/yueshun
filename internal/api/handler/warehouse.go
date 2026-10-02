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

// WarehouseHandler 仓库/资金处理器
type WarehouseHandler struct {
	db *gorm.DB
}

func NewWarehouseHandler(db *gorm.DB) *WarehouseHandler {
	return &WarehouseHandler{db: db}
}

func (h *WarehouseHandler) RegisterRoutes(r *gin.RouterGroup) {
	wh := r.Group("/warehouses")
	{
		// 仓库
		wh.GET("", h.ListWarehouses)
		wh.POST("", h.CreateWarehouse)
		wh.PUT("/:id", h.UpdateWarehouse)
		wh.DELETE("/:id", h.DeleteWarehouse)

		// 仓位
		wh.GET("/:id/positions", h.ListPositions)
		wh.POST("/:id/positions", h.CreatePosition)
		wh.PUT("/:id/positions/:pid", h.UpdatePosition)
		wh.DELETE("/:id/positions/:pid", h.DeletePosition)
	}

	// 资金账户
	account := r.Group("/accounts")
	{
		account.GET("", h.ListAccounts)
		account.POST("", h.CreateAccount)
		account.PUT("/:id", h.UpdateAccount)
		account.DELETE("/:id", h.DeleteAccount)
		account.GET("/:id", h.GetAccount)
		account.GET("/:id/flows", h.ListAccountFlows)
	}

	// 库存台账
	stocks := r.Group("/stocks")
	{
		stocks.GET("", h.ListStocks)
		stocks.POST("/import", h.ImportStockStatus)
		stocks.GET("/status", h.ListStockStatus)
		stocks.PUT("/status/limits", h.UpdateStockLimits)
		stocks.GET("/status/:productId/warehouses", h.ListStockDistribution)
		stocks.GET("/status/:productId/flows", h.ListStockFlows)
	}

	// 收支项目
	ie := r.Group("/income-expense-items")
	{
		ie.GET("", h.ListIEItems)
		ie.POST("", h.CreateIEItem)
		ie.PUT("/:id", h.UpdateIEItem)
		ie.DELETE("/:id", h.DeleteIEItem)
	}
}

// ==================== 仓库 ====================

type WarehouseListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

func (h *WarehouseHandler) ListWarehouses(c *gin.Context) {
	var req WarehouseListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Warehouse{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.Warehouse
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *WarehouseHandler) CreateWarehouse(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required,max=64"`
		Code      string `json:"code" binding:"max=64"`
		Address   string `json:"address" binding:"max=255"`
		ManagerID uint   `json:"managerId"`
		Status    int8   `json:"status" binding:"oneof=0 1"`
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

	wh := model.Warehouse{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Address:              req.Address,
		ManagerID:            req.ManagerID,
		Status:               req.Status,
	}
	if err := h.db.Create(&wh).Error; err != nil {
		log.Error().Err(err).Msg("create warehouse failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, wh)
}

func (h *WarehouseHandler) UpdateWarehouse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name      string `json:"name" binding:"max=64"`
		Code      string `json:"code" binding:"max=64"`
		Address   string `json:"address" binding:"max=255"`
		ManagerID uint   `json:"managerId"`
		Status    int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var wh model.Warehouse
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&wh).Error; err != nil {
		response.NotFound(c, "仓库不存在")
		return
	}

	if req.Name != "" {
		wh.Name = req.Name
	}
	if req.Code != "" {
		wh.Code = req.Code
	}
	wh.Address = req.Address
	wh.ManagerID = req.ManagerID
	wh.Status = req.Status

	if err := h.db.Save(&wh).Error; err != nil {
		log.Error().Err(err).Msg("update warehouse failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, wh)
}

func (h *WarehouseHandler) DeleteWarehouse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var count int64
	h.db.Model(&model.WarehousePosition{}).Where("warehouse_id = ?", id).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "该仓库下存在仓位，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Warehouse{}).Error; err != nil {
		log.Error().Err(err).Msg("delete warehouse failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 仓位 ====================

func (h *WarehouseHandler) ListPositions(c *gin.Context) {
	warehouseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "仓库ID格式错误")
		return
	}

	var list []model.WarehousePosition
	h.db.Where("warehouse_id = ? AND status = 1", warehouseID).Order("created_at DESC").Find(&list)
	response.Ok(c, list)
}

func (h *WarehouseHandler) CreatePosition(c *gin.Context) {
	warehouseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "仓库ID格式错误")
		return
	}
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Code   string `json:"code" binding:"max=64"`
		Status int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	pos := model.WarehousePosition{
		WarehouseID: uint(warehouseID),
		Name:        req.Name,
		Code:        req.Code,
		Status:      req.Status,
	}
	if err := h.db.Create(&pos).Error; err != nil {
		log.Error().Err(err).Msg("create position failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, pos)
}

func (h *WarehouseHandler) UpdatePosition(c *gin.Context) {
	pid, err := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "仓位ID格式错误")
		return
	}
	var req struct {
		Name   string `json:"name" binding:"max=64"`
		Code   string `json:"code" binding:"max=64"`
		Status int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var pos model.WarehousePosition
	if err := h.db.First(&pos, pid).Error; err != nil {
		response.NotFound(c, "仓位不存在")
		return
	}

	if req.Name != "" {
		pos.Name = req.Name
	}
	if req.Code != "" {
		pos.Code = req.Code
	}
	pos.Status = req.Status

	if err := h.db.Save(&pos).Error; err != nil {
		log.Error().Err(err).Msg("update position failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, pos)
}

func (h *WarehouseHandler) DeletePosition(c *gin.Context) {
	pid, err := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "仓位ID格式错误")
		return
	}

	if err := h.db.Delete(&model.WarehousePosition{}, pid).Error; err != nil {
		log.Error().Err(err).Msg("delete position failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 资金账户 ====================

type AccountListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
}

func (h *WarehouseHandler) ListAccounts(c *gin.Context) {
	var req AccountListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Account{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	var total int64
	query.Count(&total)

	var list []model.Account
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *WarehouseHandler) CreateAccount(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required,max=64"`
		Code        string  `json:"code" binding:"max=64"`
		Type        string  `json:"type" binding:"required,oneof=cash bank alipay wechat"`
		Balance     float64 `json:"balance"`
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

	account := model.Account{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Type:                 req.Type,
		Balance:              req.Balance,
		Description:          req.Description,
		Status:               req.Status,
	}
	if err := h.db.Create(&account).Error; err != nil {
		log.Error().Err(err).Msg("create account failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, account)
}

func (h *WarehouseHandler) GetAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var account model.Account
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&account).Error; err != nil {
		response.NotFound(c, "账户不存在")
		return
	}
	response.Ok(c, account)
}

func (h *WarehouseHandler) UpdateAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"max=64"`
		Code        string  `json:"code" binding:"max=64"`
		Type        string  `json:"type" binding:"oneof=cash bank alipay wechat"`
		Balance     float64 `json:"balance"`
		Description string  `json:"description" binding:"max=255"`
		Status      int8    `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var account model.Account
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&account).Error; err != nil {
		response.NotFound(c, "账户不存在")
		return
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	if req.Code != "" {
		account.Code = req.Code
	}
	if req.Type != "" {
		account.Type = req.Type
	}
	account.Balance = req.Balance
	account.Description = req.Description
	account.Status = req.Status

	if err := h.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("update account failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, account)
}

func (h *WarehouseHandler) DeleteAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Account{}).Error; err != nil {
		log.Error().Err(err).Msg("delete account failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 资金流水 ====================

type AccountFlowListReq struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"pageSize,default=20"`
}

func (h *WarehouseHandler) ListAccountFlows(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req AccountFlowListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.AccountFlow{}).Where("account_id = ? AND company_id = ?", id, companyID)

	var total int64
	query.Count(&total)

	var list []model.AccountFlow
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// ==================== 库存台账 ====================

type StockListReq struct {
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=20"`
	WarehouseID uint   `form:"warehouseId"`
	ProductID   uint   `form:"productId"`
	Keyword     string `form:"keyword"`
}

type StockResp struct {
	model.Stock
	ProductName     string `json:"productName"`
	ProductCode     string `json:"productCode"`
	ProductUnit     string `json:"productUnit"`
	WarehouseName   string `json:"warehouseName"`
}

func (h *WarehouseHandler) ListStocks(c *gin.Context) {
	var req StockListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Stock{}).Where("stocks.company_id = ?", companyID)
	if req.WarehouseID > 0 {
		query = query.Where("stocks.warehouse_id = ?", req.WarehouseID)
	}
	if req.ProductID > 0 {
		query = query.Where("stocks.product_id = ?", req.ProductID)
	}
	if req.Keyword != "" {
		query = query.Where("products.name LIKE ? OR products.code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Joins("LEFT JOIN products ON products.id = stocks.product_id").
		Distinct("stocks.id").Count(&total)

	// 列表查询使用独立链，避免与 Count 共用 Statement 导致 DISTINCT 残留
	listQuery := h.db.Model(&model.Stock{}).Where("stocks.company_id = ?", companyID)
	if req.WarehouseID > 0 {
		listQuery = listQuery.Where("stocks.warehouse_id = ?", req.WarehouseID)
	}
	if req.ProductID > 0 {
		listQuery = listQuery.Where("stocks.product_id = ?", req.ProductID)
	}
	if req.Keyword != "" {
		listQuery = listQuery.Where("products.name LIKE ? OR products.code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var list []StockResp
	if err := listQuery.Select("stocks.*, products.name as product_name, products.code as product_code, products.unit as product_unit, warehouses.name as warehouse_name").
		Joins("LEFT JOIN products ON products.id = stocks.product_id").
		Joins("LEFT JOIN warehouses ON warehouses.id = stocks.warehouse_id").
		Order("stocks.updated_at DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list).Error; err != nil {
		log.Error().Err(err).Msg("list stocks failed")
		response.ServerError(c, "查询失败")
		return
	}

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// ==================== 收支项目 ====================

type IEItemListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
}

func (h *WarehouseHandler) ListIEItems(c *gin.Context) {
	var req IEItemListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.IncomeExpenseItem{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	var total int64
	query.Count(&total)

	var list []model.IncomeExpenseItem
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *WarehouseHandler) CreateIEItem(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required,max=64"`
		Code   string `json:"code" binding:"max=64"`
		Type   string `json:"type" binding:"required,oneof=income expense"`
		Sort   int    `json:"sort"`
		Status int8   `json:"status" binding:"oneof=0 1"`
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

	item := model.IncomeExpenseItem{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Type:                 req.Type,
		Sort:                 req.Sort,
		Status:               req.Status,
	}
	if err := h.db.Create(&item).Error; err != nil {
		log.Error().Err(err).Msg("create income expense item failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, item)
}

func (h *WarehouseHandler) UpdateIEItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Name   string `json:"name" binding:"max=64"`
		Code   string `json:"code" binding:"max=64"`
		Type   string `json:"type" binding:"oneof=income expense"`
		Sort   int    `json:"sort"`
		Status int8   `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var item model.IncomeExpenseItem
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&item).Error; err != nil {
		response.NotFound(c, "项目不存在")
		return
	}

	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Code != "" {
		item.Code = req.Code
	}
	if req.Type != "" {
		item.Type = req.Type
	}
	item.Sort = req.Sort
	item.Status = req.Status

	if err := h.db.Save(&item).Error; err != nil {
		log.Error().Err(err).Msg("update income expense item failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, item)
}

func (h *WarehouseHandler) DeleteIEItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.IncomeExpenseItem{}).Error; err != nil {
		log.Error().Err(err).Msg("delete income expense item failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}
