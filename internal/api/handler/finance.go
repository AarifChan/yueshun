package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// FinanceHandler 资金模块（费用单/其他收入单 + 资金报表）
type FinanceHandler struct {
	db *gorm.DB
}

func NewFinanceHandler(db *gorm.DB) *FinanceHandler {
	return &FinanceHandler{db: db}
}

func (h *FinanceHandler) RegisterRoutes(r *gin.RouterGroup) {
	eb := r.Group("/expense-bills")
	{
		eb.GET("", h.ListExpenseBills)
		eb.POST("", h.CreateExpenseBill)
		eb.PUT("/:id", h.UpdateExpenseBill)
		eb.DELETE("/:id", h.DeleteExpenseBill)
		eb.PUT("/:id/complete", h.CompleteExpenseBill)
	}
	ib := r.Group("/other-income-bills")
	{
		ib.GET("", h.ListIncomeBills)
		ib.POST("", h.CreateIncomeBill)
		ib.PUT("/:id", h.UpdateIncomeBill)
		ib.DELETE("/:id", h.DeleteIncomeBill)
		ib.PUT("/:id/complete", h.CompleteIncomeBill)
	}
	g := r.Group("/finance-reports")
	{
		g.GET("/receivable", h.Receivable)
		g.GET("/payable", h.Payable)
		g.GET("/balance", h.Balance)
		g.GET("/customer-advance", h.CustomerAdvance)
		g.GET("/brand-advance", h.BrandAdvance)
		g.GET("/brand-unsettled", h.BrandUnsettled)
		g.GET("/unsettled-bills", h.UnsettledBills)
		g.GET("/cash-bank", h.CashBank)
		g.GET("/account-income", h.AccountIncome)
		g.GET("/revenue-expense", h.RevenueExpense)
		g.GET("/expense-distribution", h.ExpenseDistribution)
		g.GET("/income-distribution", h.IncomeDistribution)
		g.GET("/operating-profit", h.OperatingProfit)
	}
}

// ==================== 费用单 ====================

