package handler

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// SaleReportHandler 销售报表 + 销售提成
type SaleReportHandler struct {
	db *gorm.DB
}

func NewSaleReportHandler(db *gorm.DB) *SaleReportHandler {
	return &SaleReportHandler{db: db}
}

func (h *SaleReportHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/sale-reports")
	{
		g.GET("/daily-summary", h.DailySummary)
		g.GET("/daily-close", h.DailyClose)
		g.GET("/product-stats", h.ProductStats)
		g.GET("/customer-stats", h.CustomerStats)
		g.GET("/employee-stats", h.EmployeeStats)
		g.GET("/product-detail", h.ProductDetail)
		g.GET("/order-execution", h.OrderExecution)
		g.GET("/comprehensive", h.Comprehensive)
		g.GET("/commission-stats", h.CommissionStats)
	}
	p := r.Group("/commission-plans")
	{
		p.GET("", h.PlanList)
		p.GET("/all", h.PlanAll)
		p.POST("", h.PlanCreate)
		p.PUT("/:id", h.PlanUpdate)
		p.DELETE("/:id", h.PlanDelete)
	}
}

func salePageParams(c *gin.Context) (page, pageSize int, start, end string) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 10000 {
		pageSize = 30
	}
	start = c.Query("startDate")
	end = c.Query("endDate")
	if start == "" {
		start = "2000-01-01"
	}
	if end == "" {
		end = "2099-12-31"
	}
	return
}

// ==================== 日销售统计 ====================

type dailySummaryRow struct {
	Day          string  `json:"day"`
	SaleAmount   float64 `json:"saleAmount"`
	Freight      float64 `json:"freight"`
	ReturnAmount float64 `json:"returnAmount"`
	TotalAmount  float64 `json:"totalAmount"`
	Receipt      float64 `json:"receipt"`
	Debt         float64 `json:"debt"`
	Discount     float64 `json:"discount"`
	Profit       float64 `json:"profit"`
}

func (h *SaleReportHandler) DailySummary(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)

	type agg struct {
		Day    string  `gorm:"column:day"`
		Amount float64 `gorm:"column:amount"`
	}
	days := map[string]*dailySummaryRow{}
	get := func(day string) *dailySummaryRow {
		if r, ok := days[day]; ok {
			return r
		}
		r := &dailySummaryRow{Day: day}
		days[day] = r
		return r
	}

	// 销售出库（金额 + 优惠）
	var saleRows []struct {
		Day      string  `gorm:"column:day"`
		Amount   float64 `gorm:"column:amount"`
		Discount float64 `gorm:"column:discount"`
	}
	h.db.Raw(`SELECT b.bill_date::text AS day, SUM(b.total_amount) AS amount, SUM(b.discount) AS discount
		FROM sales_out_stocks b
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.bill_date`, companyID, start, end).Scan(&saleRows)
	for _, r := range saleRows {
		row := get(r.Day)
		row.SaleAmount = r.Amount
		row.Discount = r.Discount
	}

	// 销售退货
	var retRows []agg
	h.db.Raw(`SELECT b.bill_date::text AS day, SUM(b.total_amount) AS amount
		FROM sales_returns b
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.bill_date`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		get(r.Day).ReturnAmount = r.Amount
	}

	// 收款
	var rcptRows []agg
	h.db.Raw(`SELECT b.bill_date::text AS day, SUM(b.total_amount) AS amount
		FROM sales_receipts b
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.bill_date`, companyID, start, end).Scan(&rcptRows)
	for _, r := range rcptRows {
		get(r.Day).Receipt = r.Amount
	}

	// 利润（出库毛利 - 退货毛利，成本取商品进价）
	var profitRows []agg
	h.db.Raw(`SELECT b.bill_date::text AS day, SUM(i.amount - i.quantity * p.purchase_price) AS amount
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.bill_date`, companyID, start, end).Scan(&profitRows)
	for _, r := range profitRows {
		get(r.Day).Profit = r.Amount
	}
	var retProfitRows []agg
	h.db.Raw(`SELECT b.bill_date::text AS day, SUM(i.amount - i.quantity * p.purchase_price) AS amount
		FROM sales_return_items i
		JOIN sales_returns b ON b.id = i.return_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.bill_date`, companyID, start, end).Scan(&retProfitRows)
	for _, r := range retProfitRows {
		get(r.Day).Profit -= r.Amount
	}

	list := make([]dailySummaryRow, 0, len(days))
	for _, r := range days {
		r.TotalAmount = r.SaleAmount - r.ReturnAmount
		r.Debt = r.TotalAmount - r.Receipt
		list = append(list, *r)
	}
	sortByStringDesc(list, func(r dailySummaryRow) string { return r.Day })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 日清日结 ====================

