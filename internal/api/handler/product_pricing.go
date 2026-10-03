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
		cp.POST("/import", h.CustomerPriceImport)
		cp.PUT("/:id", h.CustomerPriceUpdate)
		cp.DELETE("/:id", h.CustomerPriceDelete)
	}
	g := r.Group("/product-reports")
	{
		g.GET("/sale-price-track", h.SalePriceTrack)
		g.POST("/sale-price-track", h.SalePriceTrackCreate)
		g.DELETE("/sale-price-track/:id", h.PriceTrackDelete)
		g.POST("/sale-price-track/import", h.SalePriceTrackImport)
		g.GET("/purchase-price-track", h.PurchasePriceTrack)
		g.POST("/purchase-price-track", h.PurchasePriceTrackCreate)
		g.DELETE("/purchase-price-track/:id", h.PriceTrackDelete)
		g.POST("/purchase-price-track/import", h.PurchasePriceTrackImport)
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
		sb.POST("/import", h.BarcodeImport)
		sb.GET("/print-bills", h.PrintBillList)
		sb.GET("/print-bill-items", h.PrintBillItems)
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
	customerCategoryID, _ := strconv.Atoi(c.DefaultQuery("customerCategoryId", "0"))

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
		AND (? = 0 OR cu.category_id = ?)
		ORDER BY cp.id DESC`,
		companyID,
		customerKeyword, "%"+customerKeyword+"%", "%"+customerKeyword+"%", "%"+customerKeyword+"%", "%"+customerKeyword+"%",
		productKeyword, "%"+productKeyword+"%", "%"+productKeyword+"%",
		categoryID, categoryID, customerCategoryID, customerCategoryID).Scan(&all)
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

// customerTagIDSet 返回打了指定客户标签的客户 id 集合；tagID<=0 时返回 nil（不过滤）
func (h *ProductExtHandler) customerTagIDSet(companyID uint, tagID int) map[uint]bool {
	if tagID <= 0 {
		return nil
	}
	var ids []uint
	h.db.Model(&model.CustomerTagRelation{}).
		Where("company_id = ? AND tag_id = ?", companyID, tagID).
		Pluck("customer_id", &ids)
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

// customerTagNameMap 批量取客户标签名（逗号拼接）
func (h *ProductExtHandler) customerTagNameMap(companyID uint, customerIDs []uint) map[uint]string {
	result := make(map[uint]string, len(customerIDs))
	if len(customerIDs) == 0 {
		return result
	}
	type tagRow struct {
		CustomerID uint
		Name       string
	}
	var rows []tagRow
	h.db.Model(&model.CustomerTagRelation{}).
		Select("customer_tag_relations.customer_id, customer_tags.name").
		Joins("JOIN customer_tags ON customer_tags.id = customer_tag_relations.tag_id AND customer_tags.deleted_at IS NULL").
		Where("customer_tag_relations.company_id = ? AND customer_tag_relations.customer_id IN ?", companyID, customerIDs).
		Scan(&rows)
	names := map[uint][]string{}
	for _, r := range rows {
		names[r.CustomerID] = append(names[r.CustomerID], r.Name)
	}
	for id, ns := range names {
		result[id] = strings.Join(ns, "，")
	}
	return result
}

type salePriceTrackRow struct {
	ID                 uint    `json:"id"`     // 手动记录 id；单据生成的为 0
	Source             string  `json:"source"` // auto=销售单据 manual=手动新增
	ProductID          uint    `json:"productId"`
	Product            string  `json:"product"`
	ProductCode        string  `json:"productCode"`
	Barcode            string  `json:"barcode"`
	Image              string  `json:"image"`
	Spec               string  `json:"spec"`
	Unit               string  `json:"unit"`
	CategoryID         uint    `json:"categoryId"`
	BrandID            uint    `json:"brandId"`
	Status             int8    `json:"status"`
	CustomerID         uint    `json:"customerId"`
	Customer           string  `json:"customer"`
	CustomerCode       string  `json:"customerCode"`
	RegionID           uint    `json:"regionId"`
	Region             string  `json:"region"`
	CustomerCategoryID uint    `json:"customerCategoryId"`
	CustomerCategory   string  `json:"customerCategory"`
	CustomerTags       string  `json:"customerTags"`
	Price              float64 `json:"price"`
	Amount             float64 `json:"amount"`
	Quantity           float64 `json:"quantity"`
	LastDate           string  `json:"lastDate"`
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
	customerTagID, _ := strconv.Atoi(c.DefaultQuery("customerTagId", "0"))

	type rawRow struct {
		ProductID          uint    `gorm:"column:product_id"`
		Product            string  `gorm:"column:product"`
		ProductCode        string  `gorm:"column:product_code"`
		Barcode            string  `gorm:"column:barcode"`
		Image              string  `gorm:"column:image"`
		Spec               string  `gorm:"column:spec"`
		Unit               string  `gorm:"column:unit"`
		CategoryID         uint    `gorm:"column:category_id"`
		BrandID            uint    `gorm:"column:brand_id"`
		Status             int8    `gorm:"column:status"`
		CustomerID         uint    `gorm:"column:customer_id"`
		Customer           string  `gorm:"column:customer"`
		CustomerCode       string  `gorm:"column:customer_code"`
		RegionID           uint    `gorm:"column:region_id"`
		Region             string  `gorm:"column:region"`
		CustomerCategoryID uint    `gorm:"column:customer_category_id"`
		CustomerCategory   string  `gorm:"column:customer_category"`
		Price              float64 `gorm:"column:price"`
		Amount             float64 `gorm:"column:amount"`
		Quantity           float64 `gorm:"column:quantity"`
		LastDate           string  `gorm:"column:last_date"`
	}
	var autoRows []rawRow
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
		ORDER BY i.product_id, b.customer_id, b.bill_date DESC, b.id DESC`, companyID, start, end).Scan(&autoRows)

	all := make([]salePriceTrackRow, 0, len(autoRows))
	for _, r := range autoRows {
		all = append(all, salePriceTrackRow{
			Source:    "auto",
			ProductID: r.ProductID, Product: r.Product, ProductCode: r.ProductCode, Barcode: r.Barcode,
			Image: r.Image, Spec: r.Spec, Unit: r.Unit, CategoryID: r.CategoryID, BrandID: r.BrandID, Status: r.Status,
			CustomerID: r.CustomerID, Customer: r.Customer, CustomerCode: r.CustomerCode,
			RegionID: r.RegionID, Region: r.Region,
			CustomerCategoryID: r.CustomerCategoryID, CustomerCategory: r.CustomerCategory,
			Price: r.Price, Amount: r.Amount, Quantity: r.Quantity, LastDate: r.LastDate,
		})
	}
	// 合并手动新增的价格跟踪
	all = append(all, h.manualSaleTrackRows(companyID, start, end)...)

	tagSet := h.tagProductIDSet(companyID, tagID)
	customerTagSet := h.customerTagIDSet(companyID, customerTagID)
	list := make([]salePriceTrackRow, 0, len(all))
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
		if customerTagSet != nil && !customerTagSet[r.CustomerID] {
			continue
		}
		list = append(list, r)
	}
	// 客户标签列
	customerIDs := make([]uint, 0, len(list))
	seen := map[uint]bool{}
	for _, r := range list {
		if !seen[r.CustomerID] {
			seen[r.CustomerID] = true
			customerIDs = append(customerIDs, r.CustomerID)
		}
	}
	tagNames := h.customerTagNameMap(companyID, customerIDs)
	for i := range list {
		list[i].CustomerTags = tagNames[list[i].CustomerID]
	}
	sortByStringDesc(list, func(r salePriceTrackRow) string { return r.LastDate })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// manualSaleTrackRows 读取手动新增的销售价格跟踪记录并补齐商品/客户信息
