package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// ==================== 库存状况表 ====================

type StockStatusReq struct {
	Page          int    `form:"page,default=1"`
	PageSize      int    `form:"pageSize,default=30"`
	Keyword       string `form:"keyword"`
	WarehouseID   uint   `form:"warehouseId"`
	CategoryID    uint   `form:"categoryId"`
	StockStatus   string `form:"stockStatus"`   // in 有库存 / out 无库存 / belowMin 低于下限 / aboveMax 高于上限
	ProductStatus string `form:"productStatus"` // "" 全部 / 1 已启用 / 0 已停用
	OnlyWithStock bool   `form:"onlyWithStock"` // 仅显示有账面库存商品
	GroupBy       string `form:"groupBy,default=product"`
}

type stockStatusRow struct {
	ProductID     uint    `json:"productId"`
	ProductName   string  `json:"productName"`
	Code          string  `json:"code"`
	Specification string  `json:"specification"`
	Unit          string  `json:"unit"`
	Image         string  `json:"image"`
	CategoryID    uint    `json:"categoryId"`
	CategoryName  string  `json:"categoryName"`
	CostPrice     float64 `json:"costPrice"`
	MinStock      float64 `json:"minStock"`
	MaxStock      float64 `json:"maxStock"`
	Status        int8    `json:"status"`
	Quantity      float64 `json:"quantity"`
}

type StockStatusResp struct {
	stockStatusRow
	CostAmount    float64 `json:"costAmount"`
	ReservedQty   float64 `json:"reservedQty"`   // 预订量（暂无预订业务，固定 0）
	PendingOutQty float64 `json:"pendingOutQty"` // 待出库数量（暂无占用业务，固定 0）
	AvailableQty  float64 `json:"availableQty"`  // 可用库存量 = 库存总量 - 预订量 - 待出库数量
}

type StockStatusSummary struct {
	TotalQuantity float64 `json:"totalQuantity"`
	TotalAmount   float64 `json:"totalAmount"`
}