func (h *SaleReportHandler) DailyClose(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	keyword := c.Query("keyword")

	// 企业账户列表（动态列）
	var accounts []model.Account
	h.db.Where("company_id = ? AND status = 1", companyID).Order("id").Find(&accounts)

	type closeRow struct {
		BillType    string             `json:"billType"`
		BillNo      string             `json:"billNo"`
		Status      string             `json:"status"`
		StatusLabel string             `json:"statusLabel"`
		Manager     string             `json:"manager"`
		HandlerName string             `json:"handlerName"`
		PayStatus   string             `json:"payStatus"`
		Amount      float64            `json:"amount"`
		Balance     float64            `json:"balance"`
		Accounts    map[string]float64 `json:"accounts"`
		CreatedAt   string             `json:"createdAt"`
	}
	rows := []closeRow{}
	match := func(billNo string) bool {
		return keyword == "" || strings.Contains(billNo, keyword)
	}

	// 销售出库单
	var outRows []struct {
		BillNo     string  `gorm:"column:bill_no"`
		Status     string  `gorm:"column:status"`
		Handler    string  `gorm:"column:handler"`
		Amount     float64 `gorm:"column:amount"`
		Paid       float64 `gorm:"column:paid"`
		CreatedAt  string  `gorm:"column:created_at"`
	}
	h.db.Raw(`SELECT b.bill_no, b.status, COALESCE(e.name,'') AS handler,
		b.total_amount AS amount, b.paid_amount AS paid, b.created_at::text
		FROM sales_out_stocks b
		LEFT JOIN employees e ON e.id = b.operator_id
		WHERE b.company_id = ? AND b.bill_date = ? AND b.deleted_at IS NULL`, companyID, date).Scan(&outRows)
	for _, r := range outRows {
		if !match(r.BillNo) {
			continue
		}
		rows = append(rows, closeRow{
			BillType: "销售出库单", BillNo: r.BillNo, Status: r.Status, StatusLabel: statusLabelOf(r.Status),
			HandlerName: r.Handler, PayStatus: payStatusOf(r.Amount, r.Paid),
			Amount: r.Amount, Balance: r.Amount - r.Paid, Accounts: map[string]float64{}, CreatedAt: r.CreatedAt,
		})
	}

	// 销售退货单
	var retRows []struct {
		BillNo    string  `gorm:"column:bill_no"`
		Status    string  `gorm:"column:status"`
		Handler   string  `gorm:"column:handler"`
		Amount    float64 `gorm:"column:amount"`
		CreatedAt string  `gorm:"column:created_at"`
	}
	h.db.Raw(`SELECT b.bill_no, b.status, COALESCE(e.name,'') AS handler,
		b.total_amount AS amount, b.created_at::text
		FROM sales_returns b
		LEFT JOIN employees e ON e.id = b.operator_id
		WHERE b.company_id = ? AND b.bill_date = ? AND b.deleted_at IS NULL`, companyID, date).Scan(&retRows)
	for _, r := range retRows {
		if !match(r.BillNo) {
			continue
		}
		rows = append(rows, closeRow{
			BillType: "销售退货单", BillNo: r.BillNo, Status: r.Status, StatusLabel: statusLabelOf(r.Status),
			HandlerName: r.Handler, PayStatus: "-", Amount: -r.Amount, Accounts: map[string]float64{}, CreatedAt: r.CreatedAt,
		})
	}

	// 销售收款单（按账户分摊）
	var rcptRows []struct {
		BillNo    string  `gorm:"column:bill_no"`
		Status    string  `gorm:"column:status"`
		Handler   string  `gorm:"column:handler"`
		Amount    float64 `gorm:"column:amount"`
		Account   string  `gorm:"column:account"`
		CreatedAt string  `gorm:"column:created_at"`
	}
	h.db.Raw(`SELECT b.bill_no, b.status, COALESCE(e.name,'') AS handler,
		b.total_amount AS amount, COALESCE(a.name,'') AS account, b.created_at::text
		FROM sales_receipts b
		LEFT JOIN employees e ON e.id = b.operator_id
		LEFT JOIN accounts a ON a.id = b.account_id
		WHERE b.company_id = ? AND b.bill_date = ? AND b.deleted_at IS NULL`, companyID, date).Scan(&rcptRows)
	for _, r := range rcptRows {
		if !match(r.BillNo) {
			continue
		}
		accs := map[string]float64{}
		if r.Account != "" && r.Status == "completed" {
			accs[r.Account] = r.Amount
		}
		rows = append(rows, closeRow{
			BillType: "销售收款单", BillNo: r.BillNo, Status: r.Status, StatusLabel: statusLabelOf(r.Status),
			HandlerName: r.Handler, PayStatus: "-", Amount: r.Amount, Accounts: accs, CreatedAt: r.CreatedAt,
		})
	}

	accountNames := make([]string, 0, len(accounts))
	for _, a := range accounts {
		accountNames = append(accountNames, a.Name)
	}
	response.Ok(c, gin.H{
		"list": rows, "total": len(rows), "page": 1, "pageSize": len(rows),
		"accounts": accountNames,
	})
}

