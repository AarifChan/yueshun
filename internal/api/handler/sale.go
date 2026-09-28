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

// SaleHandler 销售模块处理器
type SaleHandler struct {
	db *gorm.DB
}

func NewSaleHandler(db *gorm.DB) *SaleHandler {
	return &SaleHandler{db: db}
}

func (h *SaleHandler) RegisterRoutes(r *gin.RouterGroup) {
	so := r.Group("/sales-orders")
	{
		so.GET("", h.ListSaleOrders)
		so.POST("", h.CreateSaleOrder)
		so.GET("/history-price", h.HistoryPrice)
		so.GET("/:id", h.GetSaleOrder)
		so.PUT("/:id", h.UpdateSaleOrder)
		so.DELETE("/:id", h.DeleteSaleOrder)
		so.PUT("/:id/confirm", h.ConfirmSaleOrder)
		so.PUT("/:id/cancel", h.CancelSaleOrder)
	}

	so2 := r.Group("/sales-outstock")
	{
		so2.GET("", h.ListOutStocks)
		so2.POST("", h.CreateOutStock)
		so2.GET("/:id", h.GetOutStock)
		so2.PUT("/:id", h.UpdateOutStock)
		so2.DELETE("/:id", h.DeleteOutStock)
		so2.PUT("/:id/complete", h.CompleteOutStock)
	}

	sr := r.Group("/sales-returns")
	{
		sr.GET("", h.ListSaleReturns)
		sr.POST("", h.CreateSaleReturn)
		sr.GET("/:id", h.GetSaleReturn)
		sr.PUT("/:id", h.UpdateSaleReturn)
		sr.DELETE("/:id", h.DeleteSaleReturn)
		sr.PUT("/:id/complete", h.CompleteSaleReturn)
	}

	sr2 := r.Group("/sales-receipts")
	{
		sr2.GET("", h.ListReceipts)
		sr2.POST("", h.CreateReceipt)
		sr2.GET("/:id", h.GetReceipt)
		sr2.PUT("/:id", h.UpdateReceipt)
		sr2.DELETE("/:id", h.DeleteReceipt)
		sr2.PUT("/:id/complete", h.CompleteReceipt)
	}
}

// ==================== 销售订单 ====================

type SaleOrderListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CustomerID uint   `form:"customerId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *SaleHandler) ListSaleOrders(c *gin.Context) {
	var req SaleOrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.SalesOrder{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("order_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
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

	var list []model.SalesOrder
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// HistoryPrice 历史售价：查询客户最近一次购买某商品的价格
func (h *SaleHandler) HistoryPrice(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Query("customerId"), 10, 64)
	if err != nil || customerID == 0 {
		response.BadRequest(c, "customerId 必填")
		return
	}
	productID, err := strconv.ParseUint(c.Query("productId"), 10, 64)
	if err != nil || productID == 0 {
		response.BadRequest(c, "productId 必填")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var result struct {
		Price     float64   `json:"price"`
		OrderID   uint      `json:"orderId"`
		OrderNo   string    `json:"orderNo"`
		OrderDate time.Time `json:"date"`
	}
	query := h.db.Table("sales_order_items AS i").
		Select("i.price, o.id AS order_id, o.order_no, o.order_date").
		Joins("JOIN sales_orders o ON o.id = i.order_id").
		Where("o.company_id = ? AND o.customer_id = ? AND o.status <> 'cancelled' AND i.product_id = ?", companyID, customerID, productID).
		Order("o.created_at DESC").
		Limit(1).
		Scan(&result)
	if query.Error != nil {
		log.Error().Err(query.Error).Msg("query history price failed")
		response.ServerError(c, "查询失败")
		return
	}
	if query.RowsAffected == 0 {
		response.Ok(c, nil)
		return
	}

	response.Ok(c, result)
}

func (h *SaleHandler) GetSaleOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.SalesOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	response.Ok(c, order)
}

type SaleOrderItemReq struct {
	ProductID uint    `json:"productId" binding:"required"`
	UnitID    uint    `json:"unitId"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gte=0"`
	Discount  float64 `json:"discount"`
	TaxRate   float64 `json:"taxRate"`
	Remark    string  `json:"remark"`
}

