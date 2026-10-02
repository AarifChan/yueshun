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

// CRMExtHandler CRM 增强（客户公海/汇报/模板/自定义字段/供应商分类/业务驾驶舱）
type CRMExtHandler struct {
	db *gorm.DB
}

func NewCRMExtHandler(db *gorm.DB) *CRMExtHandler {
	return &CRMExtHandler{db: db}
}

func (h *CRMExtHandler) RegisterRoutes(r *gin.RouterGroup) {
	sea := r.Group("/crm/sea")
	{
		sea.GET("/customers", h.SeaCustomers)
		sea.POST("/recycle", h.SeaRecycle)
		sea.POST("/claim", h.SeaClaim)
		sea.POST("/assign", h.SeaAssign)
	}
	sc := r.Group("/supplier-categories")
	{
		sc.GET("", h.SupplierCategoryList)
		sc.POST("", h.SupplierCategoryCreate)
		sc.PUT("/:id", h.SupplierCategoryUpdate)
		sc.DELETE("/:id", h.SupplierCategoryDelete)
	}
	wr := r.Group("/work-reports")
	{
		wr.GET("", h.WorkReportList)
		wr.POST("", h.WorkReportCreate)
		wr.DELETE("/:id", h.WorkReportDelete)
		wr.GET("/stats", h.WorkReportStats)
	}
	rt := r.Group("/report-templates")
	{
		rt.GET("", h.TemplateList)
		rt.POST("", h.TemplateCreate)
		rt.PUT("/:id", h.TemplateUpdate)
		rt.DELETE("/:id", h.TemplateDelete)
	}
	ot := r.Group("/opportunity-templates")
	{
		ot.GET("", h.OppTemplateList)
		ot.POST("", h.OppTemplateCreate)
		ot.PUT("/:id", h.OppTemplateUpdate)
		ot.DELETE("/:id", h.OppTemplateDelete)
	}
	cf := r.Group("/custom-fields")
	{
		cf.GET("", h.CustomFieldList)
		cf.POST("", h.CustomFieldCreate)
		cf.PUT("/:id", h.CustomFieldUpdate)
		cf.DELETE("/:id", h.CustomFieldDelete)
	}
	sp := r.Group("/sales-plans")
	{
		sp.GET("", h.SalesPlanList)
		sp.PUT("", h.SalesPlanSave)
	}
	ct := r.Group("/crm/customer-tags")
	{
		ct.GET("", h.CustomerTagList)
		ct.POST("", h.CustomerTagCreate)
		ct.PUT("/:id", h.CustomerTagUpdate)
		ct.DELETE("/:id", h.CustomerTagDelete)
	}
}

// ==================== 客户公海 ====================

func (h *CRMExtHandler) seaCustomerRow(cu model.Customer) gin.H {
	var category, region, owner string
	if cu.CategoryID > 0 {
		h.db.Model(&model.CustomerCategory{}).Where("id = ?", cu.CategoryID).Select("name").Scan(&category)
	}
	if cu.RegionID > 0 {
		h.db.Model(&model.Region{}).Where("id = ?", cu.RegionID).Select("name").Scan(&region)
	}
	if cu.OwnerID > 0 {
		h.db.Model(&model.Employee{}).Where("id = ?", cu.OwnerID).Select("name").Scan(&owner)
	}
	// 最后跟进人
	var lastFollower string
	h.db.Raw(`SELECT e.name FROM follow_ups f JOIN employees e ON e.id = f.employee_id
		WHERE f.customer_id = ? ORDER BY f.created_at DESC LIMIT 1`, cu.ID).Scan(&lastFollower)
	row := gin.H{
		"id": cu.ID, "name": cu.Name, "code": cu.Code, "contact": cu.Contact, "phone": cu.Phone,
		"category": category, "region": region, "address": cu.Address, "ownerName": owner,
		"recycleCount": cu.RecycleCount, "claimCount": cu.ClaimCount,
		"lastRecycleType": cu.LastRecycleType, "lastRecycleReason": cu.LastRecycleReason,
		"lastFollower": lastFollower, "seaStatus": cu.SeaStatus,
	}
	if cu.LastRecycleAt != nil {
		row["lastRecycleAt"] = cu.LastRecycleAt.Format("2006-01-02 15:04")
	}
	if cu.LastClaimAt != nil {
		row["lastClaimAt"] = cu.LastClaimAt.Format("2006-01-02 15:04")
	}
	return row
}

