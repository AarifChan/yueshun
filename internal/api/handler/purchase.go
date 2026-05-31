package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// PurchaseHandler 采购模块处理器
type PurchaseHandler struct {
	db *gorm.DB
}

func NewPurchaseHandler(db *gorm.DB) *PurchaseHandler {
	return &PurchaseHandler{db: db}
}

func (h *PurchaseHandler) RegisterRoutes(r *gin.RouterGroup) {
	po := r.Group("/purchase-orders")
	{
		po.GET("", h.ListOrders)
		po.POST("", h.CreateOrder)
		po.GET("/:id", h.GetOrder)
		po.PUT("/:id", h.UpdateOrder)
		po.DELETE("/:id", h.DeleteOrder)
		po.PUT("/:id/confirm", h.ConfirmOrder)
		po.PUT("/:id/cancel", h.CancelOrder)
	}

	pi := r.Group("/purchase-instock")
	{
		pi.GET("", h.ListInStocks)
		pi.POST("", h.CreateInStock)
		pi.GET("/:id", h.GetInStock)
		pi.PUT("/:id", h.UpdateInStock)
		pi.DELETE("/:id", h.DeleteInStock)
		pi.PUT("/:id/complete", h.CompleteInStock)
	}

	pr := r.Group("/purchase-returns")
	{
		pr.GET("", h.ListReturns)
		pr.POST("", h.CreateReturn)
		pr.GET("/:id", h.GetReturn)
		pr.PUT("/:id", h.UpdateReturn)
		pr.DELETE("/:id", h.DeleteReturn)
		pr.PUT("/:id/complete", h.CompleteReturn)
	}

	pp := r.Group("/purchase-payments")
	{
		pp.GET("", h.ListPayments)
		pp.POST("", h.CreatePayment)
		pp.GET("/:id", h.GetPayment)
		pp.PUT("/:id", h.UpdatePayment)
		pp.DELETE("/:id", h.DeletePayment)
		pp.PUT("/:id/complete", h.CompletePayment)
	}
}

// ==================== 采购订单 ====================

type OrderListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	SupplierID uint   `form:"supplierId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *PurchaseHandler) ListOrders(c *gin.Context) {
	var req OrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.PurchaseOrder{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("order_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.SupplierID > 0 {
		query = query.Where("supplier_id = ?", req.SupplierID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("order_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("order_date <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	var list []model.PurchaseOrder
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *PurchaseHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.PurchaseOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	response.Ok(c, order)
}

type OrderItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	UnitID    uint    `json:"unitId"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gte=0"`
	Discount  float64 `json:"discount"`
	TaxRate   float64 `json:"taxRate"`
	Remark    string  `json:"remark"`
}

type CreateOrderReq struct {
	SupplierID   uint           `json:"supplierId" binding:"required"`
	WarehouseID  uint           `json:"warehouseId" binding:"required"`
	OrderDate    string         `json:"orderDate" binding:"required"`
	DeliveryDate string         `json:"deliveryDate"`
	Discount     float64        `json:"discount"`
	Remark       string         `json:"remark"`
	Items        []OrderItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *PurchaseHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	orderDate, _ := time.Parse("2006-01-02", req.OrderDate)
	deliveryDate, _ := time.Parse("2006-01-02", req.DeliveryDate)
	orderNo := generateOrderNo("PO")

	order := model.PurchaseOrder{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		SupplierID:           req.SupplierID,
		WarehouseID:          req.WarehouseID,
		OrderNo:              orderNo,
		OrderDate:            orderDate,
		DeliveryDate:         deliveryDate,
		Discount:             req.Discount,
		Remark:               req.Remark,
		Status:               "draft",
		OperatorID:           middleware.GetUserID(c),
	}

	var amount, taxAmount, totalAmount float64
	items := make([]model.PurchaseOrderItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		itemTax := itemAmount * item.TaxRate / 100
		itemTotal := itemAmount + itemTax - item.Discount
		amount += itemAmount
		taxAmount += itemTax
		totalAmount += itemTotal

		items[i] = model.PurchaseOrderItem{
			ProductID:   item.ProductID,
			UnitID:      item.UnitID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Amount:      itemAmount,
			Discount:    item.Discount,
			TaxRate:     item.TaxRate,
			TaxAmount:   itemTax,
			TotalAmount: itemTotal,
			Remark:      item.Remark,
		}
	}
	order.Amount = amount
	order.TaxAmount = taxAmount
	order.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create purchase order failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, order)
}

