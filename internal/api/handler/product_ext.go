package handler

import (
	"path/filepath"
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

// ImportProductsCustom 自定义模板导入（按表头匹配，支持更新）
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

	colIndex := map[string]int{}
	for idx, header := range rows[0] {
		header = strings.TrimSpace(header)
		switch header {
		case "商品编号", "商品名称", "商品分类", "商品品牌", "商品单位", "条码", "拼音码", "商城展示名称", "批发价", "零售价", "采购价", "规格":
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

	created := 0
	updated := 0
	failed := 0
	errors := make([]PriceImportError, 0, 20)

	for i := 1; i < len(rows); i++ {
		raw := rows[i]
		rowNo := i + 1
		row := productImportRow{}
		row.Code, _ = cell(raw, "商品编号")
		row.Name, _ = cell(raw, "商品名称")
		row.Category, _ = cell(raw, "商品分类")
		row.Brand, _ = cell(raw, "商品品牌")
		row.Unit, _ = cell(raw, "商品单位")
		row.Barcode, _ = cell(raw, "条码")
		row.PinyinCode, _ = cell(raw, "拼音码")
		row.MallName, _ = cell(raw, "商城展示名称")
		row.Wholesale, _ = cell(raw, "批发价")
		row.Retail, _ = cell(raw, "零售价")
		row.Purchase, _ = cell(raw, "采购价")
		row.Spec, _ = cell(raw, "规格")
		if row.empty() {
			continue
		}
		fail := func(reason string) {
			failed++
			if len(errors) < 20 {
				errors = append(errors, PriceImportError{Row: rowNo, Code: row.Code, Reason: reason})
			}
		}

		var product model.Product
		exists := false
		if row.Code != "" {
			if err := h.db.Where("company_id = ? AND code = ?", companyID, row.Code).First(&product).Error; err == nil {
				exists = true
			} else if err != gorm.ErrRecordNotFound {
				fail("查询商品失败")
				continue
			}
		}

		if !exists {
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
			continue
		}

		updates := map[string]interface{}{}
		rowFailed := false
		setStr := func(col, field, value string, matched bool) {
			if rowFailed || !matched {
				return
			}
			if value == "" && !overwriteEmpty {
				return
			}
			updates[field] = value
		}
		setPrice := func(col, field, value string, matched bool) {
			if rowFailed || !matched {
				return
			}
			if value == "" {
				if overwriteEmpty {
					updates[field] = 0
				}
				return
			}
			v, err := parsePriceCell(value)
			if err != nil {
				fail(col + err.Error())
				rowFailed = true
				return
			}
			updates[field] = *v
		}

		if v, m := cell(raw, "商品名称"); m {
			setStr("商品名称", "name", v, true)
		}
		if v, m := cell(raw, "商品单位"); m {
			setStr("商品单位", "unit", v, true)
		}
		if v, m := cell(raw, "条码"); m {
			setStr("条码", "barcode", v, true)
		}
		if v, m := cell(raw, "拼音码"); m {
			setStr("拼音码", "pinyin_code", v, true)
		}
		if v, m := cell(raw, "商城展示名称"); m {
			setStr("商城展示名称", "mall_name", v, true)
		}
		if v, m := cell(raw, "规格"); m {
			setStr("规格", "specification", v, true)
		}
		if v, m := cell(raw, "批发价"); m {
			setPrice("批发价", "wholesale_price", v, true)
		}
		if v, m := cell(raw, "零售价"); m {
			setPrice("零售价", "retail_price", v, true)
		}
		if v, m := cell(raw, "采购价"); m {
			setPrice("采购价", "purchase_price", v, true)
		}
		if rowFailed {
			continue
		}

		err := h.db.Transaction(func(tx *gorm.DB) error {
			if v, m := cell(raw, "商品分类"); m {
				if v == "" {
					if overwriteEmpty {
						updates["category_id"] = 0
					}
				} else {
					catID, err := resolveCategoryByName(tx, companyID, v)
					if err != nil {
						return err
					}
					updates["category_id"] = catID
				}
			}
			if v, m := cell(raw, "商品品牌"); m {
				if v == "" {
					if overwriteEmpty {
						updates["brand_id"] = 0
					}
				} else {
					brandID, err := resolveBrandByName(tx, companyID, v)
					if err != nil {
						return err
					}
					updates["brand_id"] = brandID
				}
			}
			if len(updates) == 0 {
				return nil
			}
			return tx.Model(&model.Product{}).
				Where("id = ? AND company_id = ?", product.ID, companyID).
				Updates(updates).Error
		})
		if err != nil {
			log.Error().Err(err).Int("row", rowNo).Msg("import update product failed")
			fail("更新失败")
			continue
		}
		updated++
	}

	response.Ok(c, gin.H{
		"created": created,
		"updated": updated,
		"failed":  failed,
		"errors":  errors,
	})
}