func (h *ProductExtHandler) manualSaleTrackRows(companyID uint, start, end string) []salePriceTrackRow {
	var tracks []model.ProductPriceTrack
	h.db.Where("company_id = ? AND type = ? AND track_date >= ? AND track_date <= ?", companyID, "sale", start, end).
		Order("id DESC").Find(&tracks)
	if len(tracks) == 0 {
		return nil
	}
	productIDs := make([]uint, 0, len(tracks))
	customerIDs := make([]uint, 0, len(tracks))
	for _, t := range tracks {
		productIDs = append(productIDs, t.ProductID)
		customerIDs = append(customerIDs, t.CustomerID)
	}
	products := map[uint]model.Product{}
	var plist []model.Product
	h.db.Where("company_id = ? AND id IN ?", companyID, productIDs).Find(&plist)
	for _, p := range plist {
		products[p.ID] = p
	}
	customers := map[uint]model.Customer{}
	var clist []model.Customer
	h.db.Where("company_id = ? AND id IN ?", companyID, customerIDs).Find(&clist)
	for _, cu := range clist {
		customers[cu.ID] = cu
	}
	regionNames := map[uint]string{}
	categoryNames := map[uint]string{}
	for _, cu := range clist {
		if cu.RegionID > 0 {
			var name string
			h.db.Model(&model.Region{}).Where("id = ?", cu.RegionID).Select("name").Scan(&name)
			regionNames[cu.ID] = name
		}
		if cu.CategoryID > 0 {
			var name string
			h.db.Model(&model.CustomerCategory{}).Where("id = ?", cu.CategoryID).Select("name").Scan(&name)
			categoryNames[cu.ID] = name
		}
	}
	rows := make([]salePriceTrackRow, 0, len(tracks))
	for _, t := range tracks {
		p := products[t.ProductID]
		cu := customers[t.CustomerID]
		rows = append(rows, salePriceTrackRow{
			ID: t.ID, Source: "manual",
			ProductID: t.ProductID, Product: p.Name, ProductCode: p.Code, Barcode: p.Barcode,
			Image: p.Image, Spec: p.Specification, Unit: p.Unit, CategoryID: p.CategoryID, BrandID: p.BrandID, Status: p.Status,
			CustomerID: t.CustomerID, Customer: cu.Name, CustomerCode: cu.Code,
			RegionID: cu.RegionID, Region: regionNames[cu.ID],
			CustomerCategoryID: cu.CategoryID, CustomerCategory: categoryNames[cu.ID],
			Price: t.Price, Amount: t.Price * t.Quantity, Quantity: t.Quantity, LastDate: t.TrackDate,
		})
	}
	return rows
}

