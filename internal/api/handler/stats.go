package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// StatsHandler 数据统计处理器
type StatsHandler struct {
	db *gorm.DB
}

func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

func (h *StatsHandler) RegisterRoutes(r *gin.RouterGroup) {
	s := r.Group("/stats")
	{
		s.GET("/dashboard", h.Dashboard)
	}
}

// Dashboard 数据看板
func (h *StatsHandler) Dashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var productCount, customerCount int64
	h.db.Model(&model.Product{}).Where("company_id = ?", companyID).Count(&productCount)
	h.db.Model(&model.Customer{}).Where("company_id = ?", companyID).Count(&customerCount)

	var purchaseDraft, salesDraft, approvalPending int64
	h.db.Model(&model.PurchaseOrder{}).Where("company_id = ? AND status = 'draft'", companyID).Count(&purchaseDraft)
	h.db.Model(&model.SalesOrder{}).Where("company_id = ? AND status = 'draft'", companyID).Count(&salesDraft)
	h.db.Model(&model.ApprovalRecord{}).Where("company_id = ? AND action = 'pending'", companyID).Count(&approvalPending)
	pendingBills := purchaseDraft + salesDraft + approvalPending

	var warningCount int64
	h.db.Model(&model.InventoryWarning{}).Where("company_id = ? AND status <> 'resolved'", companyID).Count(&warningCount)

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var todaySales float64
	var todayOrders int64
	h.db.Model(&model.SalesOrder{}).
		Where("company_id = ? AND status <> 'cancelled' AND created_at >= ?", companyID, dayStart).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&todaySales)
	h.db.Model(&model.SalesOrder{}).
		Where("company_id = ? AND status <> 'cancelled' AND created_at >= ?", companyID, dayStart).
		Count(&todayOrders)

	var monthSales float64
	h.db.Model(&model.SalesOrder{}).
		Where("company_id = ? AND status <> 'cancelled' AND created_at >= ?", companyID, monthStart).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&monthSales)

	var totalReceivable float64
	h.db.Model(&model.Customer{}).
		Where("company_id = ?", companyID).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&totalReceivable)

	log.Debug().Uint("companyID", companyID).Msg("dashboard stats")

	response.Ok(c, gin.H{
		"productCount":    productCount,
		"customerCount":   customerCount,
		"pendingBills":    pendingBills,
		"warningCount":    warningCount,
		"todaySales":      todaySales,
		"todayOrders":     todayOrders,
		"monthSales":      monthSales,
		"totalReceivable": totalReceivable,
	})
}
