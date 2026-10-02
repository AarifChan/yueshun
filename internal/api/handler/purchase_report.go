package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/pkg/response"
)

// PurchaseReportHandler 采购报表
type PurchaseReportHandler struct {
	db *gorm.DB
}

func NewPurchaseReportHandler(db *gorm.DB) *PurchaseReportHandler {
	return &PurchaseReportHandler{db: db}
}

func (h *PurchaseReportHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/purchase-reports")
	{
		g.GET("/product-stats", h.ProductStats)
		g.GET("/supplier-stats", h.SupplierStats)
		g.GET("/product-detail", h.ProductDetail)
		g.GET("/order-execution", h.OrderExecution)
	}
}

// ==================== 商品采购统计 ====================

func (h *PurchaseReportHandler) ProductStats(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	supplierKeyword := c.Query("supplierKeyword")
	includeZero := c.Query("includeZero") == "true"

	type stat struct {
		ProductID uint    `json:"productId"`
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		Spec      string  `json:"spec"`
		Unit      string  `json:"unit"`
		Qty       float64 `json:"qty"`
		Amount    float64 `json:"amount"`
		AvgPrice  float64 `json:"avgPrice"`
		GiftQty   float64 `json:"giftQty"`
	}
	stats := map[uint]*stat{}

	var inRows []struct {
		ProductID uint    `gorm:"column:product_id"`
		Code      string  `gorm:"column:code"`
		Name      string  `gorm:"column:name"`
		Spec      string  `gorm:"column:spec"`
		Unit      string  `gorm:"column:unit"`
		Qty       float64 `gorm:"column:qty"`
		Amount    float64 `gorm:"column:amount"`
		Supplier  string  `gorm:"column:supplier"`
	}
	h.db.Raw(`SELECT p.id AS product_id, p.code, p.name, p.specification AS spec, p.unit,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount, COALESCE(s.name,'') AS supplier
		FROM purchase_in_stock_items i
		JOIN purchase_in_stocks b ON b.id = i.in_stock_id
		JOIN products p ON p.id = i.product_id
		LEFT JOIN suppliers s ON s.id = b.supplier_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY p.id, p.code, p.name, p.specification, p.unit, s.name`, companyID, start, end).Scan(&inRows)
	for _, r := range inRows {
		if supplierKeyword != "" && !strings.Contains(r.Supplier, supplierKeyword) {
			continue
		}
		s, ok := stats[r.ProductID]
		if !ok {
			s = &stat{ProductID: r.ProductID, Code: r.Code, Name: r.Name, Spec: r.Spec, Unit: r.Unit}
			stats[r.ProductID] = s
		}
		s.Qty += r.Qty
		s.Amount += r.Amount
	}

	// 退货冲减
	var retRows []struct {
		ProductID uint    `gorm:"column:product_id"`
		Qty       float64 `gorm:"column:qty"`
		Amount    float64 `gorm:"column:amount"`
		Supplier  string  `gorm:"column:supplier"`
	}
	h.db.Raw(`SELECT i.product_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amount, COALESCE(s.name,'') AS supplier
		FROM purchase_return_items i
		JOIN purchase_returns b ON b.id = i.return_id
		LEFT JOIN suppliers s ON s.id = b.supplier_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY i.product_id, s.name`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if supplierKeyword != "" && !strings.Contains(r.Supplier, supplierKeyword) {
			continue
		}
		if s, ok := stats[r.ProductID]; ok {
			s.Qty -= r.Qty
			s.Amount -= r.Amount
		}
	}

	list := make([]stat, 0, len(stats))
	for _, s := range stats {
		if s.Qty != 0 {
			s.AvgPrice = s.Amount / s.Qty
		}
		if !includeZero && s.Qty == 0 {
			continue
		}
		if keyword == "" || strings.Contains(s.Name, keyword) || strings.Contains(s.Code, keyword) || strings.Contains(s.Spec, keyword) {
			list = append(list, *s)
		}
	}
	sortByFloatDesc(list, func(r stat) float64 { return r.Amount })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 供应商采购统计 ====================

func (h *PurchaseReportHandler) SupplierStats(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	productKeyword := c.Query("productKeyword")
	includeZero := c.Query("includeZero") == "true"

	type stat struct {
		SupplierID uint    `json:"supplierId"`
		Name       string  `json:"name"`
		Qty        float64 `json:"qty"`
		Amount     float64 `json:"amount"`
		GiftQty    float64 `json:"giftQty"`
	}
	stats := map[uint]*stat{}

	var inRows []struct {
		SupplierID uint    `gorm:"column:supplier_id"`
		Name       string  `gorm:"column:name"`
		Qty        float64 `gorm:"column:qty"`
		Amount     float64 `gorm:"column:amount"`
		Product    string  `gorm:"column:product"`
	}
	h.db.Raw(`SELECT s.id AS supplier_id, s.name,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount, COALESCE(p.name,'') AS product
		FROM purchase_in_stock_items i
		JOIN purchase_in_stocks b ON b.id = i.in_stock_id
		JOIN suppliers s ON s.id = b.supplier_id
		LEFT JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY s.id, s.name, p.name`, companyID, start, end).Scan(&inRows)
	for _, r := range inRows {
		if productKeyword != "" && !strings.Contains(r.Product, productKeyword) {
			continue
		}
		s, ok := stats[r.SupplierID]
		if !ok {
			s = &stat{SupplierID: r.SupplierID, Name: r.Name}
			stats[r.SupplierID] = s
		}
		s.Qty += r.Qty
		s.Amount += r.Amount
	}

	// 退货冲减
	var retRows []struct {
		SupplierID uint    `gorm:"column:supplier_id"`
		Qty        float64 `gorm:"column:qty"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT b.supplier_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amount
		FROM purchase_return_items i
		JOIN purchase_returns b ON b.id = i.return_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.supplier_id`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if s, ok := stats[r.SupplierID]; ok {
			s.Qty -= r.Qty
			s.Amount -= r.Amount
		}
	}

	list := make([]stat, 0, len(stats))
	for _, s := range stats {
		if !includeZero && s.Qty == 0 {
			continue
		}
		if keyword == "" || strings.Contains(s.Name, keyword) {
			list = append(list, *s)
		}
	}
	sortByFloatDesc(list, func(r stat) float64 { return r.Amount })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 商品采购明细统计 ====================

func (h *PurchaseReportHandler) ProductDetail(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	billType := c.Query("billType")

	type row struct {
		BillNo    string  `json:"billNo"`
		CreatedAt string  `json:"createdAt"`
		BillType  string  `json:"billType"`
		Handler   string  `json:"handler"`
		Product   string  `json:"product"`
		Spec      string  `json:"spec"`
		Warehouse string  `json:"warehouse"`
		Unit      string  `json:"unit"`
		Qty       float64 `json:"qty"`
		Price     float64 `json:"price"`
		Amount    float64 `json:"amount"`
		Remark    string  `json:"remark"`
	}
	rows := []row{}

	type detailRow struct {
		BillNo    string  `gorm:"column:bill_no"`
		CreatedAt string  `gorm:"column:created_at"`
		Handler   string  `gorm:"column:handler"`
		Product   string  `gorm:"column:product"`
		Spec      string  `gorm:"column:spec"`
		Warehouse string  `gorm:"column:warehouse"`
		Unit      string  `gorm:"column:unit"`
		Qty       float64 `gorm:"column:qty"`
		Price     float64 `gorm:"column:price"`
		Amount    float64 `gorm:"column:amount"`
		Remark    string  `gorm:"column:remark"`
	}

	if billType == "" || billType == "采购入库单" {
		var inRows []detailRow
		h.db.Raw(`SELECT b.bill_no, b.created_at::text, COALESCE(e.name,'') AS handler,
			p.name AS product, p.specification AS spec, COALESCE(w.name,'') AS warehouse, p.unit,
			i.quantity AS qty, i.price, i.amount, COALESCE(i.remark,'') AS remark
			FROM purchase_in_stock_items i
			JOIN purchase_in_stocks b ON b.id = i.in_stock_id
			LEFT JOIN employees e ON e.id = b.operator_id
			JOIN products p ON p.id = i.product_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
			AND b.bill_date >= ? AND b.bill_date <= ?`, companyID, start, end).Scan(&inRows)
		for _, r := range inRows {
			if keyword != "" && !strings.Contains(r.Product, keyword) {
				continue
			}
			rows = append(rows, row{
				BillNo: r.BillNo, CreatedAt: r.CreatedAt, BillType: "采购入库单", Handler: r.Handler,
				Product: r.Product, Spec: r.Spec, Warehouse: r.Warehouse, Unit: r.Unit,
				Qty: r.Qty, Price: r.Price, Amount: r.Amount, Remark: r.Remark,
			})
		}
	}

	if billType == "" || billType == "采购退货单" {
		var retRows []detailRow
		h.db.Raw(`SELECT b.bill_no, b.created_at::text, COALESCE(e.name,'') AS handler,
			p.name AS product, p.specification AS spec, COALESCE(w.name,'') AS warehouse, p.unit,
			i.quantity AS qty, i.price, i.amount, COALESCE(i.remark,'') AS remark
			FROM purchase_return_items i
			JOIN purchase_returns b ON b.id = i.return_id
			LEFT JOIN employees e ON e.id = b.operator_id
			JOIN products p ON p.id = i.product_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
			AND b.bill_date >= ? AND b.bill_date <= ?`, companyID, start, end).Scan(&retRows)
		for _, r := range retRows {
			if keyword != "" && !strings.Contains(r.Product, keyword) {
				continue
			}
			rows = append(rows, row{
				BillNo: r.BillNo, CreatedAt: r.CreatedAt, BillType: "采购退货单", Handler: r.Handler,
				Product: r.Product, Spec: r.Spec, Warehouse: r.Warehouse, Unit: r.Unit,
				Qty: -r.Qty, Price: r.Price, Amount: -r.Amount, Remark: r.Remark,
			})
		}
	}

	sortByStringDesc(rows, func(r row) string { return r.CreatedAt })
	response.Ok(c, paginateSaleRows(rows, page, pageSize))
}

// ==================== 采购订单执行明细 ====================

func (h *PurchaseReportHandler) OrderExecution(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	supplierKeyword := c.Query("supplierKeyword")
	status := c.Query("status")

	type row struct {
		OrderNo    string  `json:"orderNo"`
		CreatedAt  string  `json:"createdAt"`
		Supplier   string  `json:"supplier"`
		Status     string  `json:"status"`
		StatusLabel string `json:"statusLabel"`
		Product    string  `json:"product"`
		Spec       string  `json:"spec"`
		Warehouse  string  `json:"warehouse"`
		Unit       string  `json:"unit"`
		OrderQty   float64 `json:"orderQty"`
		Price      float64 `json:"price"`
		Amount     float64 `json:"amount"`
		InQty      float64 `json:"inQty"`
		PendingQty float64 `json:"pendingQty"`
	}

	var raw []struct {
		OrderNo   string  `gorm:"column:order_no"`
		CreatedAt string  `gorm:"column:created_at"`
		Supplier  string  `gorm:"column:supplier"`
		Status    string  `gorm:"column:status"`
		Product   string  `gorm:"column:product"`
		Spec      string  `gorm:"column:spec"`
		Warehouse string  `gorm:"column:warehouse"`
		Unit      string  `gorm:"column:unit"`
		Qty       float64 `gorm:"column:qty"`
		Price     float64 `gorm:"column:price"`
		Amount    float64 `gorm:"column:amount"`
		Received  float64 `gorm:"column:received"`
	}
	h.db.Raw(`SELECT b.order_no, b.created_at::text, COALESCE(s.name,'') AS supplier, b.status,
		p.name AS product, p.specification AS spec, COALESCE(w.name,'') AS warehouse, p.unit,
		i.quantity AS qty, i.price, i.amount, i.received_qty AS received
		FROM purchase_order_items i
		JOIN purchase_orders b ON b.id = i.order_id
		LEFT JOIN suppliers s ON s.id = b.supplier_id
		JOIN products p ON p.id = i.product_id
		LEFT JOIN warehouses w ON w.id = b.warehouse_id
		WHERE b.company_id = ? AND b.deleted_at IS NULL AND b.status NOT IN ('draft','cancelled')
		AND b.order_date >= ? AND b.order_date <= ?`, companyID, start, end).Scan(&raw)

	rows := make([]row, 0, len(raw))
	for _, r := range raw {
		if keyword != "" && !strings.Contains(r.Product, keyword) && !strings.Contains(r.OrderNo, keyword) {
			continue
		}
		if supplierKeyword != "" && !strings.Contains(r.Supplier, supplierKeyword) {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		pending := r.Qty - r.Received
		if pending < 0 {
			pending = 0
		}
		rows = append(rows, row{
			OrderNo: r.OrderNo, CreatedAt: r.CreatedAt, Supplier: r.Supplier,
			Status: r.Status, StatusLabel: statusLabelOf(r.Status),
			Product: r.Product, Spec: r.Spec, Warehouse: r.Warehouse, Unit: r.Unit,
			OrderQty: r.Qty, Price: r.Price, Amount: r.Amount, InQty: r.Received, PendingQty: pending,
		})
	}
	sortByStringDesc(rows, func(r row) string { return r.CreatedAt })
	response.Ok(c, paginateSaleRows(rows, page, pageSize))
}