// priceTrackCreateReq 新增价格跟踪请求
type priceTrackCreateReq struct {
	CustomerID uint    `json:"customerId"`
	SupplierID uint    `json:"supplierId"`
	ProductID  uint    `json:"productId" binding:"required"`
	Price      float64 `json:"price" binding:"gte=0"`
	Quantity   float64 `json:"quantity" binding:"gte=0"`
	TrackDate  string  `json:"trackDate" binding:"max=10"`
	Remark     string  `json:"remark" binding:"max=255"`
}

func (h *ProductExtHandler) createPriceTrack(c *gin.Context, trackType string, req priceTrackCreateReq) {
	companyID := middleware.GetCompanyID(c)
	if req.ProductID == 0 {
		response.Fail(c, 4000, "请选择商品")
		return
	}
	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", req.ProductID, companyID).First(&product).Error; err != nil {
		response.Fail(c, 4004, "商品不存在")
		return
	}
	track := model.ProductPriceTrack{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Type:                 trackType,
		CustomerID:           req.CustomerID,
		SupplierID:           req.SupplierID,
		ProductID:            req.ProductID,
		Price:                req.Price,
		Quantity:             req.Quantity,
		TrackDate:            req.TrackDate,
		Remark:               req.Remark,
	}
	if trackType == "sale" {
		if req.CustomerID == 0 {
			response.Fail(c, 4000, "请选择客户")
			return
		}
		var count int64
		h.db.Model(&model.Customer{}).Where("id = ? AND company_id = ?", req.CustomerID, companyID).Count(&count)
		if count == 0 {
			response.Fail(c, 4004, "客户不存在")
			return
		}
	} else {
		if req.SupplierID == 0 {
			response.Fail(c, 4000, "请选择供应商")
			return
		}
		var count int64
		h.db.Model(&model.Supplier{}).Where("id = ? AND company_id = ?", req.SupplierID, companyID).Count(&count)
		if count == 0 {
			response.Fail(c, 4004, "供应商不存在")
			return
		}
	}
	if track.TrackDate == "" {
		track.TrackDate = time.Now().Format("2006-01-02")
	}
	if err := h.db.Create(&track).Error; err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, track)
}

