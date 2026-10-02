package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// BaseDataHandler 资料模块处理器（出入库类型/发货方式/物流公司/素材库/业务期初）
type BaseDataHandler struct {
	db *gorm.DB
}

func NewBaseDataHandler(db *gorm.DB) *BaseDataHandler {
	return &BaseDataHandler{db: db}
}

func (h *BaseDataHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/base-data")
	{
		g.GET("/stock-op-types", h.ListStockOpTypes)
		g.POST("/stock-op-types", h.CreateStockOpType)
		g.PUT("/stock-op-types/:id", h.UpdateStockOpType)
		g.DELETE("/stock-op-types/:id", h.DeleteStockOpType)

		g.GET("/delivery-methods", h.ListDeliveryMethods)
		g.POST("/delivery-methods", h.CreateDeliveryMethod)
		g.PUT("/delivery-methods/:id", h.UpdateDeliveryMethod)
		g.DELETE("/delivery-methods/:id", h.DeleteDeliveryMethod)

		g.GET("/logistics-companies", h.ListLogisticsCompanies)
		g.POST("/logistics-companies", h.CreateLogisticsCompany)
		g.PUT("/logistics-companies/:id", h.UpdateLogisticsCompany)
		g.DELETE("/logistics-companies/:id", h.DeleteLogisticsCompany)

		g.GET("/materials", h.ListMaterials)
		g.POST("/materials", h.CreateMaterial)
		g.DELETE("/materials/:id", h.DeleteMaterial)

		g.GET("/initial-stocks", h.ListInitialStocks)
		g.PUT("/initial-stocks", h.SaveInitialStock)
		g.GET("/initial-balances", h.ListInitialBalances)
		g.PUT("/initial-balances", h.SaveInitialBalance)
		g.GET("/initial-accounts", h.ListInitialAccounts)
		g.PUT("/initial-accounts", h.SaveInitialAccount)
	}
}

// ---------- 出入库类型 ----------

type baseDataListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
	Status   *int8  `form:"status"`
}

