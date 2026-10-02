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

// ProductExtHandler 商品增强（客户单独定价/价格跟踪/商品审核/规格单位条码）
type ProductExtHandler struct {
	db *gorm.DB
}

func NewProductExtHandler(db *gorm.DB) *ProductExtHandler {
	return &ProductExtHandler{db: db}
}

func (h *ProductExtHandler) RegisterRoutes(r *gin.RouterGroup) {
	cp := r.Group("/customer-prices")
	{
		cp.GET("", h.CustomerPriceList)
		cp.POST("", h.CustomerPriceCreate)
		cp.PUT("/:id", h.CustomerPriceUpdate)
		cp.DELETE("/:id", h.CustomerPriceDelete)
	}
	g := r.Group("/product-reports")
	{
		g.GET("/sale-price-track", h.SalePriceTrack)
		g.GET("/purchase-price-track", h.PurchasePriceTrack)
	}
	pa := r.Group("/product-audit")
	{
		pa.GET("", h.AuditList)
		pa.PUT("/:id/approve", h.AuditApprove)
		pa.PUT("/:id/revoke", h.AuditRevoke)
	}
	sb := r.Group("/product-barcodes")
	{
		sb.GET("", h.BarcodeList)
		sb.PUT("/:id", h.BarcodeUpdate)
	}
}

// ==================== 客户单独定价 ====================

func (h *ProductExtHandler) CustomerPriceList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	customerKeyword := c.Query("customerKeyword")
	productKeyword := c.Query("productKeyword")
	categoryID, _ := strconv.Atoi(c.DefaultQuery("categoryId", "0"))

	type row struct {
		ID           uint    `gorm:"column:id" json:"id"`
		CustomerID   uint    `gorm:"column:customer_id" json:"customerId"`
		Customer     string  `gorm:"column:customer" json:"customer"`
		CustomerCode string  `gorm:"column:customer_code" json:"customerCode"`
		ProductID    uint    `gorm:"column:product_id" json:"productId"`
		Product      string  `gorm:"column:product" json:"product"`
		ProductCode  string  `gorm:"column:product_code" json:"productCode"`
		Spec         string  `gorm:"column:spec" json:"spec"`
		Unit         string  `gorm:"column:unit" json:"unit"`
		RetailPrice  float64 `gorm:"column:retail_price" json:"retailPrice"`
		Price        float64 `gorm:"column:price" json:"price"`
		Remark       string  `gorm:"column:remark" json:"remark"`
	}
	var all []row
	h.db.Raw(`SELECT cp.id, cp.customer_id, cu.name AS customer, cu.code AS customer_code,
		cp.product_id, p.name AS product, p.code AS product_code, p.specification AS spec, p.unit,
		p.retail_price, cp.price, COALESCE(cp.remark,'') AS remark
		FROM customer_product_prices cp
		JOIN customers cu ON cu.id = cp.customer_id
		JOIN products p ON p.id = cp.product_id
		WHERE cp.company_id = ? AND cp.deleted_at IS NULL
		AND (? = '' OR cu.name ILIKE ? OR cu.code ILIKE ? OR cu.contact ILIKE ? OR cu.phone ILIKE ?)
		AND (? = '' OR p.name ILIKE ? OR p.code ILIKE ?)
		AND (? = 0 OR p.category_id = ?)
		ORDER BY cp.id DESC`,
		companyID,
		customerKeyword, "%"+customerKeyword+"%", "%"+customerKeyword+"%", "%"+customerKeyword+"%", "%"+customerKeyword+"%",
		productKeyword, "%"+productKeyword+"%", "%"+productKeyword+"%",
		categoryID, categoryID).Scan(&all)
	response.Ok(c, paginateSaleRows(all, page, pageSize))
}