func (h *ProductExtHandler) SalePriceTrackCreate(c *gin.Context) {
	var req priceTrackCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	h.createPriceTrack(c, "sale", req)
}

func (h *ProductExtHandler) PurchasePriceTrackCreate(c *gin.Context) {
	var req priceTrackCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误: "+err.Error())
		return
	}
	h.createPriceTrack(c, "purchase", req)
}

// PriceTrackDelete 删除手动新增的价格跟踪（type 由路由前缀区分）
func (h *ProductExtHandler) PriceTrackDelete(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	trackType := "sale"
	if strings.Contains(c.FullPath(), "purchase") {
		trackType = "purchase"
	}
	result := h.db.Where("id = ? AND company_id = ? AND type = ?", id, companyID, trackType).Delete(&model.ProductPriceTrack{})
	if result.RowsAffected == 0 {
		response.Fail(c, 4004, "记录不存在")
		return
	}
	response.Ok(c, nil)
}

// ==================== 采购价格跟踪 ====================

type purchasePriceTrackRow struct {
	ID          uint    `json:"id"`     // 手动记录 id；单据生成的为 0
	Source      string  `json:"source"` // auto=采购单据 manual=手动新增
	ProductID   uint    `json:"productId"`
	Product     string  `json:"product"`
	ProductCode string  `json:"productCode"`
	Barcode     string  `json:"barcode"`
	Spec        string  `json:"spec"`
	Unit        string  `json:"unit"`
	CategoryID  uint    `json:"categoryId"`
	BrandID     uint    `json:"brandId"`
	Status      int8    `json:"status"`
	SupplierID  uint    `json:"supplierId"`
	Supplier    string  `json:"supplier"`
	Price       float64 `json:"price"`
	Quantity    float64 `json:"quantity"`
	LastDate    string  `json:"lastDate"`
}

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

	type rawRow struct {
		ProductID   uint    `gorm:"column:product_id"`
		Product     string  `gorm:"column:product"`
		ProductCode string  `gorm:"column:product_code"`
		Barcode     string  `gorm:"column:barcode"`
		Spec        string  `gorm:"column:spec"`
		Unit        string  `gorm:"column:unit"`
		CategoryID  uint    `gorm:"column:category_id"`
		BrandID     uint    `gorm:"column:brand_id"`
		Status      int8    `gorm:"column:status"`
		SupplierID  uint    `gorm:"column:supplier_id"`
		Supplier    string  `gorm:"column:supplier"`
		Price       float64 `gorm:"column:price"`
		Quantity    float64 `gorm:"column:quantity"`
		LastDate    string  `gorm:"column:last_date"`
	}
	var autoRows []rawRow
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
		ORDER BY i.product_id, b.supplier_id, b.bill_date DESC, b.id DESC`, companyID, start, end).Scan(&autoRows)

	all := make([]purchasePriceTrackRow, 0, len(autoRows))
	for _, r := range autoRows {
		all = append(all, purchasePriceTrackRow{
			Source:    "auto",
			ProductID: r.ProductID, Product: r.Product, ProductCode: r.ProductCode, Barcode: r.Barcode,
			Spec: r.Spec, Unit: r.Unit, CategoryID: r.CategoryID, BrandID: r.BrandID, Status: r.Status,
			SupplierID: r.SupplierID, Supplier: r.Supplier,
			Price: r.Price, Quantity: r.Quantity, LastDate: r.LastDate,
		})
	}
	// 合并手动新增的价格跟踪
	all = append(all, h.manualPurchaseTrackRows(companyID, start, end)...)

	tagSet := h.tagProductIDSet(companyID, tagID)
	list := make([]purchasePriceTrackRow, 0, len(all))
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
	sortByStringDesc(list, func(r purchasePriceTrackRow) string { return r.LastDate })
	response.Ok(c, paginateSaleRows(list, page, pageSize))
}

// manualPurchaseTrackRows 读取手动新增的采购价格跟踪记录并补齐商品/供应商信息
func (h *ProductExtHandler) manualPurchaseTrackRows(companyID uint, start, end string) []purchasePriceTrackRow {
	var tracks []model.ProductPriceTrack
	h.db.Where("company_id = ? AND type = ? AND track_date >= ? AND track_date <= ?", companyID, "purchase", start, end).
		Order("id DESC").Find(&tracks)
	if len(tracks) == 0 {
		return nil
	}
	productIDs := make([]uint, 0, len(tracks))
	supplierIDs := make([]uint, 0, len(tracks))
	for _, t := range tracks {
		productIDs = append(productIDs, t.ProductID)
		supplierIDs = append(supplierIDs, t.SupplierID)
	}
	products := map[uint]model.Product{}
	var plist []model.Product
	h.db.Where("company_id = ? AND id IN ?", companyID, productIDs).Find(&plist)
	for _, p := range plist {
		products[p.ID] = p
	}
	suppliers := map[uint]model.Supplier{}
	var slist []model.Supplier
	h.db.Where("company_id = ? AND id IN ?", companyID, supplierIDs).Find(&slist)
	for _, s := range slist {
		suppliers[s.ID] = s
	}
	rows := make([]purchasePriceTrackRow, 0, len(tracks))
	for _, t := range tracks {
		p := products[t.ProductID]
		s := suppliers[t.SupplierID]
		rows = append(rows, purchasePriceTrackRow{
			ID: t.ID, Source: "manual",
			ProductID: t.ProductID, Product: p.Name, ProductCode: p.Code, Barcode: p.Barcode,
			Spec: p.Specification, Unit: p.Unit, CategoryID: p.CategoryID, BrandID: p.BrandID, Status: p.Status,
			SupplierID: t.SupplierID, Supplier: s.Name,
			Price: t.Price, Quantity: t.Quantity, LastDate: t.TrackDate,
		})
	}
	return rows
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

// ==================== 导入（客户定价 / 价格跟踪 / 条码） ====================

// importResultRow 导入结果明细
type importResultRow struct {
	Created int              `json:"created"`
	Updated int              `json:"updated"`
	Failed  int              `json:"failed"`
	Errors  []importErrorRow `json:"errors"`
}

type importErrorRow struct {
	Row    int    `json:"row"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

