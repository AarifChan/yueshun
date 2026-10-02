package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/pkg/response"
)

// InventoryReportHandler 库存报表处理器
type InventoryReportHandler struct {
	db *gorm.DB
}

func NewInventoryReportHandler(db *gorm.DB) *InventoryReportHandler {
	return &InventoryReportHandler{db: db}
}

func (h *InventoryReportHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/inventory-reports")
	{
		g.GET("/stock-distribution", h.StockDistribution)
		g.GET("/inout-summary", h.InoutSummary)
		g.GET("/transfer-summary", h.TransferSummary)
		g.GET("/other-inout-stats", h.OtherInoutStats)
		g.GET("/batch-stats", h.BatchStats)
		g.GET("/batch-trace", h.BatchTrace)
		g.GET("/expiry-warning", h.ExpiryWarning)
		g.GET("/inout-flow", h.InoutFlow)
	}
}

type reportReq struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"pageSize,default=30"`
	Keyword   string `form:"keyword"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
}

func dateRangeWhere(column, start, end string) (string, []interface{}) {
	clause := ""
	var args []interface{}
	if start != "" {
		clause += fmt.Sprintf(" AND %s >= ?", column)
		args = append(args, start+" 00:00:00")
	}
	if end != "" {
		clause += fmt.Sprintf(" AND %s <= ?", column)
		args = append(args, end+" 23:59:59")
	}
	return clause, args
}

// StockDistribution 库存分布表
// @Summary 库存分布表（各仓库数量列动态展开）
// @Tags 库存报表
// @Router /api/v1/inventory-reports/stock-distribution [get]
func (h *InventoryReportHandler) StockDistribution(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)

	// 仓库列
	type whCol struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	var warehouses []whCol
	h.db.Model(&struct {
		ID   uint
		Name string
	}{}).Raw(`SELECT id, name FROM warehouses WHERE company_id = ? AND deleted_at IS NULL ORDER BY id`, companyID).Scan(&warehouses)

	type row struct {
		ProductID     uint    `json:"productId"`
		ProductName   string  `json:"productName"`
		Code          string  `json:"code"`
		Specification string  `json:"specification"`
		Unit          string  `json:"unit"`
		TotalQty      float64 `json:"totalQty"`
		CostPrice     float64 `json:"costPrice"`
		Amount        float64 `json:"amount"`
	}
	base := `FROM products p
		LEFT JOIN (SELECT product_id, SUM(quantity) AS total_qty,
			CASE WHEN SUM(quantity) <> 0 THEN SUM(amount) / SUM(quantity) ELSE 0 END AS cost_price,
			SUM(amount) AS amount
			FROM stocks WHERE company_id = ? AND deleted_at IS NULL GROUP BY product_id) s ON s.product_id = p.id
		WHERE p.company_id = ? AND p.deleted_at IS NULL`
	args := []interface{}{companyID, companyID}
	if req.Keyword != "" {
		base += " AND (p.name LIKE ? OR p.code LIKE ? OR p.barcode LIKE ? OR p.specification LIKE ?)"
		kw := "%" + req.Keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}
	var total int64
	h.db.Raw("SELECT COUNT(*) "+base, args...).Scan(&total)

	var list []row
	h.db.Raw(`SELECT p.id AS product_id, p.name AS product_name, p.code, p.specification, p.unit,
		COALESCE(s.total_qty,0) AS total_qty, COALESCE(s.cost_price,0) AS cost_price, COALESCE(s.amount,0) AS amount `+base+
		" ORDER BY p.id LIMIT ? OFFSET ?", append(args, req.PageSize, (req.Page-1)*req.PageSize)...).Scan(&list)

	// 各仓库数量
	type whQty struct {
		ProductID uint    `json:"productId"`
		Qty       float64 `json:"qty"`
		Warehouse string  `json:"warehouse"`
	}
	var qtys []whQty
	if len(list) > 0 {
		ids := make([]uint, 0, len(list))
		for _, r := range list {
			ids = append(ids, r.ProductID)
		}
		h.db.Raw(`SELECT s.product_id, s.quantity AS qty, w.name AS warehouse
			FROM stocks s JOIN warehouses w ON w.id = s.warehouse_id
			WHERE s.company_id = ? AND s.product_id IN ? AND s.deleted_at IS NULL`, companyID, ids).Scan(&qtys)
	}
	type outRow struct {
		row
		Warehouses map[string]float64 `json:"warehouses"`
	}
	out := make([]outRow, 0, len(list))
	qtyMap := map[uint]map[string]float64{}
	for _, q := range qtys {
		if qtyMap[q.ProductID] == nil {
			qtyMap[q.ProductID] = map[string]float64{}
		}
		qtyMap[q.ProductID][q.Warehouse] = q.Qty
	}
	for _, r := range list {
		out = append(out, outRow{row: r, Warehouses: qtyMap[r.ProductID]})
	}
	response.Ok(c, gin.H{"list": out, "total": total, "page": req.Page, "pageSize": req.PageSize, "warehouses": warehouses})
}

