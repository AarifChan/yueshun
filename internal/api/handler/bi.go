package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// BIHandler BI 分析（商城上线/新客跟踪/老客增品/流失拉回/商品铺市）
type BIHandler struct {
	db *gorm.DB
}

func NewBIHandler(db *gorm.DB) *BIHandler {
	return &BIHandler{db: db}
}

func (h *BIHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/bi")
	{
		g.GET("/mall-online/summary", h.MallOnlineSummary)
		g.GET("/mall-online/detail", h.MallOnlineDetail)
		g.GET("/new-customers/summary", h.NewCustomerSummary)
		g.GET("/new-customers/detail", h.NewCustomerDetail)
		g.GET("/old-customers/summary", h.OldCustomerSummary)
		g.GET("/old-customers/detail", h.OldCustomerDetail)
		g.GET("/lost-customers/summary", h.LostCustomerSummary)
		g.GET("/lost-customers/detail", h.LostCustomerDetail)
		g.GET("/product-distribution/summary", h.ProductDistSummary)
		g.GET("/product-distribution/detail", h.ProductDistDetail)
	}
}

// dimInfo 客户维度映射
type dimInfo struct {
	owners  map[uint]string // 职员 id→name
	depts   map[uint]string // 部门 id→name
	empDept map[uint]uint   // 职员 id→部门 id
	regions map[uint]string
	cats    map[uint]string
}

func (h *BIHandler) loadDimInfo(companyID uint) *dimInfo {
	d := &dimInfo{
		owners: map[uint]string{}, depts: map[uint]string{}, empDept: map[uint]uint{},
		regions: map[uint]string{}, cats: map[uint]string{},
	}
	var emps []model.Employee
	h.db.Where("company_id = ?", companyID).Find(&emps)
	for _, e := range emps {
		d.owners[e.ID] = e.Name
		d.empDept[e.ID] = e.DeptID
	}
	var depts []model.Department
	h.db.Where("company_id = ?", companyID).Find(&depts)
	for _, x := range depts {
		d.depts[x.ID] = x.Name
	}
	var regions []model.Region
	h.db.Where("company_id = ?", companyID).Find(&regions)
	for _, x := range regions {
		d.regions[x.ID] = x.Name
	}
	var cats []model.CustomerCategory
	h.db.Where("company_id = ?", companyID).Find(&cats)
	for _, x := range cats {
		d.cats[x.ID] = x.Name
	}
	return d
}

// dimKeyName 按维度返回分组 key 与名称
func (d *dimInfo) dimKeyName(dimension string, ownerID, regionID, categoryID uint) (string, string) {
	switch dimension {
	case "dept":
		deptID := d.empDept[ownerID]
		return fmt.Sprintf("d%d", deptID), valueOr(d.depts[deptID], "未分配部门")
	case "region":
		return fmt.Sprintf("r%d", regionID), valueOr(d.regions[regionID], "未设置区域")
	case "category":
		return fmt.Sprintf("c%d", categoryID), valueOr(d.cats[categoryID], "未分类")
	default: // employee / manager 都按客户负责人
		return fmt.Sprintf("e%d", ownerID), valueOr(d.owners[ownerID], "未分配职员")
	}
}

func valueOr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// ==================== 客户商城上线 ====================

// mallChannelRow 期内/期初每客户商城与线下单数
type mallChannelRow struct {
	CustomerID uint `gorm:"column:customer_id"`
	MallCnt    int  `gorm:"column:mall_cnt"`
	OffCnt     int  `gorm:"column:off_cnt"`
}

