package handler

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// ==================== 按规格列表 ====================

type SpecItemListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CategoryID uint   `form:"categoryId"`
	BrandID    uint   `form:"brandId"`
	Status     *int8  `form:"status"`
}

type SpecItemResp struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"productId"`
	SpecValue    string `json:"specValue"`
	Code         string `json:"code"`
	Barcode      string `json:"barcode"`
	OnShelf      int8   `json:"onShelf"`
	ProductName  string `json:"productName"`
	ProductImage string `json:"productImage"`
	BrandName    string `json:"brandName"`
	CategoryName string `json:"categoryName"`
	Unit         string `json:"unit"`
}

func (h *ProductHandler) ListProductSpecItems(c *gin.Context) {
	var req SpecItemListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.ProductSpecItem{}).
		Joins("JOIN products ON products.id = product_spec_items.product_id AND products.deleted_at IS NULL").
		Joins("LEFT JOIN brands ON brands.id = products.brand_id").
		Joins("LEFT JOIN product_categories ON product_categories.id = products.category_id").
		Where("product_spec_items.company_id = ?", companyID)
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("products.name LIKE ? OR products.code LIKE ? OR products.barcode LIKE ? OR product_spec_items.spec_value LIKE ?",
			like, like, like, like)
	}
	if req.CategoryID > 0 {
		query = query.Where("products.category_id = ?", req.CategoryID)
	}
	if req.BrandID > 0 {
		query = query.Where("products.brand_id = ?", req.BrandID)
	}
	if req.Status != nil {
		query = query.Where("products.status = ?", *req.Status)
	}

	var total int64
	query.Count(&total)

	list := make([]SpecItemResp, 0)
	query.Select("product_spec_items.id, product_spec_items.product_id, product_spec_items.spec_value, product_spec_items.code, product_spec_items.barcode, product_spec_items.on_shelf, products.name AS product_name, products.image AS product_image, brands.name AS brand_name, product_categories.name AS category_name, products.unit").
		Order("product_spec_items.id DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

// ==================== 专题分类 ====================

func (h *ProductHandler) ListTopicCategories(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	names := make([]string, 0)
	h.db.Model(&model.Product{}).
		Where("company_id = ? AND topic_category <> ''", companyID).
		Distinct().
		Pluck("topic_category", &names)
	response.Ok(c, names)
}

// ==================== Excel 导入 ====================

func resolveCategoryByName(tx *gorm.DB, companyID uint, name string) (uint, error) {
	if name == "" {
		return 0, nil
	}
	var cat model.ProductCategory
	err := tx.Where("company_id = ? AND name = ?", companyID, name).First(&cat).Error
	if err == nil {
		return cat.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, err
	}
	cat = model.ProductCategory{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 name,
		Status:               1,
	}
	if err := tx.Create(&cat).Error; err != nil {
		return 0, err
	}
	return cat.ID, nil
}

func resolveBrandByName(tx *gorm.DB, companyID uint, name string) (uint, error) {
	if name == "" {
		return 0, nil
	}
	var brand model.Brand
	err := tx.Where("company_id = ? AND name = ?", companyID, name).First(&brand).Error
	if err == nil {
		return brand.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, err
	}
	brand = model.Brand{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 name,
		Status:               1,
	}
	if err := tx.Create(&brand).Error; err != nil {
		return 0, err
	}
	return brand.ID, nil
}

type productImportRow struct {
	Code        string
	Name        string
	Category    string
	Brand       string
	Unit        string
	Barcode     string
	PinyinCode  string
	MallName    string
	Wholesale   string
	Retail      string
	Purchase    string
	Spec        string
}

func (r *productImportRow) empty() bool {
	return r.Code == "" && r.Name == "" && r.Category == "" && r.Brand == "" &&
		r.Unit == "" && r.Barcode == "" && r.PinyinCode == "" && r.MallName == "" &&
		r.Wholesale == "" && r.Retail == "" && r.Purchase == "" && r.Spec == ""
}

func (h *ProductHandler) readImportFile(c *gin.Context) (*excelize.File, [][]string, bool) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return nil, nil, false
	}
	if file.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小不能超过 10MB")
		return nil, nil, false
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".xlsx" {
		response.BadRequest(c, "仅支持 .xlsx 文件")
		return nil, nil, false
	}
	src, err := file.Open()
	if err != nil {
		response.ServerError(c, "读取文件失败")
		return nil, nil, false
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		response.BadRequest(c, "Excel 文件解析失败")
		return nil, nil, false
	}
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil || len(rows) < 2 {
		f.Close()
		response.BadRequest(c, "文件内容为空")
		return nil, nil, false
	}
	return f, rows, true
}

// createImportedProduct 在事务内创建导入的商品，返回错误原因（业务失败）或 error
func (h *ProductHandler) createImportedProduct(companyID uint, row *productImportRow) (string, error) {
	if row.Name == "" {
		return "商品名称不能为空", nil
	}
	if row.Unit == "" {
		return "商品单位不能为空", nil
	}
	wholesale, err := parsePriceCell(row.Wholesale)
	if err != nil {
		return "批发价" + err.Error(), nil
	}
	retail, err := parsePriceCell(row.Retail)
	if err != nil {
		return "零售价" + err.Error(), nil
	}
	purchase, err := parsePriceCell(row.Purchase)
	if err != nil {
		return "采购价" + err.Error(), nil
	}

	product := model.Product{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 row.Name,
		Code:                 row.Code,
		Barcode:              row.Barcode,
		Specification:        row.Spec,
		Unit:                 row.Unit,
		PinyinCode:           row.PinyinCode,
		MallName:             row.MallName,
		Status:               1,
	}
	if wholesale != nil {
		product.WholesalePrice = *wholesale
	}
	if retail != nil {
		product.RetailPrice = *retail
	}
	if purchase != nil {
		product.PurchasePrice = *purchase
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if product.Code == "" {
			code, err := nextProductCode(tx, companyID)
			if err != nil {
				return err
			}
			product.Code = code
		}
		catID, err := resolveCategoryByName(tx, companyID, row.Category)
		if err != nil {
			return err
		}
		product.CategoryID = catID
		brandID, err := resolveBrandByName(tx, companyID, row.Brand)
		if err != nil {
			return err
		}
		product.BrandID = brandID
		return tx.Create(&product).Error
	})
	return "", err
}

// ImportProductsSystem 系统模板导入（固定列序，仅新增）
func (h *ProductHandler) ImportProductsSystem(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	f, rows, ok := h.readImportFile(c)
	if !ok {
		return
	}
	defer f.Close()

	created := 0
	failed := 0
	errors := make([]PriceImportError, 0, 20)

	for i := 1; i < len(rows); i++ {
		raw := rows[i]
		rowNo := i + 1
		getCell := func(idx int) string {
			if idx < len(raw) {
				return strings.TrimSpace(raw[idx])
			}
			return ""
		}
		row := productImportRow{
			Code:       getCell(0),
			Name:       getCell(1),
			Category:   getCell(2),
			Brand:      getCell(3),
			Unit:       getCell(4),
			Barcode:    getCell(5),
			PinyinCode: getCell(6),
			MallName:   getCell(7),
			Wholesale:  getCell(8),
			Retail:     getCell(9),
			Purchase:   getCell(10),
			Spec:       getCell(11),
		}
		if row.empty() {
			continue
		}
		fail := func(reason string) {
			failed++
			if len(errors) < 20 {
				errors = append(errors, PriceImportError{Row: rowNo, Code: row.Code, Reason: reason})
			}
		}

		if row.Code != "" {
			var count int64
			h.db.Model(&model.Product{}).Where("company_id = ? AND code = ?", companyID, row.Code).Count(&count)
			if count > 0 {
				fail("商品编号已存在")
				continue
			}
		}

		reason, err := h.createImportedProduct(companyID, &row)
		if reason != "" {
			fail(reason)
			continue
		}
		if err != nil {
			log.Error().Err(err).Int("row", rowNo).Msg("import product failed")
			fail("保存失败")
			continue
		}
		created++
	}

	response.Ok(c, gin.H{
		"created": created,
		"failed":  failed,
		"errors":  errors,
	})
}

// ==================== 自定义模板更新导入 ====================

var customCategoryCols = []string{"一级商品分类", "二级商品分类", "三级商品分类", "四级商品分类", "五级商品分类"}

var customPriceKinds = []struct {
	name         string
	productField string
	unitField    string
}{
	{"默认订货价", "", "default_price"},
	{"批发价", "wholesale_price", "wholesale_price"},
	{"零售价", "retail_price", "retail_price"},
	{"参考采购价", "purchase_price", "ref_price"},
	{"最低售价", "min_sale_price", "min_sale_price"},
}

const customAuxUnitMax = 3

func customRecognizedHeaders() map[string]bool {
	set := map[string]bool{
		"商品编号": true, "商品名称": true, "商品品牌": true, "拼音码": true,
		"商城展示名称": true, "搜索关键词": true, "默认供应商": true, "商城排序权重": true,
		"出库仓库": true, "商品描述": true, "专题分类": true, "商品标签": true,
		"起订量": true, "限订量": true,
		"按起订量倍数订购": true, "无需管控可用库存": true, "无需管控账面库存": true,
		"上下架状态": true,
		"商品条码": true, "规格备注1": true, "规格备注2": true,
		"重量(kg)": true, "重量（kg）": true,
		"体积(m³)": true, "体积（m³）": true, "体积(m3)": true, "体积（m3）": true,
	}
	for _, col := range customCategoryCols {
		set[col] = true
	}
	for _, pk := range customPriceKinds {
		set[pk.name+"-基本单位价格"] = true
		for n := 1; n <= customAuxUnitMax; n++ {
			set[fmt.Sprintf("%s-辅助单位%d价格", pk.name, n)] = true
		}
	}
	for n := 1; n <= customAuxUnitMax; n++ {
		set[fmt.Sprintf("辅助单位%d-单位名称", n)] = true
	}
	return set
}

type customImportCaches struct {
	categories map[string]uint
	brands     map[string]uint
	tags       map[string]uint
	suppliers  map[string]uint
	warehouses map[string]uint
}

func newCustomImportCaches() *customImportCaches {
	return &customImportCaches{
		categories: map[string]uint{},
		brands:     map[string]uint{},
		tags:       map[string]uint{},
		suppliers:  map[string]uint{},
		warehouses: map[string]uint{},
	}
}

func (caches *customImportCaches) resolveCategoryPath(tx *gorm.DB, companyID uint, path []string) (uint, error) {
	parentID := uint(0)
	for _, name := range path {
		key := fmt.Sprintf("%d|%s", parentID, name)
		if id, ok := caches.categories[key]; ok {
			parentID = id
			continue
		}
		var cat model.ProductCategory
		err := tx.Where("company_id = ? AND name = ? AND parent_id = ?", companyID, name, parentID).First(&cat).Error
		if err == gorm.ErrRecordNotFound {
			cat = model.ProductCategory{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				ParentID:             parentID,
				Name:                 name,
				Status:               1,
			}
			if err := tx.Create(&cat).Error; err != nil {
				return 0, err
			}
		} else if err != nil {
			return 0, err
		}
		caches.categories[key] = cat.ID
		parentID = cat.ID
	}
	return parentID, nil
}

func (caches *customImportCaches) resolveBrand(tx *gorm.DB, companyID uint, name string) (uint, error) {
	if id, ok := caches.brands[name]; ok {
		return id, nil
	}
	var brand model.Brand
	err := tx.Where("company_id = ? AND name = ?", companyID, name).First(&brand).Error
	if err == gorm.ErrRecordNotFound {
		brand = model.Brand{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Name:                 name,
			Status:               1,
		}
		if err := tx.Create(&brand).Error; err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}
	caches.brands[name] = brand.ID
	return brand.ID, nil
}

func (caches *customImportCaches) resolveTag(tx *gorm.DB, companyID uint, name string) (uint, error) {
	if id, ok := caches.tags[name]; ok {
		return id, nil
	}
	var tag model.ProductTag
	err := tx.Where("company_id = ? AND name = ?", companyID, name).First(&tag).Error
	if err == gorm.ErrRecordNotFound {
		tag = model.ProductTag{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Name:                 name,
			Type:                 1,
			Status:               1,
		}
		if err := tx.Create(&tag).Error; err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}
	caches.tags[name] = tag.ID
	return tag.ID, nil
}

func (caches *customImportCaches) resolveSupplier(tx *gorm.DB, companyID uint, name string) (uint, error) {
	if id, ok := caches.suppliers[name]; ok {
		return id, nil
	}
	var supplier model.Supplier
	err := tx.Where("company_id = ? AND name = ?", companyID, name).First(&supplier).Error
	if err == gorm.ErrRecordNotFound {
		supplier = model.Supplier{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Name:                 name,
			Status:               1,
		}
		if err := tx.Create(&supplier).Error; err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}
	caches.suppliers[name] = supplier.ID
	return supplier.ID, nil
}

func (caches *customImportCaches) resolveWarehouse(tx *gorm.DB, companyID uint, name string) (uint, error) {
	if id, ok := caches.warehouses[name]; ok {
		return id, nil
	}
	var wh model.Warehouse
	err := tx.Where("company_id = ? AND name = ?", companyID, name).First(&wh).Error
	if err == gorm.ErrRecordNotFound {
		wh = model.Warehouse{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Name:                 name,
			Status:               1,
		}
		if err := tx.Create(&wh).Error; err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}
	caches.warehouses[name] = wh.ID
	return wh.ID, nil
}

func parseYesNoCell(v string) (int8, bool) {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "是", "Y", "1":
		return 1, true
	case "否", "N", "0":
		return 0, true
	}
	return 0, false
}

func parseShelfCell(v string) (int8, bool) {
	switch strings.TrimSpace(v) {
	case "上架":
		return 1, true
	case "下架":
		return 0, true
	}
	return 0, false
}

// ImportProductsCustom 自定义模板导入（按表头匹配，仅更新已有商品）
func (h *ProductHandler) ImportProductsCustom(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	overwriteEmpty := c.PostForm("overwriteEmpty") == "1"

	f, rows, ok := h.readImportFile(c)
	if !ok {
		return
	}
	defer f.Close()

	recognized := customRecognizedHeaders()
	colIndex := map[string]int{}
	for idx, header := range rows[0] {
		header = strings.TrimSpace(header)
		if recognized[header] {
			if _, exists := colIndex[header]; !exists {
				colIndex[header] = idx
			}
		}
	}

	cell := func(raw []string, name string) (string, bool) {
		idx, exists := colIndex[name]
		if !exists {
			return "", false
		}
		if idx < len(raw) {
			return strings.TrimSpace(raw[idx]), true
		}
		return "", true
	}
	cellFirst := func(raw []string, names ...string) (string, bool) {
		for _, name := range names {
			if v, m := cell(raw, name); m {
				return v, true
			}
		}
		return "", false
	}

	var defaultLevelID uint
	h.db.Model(&model.PriceLevel{}).Where("company_id = ? AND is_default = ?", companyID, true).Select("id").Scan(&defaultLevelID)

	caches := newCustomImportCaches()
	updated := 0
	failed := 0
	errors := make([]PriceImportError, 0, 20)

	for i := 1; i < len(rows); i++ {
		raw := rows[i]
		rowNo := i + 1
		blank := true
		for _, v := range raw {
			if strings.TrimSpace(v) != "" {
				blank = false
				break
			}
		}
		if blank {
			continue
		}

		code, _ := cell(raw, "商品编号")
		fail := func(reason string) {
			failed++
			if len(errors) < 20 {
				errors = append(errors, PriceImportError{Row: rowNo, Code: code, Reason: reason})
			}
		}
		if code == "" {
			fail("商品编号为空，无法更新")
			continue
		}

		var product model.Product
		var specItem model.ProductSpecItem
		isSpecRow := false
		err := h.db.Where("company_id = ? AND code = ?", companyID, code).First(&product).Error
		if err == gorm.ErrRecordNotFound {
			if err := h.db.Model(&model.ProductSpecItem{}).
				Joins("JOIN products ON products.id = product_spec_items.product_id AND products.deleted_at IS NULL").
				Where("product_spec_items.company_id = ? AND product_spec_items.code = ?", companyID, code).
				Order("product_spec_items.id").
				First(&specItem).Error; err != nil {
				fail("商品编号不存在，无法更新")
				continue
			}
			isSpecRow = true
			if err := h.db.Where("id = ? AND company_id = ?", specItem.ProductID, companyID).First(&product).Error; err != nil {
				fail("规格所属商品不存在")
				continue
			}
		} else if err != nil {
			fail("查询商品失败")
			continue
		}

		updates := map[string]interface{}{}
		specUpdates := map[string]interface{}{}
		unitUpdates := map[uint]map[string]interface{}{}
		var defaultPrice *float64
		tagMatched := false
		tagNames := make([]string, 0)
		var catPath []string
		rowErr := ""

		priceValue := func(colName, v string) (float64, bool) {
			if v == "" {
				if overwriteEmpty {
					return 0, true
				}
				return 0, false
			}
			fv, err := strconv.ParseFloat(v, 64)
			if err != nil || fv < 0 {
				rowErr = "「" + colName + "」数值格式错误"
				return 0, false
			}
			return fv, true
		}
		setStr := func(target map[string]interface{}, field string, v string, matched bool) {
			if rowErr != "" || !matched {
				return
			}
			if v == "" && !overwriteEmpty {
				return
			}
			target[field] = v
		}
		setNum := func(target map[string]interface{}, colName, field string, v string, matched bool) {
			if rowErr != "" || !matched {
				return
			}
			if fv, apply := priceValue(colName, v); apply {
				target[field] = fv
			}
		}
		setBool := func(colName, field string, v string, matched bool) {
			if rowErr != "" || !matched {
				return
			}
			if v == "" {
				if overwriteEmpty {
					updates[field] = int8(0)
				}
				return
			}
			bv, ok := parseYesNoCell(v)
			if !ok {
				rowErr = "「" + colName + "」取值无效（是/否）"
				return
			}
			updates[field] = bv
		}

		if v, m := cell(raw, "商品名称"); m {
			setStr(updates, "name", v, true)
		}
		if v, m := cell(raw, "拼音码"); m {
			setStr(updates, "pinyin_code", v, true)
		}
		if v, m := cell(raw, "商城展示名称"); m {
			setStr(updates, "mall_name", v, true)
		}
		if v, m := cell(raw, "搜索关键词"); m {
			setStr(updates, "search_keywords", v, true)
		}
		if v, m := cell(raw, "商品描述"); m {
			setStr(updates, "description", v, true)
		}
		if v, m := cell(raw, "专题分类"); m {
			setStr(updates, "topic_category", v, true)
		}
		if v, m := cell(raw, "商城排序权重"); m && rowErr == "" {
			if v == "" {
				if overwriteEmpty {
					updates["mall_sort_weight"] = 0
				}
			} else if iv, err := strconv.Atoi(v); err != nil {
				rowErr = "「商城排序权重」格式错误"
			} else {
				updates["mall_sort_weight"] = iv
			}
		}
		if v, m := cell(raw, "起订量"); m {
			setNum(updates, "起订量", "min_order_qty", v, true)
		}
		if v, m := cell(raw, "限订量"); m {
			setNum(updates, "限订量", "max_order_qty", v, true)
		}
		if v, m := cell(raw, "按起订量倍数订购"); m {
			setBool("按起订量倍数订购", "order_by_multiple", v, true)
		}
		if v, m := cell(raw, "无需管控可用库存"); m {
			setBool("无需管控可用库存", "no_available_stock_control", v, true)
		}
		if v, m := cell(raw, "无需管控账面库存"); m {
			setBool("无需管控账面库存", "no_book_stock_control", v, true)
		}
		if v, m := cell(raw, "上下架状态"); m && rowErr == "" {
			if v == "" {
				if overwriteEmpty {
					updates["status"] = int8(0)
					if isSpecRow {
						specUpdates["on_shelf"] = int8(0)
					}
				}
			} else if sv, ok := parseShelfCell(v); ok {
				updates["status"] = sv
				if isSpecRow {
					specUpdates["on_shelf"] = sv
				}
			} else {
				rowErr = "「上下架状态」取值无效（上架/下架）"
			}
		}

		// 分类链（整组为空则不更新）
		if rowErr == "" {
			levels := make([]string, 5)
			lastFilled := -1
			for lvl := 0; lvl < 5; lvl++ {
				if v, m := cell(raw, customCategoryCols[lvl]); m {
					levels[lvl] = v
					if v != "" {
						lastFilled = lvl
					}
				}
			}
			if lastFilled >= 0 {
				gap := -1
				for lvl := 0; lvl < lastFilled; lvl++ {
					if levels[lvl] == "" {
						gap = lvl
						break
					}
				}
				if gap >= 0 {
					rowErr = "商品分类必须先填写「" + customCategoryCols[gap] + "」"
				} else {
					catPath = levels[:lastFilled+1]
				}
			}
		}

		// 价格
		if rowErr == "" {
			for _, pk := range customPriceKinds {
				if v, m := cell(raw, pk.name+"-基本单位价格"); m {
					if fv, apply := priceValue(pk.name+"-基本单位价格", v); apply {
						if pk.productField != "" {
							updates[pk.productField] = fv
						} else {
							pv := fv
							defaultPrice = &pv
						}
					}
					if rowErr != "" {
						break
					}
				}
			}
		}

		// 商品标签（| 分隔，全量替换）
		if v, m := cell(raw, "商品标签"); m && rowErr == "" {
			if v == "" {
				if overwriteEmpty {
					tagMatched = true
				}
			} else {
				tagMatched = true
				for _, name := range strings.Split(v, "|") {
					if name = strings.TrimSpace(name); name != "" {
						tagNames = append(tagNames, name)
					}
				}
			}
		}

		// 规格级字段
		if isSpecRow && rowErr == "" {
			if v, m := cell(raw, "商品条码"); m {
				setStr(specUpdates, "barcode", v, true)
			}
			if v, m := cellFirst(raw, "重量(kg)", "重量（kg）"); m {
				setNum(specUpdates, "重量(kg)", "weight", v, true)
			}
			if v, m := cellFirst(raw, "体积(m³)", "体积（m³）", "体积(m3)", "体积（m3）"); m {
				setNum(specUpdates, "体积(m³)", "volume", v, true)
			}
			if v, m := cell(raw, "规格备注1"); m {
				setStr(specUpdates, "remark1", v, true)
			}
			if v, m := cell(raw, "规格备注2"); m {
				setStr(specUpdates, "remark2", v, true)
			}
		}

		if rowErr != "" {
			fail(rowErr)
			continue
		}

		// 辅助单位价格：按「辅助单位N-单位名称」定位单位
		var productUnits []model.ProductUnit
		unitsLoaded := false
		for n := 1; n <= customAuxUnitMax && rowErr == ""; n++ {
			unitName, nameMatched := cell(raw, fmt.Sprintf("辅助单位%d-单位名称", n))
			for _, pk := range customPriceKinds {
				v, m := cell(raw, fmt.Sprintf("%s-辅助单位%d价格", pk.name, n))
				if !m {
					continue
				}
				if !nameMatched || unitName == "" {
					continue
				}
				if !unitsLoaded {
					h.db.Where("product_id = ?", product.ID).Find(&productUnits)
					unitsLoaded = true
				}
				var unitID uint
				for _, u := range productUnits {
					if u.Name == unitName {
						unitID = u.ID
						break
					}
				}
				if unitID == 0 {
					continue
				}
				if fv, apply := priceValue(fmt.Sprintf("%s-辅助单位%d价格", pk.name, n), v); apply {
					if unitUpdates[unitID] == nil {
						unitUpdates[unitID] = map[string]interface{}{}
					}
					unitUpdates[unitID][pk.unitField] = fv
				}
			}
		}
		if rowErr != "" {
			fail(rowErr)
			continue
		}

		err = h.db.Transaction(func(tx *gorm.DB) error {
			if len(catPath) > 0 {
				catID, err := caches.resolveCategoryPath(tx, companyID, catPath)
				if err != nil {
					return err
				}
				updates["category_id"] = catID
			}
			if v, m := cell(raw, "商品品牌"); m {
				if v == "" {
					if overwriteEmpty {
						updates["brand_id"] = 0
					}
				} else {
					brandID, err := caches.resolveBrand(tx, companyID, v)
					if err != nil {
						return err
					}
					updates["brand_id"] = brandID
				}
			}
			if v, m := cell(raw, "默认供应商"); m {
				if v == "" {
					if overwriteEmpty {
						updates["supplier_id"] = 0
					}
				} else {
					supplierID, err := caches.resolveSupplier(tx, companyID, v)
					if err != nil {
						return err
					}
					updates["supplier_id"] = supplierID
				}
			}
			if v, m := cell(raw, "出库仓库"); m {
				if v == "" {
					if overwriteEmpty {
						updates["warehouse_id"] = 0
					}
				} else {
					warehouseID, err := caches.resolveWarehouse(tx, companyID, v)
					if err != nil {
						return err
					}
					updates["warehouse_id"] = warehouseID
				}
			}

			if len(updates) > 0 {
				if err := tx.Model(&model.Product{}).
					Where("id = ? AND company_id = ?", product.ID, companyID).
					Updates(updates).Error; err != nil {
					return err
				}
			}
			if isSpecRow && len(specUpdates) > 0 {
				if err := tx.Model(&model.ProductSpecItem{}).
					Where("id = ? AND company_id = ?", specItem.ID, companyID).
					Updates(specUpdates).Error; err != nil {
					return err
				}
			}
			for unitID, fields := range unitUpdates {
				if err := tx.Model(&model.ProductUnit{}).
					Where("id = ? AND product_id = ?", unitID, product.ID).
					Updates(fields).Error; err != nil {
					return err
				}
			}

			if defaultPrice != nil && defaultLevelID > 0 {
				var pp model.ProductPrice
				err := tx.Where("product_id = ? AND unit_id = 0 AND level_id = ? AND customer_id = 0", product.ID, defaultLevelID).First(&pp).Error
				if err == nil {
					if err := tx.Model(&pp).Update("price", *defaultPrice).Error; err != nil {
						return err
					}
				} else if err == gorm.ErrRecordNotFound {
					pp = model.ProductPrice{
						ProductID: product.ID,
						UnitID:    0,
						LevelID:   defaultLevelID,
						Price:     *defaultPrice,
						Status:    1,
					}
					if err := tx.Create(&pp).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if tagMatched {
				if err := tx.Where("company_id = ? AND product_id = ?", companyID, product.ID).
					Delete(&model.ProductTagRelation{}).Error; err != nil {
					return err
				}
				for _, name := range tagNames {
					tagID, err := caches.resolveTag(tx, companyID, name)
					if err != nil {
						return err
					}
					rel := model.ProductTagRelation{
						BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
						ProductID:            product.ID,
						TagID:                tagID,
					}
					if err := tx.Create(&rel).Error; err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			log.Error().Err(err).Int("row", rowNo).Msg("import custom update product failed")
			fail("更新失败")
			continue
		}
		updated++
	}

	response.Ok(c, gin.H{
		"created": 0,
		"updated": updated,
		"failed":  failed,
		"errors":  errors,
	})
}