type expenseBillReq struct {
	BillDate    string  `json:"billDate" binding:"required"`
	AccountID   uint    `json:"accountId" binding:"required"`
	ItemID      uint    `json:"itemId" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Counterpart string  `json:"counterpart"`
	HandlerID   uint    `json:"handlerId"`
	DeptID      uint    `json:"deptId"`
	Remark      string  `json:"remark"`
}

func (h *FinanceHandler) ListExpenseBills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	status := c.Query("status")
	start := c.DefaultQuery("startDate", "2000-01-01")
	end := c.DefaultQuery("endDate", "2099-12-31")

	q := h.db.Model(&model.ExpenseBill{}).Where("company_id = ? AND bill_date >= ? AND bill_date <= ?", companyID, start, end)
	if keyword != "" {
		q = q.Where("bill_no ILIKE ? OR counterpart ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var list []model.ExpenseBill
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	// 补充名称
	rows := make([]gin.H, 0, len(list))
	for _, b := range list {
		rows = append(rows, h.expenseBillRow(b))
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *FinanceHandler) expenseBillRow(b model.ExpenseBill) gin.H {
	var account, item, handler string
	h.db.Model(&model.Account{}).Where("id = ?", b.AccountID).Select("name").Scan(&account)
	h.db.Model(&model.IncomeExpenseItem{}).Where("id = ?", b.ItemID).Select("name").Scan(&item)
	if b.HandlerID > 0 {
		h.db.Model(&model.Employee{}).Where("id = ?", b.HandlerID).Select("name").Scan(&handler)
	}
	return gin.H{
		"id": b.ID, "billNo": b.BillNo, "billDate": b.BillDate.Format("2006-01-02"),
		"accountId": b.AccountID, "accountName": account, "itemId": b.ItemID, "itemName": item,
		"amount": b.Amount, "counterpart": b.Counterpart, "handlerId": b.HandlerID, "handlerName": handler,
		"deptId": b.DeptID, "status": b.Status, "statusLabel": statusLabelOf(b.Status),
		"remark": b.Remark, "createdAt": b.CreatedAt,
	}
}

func (h *FinanceHandler) CreateExpenseBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req expenseBillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	bill := model.ExpenseBill{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		BillNo:               genDailyBillNo(h.db, "expense_bills", "FY"),
		BillDate:             billDate,
		AccountID:            req.AccountID,
		ItemID:               req.ItemID,
		Amount:               req.Amount,
		Counterpart:          req.Counterpart,
		HandlerID:            req.HandlerID,
		DeptID:               req.DeptID,
		Status:               "pending",
		Remark:               req.Remark,
	}
	if err := h.db.Create(&bill).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, h.expenseBillRow(bill))
}

func (h *FinanceHandler) UpdateExpenseBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var bill model.ExpenseBill
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.Fail(c, 4004, "费用单不存在")
		return
	}
	if bill.Status != "pending" {
		response.Fail(c, 4000, "只有待审核状态可编辑")
		return
	}
	var req expenseBillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	bill.BillDate, _ = time.Parse("2006-01-02", req.BillDate)
	bill.AccountID = req.AccountID
	bill.ItemID = req.ItemID
	bill.Amount = req.Amount
	bill.Counterpart = req.Counterpart
	bill.HandlerID = req.HandlerID
	bill.DeptID = req.DeptID
	bill.Remark = req.Remark
	if err := h.db.Save(&bill).Error; err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, h.expenseBillRow(bill))
}

func (h *FinanceHandler) DeleteExpenseBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var bill model.ExpenseBill
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.Fail(c, 4004, "费用单不存在")
		return
	}
	if bill.Status != "pending" {
		response.Fail(c, 4000, "只有待审核状态可删除")
		return
	}
	h.db.Delete(&bill)
	response.Ok(c, nil)
}

func (h *FinanceHandler) CompleteExpenseBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var bill model.ExpenseBill
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.Fail(c, 4004, "费用单不存在")
		return
	}
	if bill.Status != "pending" {
		response.Fail(c, 4000, "只有待审核状态可完成")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		bill.Status = "completed"
		if err := tx.Save(&bill).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Account{}).Where("id = ? AND company_id = ?", bill.AccountID, companyID).
			UpdateColumn("balance", gorm.Expr("balance - ?", bill.Amount)).Error; err != nil {
			return err
		}
		return tx.Create(&model.AccountFlow{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			AccountID:            bill.AccountID,
			Type:                 "expense",
			Amount:               bill.Amount,
			RefType:              "expense_bill",
			RefID:                bill.ID,
			Remark:               bill.BillNo,
		}).Error
	})
	if err != nil {
		response.Fail(c, 4000, "完成失败: "+err.Error())
		return
	}
	response.Ok(c, h.expenseBillRow(bill))
}

// ==================== 其他收入单 ====================

func (h *FinanceHandler) ListIncomeBills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	status := c.Query("status")
	start := c.DefaultQuery("startDate", "2000-01-01")
	end := c.DefaultQuery("endDate", "2099-12-31")

	q := h.db.Model(&model.OtherIncomeBill{}).Where("company_id = ? AND bill_date >= ? AND bill_date <= ?", companyID, start, end)
	if keyword != "" {
		q = q.Where("bill_no ILIKE ? OR counterpart ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var list []model.OtherIncomeBill
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	rows := make([]gin.H, 0, len(list))
	for _, b := range list {
		rows = append(rows, h.incomeBillRow(b))
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *FinanceHandler) incomeBillRow(b model.OtherIncomeBill) gin.H {
	var account, item, handler string
	h.db.Model(&model.Account{}).Where("id = ?", b.AccountID).Select("name").Scan(&account)
	h.db.Model(&model.IncomeExpenseItem{}).Where("id = ?", b.ItemID).Select("name").Scan(&item)
	if b.HandlerID > 0 {
		h.db.Model(&model.Employee{}).Where("id = ?", b.HandlerID).Select("name").Scan(&handler)
	}
	return gin.H{
		"id": b.ID, "billNo": b.BillNo, "billDate": b.BillDate.Format("2006-01-02"),
		"accountId": b.AccountID, "accountName": account, "itemId": b.ItemID, "itemName": item,
		"amount": b.Amount, "counterpart": b.Counterpart, "handlerId": b.HandlerID, "handlerName": handler,
		"deptId": b.DeptID, "status": b.Status, "statusLabel": statusLabelOf(b.Status),
		"remark": b.Remark, "createdAt": b.CreatedAt,
	}
}

func (h *FinanceHandler) CreateIncomeBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req expenseBillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	billDate, _ := time.Parse("2006-01-02", req.BillDate)
	bill := model.OtherIncomeBill{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		BillNo:               genDailyBillNo(h.db, "other_income_bills", "SR"),
		BillDate:             billDate,
		AccountID:            req.AccountID,
		ItemID:               req.ItemID,
		Amount:               req.Amount,
		Counterpart:          req.Counterpart,
		HandlerID:            req.HandlerID,
		DeptID:               req.DeptID,
		Status:               "pending",
		Remark:               req.Remark,
	}
	if err := h.db.Create(&bill).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, h.incomeBillRow(bill))
}

func (h *FinanceHandler) UpdateIncomeBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var bill model.OtherIncomeBill
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.Fail(c, 4004, "收入单不存在")
		return
	}
	if bill.Status != "pending" {
		response.Fail(c, 4000, "只有待审核状态可编辑")
		return
	}
	var req expenseBillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	bill.BillDate, _ = time.Parse("2006-01-02", req.BillDate)
	bill.AccountID = req.AccountID
	bill.ItemID = req.ItemID
	bill.Amount = req.Amount
	bill.Counterpart = req.Counterpart
	bill.HandlerID = req.HandlerID
	bill.DeptID = req.DeptID
	bill.Remark = req.Remark
	if err := h.db.Save(&bill).Error; err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, h.incomeBillRow(bill))
}

func (h *FinanceHandler) DeleteIncomeBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var bill model.OtherIncomeBill
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.Fail(c, 4004, "收入单不存在")
		return
	}
	if bill.Status != "pending" {
		response.Fail(c, 4000, "只有待审核状态可删除")
		return
	}
	h.db.Delete(&bill)
	response.Ok(c, nil)
}

func (h *FinanceHandler) CompleteIncomeBill(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var bill model.OtherIncomeBill
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&bill).Error; err != nil {
		response.Fail(c, 4004, "收入单不存在")
		return
	}
	if bill.Status != "pending" {
		response.Fail(c, 4000, "只有待审核状态可完成")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		bill.Status = "completed"
		if err := tx.Save(&bill).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Account{}).Where("id = ? AND company_id = ?", bill.AccountID, companyID).
			UpdateColumn("balance", gorm.Expr("balance + ?", bill.Amount)).Error; err != nil {
			return err
		}
		return tx.Create(&model.AccountFlow{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			AccountID:            bill.AccountID,
			Type:                 "income",
			Amount:               bill.Amount,
			RefType:              "other_income_bill",
			RefID:                bill.ID,
			Remark:               bill.BillNo,
		}).Error
	})
	if err != nil {
		response.Fail(c, 4000, "完成失败: "+err.Error())
		return
	}
	response.Ok(c, h.incomeBillRow(bill))
}

// ==================== 应收查询 ====================

func (h *FinanceHandler) Receivable(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("customerKeyword")
	includeZero := c.Query("includeZero") == "true"

	type row struct {
		CustomerID     uint    `json:"customerId"`
		Code           string  `json:"code"`
		Name           string  `json:"name"`
		InitReceivable float64 `json:"initReceivable"`
		PeriodSale     float64 `json:"periodSale"`
		PeriodReceipt  float64 `json:"periodReceipt"`
		ReceivableBal  float64 `json:"receivableBal"`
		InitAdvance    float64 `json:"initAdvance"`
		PeriodAdvance  float64 `json:"periodAdvance"`
		AdvanceBal     float64 `json:"advanceBal"`
		Debt           float64 `json:"debt"`
	}
	stats := map[uint]*row{}

	var customers []model.Customer
	h.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Find(&customers)
	for _, cu := range customers {
		if keyword != "" && !strings.Contains(cu.Name, keyword) && !strings.Contains(cu.Code, keyword) {
			continue
		}
		stats[cu.ID] = &row{CustomerID: cu.ID, Code: cu.Code, Name: cu.Name}
	}

	// 期初
	var initRows []struct {
		TargetID   uint    `gorm:"column:target_id"`
		Receivable float64 `gorm:"column:receivable"`
		Advance    float64 `gorm:"column:advance"`
	}
	h.db.Raw(`SELECT target_id, receivable, advance FROM initial_balances
		WHERE company_id = ? AND biz_type = 'customer' AND deleted_at IS NULL`, companyID).Scan(&initRows)
	for _, r := range initRows {
		if s, ok := stats[r.TargetID]; ok {
			s.InitReceivable = r.Receivable
			s.InitAdvance = r.Advance
			s.AdvanceBal = r.Advance
		}
	}

	// 本期销售（出库 - 退货）
	var saleRows []struct {
		CustomerID uint    `gorm:"column:customer_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT customer_id, SUM(total_amount) AS amount FROM sales_out_stocks
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY customer_id`, companyID, start, end).Scan(&saleRows)
	for _, r := range saleRows {
		if s, ok := stats[r.CustomerID]; ok {
			s.PeriodSale = r.Amount
		}
	}
	var retRows []struct {
		CustomerID uint    `gorm:"column:customer_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT customer_id, SUM(total_amount) AS amount FROM sales_returns
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY customer_id`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if s, ok := stats[r.CustomerID]; ok {
			s.PeriodSale -= r.Amount
		}
	}

	// 本期收款
	var rcptRows []struct {
		CustomerID uint    `gorm:"column:customer_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT customer_id, SUM(total_amount) AS amount FROM sales_receipts
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY customer_id`, companyID, start, end).Scan(&rcptRows)
	for _, r := range rcptRows {
		if s, ok := stats[r.CustomerID]; ok {
			s.PeriodReceipt = r.Amount
		}
	}

	list := make([]row, 0, len(stats))
	for _, s := range stats {
		s.ReceivableBal = s.InitReceivable + s.PeriodSale - s.PeriodReceipt
		s.Debt = s.ReceivableBal - s.AdvanceBal
		if !includeZero && s.Debt == 0 && s.PeriodSale == 0 && s.PeriodReceipt == 0 {
			continue
		}
		list = append(list, *s)
	}
	sortByFloatDesc(list, func(r row) float64 { return r.Debt })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 应付查询 ====================

func (h *FinanceHandler) Payable(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("supplierKeyword")
	includeZero := c.Query("includeZero") == "true"

	type row struct {
		SupplierID   uint    `json:"supplierId"`
		Code         string  `json:"code"`
		Name         string  `json:"name"`
		InitPayable  float64 `json:"initPayable"`
		PeriodPurchase float64 `json:"periodPurchase"`
		PeriodPayment float64 `json:"periodPayment"`
		PayableBal   float64 `json:"payableBal"`
		InitAdvance  float64 `json:"initAdvance"`
		AdvanceBal   float64 `json:"advanceBal"`
		Debt         float64 `json:"debt"`
	}
	stats := map[uint]*row{}

	var suppliers []model.Supplier
	h.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Find(&suppliers)
	for _, sp := range suppliers {
		if keyword != "" && !strings.Contains(sp.Name, keyword) && !strings.Contains(sp.Code, keyword) {
			continue
		}
		stats[sp.ID] = &row{SupplierID: sp.ID, Code: sp.Code, Name: sp.Name}
	}

	var initRows []struct {
		TargetID   uint    `gorm:"column:target_id"`
		Receivable float64 `gorm:"column:receivable"`
		Advance    float64 `gorm:"column:advance"`
	}
	h.db.Raw(`SELECT target_id, receivable, advance FROM initial_balances
		WHERE company_id = ? AND biz_type = 'supplier' AND deleted_at IS NULL`, companyID).Scan(&initRows)
	for _, r := range initRows {
		if s, ok := stats[r.TargetID]; ok {
			s.InitPayable = r.Receivable
			s.InitAdvance = r.Advance
			s.AdvanceBal = r.Advance
		}
	}

	var purRows []struct {
		SupplierID uint    `gorm:"column:supplier_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT supplier_id, SUM(total_amount) AS amount FROM purchase_in_stocks
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY supplier_id`, companyID, start, end).Scan(&purRows)
	for _, r := range purRows {
		if s, ok := stats[r.SupplierID]; ok {
			s.PeriodPurchase = r.Amount
		}
	}
	var retRows []struct {
		SupplierID uint    `gorm:"column:supplier_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT supplier_id, SUM(total_amount) AS amount FROM purchase_returns
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY supplier_id`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		if s, ok := stats[r.SupplierID]; ok {
			s.PeriodPurchase -= r.Amount
		}
	}
	var payRows []struct {
		SupplierID uint    `gorm:"column:supplier_id"`
		Amount     float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT supplier_id, SUM(total_amount) AS amount FROM purchase_payments
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY supplier_id`, companyID, start, end).Scan(&payRows)
	for _, r := range payRows {
		if s, ok := stats[r.SupplierID]; ok {
			s.PeriodPayment = r.Amount
		}
	}

	list := make([]row, 0, len(stats))
	for _, s := range stats {
		s.PayableBal = s.InitPayable + s.PeriodPurchase - s.PeriodPayment
		s.Debt = s.PayableBal - s.AdvanceBal
		if !includeZero && s.Debt == 0 && s.PeriodPurchase == 0 && s.PeriodPayment == 0 {
			continue
		}
		list = append(list, *s)
	}
	sortByFloatDesc(list, func(r row) float64 { return r.Debt })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 往来余额查询 ====================

