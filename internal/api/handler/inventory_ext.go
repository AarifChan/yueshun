package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/database"
	"zhizhang-server/internal/pkg/response"
)

// InventoryExtHandler 库存扩展处理器（成本调价单/拆装模板/组装拆装单）
type InventoryExtHandler struct {
	db *gorm.DB
}

func NewInventoryExtHandler(db *gorm.DB) *InventoryExtHandler {
	return &InventoryExtHandler{db: db}
}

func (h *InventoryExtHandler) RegisterRoutes(r *gin.RouterGroup) {
	ca := r.Group("/cost-adjusts")
	{
		ca.GET("", h.ListCostAdjusts)
		ca.POST("", h.CreateCostAdjust)
		ca.GET("/:id", h.GetCostAdjust)
		ca.PUT("/:id", h.UpdateCostAdjust)
		ca.DELETE("/:id", h.DeleteCostAdjust)
		ca.PUT("/:id/complete", h.CompleteCostAdjust)
	}
	tpl := r.Group("/assembly-templates")
	{
		tpl.GET("", h.ListAssemblyTemplates)
		tpl.POST("", h.CreateAssemblyTemplate)
		tpl.GET("/:id", h.GetAssemblyTemplate)
		tpl.PUT("/:id", h.UpdateAssemblyTemplate)
		tpl.DELETE("/:id", h.DeleteAssemblyTemplate)
	}
	ao := r.Group("/assembly-orders")
	{
		ao.GET("", h.ListAssemblyOrders)
		ao.POST("", h.CreateAssemblyOrder)
		ao.GET("/:id", h.GetAssemblyOrder)
		ao.PUT("/:id", h.UpdateAssemblyOrder)
		ao.DELETE("/:id", h.DeleteAssemblyOrder)
		ao.PUT("/:id/complete", h.CompleteAssemblyOrder)
	}
}

// ==================== 成本调价单 ====================

type costAdjustListReq struct {
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=20"`
	Keyword     string `form:"keyword"`
	ProductKw   string `form:"productKw"`
	WarehouseID uint   `form:"warehouseId"`
	Status      string `form:"status"`
	ShowCancelled bool `form:"showCancelled"`
	StartDate   string `form:"startDate"`
	EndDate     string `form:"endDate"`
}

type costAdjustRow struct {
	model.CostAdjust
	WarehouseName string `json:"warehouseName"`
	HandlerName   string `json:"handlerName"`
	OperatorName  string `json:"operatorName"`
}