// InoutSummary 商品进销存汇总
// @Summary 商品进销存汇总（期初/本期入库按单据类型/本期出库按单据类型/结存）
// @Tags 库存报表
// @Router /api/v1/inventory-reports/inout-summary [get]
func (h *InventoryReportHandler) InoutSummary(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)

	// 入库流水（单据类型, 商品, 数量, 金额, 时间）
	inFlows := []struct{ typeName, sql string }{
		{"采购入库单", `SELECT '采购入库单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM purchase_in_stock_items i JOIN purchase_in_stocks b ON b.id = i.in_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"销售退货单", `SELECT '销售退货单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM sales_return_items i JOIN sales_returns b ON b.id = i.return_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"其他入库单", `SELECT '其他入库单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM other_in_stock_items i JOIN other_in_stocks b ON b.id = i.in_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"同价调拨单", `SELECT '同价调拨单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM transfer_in_items i JOIN transfer_ins b ON b.id = i.in_id AND b.status = 'completed' AND b.transfer_type = 'same' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"变价调拨单", `SELECT '变价调拨单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM transfer_in_items i JOIN transfer_ins b ON b.id = i.in_id AND b.status = 'completed' AND b.transfer_type = 'diff' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"同价调拨单(旧)", `SELECT '同价调拨单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM inventory_transfer_items i JOIN inventory_transfers b ON b.id = i.transfer_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"组装拆装单", `SELECT '组装拆装单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM assembly_order_items i JOIN assembly_orders b ON b.id = i.order_id AND b.status = 'completed' AND i.direction = 'in' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
	}
	outFlows := []struct{ typeName, sql string }{
		{"销售出库单", `SELECT '销售出库单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM sales_out_stock_items i JOIN sales_out_stocks b ON b.id = i.out_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"采购退货单", `SELECT '采购退货单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM purchase_return_items i JOIN purchase_returns b ON b.id = i.return_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"其他出库单", `SELECT '其他出库单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM other_out_stock_items i JOIN other_out_stocks b ON b.id = i.out_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"同价调拨单", `SELECT '同价调拨单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM transfer_out_items i JOIN transfer_outs b ON b.id = i.out_id AND b.status = 'completed' AND b.transfer_type = 'same' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"变价调拨单", `SELECT '变价调拨单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM transfer_out_items i JOIN transfer_outs b ON b.id = i.out_id AND b.status = 'completed' AND b.transfer_type = 'diff' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"同价调拨单(旧)", `SELECT '同价调拨单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM inventory_transfer_items i JOIN inventory_transfers b ON b.id = i.transfer_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
		{"组装拆装单", `SELECT '组装拆装单' AS t, i.product_id, i.quantity AS qty, i.amount AS amt, b.created_at FROM assembly_order_items i JOIN assembly_orders b ON b.id = i.order_id AND b.status = 'completed' AND i.direction = 'out' AND b.deleted_at IS NULL WHERE b.company_id = ?`},
	}

	type flow struct {
		T         string  `gorm:"column:t"`
		ProductID uint    `gorm:"column:product_id"`
		Qty       float64 `gorm:"column:qty"`
		Amt       float64 `gorm:"column:amt"`
	}
	var inRows, outRows []flow
	for _, f := range inFlows {
		var rows []flow
		h.db.Raw(f.sql, companyID).Scan(&rows)
		inRows = append(inRows, rows...)
	}
	for _, f := range outFlows {
		var rows []flow
		h.db.Raw(f.sql, companyID).Scan(&rows)
		outRows = append(outRows, rows...)
	}

	// 产品清单（分页）
	type prod struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		Code          string `json:"code"`
		Specification string `json:"specification"`
		Unit          string `json:"unit"`
	}
	pq := h.db.Table("products").Where("company_id = ? AND deleted_at IS NULL", companyID)
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		pq = pq.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ? OR specification LIKE ?", kw, kw, kw, kw)
	}
	var total int64
	pq.Count(&total)
	var prods []prod
	pq.Select("id, name, code, specification, unit").Order("id").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Scan(&prods)
	prodSet := map[uint]bool{}
	for _, p := range prods {
		prodSet[p.ID] = true
	}

	type bucket struct{ Qty, Amt float64 }
	type agg struct {
		PriorIn, PriorOut             bucket
		InByType, OutByType           map[string]*bucket
		PeriodIn, PeriodOut, Balance  bucket
	}
	aggs := map[uint]*agg{}
	getAgg := func(pid uint) *agg {
		if aggs[pid] == nil {
			aggs[pid] = &agg{InByType: map[string]*bucket{}, OutByType: map[string]*bucket{}}
		}
		return aggs[pid]
	}
	startTs := req.StartDate + " 00:00:00"
	endTs := req.EndDate + " 23:59:59"
	_ = startTs

	// 需要 created_at 的原始行：重查带时间版本
	type flowT struct {
		T         string  `gorm:"column:t"`
		ProductID uint    `gorm:"column:product_id"`
		Qty       float64 `gorm:"column:qty"`
		Amt       float64 `gorm:"column:amt"`
		CreatedAt string  `gorm:"column:created_at"`
	}
	collect := func(sqls []struct{ typeName, sql string }, in bool) {
		for _, f := range sqls {
			var rows []flowT
			h.db.Raw(f.sql, companyID).Scan(&rows)
			for _, r := range rows {
				if !prodSet[r.ProductID] {
					continue
				}
				a := getAgg(r.ProductID)
				ts := r.CreatedAt
				if req.StartDate != "" && ts < startTs {
					if in {
						a.PriorIn.Qty += r.Qty
						a.PriorIn.Amt += r.Amt
					} else {
						a.PriorOut.Qty += r.Qty
						a.PriorOut.Amt += r.Amt
					}
					continue
				}
				if req.EndDate != "" && ts > endTs {
					continue
				}
				m := a.InByType
				if !in {
					m = a.OutByType
				}
				if m[r.T] == nil {
					m[r.T] = &bucket{}
				}
				m[r.T].Qty += r.Qty
				m[r.T].Amt += r.Amt
				if in {
					a.PeriodIn.Qty += r.Qty
					a.PeriodIn.Amt += r.Amt
				} else {
					a.PeriodOut.Qty += r.Qty
					a.PeriodOut.Amt += r.Amt
				}
			}
		}
	}
	collect(inFlows, true)
	collect(outFlows, false)

	type row struct {
		prod
		PriorQty   float64            `json:"priorQty"`
		PriorAmt   float64            `json:"priorAmt"`
		InByType   map[string]*bucket `json:"inByType"`
		OutByType  map[string]*bucket `json:"outByType"`
		PeriodInQty  float64          `json:"periodInQty"`
		PeriodInAmt  float64          `json:"periodInAmt"`
		PeriodOutQty float64          `json:"periodOutQty"`
		PeriodOutAmt float64          `json:"periodOutAmt"`
		BalanceQty float64            `json:"balanceQty"`
		BalanceAmt float64            `json:"balanceAmt"`
	}
	var out []row
	for _, p := range prods {
		a := getAgg(p.ID)
		out = append(out, row{
			prod:         p,
			PriorQty:     a.PriorIn.Qty - a.PriorOut.Qty,
			PriorAmt:     a.PriorIn.Amt - a.PriorOut.Amt,
			InByType:     a.InByType,
			OutByType:    a.OutByType,
			PeriodInQty:  a.PeriodIn.Qty,
			PeriodInAmt:  a.PeriodIn.Amt,
			PeriodOutQty: a.PeriodOut.Qty,
			PeriodOutAmt: a.PeriodOut.Amt,
			BalanceQty:   a.PriorIn.Qty - a.PriorOut.Qty + a.PeriodIn.Qty - a.PeriodOut.Qty,
			BalanceAmt:   a.PriorIn.Amt - a.PriorOut.Amt + a.PeriodIn.Amt - a.PeriodOut.Amt,
		})
	}
	response.Ok(c, gin.H{"list": out, "total": total, "page": req.Page, "pageSize": req.PageSize})
}

// TransferSummary 调拨汇总表
// @Summary 调拨汇总表
// @Tags 库存报表
// @Router /api/v1/inventory-reports/transfer-summary [get]
func (h *InventoryReportHandler) TransferSummary(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)
	transferType := c.Query("transferType")
	fromWarehouseID, _ := strconv.Atoi(c.DefaultQuery("fromWarehouseId", "0"))
	toWarehouseID, _ := strconv.Atoi(c.DefaultQuery("toWarehouseId", "0"))

	outWhere := "b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL"
	outArgs := []interface{}{companyID}
	if transferType != "" {
		outWhere += " AND b.transfer_type = ?"
		outArgs = append(outArgs, transferType)
	}
	if fromWarehouseID > 0 {
		outWhere += " AND b.from_warehouse_id = ?"
		outArgs = append(outArgs, fromWarehouseID)
	}
	if toWarehouseID > 0 {
		outWhere += " AND b.to_warehouse_id = ?"
		outArgs = append(outArgs, toWarehouseID)
	}
	dw, dargs := dateRangeWhere("b.created_at", req.StartDate, req.EndDate)
	outWhere += dw
	outArgs = append(outArgs, dargs...)

	type row struct {
		ProductID     uint    `json:"productId"`
		ProductName   string  `json:"productName"`
		Specification string  `json:"specification"`
		Unit          string  `json:"unit"`
		OutQty        float64 `json:"outQty"`
		InQty         float64 `json:"inQty"`
		OutCostPrice  float64 `json:"outCostPrice"`
		OutCostAmt    float64 `json:"outCostAmt"`
		InPrice       float64 `json:"inPrice"`
		InAmt         float64 `json:"inAmt"`
		DiffAmt       float64 `json:"diffAmt"`
		Code          string  `json:"code"`
	}
	base := fmt.Sprintf(`FROM products p
		LEFT JOIN (SELECT i.product_id, SUM(i.quantity) AS out_qty, SUM(i.amount) AS out_amt,
			CASE WHEN SUM(i.quantity) <> 0 THEN SUM(i.amount)/SUM(i.quantity) ELSE 0 END AS out_price
			FROM transfer_out_items i JOIN transfer_outs b ON b.id = i.out_id WHERE %s GROUP BY i.product_id) o ON o.product_id = p.id
		LEFT JOIN (SELECT i.product_id, SUM(i.quantity) AS in_qty, SUM(i.amount) AS in_amt,
			CASE WHEN SUM(i.quantity) <> 0 THEN SUM(i.amount)/SUM(i.quantity) ELSE 0 END AS in_price
			FROM transfer_in_items i JOIN transfer_ins b ON b.id = i.in_id WHERE %s GROUP BY i.product_id) n ON n.product_id = p.id
		WHERE p.company_id = ? AND p.deleted_at IS NULL AND (o.out_qty IS NOT NULL OR n.in_qty IS NOT NULL)`, outWhere, outWhere)
	args := append(append([]interface{}{}, outArgs...), outArgs...)
	args = append(args, companyID)
	if req.Keyword != "" {
		base += " AND (p.name LIKE ? OR p.code LIKE ? OR p.specification LIKE ?)"
		kw := "%" + req.Keyword + "%"
		args = append(args, kw, kw, kw)
	}
	var total int64
	h.db.Raw("SELECT COUNT(*) "+base, args...).Scan(&total)
	var list []row
	h.db.Raw(`SELECT p.id AS product_id, p.name AS product_name, p.specification, p.unit, p.code,
		COALESCE(o.out_qty,0) AS out_qty, COALESCE(n.in_qty,0) AS in_qty,
		COALESCE(o.out_price,0) AS out_cost_price, COALESCE(o.out_amt,0) AS out_cost_amt,
		COALESCE(n.in_price,0) AS in_price, COALESCE(n.in_amt,0) AS in_amt,
		COALESCE(n.in_amt,0) - COALESCE(o.out_amt,0) AS diff_amt `+base+" ORDER BY p.id LIMIT ? OFFSET ?",
		append(args, req.PageSize, (req.Page-1)*req.PageSize)...).Scan(&list)
	response.Ok(c, gin.H{"list": list, "total": total, "page": req.Page, "pageSize": req.PageSize})
}

// OtherInoutStats 其他出入库统计
// @Summary 其他出入库统计
// @Tags 库存报表
// @Router /api/v1/inventory-reports/other-inout-stats [get]
func (h *InventoryReportHandler) OtherInoutStats(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)
	warehouseID, _ := strconv.Atoi(c.DefaultQuery("warehouseId", "0"))

	inWhere := "b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL"
	outWhere := "b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL"
	args := []interface{}{companyID}
	if warehouseID > 0 {
		inWhere += " AND b.warehouse_id = ?"
		outWhere += " AND b.warehouse_id = ?"
		args = append(args, warehouseID)
	}
	dw, dargs := dateRangeWhere("b.created_at", req.StartDate, req.EndDate)
	inWhere += dw
	outWhere += dw
	args = append(args, dargs...)

	type row struct {
		ProductID   uint    `json:"productId"`
		Code        string  `json:"code"`
		ProductName string  `json:"productName"`
		Unit        string  `json:"unit"`
		InQty       float64 `json:"inQty"`
		InPrice     float64 `json:"inPrice"`
		InAmt       float64 `json:"inAmt"`
		OutQty      float64 `json:"outQty"`
		OutPrice    float64 `json:"outPrice"`
		OutAmt      float64 `json:"outAmt"`
	}
	inSQL := fmt.Sprintf(`SELECT i.product_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amt,
		CASE WHEN SUM(i.quantity) <> 0 THEN SUM(i.amount)/SUM(i.quantity) ELSE 0 END AS price
		FROM other_in_stock_items i JOIN other_in_stocks b ON b.id = i.in_stock_id WHERE %s GROUP BY i.product_id`, inWhere)
	outSQL := fmt.Sprintf(`SELECT i.product_id, SUM(i.quantity) AS qty, SUM(i.amount) AS amt,
		CASE WHEN SUM(i.quantity) <> 0 THEN SUM(i.amount)/SUM(i.quantity) ELSE 0 END AS price
		FROM other_out_stock_items i JOIN other_out_stocks b ON b.id = i.out_stock_id WHERE %s GROUP BY i.product_id`, outWhere)
	type qa struct {
		ProductID uint    `gorm:"column:product_id"`
		Qty       float64 `gorm:"column:qty"`
		Amt       float64 `gorm:"column:amt"`
		Price     float64 `gorm:"column:price"`
	}
	var inRows, outRows []qa
	h.db.Raw(inSQL, args...).Scan(&inRows)
	h.db.Raw(outSQL, args...).Scan(&outRows)
	inMap := map[uint]qa{}
	for _, r := range inRows {
		inMap[r.ProductID] = r
	}
	outMap := map[uint]qa{}
	for _, r := range outRows {
		outMap[r.ProductID] = r
	}
	pidSet := map[uint]bool{}
	for pid := range inMap {
		pidSet[pid] = true
	}
	for pid := range outMap {
		pidSet[pid] = true
	}
	pids := make([]uint, 0, len(pidSet))
	for pid := range pidSet {
		pids = append(pids, pid)
	}
	type prod struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Code string `gorm:"column:code"`
		Unit string `gorm:"column:unit"`
	}
	var prods []prod
	if len(pids) > 0 {
		pq := h.db.Table("products").Where("id IN ? AND company_id = ?", pids, companyID)
		if req.Keyword != "" {
			kw := "%" + req.Keyword + "%"
			pq = pq.Where("name LIKE ? OR code LIKE ?", kw, kw)
		}
		pq.Select("id, name, code, unit").Order("id").Scan(&prods)
	}
	total := len(prods)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	var list []row
	for _, p := range prods[start:end] {
		i := inMap[p.ID]
		o := outMap[p.ID]
		list = append(list, row{
			ProductID: p.ID, Code: p.Code, ProductName: p.Name, Unit: p.Unit,
			InQty: i.Qty, InPrice: i.Price, InAmt: i.Amt,
			OutQty: o.Qty, OutPrice: o.Price, OutAmt: o.Amt,
		})
	}
	response.Ok(c, gin.H{"list": list, "total": total, "page": req.Page, "pageSize": req.PageSize})
}

// BatchStats 库存批号统计
// @Summary 库存批号统计
// @Tags 库存报表
// @Router /api/v1/inventory-reports/batch-stats [get]
func (h *InventoryReportHandler) BatchStats(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)
	warehouseID, _ := strconv.Atoi(c.DefaultQuery("warehouseId", "0"))
	batchNo := c.Query("batchNo")
	onlyNoBatch := c.Query("onlyNoBatch") == "true"

	q := h.db.Table("stock_batches b").
		Select(`b.id, p.name AS product_name, p.specification, p.unit, b.batch_no, b.quantity, b.cost_price,
			(b.quantity * b.cost_price) AS cost_amount, b.produce_date, b.expiry_date, b.shelf_life_days, b.in_date,
			COALESCE(w.name,'') AS warehouse_name, COALESCE(s.name,'') AS supplier_name, p.code`).
		Joins("JOIN products p ON p.id = b.product_id").
		Joins("LEFT JOIN warehouses w ON w.id = b.warehouse_id").
		Joins("LEFT JOIN suppliers s ON s.id = b.supplier_id").
		Where("b.company_id = ?", companyID)
	if warehouseID > 0 {
		q = q.Where("b.warehouse_id = ?", warehouseID)
	}
	if batchNo != "" {
		q = q.Where("b.batch_no LIKE ?", "%"+batchNo+"%")
	}
	if onlyNoBatch {
		q = q.Where("b.batch_no = ''")
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		q = q.Where("(p.name LIKE ? OR p.code LIKE ?)", kw, kw)
	}
	var total int64
	q.Count(&total)
	type row struct {
		ID            uint    `json:"id"`
		ProductName   string  `json:"productName"`
		Specification string  `json:"specification"`
		Unit          string  `json:"unit"`
		BatchNo       string  `json:"batchNo"`
		Quantity      float64 `json:"quantity"`
		CostPrice     float64 `json:"costPrice"`
		CostAmount    float64 `json:"costAmount"`
		ProduceDate   *string `json:"produceDate"`
		ExpiryDate    *string `json:"expiryDate"`
		ShelfLifeDays int     `json:"shelfLifeDays"`
		InDate        *string `json:"inDate"`
		WarehouseName string  `json:"warehouseName"`
		SupplierName  string  `json:"supplierName"`
		Code          string  `json:"code"`
	}
	var list []row
	q.Order("b.id DESC").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Scan(&list)
	response.Ok(c, gin.H{"list": list, "total": total, "page": req.Page, "pageSize": req.PageSize})
}

// BatchTrace 批号跟踪详情（其他入库单批次行）
// @Summary 批号跟踪详情
// @Tags 库存报表
// @Router /api/v1/inventory-reports/batch-trace [get]
func (h *InventoryReportHandler) BatchTrace(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)
	batchNo := c.Query("batchNo")
	warehouseID, _ := strconv.Atoi(c.DefaultQuery("warehouseId", "0"))

	q := h.db.Table("other_in_stock_items i").
		Select(`'其他入库单' AS bill_type, b.counterpart AS counterpart, COALESCE(e.name,'') AS handler_name,
			p.name AS product_name, p.specification, p.unit, COALESCE(w.name,'') AS warehouse_name,
			i.batch_no, i.produce_date, b.bill_date AS in_date, i.shelf_life_days,
			'已过账' AS status, i.quantity, i.price AS cost_price, i.amount AS cost_amount,
			b.bill_no, b.created_at`).
		Joins("JOIN other_in_stocks b ON b.id = i.in_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL").
		Joins("JOIN products p ON p.id = i.product_id").
		Joins("LEFT JOIN warehouses w ON w.id = b.warehouse_id").
		Joins("LEFT JOIN employees e ON e.id = b.handler_id").
		Where("b.company_id = ? AND i.batch_no != ''", companyID)
	if batchNo != "" {
		q = q.Where("i.batch_no LIKE ?", "%"+batchNo+"%")
	}
	if warehouseID > 0 {
		q = q.Where("b.warehouse_id = ?", warehouseID)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		q = q.Where("(p.name LIKE ? OR p.code LIKE ?)", kw, kw)
	}
	if req.StartDate != "" {
		q = q.Where("b.created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		q = q.Where("b.created_at <= ?", req.EndDate+" 23:59:59")
	}
	var total int64
	q.Count(&total)
	type row struct {
		BillType    string  `json:"billType"`
		Counterpart string  `json:"counterpart"`
		HandlerName string  `json:"handlerName"`
		ProductName string  `json:"productName"`
		Specification string `json:"specification"`
		Unit        string  `json:"unit"`
		WarehouseName string `json:"warehouseName"`
		BatchNo     string  `json:"batchNo"`
		ProduceDate *string `json:"produceDate"`
		InDate      string  `json:"inDate"`
		ShelfLifeDays int   `json:"shelfLifeDays"`
		Status      string  `json:"status"`
		Quantity    float64 `json:"quantity"`
		CostPrice   float64 `json:"costPrice"`
		CostAmount  float64 `json:"costAmount"`
		BillNo      string  `json:"billNo"`
		CreatedAt   string  `json:"createdAt"`
	}
	var list []row
	q.Order("b.created_at DESC").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Scan(&list)
	response.Ok(c, gin.H{"list": list, "total": total, "page": req.Page, "pageSize": req.PageSize})
}

// ExpiryWarning 商品近效期预警
// @Summary 商品近效期预警
// @Tags 库存报表
// @Param days query int false "预警天数 默认90"
// @Router /api/v1/inventory-reports/expiry-warning [get]
func (h *InventoryReportHandler) ExpiryWarning(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
	warehouseID, _ := strconv.Atoi(c.DefaultQuery("warehouseId", "0"))

	q := h.db.Table("stock_batches b").
		Select(`b.id, p.name AS product_name, p.specification, b.batch_no, b.produce_date, b.shelf_life_days, b.expiry_date,
			(b.expiry_date::date - CURRENT_DATE) AS remain_days,
			CASE WHEN b.expiry_date::date < CURRENT_DATE THEN (CURRENT_DATE - b.expiry_date::date) ELSE 0 END AS expired_days,
			p.unit, COALESCE(w.name,'') AS warehouse_name, b.quantity, b.cost_price, (b.quantity*b.cost_price) AS cost_amount,
			COALESCE(s.name,'') AS supplier_name, p.code`).
		Joins("JOIN products p ON p.id = b.product_id").
		Joins("LEFT JOIN warehouses w ON w.id = b.warehouse_id").
		Joins("LEFT JOIN suppliers s ON s.id = b.supplier_id").
		Where("b.company_id = ? AND b.expiry_date IS NOT NULL AND b.quantity > 0", companyID).
		Where("b.expiry_date::date <= CURRENT_DATE + ?", days)
	if warehouseID > 0 {
		q = q.Where("b.warehouse_id = ?", warehouseID)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		q = q.Where("(p.name LIKE ? OR p.code LIKE ? OR b.batch_no LIKE ?)", kw, kw, kw)
	}
	var total int64
	q.Count(&total)
	type row struct {
		ID            uint    `json:"id"`
		ProductName   string  `json:"productName"`
		Specification string  `json:"specification"`
		BatchNo       string  `json:"batchNo"`
		ProduceDate   *string `json:"produceDate"`
		ShelfLifeDays int     `json:"shelfLifeDays"`
		ExpiryDate    *string `json:"expiryDate"`
		RemainDays    int     `json:"remainDays"`
		ExpiredDays   int     `json:"expiredDays"`
		Unit          string  `json:"unit"`
		WarehouseName string  `json:"warehouseName"`
		Quantity      float64 `json:"quantity"`
		CostPrice     float64 `json:"costPrice"`
		CostAmount    float64 `json:"costAmount"`
		SupplierName  string  `json:"supplierName"`
		Code          string  `json:"code"`
	}
	var list []row
	q.Order("b.expiry_date").Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Scan(&list)
	response.Ok(c, gin.H{"list": list, "total": total, "page": req.Page, "pageSize": req.PageSize})
}

// InoutFlow 商品出入库流水
// @Summary 商品出入库流水（行级，跨单据类型）
// @Tags 库存报表
// @Param billType query string false "单据类型"
// @Router /api/v1/inventory-reports/inout-flow [get]
func (h *InventoryReportHandler) InoutFlow(c *gin.Context) {
	var req reportReq
	_ = c.ShouldBindQuery(&req)
	companyID := middleware.GetCompanyID(c)
	billType := c.Query("billType")
	billNo := c.Query("billNo")

	flows := []struct {
		typeName, sql string
		in            bool
	}{
		{"采购入库单", `SELECT '采购入库单' AS bill_type, i.product_id, NULL AS batch_no, i.quantity AS in_qty, i.amount AS in_amt, 0 AS out_qty, 0 AS out_amt, b.bill_no, b.created_at FROM purchase_in_stock_items i JOIN purchase_in_stocks b ON b.id = i.in_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, true},
		{"销售退货单", `SELECT '销售退货单' AS bill_type, i.product_id, NULL AS batch_no, i.quantity AS in_qty, i.amount AS in_amt, 0 AS out_qty, 0 AS out_amt, b.bill_no, b.created_at FROM sales_return_items i JOIN sales_returns b ON b.id = i.return_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, true},
		{"其他入库单", `SELECT '其他入库单' AS bill_type, i.product_id, i.batch_no, i.quantity AS in_qty, i.amount AS in_amt, 0 AS out_qty, 0 AS out_amt, b.bill_no, b.created_at FROM other_in_stock_items i JOIN other_in_stocks b ON b.id = i.in_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, true},
		{"调拨入库单", `SELECT '调拨入库单' AS bill_type, i.product_id, NULL AS batch_no, i.quantity AS in_qty, i.amount AS in_amt, 0 AS out_qty, 0 AS out_amt, b.bill_no, b.created_at FROM transfer_in_items i JOIN transfer_ins b ON b.id = i.in_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, true},
		{"组装拆装单(入)", `SELECT '组装拆装单' AS bill_type, i.product_id, NULL AS batch_no, i.quantity AS in_qty, i.amount AS in_amt, 0 AS out_qty, 0 AS out_amt, b.bill_no, b.created_at FROM assembly_order_items i JOIN assembly_orders b ON b.id = i.order_id AND b.status = 'completed' AND i.direction = 'in' AND b.deleted_at IS NULL WHERE b.company_id = ?`, true},
		{"销售出库单", `SELECT '销售出库单' AS bill_type, i.product_id, NULLIF(i.batch_no,'') AS batch_no, 0 AS in_qty, 0 AS in_amt, i.quantity AS out_qty, i.amount AS out_amt, b.bill_no, b.created_at FROM sales_out_stock_items i JOIN sales_out_stocks b ON b.id = i.out_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, false},
		{"采购退货单", `SELECT '采购退货单' AS bill_type, i.product_id, NULL AS batch_no, 0 AS in_qty, 0 AS in_amt, i.quantity AS out_qty, i.amount AS out_amt, b.bill_no, b.created_at FROM purchase_return_items i JOIN purchase_returns b ON b.id = i.return_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, false},
		{"其他出库单", `SELECT '其他出库单' AS bill_type, i.product_id, NULL AS batch_no, 0 AS in_qty, 0 AS in_amt, i.quantity AS out_qty, i.amount AS out_amt, b.bill_no, b.created_at FROM other_out_stock_items i JOIN other_out_stocks b ON b.id = i.out_stock_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, false},
		{"调拨出库单", `SELECT '调拨出库单' AS bill_type, i.product_id, NULL AS batch_no, 0 AS in_qty, 0 AS in_amt, i.quantity AS out_qty, i.amount AS out_amt, b.bill_no, b.created_at FROM transfer_out_items i JOIN transfer_outs b ON b.id = i.out_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, false},
		{"库存调拨单(出)", `SELECT '同价调拨单' AS bill_type, i.product_id, NULL AS batch_no, 0 AS in_qty, 0 AS in_amt, i.quantity AS out_qty, i.amount AS out_amt, b.bill_no, b.created_at FROM inventory_transfer_items i JOIN inventory_transfers b ON b.id = i.transfer_id AND b.status = 'completed' AND b.deleted_at IS NULL WHERE b.company_id = ?`, false},
		{"组装拆装单(出)", `SELECT '组装拆装单' AS bill_type, i.product_id, NULL AS batch_no, 0 AS in_qty, 0 AS in_amt, i.quantity AS out_qty, i.amount AS out_amt, b.bill_no, b.created_at FROM assembly_order_items i JOIN assembly_orders b ON b.id = i.order_id AND b.status = 'completed' AND i.direction = 'out' AND b.deleted_at IS NULL WHERE b.company_id = ?`, false},
	}

	type flowRow struct {
		BillType  string   `gorm:"column:bill_type" json:"billType"`
		ProductID uint     `gorm:"column:product_id" json:"productId"`
		BatchNo   *string  `gorm:"column:batch_no" json:"batchNo"`
		InQty     float64  `gorm:"column:in_qty" json:"inQty"`
		InAmt     float64  `gorm:"column:in_amt" json:"inAmt"`
		OutQty    float64  `gorm:"column:out_qty" json:"outQty"`
		OutAmt    float64  `gorm:"column:out_amt" json:"outAmt"`
		BillNo    string   `gorm:"column:bill_no" json:"billNo"`
		CreatedAt string   `gorm:"column:created_at" json:"createdAt"`
	}
	var all []flowRow
	for _, f := range flows {
		if billType != "" && f.typeName != billType && !(billType == "组装拆装单" && (f.typeName == "组装拆装单(入)" || f.typeName == "组装拆装单(出)")) && !(billType == "同价调拨单" && f.typeName == "库存调拨单(出)") {
			continue
		}
		var rows []flowRow
		h.db.Raw(f.sql, companyID).Scan(&rows)
		all = append(all, rows...)
	}
	// 过滤
	filtered := all[:0]
	for _, r := range all {
		if billNo != "" && !strings.Contains(r.BillNo, billNo) {
			continue
		}
		if req.StartDate != "" && r.CreatedAt < req.StartDate+" 00:00:00" {
			continue
		}
		if req.EndDate != "" && r.CreatedAt > req.EndDate+" 23:59:59" {
			continue
		}
		filtered = append(filtered, r)
	}
	// 关联商品信息 + 关键字过滤
	pidSet := map[uint]bool{}
	for _, r := range filtered {
		pidSet[r.ProductID] = true
	}
	pids := make([]uint, 0, len(pidSet))
	for pid := range pidSet {
		pids = append(pids, pid)
	}
	type prod struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Unit string `gorm:"column:unit"`
	}
	prodMap := map[uint]prod{}
	if len(pids) > 0 {
		var prods []prod
		h.db.Table("products").Select("id, name, unit").Where("id IN ?", pids).Scan(&prods)
		for _, p := range prods {
			prodMap[p.ID] = p
		}
	}
	type outRow2 struct {
		flowRow
		ProductName string `json:"productName"`
		Unit        string `json:"unit"`
	}
	var rows2 []outRow2
	for _, r := range filtered {
		p := prodMap[r.ProductID]
		if req.Keyword != "" && !strings.Contains(p.Name, req.Keyword) {
			continue
		}
		rows2 = append(rows2, outRow2{flowRow: r, ProductName: p.Name, Unit: p.Unit})
	}
	// 排序：时间倒序
	for i := 0; i < len(rows2); i++ {
		for j := i + 1; j < len(rows2); j++ {
			if rows2[j].CreatedAt > rows2[i].CreatedAt {
				rows2[i], rows2[j] = rows2[j], rows2[i]
			}
		}
	}
	total := len(rows2)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	response.Ok(c, gin.H{"list": rows2[start:end], "total": total, "page": req.Page, "pageSize": req.PageSize})
}