// readImportRows 读取上传的 Excel 并按表头名归一化
func readImportRows(c *gin.Context, headers ...string) ([][]string, int, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, -1, err
	}
	if file.Size > 10*1024*1024 {
		return nil, -1, errTooLarge
	}
	src, err := file.Open()
	if err != nil {
		return nil, -1, err
	}
	defer src.Close()
	rows, err := openSpreadsheetRows(src, file.Filename)
	if err != nil {
		return nil, -1, err
	}
	for i := range rows {
		for j := range rows[i] {
			rows[i][j] = cleanHeaderCell(rows[i][j])
		}
	}
	headerIdx := findHeaderRow(rows, 5, headers...)
	if headerIdx < 0 {
		return nil, -1, errNoHeader
	}
	return rows, headerIdx, nil
}

var (
	errTooLarge = errImport("文件大小不能超过 10MB")
	errNoHeader = errImport("未找到表头行")
)

type errImport string

func (e errImport) Error() string { return string(e) }

// headerColIdx 按表头名定位列索引（找不到为 -1）
func headerColIdx(header []string, names ...string) map[string]int {
	idx := map[string]int{}
	for _, n := range names {
		idx[n] = -1
	}
	for j, cell := range header {
		if _, ok := idx[cell]; ok {
			idx[cell] = j
		}
	}
	return idx
}