// ListCostAdjusts 成本调价单列表
// @Summary 成本调价单列表
// @Tags 库存
// @Param keyword query string false "单号"
// @Param warehouseId query int false "仓库"
// @Param status query string false "状态"
// @Router /api/v1/cost-adjusts [get]
func (h *InventoryExtHandler) ListCostAdjusts(c *gin.Context) {
	var req costAdjustListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	q := h.db.Model(&model.CostAdjust{}).
		Select(`cost_adjusts.*, COALESCE(w.name,'') AS warehouse_name,
			COALESCE(hd.name,'') AS handler_name, COALESCE(op.name,'') AS operator_name`).
		Joins("LEFT JOIN warehouses w ON w.id = cost_adjusts.warehouse_id").
		Joins("LEFT JOIN employees hd ON hd.id = cost_adjusts.handler_id").
		Joins("LEFT JOIN employees op ON op.id = cost_adjusts.operator_id").
		Where("cost_adjusts.company_id = ?", companyID)
	if !req.ShowCancelled {
		q = q.Where("cost_adjusts.status != 'cancelled'")
	}
	if req.Keyword != "" {
		q = q.Where("cost_adjusts.bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.WarehouseID > 0 {
		q = q.Where("cost_adjusts.warehouse_id = ?", req.WarehouseID)
	}
	if req.Status != "" {
		q = q.Where("cost_adjusts.status = ?", req.Status)
	}
	if req.StartDate != "" {
		q = q.Where("cost_adjusts.created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		q = q.Where("cost_adjusts.created_at <= ?", req.EndDate+" 23:59:59")
	}
	if req.ProductKw != "" {
		q = q.Where(`EXISTS (SELECT 1 FROM cost_adjust_items ci JOIN products p ON p.id = ci.product_id
			WHERE ci.adjust_id = cost_adjusts.id AND (p.name LIKE ? OR p.code LIKE ?))`,
			"%"+req.ProductKw+"%", "%"+req.ProductKw+"%")
	}
	var total int64
	q.Count(&total)
	var list []costAdjustRow
	q.Order("cost_adjusts.id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// GetCostAdjust 成本调价单详情
// @Summary 成本调价单详情
// @Tags 库存
// @Router /api/v1/cost-adjusts/:id [get]
func (h *InventoryExtHandler) GetCostAdjust(c *gin.Context) {
	var bill model.CostAdjust
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).
		Preload("Items.Product").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	response.Ok(c, bill)
}

type costAdjustItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity"`
	OldPrice  float64 `json:"oldPrice"`
	NewPrice  float64 `json:"newPrice" binding:"required"`
	Remark    string  `json:"remark"`
}

type costAdjustReq struct {
	WarehouseID uint                `json:"warehouseId" binding:"required"`
	BillDate    string              `json:"billDate"`
	HandlerID   uint                `json:"handlerId"`
	RefBillNo   string              `json:"refBillNo"`
	Remark      string              `json:"remark"`
	Items       []costAdjustItemReq `json:"items" binding:"required,min=1"`
}

func (r *costAdjustReq) billDate() time.Time {
	if t, err := time.Parse("2006-01-02", r.BillDate); err == nil {
		return t
	}
	return time.Now()
}

// CreateCostAdjust 新增成本调价单
// @Summary 新增成本调价单
// @Tags 库存
// @Router /api/v1/cost-adjusts [post]
func (h *InventoryExtHandler) CreateCostAdjust(c *gin.Context) {
	var req costAdjustReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)

	var bill model.CostAdjust
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		bill = model.CostAdjust{
			WarehouseID: req.WarehouseID,
			BillNo:      genDailyBillNo(tx, "cost_adjusts", "CBTJ"),
			BillDate:    req.billDate(),
			Status:      "draft",
			HandlerID:   req.HandlerID,
			OperatorID:  userID,
			RefBillNo:   req.RefBillNo,
			Remark:      req.Remark,
		}
		bill.CompanyID = companyID
		var total float64
		for _, it := range req.Items {
			diff := (it.NewPrice - it.OldPrice) * it.Quantity
			total += diff
			bill.Items = append(bill.Items, model.CostAdjustItem{
				ProductID: it.ProductID, Quantity: it.Quantity,
				OldPrice: it.OldPrice, NewPrice: it.NewPrice, DiffAmount: diff, Remark: it.Remark,
			})
		}
		bill.Amount = total
		return tx.Create(&bill).Error
	}); err != nil {
		log.Error().Err(err).Msg("create cost adjust failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, bill)
}

// UpdateCostAdjust 修改成本调价单（草稿）
// @Summary 修改成本调价单
// @Tags 库存
// @Router /api/v1/cost-adjusts/:id [put]
func (h *InventoryExtHandler) UpdateCostAdjust(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req costAdjustReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var bill model.CostAdjust
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可编辑")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("adjust_id = ?", bill.ID).Delete(&model.CostAdjustItem{}).Error; err != nil {
			return err
		}
		var total float64
		var items []model.CostAdjustItem
		for _, it := range req.Items {
			diff := (it.NewPrice - it.OldPrice) * it.Quantity
			total += diff
			items = append(items, model.CostAdjustItem{
				AdjustID: bill.ID, ProductID: it.ProductID, Quantity: it.Quantity,
				OldPrice: it.OldPrice, NewPrice: it.NewPrice, DiffAmount: diff, Remark: it.Remark,
			})
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return tx.Model(&bill).Updates(map[string]interface{}{
			"warehouse_id": req.WarehouseID, "bill_date": req.billDate(), "handler_id": req.HandlerID,
			"ref_bill_no": req.RefBillNo, "remark": req.Remark, "amount": total,
		}).Error
	}); err != nil {
		log.Error().Err(err).Msg("update cost adjust failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

// DeleteCostAdjust 删除成本调价单（草稿）
// @Summary 删除成本调价单
// @Tags 库存
// @Router /api/v1/cost-adjusts/:id [delete]
func (h *InventoryExtHandler) DeleteCostAdjust(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.CostAdjust
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可删除")
		return
	}
	h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("adjust_id = ?", bill.ID).Delete(&model.CostAdjustItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&bill).Error
	})
	response.Ok(c, nil)
}