func (h *FinanceHandler) Balance(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	unitType := c.Query("unitType") // customer|supplier|''

	type row struct {
		Name      string  `json:"name"`
		Type      string  `json:"type"`
		TypeLabel string  `json:"typeLabel"`
		Receivable float64 `json:"receivable"`
		Advance   float64 `json:"advance"`
		Payable   float64 `json:"payable"`
		Prepay    float64 `json:"prepay"`
		Balance   float64 `json:"balance"`
	}
	rows := []row{}

	if unitType == "" || unitType == "customer" {
		var customers []model.Customer
		h.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Find(&customers)
		for _, cu := range customers {
			if keyword != "" && !strings.Contains(cu.Name, keyword) {
				continue
			}
			r := row{Name: cu.Name, Type: "customer", TypeLabel: "客户"}
			var init struct {
				Receivable float64 `gorm:"column:receivable"`
				Advance    float64 `gorm:"column:advance"`
			}
			h.db.Raw(`SELECT receivable, advance FROM initial_balances
				WHERE company_id = ? AND biz_type = 'customer' AND target_id = ? AND deleted_at IS NULL`,
				companyID, cu.ID).Scan(&init)
			var sale, ret, rcpt float64
			h.db.Raw(`SELECT COALESCE(SUM(total_amount),0) FROM sales_out_stocks
				WHERE company_id = ? AND customer_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?`,
				companyID, cu.ID, start, end).Scan(&sale)
			h.db.Raw(`SELECT COALESCE(SUM(total_amount),0) FROM sales_returns
				WHERE company_id = ? AND customer_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?`,
				companyID, cu.ID, start, end).Scan(&ret)
			h.db.Raw(`SELECT COALESCE(SUM(total_amount),0) FROM sales_receipts
				WHERE company_id = ? AND customer_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?`,
				companyID, cu.ID, start, end).Scan(&rcpt)
			r.Receivable = init.Receivable + sale - ret - rcpt
			r.Advance = init.Advance
			r.Balance = r.Receivable - r.Advance
			if r.Receivable != 0 || r.Advance != 0 {
				rows = append(rows, r)
			}
		}
	}

	if unitType == "" || unitType == "supplier" {
		var suppliers []model.Supplier
		h.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Find(&suppliers)
		for _, sp := range suppliers {
			if keyword != "" && !strings.Contains(sp.Name, keyword) {
				continue
			}
			r := row{Name: sp.Name, Type: "supplier", TypeLabel: "供应商"}
			var init struct {
				Receivable float64 `gorm:"column:receivable"`
				Advance    float64 `gorm:"column:advance"`
			}
			h.db.Raw(`SELECT receivable, advance FROM initial_balances
				WHERE company_id = ? AND biz_type = 'supplier' AND target_id = ? AND deleted_at IS NULL`,
				companyID, sp.ID).Scan(&init)
			var pur, ret, pay float64
			h.db.Raw(`SELECT COALESCE(SUM(total_amount),0) FROM purchase_in_stocks
				WHERE company_id = ? AND supplier_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?`,
				companyID, sp.ID, start, end).Scan(&pur)
			h.db.Raw(`SELECT COALESCE(SUM(total_amount),0) FROM purchase_returns
				WHERE company_id = ? AND supplier_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?`,
				companyID, sp.ID, start, end).Scan(&ret)
			h.db.Raw(`SELECT COALESCE(SUM(total_amount),0) FROM purchase_payments
				WHERE company_id = ? AND supplier_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?`,
				companyID, sp.ID, start, end).Scan(&pay)
			r.Payable = init.Receivable + pur - ret - pay
			r.Prepay = init.Advance
			r.Balance = r.Payable - r.Prepay
			if r.Payable != 0 || r.Prepay != 0 {
				rows = append(rows, r)
			}
		}
	}

	sortByFloatDesc(rows, func(r row) float64 { return r.Balance })
	response.Ok(c, paginateSaleRows(rows, page, pageSize))
}