// ==================== 商品销售统计 ====================

func (h *SaleReportHandler) ProductStats(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	type stat struct {
		ProductID uint    `json:"productId"`
		Name      string  `json:"name"`
		Spec      string  `json:"spec"`
		Unit      string  `json:"unit"`
		Code      string  `json:"code"`
		Qty       float64 `json:"qty"`
		Amount    float64 `json:"amount"`
		Cost      float64 `json:"cost"`
		Profit    float64 `json:"profit"`
		ProfitPct float64 `json:"profitPct"`
		SalePct   float64 `json:"salePct"`
		BillCount int     `json:"billCount"`
	}
	stats := map[uint]*stat{}

	var outRows []struct {
		ProductID uint    `gorm:"column:product_id"`
		Name      string  `gorm:"column:name"`
		Spec      string  `gorm:"column:spec"`
		Unit      string  `gorm:"column:unit"`
		Code      string  `gorm:"column:code"`
		Qty       float64 `gorm:"column:qty"`
		Amount    float64 `gorm:"column:amount"`
		Cost      float64 `gorm:"column:cost"`
		BillCount int     `gorm:"column:bill_count"`
	}
	h.db.Raw(`SELECT p.id AS product_id, p.name, p.specification AS spec, p.unit, p.code,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount,
		SUM(i.quantity * p.purchase_price) AS cost, COUNT(DISTINCT b.id) AS bill_count
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY p.id, p.name, p.specification, p.unit, p.code`, companyID, start, end).Scan(&outRows)
	for _, r := range outRows {
		stats[r.ProductID] = &stat{
			ProductID: r.ProductID, Name: r.Name, Spec: r.Spec, Unit: r.Unit, Code: r.Code,
			Qty: r.Qty, Amount: r.Amount, Cost: r.Cost, BillCount: r.BillCount,
		}
	}

	// 退货冲减
	var retRows []struct {
		ProductID uint    `gorm:"column:product_id"`
		Qty       float64 `gorm:"column:qty"`
		Amount    float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT i.product_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amount
		FROM sales_return_items i
		JOIN sales_returns b ON b.id = i.return_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY i.product_id`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if s, ok := stats[r.ProductID]; ok {
			s.Qty -= r.Qty
			s.Amount -= r.Amount
			if s.Qty != 0 {
				s.Cost = s.Qty * (s.Cost / (s.Qty + r.Qty))
			}
		}
	}

	list := make([]stat, 0, len(stats))
	totalAmount := 0.0
	for _, s := range stats {
		s.Profit = s.Amount - s.Cost
		if s.Amount != 0 {
			s.ProfitPct = s.Profit / s.Amount * 100
		}
		if keyword == "" || strings.Contains(s.Name, keyword) || strings.Contains(s.Code, keyword) || strings.Contains(s.Spec, keyword) {
			totalAmount += s.Amount
			list = append(list, *s)
		}
	}
	for i := range list {
		if totalAmount != 0 {
			list[i].SalePct = list[i].Amount / totalAmount * 100
		}
	}
	sortByFloatDesc(list, func(r stat) float64 { return r.Amount })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 客户销售统计 ====================

func (h *SaleReportHandler) CustomerStats(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	type stat struct {
		CustomerID uint    `json:"customerId"`
		Name       string  `json:"name"`
		Code       string  `json:"code"`
		Category   string  `json:"category"`
		Qty        float64 `json:"qty"`
		Amount     float64 `json:"amount"`
		Receipt    float64 `json:"receipt"`
		SalePct    float64 `json:"salePct"`
		BillCount  int     `json:"billCount"`
	}
	stats := map[uint]*stat{}

	var outRows []struct {
		CustomerID uint    `gorm:"column:customer_id"`
		Name       string  `gorm:"column:name"`
		Code       string  `gorm:"column:code"`
		Category   string  `gorm:"column:category"`
		Qty        float64 `gorm:"column:qty"`
		Amount     float64 `gorm:"column:amount"`
		BillCount  int     `gorm:"column:bill_count"`
	}
	h.db.Raw(`SELECT cu.id AS customer_id, cu.name, cu.code, COALESCE(cc.name,'') AS category,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount, COUNT(DISTINCT b.id) AS bill_count
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN customers cu ON cu.id = b.customer_id
		LEFT JOIN customer_categories cc ON cc.id = cu.category_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY cu.id, cu.name, cu.code, cc.name`, companyID, start, end).Scan(&outRows)
	for _, r := range outRows {
		stats[r.CustomerID] = &stat{
			CustomerID: r.CustomerID, Name: r.Name, Code: r.Code, Category: r.Category,
			Qty: r.Qty, Amount: r.Amount, BillCount: r.BillCount,
		}
	}

	// 退货冲减
	var retRows []struct {
		CustomerID uint    `gorm:"column:customer_id"`
		Qty        float64 `gorm:"column:qty"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT b.customer_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amount
		FROM sales_return_items i
		JOIN sales_returns b ON b.id = i.return_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.customer_id`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if s, ok := stats[r.CustomerID]; ok {
			s.Qty -= r.Qty
			s.Amount -= r.Amount
		}
	}

	// 收款
	var rcptRows []struct {
		CustomerID uint    `gorm:"column:customer_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT customer_id, SUM(total_amount) AS amount
		FROM sales_receipts
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL
		AND bill_date >= ? AND bill_date <= ?
		GROUP BY customer_id`, companyID, start, end).Scan(&rcptRows)
	for _, r := range rcptRows {
		if s, ok := stats[r.CustomerID]; ok {
			s.Receipt = r.Amount
		} else {
			stats[r.CustomerID] = &stat{CustomerID: r.CustomerID, Receipt: r.Amount}
		}
	}

	list := make([]stat, 0, len(stats))
	totalAmount := 0.0
	for _, s := range stats {
		if keyword == "" || strings.Contains(s.Name, keyword) || strings.Contains(s.Code, keyword) {
			totalAmount += s.Amount
			list = append(list, *s)
		}
	}
	for i := range list {
		if totalAmount != 0 {
			list[i].SalePct = list[i].Amount / totalAmount * 100
		}
	}
	sortByFloatDesc(list, func(r stat) float64 { return r.Amount })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 职员销售业绩统计 ====================

func (h *SaleReportHandler) EmployeeStats(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	type stat struct {
		EmployeeID uint    `json:"employeeId"`
		Name       string  `json:"name"`
		Dept       string  `json:"dept"`
		Qty        float64 `json:"qty"`
		Amount     float64 `json:"amount"`
		Receipt    float64 `json:"receipt"`
		Customers  int     `json:"customers"`
		BillCount  int     `json:"billCount"`
		SalePct    float64 `json:"salePct"`
	}
	stats := map[uint]*stat{}

	var outRows []struct {
		EmployeeID uint    `gorm:"column:employee_id"`
		Name       string  `gorm:"column:name"`
		Dept       string  `gorm:"column:dept"`
		Qty        float64 `gorm:"column:qty"`
		Amount     float64 `gorm:"column:amount"`
		Customers  int     `gorm:"column:customers"`
		BillCount  int     `gorm:"column:bill_count"`
	}
	h.db.Raw(`SELECT e.id AS employee_id, e.name, COALESCE(d.name,'') AS dept,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount,
		COUNT(DISTINCT b.customer_id) AS customers, COUNT(DISTINCT b.id) AS bill_count
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN employees e ON e.id = b.operator_id
		LEFT JOIN departments d ON d.id = e.dept_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY e.id, e.name, d.name`, companyID, start, end).Scan(&outRows)
	for _, r := range outRows {
		stats[r.EmployeeID] = &stat{
			EmployeeID: r.EmployeeID, Name: r.Name, Dept: r.Dept,
			Qty: r.Qty, Amount: r.Amount, Customers: r.Customers, BillCount: r.BillCount,
		}
	}

	// 退货冲减
	var retRows []struct {
		EmployeeID uint    `gorm:"column:employee_id"`
		Qty        float64 `gorm:"column:qty"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT b.operator_id AS employee_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amount
		FROM sales_return_items i
		JOIN sales_returns b ON b.id = i.return_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.operator_id`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if s, ok := stats[r.EmployeeID]; ok {
			s.Qty -= r.Qty
			s.Amount -= r.Amount
		}
	}

	// 回款
	var rcptRows []struct {
		EmployeeID uint    `gorm:"column:employee_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT operator_id AS employee_id, SUM(total_amount) AS amount
		FROM sales_receipts
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL
		AND bill_date >= ? AND bill_date <= ?
		GROUP BY operator_id`, companyID, start, end).Scan(&rcptRows)
	for _, r := range rcptRows {
		if s, ok := stats[r.EmployeeID]; ok {
			s.Receipt = r.Amount
		}
	}

	list := make([]stat, 0, len(stats))
	totalAmount := 0.0
	for _, s := range stats {
		if keyword == "" || strings.Contains(s.Name, keyword) {
			totalAmount += s.Amount
			list = append(list, *s)
		}
	}
	for i := range list {
		if totalAmount != 0 {
			list[i].SalePct = list[i].Amount / totalAmount * 100
		}
	}
	sortByFloatDesc(list, func(r stat) float64 { return r.Amount })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 商品销售明细统计 ====================

func (h *SaleReportHandler) ProductDetail(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	customerKeyword := c.Query("customerKeyword")

	type row struct {
		BillType   string  `json:"billType"`
		BillNo     string  `json:"billNo"`
		Customer   string  `json:"customer"`
		Handler    string  `json:"handler"`
		Product    string  `json:"product"`
		Spec       string  `json:"spec"`
		Warehouse  string  `json:"warehouse"`
		Qty        float64 `json:"qty"`
		Unit       string  `json:"unit"`
		Price      float64 `json:"price"`
		Amount     float64 `json:"amount"`
		CostPrice  float64 `json:"costPrice"`
		CostAmount float64 `json:"costAmount"`
		Profit     float64 `json:"profit"`
		ProfitPct  float64 `json:"profitPct"`
		Remark     string  `json:"remark"`
		CreatedAt  string  `json:"createdAt"`
	}
	rows := []row{}

	var outRows []struct {
		BillNo    string  `gorm:"column:bill_no"`
		Customer  string  `gorm:"column:customer"`
		Handler   string  `gorm:"column:handler"`
		Product   string  `gorm:"column:product"`
		Spec      string  `gorm:"column:spec"`
		Warehouse string  `gorm:"column:warehouse"`
		Qty       float64 `gorm:"column:qty"`
		Unit      string  `gorm:"column:unit"`
		Price     float64 `gorm:"column:price"`
		Amount    float64 `gorm:"column:amount"`
		CostPrice float64 `gorm:"column:cost_price"`
		Remark    string  `gorm:"column:remark"`
		CreatedAt string  `gorm:"column:created_at"`
	}
	h.db.Raw(`SELECT b.bill_no, COALESCE(cu.name,'') AS customer, COALESCE(e.name,'') AS handler,
		p.name AS product, p.specification AS spec, COALESCE(w.name,'') AS warehouse,
		i.quantity AS qty, p.unit, i.price, i.amount, p.purchase_price AS cost_price,
		COALESCE(i.remark,'') AS remark, b.created_at::text
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		LEFT JOIN customers cu ON cu.id = b.customer_id
		LEFT JOIN employees e ON e.id = b.operator_id
		JOIN products p ON p.id = i.product_id
		LEFT JOIN warehouses w ON w.id = b.warehouse_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?`, companyID, start, end).Scan(&outRows)
	for _, r := range outRows {
		if keyword != "" && !strings.Contains(r.Product, keyword) {
			continue
		}
		if customerKeyword != "" && !strings.Contains(r.Customer, customerKeyword) {
			continue
		}
		cost := r.Qty * r.CostPrice
		profit := r.Amount - cost
		pct := 0.0
		if r.Amount != 0 {
			pct = profit / r.Amount * 100
		}
		rows = append(rows, row{
			BillType: "销售出库单", BillNo: r.BillNo, Customer: r.Customer, Handler: r.Handler,
			Product: r.Product, Spec: r.Spec, Warehouse: r.Warehouse, Qty: r.Qty, Unit: r.Unit,
			Price: r.Price, Amount: r.Amount, CostPrice: r.CostPrice, CostAmount: cost,
			Profit: profit, ProfitPct: pct, Remark: r.Remark, CreatedAt: r.CreatedAt,
		})
	}

	var retRows []struct {
		BillNo    string  `gorm:"column:bill_no"`
		Customer  string  `gorm:"column:customer"`
		Handler   string  `gorm:"column:handler"`
		Product   string  `gorm:"column:product"`
		Spec      string  `gorm:"column:spec"`
		Warehouse string  `gorm:"column:warehouse"`
		Qty       float64 `gorm:"column:qty"`
		Unit      string  `gorm:"column:unit"`
		Price     float64 `gorm:"column:price"`
		Amount    float64 `gorm:"column:amount"`
		CostPrice float64 `gorm:"column:cost_price"`
		Remark    string  `gorm:"column:remark"`
		CreatedAt string  `gorm:"column:created_at"`
	}
	h.db.Raw(`SELECT b.bill_no, COALESCE(cu.name,'') AS customer, COALESCE(e.name,'') AS handler,
		p.name AS product, p.specification AS spec, COALESCE(w.name,'') AS warehouse,
		i.quantity AS qty, p.unit, i.price, i.amount, p.purchase_price AS cost_price,
		COALESCE(i.remark,'') AS remark, b.created_at::text
		FROM sales_return_items i
		JOIN sales_returns b ON b.id = i.return_id
		LEFT JOIN customers cu ON cu.id = b.customer_id
		LEFT JOIN employees e ON e.id = b.operator_id
		JOIN products p ON p.id = i.product_id
		LEFT JOIN warehouses w ON w.id = b.warehouse_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if keyword != "" && !strings.Contains(r.Product, keyword) {
			continue
		}
		if customerKeyword != "" && !strings.Contains(r.Customer, customerKeyword) {
			continue
		}
		cost := r.Qty * r.CostPrice
		profit := -(r.Amount - cost)
		pct := 0.0
		if r.Amount != 0 {
			pct = profit / r.Amount * 100
		}
		rows = append(rows, row{
			BillType: "销售退货单", BillNo: r.BillNo, Customer: r.Customer, Handler: r.Handler,
			Product: r.Product, Spec: r.Spec, Warehouse: r.Warehouse, Qty: -r.Qty, Unit: r.Unit,
			Price: r.Price, Amount: -r.Amount, CostPrice: r.CostPrice, CostAmount: -cost,
			Profit: profit, ProfitPct: pct, Remark: r.Remark, CreatedAt: r.CreatedAt,
		})
	}

	sortByStringDesc(rows, func(r row) string { return r.CreatedAt })
	response.Ok(c, paginateSaleRows(rows, page, pageSize))
}