type CreateSaleOrderReq struct {
	CustomerID   uint              `json:"customerId" binding:"required"`
	WarehouseID  uint              `json:"warehouseId" binding:"required"`
	OrderDate    string            `json:"orderDate" binding:"required"`
	DeliveryDate string            `json:"deliveryDate"`
	Discount     float64           `json:"discount"`
	Remark       string            `json:"remark"`
	Items        []SaleOrderItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *SaleHandler) CreateSaleOrder(c *gin.Context) {
	var req CreateSaleOrderReq
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
	orderNo := generateOrderNo("SO")

	order := model.SalesOrder{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
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
	items := make([]model.SalesOrderItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		itemTax := itemAmount * item.TaxRate / 100
		itemTotal := itemAmount + itemTax - item.Discount
		amount += itemAmount
		taxAmount += itemTax
		totalAmount += itemTotal

		items[i] = model.SalesOrderItem{
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
		log.Error().Err(err).Msg("create sale order failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, order)
}

func (h *SaleHandler) UpdateSaleOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateSaleOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.SalesOrder
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

	order.CustomerID = req.CustomerID
	order.WarehouseID = req.WarehouseID
	order.OrderDate = orderDate
	order.DeliveryDate = deliveryDate
	order.Discount = req.Discount
	order.Remark = req.Remark

	var amount, taxAmount, totalAmount float64
	items := make([]model.SalesOrderItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		itemTax := itemAmount * item.TaxRate / 100
		itemTotal := itemAmount + itemTax - item.Discount
		amount += itemAmount
		taxAmount += itemTax
		totalAmount += itemTotal

		items[i] = model.SalesOrderItem{
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
		if err := tx.Where("order_id = ?", id).Delete(&model.SalesOrderItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update sale order failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, order)
}

func (h *SaleHandler) DeleteSaleOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.SalesOrder
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&order).Error; err != nil {
		response.NotFound(c, "订单不存在")
		return
	}
	if order.Status != "draft" {
		response.Fail(c, response.CodeBadRequest, "只有草稿状态的订单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", id).Delete(&model.SalesOrderItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&order).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete sale order failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *SaleHandler) ConfirmSaleOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.SalesOrder
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
		log.Error().Err(err).Msg("confirm sale order failed")
		response.ServerError(c, "确认失败")
		return
	}
	// 回写客户最近下单时间
	now := time.Now()
	if err := h.db.Model(&model.Customer{}).Where("id = ? AND company_id = ?", order.CustomerID, companyID).UpdateColumn("last_order_at", now).Error; err != nil {
		log.Error().Err(err).Msg("update customer last order at failed")
	}
	response.OkWithMessage(c, "确认成功", nil)
}

func (h *SaleHandler) CancelSaleOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var order model.SalesOrder
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
		log.Error().Err(err).Msg("cancel sale order failed")
		response.ServerError(c, "取消失败")
		return
	}
	response.OkWithMessage(c, "取消成功", nil)
}

// ==================== 销售出库 ====================

type OutStockListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CustomerID uint   `form:"customerId"`
	OrderID    uint   `form:"orderId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *SaleHandler) ListOutStocks(c *gin.Context) {
	var req OutStockListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.SalesOutStock{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
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

	var list []model.SalesOutStock
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *SaleHandler) GetOutStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var outStock model.SalesOutStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&outStock).Error; err != nil {
		response.NotFound(c, "出库单不存在")
		return
	}
	response.Ok(c, outStock)
}

type OutStockItemReq struct {
	OrderItemID uint    `json:"orderItemId"`
	ProductID   uint    `json:"productId" binding:"required"`
	UnitID      uint    `json:"unitId"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Price       float64 `json:"price" binding:"required,gte=0"`
	BatchNo     string  `json:"batchNo"`
	PositionID  uint    `json:"positionId"`
	Remark      string  `json:"remark"`
}