// ListStockOpTypes 出入库类型列表
// @Summary 出入库类型列表
// @Tags 资料
// @Param keyword query string false "名称/编号"
// @Param type query string false "in入库 out出库"
// @Router /api/v1/base-data/stock-op-types [get]
func (h *BaseDataHandler) ListStockOpTypes(c *gin.Context) {
	var req baseDataListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	q := h.db.Model(&model.StockOpType{}).Where("company_id = ?", middleware.GetCompanyID(c))
	if req.Keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		q = q.Where("type = ?", req.Type)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	var total int64
	q.Count(&total)
	var list []model.StockOpType
	q.Order("sort, id").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type stockOpTypeReq struct {
	Name   string `json:"name" binding:"required,max=64"`
	Code   string `json:"code" binding:"max=64"`
	Type   string `json:"type" binding:"required,oneof=in out"`
	Sort   int    `json:"sort"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

// CreateStockOpType 新增出入库类型
// @Summary 新增出入库类型
// @Tags 资料
// @Router /api/v1/base-data/stock-op-types [post]
func (h *BaseDataHandler) CreateStockOpType(c *gin.Context) {
	var req stockOpTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	item := model.StockOpType{
		Name: req.Name, Code: req.Code, Type: req.Type, Sort: req.Sort, Status: req.Status,
	}
	item.CompanyID = middleware.GetCompanyID(c)
	if err := h.db.Create(&item).Error; err != nil {
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, item)
}

// UpdateStockOpType 修改出入库类型
// @Summary 修改出入库类型
// @Tags 资料
// @Router /api/v1/base-data/stock-op-types/:id [put]
func (h *BaseDataHandler) UpdateStockOpType(c *gin.Context) {
	var req stockOpTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	var item model.StockOpType
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).First(&item).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	h.db.Model(&item).Updates(map[string]interface{}{
		"name": req.Name, "code": req.Code, "type": req.Type, "sort": req.Sort, "status": req.Status,
	})
	response.Ok(c, item)
}

// DeleteStockOpType 删除出入库类型
// @Summary 删除出入库类型
// @Tags 资料
// @Router /api/v1/base-data/stock-op-types/:id [delete]
func (h *BaseDataHandler) DeleteStockOpType(c *gin.Context) {
	h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).Delete(&model.StockOpType{})
	response.Ok(c, nil)
}

// ---------- 发货方式 ----------

// ListDeliveryMethods 发货方式列表
// @Summary 发货方式列表
// @Tags 资料
// @Router /api/v1/base-data/delivery-methods [get]
func (h *BaseDataHandler) ListDeliveryMethods(c *gin.Context) {
	var req baseDataListReq
	_ = c.ShouldBindQuery(&req)
	q := h.db.Model(&model.DeliveryMethod{}).Where("company_id = ?", middleware.GetCompanyID(c))
	if req.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.DeliveryMethod
	q.Order("sort, id").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type deliveryMethodReq struct {
	Name         string `json:"name" binding:"required,max=64"`
	NeedLogistic bool   `json:"needLogistic"`
	Sort         int    `json:"sort"`
	Status       int8   `json:"status" binding:"oneof=0 1"`
}

// CreateDeliveryMethod 新增发货方式
// @Summary 新增发货方式
// @Tags 资料
// @Router /api/v1/base-data/delivery-methods [post]
func (h *BaseDataHandler) CreateDeliveryMethod(c *gin.Context) {
	var req deliveryMethodReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	item := model.DeliveryMethod{Name: req.Name, NeedLogistic: req.NeedLogistic, Sort: req.Sort, Status: req.Status}
	item.CompanyID = middleware.GetCompanyID(c)
	if err := h.db.Create(&item).Error; err != nil {
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, item)
}

// UpdateDeliveryMethod 修改发货方式
// @Summary 修改发货方式
// @Tags 资料
// @Router /api/v1/base-data/delivery-methods/:id [put]
func (h *BaseDataHandler) UpdateDeliveryMethod(c *gin.Context) {
	var req deliveryMethodReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	var item model.DeliveryMethod
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).First(&item).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	h.db.Model(&item).Updates(map[string]interface{}{
		"name": req.Name, "need_logistic": req.NeedLogistic, "sort": req.Sort, "status": req.Status,
	})
	response.Ok(c, item)
}

// DeleteDeliveryMethod 删除发货方式
// @Summary 删除发货方式
// @Tags 资料
// @Router /api/v1/base-data/delivery-methods/:id [delete]
func (h *BaseDataHandler) DeleteDeliveryMethod(c *gin.Context) {
	h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).Delete(&model.DeliveryMethod{})
	response.Ok(c, nil)
}

// ---------- 物流公司 ----------

// ListLogisticsCompanies 物流公司列表
// @Summary 物流公司列表
// @Tags 资料
// @Router /api/v1/base-data/logistics-companies [get]
func (h *BaseDataHandler) ListLogisticsCompanies(c *gin.Context) {
	var req baseDataListReq
	_ = c.ShouldBindQuery(&req)
	q := h.db.Model(&model.LogisticsCompany{}).Where("company_id = ?", middleware.GetCompanyID(c))
	if req.Keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.LogisticsCompany
	q.Order("sort, id").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type logisticsCompanyReq struct {
	Name     string `json:"name" binding:"required,max=64"`
	Code     string `json:"code" binding:"max=64"`
	Tracking bool   `json:"tracking"`
	IsCustom bool   `json:"isCustom"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

// CreateLogisticsCompany 新增物流公司
// @Summary 新增物流公司
// @Tags 资料
// @Router /api/v1/base-data/logistics-companies [post]
func (h *BaseDataHandler) CreateLogisticsCompany(c *gin.Context) {
	var req logisticsCompanyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	item := model.LogisticsCompany{
		Name: req.Name, Code: req.Code, Tracking: req.Tracking, IsCustom: req.IsCustom, Sort: req.Sort, Status: req.Status,
	}
	item.CompanyID = middleware.GetCompanyID(c)
	if err := h.db.Create(&item).Error; err != nil {
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, item)
}

// UpdateLogisticsCompany 修改物流公司
// @Summary 修改物流公司
// @Tags 资料
// @Router /api/v1/base-data/logistics-companies/:id [put]
func (h *BaseDataHandler) UpdateLogisticsCompany(c *gin.Context) {
	var req logisticsCompanyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	var item model.LogisticsCompany
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).First(&item).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	h.db.Model(&item).Updates(map[string]interface{}{
		"name": req.Name, "code": req.Code, "tracking": req.Tracking, "is_custom": req.IsCustom, "sort": req.Sort, "status": req.Status,
	})
	response.Ok(c, item)
}

// DeleteLogisticsCompany 删除物流公司
// @Summary 删除物流公司
// @Tags 资料
// @Router /api/v1/base-data/logistics-companies/:id [delete]
func (h *BaseDataHandler) DeleteLogisticsCompany(c *gin.Context) {
	h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).Delete(&model.LogisticsCompany{})
	response.Ok(c, nil)
}

// ---------- 素材库 ----------