// ==================== 销售订单执行明细 ====================

func (h *SaleReportHandler) OrderExecution(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	type row struct {
		OrderNo    string  `json:"orderNo"`
		Customer   string  `json:"customer"`
		Status     string  `json:"status"`
		StatusLabel string `json:"statusLabel"`
		Paid       float64 `json:"paid"`
		Product    string  `json:"product"`
		Spec       string  `json:"spec"`
		Unit       string  `json:"unit"`
		Warehouse  string  `json:"warehouse"`
		OrderQty   float64 `json:"orderQty"`
		StoppedQty float64 `json:"stoppedQty"`
		Price      float64 `json:"price"`
		Amount     float64 `json:"amount"`
		OutQty     float64 `json:"outQty"`
		PendingQty float64 `json:"pendingQty"`
		CreatedAt  string  `json:"createdAt"`
	}

	var raw []struct {
		OrderNo   string  `gorm:"column:order_no"`
		Customer  string  `gorm:"column:customer"`
		Status    string  `gorm:"column:status"`
		Paid      float64 `gorm:"column:paid"`
		Product   string  `gorm:"column:product"`
		Spec      string  `gorm:"column:spec"`
		Unit      string  `gorm:"column:unit"`
		Warehouse string  `gorm:"column:warehouse"`
		Qty       float64 `gorm:"column:qty"`
		Price     float64 `gorm:"column:price"`
		Amount    float64 `gorm:"column:amount"`
		Delivered float64 `gorm:"column:delivered"`
		CreatedAt string  `gorm:"column:created_at"`
	}
	h.db.Raw(`SELECT b.order_no, COALESCE(cu.name,'') AS customer, b.status, b.paid_amount AS paid,
		p.name AS product, p.specification AS spec, p.unit, COALESCE(w.name,'') AS warehouse,
		i.quantity AS qty, i.price, i.amount, i.delivered_qty AS delivered, b.created_at::text
		FROM sales_order_items i
		JOIN sales_orders b ON b.id = i.order_id
		LEFT JOIN customers cu ON cu.id = b.customer_id
		JOIN products p ON p.id = i.product_id
		LEFT JOIN warehouses w ON w.id = b.warehouse_id
		WHERE b.company_id = ? AND b.deleted_at IS NULL AND b.status NOT IN ('draft','cancelled')
		AND b.order_date >= ? AND b.order_date <= ?`, companyID, start, end).Scan(&raw)

	rows := make([]row, 0, len(raw))
	for _, r := range raw {
		if keyword != "" && !strings.Contains(r.Product, keyword) && !strings.Contains(r.OrderNo, keyword) {
			continue
		}
		pending := r.Qty - r.Delivered
		if pending < 0 {
			pending = 0
		}
		rows = append(rows, row{
			OrderNo: r.OrderNo, Customer: r.Customer, Status: r.Status, StatusLabel: statusLabelOf(r.Status),
			Paid: r.Paid, Product: r.Product, Spec: r.Spec, Unit: r.Unit, Warehouse: r.Warehouse,
			OrderQty: r.Qty, StoppedQty: 0, Price: r.Price, Amount: r.Amount,
			OutQty: r.Delivered, PendingQty: pending, CreatedAt: r.CreatedAt,
		})
	}
	sortByStringDesc(rows, func(r row) string { return r.CreatedAt })
	response.Ok(c, paginateSaleRows(rows, page, pageSize))
}