// CompleteCostAdjust 成本调价单过账：更新库存台账成本与商品成本价
// @Summary 成本调价单过账
// @Tags 库存
// @Router /api/v1/cost-adjusts/:id/complete [put]
func (h *InventoryExtHandler) CompleteCostAdjust(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.CostAdjust
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).Preload("Items").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可过账")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range bill.Items {
			// 更新库存台账成本（若台账存在）
			if err := tx.Model(&model.Stock{}).
				Where("company_id = ? AND warehouse_id = ? AND product_id = ?", companyID, bill.WarehouseID, item.ProductID).
				Updates(map[string]interface{}{
					"cost_price": item.NewPrice,
					"amount":     gorm.Expr("quantity * ?", item.NewPrice),
				}).Error; err != nil {
				return err
			}
			// 同步商品档案参考成本（参考采购价）
			if err := tx.Model(&model.Product{}).
				Where("id = ? AND company_id = ?", item.ProductID, companyID).
				UpdateColumn("purchase_price", item.NewPrice).Error; err != nil {
				return err
			}
		}
		return tx.Model(&bill).UpdateColumn("status", "completed").Error
	}); err != nil {
		log.Error().Err(err).Msg("complete cost adjust failed")
		response.ServerError(c, "过账失败")
		return
	}
	response.OkWithMessage(c, "过账成功", nil)
}

// ==================== 拆装模板 ====================

// ListAssemblyTemplates 拆装模板列表
// @Summary 拆装模板列表
// @Tags 库存
// @Router /api/v1/assembly-templates [get]
func (h *InventoryExtHandler) ListAssemblyTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	q := h.db.Model(&model.AssemblyTemplate{}).Where("company_id = ?", middleware.GetCompanyID(c))
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.AssemblyTemplate
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OkWithPage(c, list, page, pageSize, int(total))
}

// GetAssemblyTemplate 拆装模板详情
// @Summary 拆装模板详情
// @Tags 库存
// @Router /api/v1/assembly-templates/:id [get]
func (h *InventoryExtHandler) GetAssemblyTemplate(c *gin.Context) {
	var tpl model.AssemblyTemplate
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).First(&tpl).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}
	response.Ok(c, tpl)
}