// ListMaterials 素材库列表
// @Summary 素材库列表
// @Tags 资料
// @Router /api/v1/base-data/materials [get]
func (h *BaseDataHandler) ListMaterials(c *gin.Context) {
	var req baseDataListReq
	_ = c.ShouldBindQuery(&req)
	q := h.db.Model(&model.Material{}).Where("company_id = ?", middleware.GetCompanyID(c))
	if req.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		q = q.Where("type = ?", req.Type)
	}
	var total int64
	q.Count(&total)
	var list []model.Material
	q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type materialReq struct {
	Name string `json:"name" binding:"required,max=128"`
	Type string `json:"type" binding:"required,oneof=image video file"`
	URL  string `json:"url" binding:"required,max=512"`
	Size int64  `json:"size"`
}

// CreateMaterial 新增素材
// @Summary 新增素材
// @Tags 资料
// @Router /api/v1/base-data/materials [post]
func (h *BaseDataHandler) CreateMaterial(c *gin.Context) {
	var req materialReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	item := model.Material{Name: req.Name, Type: req.Type, URL: req.URL, Size: req.Size}
	item.CompanyID = middleware.GetCompanyID(c)
	if err := h.db.Create(&item).Error; err != nil {
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, item)
}

// DeleteMaterial 删除素材
// @Summary 删除素材
// @Tags 资料
// @Router /api/v1/base-data/materials/:id [delete]
func (h *BaseDataHandler) DeleteMaterial(c *gin.Context) {
	h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).Delete(&model.Material{})
	response.Ok(c, nil)
}

// ---------- 业务期初 ----------

type initialStockRow struct {
	model.InitialStock
	ProductName   string `json:"productName"`
	ProductCode   string `json:"productCode"`
	WarehouseName string `json:"warehouseName"`
}