// ==================== 订销退综合分析 ====================

func (h *SaleReportHandler) Comprehensive(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	type stat struct {
		ProductID   uint    `json:"productId"`
		Name        string  `json:"name"`
		Spec        string  `json:"spec"`
		Unit        string  `json:"unit"`
		Code        string  `json:"code"`
		OrderQty    float64 `json:"orderQty"`
		OrderAmount float64 `json:"orderAmount"`
		OrderCount  int     `json:"orderCount"`
		SaleQty     float64 `json:"saleQty"`
		SaleAmount  float64 `json:"saleAmount"`
		SaleCount   int     `json:"saleCount"`
		ReturnQty   float64 `json:"returnQty"`
		ReturnAmount float64 `json:"returnAmount"`
		ReturnCount int     `json:"returnCount"`
	}
	stats := map[uint]*stat{}

	type prodAgg struct {
		ProductID uint    `gorm:"column:product_id"`
		Name      string  `gorm:"column:name"`
		Spec      string  `gorm:"column:spec"`
		Unit      string  `gorm:"column:unit"`
		Code      string  `gorm:"column:code"`
		Qty       float64 `gorm:"column:qty"`
		Amount    float64 `gorm:"column:amount"`
		BillCount int     `gorm:"column:bill_count"`
	}
	get := func(p prodAgg) *stat {
		if s, ok := stats[p.ProductID]; ok {
			return s
		}
		s := &stat{ProductID: p.ProductID, Name: p.Name, Spec: p.Spec, Unit: p.Unit, Code: p.Code}
		stats[p.ProductID] = s
		return s
	}

	// 订货
	var orderRows []prodAgg
	h.db.Raw(`SELECT p.id AS product_id, p.name, p.specification AS spec, p.unit, p.code,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount, COUNT(DISTINCT b.id) AS bill_count
		FROM sales_order_items i
		JOIN sales_orders b ON b.id = i.order_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.deleted_at IS NULL AND b.status NOT IN ('draft','cancelled')
		AND b.order_date >= ? AND b.order_date <= ?
		GROUP BY p.id, p.name, p.specification, p.unit, p.code`, companyID, start, end).Scan(&orderRows)
	for _, r := range orderRows {
		s := get(r)
		s.OrderQty = r.Qty
		s.OrderAmount = r.Amount
		s.OrderCount = r.BillCount
	}

	// 销售
	var saleRows []prodAgg
	h.db.Raw(`SELECT p.id AS product_id, p.name, p.specification AS spec, p.unit, p.code,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount, COUNT(DISTINCT b.id) AS bill_count
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY p.id, p.name, p.specification, p.unit, p.code`, companyID, start, end).Scan(&saleRows)
	for _, r := range saleRows {
		s := get(r)
		s.SaleQty = r.Qty
		s.SaleAmount = r.Amount
		s.SaleCount = r.BillCount
	}

	// 退货
	var retRows []prodAgg
	h.db.Raw(`SELECT p.id AS product_id, p.name, p.specification AS spec, p.unit, p.code,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount, COUNT(DISTINCT b.id) AS bill_count
		FROM sales_return_items i
		JOIN sales_returns b ON b.id = i.return_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY p.id, p.name, p.specification, p.unit, p.code`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		s := get(r)
		s.ReturnQty = r.Qty
		s.ReturnAmount = r.Amount
		s.ReturnCount = r.BillCount
	}

	list := make([]stat, 0, len(stats))
	for _, s := range stats {
		if keyword == "" || strings.Contains(s.Name, keyword) || strings.Contains(s.Code, keyword) {
			list = append(list, *s)
		}
	}
	sortByFloatDesc(list, func(r stat) float64 { return r.SaleAmount })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 提成方案 CRUD ====================

func (h *SaleReportHandler) PlanList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	status := c.Query("status")

	q := h.db.Model(&model.CommissionPlan{}).Where("company_id = ?", companyID)
	if keyword != "" {
		q = q.Where("name ILIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var list []model.CommissionPlan
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OkWithPage(c, list, page, pageSize, int(total))
}

func (h *SaleReportHandler) PlanAll(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var list []model.CommissionPlan
	h.db.Where("company_id = ? AND status = 1", companyID).Order("id DESC").Find(&list)
	response.Ok(c, gin.H{"list": list})
}

func (h *SaleReportHandler) PlanCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var plan model.CommissionPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	plan.ID = 0
	plan.CompanyID = companyID
	if plan.Name == "" {
		response.Fail(c, 4000, "方案名称不能为空")
		return
	}
	if plan.CalcType == "" {
		plan.CalcType = "amount_percent"
	}
	if err := h.db.Create(&plan).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, plan)
}