func (h *CRMExtHandler) SeaCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	tab := c.DefaultQuery("tab", "all")
	keyword := c.Query("keyword")

	q := h.db.Model(&model.Customer{}).Where("company_id = ? AND deleted_at IS NULL", companyID)
	switch tab {
	case "noTrade": // 规定时间未有交易（90 天）
		cutoff := time.Now().AddDate(0, 0, -90)
		q = q.Where("sea_status = 'public' AND (last_order_at IS NULL OR last_order_at < ?)", cutoff)
	case "pending": // 待领取/分配：在公海且回收后未被领取
		q = q.Where("sea_status = 'public' AND (last_claim_at IS NULL OR last_claim_at < last_recycle_at)")
	case "claimed": // 已领取/分配
		q = q.Where("claim_count > 0")
	default: // 全部公海客户
		q = q.Where("sea_status = 'public'")
	}
	if keyword != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ? OR contact ILIKE ? OR phone ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Customer
	q.Order("last_recycle_at DESC NULLS LAST").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	rows := make([]gin.H, 0, len(list))
	for _, cu := range list {
		rows = append(rows, h.seaCustomerRow(cu))
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *CRMExtHandler) SeaRecycle(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		CustomerID uint   `json:"customerId" binding:"required"`
		Reason     string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	var cu model.Customer
	if err := h.db.Where("id = ? AND company_id = ?", req.CustomerID, companyID).First(&cu).Error; err != nil {
		response.Fail(c, 4004, "客户不存在")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{
		"sea_status": "public", "recycle_count": cu.RecycleCount + 1,
		"last_recycle_at": &now, "last_recycle_type": "手动回收", "last_recycle_reason": req.Reason,
		"owner_id": 0,
	}
	if err := h.db.Model(&cu).Updates(updates).Error; err != nil {
		response.Fail(c, 4000, "回收失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *CRMExtHandler) SeaClaim(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	var req struct {
		CustomerID uint `json:"customerId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	var cu model.Customer
	if err := h.db.Where("id = ? AND company_id = ? AND sea_status = 'public'", req.CustomerID, companyID).First(&cu).Error; err != nil {
		response.Fail(c, 4004, "公海客户不存在")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{
		"sea_status": "private", "claim_count": cu.ClaimCount + 1, "last_claim_at": &now, "owner_id": userID,
	}
	if err := h.db.Model(&cu).Updates(updates).Error; err != nil {
		response.Fail(c, 4000, "领取失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *CRMExtHandler) SeaAssign(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		CustomerID uint `json:"customerId" binding:"required"`
		EmployeeID uint `json:"employeeId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	var cu model.Customer
	if err := h.db.Where("id = ? AND company_id = ? AND sea_status = 'public'", req.CustomerID, companyID).First(&cu).Error; err != nil {
		response.Fail(c, 4004, "公海客户不存在")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{
		"sea_status": "private", "claim_count": cu.ClaimCount + 1, "last_claim_at": &now, "owner_id": req.EmployeeID,
	}
	if err := h.db.Model(&cu).Updates(updates).Error; err != nil {
		response.Fail(c, 4000, "分配失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

// ==================== 供应商分类 ====================

func (h *CRMExtHandler) SupplierCategoryList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	q := h.db.Model(&model.SupplierCategory{}).Where("company_id = ?", companyID)
	if keyword != "" {
		q = q.Where("name ILIKE ?", "%"+keyword+"%")
	}
	var list []model.SupplierCategory
	q.Order("sort, id").Find(&list)
	rows := make([]gin.H, 0, len(list))
	for _, sc := range list {
		var count int64
		h.db.Model(&model.Supplier{}).Where("category_id = ? AND company_id = ?", sc.ID, companyID).Count(&count)
		rows = append(rows, gin.H{
			"id": sc.ID, "name": sc.Name, "code": sc.Code, "sort": sc.Sort,
			"status": sc.Status, "supplierCount": count, "createdAt": sc.CreatedAt,
		})
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}

func (h *CRMExtHandler) SupplierCategoryCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var sc model.SupplierCategory
	if err := c.ShouldBindJSON(&sc); err != nil || sc.Name == "" {
		response.Fail(c, 4000, "分类名称不能为空")
		return
	}
	sc.ID = 0
	sc.CompanyID = companyID
	if err := h.db.Create(&sc).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, sc)
}

func (h *CRMExtHandler) SupplierCategoryUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var sc model.SupplierCategory
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&sc).Error; err != nil {
		response.Fail(c, 4004, "分类不存在")
		return
	}
	var req model.SupplierCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	sc.Name = req.Name
	sc.Code = req.Code
	sc.Sort = req.Sort
	sc.Status = req.Status
	h.db.Save(&sc)
	response.Ok(c, sc)
}

func (h *CRMExtHandler) SupplierCategoryDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var count int64
	h.db.Model(&model.Supplier{}).Where("category_id = ? AND company_id = ?", id, companyID).Count(&count)
	if count > 0 {
		response.Fail(c, 4000, "该分类下还有供应商，无法删除")
		return
	}
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.SupplierCategory{})
	response.Ok(c, nil)
}

// ==================== 工作汇报 ====================

func (h *CRMExtHandler) workReportRow(r model.WorkReport) gin.H {
	var author string
	h.db.Model(&model.Employee{}).Where("id = ?", r.AuthorID).Select("name").Scan(&author)
	ccNames := []string{}
	if r.CcIDs != "" {
		var names []string
		h.db.Raw(`SELECT name FROM employees WHERE id = ANY(string_to_array(?, ',')::int[])`, r.CcIDs).Scan(&names)
		ccNames = names
	}
	typeLabel := map[string]string{"log": "日志", "week": "周报", "month": "月报"}[r.Type]
	return gin.H{
		"id": r.ID, "type": r.Type, "typeLabel": typeLabel,
		"reportDate": r.ReportDate.Format("2006-01-02"), "content": r.Content,
		"authorId": r.AuthorID, "authorName": author, "ccNames": ccNames,
		"createdAt": r.CreatedAt,
	}
}

func (h *CRMExtHandler) WorkReportList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	tab := c.DefaultQuery("tab", "mine") // mine|cc|others
	reportType := c.Query("type")
	date := c.Query("date")

	q := h.db.Model(&model.WorkReport{}).Where("company_id = ?", companyID)
	switch tab {
	case "cc":
		q = q.Where("',' || COALESCE(cc_ids,'') || ',' LIKE ?", "%,"+strconv.Itoa(int(userID))+",%")
	case "others":
		q = q.Where("author_id != ?", userID)
	default:
		q = q.Where("author_id = ?", userID)
	}
	if reportType != "" {
		q = q.Where("type = ?", reportType)
	}
	if date != "" {
		q = q.Where("report_date = ?", date)
	}
	var total int64
	q.Count(&total)
	var list []model.WorkReport
	q.Order("report_date DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	rows := make([]gin.H, 0, len(list))
	for _, r := range list {
		rows = append(rows, h.workReportRow(r))
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *CRMExtHandler) WorkReportCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	var req struct {
		Type       string `json:"type" binding:"required"`
		ReportDate string `json:"reportDate" binding:"required"`
		Content    string `json:"content" binding:"required"`
		CcIDs      string `json:"ccIds"`
		TemplateID uint   `json:"templateId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	if req.Type != "log" && req.Type != "week" && req.Type != "month" {
		response.Fail(c, 4000, "汇报类型必须是 log/week/month")
		return
	}
	reportDate, _ := time.Parse("2006-01-02", req.ReportDate)
	report := model.WorkReport{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Type:                 req.Type,
		ReportDate:           reportDate,
		Content:              req.Content,
		AuthorID:             userID,
		CcIDs:                req.CcIDs,
		TemplateID:           req.TemplateID,
	}
	if err := h.db.Create(&report).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, h.workReportRow(report))
}

func (h *CRMExtHandler) WorkReportDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Where("id = ? AND company_id = ? AND author_id = ?", id, companyID, userID).Delete(&model.WorkReport{})
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "汇报不存在或无权限删除")
		return
	}
	response.Ok(c, nil)
}

func (h *CRMExtHandler) WorkReportStats(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	type row struct {
		EmployeeID uint   `gorm:"column:employee_id" json:"employeeId"`
		Name       string `gorm:"column:name" json:"name"`
		LogCount   int    `gorm:"column:log_count" json:"logCount"`
		WeekCount  int    `gorm:"column:week_count" json:"weekCount"`
		MonthCount int    `gorm:"column:month_count" json:"monthCount"`
		Total      int    `gorm:"column:total" json:"total"`
	}
	var rows []row
	h.db.Raw(`SELECT e.id AS employee_id, e.name,
		SUM(CASE WHEN r.type = 'log' THEN 1 ELSE 0 END) AS log_count,
		SUM(CASE WHEN r.type = 'week' THEN 1 ELSE 0 END) AS week_count,
		SUM(CASE WHEN r.type = 'month' THEN 1 ELSE 0 END) AS month_count,
		COUNT(r.id) AS total
		FROM work_reports r
		JOIN employees e ON e.id = r.author_id
		WHERE r.company_id = ? AND to_char(r.report_date, 'YYYY-MM') = ? AND r.deleted_at IS NULL
		GROUP BY e.id, e.name
		ORDER BY total DESC`, companyID, month).Scan(&rows)
	response.Ok(c, gin.H{"list": rows, "total": len(rows), "month": month})
}

// ==================== 汇报模板 ====================

func (h *CRMExtHandler) TemplateList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	status := c.DefaultQuery("status", "1")
	q := h.db.Model(&model.ReportTemplate{}).Where("company_id = ?", companyID)
	if status != "all" {
		q = q.Where("status = ?", status)
	}
	var list []model.ReportTemplate
	q.Order("sort, id").Find(&list)
	response.Ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *CRMExtHandler) TemplateCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var t model.ReportTemplate
	if err := c.ShouldBindJSON(&t); err != nil || t.Name == "" {
		response.Fail(c, 4000, "模板名称不能为空")
		return
	}
	t.ID = 0
	t.CompanyID = companyID
	if err := h.db.Create(&t).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, t)
}

func (h *CRMExtHandler) TemplateUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var t model.ReportTemplate
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&t).Error; err != nil {
		response.Fail(c, 4004, "模板不存在")
		return
	}
	var req model.ReportTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	t.Category = req.Category
	t.Name = req.Name
	t.Description = req.Description
	t.UserIDs = req.UserIDs
	t.Status = req.Status
	t.Sort = req.Sort
	h.db.Save(&t)
	response.Ok(c, t)
}

func (h *CRMExtHandler) TemplateDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.ReportTemplate{})
	response.Ok(c, nil)
}

// ==================== 商机设置模板 ====================

func (h *CRMExtHandler) OppTemplateList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var list []model.OpportunityTemplate
	h.db.Where("company_id = ?", companyID).Order("sort, id").Find(&list)
	response.Ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *CRMExtHandler) OppTemplateCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var t model.OpportunityTemplate
	if err := c.ShouldBindJSON(&t); err != nil || t.Name == "" {
		response.Fail(c, 4000, "模板名称不能为空")
		return
	}
	t.ID = 0
	t.CompanyID = companyID
	if err := h.db.Create(&t).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, t)
}

func (h *CRMExtHandler) OppTemplateUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var t model.OpportunityTemplate
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&t).Error; err != nil {
		response.Fail(c, 4004, "模板不存在")
		return
	}
	var req model.OpportunityTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	t.Name = req.Name
	t.Stages = req.Stages
	t.UserIDs = req.UserIDs
	t.Description = req.Description
	t.Status = req.Status
	t.Sort = req.Sort
	h.db.Save(&t)
	response.Ok(c, t)
}

func (h *CRMExtHandler) OppTemplateDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.OpportunityTemplate{})
	response.Ok(c, nil)
}

// ==================== 客户自定义字段 ====================

func (h *CRMExtHandler) CustomFieldList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	scope := c.DefaultQuery("scope", "customer")
	keyword := c.Query("keyword")
	q := h.db.Where("company_id = ? AND scope = ?", companyID, scope)
	if keyword != "" {
		q = q.Where("name ILIKE ?", "%"+keyword+"%")
	}
	var list []model.CustomFieldDef
	q.Order("sort, id").Find(&list)
	response.Ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *CRMExtHandler) CustomFieldCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var f model.CustomFieldDef
	if err := c.ShouldBindJSON(&f); err != nil || f.Name == "" {
		response.Fail(c, 4000, "字段名称不能为空")
		return
	}
	f.ID = 0
	f.CompanyID = companyID
	if f.Scope == "" {
		f.Scope = "customer"
	}
	if err := h.db.Create(&f).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, f)
}

func (h *CRMExtHandler) CustomFieldUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var f model.CustomFieldDef
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&f).Error; err != nil {
		response.Fail(c, 4004, "字段不存在")
		return
	}
	var req model.CustomFieldDef
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	f.Name = req.Name
	f.FieldType = req.FieldType
	f.Options = req.Options
	f.Enabled = req.Enabled
	f.Required = req.Required
	f.DefaultValue = req.DefaultValue
	f.ShowInDetail = req.ShowInDetail
	f.ShowInList = req.ShowInList
	f.Description = req.Description
	f.Sort = req.Sort
	h.db.Save(&f)
	response.Ok(c, f)
}

func (h *CRMExtHandler) CustomFieldDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.CustomFieldDef{})
	response.Ok(c, nil)
}

// ==================== 业务驾驶舱（年度销售计划） ====================

func (h *CRMExtHandler) SalesPlanList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	employeeKeyword := c.Query("employeeKeyword")

	type row struct {
		EmployeeID uint     `json:"employeeId"`
		Name       string   `json:"name"`
		Dept       string   `json:"dept"`
		PlanID     uint     `json:"planId"`
		Months     [12]float64 `json:"months"`
		Total      float64  `json:"total"`
	}
	rows := []row{}

	var employees []model.Employee
	h.db.Where("company_id = ? AND status = 1", companyID).Order("id").Find(&employees)
	plans := map[uint]model.SalesPlan{}
	var planList []model.SalesPlan
	h.db.Where("company_id = ? AND year = ?", companyID, year).Find(&planList)
	for _, p := range planList {
		plans[p.EmployeeID] = p
	}
	for _, e := range employees {
		if employeeKeyword != "" && !strings.Contains(e.Name, employeeKeyword) {
			continue
		}
		var dept string
		h.db.Model(&model.Department{}).Where("id = ?", e.DeptID).Select("name").Scan(&dept)
		r := row{EmployeeID: e.ID, Name: e.Name, Dept: dept}
		if p, ok := plans[e.ID]; ok {
			r.PlanID = p.ID
			r.Months = [12]float64{p.M1, p.M2, p.M3, p.M4, p.M5, p.M6, p.M7, p.M8, p.M9, p.M10, p.M11, p.M12}
		}
		for _, m := range r.Months {
			r.Total += m
		}
		rows = append(rows, r)
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows), "year": year})
}

func (h *CRMExtHandler) SalesPlanSave(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		Year       int          `json:"year" binding:"required"`
		EmployeeID uint         `json:"employeeId" binding:"required"`
		Months     [12]float64  `json:"months" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	var plan model.SalesPlan
	err := h.db.Where("company_id = ? AND year = ? AND employee_id = ?", companyID, req.Year, req.EmployeeID).First(&plan).Error
	if err != nil {
		plan = model.SalesPlan{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Year:                 req.Year,
			EmployeeID:           req.EmployeeID,
		}
	}
	plan.M1, plan.M2, plan.M3, plan.M4 = req.Months[0], req.Months[1], req.Months[2], req.Months[3]
	plan.M5, plan.M6, plan.M7, plan.M8 = req.Months[4], req.Months[5], req.Months[6], req.Months[7]
	plan.M9, plan.M10, plan.M11, plan.M12 = req.Months[8], req.Months[9], req.Months[10], req.Months[11]
	if plan.ID == 0 {
		err = h.db.Create(&plan).Error
	} else {
		err = h.db.Save(&plan).Error
	}
	if err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, plan)
}

// ==================== 客户标签 ====================

func (h *CRMExtHandler) CustomerTagList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	q := h.db.Model(&model.CustomerTag{}).Where("company_id = ?", companyID)
	if keyword != "" {
		q = q.Where("name ILIKE ?", "%"+keyword+"%")
	}
	var list []model.CustomerTag
	q.Order("id").Find(&list)
	rows := make([]gin.H, 0, len(list))
	for _, t := range list {
		rows = append(rows, gin.H{
			"id": t.ID, "name": t.Name, "status": t.Status,
			"tagType": "手动标签", "customerCount": 0, "createdAt": t.CreatedAt,
		})
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}

func (h *CRMExtHandler) CustomerTagCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var t model.CustomerTag
	if err := c.ShouldBindJSON(&t); err != nil || t.Name == "" {
		response.Fail(c, 4000, "标签名称不能为空")
		return
	}
	t.ID = 0
	t.CompanyID = companyID
	if err := h.db.Create(&t).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, t)
}

func (h *CRMExtHandler) CustomerTagUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var t model.CustomerTag
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&t).Error; err != nil {
		response.Fail(c, 4004, "标签不存在")
		return
	}
	var req model.CustomerTag
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	t.Name = req.Name
	t.Status = req.Status
	h.db.Save(&t)
	response.Ok(c, t)
}

func (h *CRMExtHandler) CustomerTagDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.CustomerTag{})
	response.Ok(c, nil)
}