// ==================== 客户预收查询 ====================

func (h *FinanceHandler) CustomerAdvance(c *gin.Context) {
	page, pageSize, _, _ := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("customerKeyword")

	type row struct {
		CustomerID  uint    `json:"customerId"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		InitAdvance float64 `json:"initAdvance"`
		PeriodAdvance float64 `json:"periodAdvance"`
		AdvanceBal  float64 `json:"advanceBal"`
	}
	rows := []row{}

	var initRows []struct {
		TargetID uint    `gorm:"column:target_id"`
		Advance  float64 `gorm:"column:advance"`
		Name     string  `gorm:"column:name"`
		Code     string  `gorm:"column:code"`
	}
	h.db.Raw(`SELECT b.target_id, b.advance, cu.name, cu.code
		FROM initial_balances b
		JOIN customers cu ON cu.id = b.target_id
		WHERE b.company_id = ? AND b.biz_type = 'customer' AND b.advance != 0 AND b.deleted_at IS NULL`, companyID).Scan(&initRows)
	for _, r := range initRows {
		if keyword != "" && !strings.Contains(r.Name, keyword) && !strings.Contains(r.Code, keyword) {
			continue
		}
		rows = append(rows, row{
			CustomerID: r.TargetID, Code: r.Code, Name: r.Name,
			InitAdvance: r.Advance, PeriodAdvance: 0, AdvanceBal: r.Advance,
		})
	}
	response.Ok(c, paginateSaleRows(rows, page, pageSize))
}

// ==================== 品牌预收/待结算（暂无品牌级预收业务，返回空结构） ====================

func (h *FinanceHandler) BrandAdvance(c *gin.Context) {
	response.Ok(c, gin.H{"list": []gin.H{}, "total": 0, "page": 1, "pageSize": 30})
}

func (h *FinanceHandler) BrandUnsettled(c *gin.Context) {
	response.Ok(c, gin.H{"list": []gin.H{}, "total": 0, "page": 1, "pageSize": 30})
}

// ==================== 单据待结算查询 ====================

func (h *FinanceHandler) UnsettledBills(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	type row struct {
		BillNo    string  `json:"billNo"`
		BillType  string  `json:"billType"`
		Customer  string  `json:"customer"`
		Amount    float64 `json:"amount"`
		Paid      float64 `json:"paid"`
		Unsettled float64 `json:"unsettled"`
		CreatedAt string  `json:"createdAt"`
	}
	rows := []row{}

	var raw []struct {
		BillNo    string  `gorm:"column:bill_no"`
		Customer  string  `gorm:"column:customer"`
		Amount    float64 `gorm:"column:amount"`
		Paid      float64 `gorm:"column:paid"`
		CreatedAt string  `gorm:"column:created_at"`
	}
	h.db.Raw(`SELECT b.bill_no, COALESCE(cu.name,'') AS customer, b.total_amount AS amount,
		b.paid_amount AS paid, b.created_at::text
		FROM sales_out_stocks b
		LEFT JOIN customers cu ON cu.id = b.customer_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.paid_amount < b.total_amount
		AND b.bill_date >= ? AND b.bill_date <= ?`, companyID, start, end).Scan(&raw)
	for _, r := range raw {
		if keyword != "" && !strings.Contains(r.BillNo, keyword) && !strings.Contains(r.Customer, keyword) {
			continue
		}
		rows = append(rows, row{
			BillNo: r.BillNo, BillType: "销售出库单", Customer: r.Customer,
			Amount: r.Amount, Paid: r.Paid, Unsettled: r.Amount - r.Paid, CreatedAt: r.CreatedAt,
		})
	}
	sortByStringDesc(rows, func(r row) string { return r.CreatedAt })
	response.Ok(c, paginateSaleRows(rows, page, pageSize))
}

// ==================== 现金银行账户统计 ====================

func (h *FinanceHandler) CashBank(c *gin.Context) {
	_, _, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)

	type row struct {
		AccountID uint    `json:"accountId"`
		Name      string  `json:"name"`
		Type      string  `json:"type"`
		InitBal   float64 `json:"initBal"`
		Income    float64 `json:"income"`
		Expense   float64 `json:"expense"`
		FinalBal  float64 `json:"finalBal"`
	}
	rows := []row{}

	var accounts []model.Account
	h.db.Where("company_id = ? AND status = 1", companyID).Order("id").Find(&accounts)
	for _, a := range accounts {
		r := row{AccountID: a.ID, Name: a.Name, Type: a.Type}
		// 期初 = 业务期初 + 期初前流水
		var init float64
		h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM initial_account_balances
			WHERE company_id = ? AND account_id = ? AND deleted_at IS NULL`, companyID, a.ID).Scan(&init)
		var preIn, preOut float64
		h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM account_flows
			WHERE company_id = ? AND account_id = ? AND type = 'income' AND created_at < ? AND deleted_at IS NULL`,
			companyID, a.ID, start).Scan(&preIn)
		h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM account_flows
			WHERE company_id = ? AND account_id = ? AND type = 'expense' AND created_at < ? AND deleted_at IS NULL`,
			companyID, a.ID, start).Scan(&preOut)
		r.InitBal = init + preIn - preOut
		// 本期
		h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM account_flows
			WHERE company_id = ? AND account_id = ? AND type = 'income' AND created_at >= ? AND created_at < ?::date + 1 AND deleted_at IS NULL`,
			companyID, a.ID, start, end).Scan(&r.Income)
		h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM account_flows
			WHERE company_id = ? AND account_id = ? AND type = 'expense' AND created_at >= ? AND created_at < ?::date + 1 AND deleted_at IS NULL`,
			companyID, a.ID, start, end).Scan(&r.Expense)
		r.FinalBal = r.InitBal + r.Income - r.Expense
		rows = append(rows, r)
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "pageSize": len(rows)})
}

