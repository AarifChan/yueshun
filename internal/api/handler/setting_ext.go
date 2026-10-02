package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// SettingExtHandler 设置完善（操作日志/专题分类/打印机/关联商品/职员权限/应用授权/企业信息）
type SettingExtHandler struct {
	db *gorm.DB
}

func NewSettingExtHandler(db *gorm.DB) *SettingExtHandler {
	return &SettingExtHandler{db: db}
}

func (h *SettingExtHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/operation-logs", h.OperationLogList)

	sc := r.Group("/subject-categories")
	{
		sc.GET("", h.SubjectCategoryList)
		sc.POST("", h.SubjectCategoryCreate)
		sc.PUT("/:id", h.SubjectCategoryUpdate)
		sc.DELETE("/:id", h.SubjectCategoryDelete)
	}
	pr := r.Group("/printers")
	{
		pr.GET("", h.PrinterList)
		pr.POST("", h.PrinterCreate)
		pr.PUT("/:id", h.PrinterUpdate)
		pr.DELETE("/:id", h.PrinterDelete)
	}
	r.GET("/product-relations", h.ProductRelationList)
	r.PUT("/product-relations", h.ProductRelationSave)

	ep := r.Group("/employee-permissions")
	{
		ep.GET("", h.EmployeePermissionList)
		ep.PUT("/:employeeId", h.EmployeePermissionSave)
	}
	r.GET("/app-licenses", h.AppLicenseList)
	r.GET("/company-info", h.CompanyInfoGet)
	r.PUT("/company-info", h.CompanyInfoSave)
}

// ==================== 操作日志 ====================

func (h *SettingExtHandler) OperationLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	objectType := c.Query("objectType")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	q := h.db.Model(&model.OperationLog{}).Where("company_id = ?", companyID)
	if keyword != "" {
		q = q.Where("detail ILIKE ?", "%"+keyword+"%")
	}
	if objectType != "" {
		q = q.Where("object_type = ?", objectType)
	}
	if startDate != "" {
		q = q.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		q = q.Where("created_at <= ?", endDate+" 23:59:59")
	}
	var total int64
	q.Count(&total)
	var list []model.OperationLog
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	empNames := map[uint]string{}
	var emps []model.Employee
	h.db.Where("company_id = ?", companyID).Find(&emps)
	for _, e := range emps {
		empNames[e.ID] = e.Name
	}

	rows := make([]gin.H, 0, len(list))
	for _, x := range list {
		rows = append(rows, gin.H{
			"id": x.ID, "objectType": x.ObjectType, "action": x.Action,
			"userId": x.UserID, "operator": empNames[x.UserID],
			"ip": x.IP, "source": x.Source, "detail": x.Detail, "createdAt": x.CreatedAt,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

// ==================== 专题分类 ====================

func (h *SettingExtHandler) SubjectCategoryList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var list []model.SubjectCategory
	h.db.Where("company_id = ?", companyID).Order("sort, id").Find(&list)
	rows := make([]gin.H, 0, len(list))
	for _, x := range list {
		rows = append(rows, gin.H{
			"id": x.ID, "name": x.Name, "image": x.Image,
			"visibleScope": x.VisibleScope, "showInMall": x.ShowInMall,
			"productIds": x.ProductIDs, "productCount": len(parseIDList(x.ProductIDs)),
			"sort": x.Sort, "status": x.Status,
		})
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}

func (h *SettingExtHandler) SubjectCategoryCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var x model.SubjectCategory
	if err := c.ShouldBindJSON(&x); err != nil || x.Name == "" {
		response.Fail(c, 4000, "专题分类名称不能为空")
		return
	}
	x.ID = 0
	x.CompanyID = companyID
	if err := h.db.Create(&x).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, x)
}

func (h *SettingExtHandler) SubjectCategoryUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var x model.SubjectCategory
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&x).Error; err != nil {
		response.Fail(c, 4004, "专题分类不存在")
		return
	}
	var req model.SubjectCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	req.ID = x.ID
	req.CompanyID = companyID
	h.db.Save(&req)
	response.Ok(c, req)
}

func (h *SettingExtHandler) SubjectCategoryDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.SubjectCategory{})
	response.Ok(c, nil)
}

// ==================== 打印机 ====================

func (h *SettingExtHandler) PrinterList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var list []model.Printer
	h.db.Where("company_id = ?", companyID).Order("id").Find(&list)
	response.Ok(c, gin.H{"list": list, "total": len(list)})
}