// ListInitialStocks 商品库存期初列表
// @Summary 商品库存期初列表
// @Tags 资料
// @Router /api/v1/base-data/initial-stocks [get]
func (h *BaseDataHandler) ListInitialStocks(c *gin.Context) {
	var req baseDataListReq
	_ = c.ShouldBindQuery(&req)
	q := h.db.Model(&model.InitialStock{}).
		Select(`initial_stocks.*, COALESCE(p.name,'') AS product_name, COALESCE(p.code,'') AS product_code, COALESCE(w.name,'') AS warehouse_name`).
		Joins("LEFT JOIN products p ON p.id = initial_stocks.product_id").
		Joins("LEFT JOIN warehouses w ON w.id = initial_stocks.warehouse_id").
		Where("initial_stocks.company_id = ?", middleware.GetCompanyID(c))
	if req.Keyword != "" {
		q = q.Where("p.name LIKE ? OR p.code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []initialStockRow
	q.Order("initial_stocks.id").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type initialStockReq struct {
	ProductID   uint    `json:"productId" binding:"required"`
	WarehouseID uint    `json:"warehouseId" binding:"required"`
	Quantity    float64 `json:"quantity"`
	CostPrice   float64 `json:"costPrice"`
}

// SaveInitialStock 保存商品库存期初（按 商品+仓库 幂等更新）
// @Summary 保存商品库存期初
// @Tags 资料
// @Router /api/v1/base-data/initial-stocks [put]
func (h *BaseDataHandler) SaveInitialStock(c *gin.Context) {
	var req initialStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var item model.InitialStock
	err := h.db.Where("company_id = ? AND product_id = ? AND warehouse_id = ?", companyID, req.ProductID, req.WarehouseID).First(&item).Error
	costAmount := req.Quantity * req.CostPrice
	if err != nil {
		item = model.InitialStock{
			ProductID: req.ProductID, WarehouseID: req.WarehouseID,
			Quantity: req.Quantity, CostPrice: req.CostPrice, CostAmount: costAmount,
		}
		item.CompanyID = companyID
		if cerr := h.db.Create(&item).Error; cerr != nil {
			response.ServerError(c, "保存失败")
			return
		}
	} else {
		h.db.Model(&item).Updates(map[string]interface{}{
			"quantity": req.Quantity, "cost_price": req.CostPrice, "cost_amount": costAmount,
		})
	}
	response.Ok(c, item)
}

type initialBalanceRow struct {
	model.InitialBalance
	TargetName string `json:"targetName"`
	TargetCode string `json:"targetCode"`
}

// ListInitialBalances 往来期初列表（应收预收/应付预付）
// @Summary 往来期初列表
// @Tags 资料
// @Param bizType query string true "customer客户 supplier供应商"
// @Router /api/v1/base-data/initial-balances [get]
func (h *BaseDataHandler) ListInitialBalances(c *gin.Context) {
	var req baseDataListReq
	_ = c.ShouldBindQuery(&req)
	bizType := c.DefaultQuery("bizType", "customer")
	companyID := middleware.GetCompanyID(c)
	var total int64
	var list []initialBalanceRow
	if bizType == "supplier" {
		q := h.db.Model(&model.InitialBalance{}).
			Select(`initial_balances.*, COALESCE(s.name,'') AS target_name, COALESCE(s.code,'') AS target_code`).
			Joins("LEFT JOIN suppliers s ON s.id = initial_balances.target_id").
			Where("initial_balances.company_id = ? AND initial_balances.biz_type = ?", companyID, bizType)
		if req.Keyword != "" {
			q = q.Where("s.name LIKE ? OR s.code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
		}
		q.Count(&total)
		q.Order("initial_balances.id").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	} else {
		q := h.db.Model(&model.InitialBalance{}).
			Select(`initial_balances.*, COALESCE(cu.name,'') AS target_name, COALESCE(cu.code,'') AS target_code`).
			Joins("LEFT JOIN customers cu ON cu.id = initial_balances.target_id").
			Where("initial_balances.company_id = ? AND initial_balances.biz_type = ?", companyID, bizType)
		if req.Keyword != "" {
			q = q.Where("cu.name LIKE ? OR cu.code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
		}
		q.Count(&total)
		q.Order("initial_balances.id").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	}
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type initialBalanceReq struct {
	BizType    string  `json:"bizType" binding:"required,oneof=customer supplier"`
	TargetID   uint    `json:"targetId" binding:"required"`
	Receivable float64 `json:"receivable"`
	Advance    float64 `json:"advance"`
}

// SaveInitialBalance 保存往来期初（按 对象+类型 幂等更新）
// @Summary 保存往来期初
// @Tags 资料
// @Router /api/v1/base-data/initial-balances [put]
func (h *BaseDataHandler) SaveInitialBalance(c *gin.Context) {
	var req initialBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var item model.InitialBalance
	err := h.db.Where("company_id = ? AND biz_type = ? AND target_id = ?", companyID, req.BizType, req.TargetID).First(&item).Error
	if err != nil {
		item = model.InitialBalance{BizType: req.BizType, TargetID: req.TargetID, Receivable: req.Receivable, Advance: req.Advance}
		item.CompanyID = companyID
		if cerr := h.db.Create(&item).Error; cerr != nil {
			response.ServerError(c, "保存失败")
			return
		}
	} else {
		h.db.Model(&item).Updates(map[string]interface{}{"receivable": req.Receivable, "advance": req.Advance})
	}
	response.Ok(c, item)
}

type initialAccountRow struct {
	model.InitialAccountBalance
	AccountName string `json:"accountName"`
	AccountType string `json:"accountType"`
}

// ListInitialAccounts 现金银行期初列表
// @Summary 现金银行期初列表
// @Tags 资料
// @Router /api/v1/base-data/initial-accounts [get]
func (h *BaseDataHandler) ListInitialAccounts(c *gin.Context) {
	var req baseDataListReq
	_ = c.ShouldBindQuery(&req)
	q := h.db.Model(&model.InitialAccountBalance{}).
		Select(`initial_account_balances.*, COALESCE(a.name,'') AS account_name, COALESCE(a.type,'') AS account_type`).
		Joins("LEFT JOIN accounts a ON a.id = initial_account_balances.account_id").
		Where("initial_account_balances.company_id = ?", middleware.GetCompanyID(c))
	if req.Keyword != "" {
		q = q.Where("a.name LIKE ?", "%"+req.Keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []initialAccountRow
	q.Order("initial_account_balances.id").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type initialAccountReq struct {
	AccountID uint    `json:"accountId" binding:"required"`
	Amount    float64 `json:"amount"`
}

// SaveInitialAccount 保存现金银行期初（按账户幂等更新）
// @Summary 保存现金银行期初
// @Tags 资料
// @Router /api/v1/base-data/initial-accounts [put]
func (h *BaseDataHandler) SaveInitialAccount(c *gin.Context) {
	var req initialAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var item model.InitialAccountBalance
	err := h.db.Where("company_id = ? AND account_id = ?", companyID, req.AccountID).First(&item).Error
	if err != nil {
		item = model.InitialAccountBalance{AccountID: req.AccountID, Amount: req.Amount}
		item.CompanyID = companyID
		if cerr := h.db.Create(&item).Error; cerr != nil {
			response.ServerError(c, "保存失败")
			return
		}
	} else {
		h.db.Model(&item).Updates(map[string]interface{}{"amount": req.Amount})
	}
	response.Ok(c, item)
}