type StockStatusCategoryResp struct {
	CategoryID   uint    `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	ProductCount int     `json:"productCount"`
	Quantity     float64 `json:"quantity"`
	CostAmount   float64 `json:"costAmount"`
}

// ListStockStatus 库存状况表
// @Summary 库存状况表（按商品/按分类）
// @Tags 库存
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "商品名称/编号"
// @Param warehouseId query int false "仓库ID"
// @Param categoryId query int false "商品分类ID（含子分类）"
// @Param stockStatus query string false "库存状态 in/out/belowMin/aboveMax"
// @Param productStatus query string false "商品状态 1启用 0停用"
// @Param onlyWithStock query bool false "仅显示有账面库存商品"
// @Param groupBy query string false "product 按商品 / category 按分类"
// @Success 200 {object} response.Response
// @Router /stocks/status [get]
func (h *WarehouseHandler) ListStockStatus(c *gin.Context) {
	var req StockStatusReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 库存 JOIN 条件放在 ON 子句，保证无库存商品（quantity=0）也能查出
	stockJoin := "LEFT JOIN stocks s ON s.product_id = products.id AND s.company_id = products.company_id"
	joinArgs := []interface{}{}
	if req.WarehouseID > 0 {
		stockJoin += " AND s.warehouse_id = ?"
		joinArgs = append(joinArgs, req.WarehouseID)
	}

	query := h.db.Model(&model.Product{}).
		Select(`products.id AS product_id, products.name AS product_name, products.code,
			products.specification, products.unit, products.image, products.category_id,
			products.purchase_price AS cost_price, products.min_stock, products.max_stock, products.status,
			COALESCE(pc.name, '') AS category_name,
			COALESCE(SUM(s.quantity), 0) AS quantity`).
		Joins(stockJoin, joinArgs...).
		Joins("LEFT JOIN product_categories pc ON pc.id = products.category_id").
		Where("products.company_id = ?", companyID).
		Group("products.id").Group("pc.name").
		Order("products.id")

	if req.Keyword != "" {
		query = query.Where("products.name LIKE ? OR products.code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.CategoryID > 0 {
		catIDs := h.descendantCategoryIDs(companyID, req.CategoryID)
		query = query.Where("products.category_id IN ?", catIDs)
	}
	if req.ProductStatus == "0" || req.ProductStatus == "1" {
		query = query.Where("products.status = ?", req.ProductStatus)
	}

	var rows []stockStatusRow
	if err := query.Scan(&rows).Error; err != nil {
		log.Error().Err(err).Msg("list stock status failed")
		response.ServerError(c, "查询失败")
		return
	}

	// 库存量相关过滤（聚合结果，在内存中处理）
	filtered := rows[:0]
	for _, r := range rows {
		if req.OnlyWithStock && r.Quantity == 0 {
			continue
		}
		switch req.StockStatus {
		case "in":
			if r.Quantity <= 0 {
				continue
			}
		case "out":
			if r.Quantity > 0 {
				continue
			}
		case "belowMin":
			if !(r.MinStock > 0 && r.Quantity < r.MinStock) {
				continue
			}
		case "aboveMax":
			if !(r.MaxStock > 0 && r.Quantity > r.MaxStock) {
				continue
			}
		}
		filtered = append(filtered, r)
	}

	summary := StockStatusSummary{}
	for _, r := range filtered {
		summary.TotalQuantity += r.Quantity
		summary.TotalAmount += r.Quantity * r.CostPrice
	}

	if req.GroupBy == "category" {
		h.replyStockStatusByCategory(c, filtered, summary)
		return
	}

	total := len(filtered)
	start := (req.Page - 1) * req.PageSize
	if start > total {
		start = total
	}
	end := start + req.PageSize
	if end > total {
		end = total
	}

	list := make([]StockStatusResp, 0, end-start)
	for _, r := range filtered[start:end] {
		list = append(list, StockStatusResp{
			stockStatusRow: r,
			CostAmount:     r.Quantity * r.CostPrice,
			ReservedQty:    0,
			PendingOutQty:  0,
			AvailableQty:   r.Quantity,
		})
	}

	response.Ok(c, gin.H{
		"list":     list,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
		"summary":  summary,
	})
}

func (h *WarehouseHandler) replyStockStatusByCategory(c *gin.Context, rows []stockStatusRow, summary StockStatusSummary) {
	type agg struct {
		name  string
		count int
		qty   float64
		amt   float64
		order int
	}
	aggs := map[uint]*agg{}
	order := []uint{}
	for _, r := range rows {
		a, ok := aggs[r.CategoryID]
		if !ok {
			a = &agg{name: r.CategoryName}
			if a.name == "" {
				a.name = "未分类"
			}
			aggs[r.CategoryID] = a
			order = append(order, r.CategoryID)
		}
		a.count++
		a.qty += r.Quantity
		a.amt += r.Quantity * r.CostPrice
	}
	list := make([]StockStatusCategoryResp, 0, len(order))
	for _, id := range order {
		a := aggs[id]
		list = append(list, StockStatusCategoryResp{
			CategoryID:   id,
			CategoryName: a.name,
			ProductCount: a.count,
			Quantity:     a.qty,
			CostAmount:   a.amt,
		})
	}
	response.Ok(c, gin.H{
		"list":     list,
		"total":    len(list),
		"page":     1,
		"pageSize": len(list),
		"summary":  summary,
	})
}

// descendantCategoryIDs 返回分类及其全部子孙分类 ID
func (h *WarehouseHandler) descendantCategoryIDs(companyID, rootID uint) []uint {
	var cats []model.ProductCategory
	if err := h.db.Where("company_id = ?", companyID).Find(&cats).Error; err != nil {
		return []uint{rootID}
	}
	children := map[uint][]uint{}
	for _, cat := range cats {
		children[cat.ParentID] = append(children[cat.ParentID], cat.ID)
	}
	ids := []uint{rootID}
	queue := []uint{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range children[cur] {
			ids = append(ids, child)
			queue = append(queue, child)
		}
	}
	return ids
}

type StockLimitReq struct {
	ProductIDs []uint  `json:"productIds" binding:"required"`
	MinStock   float64 `json:"minStock"`
	MaxStock   float64 `json:"maxStock"`
}

// UpdateStockLimits 批量设置库存上下限
// @Summary 批量设置商品库存上下限
// @Tags 库存
// @Param body body StockLimitReq true "商品ID列表与上下限"
// @Success 200 {object} response.Response
// @Router /stocks/status/limits [put]
func (h *WarehouseHandler) UpdateStockLimits(c *gin.Context) {
	var req StockLimitReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.ProductIDs) == 0 {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	if err := h.db.Model(&model.Product{}).
		Where("company_id = ? AND id IN ?", companyID, req.ProductIDs).
		Updates(map[string]interface{}{"min_stock": req.MinStock, "max_stock": req.MaxStock}).Error; err != nil {
		log.Error().Err(err).Msg("update stock limits failed")
		response.ServerError(c, "设置失败")
		return
	}
	response.OkWithMessage(c, "设置成功", nil)
}

// ==================== 库存状况表-行操作 ====================

type StockDistributionItem struct {
	WarehouseID   uint    `json:"warehouseId"`
	WarehouseName string  `json:"warehouseName"`
	Quantity      float64 `json:"quantity"`
	CostAmount    float64 `json:"costAmount"`
}

// ListStockDistribution 单商品库存分布（按仓库）
// @Summary 商品库存分布
// @Tags 库存
// @Param productId path int true "商品ID"
// @Success 200 {object} response.Response
// @Router /stocks/status/{productId}/warehouses [get]
func (h *WarehouseHandler) ListStockDistribution(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	productID, _ := strconv.Atoi(c.Param("productId"))
	if productID <= 0 {
		response.BadRequest(c, "商品ID无效")
		return
	}

	var product model.Product
	if err := h.db.Select("id", "purchase_price").
		Where("company_id = ? AND id = ?", companyID, productID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	type row struct {
		WarehouseID   uint    `json:"warehouseId"`
		WarehouseName string  `json:"warehouseName"`
		Quantity      float64 `json:"quantity"`
	}
	var rows []row
	if err := h.db.Raw(`
		SELECT w.id AS warehouse_id, w.name AS warehouse_name, COALESCE(s.quantity, 0) AS quantity
		FROM warehouses w
		LEFT JOIN stocks s ON s.warehouse_id = w.id AND s.product_id = ? AND s.company_id = ?
		WHERE w.company_id = ? AND w.deleted_at IS NULL
		ORDER BY w.id`, productID, companyID, companyID).Scan(&rows).Error; err != nil {
		log.Error().Err(err).Msg("list stock distribution failed")
		response.ServerError(c, "查询失败")
		return
	}

	list := make([]StockDistributionItem, 0, len(rows))
	var totalQty float64
	for _, r := range rows {
		totalQty += r.Quantity
		list = append(list, StockDistributionItem{
			WarehouseID:   r.WarehouseID,
			WarehouseName: r.WarehouseName,
			Quantity:      r.Quantity,
			CostAmount:    r.Quantity * product.PurchasePrice,
		})
	}
	response.Ok(c, gin.H{"list": list, "totalQuantity": totalQty, "costPrice": product.PurchasePrice})
}

type StockFlowItem struct {
	BillDate      string  `json:"billDate"`
	BillNo        string  `json:"billNo"`
	BillType      string  `json:"billType"`
	WarehouseID   uint    `json:"warehouseId"`
	WarehouseName string  `json:"warehouseName"`
	InQty         float64 `json:"inQty"`
	OutQty        float64 `json:"outQty"`
	Price         float64 `json:"price"`
}

// 已完成单据的出入库流水 UNION 子查询（每段参数：companyID, productID）
const stockFlowUnionSQL = `
	SELECT b.bill_date, b.bill_no, '采购入库' AS bill_type, b.warehouse_id, i.quantity AS in_qty, 0::numeric AS out_qty, i.price, i.created_at
	FROM purchase_in_stock_items i JOIN purchase_in_stocks b ON b.id = i.in_stock_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '采购退货', b.warehouse_id, 0::numeric, i.quantity, i.price, i.created_at
	FROM purchase_return_items i JOIN purchase_returns b ON b.id = i.return_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '销售出库', b.warehouse_id, 0::numeric, i.quantity, i.price, i.created_at
	FROM sales_out_stock_items i JOIN sales_out_stocks b ON b.id = i.out_stock_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '销售退货', b.warehouse_id, i.quantity, 0::numeric, i.price, i.created_at
	FROM sales_return_items i JOIN sales_returns b ON b.id = i.return_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '库存盘点', b.warehouse_id,
		CASE WHEN i.diff_qty > 0 THEN i.diff_qty ELSE 0::numeric END,
		CASE WHEN i.diff_qty < 0 THEN -i.diff_qty ELSE 0::numeric END, i.price, i.created_at
	FROM inventory_check_items i JOIN inventory_checks b ON b.id = i.check_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND i.diff_qty <> 0 AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '调拨出库', b.from_warehouse_id, 0::numeric, i.quantity, i.price, i.created_at
	FROM inventory_transfer_items i JOIN inventory_transfers b ON b.id = i.transfer_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '调拨入库', b.to_warehouse_id, i.quantity, 0::numeric, i.price, i.created_at
	FROM inventory_transfer_items i JOIN inventory_transfers b ON b.id = i.transfer_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '其他入库', b.warehouse_id, i.quantity, 0::numeric, i.price, i.created_at
	FROM other_in_stock_items i JOIN other_in_stocks b ON b.id = i.in_stock_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL
	UNION ALL
	SELECT b.bill_date, b.bill_no, '其他出库', b.warehouse_id, 0::numeric, i.quantity, i.price, i.created_at
	FROM other_out_stock_items i JOIN other_out_stocks b ON b.id = i.out_stock_id
	WHERE b.company_id = ? AND i.product_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND i.deleted_at IS NULL`

// ListStockFlows 单商品出入库明细（流水）
// @Summary 商品出入库明细
// @Tags 库存
// @Param productId path int true "商品ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /stocks/status/{productId}/flows [get]
func (h *WarehouseHandler) ListStockFlows(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	productID, _ := strconv.Atoi(c.Param("productId"))
	if productID <= 0 {
		response.BadRequest(c, "商品ID无效")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 9 个 UNION 段，每段参数为 (companyID, productID)
	args := make([]interface{}, 0, 18)
	for i := 0; i < 9; i++ {
		args = append(args, companyID, productID)
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM (" + stockFlowUnionSQL + ") t"
	if err := h.db.Raw(countSQL, args...).Scan(&total).Error; err != nil {
		log.Error().Err(err).Msg("count stock flows failed")
		response.ServerError(c, "查询失败")
		return
	}

	type flowRow struct {
		BillDate      time.Time `json:"billDate"`
		BillNo        string    `json:"billNo"`
		BillType      string    `json:"billType"`
		WarehouseID   uint      `json:"warehouseId"`
		WarehouseName string    `json:"warehouseName"`
		InQty         float64   `json:"inQty"`
		OutQty        float64   `json:"outQty"`
		Price         float64   `json:"price"`
	}
	var rows []flowRow
	listSQL := `
		SELECT t.bill_date, t.bill_no, t.bill_type, t.warehouse_id,
			COALESCE(w.name, '') AS warehouse_name, t.in_qty, t.out_qty, t.price
		FROM (` + stockFlowUnionSQL + `) t
		LEFT JOIN warehouses w ON w.id = t.warehouse_id
		ORDER BY t.bill_date DESC, t.created_at DESC
		LIMIT ? OFFSET ?`
	listArgs := append(append([]interface{}{}, args...), pageSize, (page-1)*pageSize)
	if err := h.db.Raw(listSQL, listArgs...).Scan(&rows).Error; err != nil {
		log.Error().Err(err).Msg("list stock flows failed")
		response.ServerError(c, "查询失败")
		return
	}

	list := make([]StockFlowItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, StockFlowItem{
			BillDate:      r.BillDate.Format("2006-01-02"),
			BillNo:        r.BillNo,
			BillType:      r.BillType,
			WarehouseID:   r.WarehouseID,
			WarehouseName: r.WarehouseName,
			InQty:         r.InQty,
			OutQty:        r.OutQty,
			Price:         r.Price,
		})
	}
	response.OkWithPage(c, list, page, pageSize, int(total))
}