type CreateOutStockReq struct {
	CustomerID  uint              `json:"customerId" binding:"required"`
	WarehouseID uint             `json:"warehouseId" binding:"required"`
	OrderID     uint              `json:"orderId"`
	BillDate    string            `json:"billDate" binding:"required"`
	Discount    float64           `json:"discount"`
	Remark      string            `json:"remark"`
	Items       []OutStockItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *SaleHandler) CreateOutStock(c *gin.Context) {
	var req CreateOutStockReq
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
	billNo := generateOrderNo("SO")

	outStock := model.SalesOutStock{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
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
	items := make([]model.SalesOutStockItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		items[i] = model.SalesOutStockItem{
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			UnitID:      item.UnitID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Amount:      itemAmount,
			BatchNo:     item.BatchNo,
			PositionID:  item.PositionID,
			Remark:      item.Remark,
		}
	}
	outStock.Amount = amount
	outStock.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&outStock).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OutStockID = outStock.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		if req.OrderID > 0 {
			for _, item := range req.Items {
				if item.OrderItemID > 0 {
					if err := tx.Model(&model.SalesOrderItem{}).Where("id = ?", item.OrderItemID).UpdateColumn("delivered_qty", gorm.Expr("delivered_qty + ?", item.Quantity)).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create sale outstock failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, outStock)
}

func (h *SaleHandler) UpdateOutStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateOutStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var outStock model.SalesOutStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&outStock).Error; err != nil {
		response.NotFound(c, "出库单不存在")
		return
	}
	if outStock.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的出库单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	outStock.CustomerID = req.CustomerID
	outStock.WarehouseID = req.WarehouseID
	outStock.OrderID = req.OrderID
	outStock.BillDate = billDate
	outStock.Discount = req.Discount
	outStock.Remark = req.Remark

	var amount, totalAmount float64
	items := make([]model.SalesOutStockItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		items[i] = model.SalesOutStockItem{
			OutStockID:  uint(id),
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			UnitID:      item.UnitID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Amount:      itemAmount,
			BatchNo:     item.BatchNo,
			PositionID:  item.PositionID,
			Remark:      item.Remark,
		}
	}
	outStock.Amount = amount
	outStock.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&outStock).Error; err != nil {
			return err
		}
		if err := tx.Where("out_stock_id = ?", id).Delete(&model.SalesOutStockItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update sale outstock failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, outStock)
}

func (h *SaleHandler) DeleteOutStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var outStock model.SalesOutStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&outStock).Error; err != nil {
		response.NotFound(c, "出库单不存在")
		return
	}
	if outStock.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的出库单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("out_stock_id = ?", id).Delete(&model.SalesOutStockItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&outStock).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete sale outstock failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *SaleHandler) CompleteOutStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var outStock model.SalesOutStock
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&outStock).Error; err != nil {
		response.NotFound(c, "出库单不存在")
		return
	}
	if outStock.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的出库单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		outStock.Status = "completed"
		if err := tx.Save(&outStock).Error; err != nil {
			return err
		}
		for _, item := range outStock.Items {
			if err := database.ChangeStock(tx, companyID, outStock.WarehouseID, item.ProductID, -item.Quantity); err != nil {
				return err
			}
		}
		// 出库挂账：客户应收（欠款）增加
		if err := tx.Model(&model.Customer{}).Where("id = ?", outStock.CustomerID).
			UpdateColumn("balance", gorm.Expr("balance + ?", outStock.TotalAmount)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, database.ErrStockNotEnough) {
			response.Fail(c, response.CodeBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Msg("complete sale outstock failed")
		response.ServerError(c, "完成出库失败")
		return
	}
	response.OkWithMessage(c, "出库完成", nil)
}

// ==================== 销售退货 ====================

type SaleReturnListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CustomerID uint   `form:"customerId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *SaleHandler) ListSaleReturns(c *gin.Context) {
	var req SaleReturnListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.SalesReturn{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
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

	var list []model.SalesReturn
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *SaleHandler) GetSaleReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var saleRet model.SalesReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items.Product").First(&saleRet).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	response.Ok(c, saleRet)
}

type SaleReturnItemReq struct {
	OutStockItemID uint    `json:"outStockItemId"`
	ProductID      uint    `json:"productId" binding:"required"`
	UnitID         uint    `json:"unitId"`
	Quantity       float64 `json:"quantity" binding:"required,gt=0"`
	Price          float64 `json:"price" binding:"required,gte=0"`
	Remark         string  `json:"remark"`
}

type CreateSaleReturnReq struct {
	CustomerID  uint               `json:"customerId" binding:"required"`
	WarehouseID uint              `json:"warehouseId" binding:"required"`
	OutStockID  uint              `json:"outStockId"`
	BillDate    string            `json:"billDate" binding:"required"`
	Remark      string            `json:"remark"`
	Items       []SaleReturnItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *SaleHandler) CreateSaleReturn(c *gin.Context) {
	var req CreateSaleReturnReq
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
	billNo := generateOrderNo("SR")

	saleRet := model.SalesReturn{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
		WarehouseID:          req.WarehouseID,
		OutStockID:           req.OutStockID,
		BillNo:               billNo,
		BillDate:             billDate,
		Remark:               req.Remark,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
	}

	var amount, totalAmount float64
	items := make([]model.SalesReturnItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		items[i] = model.SalesReturnItem{
			OutStockItemID: item.OutStockItemID,
			ProductID:      item.ProductID,
			UnitID:         item.UnitID,
			Quantity:       item.Quantity,
			Price:          item.Price,
			Amount:         itemAmount,
			Remark:         item.Remark,
		}
	}
	saleRet.Amount = amount
	saleRet.TotalAmount = totalAmount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&saleRet).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ReturnID = saleRet.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create sale return failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, saleRet)
}

func (h *SaleHandler) UpdateSaleReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateSaleReturnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var saleRet model.SalesReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&saleRet).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if saleRet.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的退货单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	saleRet.CustomerID = req.CustomerID
	saleRet.WarehouseID = req.WarehouseID
	saleRet.OutStockID = req.OutStockID
	saleRet.BillDate = billDate
	saleRet.Remark = req.Remark

	var amount, totalAmount float64
	items := make([]model.SalesReturnItem, len(req.Items))
	for i, item := range req.Items {
		itemAmount := item.Quantity * item.Price
		amount += itemAmount
		totalAmount += itemAmount

		items[i] = model.SalesReturnItem{
			ReturnID:       uint(id),
			OutStockItemID: item.OutStockItemID,
			ProductID:      item.ProductID,
			UnitID:         item.UnitID,
			Quantity:       item.Quantity,
			Price:          item.Price,
			Amount:         itemAmount,
			Remark:         item.Remark,
		}
	}
	saleRet.Amount = amount
	saleRet.TotalAmount = totalAmount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&saleRet).Error; err != nil {
			return err
		}
		if err := tx.Where("return_id = ?", id).Delete(&model.SalesReturnItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update sale return failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, saleRet)
}

func (h *SaleHandler) DeleteSaleReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var saleRet model.SalesReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&saleRet).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if saleRet.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的退货单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("return_id = ?", id).Delete(&model.SalesReturnItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&saleRet).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete sale return failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *SaleHandler) CompleteSaleReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var saleRet model.SalesReturn
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&saleRet).Error; err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if saleRet.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的退货单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		saleRet.Status = "completed"
		if err := tx.Save(&saleRet).Error; err != nil {
			return err
		}
		for _, item := range saleRet.Items {
			if err := database.ChangeStock(tx, companyID, saleRet.WarehouseID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		// 退货冲减客户应收（欠款）
		if err := tx.Model(&model.Customer{}).Where("id = ?", saleRet.CustomerID).
			UpdateColumn("balance", gorm.Expr("balance - ?", saleRet.TotalAmount)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("complete sale return failed")
		response.ServerError(c, "完成退货失败")
		return
	}
	response.OkWithMessage(c, "退货完成", nil)
}

// ==================== 销售收款 ====================

type ReceiptListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CustomerID uint   `form:"customerId"`
	AccountID  uint   `form:"accountId"`
	Status     string `form:"status"`
	StartDate  string `form:"startDate"`
	EndDate    string `form:"endDate"`
}

func (h *SaleHandler) ListReceipts(c *gin.Context) {
	var req ReceiptListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.SalesReceipt{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("bill_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.CustomerID > 0 {
		query = query.Where("customer_id = ?", req.CustomerID)
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

	var list []model.SalesReceipt
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *SaleHandler) GetReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var receipt model.SalesReceipt
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&receipt).Error; err != nil {
		response.NotFound(c, "收款单不存在")
		return
	}
	response.Ok(c, receipt)
}

type ReceiptItemReq struct {
	OrderID    uint    `json:"orderId"`
	OutStockID uint    `json:"outStockId"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Discount   float64 `json:"discount"`
	Remark     string  `json:"remark"`
}

type CreateReceiptReq struct {
	CustomerID uint              `json:"customerId" binding:"required"`
	AccountID  uint              `json:"accountId" binding:"required"`
	BillDate   string            `json:"billDate" binding:"required"`
	Discount   float64           `json:"discount"`
	Remark     string            `json:"remark"`
	Items      []ReceiptItemReq  `json:"items" binding:"required,min=1,dive"`
}