func (h *BIHandler) mallChannelCounts(companyID uint, start, end string, before bool) map[uint]mallChannelRow {
	cond := "order_date >= ? AND order_date <= ?"
	args := []interface{}{companyID, start, end, companyID, start, end}
	if before {
		cond = "order_date < ?"
		args = []interface{}{companyID, start, companyID, start}
	}
	rows := []mallChannelRow{}
	h.db.Raw(`SELECT customer_id, COALESCE(SUM(mall_cnt),0)::int AS mall_cnt, COALESCE(SUM(off_cnt),0)::int AS off_cnt FROM (
		SELECT customer_id, COUNT(*) AS mall_cnt, 0 AS off_cnt FROM mall_orders
		WHERE company_id = ? AND status NOT IN ('cancelled') AND deleted_at IS NULL AND `+cond+` GROUP BY customer_id
		UNION ALL
		SELECT customer_id, 0 AS mall_cnt, COUNT(*) AS off_cnt FROM sales_orders
		WHERE company_id = ? AND status NOT IN ('draft','cancelled') AND deleted_at IS NULL AND `+cond+` GROUP BY customer_id
	) t GROUP BY customer_id`, args...).Scan(&rows)
	m := map[uint]mallChannelRow{}
	for _, r := range rows {
		m[r.CustomerID] = r
	}
	return m
}

// channelClass 全商城/部分商城/全线下/无交易
func channelClass(mallCnt, offCnt int) string {
	switch {
	case mallCnt > 0 && offCnt > 0:
		return "部分商城下单"
	case mallCnt > 0:
		return "全商城下单"
	case offCnt > 0:
		return "全线下下单"
	default:
		return ""
	}
}

func (h *BIHandler) mallOnlineRows(c *gin.Context) ([]gin.H, *dimInfo, string, string) {
	companyID := middleware.GetCompanyID(c)
	_, _, start, end := salePageParams(c)
	customerType := c.DefaultQuery("customerType", "all") // new=本期新注册 all=所有注册
	dimension := c.DefaultQuery("dimension", "employee")

	dim := h.loadDimInfo(companyID)
	period := h.mallChannelCounts(companyID, start, end, false)
	initial := h.mallChannelCounts(companyID, start, end, true)

	var customers []model.Customer
	q := h.db.Where("company_id = ? AND deleted_at IS NULL AND type != 'supplier'", companyID)
	if customerType == "new" {
		q = q.Where("created_at::date >= ? AND created_at::date <= ?", start, end)
	}
	q.Find(&customers)

	rows := make([]gin.H, 0, len(customers))
	for _, cu := range customers {
		p := period[cu.ID]
		i := initial[cu.ID]
		pClass := channelClass(p.MallCnt, p.OffCnt)
		iClass := channelClass(i.MallCnt, i.OffCnt)
		key, name := dim.dimKeyName(dimension, cu.OwnerID, cu.RegionID, cu.CategoryID)
		isNew := cu.CreatedAt.Format("2006-01-02") >= start && cu.CreatedAt.Format("2006-01-02") <= end
		rows = append(rows, gin.H{
			"customerId": cu.ID, "customer": cu.Name, "code": cu.Code,
			"dimKey": key, "dimName": name, "ownerId": cu.OwnerID,
			"mallCnt": p.MallCnt, "offCnt": p.OffCnt, "class": pClass,
			"initMallCnt": i.MallCnt, "initOffCnt": i.OffCnt, "initClass": iClass,
			"isNew": isNew, "createdAt": cu.CreatedAt,
			"newAllMall":  pClass == "全商城下单" && iClass != "全商城下单",
			"lostAllMall": iClass == "全商城下单" && pClass != "全商城下单",
		})
	}
	return rows, dim, start, end
}