type assemblyTemplateReq struct {
	Name     string `json:"name" binding:"required,max=128"`
	OutItems string `json:"outItems"`
	InItems  string `json:"inItems"`
	Remark   string `json:"remark"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

// CreateAssemblyTemplate 新增拆装模板
// @Summary 新增拆装模板
// @Tags 库存
// @Router /api/v1/assembly-templates [post]
func (h *InventoryExtHandler) CreateAssemblyTemplate(c *gin.Context) {
	var req assemblyTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	tpl := model.AssemblyTemplate{Name: req.Name, OutItems: req.OutItems, InItems: req.InItems, Remark: req.Remark, Status: req.Status}
	tpl.CompanyID = middleware.GetCompanyID(c)
	if err := h.db.Create(&tpl).Error; err != nil {
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, tpl)
}

// UpdateAssemblyTemplate 修改拆装模板
// @Summary 修改拆装模板
// @Tags 库存
// @Router /api/v1/assembly-templates/:id [put]
func (h *InventoryExtHandler) UpdateAssemblyTemplate(c *gin.Context) {
	var req assemblyTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	var tpl model.AssemblyTemplate
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).First(&tpl).Error; err != nil {
		response.NotFound(c, "模板不存在")
		return
	}
	h.db.Model(&tpl).Updates(map[string]interface{}{
		"name": req.Name, "out_items": req.OutItems, "in_items": req.InItems, "remark": req.Remark, "status": req.Status,
	})
	response.Ok(c, tpl)
}

// DeleteAssemblyTemplate 删除拆装模板
// @Summary 删除拆装模板
// @Tags 库存
// @Router /api/v1/assembly-templates/:id [delete]
func (h *InventoryExtHandler) DeleteAssemblyTemplate(c *gin.Context) {
	h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).Delete(&model.AssemblyTemplate{})
	response.Ok(c, nil)
}

// ==================== 组装拆装单 ====================

type assemblyOrderListReq struct {
	Page           int    `form:"page,default=1"`
	PageSize       int    `form:"pageSize,default=30"`
	Keyword        string `form:"keyword"`
	OutWarehouseID uint   `form:"outWarehouseId"`
	InWarehouseID  uint   `form:"inWarehouseId"`
	Status         string `form:"status"`
	ShowCancelled  bool   `form:"showCancelled"`
}

type assemblyOrderRow struct {
	model.AssemblyOrder
	OutWarehouseName string `json:"outWarehouseName"`
	InWarehouseName  string `json:"inWarehouseName"`
	HandlerName      string `json:"handlerName"`
	OperatorName     string `json:"operatorName"`
}

// ListAssemblyOrders 组装拆装单列表
// @Summary 组装拆装单列表
// @Tags 库存
// @Router /api/v1/assembly-orders [get]
func (h *InventoryExtHandler) ListAssemblyOrders(c *gin.Context) {
	var req assemblyOrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	q := h.db.Model(&model.AssemblyOrder{}).
		Select(`assembly_orders.*, COALESCE(ow.name,'') AS out_warehouse_name, COALESCE(iw.name,'') AS in_warehouse_name,
			COALESCE(hd.name,'') AS handler_name, COALESCE(op.name,'') AS operator_name`).
		Joins("LEFT JOIN warehouses ow ON ow.id = assembly_orders.out_warehouse_id").
		Joins("LEFT JOIN warehouses iw ON iw.id = assembly_orders.in_warehouse_id").
		Joins("LEFT JOIN employees hd ON hd.id = assembly_orders.handler_id").
		Joins("LEFT JOIN employees op ON op.id = assembly_orders.operator_id").
		Where("assembly_orders.company_id = ?", middleware.GetCompanyID(c))
	if !req.ShowCancelled {
		q = q.Where("assembly_orders.status != 'cancelled'")
	}
	if req.Keyword != "" {
		q = q.Where("assembly_orders.bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.OutWarehouseID > 0 {
		q = q.Where("assembly_orders.out_warehouse_id = ?", req.OutWarehouseID)
	}
	if req.InWarehouseID > 0 {
		q = q.Where("assembly_orders.in_warehouse_id = ?", req.InWarehouseID)
	}
	if req.Status != "" {
		q = q.Where("assembly_orders.status = ?", req.Status)
	}
	var total int64
	q.Count(&total)
	var list []assemblyOrderRow
	q.Order("assembly_orders.id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Scan(&list)
	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// GetAssemblyOrder 组装拆装单详情
// @Summary 组装拆装单详情
// @Tags 库存
// @Router /api/v1/assembly-orders/:id [get]
func (h *InventoryExtHandler) GetAssemblyOrder(c *gin.Context) {
	var bill model.AssemblyOrder
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), middleware.GetCompanyID(c)).
		Preload("Items.Product").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	response.Ok(c, bill)
}

type assemblyOrderItemReq struct {
	Direction string  `json:"direction" binding:"required,oneof=out in"`
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price"`
	Remark    string  `json:"remark"`
}

type assemblyOrderReq struct {
	TemplateID     uint                   `json:"templateId"`
	OutWarehouseID uint                   `json:"outWarehouseId" binding:"required"`
	InWarehouseID  uint                   `json:"inWarehouseId" binding:"required"`
	BillDate       string                 `json:"billDate"`
	Fee            float64                `json:"fee"`
	HandlerID      uint                   `json:"handlerId"`
	Remark         string                 `json:"remark"`
	Items          []assemblyOrderItemReq `json:"items" binding:"required,min=1"`
}

func (r *assemblyOrderReq) billDate() time.Time {
	if t, err := time.Parse("2006-01-02", r.BillDate); err == nil {
		return t
	}
	return time.Now()
}