func (h *ProductExtHandler) CustomerPriceCreate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		CustomerID uint    `json:"customerId" binding:"required"`
		ProductID  uint    `json:"productId" binding:"required"`
		Price      float64 `json:"price" binding:"gte=0"`
		Remark     string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	// 同客户同商品已存在则更新价格
	var exist model.CustomerProductPrice
	err := h.db.Where("company_id = ? AND customer_id = ? AND product_id = ?", companyID, req.CustomerID, req.ProductID).First(&exist).Error
	if err == nil {
		exist.Price = req.Price
		exist.Remark = req.Remark
		h.db.Save(&exist)
		response.Ok(c, exist)
		return
	}
	cp := model.CustomerProductPrice{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CustomerID:           req.CustomerID,
		ProductID:            req.ProductID,
		Price:                req.Price,
		Remark:               req.Remark,
	}
	if err := h.db.Create(&cp).Error; err != nil {
		response.Fail(c, 4000, "创建失败: "+err.Error())
		return
	}
	response.Ok(c, cp)
}

func (h *ProductExtHandler) CustomerPriceUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var cp model.CustomerProductPrice
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&cp).Error; err != nil {
		response.Fail(c, 4004, "定价不存在")
		return
	}
	var req struct {
		Price  float64 `json:"price" binding:"gte=0"`
		Remark string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	cp.Price = req.Price
	cp.Remark = req.Remark
	h.db.Save(&cp)
	response.Ok(c, cp)
}

func (h *ProductExtHandler) CustomerPriceDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.CustomerProductPrice{})
	response.Ok(c, nil)
}

// ==================== 销售价格跟踪 ====================