func (h *PurchaseHandler) UpdateOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.PurchaseOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态的订单可编辑")
		return
	}

	orderDate, _ := time.Parse("2006-01-02", req.OrderDate)
	deliveryDate, _ := time.Parse("2006-01-02", req.DeliveryDate)

	order.SupplierID = req.SupplierID
	order.WarehouseID = req.WarehouseID
	order.OrderDate = orderDate
	order.DeliveryDate = deliveryDate
	order.Discount = req.Discount
	order.Remark = req.Remark

	var amount, taxAmount, totalAmount float64
	items := make([]model.PurchaseOrderItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		itemTax := itemAmount * item.TaxRate / 100
		itemTotal := itemAmount + itemTax - item.Discount
		amount += itemAmount
		taxAmount += itemTax
		totalAmount += itemTotal

		items[i] = model.PurchaseOrderItem{
			OrderID:     uint(id),
			ProductID:   item.ProductID,
			UnitID:      item.UnitID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Amount:      itemAmount,
			Discount:    item.Discount,
			TaxRate:     item.TaxRate,
			TaxAmount:   itemTax,
			TotalAmount: itemTotal,
			Remark:      item.Remark,
		}
	}
	order.Amount = amount
	order.TaxAmount = taxAmount
	order.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		if err := tx.Where("order_id = ?", id).Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update purchase order failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, order)
}

func (h *PurchaseHandler) DeleteOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.PurchaseOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态的订单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&order).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete purchase order failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *PurchaseHandler) ConfirmOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.PurchaseOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态的订单可确认")
		return
	}

	order.Status = "confirmed"
	if err := h.db.Save(&order).Error; err != nil {
		log.Error().Err(err).Msg("confirm purchase order failed")
		response.ServerError(c, "确认失败")
		return
	}
	response.OkWithMessage(c, "确认成功", nil)
}

func (h *PurchaseHandler) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.PurchaseOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "draft" && order.Status != "confirmed" {
		response.Fail(c, response.CodeBadRequest, "当前状态不可取消")
		return
	}

	order.Status = "cancelled"
	if err := h.db.Save(&order).Error; err != nil {
		log.Error().Err(err).Msg("cancel purchase order failed")
		response.ServerError(c, "取消失败")
		return
	}
	response.OkWithMessage(c, "取消成功", nil)
}

// ==================== 采购入库 ====================

type InStockListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	SupplierID uint   `form:"supplierId"`
	OrderID    uint   `form:"orderId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *PurchaseHandler) ListInStocks(c *gin.Context) {
	var req InStockListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.PurchaseInStock{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.SupplierID > 0 {
		query = query.Where("supplier_id = ?", req.SupplierID)
	}
	if req.OrderID > 0 {
		query = query.Where("order_id = ?", req.OrderID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("bill_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("bill_date <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	var list []model.PurchaseInStock
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *PurchaseHandler) GetInStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var inStock model.PurchaseInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&inStock).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	response.Ok(c, inStock)
}

type InStockItemReq struct {
	OrderItemID uint    `json:"orderItemId"`
	ProductID   uint    `json:"productId" binding:"required"`
	UnitID      uint    `json:"unitId"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Price       float64 `json:"price" binding:"required,gte=0"`
	BatchNo     string  `json:"batchNo"`
	ExpiryDate  string  `json:"expiryDate"`
	PositionID  uint    `json:"positionId"`
	Remark      string  `json:"remark"`
}