func buildAssemblyItems(orderID uint, items []assemblyOrderItemReq) ([]model.AssemblyOrderItem, float64, float64) {
	var out []model.AssemblyOrderItem
	var outAmount, inAmount float64
	for _, it := range items {
		amount := it.Quantity * it.Price
		if it.Direction == "out" {
			outAmount += amount
		} else {
			inAmount += amount
		}
		out = append(out, model.AssemblyOrderItem{
			OrderID: orderID, Direction: it.Direction, ProductID: it.ProductID,
			Quantity: it.Quantity, Price: it.Price, Amount: amount, Remark: it.Remark,
		})
	}
	return out, outAmount, inAmount
}

// CreateAssemblyOrder 新增组装拆装单
// @Summary 新增组装拆装单
// @Tags 库存
// @Router /api/v1/assembly-orders [post]
func (h *InventoryExtHandler) CreateAssemblyOrder(c *gin.Context) {
	var req assemblyOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	var bill model.AssemblyOrder
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		bill = model.AssemblyOrder{
			TemplateID: req.TemplateID, OutWarehouseID: req.OutWarehouseID, InWarehouseID: req.InWarehouseID,
			BillNo: genDailyBillNo(tx, "assembly_orders", "ZZCX"), BillDate: req.billDate(),
			Fee: req.Fee, Status: "draft", HandlerID: req.HandlerID, OperatorID: userID, Remark: req.Remark,
		}
		bill.CompanyID = companyID
		items, outAmount, inAmount := buildAssemblyItems(0, req.Items)
		bill.OutAmount, bill.InAmount = outAmount, inAmount
		bill.Items = items
		return tx.Create(&bill).Error
	}); err != nil {
		log.Error().Err(err).Msg("create assembly order failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, bill)
}

// UpdateAssemblyOrder 修改组装拆装单（草稿）
// @Summary 修改组装拆装单
// @Tags 库存
// @Router /api/v1/assembly-orders/:id [put]
func (h *InventoryExtHandler) UpdateAssemblyOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req assemblyOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var bill model.AssemblyOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可编辑")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", bill.ID).Delete(&model.AssemblyOrderItem{}).Error; err != nil {
			return err
		}
		items, outAmount, inAmount := buildAssemblyItems(bill.ID, req.Items)
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return tx.Model(&bill).Updates(map[string]interface{}{
			"template_id": req.TemplateID, "out_warehouse_id": req.OutWarehouseID, "in_warehouse_id": req.InWarehouseID,
			"bill_date": req.billDate(), "fee": req.Fee, "handler_id": req.HandlerID, "remark": req.Remark,
			"out_amount": outAmount, "in_amount": inAmount,
		}).Error
	}); err != nil {
		log.Error().Err(err).Msg("update assembly order failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

// DeleteAssemblyOrder 删除组装拆装单（草稿）
// @Summary 删除组装拆装单
// @Tags 库存
// @Router /api/v1/assembly-orders/:id [delete]
func (h *InventoryExtHandler) DeleteAssemblyOrder(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.AssemblyOrder
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可删除")
		return
	}
	h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", bill.ID).Delete(&model.AssemblyOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&bill).Error
	})
	response.Ok(c, nil)
}

// CompleteAssemblyOrder 组装拆装单过账：出库仓减出库行、入库仓加入库行
// @Summary 组装拆装单过账
// @Tags 库存
// @Router /api/v1/assembly-orders/:id/complete [put]
func (h *InventoryExtHandler) CompleteAssemblyOrder(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var bill model.AssemblyOrder
	if err := h.db.Where("id = ? AND company_id = ?", c.Param("id"), companyID).Preload("Items").First(&bill).Error; err != nil {
		response.NotFound(c, "单据不存在")
		return
	}
	if bill.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态可过账")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range bill.Items {
			if item.Direction == "out" {
				if err := database.ChangeStock(tx, companyID, bill.OutWarehouseID, item.ProductID, -item.Quantity); err != nil {
					return err
				}
			} else {
				if err := database.ChangeStock(tx, companyID, bill.InWarehouseID, item.ProductID, item.Quantity); err != nil {
					return err
				}
			}
		}
		return tx.Model(&bill).UpdateColumn("status", "completed").Error
	}); err != nil {
		if errors.Is(err, database.ErrStockNotEnough) {
			response.Fail(c, response.CodeBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Msg("complete assembly order failed")
		response.ServerError(c, "过账失败")
		return
	}
	response.OkWithMessage(c, "过账成功", nil)
}