func cellAt(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

// parseImportPrice 解析导入的数值（复用 other_instock_import.go 的 parseImportFloat）
func parseImportPrice(s string) float64 {
	f, _ := parseImportFloat(s)
	return f
}

// findProductForImport 按编号优先、名称其次匹配商品
func (h *ProductExtHandler) findProductForImport(companyID uint, code, name string) *model.Product {
	var p model.Product
	if code != "" {
		if err := h.db.Where("company_id = ? AND code = ?", companyID, code).First(&p).Error; err == nil {
			return &p
		}
	}
	if name != "" {
		if err := h.db.Where("company_id = ? AND name = ?", companyID, name).First(&p).Error; err == nil {
			return &p
		}
	}
	return nil
}

// CustomerPriceImport 客户单独定价 Excel 导入（列：客户编号/客户名称、商品编号/商品名称、客户定价、备注）
func (h *ProductExtHandler) CustomerPriceImport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	rows, headerIdx, err := readImportRows(c, "客户定价")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cols := headerColIdx(rows[headerIdx], "客户编号", "客户名称", "商品编号", "商品名称", "客户定价", "备注")
	if cols["客户定价"] < 0 || (cols["客户编号"] < 0 && cols["客户名称"] < 0) || (cols["商品编号"] < 0 && cols["商品名称"] < 0) {
		response.BadRequest(c, "表头需包含：客户编号/客户名称、商品编号/商品名称、客户定价")
		return
	}
	result := importResultRow{Errors: []importErrorRow{}}
	for i := headerIdx + 1; i < len(rows); i++ {
		raw := rows[i]
		customerCode := cellAt(raw, cols["客户编号"])
		customerName := cellAt(raw, cols["客户名称"])
		productCode := cellAt(raw, cols["商品编号"])
		productName := cellAt(raw, cols["商品名称"])
		if customerCode == "" && customerName == "" && productCode == "" && productName == "" {
			continue
		}
		codeLabel := productCode
		if codeLabel == "" {
			codeLabel = productName
		}
		var customer model.Customer
		found := false
		if customerCode != "" {
			found = h.db.Where("company_id = ? AND code = ?", companyID, customerCode).First(&customer).Error == nil
		}
		if !found && customerName != "" {
			found = h.db.Where("company_id = ? AND name = ?", companyID, customerName).First(&customer).Error == nil
		}
		if !found {
			result.Failed++
			result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "客户不存在: " + customerCode + customerName})
			continue
		}
		product := h.findProductForImport(companyID, productCode, productName)
		if product == nil {
			result.Failed++
			result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "商品不存在"})
			continue
		}
		price := parseImportPrice(cellAt(raw, cols["客户定价"]))
		remark := cellAt(raw, cols["备注"])
		var exist model.CustomerProductPrice
		if err := h.db.Where("company_id = ? AND customer_id = ? AND product_id = ?", companyID, customer.ID, product.ID).First(&exist).Error; err == nil {
			exist.Price = price
			exist.Remark = remark
			h.db.Save(&exist)
			result.Updated++
			continue
		}
		cp := model.CustomerProductPrice{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			CustomerID:           customer.ID,
			ProductID:            product.ID,
			Price:                price,
			Remark:               remark,
		}
		if err := h.db.Create(&cp).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "保存失败"})
			continue
		}
		result.Created++
	}
	response.Ok(c, result)
}