func (h *SettingExtHandler) PrinterCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var x model.Printer
	if err := c.ShouldBindJSON(&x); err != nil || x.Name == "" {
		response.Fail(c, 4000, "打印机名称不能为空")
		return
	}
	x.ID = 0
	x.CompanyID = companyID
	if err := h.db.Create(&x).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, x)
}

func (h *SettingExtHandler) PrinterUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var x model.Printer
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&x).Error; err != nil {
		response.Fail(c, 4004, "打印机不存在")
		return
	}
	var req model.Printer
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	x.Name = req.Name
	x.Type = req.Type
	x.DeviceNo = req.DeviceNo
	x.IsDefault = req.IsDefault
	x.Status = req.Status
	x.Remark = req.Remark
	h.db.Save(&x)
	response.Ok(c, x)
}

func (h *SettingExtHandler) PrinterDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Printer{})
	response.Ok(c, nil)
}

// ==================== 关联商品 ====================

// ProductRelationList 查询某商品的关联商品（含商品信息）
func (h *SettingExtHandler) ProductRelationList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	productID, _ := strconv.Atoi(c.DefaultQuery("productId", "0"))
	if productID == 0 {
		response.Fail(c, 4000, "请指定商品")
		return
	}
	type row struct {
		ID        uint   `gorm:"column:id" json:"id"`
		RelatedID uint   `gorm:"column:related_product_id" json:"relatedProductId"`
		Name      string `gorm:"column:name" json:"name"`
		Code      string `gorm:"column:code" json:"code"`
		Spec      string `gorm:"column:spec" json:"spec"`
		Image     string `gorm:"column:image" json:"image"`
		Category  string `gorm:"column:category" json:"category"`
		Brand     string `gorm:"column:brand" json:"brand"`
		Sort      int    `gorm:"column:sort" json:"sort"`
	}
	var rows []row
	h.db.Raw(`SELECT r.id, r.related_product_id, p.name, p.code, p.specification AS spec, p.image,
		COALESCE(pc.name,'') AS category, COALESCE(b.name,'') AS brand, r.sort
		FROM product_relations r
		JOIN products p ON p.id = r.related_product_id
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN brands b ON b.id = p.brand_id
		WHERE r.company_id = ? AND r.product_id = ? AND r.deleted_at IS NULL
		ORDER BY r.sort, r.id`, companyID, productID).Scan(&rows)
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}