// ==================== 账户收支统计 ====================

func (h *FinanceHandler) AccountIncome(c *gin.Context) {
	_, _, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	accountKeyword := c.Query("accountKeyword")
	view := c.DefaultQuery("view", "day")

	trunc := "day"
	if view == "week" {
		trunc = "week"
	} else if view == "month" {
		trunc = "month"
	}

	type row struct {
		Period   string  `json:"period"`
		Account  string  `json:"account"`
		Income   float64 `json:"income"`
		Expense  float64 `json:"expense"`
		Net      float64 `json:"net"`
	}
	var rows []row
	h.db.Raw(`SELECT date_trunc('`+trunc+`', f.created_at)::date::text AS period,
		a.name AS account,
		SUM(CASE WHEN f.type = 'income' THEN f.amount ELSE 0 END) AS income,
		SUM(CASE WHEN f.type = 'expense' THEN f.amount ELSE 0 END) AS expense
		FROM account_flows f
		JOIN accounts a ON a.id = f.account_id
		WHERE f.company_id = ? AND f.deleted_at IS NULL
		AND f.created_at >= ? AND f.created_at < ?::date + 1
		GROUP BY period, a.name
		ORDER BY period DESC, a.name`, companyID, start, end).Scan(&rows)

	list := make([]row, 0, len(rows))
	for _, r := range rows {
		if accountKeyword != "" && !strings.Contains(r.Account, accountKeyword) {
			continue
		}
		r.Net = r.Income - r.Expense
		list = append(list, r)
	}
	response.Ok(c, gin.H{"list": list, "total": len(list), "page": 1, "pageSize": len(list)})
}