// tagProductIDSet 返回打有指定商品标签的商品 id 集合；tagID<=0 时返回 nil（不过滤）
func (h *ProductExtHandler) tagProductIDSet(companyID uint, tagID int) map[uint]bool {
	if tagID <= 0 {
		return nil
	}
	var ids []uint
	h.db.Model(&model.ProductTagRelation{}).
		Where("company_id = ? AND tag_id = ?", companyID, tagID).
		Pluck("product_id", &ids)
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func (h *ProductExtHandler) SalePriceTrack(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	customerKeyword := c.Query("customerKeyword")
	categoryID, _ := strconv.Atoi(c.DefaultQuery("categoryId", "0"))
	brandID, _ := strconv.Atoi(c.DefaultQuery("brandId", "0"))
	tagID, _ := strconv.Atoi(c.DefaultQuery("tagId", "0"))
	unit := c.Query("unit")
	status := c.Query("status")
	regionID, _ := strconv.Atoi(c.DefaultQuery("regionId", "0"))
	customerCategoryID, _ := strconv.Atoi(c.DefaultQuery("customerCategoryId", "0"))

	type row struct {
		ProductID          uint    `gorm:"column:product_id" json:"productId"`
		Product            string  `gorm:"column:product" json:"product"`
		ProductCode        string  `gorm:"column:product_code" json:"productCode"`
		Barcode            string  `gorm:"column:barcode" json:"barcode"`
		Image              string  `gorm:"column:image" json:"image"`
		Spec               string  `gorm:"column:spec" json:"spec"`
		Unit               string  `gorm:"column:unit" json:"unit"`
		CategoryID         uint    `gorm:"column:category_id" json:"categoryId"`
		BrandID            uint    `gorm:"column:brand_id" json:"brandId"`
		Status             int8    `gorm:"column:status" json:"status"`
		CustomerID         uint    `gorm:"column:customer_id" json:"customerId"`
		Customer           string  `gorm:"column:customer" json:"customer"`
		CustomerCode       string  `gorm:"column:customer_code" json:"customerCode"`
		RegionID           uint    `gorm:"column:region_id" json:"regionId"`
		Region             string  `gorm:"column:region" json:"region"`
		CustomerCategoryID uint    `gorm:"column:customer_category_id" json:"customerCategoryId"`
		CustomerCategory   string  `gorm:"column:customer_category" json:"customerCategory"`
		Price              float64 `gorm:"column:price" json:"price"`
		Amount             float64 `gorm:"column:amount" json:"amount"`
		Quantity           float64 `gorm:"column:quantity" json:"quantity"`
		LastDate           string  `gorm:"column:last_date" json:"lastDate"`
	}
	var all []row
	h.db.Raw(`SELECT DISTINCT ON (i.product_id, b.customer_id)
		i.product_id, p.name AS product, p.code AS product_code, p.barcode, p.image,
		p.specification AS spec, p.unit, p.category_id, p.brand_id, p.status,
		b.customer_id, cu.name AS customer, cu.code AS customer_code,
		cu.region_id, COALESCE(rg.name,'') AS region,
		cu.category_id AS customer_category_id, COALESCE(cc.name,'') AS customer_category,
		i.price, i.amount, i.quantity, b.bill_date::text AS last_date
		FROM sales_out_stock_items i
		JOIN sales_out_stocks b ON b.id = i.out_stock_id
		JOIN products p ON p.id = i.product_id
		JOIN customers cu ON cu.id = b.customer_id
		LEFT JOIN regions rg ON rg.id = cu.region_id
		LEFT JOIN customer_categories cc ON cc.id = cu.category_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		ORDER BY i.product_id, b.customer_id, b.bill_date DESC, b.id DESC`, companyID, start, end).Scan(&all)

	tagSet := h.tagProductIDSet(companyID, tagID)
	list := make([]row, 0, len(all))
	for _, r := range all {
		if keyword != "" && !strings.Contains(r.Product, keyword) && !strings.Contains(r.ProductCode, keyword) && !strings.Contains(r.Spec, keyword) && !strings.Contains(r.Barcode, keyword) {
			continue
		}
		if customerKeyword != "" && !strings.Contains(r.Customer, customerKeyword) && !strings.Contains(r.CustomerCode, customerKeyword) {
			continue
		}
		if categoryID > 0 && r.CategoryID != uint(categoryID) {
			continue
		}
		if brandID > 0 && r.BrandID != uint(brandID) {
			continue
		}
		if tagSet != nil && !tagSet[r.ProductID] {
			continue
		}
		if unit != "" && r.Unit != unit {
			continue
		}
		if status != "" && strconv.Itoa(int(r.Status)) != status {
			continue
		}
		if regionID > 0 && r.RegionID != uint(regionID) {
			continue
		}
		if customerCategoryID > 0 && r.CustomerCategoryID != uint(customerCategoryID) {
			continue
		}
		list = append(list, r)
	}
	sortByStringDesc(list, func(r row) string { return r.LastDate })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 采购价格跟踪 ====================

func (h *ProductExtHandler) PurchasePriceTrack(c *gin.Context) {
	page, pageSize, start, end := salePageParams(c)
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")
	supplierKeyword := c.Query("supplierKeyword")
	categoryID, _ := strconv.Atoi(c.DefaultQuery("categoryId", "0"))
	brandID, _ := strconv.Atoi(c.DefaultQuery("brandId", "0"))
	tagID, _ := strconv.Atoi(c.DefaultQuery("tagId", "0"))
	unit := c.Query("unit")
	status := c.Query("status")

	type row struct {
		ProductID   uint    `gorm:"column:product_id" json:"productId"`
		Product     string  `gorm:"column:product" json:"product"`
		ProductCode string  `gorm:"column:product_code" json:"productCode"`
		Barcode     string  `gorm:"column:barcode" json:"barcode"`
		Spec        string  `gorm:"column:spec" json:"spec"`
		Unit        string  `gorm:"column:unit" json:"unit"`
		CategoryID  uint    `gorm:"column:category_id" json:"categoryId"`
		BrandID     uint    `gorm:"column:brand_id" json:"brandId"`
		Status      int8    `gorm:"column:status" json:"status"`
		SupplierID  uint    `gorm:"column:supplier_id" json:"supplierId"`
		Supplier    string  `gorm:"column:supplier" json:"supplier"`
		Price       float64 `gorm:"column:price" json:"price"`
		Quantity    float64 `gorm:"column:quantity" json:"quantity"`
		LastDate    string  `gorm:"column:last_date" json:"lastDate"`
	}
	var all []row
	h.db.Raw(`SELECT DISTINCT ON (i.product_id, b.supplier_id)
		i.product_id, p.name AS product, p.code AS product_code, p.barcode,
		p.specification AS spec, p.unit, p.category_id, p.brand_id, p.status,
		b.supplier_id, s.name AS supplier,
		i.price, i.quantity, b.bill_date::text AS last_date
		FROM purchase_in_stock_items i
		JOIN purchase_in_stocks b ON b.id = i.in_stock_id
		JOIN products p ON p.id = i.product_id
		JOIN suppliers s ON s.id = b.supplier_id
		WHERE b.company_id = ? AND b.status = 'completed' AND b.deleted_at IS NULL
		AND b.bill_date >= ? AND b.bill_date <= ?
		ORDER BY i.product_id, b.supplier_id, b.bill_date DESC, b.id DESC`, companyID, start, end).Scan(&all)

	tagSet := h.tagProductIDSet(companyID, tagID)
	list := make([]row, 0, len(all))
	for _, r := range all {
		if keyword != "" && !strings.Contains(r.Product, keyword) && !strings.Contains(r.ProductCode, keyword) && !strings.Contains(r.Spec, keyword) && !strings.Contains(r.Barcode, keyword) {
			continue
		}
		if supplierKeyword != "" && !strings.Contains(r.Supplier, supplierKeyword) {
			continue
		}
		if categoryID > 0 && r.CategoryID != uint(categoryID) {
			continue
		}
		if brandID > 0 && r.BrandID != uint(brandID) {
			continue
		}
		if tagSet != nil && !tagSet[r.ProductID] {
			continue
		}
		if unit != "" && r.Unit != unit {
			continue
		}
		if status != "" && strconv.Itoa(int(r.Status)) != status {
			continue
		}
		list = append(list, r)
	}
	sortByStringDesc(list, func(r row) string { return r.LastDate })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// ==================== 商品审核 ====================

func (h *ProductExtHandler) AuditList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	status := c.DefaultQuery("status", "pending")
	keyword := c.Query("keyword")

	q := h.db.Model(&model.Product{}).Where("company_id = ? AND deleted_at IS NULL", companyID)
	if status != "all" {
		q = q.Where("audit_status = ?", status)
	}
	if keyword != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ? OR barcode ILIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Product
	q.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	catNames := map[uint]string{}
	var cats []model.ProductCategory
	h.db.Where("company_id = ?", companyID).Find(&cats)
	for _, cat := range cats {
		catNames[cat.ID] = cat.Name
	}

	rows := make([]gin.H, 0, len(list))
	for _, p := range list {
		rows = append(rows, gin.H{
			"id": p.ID, "code": p.Code, "barcode": p.Barcode, "name": p.Name,
			"spec": p.Specification, "unit": p.Unit, "category": catNames[p.CategoryID],
			"auditStatus": p.AuditStatus, "updatedAt": p.UpdatedAt, "createdAt": p.CreatedAt,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *ProductExtHandler) AuditApprove(c *gin.Context) {
	h.setAuditStatus(c, "approved")
}

func (h *ProductExtHandler) AuditRevoke(c *gin.Context) {
	h.setAuditStatus(c, "pending")
}

func (h *ProductExtHandler) setAuditStatus(c *gin.Context, status string) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	result := h.db.Model(&model.Product{}).
		Where("id = ? AND company_id = ?", id, companyID).
		Update("audit_status", status)
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "商品不存在")
		return
	}
	response.Ok(c, nil)
}

// ==================== 规格单位条码 ====================

func (h *ProductExtHandler) BarcodeList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	companyID := middleware.GetCompanyID(c)
	keyword := c.Query("keyword")

	q := h.db.Model(&model.Product{}).Where("company_id = ? AND deleted_at IS NULL", companyID)
	if keyword != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ? OR barcode ILIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Product
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	rows := make([]gin.H, 0, len(list))
	for _, p := range list {
		rows = append(rows, gin.H{
			"id": p.ID, "code": p.Code, "barcode": p.Barcode,
			"name": p.Name, "spec": p.Specification, "unit": p.Unit,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

func (h *ProductExtHandler) BarcodeUpdate(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Barcode string `json:"barcode" binding:"max=64"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	result := h.db.Model(&model.Product{}).
		Where("id = ? AND company_id = ?", id, companyID).
		Update("barcode", req.Barcode)
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "商品不存在")
		return
	}
	response.Ok(c, nil)
}
