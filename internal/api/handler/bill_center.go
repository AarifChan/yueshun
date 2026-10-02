package handler

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/pkg/response"
)

// BillCenterHandler 单据中心处理器
type BillCenterHandler struct {
	db *gorm.DB
}

func NewBillCenterHandler(db *gorm.DB) *BillCenterHandler {
	return &BillCenterHandler{db: db}
}

func (h *BillCenterHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/bill-center", h.List)
}

// BillRow 单据中心统一行
type BillRow struct {
	BillType      string  `json:"billType"`
	BillID        uint    `json:"billId"`
	BillNo        string  `json:"billNo"`
	Status        string  `json:"status"`
	StatusLabel   string  `json:"statusLabel"`
	Counterpart   string  `json:"counterpart"`
	HandlerName   string  `json:"handlerName"`
	Quantity      float64 `json:"quantity"`
	InWarehouse   string  `json:"inWarehouse"`
	OutWarehouse  string  `json:"outWarehouse"`
	Amount        float64 `json:"amount"`
	AccountName   string  `json:"accountName"`
	PaidAmount    float64 `json:"paidAmount"`
	PayStatus     string  `json:"payStatus"`
	Remark        string  `json:"remark"`
	PostedAt      string  `json:"postedAt"`
	PrintCount    int     `json:"printCount"`
	CreatedAt     string  `json:"createdAt"`
}

type billTypeQuery struct {
	name string
	sql  string
}

func payStatusOf(total, paid float64) string {
	if total <= 0 {
		return "-"
	}
	if paid <= 0 {
		return "未支付"
	}
	if paid+0.0001 < total {
		return "部分支付"
	}
	return "已支付"
}

func statusLabelOf(s string) string {
	switch s {
	case "draft":
		return "草稿"
	case "pending":
		return "待审核"
	case "confirmed":
		return "已确认"
	case "partial":
		return "部分完成"
	case "completed":
		return "已过账"
	case "cancelled":
		return "已撤销"
	case "approved":
		return "待出库"
	case "shipped":
		return "待入库"
	default:
		return s
	}
}