func (h *BIHandler) MallOnlineSummary(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	rows, _, _, _ := h.mallOnlineRows(c)

	type agg struct {
		Customers, AllMall, PartMall, AllOff, NotAllMall int
		MallBills, TotalBills                            int
		NewAllMall, LostAllMall, NetAllMall, NewReg      int
	}
	groups := map[string]*agg{}
	names := map[string]string{}
	order := []string{}
	for _, r := range rows {
		key := r["dimKey"].(string)
		a, ok := groups[key]
		if !ok {
			a = &agg{}
			groups[key] = a
			names[key] = r["dimName"].(string)
			order = append(order, key)
		}
		a.Customers++
		switch r["class"] {
		case "全商城下单":
			a.AllMall++
		case "部分商城下单":
			a.PartMall++
		case "全线下下单":
			a.AllOff++
		}
		if r["class"] == "部分商城下单" || r["class"] == "全线下下单" {
			a.NotAllMall++
		}
		a.MallBills += r["mallCnt"].(int)
		a.TotalBills += r["mallCnt"].(int) + r["offCnt"].(int)
		if r["newAllMall"].(bool) {
			a.NewAllMall++
		}
		if r["lostAllMall"].(bool) {
			a.LostAllMall++
		}
		if r["isNew"].(bool) {
			a.NewReg++
		}
	}
	list := make([]gin.H, 0, len(groups))
	for _, key := range order {
		a := groups[key]
		a.NetAllMall = a.NewAllMall - a.LostAllMall
		ratio := 0.0
		if a.TotalBills > 0 {
			ratio = float64(a.MallBills) / float64(a.TotalBills) * 100
		}
		list = append(list, gin.H{
			"dimName": names[key], "customers": a.Customers,
			"allMall": a.AllMall, "partMall": a.PartMall, "allOff": a.AllOff, "notAllMall": a.NotAllMall,
			"mallBills": a.MallBills, "totalBills": a.TotalBills, "mallRatio": ratio,
			"newAllMall": a.NewAllMall, "lostAllMall": a.LostAllMall, "netAllMall": a.NetAllMall, "newReg": a.NewReg,
		})
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

func (h *BIHandler) MallOnlineDetail(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	keyword := c.Query("keyword")
	rows, _, _, _ := h.mallOnlineRows(c)
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		if r["class"] == "" {
			continue
		}
		if keyword != "" && !containsAny(r["customer"].(string)+r["code"].(string), keyword) {
			continue
		}
		list = append(list, gin.H{
			"customer": r["customer"], "code": r["code"], "dimName": r["dimName"],
			"mallCnt": r["mallCnt"], "offCnt": r["offCnt"], "class": r["class"],
			"createdAt": r["createdAt"],
		})
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

func containsAny(s, kw string) bool {
	return kw == "" || strings.Contains(s, kw)
}

// ==================== 新客跟踪 ====================

type newCustRow struct {
	CustomerID  uint    `gorm:"column:customer_id"`
	Customer    string  `gorm:"column:customer"`
	Code        string  `gorm:"column:code"`
	OwnerID     uint    `gorm:"column:owner_id"`
	RegionID    uint    `gorm:"column:region_id"`
	CategoryID  uint    `gorm:"column:category_id"`
	Created     string  `gorm:"column:created"`
	BillDays    int     `gorm:"column:bill_days"`
	TradeAmount float64 `gorm:"column:trade_amount"`
	FirstBill   *string `gorm:"column:first_bill"`
	LastBill    *string `gorm:"column:last_bill"`
}

func (h *BIHandler) newCustomerRows(c *gin.Context) ([]newCustRow, *dimInfo) {
	companyID := middleware.GetCompanyID(c)
	_, _, start, end := salePageParams(c)
	mode := c.DefaultQuery("mode", "new") // total=看总数 new=看新增
	billCond := "AND b.bill_date >= ? AND b.bill_date <= ?"
	args := []interface{}{companyID, start, end, start, end}
	if mode == "total" {
		billCond = "AND b.bill_date <= ?"
		args = []interface{}{companyID, start, end, end}
	}
	var rows []newCustRow
	h.db.Raw(`SELECT c.id AS customer_id, c.name AS customer, c.code, c.owner_id, c.region_id, c.category_id,
		c.created_at::text AS created,
		COUNT(DISTINCT b.bill_date) AS bill_days,
		COALESCE(SUM(b.total_amount),0) AS trade_amount,
		MIN(b.bill_date)::text AS first_bill, MAX(b.bill_date)::text AS last_bill
		FROM customers c
		LEFT JOIN sales_out_stocks b ON b.customer_id = c.id AND b.company_id = c.company_id
			AND b.status = 'completed' AND b.deleted_at IS NULL `+billCond+`
		WHERE c.company_id = ? AND c.deleted_at IS NULL AND c.type != 'supplier'
		AND c.created_at::date >= ? AND c.created_at::date <= ?
		GROUP BY c.id, c.name, c.code, c.owner_id, c.region_id, c.category_id, c.created_at`, args...).Scan(&rows)
	return rows, h.loadDimInfo(companyID)
}

// newCustStatus 未复购/已复购/已留存
func newCustStatus(r newCustRow) string {
	if r.BillDays >= 2 && r.LastBill != nil && r.Created != "" && *r.LastBill >= addDays(r.Created[:10], 30) {
		return "已留存"
	}
	if r.BillDays >= 2 {
		return "已复购"
	}
	return "未复购"
}

func addDays(date string, days int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func (h *BIHandler) NewCustomerSummary(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	dimension := c.DefaultQuery("dimension", "employee")
	rows, dim := h.newCustomerRows(c)

	type agg struct {
		Total, NotRepeat, Repeated, Retained int
		TradeAmount                          float64
	}
	groups := map[string]*agg{}
	names := map[string]string{}
	order := []string{}
	for _, r := range rows {
		key, name := dim.dimKeyName(dimension, r.OwnerID, r.RegionID, r.CategoryID)
		a, ok := groups[key]
		if !ok {
			a = &agg{}
			groups[key] = a
			names[key] = name
			order = append(order, key)
		}
		a.Total++
		switch newCustStatus(r) {
		case "已留存":
			a.Retained++
		case "已复购":
			a.Repeated++
		default:
			a.NotRepeat++
		}
		a.TradeAmount += r.TradeAmount
	}
	list := make([]gin.H, 0, len(groups))
	for _, key := range order {
		a := groups[key]
		retainRate := 0.0
		if a.Total > 0 {
			retainRate = float64(a.Retained) / float64(a.Total) * 100
		}
		list = append(list, gin.H{
			"dimName": names[key], "customers": a.Total,
			"notRepeat": a.NotRepeat, "repeated": a.Repeated, "retained": a.Retained,
			"tradeAmount": a.TradeAmount, "retainRate": retainRate,
		})
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

func (h *BIHandler) NewCustomerDetail(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	keyword := c.Query("keyword")
	rows, _ := h.newCustomerRows(c)
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		if keyword != "" && !containsAny(r.Customer+r.Code, keyword) {
			continue
		}
		list = append(list, gin.H{
			"customer": r.Customer, "code": r.Code, "created": r.Created,
			"billDays": r.BillDays, "tradeAmount": r.TradeAmount,
			"firstBill": r.FirstBill, "lastBill": r.LastBill,
			"status": newCustStatus(r),
		})
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 老客增品 ====================

type incrementRow struct {
	CustomerID  uint    `gorm:"column:customer_id"`
	Customer    string  `gorm:"column:customer"`
	Code        string  `gorm:"column:code"`
	OwnerID     uint    `gorm:"column:owner_id"`
	RegionID    uint    `gorm:"column:region_id"`
	CategoryID  uint    `gorm:"column:category_id"`
	ProductID   uint    `gorm:"column:product_id"`
	Product     string  `gorm:"column:product"`
	ProductCode string  `gorm:"column:product_code"`
	Spec        string  `gorm:"column:spec"`
	Unit        string  `gorm:"column:unit"`
	FirstDate   string  `gorm:"column:first_date"`
	Qty         float64 `gorm:"column:qty"`
	Amount      float64 `gorm:"column:amount"`
}

func (h *BIHandler) incrementRows(c *gin.Context) ([]incrementRow, *dimInfo) {
	companyID := middleware.GetCompanyID(c)
	_, _, start, end := salePageParams(c)
	var rows []incrementRow
	h.db.Raw(`WITH before AS (
		SELECT DISTINCT b.customer_id, i.product_id FROM sales_out_stocks b
		JOIN sales_out_stock_items i ON i.out_stock_id = b.id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND b.bill_date < ?
	),
	period AS (
		SELECT b.customer_id, i.product_id, MIN(b.bill_date)::text AS first_date,
			SUM(i.quantity) AS qty, SUM(i.amount) AS amount
		FROM sales_out_stocks b
		JOIN sales_out_stock_items i ON i.out_stock_id = b.id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
			AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY b.customer_id, i.product_id
	),
	old_cust AS (
		SELECT DISTINCT customer_id FROM sales_out_stocks
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date < ?
	)
	SELECT p.customer_id, cu.name AS customer, cu.code, cu.owner_id, cu.region_id, cu.category_id,
		p.product_id, pr.name AS product, pr.code AS product_code, pr.specification AS spec, pr.unit,
		p.first_date, p.qty, p.amount
	FROM period p
	LEFT JOIN before bf ON bf.customer_id = p.customer_id AND bf.product_id = p.product_id
	JOIN old_cust oc ON oc.customer_id = p.customer_id
	JOIN customers cu ON cu.id = p.customer_id
	JOIN products pr ON pr.id = p.product_id
	WHERE bf.customer_id IS NULL`, companyID, start, companyID, start, end, companyID, start).Scan(&rows)
	return rows, h.loadDimInfo(companyID)
}

func (h *BIHandler) OldCustomerSummary(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	dimension := c.DefaultQuery("dimension", "employee")
	productKeyword := c.Query("productKeyword")
	rows, dim := h.incrementRows(c)

	type agg struct {
		Customers map[uint]bool
		Skus      map[uint]bool
		Amount    float64
	}
	groups := map[string]*agg{}
	names := map[string]string{}
	order := []string{}
	for _, r := range rows {
		if productKeyword != "" && !containsAny(r.Product+r.ProductCode+r.Spec, productKeyword) {
			continue
		}
		key, name := dim.dimKeyName(dimension, r.OwnerID, r.RegionID, r.CategoryID)
		a, ok := groups[key]
		if !ok {
			a = &agg{Customers: map[uint]bool{}, Skus: map[uint]bool{}}
			groups[key] = a
			names[key] = name
			order = append(order, key)
		}
		a.Customers[r.CustomerID] = true
		a.Skus[r.ProductID] = true
		a.Amount += r.Amount
	}
	list := make([]gin.H, 0, len(groups))
	for _, key := range order {
		a := groups[key]
		list = append(list, gin.H{
			"dimName": names[key], "customers": len(a.Customers), "skus": len(a.Skus), "tradeAmount": a.Amount,
		})
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

func (h *BIHandler) OldCustomerDetail(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	keyword := c.Query("keyword")
	productKeyword := c.Query("productKeyword")
	rows, _ := h.incrementRows(c)
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		if keyword != "" && !containsAny(r.Customer+r.Code, keyword) {
			continue
		}
		if productKeyword != "" && !containsAny(r.Product+r.ProductCode+r.Spec, productKeyword) {
			continue
		}
		list = append(list, gin.H{
			"customer": r.Customer, "code": r.Code,
			"product": r.Product, "productCode": r.ProductCode, "spec": r.Spec, "unit": r.Unit,
			"firstDate": r.FirstDate, "qty": r.Qty, "amount": r.Amount,
		})
	}
	sortByStringDesc(list, func(r gin.H) string { return r["firstDate"].(string) })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 流失拉回 ====================

type lostRow struct {
	CustomerID uint     `gorm:"column:customer_id"`
	Customer   string   `gorm:"column:customer"`
	Code       string   `gorm:"column:code"`
	OwnerID    uint     `gorm:"column:owner_id"`
	RegionID   uint     `gorm:"column:region_id"`
	CategoryID uint     `gorm:"column:category_id"`
	LastBill   string   `gorm:"column:last_bill"`
	BackDate   *string  `gorm:"column:back_date"`
	BackAmount *float64 `gorm:"column:back_amount"`
	BackBills  *int     `gorm:"column:back_bills"`
}

func (h *BIHandler) lostCustomerRows(c *gin.Context) ([]lostRow, *dimInfo) {
	companyID := middleware.GetCompanyID(c)
	_, _, start, end := salePageParams(c)
	lossDays, _ := strconv.Atoi(c.DefaultQuery("lossDays", "90"))
	if lossDays <= 0 {
		lossDays = 90
	}
	var rows []lostRow
	h.db.Raw(`WITH lastb AS (
		SELECT customer_id, MAX(bill_date) AS last_bill FROM sales_out_stocks
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date < ?
		GROUP BY customer_id
	),
	lost AS (
		SELECT customer_id, last_bill FROM lastb WHERE last_bill <= ?::date - ? * INTERVAL '1 day'
	),
	back AS (
		SELECT b.customer_id, MIN(b.bill_date)::text AS back_date,
			SUM(b.total_amount) AS back_amount, COUNT(*)::int AS back_bills
		FROM sales_out_stocks b
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
			AND b.bill_date >= ? AND b.bill_date <= ?
			AND b.customer_id IN (SELECT customer_id FROM lost)
		GROUP BY b.customer_id
	)
	SELECT l.customer_id, cu.name AS customer, cu.code, cu.owner_id, cu.region_id, cu.category_id,
		l.last_bill::text AS last_bill, b.back_date, b.back_amount, b.back_bills
	FROM lost l
	JOIN customers cu ON cu.id = l.customer_id
	LEFT JOIN back b ON b.customer_id = l.customer_id`,
		companyID, start, start, lossDays, companyID, start, end).Scan(&rows)
	return rows, h.loadDimInfo(companyID)
}

func (h *BIHandler) LostCustomerSummary(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	dimension := c.DefaultQuery("dimension", "employee")
	rows, dim := h.lostCustomerRows(c)

	type agg struct {
		Lost, Back int
		BackAmount float64
	}
	groups := map[string]*agg{}
	names := map[string]string{}
	order := []string{}
	for _, r := range rows {
		key, name := dim.dimKeyName(dimension, r.OwnerID, r.RegionID, r.CategoryID)
		a, ok := groups[key]
		if !ok {
			a = &agg{}
			groups[key] = a
			names[key] = name
			order = append(order, key)
		}
		a.Lost++
		if r.BackDate != nil {
			a.Back++
			a.BackAmount += *r.BackAmount
		}
	}
	list := make([]gin.H, 0, len(groups))
	for _, key := range order {
		a := groups[key]
		backRate := 0.0
		if a.Lost > 0 {
			backRate = float64(a.Back) / float64(a.Lost) * 100
		}
		list = append(list, gin.H{
			"dimName": names[key], "lost": a.Lost, "back": a.Back,
			"backRate": backRate, "backAmount": a.BackAmount,
		})
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

func (h *BIHandler) LostCustomerDetail(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	keyword := c.Query("keyword")
	onlyBack := c.DefaultQuery("onlyBack", "")
	rows, _ := h.lostCustomerRows(c)
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		if keyword != "" && !containsAny(r.Customer+r.Code, keyword) {
			continue
		}
		isBack := r.BackDate != nil
		if onlyBack == "1" && !isBack {
			continue
		}
		list = append(list, gin.H{
			"customer": r.Customer, "code": r.Code, "lastBill": r.LastBill,
			"isBack": isBack, "backDate": r.BackDate,
			"backAmount": r.BackAmount, "backBills": r.BackBills,
		})
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 商品铺市 ====================

func (h *BIHandler) ProductDistSummary(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	categoryID, _ := strconv.Atoi(c.DefaultQuery("categoryId", "0"))
	brandID, _ := strconv.Atoi(c.DefaultQuery("brandId", "0"))

	// 每商品×客户首拿日期
	type firstRow struct {
		ProductID  uint   `gorm:"column:product_id"`
		CustomerID uint   `gorm:"column:customer_id"`
		FirstDate  string `gorm:"column:first_date"`
	}
	var firsts []firstRow
	h.db.Raw(`SELECT i.product_id, b.customer_id, MIN(b.bill_date)::text AS first_date
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND b.bill_date <= ?
		GROUP BY i.product_id, b.customer_id`, companyID, end).Scan(&firsts)

	// 本期铺市（非首拿也算）
	type periodRow struct {
		ProductID uint `gorm:"column:product_id"`
		Cnt       int  `gorm:"column:cnt"`
	}
	periodCnt := map[uint]int{}
	var prs []periodRow
	h.db.Raw(`SELECT i.product_id, COUNT(DISTINCT b.customer_id)::int AS cnt
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
			AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY i.product_id`, companyID, start, end).Scan(&prs)
	for _, r := range prs {
		periodCnt[r.ProductID] = r.Cnt
	}

	type prodAgg struct {
		Total, Initial, NewIn, Period int
	}
	aggMap := map[uint]*prodAgg{}
	for _, f := range firsts {
		a, ok := aggMap[f.ProductID]
		if !ok {
			a = &prodAgg{}
			aggMap[f.ProductID] = a
		}
		a.Total++
		if f.FirstDate < start {
			a.Initial++
		} else if f.FirstDate >= start && f.FirstDate <= end {
			a.NewIn++
		}
	}
	for pid, a := range aggMap {
		a.Period = periodCnt[pid]
	}

	// 商品信息 + 过滤
	var products []model.Product
	q := h.db.Where("company_id = ? AND deleted_at IS NULL", companyID)
	if keyword != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ? OR specification ILIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	if brandID > 0 {
		q = q.Where("brand_id = ?", brandID)
	}
	q.Find(&products)

	list := make([]gin.H, 0, len(products))
	for _, p := range products {
		a := aggMap[p.ID]
		if a == nil {
			a = &prodAgg{}
		}
		list = append(list, gin.H{
			"productId": p.ID, "product": p.Name, "productCode": p.Code,
			"spec": p.Specification, "unit": p.Unit,
			"total": a.Total, "initial": a.Initial, "newIn": a.NewIn, "period": a.Period,
		})
	}
	sortByFloatDesc(list, func(r gin.H) float64 { return float64(r["total"].(int)) })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

func (h *BIHandler) ProductDistDetail(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	productKeyword := c.Query("productKeyword")

	type row struct {
		ProductID   uint    `gorm:"column:product_id"`
		Product     string  `gorm:"column:product"`
		ProductCode string  `gorm:"column:product_code"`
		Spec        string  `gorm:"column:spec"`
		CustomerID  uint    `gorm:"column:customer_id"`
		Customer    string  `gorm:"column:customer"`
		Code        string  `gorm:"column:code"`
		FirstDate   string  `gorm:"column:first_date"`
		Qty         float64 `gorm:"column:qty"`
		Amount      float64 `gorm:"column:amount"`
	}
	var rows []row
	h.db.Raw(`WITH firsts AS (
		SELECT i.product_id, b.customer_id, MIN(b.bill_date) AS first_date
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND b.bill_date <= ?
		GROUP BY i.product_id, b.customer_id
	),
	period AS (
		SELECT i.product_id, b.customer_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amount
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
			AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY i.product_id, b.customer_id
	)
	SELECT f.product_id, p.name AS product, p.code AS product_code, p.specification AS spec,
		f.customer_id, cu.name AS customer, cu.code,
		f.first_date::text AS first_date, COALESCE(pd.qty,0) AS qty, COALESCE(pd.amount,0) AS amount
	FROM firsts f
	JOIN products p ON p.id = f.product_id
	JOIN customers cu ON cu.id = f.customer_id
	LEFT JOIN period pd ON pd.product_id = f.product_id AND pd.customer_id = f.customer_id
	WHERE pd.customer_id IS NOT NULL`, companyID, end, companyID, start, end).Scan(&rows)

	list := make([]row, 0, len(rows))
	for _, r := range rows {
		if keyword != "" && !containsAny(r.Customer+r.Code, keyword) {
			continue
		}
		if productKeyword != "" && !containsAny(r.Product+r.ProductCode+r.Spec, productKeyword) {
			continue
		}
		list = append(list, r)
	}
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}