// priceTrackImport 价格跟踪导入共用逻辑
func (h *ProductExtHandler) priceTrackImport(c *gin.Context, trackType string) {
	companyID := middleware.GetCompanyID(c)
	rows, headerIdx, err := readImportRows(c, "商品编号", "商品名称")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var partyColNames []string
	priceHeader := "销售折前价"
	if trackType == "sale" {
		partyColNames = []string{"客户编号", "客户名称"}
	} else {
		partyColNames = []string{"供应商编号", "供应商名称", "供应商"}
		priceHeader = "采购价"
	}
	names := append([]string{"商品编号", "商品名称", priceHeader, "价格", "数量", "日期", "备注"}, partyColNames...)
	cols := headerColIdx(rows[headerIdx], names...)
	if cols["商品编号"] < 0 && cols["商品名称"] < 0 {
		response.BadRequest(c, "表头需包含：商品编号/商品名称")
		return
	}
	priceCol := cols[priceHeader]
	if priceCol < 0 {
		priceCol = cols["价格"]
	}
	result := importResultRow{Errors: []importErrorRow{}}
	for i := headerIdx + 1; i < len(rows); i++ {
		raw := rows[i]
		productCode := cellAt(raw, cols["商品编号"])
		productName := cellAt(raw, cols["商品名称"])
		if productCode == "" && productName == "" {
			continue
		}
		codeLabel := productCode
		if codeLabel == "" {
			codeLabel = productName
		}
		product := h.findProductForImport(companyID, productCode, productName)
		if product == nil {
			result.Failed++
			result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "商品不存在"})
			continue
		}
		track := model.ProductPriceTrack{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Type:                 trackType,
			ProductID:            product.ID,
			Price:                parseImportPrice(cellAt(raw, priceCol)),
			Quantity:             parseImportPrice(cellAt(raw, cols["数量"])),
			TrackDate:            cellAt(raw, cols["日期"]),
			Remark:               cellAt(raw, cols["备注"]),
		}
		// 客户/供应商匹配
		partyFound := false
		for _, pn := range partyColNames {
			val := cellAt(raw, cols[pn])
			if val == "" {
				continue
			}
			if trackType == "sale" {
				var customer model.Customer
				q := h.db.Where("company_id = ?", companyID)
				if strings.HasSuffix(pn, "编号") {
					q = q.Where("code = ?", val)
				} else {
					q = q.Where("name = ?", val)
				}
				if q.First(&customer).Error == nil {
					track.CustomerID = customer.ID
					partyFound = true
					break
				}
			} else {
				var supplier model.Supplier
				q := h.db.Where("company_id = ?", companyID)
				if strings.HasSuffix(pn, "编号") {
					q = q.Where("code = ?", val)
				} else {
					q = q.Where("name = ?", val)
				}
				if q.First(&supplier).Error == nil {
					track.SupplierID = supplier.ID
					partyFound = true
					break
				}
			}
		}
		if !partyFound {
			result.Failed++
			if trackType == "sale" {
				result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "客户不存在"})
			} else {
				result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "供应商不存在"})
			}
			continue
		}
		if track.TrackDate == "" {
			track.TrackDate = time.Now().Format("2006-01-02")
		}
		if err := h.db.Create(&track).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "保存失败"})
			continue
		}
		result.Created++
	}
	response.Ok(c, result)
}

func (h *ProductExtHandler) SalePriceTrackImport(c *gin.Context) {
	h.priceTrackImport(c, "sale")
}

func (h *ProductExtHandler) PurchasePriceTrackImport(c *gin.Context) {
	h.priceTrackImport(c, "purchase")
}

// BarcodeImport 条码导入（列：商品编号/商品名称、条码），按编号优先匹配更新商品条码
func (h *ProductExtHandler) BarcodeImport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	rows, headerIdx, err := readImportRows(c, "条码")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cols := headerColIdx(rows[headerIdx], "商品编号", "商品名称", "条码")
	if cols["条码"] < 0 || (cols["商品编号"] < 0 && cols["商品名称"] < 0) {
		response.BadRequest(c, "表头需包含：商品编号/商品名称、条码")
		return
	}
	result := importResultRow{Errors: []importErrorRow{}}
	for i := headerIdx + 1; i < len(rows); i++ {
		raw := rows[i]
		code := cellAt(raw, cols["商品编号"])
		name := cellAt(raw, cols["商品名称"])
		barcode := cellAt(raw, cols["条码"])
		if code == "" && name == "" && barcode == "" {
			continue
		}
		codeLabel := code
		if codeLabel == "" {
			codeLabel = name
		}
		product := h.findProductForImport(companyID, code, name)
		if product == nil {
			result.Failed++
			result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "商品不存在"})
			continue
		}
		if err := h.db.Model(&model.Product{}).Where("id = ? AND company_id = ?", product.ID, companyID).Update("barcode", barcode).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, importErrorRow{Row: i + 1, Code: codeLabel, Reason: "保存失败"})
			continue
		}
		result.Updated++
	}
	response.Ok(c, result)
}

// ==================== 条码打印·引入单据 ====================

