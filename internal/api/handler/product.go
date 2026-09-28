package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// ProductHandler 商品资料处理器
type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

func (h *ProductHandler) RegisterRoutes(r *gin.RouterGroup) {
	product := r.Group("/products")
	{
		// 商品分类
		product.GET("/categories", h.ListCategories)
		product.GET("/categories/tree", h.GetCategoryTree)
		product.POST("/categories", h.CreateCategory)
		product.PUT("/categories/:id", h.UpdateCategory)
		product.DELETE("/categories/:id", h.DeleteCategory)

		// 品牌
		product.GET("/brands", h.ListBrands)
		product.POST("/brands", h.CreateBrand)
		product.PUT("/brands/:id", h.UpdateBrand)
		product.DELETE("/brands/:id", h.DeleteBrand)

		// 商品
		product.GET("", h.ListProducts)
		product.POST("", h.CreateProduct)
		product.GET("/:id", h.GetProduct)
		product.PUT("/:id", h.UpdateProduct)
		product.DELETE("/:id", h.DeleteProduct)
		product.PUT("/:id/status", h.UpdateProductStatus)
		product.PUT("/:id/units", h.ReplaceProductUnits)
		product.PUT("/:id/barcodes", h.ReplaceProductBarcodes)
	}
}

// ==================== 商品分类 ====================

type CategoryListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

type CreateCategoryReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"required,max=64"`
	Code     string `json:"code" binding:"max=64"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateCategoryReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"max=64"`
	Code     string `json:"code" binding:"max=64"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	var req CategoryListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.ProductCategory{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.ProductCategory
	query.Order("sort ASC, created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductHandler) GetCategoryTree(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var list []model.ProductCategory
	h.db.Where("company_id = ?", companyID).Order("sort ASC, created_at DESC").Find(&list)

	tree := buildCategoryTree(list, 0)
	response.Ok(c, tree)
}

type CategoryTreeNode struct {
	model.ProductCategory
	Children []CategoryTreeNode `json:"children"`
}

func buildCategoryTree(cats []model.ProductCategory, parentID uint) []CategoryTreeNode {
	var result []CategoryTreeNode
	for _, cat := range cats {
		if cat.ParentID == parentID {
			node := CategoryTreeNode{
				ProductCategory: cat,
				Children:        buildCategoryTree(cats, cat.ID),
			}
			result = append(result, node)
		}
	}
	return result
}

func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	if req.ParentID > 0 {
		var parent model.ProductCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父分类不存在")
			return
		}
	}

	cat := model.ProductCategory{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		ParentID:             req.ParentID,
		Name:                 req.Name,
		Code:                 req.Code,
		Sort:                 req.Sort,
		Status:               req.Status,
	}
	if err := h.db.Create(&cat).Error; err != nil {
		log.Error().Err(err).Msg("create category failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, cat)
}

func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var cat model.ProductCategory
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&cat).Error; err != nil {
		response.NotFound(c, "分类不存在")
		return
	}

	if req.ParentID > 0 {
		if uint(id) == req.ParentID {
			response.BadRequest(c, "不能将自身设为父分类")
			return
		}
		var parent model.ProductCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父分类不存在")
			return
		}
	}

	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Code != "" {
		cat.Code = req.Code
	}
	cat.ParentID = req.ParentID
	cat.Sort = req.Sort
	cat.Status = req.Status

	if err := h.db.Save(&cat).Error; err != nil {
		log.Error().Err(err).Msg("update category failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, cat)
}

func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var childCount int64
	h.db.Model(&model.ProductCategory{}).Where("parent_id = ? AND company_id = ?", id, companyID).Count(&childCount)
	if childCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该分类下存在子分类，无法删除")
		return
	}

	var productCount int64
	h.db.Model(&model.Product{}).Where("category_id = ? AND company_id = ?", id, companyID).Count(&productCount)
	if productCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该分类下存在商品，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.ProductCategory{}).Error; err != nil {
		log.Error().Err(err).Msg("delete category failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 品牌 ====================

type BrandListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Keyword  string `form:"keyword"`
}

type CreateBrandReq struct {
	Name        string `json:"name" binding:"required,max=64"`
	Code        string `json:"code" binding:"max=64"`
	Description string `json:"description" binding:"max=255"`
	Status      int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateBrandReq struct {
	Name        string `json:"name" binding:"max=64"`
	Code        string `json:"code" binding:"max=64"`
	Description string `json:"description" binding:"max=255"`
	Status      int8   `json:"status" binding:"oneof=0 1"`
}

func (h *ProductHandler) ListBrands(c *gin.Context) {
	var req BrandListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Brand{}).Where("company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.Brand
	query.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductHandler) CreateBrand(c *gin.Context) {
	var req CreateBrandReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	brand := model.Brand{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Description:          req.Description,
		Status:               req.Status,
	}
	if err := h.db.Create(&brand).Error; err != nil {
		log.Error().Err(err).Msg("create brand failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, brand)
}

func (h *ProductHandler) UpdateBrand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateBrandReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var brand model.Brand
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&brand).Error; err != nil {
		response.NotFound(c, "品牌不存在")
		return
	}

	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Code != "" {
		brand.Code = req.Code
	}
	brand.Description = req.Description
	brand.Status = req.Status

	if err := h.db.Save(&brand).Error; err != nil {
		log.Error().Err(err).Msg("update brand failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, brand)
}

func (h *ProductHandler) DeleteBrand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var count int64
	h.db.Model(&model.Product{}).Where("brand_id = ? AND company_id = ?", id, companyID).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeBadRequest, "该品牌下存在商品，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Brand{}).Error; err != nil {
		log.Error().Err(err).Msg("delete brand failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 商品 ====================

type ProductListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CategoryID uint   `form:"categoryId"`
	BrandID    uint   `form:"brandId"`
	Status     int8   `form:"status"`
}

type CreateProductReq struct {
	CategoryID     uint    `json:"categoryId" binding:"required"`
	BrandID        uint    `json:"brandId"`
	SeriesID       uint    `json:"seriesId"`
	Name           string  `json:"name" binding:"required,max=128"`
	Code           string  `json:"code" binding:"max=64"`
	Barcode        string  `json:"barcode" binding:"max=64"`
	Specification  string  `json:"specification" binding:"max=128"`
	Unit           string  `json:"unit" binding:"required,max=32"`
	PurchasePrice  float64 `json:"purchasePrice"`
	RetailPrice    float64 `json:"retailPrice"`
	WholesalePrice float64 `json:"wholesalePrice"`
	MinStock       float64 `json:"minStock"`
	MaxStock       float64 `json:"maxStock"`
	Description    string  `json:"description" binding:"max=500"`
	Status         int8    `json:"status" binding:"oneof=0 1"`
	// 辅助单位（可选）
	Units []ProductUnitReq `json:"units"`
}

type ProductUnitReq struct {
	Name       string  `json:"name" binding:"required,max=32"`
	Conversion float64 `json:"conversion" binding:"required,gt=0"`
	IsDefault  bool    `json:"isDefault"`
	Barcode    string  `json:"barcode" binding:"max=64"`
}

type UpdateProductReq struct {
	CategoryID     uint    `json:"categoryId"`
	BrandID        uint    `json:"brandId"`
	SeriesID       uint    `json:"seriesId"`
	Name           string  `json:"name" binding:"max=128"`
	Code           string  `json:"code" binding:"max=64"`
	Barcode        string  `json:"barcode" binding:"max=64"`
	Specification  string  `json:"specification" binding:"max=128"`
	Unit           string  `json:"unit" binding:"max=32"`
	PurchasePrice  float64 `json:"purchasePrice"`
	RetailPrice    float64 `json:"retailPrice"`
	WholesalePrice float64 `json:"wholesalePrice"`
	MinStock       float64 `json:"minStock"`
	MaxStock       float64 `json:"maxStock"`
	Description    string  `json:"description" binding:"max=500"`
	Status         int8    `json:"status" binding:"oneof=0 1"`
}

type ProductResp struct {
	model.Product
	CategoryName string  `json:"categoryName"`
	BrandName    string  `json:"brandName"`
	TotalStock   float64 `json:"totalStock"`
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	var req ProductListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	query := h.db.Model(&model.Product{}).Where("products.company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("products.name LIKE ? OR products.code LIKE ? OR products.barcode LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.CategoryID > 0 {
		query = query.Where("products.category_id = ?", req.CategoryID)
	}
	if req.BrandID > 0 {
		query = query.Where("products.brand_id = ?", req.BrandID)
	}
	if req.Status == 0 || req.Status == 1 {
		query = query.Where("products.status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var list []ProductResp
	query.Select("products.*, product_categories.name as category_name, brands.name as brand_name, (SELECT COALESCE(SUM(stocks.quantity),0) FROM stocks WHERE stocks.product_id = products.id) as total_stock").
		Joins("LEFT JOIN product_categories ON product_categories.id = products.category_id").
		Joins("LEFT JOIN brands ON brands.id = products.brand_id").
		Order("products.created_at DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	// 获取辅助单位
	var units []model.ProductUnit
	h.db.Where("product_id = ? AND status = 1", id).Find(&units)

	// 获取条码
	var barcodes []model.ProductBarcode
	h.db.Where("product_id = ? AND status = 1", id).Find(&barcodes)

	// 获取总库存（所有仓库合计）
	var totalStock float64
	h.db.Model(&model.Stock{}).Where("product_id = ? AND company_id = ?", id, companyID).
		Select("COALESCE(SUM(quantity),0)").Scan(&totalStock)

	response.Ok(c, gin.H{
		"product":    product,
		"units":      units,
		"barcodes":   barcodes,
		"totalStock": totalStock,
	})
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	// 验证分类
	var category model.ProductCategory
	if err := h.db.Where("id = ? AND company_id = ?", req.CategoryID, companyID).First(&category).Error; err != nil {
		response.BadRequest(c, "商品分类不存在")
		return
	}
	// 验证品牌
	if req.BrandID > 0 {
		var brand model.Brand
		if err := h.db.Where("id = ? AND company_id = ?", req.BrandID, companyID).First(&brand).Error; err != nil {
			response.BadRequest(c, "品牌不存在")
			return
		}
	}

	product := model.Product{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		CategoryID:           req.CategoryID,
		BrandID:              req.BrandID,
		SeriesID:             req.SeriesID,
		Name:                 req.Name,
		Code:                 req.Code,
		Barcode:              req.Barcode,
		Specification:        req.Specification,
		Unit:                 req.Unit,
		PurchasePrice:        req.PurchasePrice,
		RetailPrice:          req.RetailPrice,
		WholesalePrice:       req.WholesalePrice,
		MinStock:             req.MinStock,
		MaxStock:             req.MaxStock,
		Description:          req.Description,
		Status:               req.Status,
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		// 创建辅助单位
		for _, u := range req.Units {
			unit := model.ProductUnit{
				ProductID:  product.ID,
				Name:       u.Name,
				Conversion: u.Conversion,
				IsDefault:  u.IsDefault,
				Status:     1,
			}
			if err := tx.Create(&unit).Error; err != nil {
				return err
			}
			// 创建条码
			if u.Barcode != "" {
				bc := model.ProductBarcode{
					ProductID: product.ID,
					UnitID:    unit.ID,
					Barcode:   u.Barcode,
					Status:    1,
				}
				if err := tx.Create(&bc).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create product failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	if req.CategoryID > 0 {
		var category model.ProductCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.CategoryID, companyID).First(&category).Error; err != nil {
			response.BadRequest(c, "商品分类不存在")
			return
		}
		product.CategoryID = req.CategoryID
	}
	if req.BrandID > 0 {
		var brand model.Brand
		if err := h.db.Where("id = ? AND company_id = ?", req.BrandID, companyID).First(&brand).Error; err != nil {
			response.BadRequest(c, "品牌不存在")
			return
		}
		product.BrandID = req.BrandID
	}
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Code != "" {
		product.Code = req.Code
	}
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}
	if req.Specification != "" {
		product.Specification = req.Specification
	}
	if req.Unit != "" {
		product.Unit = req.Unit
	}
	product.PurchasePrice = req.PurchasePrice
	product.RetailPrice = req.RetailPrice
	product.WholesalePrice = req.WholesalePrice
	product.MinStock = req.MinStock
	product.MaxStock = req.MaxStock
	if req.Description != "" {
		product.Description = req.Description
	}
	product.Status = req.Status

	if err := h.db.Save(&product).Error; err != nil {
		log.Error().Err(err).Msg("update product failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	// 检查是否有库存记录（简化：暂不实现）
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.Product{}).Error; err != nil {
		log.Error().Err(err).Msg("delete product failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (h *ProductHandler) UpdateProductStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Status int8 `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	if err := h.db.Model(&model.Product{}).Where("id = ? AND company_id = ?", id, companyID).Update("status", req.Status).Error; err != nil {
		log.Error().Err(err).Msg("update product status failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

// ReplaceProductUnits 替换商品的辅助单位（全量覆盖）
func (h *ProductHandler) ReplaceProductUnits(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Units []ProductUnitReq `json:"units" binding:"omitempty,dive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	companyID := middleware.GetCompanyID(c)

	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", product.ID).Delete(&model.ProductUnit{}).Error; err != nil {
			return err
		}
		for _, u := range req.Units {
			unit := model.ProductUnit{
				ProductID:  product.ID,
				Name:       u.Name,
				Conversion: u.Conversion,
				IsDefault:  u.IsDefault,
				Status:     1,
			}
			if err := tx.Create(&unit).Error; err != nil {
				return err
			}
			// 单位自带条码时同步到条码表
			if u.Barcode != "" {
				if err := tx.Create(&model.ProductBarcode{
					ProductID: product.ID,
					UnitID:    unit.ID,
					Barcode:   u.Barcode,
					Status:    1,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("replace product units failed")
		response.ServerError(c, "保存失败")
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}

// ReplaceProductBarcodes 替换商品的条码（全量覆盖）
func (h *ProductHandler) ReplaceProductBarcodes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Barcodes []struct {
			UnitID  uint   `json:"unitId"`
			Barcode string `json:"barcode" binding:"required,max=64"`
		} `json:"barcodes" binding:"omitempty,dive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	companyID := middleware.GetCompanyID(c)

	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", product.ID).Delete(&model.ProductBarcode{}).Error; err != nil {
			return err
		}
		for _, b := range req.Barcodes {
			if err := tx.Create(&model.ProductBarcode{
				ProductID: product.ID,
				UnitID:    b.UnitID,
				Barcode:   b.Barcode,
				Status:    1,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("replace product barcodes failed")
		response.ServerError(c, "保存失败")
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}