// List 单据中心统一查询
// @Summary 单据中心统一查询
// @Tags 单据中心
// @Param billType query string false "单据类型"
// @Param keyword query string false "单据编号"
// @Param counterpart query string false "往来单位"
// @Param status query string false "单据状态"
// @Param payStatus query string false "支付状态"
// @Param onlyDraft query bool false "只显示草稿单"
// @Router /api/v1/bill-center [get]
func (h *BillCenterHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	billType := c.Query("billType")
	keyword := c.Query("keyword")
	counterpart := c.Query("counterpart")
	status := c.Query("status")
	payStatus := c.Query("payStatus")
	onlyDraft := c.Query("onlyDraft") == "true"
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// 各单据类型的归一化查询。
	// 列顺序: bill_type, id, bill_no, status, counterpart, handler_name, quantity,
	//         in_warehouse, out_warehouse, amount, account_name, paid_amount, remark, created_at
	queries := []billTypeQuery{
		{"采购订单", `SELECT '采购订单' AS bill_type, b.id, b.order_no AS bill_no, b.status,
			COALESCE(s.name,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM purchase_order_items i WHERE i.order_id = b.id),0) AS quantity,
			COALESCE(w.name,'') AS in_warehouse, '' AS out_warehouse,
			b.total_amount AS amount, '' AS account_name, b.paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM purchase_orders b
			LEFT JOIN suppliers s ON s.id = b.supplier_id
			LEFT JOIN employees e ON e.id = b.operator_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"采购入库单", `SELECT '采购入库单' AS bill_type, b.id, b.bill_no, b.status,
			COALESCE(s.name,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM purchase_in_stock_items i WHERE i.in_stock_id = b.id),0) AS quantity,
			COALESCE(w.name,'') AS in_warehouse, '' AS out_warehouse,
			b.total_amount AS amount, '' AS account_name, b.paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM purchase_in_stocks b
			LEFT JOIN suppliers s ON s.id = b.supplier_id
			LEFT JOIN employees e ON e.id = b.operator_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"采购退货单", `SELECT '采购退货单' AS bill_type, b.id, b.bill_no, b.status,
			COALESCE(s.name,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM purchase_return_items i WHERE i.return_id = b.id),0) AS quantity,
			'' AS in_warehouse, COALESCE(w.name,'') AS out_warehouse,
			b.total_amount AS amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM purchase_returns b
			LEFT JOIN suppliers s ON s.id = b.supplier_id
			LEFT JOIN employees e ON e.id = b.operator_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"销售订单", `SELECT '销售订单' AS bill_type, b.id, b.order_no AS bill_no, b.status,
			COALESCE(cu.name,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM sales_order_items i WHERE i.order_id = b.id),0) AS quantity,
			'' AS in_warehouse, COALESCE(w.name,'') AS out_warehouse,
			b.total_amount AS amount, '' AS account_name, b.paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM sales_orders b
			LEFT JOIN customers cu ON cu.id = b.customer_id
			LEFT JOIN employees e ON e.id = b.operator_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"销售出库单", `SELECT '销售出库单' AS bill_type, b.id, b.bill_no, b.status,
			COALESCE(cu.name,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM sales_out_stock_items i WHERE i.out_stock_id = b.id),0) AS quantity,
			'' AS in_warehouse, COALESCE(w.name,'') AS out_warehouse,
			b.total_amount AS amount, '' AS account_name, b.paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM sales_out_stocks b
			LEFT JOIN customers cu ON cu.id = b.customer_id
			LEFT JOIN employees e ON e.id = b.operator_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"销售退货单", `SELECT '销售退货单' AS bill_type, b.id, b.bill_no, b.status,
			COALESCE(cu.name,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM sales_return_items i WHERE i.return_id = b.id),0) AS quantity,
			COALESCE(w.name,'') AS in_warehouse, '' AS out_warehouse,
			b.total_amount AS amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM sales_returns b
			LEFT JOIN customers cu ON cu.id = b.customer_id
			LEFT JOIN employees e ON e.id = b.operator_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"其他入库单", `SELECT '其他入库单' AS bill_type, b.id, b.bill_no, b.status,
			COALESCE(b.counterpart,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			b.total_qty AS quantity, COALESCE(w.name,'') AS in_warehouse, '' AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM other_in_stocks b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"其他出库单", `SELECT '其他出库单' AS bill_type, b.id, b.bill_no, b.status,
			COALESCE(b.counterpart,'') AS counterpart, COALESCE(e.name,'') AS handler_name,
			b.total_qty AS quantity, '' AS in_warehouse, COALESCE(w.name,'') AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM other_out_stocks b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"同价调拨单", `SELECT '同价调拨单' AS bill_type, b.id, b.bill_no, b.status,
			'' AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM inventory_transfer_items i WHERE i.transfer_id = b.id),0) AS quantity,
			COALESCE(tw.name,'') AS in_warehouse, COALESCE(fw.name,'') AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM inventory_transfers b
			LEFT JOIN employees e ON e.id = b.operator_id
			LEFT JOIN warehouses fw ON fw.id = b.from_warehouse_id
			LEFT JOIN warehouses tw ON tw.id = b.to_warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"调拨出库单(同价)", `SELECT '调拨出库单(同价)' AS bill_type, b.id, b.bill_no, b.status,
			'' AS counterpart, COALESCE(e.name,'') AS handler_name, b.total_qty AS quantity,
			'' AS in_warehouse, COALESCE(fw.name,'') AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM transfer_outs b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses fw ON fw.id = b.from_warehouse_id
			WHERE b.company_id = ? AND b.transfer_type = 'same' AND b.deleted_at IS NULL`},
		{"调拨入库单(同价)", `SELECT '调拨入库单(同价)' AS bill_type, b.id, b.bill_no, b.status,
			'' AS counterpart, COALESCE(e.name,'') AS handler_name, b.total_qty AS quantity,
			COALESCE(tw.name,'') AS in_warehouse, '' AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM transfer_ins b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses tw ON tw.id = b.to_warehouse_id
			WHERE b.company_id = ? AND b.transfer_type = 'same' AND b.deleted_at IS NULL`},
		{"调拨出库单(变价)", `SELECT '调拨出库单(变价)' AS bill_type, b.id, b.bill_no, b.status,
			'' AS counterpart, COALESCE(e.name,'') AS handler_name, b.total_qty AS quantity,
			'' AS in_warehouse, COALESCE(fw.name,'') AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM transfer_outs b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses fw ON fw.id = b.from_warehouse_id
			WHERE b.company_id = ? AND b.transfer_type = 'diff' AND b.deleted_at IS NULL`},
		{"调拨入库单(变价)", `SELECT '调拨入库单(变价)' AS bill_type, b.id, b.bill_no, b.status,
			'' AS counterpart, COALESCE(e.name,'') AS handler_name, b.total_qty AS quantity,
			COALESCE(tw.name,'') AS in_warehouse, '' AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM transfer_ins b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses tw ON tw.id = b.to_warehouse_id
			WHERE b.company_id = ? AND b.transfer_type = 'diff' AND b.deleted_at IS NULL`},
		{"成本调价单", `SELECT '成本调价单' AS bill_type, b.id, b.bill_no, b.status,
			'' AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM cost_adjust_items i WHERE i.adjust_id = b.id),0) AS quantity,
			COALESCE(w.name,'') AS in_warehouse, '' AS out_warehouse,
			b.amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM cost_adjusts b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses w ON w.id = b.warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
		{"组装拆装单", `SELECT '组装拆装单' AS bill_type, b.id, b.bill_no, b.status,
			'' AS counterpart, COALESCE(e.name,'') AS handler_name,
			COALESCE((SELECT SUM(quantity) FROM assembly_order_items i WHERE i.order_id = b.id),0) AS quantity,
			COALESCE(iw.name,'') AS in_warehouse, COALESCE(ow.name,'') AS out_warehouse,
			b.in_amount AS amount, '' AS account_name, 0 AS paid_amount, COALESCE(b.remark,'') AS remark, b.created_at
			FROM assembly_orders b
			LEFT JOIN employees e ON e.id = b.handler_id
			LEFT JOIN warehouses ow ON ow.id = b.out_warehouse_id
			LEFT JOIN warehouses iw ON iw.id = b.in_warehouse_id
			WHERE b.company_id = ? AND b.deleted_at IS NULL`},
	}

	type rawRow struct {
		BillType     string  `gorm:"column:bill_type"`
		ID           uint    `gorm:"column:id"`
		BillNo       string  `gorm:"column:bill_no"`
		Status       string  `gorm:"column:status"`
		Counterpart  string  `gorm:"column:counterpart"`
		HandlerName  string  `gorm:"column:handler_name"`
		Quantity     float64 `gorm:"column:quantity"`
		InWarehouse  string  `gorm:"column:in_warehouse"`
		OutWarehouse string  `gorm:"column:out_warehouse"`
		Amount       float64 `gorm:"column:amount"`
		AccountName  string  `gorm:"column:account_name"`
		PaidAmount   float64 `gorm:"column:paid_amount"`
		Remark       string  `gorm:"column:remark"`
		CreatedAt    string  `gorm:"column:created_at"`
	}

	var rows []BillRow
	for _, q := range queries {
		if billType != "" && q.name != billType {
			continue
		}
		var raw []rawRow
		if err := h.db.Raw(q.sql, companyID).Scan(&raw).Error; err != nil {
			continue
		}
		for _, r := range raw {
			if onlyDraft && r.Status != "draft" {
				continue
			}
			if status != "" && r.Status != status {
				continue
			}
			if keyword != "" && !strings.Contains(r.BillNo, keyword) {
				continue
			}
			if counterpart != "" && !strings.Contains(r.Counterpart, counterpart) {
				continue
			}
			if startDate != "" && r.CreatedAt < startDate+" 00:00:00" {
				continue
			}
			if endDate != "" && r.CreatedAt > endDate+" 23:59:59" {
				continue
			}
			ps := payStatusOf(r.Amount, r.PaidAmount)
			if payStatus != "" && ps != payStatus {
				continue
			}
			rows = append(rows, BillRow{
				BillType: r.BillType, BillID: r.ID, BillNo: r.BillNo,
				Status: r.Status, StatusLabel: statusLabelOf(r.Status),
				Counterpart: r.Counterpart, HandlerName: r.HandlerName,
				Quantity: r.Quantity, InWarehouse: r.InWarehouse, OutWarehouse: r.OutWarehouse,
				Amount: r.Amount, AccountName: r.AccountName, PaidAmount: r.PaidAmount,
				PayStatus: ps, Remark: r.Remark, PostedAt: "", PrintCount: 0, CreatedAt: r.CreatedAt,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].CreatedAt > rows[j].CreatedAt })
	total := len(rows)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	response.Ok(c, gin.H{"list": rows[start:end], "total": total, "page": page, "pageSize": pageSize})
}