func (h *SaleReportHandler) PlanUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var plan model.CommissionPlan
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&plan).Error; err != nil {
		response.Fail(c, 4004, "方案不存在")
		return
	}
	var req model.CommissionPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	plan.Name = req.Name
	plan.StartDate = req.StartDate
	plan.EndDate = req.EndDate
	plan.EmployeeIDs = req.EmployeeIDs
	plan.CalcType = req.CalcType
	plan.Rate = req.Rate
	plan.Status = req.Status
	plan.Remark = req.Remark
	if err := h.db.Save(&plan).Error; err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, plan)
}

func (h *SaleReportHandler) PlanDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.CommissionPlan{}).Error; err != nil {
		response.Fail(c, 4000, "删除失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

// ==================== 职员提成统计 ====================

func (h *SaleReportHandler) CommissionStats(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	planID, _ := strconv.Atoi(c.Query("planId"))
	start := c.DefaultQuery("startDate", time.Now().Format("2006-01-02"))
	end := c.DefaultQuery("endDate", time.Now().Format("2006-01-02"))

	var plan model.CommissionPlan
	if err := h.db.Where("id = ? AND company_id = ?", planID, companyID).First(&plan).Error; err != nil {
		response.Fail(c, 4004, "提成方案不存在")
		return
	}

	// 方案限定职员
	empFilter := map[uint]bool{}
	if plan.EmployeeIDs != "" {
		for _, s := range strings.Split(plan.EmployeeIDs, ",") {
			if id, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && id > 0 {
				empFilter[uint(id)] = true
			}
		}
	}

	type stat struct {
		EmployeeID uint    `json:"employeeId"`
		Name       string  `json:"name"`
		Dept       string  `json:"dept"`
		PlanName   string  `json:"planName"`
		Commission float64 `json:"commission"`
		SaleAmount float64 `json:"saleAmount"`
		SaleQty    float64 `json:"saleQty"`
		Receipt    float64 `json:"receipt"`
		Profit     float64 `json:"profit"`
	}
	stats := map[uint]*stat{}

	var outRows []struct {
		EmployeeID uint    `gorm:"column:employee_id"`
		Name       string  `gorm:"column:name"`
		Dept       string  `gorm:"column:dept"`
		Qty        float64 `gorm:"column:qty"`
		Amount     float64 `gorm:"column:amount"`
		Cost       float64 `gorm:"column:cost"`
	}
	h.db.Raw(`SELECT e.id AS employee_id, e.name, COALESCE(d.name,'') AS dept,
		SUM(i.quantity) AS qty, SUM(i.amount) AS amount, SUM(i.quantity * p.purchase_price) AS cost
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN employees e ON e.id = b.operator_id
		LEFT JOIN departments d ON d.id = e.dept_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY e.id, e.name, d.name`, companyID, start, end).Scan(&outRows)
	for _, r := range outRows {
		if len(empFilter) > 0 && !empFilter[r.EmployeeID] {
			continue
		}
		stats[r.EmployeeID] = &stat{
			EmployeeID: r.EmployeeID, Name: r.Name, Dept: r.Dept, PlanName: plan.Name,
			SaleAmount: r.Amount, SaleQty: r.Qty, Profit: r.Amount - r.Cost,
		}
	}

	// 回款
	var rcptRows []struct {
		EmployeeID uint    `gorm:"column:employee_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT operator_id AS employee_id, SUM(total_amount) AS amount
		FROM sales_receipts
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL
		AND bill_date >= ? AND bill_date <= ?
		GROUP BY operator_id`, companyID, start, end).Scan(&rcptRows)
	for _, r := range rcptRows {
		if s, ok := stats[r.EmployeeID]; ok {
			s.Receipt = r.Amount
		}
	}

	list := make([]stat, 0, len(stats))
	for _, s := range stats {
		switch plan.CalcType {
		case "profit_percent":
			s.Commission = s.Profit * plan.Rate / 100
		case "qty_fixed":
			s.Commission = s.SaleQty * plan.Rate
		default: // amount_percent
			s.Commission = s.SaleAmount * plan.Rate / 100
		}
		list = append(list, *s)
	}
	sortByFloatDesc(list, func(r stat) float64 { return r.Commission })
	response.Ok(c, gin.H{"list": list, "total": len(list), "page": 1, "pageSize": len(list), "planName": plan.Name})
}

// ==================== 通用小工具 ====================

// sortByStringDesc 按字符串 key 倒序
func sortByStringDesc[T any](list []T, key func(T) string) {
	sort.SliceStable(list, func(i, j int) bool { return key(list[i]) > key(list[j]) })
}

// sortByFloatDesc 按数值 key 倒序
func sortByFloatDesc[T any](list []T, key func(T) float64) {
	sort.SliceStable(list, func(i, j int) bool { return key(list[i]) > key(list[j]) })
}

func paginateSaleRows[T any](list []T, page, pageSize int) gin.H {
	total := len(list)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return gin.H{"list": list[start:end], "total": total, "page": page, "pageSize": pageSize}
}