type CreateInStockReq struct {
	SupplierID  uint             `json:"supplierId" binding:"required"`
	WarehouseID uint             `json:"warehouseId" binding:"required"`
	OrderID     uint             `json:"orderId"`
	BillDate    string           `json:"billDate" binding:"required"`
	Discount    float64          `json:"discount"`
	Remark      string           `json:"remark"`
	Items       []InStockItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *PurchaseHandler) CreateInStock(c *gin.Context) {
	var req CreateInStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	billNo := generateOrderNo("PI")

	inStock := model.PurchaseInStock{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		SupplierID:           req.SupplierID,
		WarehouseID:          req.WarehouseID,
		OrderID:              req.OrderID,
		BillNo:               billNo,
		BillDate:             billDate,
		Discount:             req.Discount,
		Remark:               req.Remark,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
	}

	var amount, totalAmount float64
	items := make([]model.PurchaseInStockItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		expiryDate, _ := time.Parse("2006-01-02", item.ExpiryDate)
		items[i] = model.PurchaseInStockItem{
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			UnitID:      item.UnitID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Amount:      itemAmount,
			BatchNo:     item.BatchNo,
			ExpiryDate:  expiryDate,
			PositionID:  item.PositionID,
			Remark:      item.Remark,
		}
	}
	inStock.Amount = amount
	inStock.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&inStock).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].InStockID = inStock.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		if req.OrderID > 0 {
			for _, item := range req.Items {
				if item.OrderItemID > 0 {
					if err := tx.Model(&model.PurchaseOrderItem{}).Where("id = ?", item.OrderItemID).UpdateColumn("received_qty", gorm.Expr("received_qty + ?", item.Quantity)).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create purchase instock failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, inStock)
}

func (h *PurchaseHandler) UpdateInStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateInStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var inStock model.PurchaseInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&inStock).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	if inStock.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的入库单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	inStock.SupplierID = req.SupplierID
	inStock.WarehouseID = req.WarehouseID
	inStock.OrderID = req.OrderID
	inStock.BillDate = billDate
	inStock.Discount = req.Discount
	inStock.Remark = req.Remark

	var amount, totalAmount float64
	items := make([]model.PurchaseInStockItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		expiryDate, _ := time.Parse("2006-01-02", item.ExpiryDate)
		items[i] = model.PurchaseInStockItem{
			InStockID:   uint(id),
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			UnitID:      item.UnitID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Amount:      itemAmount,
			BatchNo:     item.BatchNo,
			ExpiryDate:  expiryDate,
			PositionID:  item.PositionID,
			Remark:      item.Remark,
		}
	}
	inStock.Amount = amount
	inStock.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&inStock).Error; err != nil {
			return err
		}
		if err := tx.Where("in_stock_id = ?", id).Delete(&model.PurchaseInStockItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update purchase instock failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, inStock)
}

func (h *PurchaseHandler) DeleteInStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var inStock model.PurchaseInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&inStock).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	if inStock.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的入库单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("in_stock_id = ?", id).Delete(&model.PurchaseInStockItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&inStock).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete purchase instock failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *PurchaseHandler) CompleteInStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var inStock model.PurchaseInStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&inStock).Error; err != nil {
		response.NotFound(c, "入库单不存在")
		return
	}
	if inStock.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的入库单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		inStock.Status = "completed"
		if err := tx.Save(&inStock).Error; err != nil {
			return err
		}
		for _, item := range inStock.Items {
			if err := tx.Model(&model.Product{}).Where("id = ?", item.ProductID).UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("complete purchase instock failed")
		response.ServerError(c, "完成入库失败")
		return
	}
	response.OkWithMessage(c, "入库完成", nil)
}

// ==================== 采购退货 ====================

type ReturnListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	SupplierID uint   `form:"supplierId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *PurchaseHandler) ListReturns(c *gin.Context) {
	var req ReturnListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.PurchaseReturn{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.SupplierID > 0 {
		query = query.Where("supplier_id = ?", req.SupplierID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("bill_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("bill_date <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	var list []model.PurchaseReturn
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *PurchaseHandler) GetReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var ret model.PurchaseReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&ret).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	response.Ok(c, ret)
}

type ReturnItemReq struct {
	InStockItemID uint    `json:"inStockItemId"`
	ProductID     uint    `json:"productId" binding:"required"`
	UnitID        uint    `json:"unitId"`
	Quantity      float64 `json:"quantity" binding:"required,gt=0"`
	Price         float64 `json:"price" binding:"required,gte=0"`
	Remark        string  `json:"remark"`
}

type CreateReturnReq struct {
	SupplierID  uint            `json:"supplierId" binding:"required"`
	WarehouseID uint            `json:"warehouseId" binding:"required"`
	InStockID   uint            `json:"inStockId"`
	BillDate    string          `json:"billDate" binding:"required"`
	Remark      string          `json:"remark"`
	Items       []ReturnItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *PurchaseHandler) CreateReturn(c *gin.Context) {
	var req CreateReturnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	billNo := generateOrderNo("PR")

	ret := model.PurchaseReturn{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		SupplierID:           req.SupplierID,
		WarehouseID:          req.WarehouseID,
		InStockID:            req.InStockID,
		BillNo:               billNo,
		BillDate:             billDate,
		Remark:               req.Remark,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
	}

	var amount, totalAmount float64
	items := make([]model.PurchaseReturnItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		items[i] = model.PurchaseReturnItem{
			InStockItemID: item.InStockItemID,
			ProductID:     item.ProductID,
			UnitID:        item.UnitID,
			Quantity:      item.Quantity,
			Price:         item.Price,
			Amount:        itemAmount,
			Remark:        item.Remark,
		}
	}
	ret.Amount = amount
	ret.TotalAmount = totalAmount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ret).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ReturnID = ret.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create purchase return failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, ret)
}

func (h *PurchaseHandler) UpdateReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateReturnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var ret model.PurchaseReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&ret).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if ret.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的退货单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	ret.SupplierID = req.SupplierID
	ret.WarehouseID = req.WarehouseID
	ret.InStockID = req.InStockID
	ret.BillDate = billDate
	ret.Remark = req.Remark

	var amount, totalAmount float64
	items := make([]model.PurchaseReturnItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		items[i] = model.PurchaseReturnItem{
			ReturnID:      uint(id),
			InStockItemID: item.InStockItemID,
			ProductID:     item.ProductID,
			UnitID:        item.UnitID,
			Quantity:      item.Quantity,
			Price:         item.Price,
			Amount:        itemAmount,
			Remark:        item.Remark,
		}
	}
	ret.Amount = amount
	ret.TotalAmount = totalAmount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&ret).Error; err != nil {
			return err
		}
		if err := tx.Where("return_id = ?", id).Delete(&model.PurchaseReturnItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update purchase return failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, ret)
}

func (h *PurchaseHandler) DeleteReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var ret model.PurchaseReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&ret).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if ret.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的退货单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("return_id = ?", id).Delete(&model.PurchaseReturnItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&ret).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete purchase return failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *PurchaseHandler) CompleteReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var ret model.PurchaseReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&ret).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if ret.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的退货单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		ret.Status = "completed"
		if err := tx.Save(&ret).Error; err != nil {
			return err
		}
		for _, item := range ret.Items {
			if err := tx.Model(&model.Product{}).Where("id = ?", item.ProductID).UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("complete purchase return failed")
		response.ServerError(c, "完成退货失败")
		return
	}
	response.OkWithMessage(c, "退货完成", nil)
}

// ==================== 采购付款 ====================

type PaymentListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	SupplierID uint   `form:"supplierId"`
	AccountID  uint   `form:"accountId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *PurchaseHandler) ListPayments(c *gin.Context) {
	var req PaymentListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.PurchasePayment{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.SupplierID > 0 {
		query = query.Where("supplier_id = ?", req.SupplierID)
	}
	if req.AccountID > 0 {
		query = query.Where("account_id = ?", req.AccountID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("bill_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("bill_date <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	var list []model.PurchasePayment
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *PurchaseHandler) GetPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var payment model.PurchasePayment
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&payment).Error; err != nil {
		response.NotFound(c, "付款单不存在")
		return
	}
	response.Ok(c, payment)
}

type PaymentItemReq struct {
	OrderID   uint    `json:"orderId"`
	InStockID uint    `json:"inStockId"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Discount  float64 `json:"discount"`
	Remark    string  `json:"remark"`
}

type CreatePaymentReq struct {
	SupplierID uint             `json:"supplierId" binding:"required"`
	AccountID  uint             `json:"accountId" binding:"required"`
	BillDate   string           `json:"billDate" binding:"required"`
	Discount   float64          `json:"discount"`
	Remark     string           `json:"remark"`
	Items      []PaymentItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *PurchaseHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	billNo := generateOrderNo("PP")

	payment := model.PurchasePayment{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		SupplierID:           req.SupplierID,
		AccountID:            req.AccountID,
		BillNo:               billNo,
		BillDate:             billDate,
		Discount:             req.Discount,
		Remark:               req.Remark,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
	}

	var amount, totalAmount float64
	items := make([]model.PurchasePaymentItem, len(req.Items))
	for i, item := range req.Items {
		amount += item.Amount
		totalAmount += item.Amount - item.Discount

		items[i] = model.PurchasePaymentItem{
			OrderID:   item.OrderID,
			InStockID: item.InStockID,
			Amount:    item.Amount,
			Discount:  item.Discount,
			Remark:    item.Remark,
		}
	}
	payment.Amount = amount
	payment.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].PaymentID = payment.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create purchase payment failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, payment)
}

func (h *PurchaseHandler) UpdatePayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreatePaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var payment model.PurchasePayment
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&payment).Error; err != nil {
		response.NotFound(c, "付款单不存在")
		return
	}
	if payment.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的付款单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	payment.SupplierID = req.SupplierID
	payment.AccountID = req.AccountID
	payment.BillDate = billDate
	payment.Discount = req.Discount
	payment.Remark = req.Remark

	var amount, totalAmount float64
	items := make([]model.PurchasePaymentItem, len(req.Items))
	for i, item := range req.Items {
		amount += item.Amount
		totalAmount += item.Amount - item.Discount

		items[i] = model.PurchasePaymentItem{
			PaymentID: uint(id),
			OrderID:   item.OrderID,
			InStockID: item.InStockID,
			Amount:    item.Amount,
			Discount:  item.Discount,
			Remark:    item.Remark,
		}
	}
	payment.Amount = amount
	payment.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&payment).Error; err != nil {
			return err
		}
		if err := tx.Where("payment_id = ?", id).Delete(&model.PurchasePaymentItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update purchase payment failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, payment)
}

func (h *PurchaseHandler) DeletePayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var payment model.PurchasePayment
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&payment).Error; err != nil {
		response.NotFound(c, "付款单不存在")
		return
	}
	if payment.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的付款单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("payment_id = ?", id).Delete(&model.PurchasePaymentItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&payment).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete purchase payment failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *PurchaseHandler) CompletePayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var payment model.PurchasePayment
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&payment).Error; err != nil {
		response.NotFound(c, "付款单不存在")
		return
	}
	if payment.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的付款单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		payment.Status = "completed"
		if err := tx.Save(&payment).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Customer{}).Where("id = ?", payment.SupplierID).UpdateColumn("balance", gorm.Expr("balance - ?", payment.TotalAmount)).Error; err != nil {
			return err
		}
		for _, item := range payment.Items {
			if item.OrderID > 0 {
				if err := tx.Model(&model.PurchaseOrder{}).Where("id = ?", item.OrderID).UpdateColumn("paid_amount", gorm.Expr("paid_amount + ?", item.Amount)).Error; err != nil {
					return err
				}
			}
			if item.InStockID > 0 {
				if err := tx.Model(&model.PurchaseInStock{}).Where("id = ?", item.InStockID).UpdateColumn("paid_amount", gorm.Expr("paid_amount + ?", item.Amount)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("complete purchase payment failed")
		response.ServerError(c, "完成付款失败")
		return
	}
	response.OkWithMessage(c, "付款完成", nil)
}

// generateOrderNo 生成单据编号
func generateOrderNo(prefix string) string {
	return prefix + time.Now().Format("20060102150405") + strconv.Itoa(int(time.Now().UnixNano()%1000))
}