// ==================== 收入费用统计 ====================

func (h *FinanceHandler) RevenueExpense(c *gin.Context) {
	_, _, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	view := c.DefaultQuery("view", "expense") // expense|income

	type row struct {
		ItemID    uint    `json:"itemId"`
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		Period    float64 `json:"period"`
		Total     float64 `json:"total"`
	}
	rows := []row{}

	var items []model.IncomeExpenseItem
	h.db.Where("company_id = ? AND type = ? AND status = 1", companyID, view).Order("sort, id").Find(&items)
	for _, it := range items {
		r := row{ItemID: it.ID, Code: it.Code, Name: it.Name}
		if view == "expense" {
			h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM expense_bills
				WHERE company_id = ? AND item_id = ? AND status = 'completed' AND bill_date >= ? AND bill_date <= ? AND deleted_at IS NULL`,
				companyID, it.ID, start, end).Scan(&r.Period)
			h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM expense_bills
				WHERE company_id = ? AND item_id = ? AND status = 'completed' AND deleted_at IS NULL`,
				companyID, it.ID).Scan(&r.Total)
		} else {
			h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM other_income_bills
				WHERE company_id = ? AND item_id = ? AND status = 'completed' AND bill_date >= ? AND bill_date <= ? AND deleted_at IS NULL`,
				companyID, it.ID, start, end).Scan(&r.Period)
			h.db.Raw(`SELECT COALESCE(SUM(amount),0) FROM other_income_bills
				WHERE company_id = ? AND item_id = ? AND status = 'completed' AND deleted_at IS NULL`,
				companyID, it.ID).Scan(&r.Total)
		}
		rows = append(rows, r)
	}
	sortByFloatDesc(rows, func(r row) float64 { return r.Period })
	response.Ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "pageSize": len(rows)})
}

// ==================== 费用/收入分布 ====================

func (h *FinanceHandler) distribution(c *gin.Context, table string) {
	_, _, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	dimension := c.DefaultQuery("dimension", "handler") // handler|dept|counterpart
	itemKeyword := c.Query("itemKeyword")

	type row struct {
		Key       string  `json:"key"`
		Name      string  `json:"name"`
		Period    float64 `json:"period"`
		Total     float64 `json:"total"`
		PeriodPct float64 `json:"periodPct"`
		TotalPct  float64 `json:"totalPct"`
	}
	rows := []row{}

	var dimExpr string
	switch dimension {
	case "dept":
		dimExpr = "COALESCE(d.name, '未分配')"
	case "counterpart":
		dimExpr = "COALESCE(NULLIF(b.counterpart,''), '未填写')"
	default:
		dimExpr = "COALESCE(e.name, '未分配')"
	}

	periodRows := []struct {
		Name   string  `gorm:"column:name"`
		Amount float64 `gorm:"column:amount"`
	}{}
	h.db.Raw(`SELECT `+dimExpr+` AS name, SUM(b.amount) AS amount
		FROM `+table+` b
		LEFT JOIN employees e ON e.id = b.handler_id
		LEFT JOIN departments d ON d.id = b.dept_id
		LEFT JOIN income_expense_items it ON it.id = b.item_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		AND (? = '' OR it.name ILIKE ?)
		GROUP BY `+dimExpr, companyID, start, end, itemKeyword, "%"+itemKeyword+"%").Scan(&periodRows)

	totalRows := []struct {
		Name   string  `gorm:"column:name"`
		Amount float64 `gorm:"column:amount"`
	}{}
	h.db.Raw(`SELECT `+dimExpr+` AS name, SUM(b.amount) AS amount
		FROM `+table+` b
		LEFT JOIN employees e ON e.id = b.handler_id
		LEFT JOIN departments d ON d.id = b.dept_id
		LEFT JOIN income_expense_items it ON it.id = b.item_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND (? = '' OR it.name ILIKE ?)
		GROUP BY `+dimExpr, companyID, itemKeyword, "%"+itemKeyword+"%").Scan(&totalRows)

	periodSum, totalSum := 0.0, 0.0
	totalMap := map[string]float64{}
	for _, r := range totalRows {
		totalMap[r.Name] = r.Amount
		totalSum += r.Amount
	}
	keys := map[string]bool{}
	for _, r := range periodRows {
		periodSum += r.Amount
		keys[r.Name] = true
	}
	for k := range totalMap {
		keys[k] = true
	}
	periodMap := map[string]float64{}
	for _, r := range periodRows {
		periodMap[r.Name] = r.Amount
	}
	for k := range keys {
		r := row{Key: k, Name: k, Period: periodMap[k], Total: totalMap[k]}
		if periodSum != 0 {
			r.PeriodPct = r.Period / periodSum * 100
		}
		if totalSum != 0 {
			r.TotalPct = r.Total / totalSum * 100
		}
		rows = append(rows, r)
	}
	sortByFloatDesc(rows, func(r row) float64 { return r.Period })
	response.Ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "pageSize": len(rows)})
}

func (h *FinanceHandler) ExpenseDistribution(c *gin.Context) {
	h.distribution(c, "expense_bills")
}

func (h *FinanceHandler) IncomeDistribution(c *gin.Context) {
	h.distribution(c, "other_income_bills")
}

// ==================== 经营利润统计 ====================

func (h *FinanceHandler) OperatingProfit(c *gin.Context) {
	_, _, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	view := c.DefaultQuery("view", "day")

	trunc := "day"
	if view == "week" {
		trunc = "week"
	} else if view == "month" {
		trunc = "month"
	}

	type row struct {
		Period       string  `json:"period"`
		SaleIncome   float64 `json:"saleIncome"`
		OtherIncome  float64 `json:"otherIncome"`
		IncomeTotal  float64 `json:"incomeTotal"`
		SaleCost     float64 `json:"saleCost"`
		Expense      float64 `json:"expense"`
		ExpenseTotal float64 `json:"expenseTotal"`
		Profit       float64 `json:"profit"`
	}
	periods := map[string]*row{}
	get := func(p string) *row {
		if r, ok := periods[p]; ok {
			return r
		}
		r := &row{Period: p}
		periods[p] = r
		return r
	}

	// 销售收入（出库 - 退货）
	var saleRows []struct {
		Period string  `gorm:"column:period"`
		Amount float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT date_trunc('`+trunc+`', bill_date)::date::text AS period, SUM(total_amount) AS amount
		FROM sales_out_stocks
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY period`, companyID, start, end).Scan(&saleRows)
	for _, r := range saleRows {
		get(r.Period).SaleIncome = r.Amount
	}
	var retRows []struct {
		Period string  `gorm:"column:period"`
		Amount float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT date_trunc('`+trunc+`', bill_date)::date::text AS period, SUM(total_amount) AS amount
		FROM sales_returns
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY period`, companyID, start, end).Scan(&retRows)
	for _, r := range retRows {
		get(r.Period).SaleIncome -= r.Amount
	}

	// 其他收入
	var incomeRows []struct {
		Period string  `gorm:"column:period"`
		Amount float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT date_trunc('`+trunc+`', bill_date)::date::text AS period, SUM(amount) AS amount
		FROM other_income_bills
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY period`, companyID, start, end).Scan(&incomeRows)
	for _, r := range incomeRows {
		get(r.Period).OtherIncome = r.Amount
	}

	// 销售成本
	var costRows []struct {
		Period string  `gorm:"column:period"`
		Amount float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT date_trunc('`+trunc+`', b.bill_date)::date::text AS period, SUM(i.quantity * p.purchase_price) AS amount
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY period`, companyID, start, end).Scan(&costRows)
	for _, r := range costRows {
		get(r.Period).SaleCost = r.Amount
	}
	var retCostRows []struct {
		Period string  `gorm:"column:period"`
		Amount float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT date_trunc('`+trunc+`', b.bill_date)::date::text AS period, SUM(i.quantity * p.purchase_price) AS amount
		FROM sales_return_items i
		JOIN sales_returns b ON b.id = i.return_id
		JOIN products p ON p.id = i.product_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL AND b.bill_date >= ? AND b.bill_date <= ?
		GROUP BY period`, companyID, start, end).Scan(&retCostRows)
	for _, r := range retCostRows {
		get(r.Period).SaleCost -= r.Amount
	}

	// 费用
	var expRows []struct {
		Period string  `gorm:"column:period"`
		Amount float64 `gorm:"column:amount"`
	}
	h.db.Raw(`SELECT date_trunc('`+trunc+`', bill_date)::date::text AS period, SUM(amount) AS amount
		FROM expense_bills
		WHERE company_id = ? AND status = 'completed' AND deleted_at IS NULL AND bill_date >= ? AND bill_date <= ?
		GROUP BY period`, companyID, start, end).Scan(&expRows)
	for _, r := range expRows {
		get(r.Period).Expense = r.Amount
	}

	list := make([]row, 0, len(periods))
	for _, r := range periods {
		r.IncomeTotal = r.SaleIncome + r.OtherIncome
		r.ExpenseTotal = r.SaleCost + r.Expense
		r.Profit = r.IncomeTotal - r.ExpenseTotal
		list = append(list, *r)
	}
	sortByStringDesc(list, func(r row) string { return r.Period })
	response.Ok(c, gin.H{"list": list, "total": len(list), "page": 1, "pageSize": len(list)})
}