func (h *SaleHandler) CreateReceipt(c *gin.Context) {
	var req CreateReceiptReq
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
	billNo := generateOrderNo("SR")

	receipt := model.SalesReceipt{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
		AccountID:            req.AccountID,
		BillNo:               billNo,
		BillDate:             billDate,
		Discount:             req.Discount,
		Remark:               req.Remark,
		Status:               "pending",
		OperatorID:           middleware.GetUserID(c),
	}

	var amount, totalAmount float64
	items := make([]model.SalesReceiptItem, len(req.Items))
	for i, item := range req.Items {
		amount += item.Amount
		totalAmount += item.Amount - item.Discount

		items[i] = model.SalesReceiptItem{
			OrderID:    item.OrderID,
			OutStockID: item.OutStockID,
			Amount:     item.Amount,
			Discount:   item.Discount,
			Remark:     item.Remark,
		}
	}
	receipt.Amount = amount
	receipt.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&receipt).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ReceiptID = receipt.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create sale receipt failed")
		response.ServerError(c, "创建失败")
		return
	}

	response.Ok(c, receipt)
}

func (h *SaleHandler) UpdateReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req CreateReceiptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var receipt model.SalesReceipt
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&receipt).Error; err != nil {
		response.NotFound(c, "收款单不存在")
		return
	}
	if receipt.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的收款单可编辑")
		return
	}

	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	receipt.CustomerID = req.CustomerID
	receipt.AccountID = req.AccountID
	receipt.BillDate = billDate
	receipt.Discount = req.Discount
	receipt.Remark = req.Remark

	var amount, totalAmount float64
	items := make([]model.SalesReceiptItem, len(req.Items))
	for i, item := range req.Items {
		amount += item.Amount
		totalAmount += item.Amount - item.Discount

		items[i] = model.SalesReceiptItem{
			ReceiptID:  uint(id),
			OrderID:    item.OrderID,
			OutStockID: item.OutStockID,
			Amount:     item.Amount,
			Discount:   item.Discount,
			Remark:     item.Remark,
		}
	}
	receipt.Amount = amount
	receipt.TotalAmount = totalAmount - req.Discount

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&receipt).Error; err != nil {
			return err
		}
		if err := tx.Where("receipt_id = ?", id).Delete(&model.SalesReceiptItem{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("update sale receipt failed")
		response.ServerError(c, "更新失败")
		return
	}

	response.Ok(c, receipt)
}

func (h *SaleHandler) DeleteReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var receipt model.SalesReceipt
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&receipt).Error; err != nil {
		response.NotFound(c, "收款单不存在")
		return
	}
	if receipt.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的收款单可删除")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("receipt_id = ?", id).Delete(&model.SalesReceiptItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&receipt).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("delete sale receipt failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *SaleHandler) CompleteReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var receipt model.SalesReceipt
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Preload("Items").First(&receipt).Error; err != nil {
		response.NotFound(c, "收款单不存在")
		return
	}
	if receipt.Status != "pending" {
		response.Fail(c, response.CodeBadRequest, "只有待审核状态的收款单可完成")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		receipt.Status = "completed"
		if err := tx.Save(&receipt).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Customer{}).Where("id = ?", receipt.CustomerID).UpdateColumn("balance", gorm.Expr("balance - ?", receipt.TotalAmount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Account{}).Where("id = ? AND company_id = ?", receipt.AccountID, companyID).UpdateColumn("balance", gorm.Expr("balance + ?", receipt.TotalAmount)).Error; err != nil {
			return err
		}
		flow := model.AccountFlow{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			AccountID:            receipt.AccountID,
			Type:                 "income",
			Amount:               receipt.TotalAmount,
			RefType:              "sales_receipt",
			RefID:                receipt.ID,
			Remark:               receipt.BillNo,
		}
		if err := tx.Create(&flow).Error; err != nil {
			return err
		}
		for _, item := range receipt.Items {
			if item.OrderID > 0 {
				if err := tx.Model(&model.SalesOrder{}).Where("id = ?", item.OrderID).UpdateColumn("paid_amount", gorm.Expr("paid_amount + ?", item.Amount)).Error; err != nil {
					return err
				}
			}
			if item.OutStockID > 0 {
				if err := tx.Model(&model.SalesOutStock{}).Where("id = ?", item.OutStockID).UpdateColumn("paid_amount", gorm.Expr("paid_amount + ?", item.Amount)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("complete sale receipt failed")
		response.ServerError(c, "完成收款失败")
		return
	}
	response.OkWithMessage(c, "收款完成", nil)
}