// PrintBillList 可引入打印的单据列表（采购入库单 / 其他入库单）
func (h *ProductExtHandler) PrintBillList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	billType := c.DefaultQuery("type", "purchase-in")
	keyword := c.Query("keyword")

	type billRow struct {
		ID        uint    `json:"id"`
		BillNo    string  `json:"billNo"`
		BillDate  string  `json:"billDate"`
		Party     string  `json:"party"`
		ItemCount int64   `json:"itemCount"`
		Amount    float64 `json:"amount"`
	}
	rows := []billRow{}
	if billType == "other-in" {
		var list []model.OtherInStock
		q := h.db.Where("company_id = ? AND deleted_at IS NULL", companyID)
		if keyword != "" {
			q = q.Where("bill_no ILIKE ?", "%"+keyword+"%")
		}
		q.Order("id DESC").Limit(50).Find(&list)
		for _, b := range list {
			var cnt int64
			h.db.Model(&model.OtherInStockItem{}).Where("in_stock_id = ? AND deleted_at IS NULL", b.ID).Count(&cnt)
			rows = append(rows, billRow{ID: b.ID, BillNo: b.BillNo, BillDate: b.BillDate.Format("2006-01-02"), Party: b.Counterpart, ItemCount: cnt, Amount: b.Amount})
		}
	} else {
		var list []model.PurchaseInStock
		q := h.db.Where("company_id = ? AND deleted_at IS NULL", companyID)
		if keyword != "" {
			q = q.Where("bill_no ILIKE ?", "%"+keyword+"%")
		}
		q.Order("id DESC").Limit(50).Find(&list)
		supplierNames := map[uint]string{}
		for _, b := range list {
			if b.SupplierID > 0 {
				if _, ok := supplierNames[b.SupplierID]; !ok {
					var name string
					h.db.Model(&model.Supplier{}).Where("id = ?", b.SupplierID).Select("name").Scan(&name)
					supplierNames[b.SupplierID] = name
				}
			}
		}
		for _, b := range list {
			var cnt int64
			h.db.Model(&model.PurchaseInStockItem{}).Where("in_stock_id = ? AND deleted_at IS NULL", b.ID).Count(&cnt)
			rows = append(rows, billRow{ID: b.ID, BillNo: b.BillNo, BillDate: b.BillDate.Format("2006-01-02"), Party: supplierNames[b.SupplierID], ItemCount: cnt, Amount: b.TotalAmount})
		}
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}

// PrintBillItems 单据明细（含商品条码信息），用于引入打印
func (h *ProductExtHandler) PrintBillItems(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	billType := c.DefaultQuery("type", "purchase-in")
	id, _ := strconv.Atoi(c.Query("id"))
	if id == 0 {
		response.BadRequest(c, "缺少单据 id")
		return
	}

	type itemRow struct {
		ProductID   uint    `json:"productId"`
		Name        string  `json:"name"`
		Spec        string  `json:"spec"`
		Code        string  `json:"code"`
		Barcode     string  `json:"barcode"`
		RetailPrice float64 `json:"retailPrice"`
		Quantity    float64 `json:"quantity"`
	}
	rows := []itemRow{}
	if billType == "other-in" {
		h.db.Raw(`SELECT i.product_id, p.name, p.specification AS spec, p.code, p.barcode, p.retail_price, i.quantity
			FROM other_in_stock_items i
			JOIN other_in_stocks b ON b.id = i.in_stock_id
			JOIN products p ON p.id = i.product_id
			WHERE i.in_stock_id = ? AND b.company_id = ? AND i.deleted_at IS NULL
			ORDER BY i.id`, id, companyID).Scan(&rows)
	} else {
		h.db.Raw(`SELECT i.product_id, p.name, p.specification AS spec, p.code, p.barcode, p.retail_price, i.quantity
			FROM purchase_in_stock_items i
			JOIN purchase_in_stocks b ON b.id = i.in_stock_id
			JOIN products p ON p.id = i.product_id
			WHERE i.in_stock_id = ? AND b.company_id = ? AND i.deleted_at IS NULL
			ORDER BY i.id`, id, companyID).Scan(&rows)
	}
	response.Ok(c, gin.H{"list": rows, "total": len(rows)})
}