// ProductRelationSave 批量设置某商品的关联商品
func (h *SettingExtHandler) ProductRelationSave(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		ProductID uint   `json:"productId" binding:"required"`
		RelatedIDs []uint `json:"relatedIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ? AND product_id = ?", companyID, req.ProductID).
			Delete(&model.ProductRelation{}).Error; err != nil {
			return err
		}
		for i, rid := range req.RelatedIDs {
			if rid == req.ProductID {
				continue
			}
			rel := model.ProductRelation{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				ProductID:            req.ProductID,
				RelatedProductID:     rid,
				Sort:                 i,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

// ==================== 职员权限 ====================

func (h *SettingExtHandler) EmployeePermissionList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	eq := h.db.Model(&model.Employee{}).Where("company_id = ? AND deleted_at IS NULL", companyID)
	if keyword != "" {
		eq = eq.Where("name ILIKE ? OR username ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	eq.Count(&total)
	var emps []model.Employee
	eq.Order("id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&emps)

	roleNames := map[uint]string{}
	var roles []model.Role
	h.db.Where("company_id = ?", companyID).Find(&roles)
	for _, r := range roles {
		roleNames[r.ID] = r.Name
	}
	perms := map[uint]model.EmployeePermission{}
	var permList []model.EmployeePermission
	h.db.Where("company_id = ?", companyID).Find(&permList)
	for _, p := range permList {
		perms[p.EmployeeID] = p
	}

	rows := make([]gin.H, 0, len(emps))
	for _, e := range emps {
		p := perms[e.ID]
		rows = append(rows, gin.H{
			"employeeId": e.ID, "name": e.Name, "username": e.Username,
			"roleName": roleNames[e.RoleID],
			"appScopes": p.AppScopes, "customerScope": valueOrDefault(p.CustomerScope, "all"),
			"employeeScope": valueOrDefault(p.EmployeeScope, "all"),
			"warehouseScope": p.WarehouseScope, "supplierScope": valueOrDefault(p.SupplierScope, "all"),
			"productScope": valueOrDefault(p.ProductScope, "all"),
			"accountScope": p.AccountScope, "loginTimeRange": p.LoginTimeRange,
			"licenseType": valueOrDefault(p.LicenseType, "normal"),
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func valueOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func (h *SettingExtHandler) EmployeePermissionSave(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	employeeID, _ := strconv.Atoi(c.Param("employeeId"))
	var req model.EmployeePermission
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	var p model.EmployeePermission
	err := h.db.Where("company_id = ? AND employee_id = ?", companyID, employeeID).First(&p).Error
	if err != nil {
		p = model.EmployeePermission{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			EmployeeID:           uint(employeeID),
		}
	}
	p.AppScopes = req.AppScopes
	p.CustomerScope = req.CustomerScope
	p.EmployeeScope = req.EmployeeScope
	p.WarehouseScope = req.WarehouseScope
	p.SupplierScope = req.SupplierScope
	p.ProductScope = req.ProductScope
	p.AccountScope = req.AccountScope
	p.LoginTimeRange = req.LoginTimeRange
	p.LicenseType = req.LicenseType
	if p.ID == 0 {
		err = h.db.Create(&p).Error
	} else {
		err = h.db.Save(&p).Error
	}
	if err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, p)
}

// ==================== 应用授权 ====================

// AppLicenseList 应用授权列表（内置模块，授权用户数=职员数）
func (h *SettingExtHandler) AppLicenseList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var empCount int64
	h.db.Model(&model.Employee{}).Where("company_id = ? AND deleted_at IS NULL", companyID).Count(&empCount)

	var expireSetting model.CompanySetting
	expire := ""
	if err := h.db.Where("company_id = ? AND key = ?", companyID, "app.license.expire").First(&expireSetting).Error; err == nil {
		expire = expireSetting.Value
	}

	apps := []string{"进销存", "订货商城", "CRM", "营销", "BI 分析", "生态互联"}
	rows := make([]gin.H, 0, len(apps))
	for _, name := range apps {
		rows = append(rows, gin.H{
			"appName": name, "expireAt": expire,
			"purchasedUsers": empCount, "grantedUsers": empCount,
		})
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}

// ==================== 企业信息 ====================

func (h *SettingExtHandler) CompanyInfoGet(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var company model.Company
	if err := h.db.First(&company, companyID).Error; err != nil {
		response.Fail(c, 4004, "企业不存在")
		return
	}
	// 扩展信息存 settings
	extra := map[string]string{}
	var settings []model.CompanySetting
	h.db.Where("company_id = ? AND key LIKE ?", companyID, "company.%").Find(&settings)
	for _, s := range settings {
		extra[strings.TrimPrefix(s.Key, "company.")] = s.Value
	}
	response.Ok(c, gin.H{
		"id": company.ID, "name": company.Name, "code": company.Code,
		"contact": company.Contact, "phone": company.Phone, "address": company.Address,
		"shortName": extra["shortName"], "industry": extra["industry"],
		"taxNature": extra["taxNature"], "wecomCorpId": extra["wecomCorpId"],
	})
}

func (h *SettingExtHandler) CompanyInfoSave(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		Name        string `json:"name" binding:"required"`
		Contact     string `json:"contact"`
		Phone       string `json:"phone"`
		Address     string `json:"address"`
		ShortName   string `json:"shortName"`
		Industry    string `json:"industry"`
		TaxNature   string `json:"taxNature"`
		WecomCorpID string `json:"wecomCorpId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Company{}).Where("id = ?", companyID).
			Updates(map[string]interface{}{
				"name": req.Name, "contact": req.Contact, "phone": req.Phone, "address": req.Address,
			}).Error; err != nil {
			return err
		}
		extras := map[string]string{
			"company.shortName": req.ShortName, "company.industry": req.Industry,
			"company.taxNature": req.TaxNature, "company.wecomCorpId": req.WecomCorpID,
		}
		for k, v := range extras {
			var s model.CompanySetting
			err := tx.Where("company_id = ? AND key = ?", companyID, k).First(&s).Error
			if err != nil {
				s = model.CompanySetting{
					BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
					Key:                  k, Value: v,
				}
				if err := tx.Create(&s).Error; err != nil {
					return err
				}
			} else if err := tx.Model(&s).Update("value", v).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}
