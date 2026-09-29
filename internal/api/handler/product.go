package handler

import (
	"fmt"
	"strconv"
	"time"

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
		product.PUT("/categories/:id/status", h.UpdateCategoryStatus)
		product.POST("/categories/sort", h.SortCategories)
		product.POST("/categories/fill-images", h.FillCategoryImages)
		product.DELETE("/categories/:id", h.DeleteCategory)

		// 品牌
		product.GET("/brands", h.ListBrands)
		product.POST("/brands", h.CreateBrand)
		product.POST("/brands/sort", h.SortBrands)
		product.PUT("/brands/:id", h.UpdateBrand)
		product.DELETE("/brands/:id", h.DeleteBrand)

		// 品牌分类
		product.GET("/brand-categories/tree", h.GetBrandCategoryTree)
		product.POST("/brand-categories", h.CreateBrandCategory)
		product.PUT("/brand-categories/:id", h.UpdateBrandCategory)
		product.DELETE("/brand-categories/:id", h.DeleteBrandCategory)

		// 商品
		product.GET("", h.ListProducts)
		product.POST("", h.CreateProduct)
		product.GET("/spec-items", h.ListProductSpecItems)
		product.GET("/topic-categories", h.ListTopicCategories)
		product.POST("/opening-stock", h.CreateOpeningStock)
		product.POST("/import/system", h.ImportProductsSystem)
		product.POST("/import/custom", h.ImportProductsCustom)
		product.GET("/:id", h.GetProduct)
		product.PUT("/:id", h.UpdateProduct)
		product.DELETE("/:id", h.DeleteProduct)
		product.PUT("/:id/status", h.UpdateProductStatus)
		product.PUT("/:id/units", h.ReplaceProductUnits)
		product.PUT("/:id/specs", h.ReplaceProductSpecItems)
		product.GET("/:id/tags", h.GetProductTags)
		product.PUT("/:id/tags", h.ReplaceProductTags)
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
	Image    string `json:"image" binding:"max=255"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateCategoryReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"max=64"`
	Code     string `json:"code" binding:"max=64"`
	Image    string `json:"image" binding:"max=255"`
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

	type categoryCount struct {
		CategoryID uint
		Count      int64
	}
	var counts []categoryCount
	h.db.Model(&model.Product{}).
		Select("category_id, COUNT(*) AS count").
		Where("company_id = ?", companyID).
		Group("category_id").
		Scan(&counts)
	countMap := make(map[uint]int64, len(counts))
	for _, item := range counts {
		countMap[item.CategoryID] = item.Count
	}

	tree := buildCategoryTree(list, 0, countMap)

	parentIDStr, hasParentID := c.GetQuery("parentId")
	if !hasParentID {
		response.Ok(c, tree)
		return
	}
	parentID, err := strconv.ParseUint(parentIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	nodeMap := make(map[uint]*CategoryTreeNode)
	var flatten func(nodes []CategoryTreeNode)
	flatten = func(nodes []CategoryTreeNode) {
		for i := range nodes {
			nodeMap[nodes[i].ID] = &nodes[i]
			flatten(nodes[i].Children)
		}
	}
	flatten(tree)

	childSet := make(map[uint]bool)
	for _, cat := range list {
		childSet[cat.ParentID] = true
	}

	result := make([]CategoryTreeNode, 0)
	for _, cat := range list {
		if cat.ParentID != uint(parentID) {
			continue
		}
		node := CategoryTreeNode{
			ProductCategory: cat,
			Children:        []CategoryTreeNode{},
			HasChildren:     childSet[cat.ID],
		}
		if full, exists := nodeMap[cat.ID]; exists {
			node.ProductCount = full.ProductCount
		}
		result = append(result, node)
	}
	response.Ok(c, result)
}

type CategoryTreeNode struct {
	model.ProductCategory
	ProductCount int64              `json:"productCount"`
	HasChildren  bool               `json:"hasChildren"`
	Children     []CategoryTreeNode `json:"children"`
}

func buildCategoryTree(cats []model.ProductCategory, parentID uint, countMap map[uint]int64) []CategoryTreeNode {
	var result []CategoryTreeNode
	for _, cat := range cats {
		if cat.ParentID == parentID {
			children := buildCategoryTree(cats, cat.ID, countMap)
			total := countMap[cat.ID]
			for _, child := range children {
				total += child.ProductCount
			}
			result = append(result, CategoryTreeNode{
				ProductCategory: cat,
				ProductCount:    total,
				HasChildren:     len(children) > 0,
				Children:        children,
			})
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
		Image:                req.Image,
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
	cat.Image = req.Image
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

type UpdateCategoryStatusReq struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

func (h *ProductHandler) UpdateCategoryStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateCategoryStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	result := h.db.Model(&model.ProductCategory{}).
		Where("id = ? AND company_id = ?", id, companyID).
		Update("status", req.Status)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("update category status failed")
		response.ServerError(c, "更新失败")
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, "分类不存在")
		return
	}
	response.Ok(c, gin.H{"id": id, "status": req.Status})
}

type SortCategoriesReq struct {
	Items []struct {
		ID   uint `json:"id" binding:"required"`
		Sort int  `json:"sort"`
	} `json:"items" binding:"required"`
}

func (h *ProductHandler) SortCategories(c *gin.Context) {
	var req SortCategoriesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			if err := tx.Model(&model.ProductCategory{}).
				Where("id = ? AND company_id = ?", item.ID, companyID).
				Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("sort categories failed")
		response.ServerError(c, "排序失败")
		return
	}
	response.OkWithMessage(c, "排序成功", nil)
}

func (h *ProductHandler) FillCategoryImages(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var cats []model.ProductCategory
	h.db.Where("company_id = ? AND (image = '' OR image IS NULL)", companyID).Find(&cats)

	updated := 0
	for _, cat := range cats {
		var mp model.MallProduct
		err := h.db.
			Select("mall_products.image_url").
			Joins("JOIN products ON products.id = mall_products.product_id").
			Where("mall_products.company_id = ? AND products.category_id = ? AND products.deleted_at IS NULL AND mall_products.image_url != ''", companyID, cat.ID).
			Order("mall_products.id ASC").
			Take(&mp).Error
		if err != nil || mp.ImageURL == "" {
			continue
		}
		if err := h.db.Model(&model.ProductCategory{}).
			Where("id = ? AND company_id = ?", cat.ID, companyID).
			Update("image", mp.ImageURL).Error; err != nil {
			log.Error().Err(err).Uint("categoryId", cat.ID).Msg("fill category image failed")
			continue
		}
		updated++
	}
	response.Ok(c, gin.H{"updated": updated})
}

// ==================== 品牌分类 ====================

type CreateBrandCategoryReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"required,max=64"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateBrandCategoryReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"max=64"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type BrandCategoryTreeNode struct {
	model.BrandCategory
	Children []BrandCategoryTreeNode `json:"children"`
}

func buildBrandCategoryTree(cats []model.BrandCategory, parentID uint) []BrandCategoryTreeNode {
	var result []BrandCategoryTreeNode
	for _, cat := range cats {
		if cat.ParentID == parentID {
			result = append(result, BrandCategoryTreeNode{
				BrandCategory: cat,
				Children:      buildBrandCategoryTree(cats, cat.ID),
			})
		}
	}
	return result
}

func (h *ProductHandler) GetBrandCategoryTree(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var list []model.BrandCategory
	h.db.Where("company_id = ?", companyID).Order("sort ASC, created_at DESC").Find(&list)
	response.Ok(c, buildBrandCategoryTree(list, 0))
}

func (h *ProductHandler) CreateBrandCategory(c *gin.Context) {
	var req CreateBrandCategoryReq
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
		var parent model.BrandCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父分类不存在")
			return
		}
	}

	cat := model.BrandCategory{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		ParentID:             req.ParentID,
		Name:                 req.Name,
		Sort:                 req.Sort,
		Status:               req.Status,
	}
	if err := h.db.Create(&cat).Error; err != nil {
		log.Error().Err(err).Msg("create brand category failed")
		response.ServerError(c, "创建失败")
		return
	}
	response.Ok(c, cat)
}

func (h *ProductHandler) UpdateBrandCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req UpdateBrandCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var cat model.BrandCategory
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&cat).Error; err != nil {
		response.NotFound(c, "分类不存在")
		return
	}

	if req.ParentID > 0 {
		if uint(id) == req.ParentID {
			response.BadRequest(c, "不能将自身设为父分类")
			return
		}
		var parent model.BrandCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.ParentID, companyID).First(&parent).Error; err != nil {
			response.BadRequest(c, "父分类不存在")
			return
		}
	}

	if req.Name != "" {
		cat.Name = req.Name
	}
	cat.ParentID = req.ParentID
	cat.Sort = req.Sort
	cat.Status = req.Status

	if err := h.db.Save(&cat).Error; err != nil {
		log.Error().Err(err).Msg("update brand category failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.Ok(c, cat)
}

func (h *ProductHandler) DeleteBrandCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	var childCount int64
	h.db.Model(&model.BrandCategory{}).Where("parent_id = ? AND company_id = ?", id, companyID).Count(&childCount)
	if childCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该分类下存在子分类，无法删除")
		return
	}

	var brandCount int64
	h.db.Model(&model.Brand{}).Where("category_id = ? AND company_id = ?", id, companyID).Count(&brandCount)
	if brandCount > 0 {
		response.Fail(c, response.CodeBadRequest, "该分类下存在品牌，无法删除")
		return
	}

	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).Delete(&model.BrandCategory{}).Error; err != nil {
		log.Error().Err(err).Msg("delete brand category failed")
		response.ServerError(c, "删除失败")
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// ==================== 品牌 ====================

type BrandListReq struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=20"`
	Keyword    string `form:"keyword"`
	CategoryID *uint  `form:"categoryId"`
}

type CreateBrandReq struct {
	Name        string `json:"name" binding:"required,max=64"`
	Code        string `json:"code" binding:"max=64"`
	Description string `json:"description" binding:"max=255"`
	Image       string `json:"image" binding:"max=255"`
	CategoryID  uint   `json:"categoryId"`
	ApplyImage  int8   `json:"applyImage" binding:"oneof=0 1"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateBrandReq struct {
	Name        string `json:"name" binding:"max=64"`
	Code        string `json:"code" binding:"max=64"`
	Description string `json:"description" binding:"max=255"`
	Image       string `json:"image" binding:"max=255"`
	CategoryID  uint   `json:"categoryId"`
	ApplyImage  int8   `json:"applyImage" binding:"oneof=0 1"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status" binding:"oneof=0 1"`
}

type BrandResp struct {
	model.Brand
	CategoryName string `json:"categoryName"`
	ProductCount int64  `json:"productCount"`
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

	query := h.db.Model(&model.Brand{}).Where("brands.company_id = ?", companyID)
	if req.Keyword != "" {
		query = query.Where("brands.name LIKE ? OR brands.code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.CategoryID != nil {
		query = query.Where("brands.category_id = ?", *req.CategoryID)
	}

	var total int64
	query.Count(&total)

	var list []BrandResp
	query.Select("brands.*, brand_categories.name AS category_name, (SELECT COUNT(*) FROM products WHERE products.brand_id = brands.id AND products.company_id = ? AND products.deleted_at IS NULL) AS product_count", companyID).
		Joins("LEFT JOIN brand_categories ON brand_categories.id = brands.category_id").
		Order("brands.sort ASC, brands.created_at DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list)

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

	if req.CategoryID > 0 {
		var cat model.BrandCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.CategoryID, companyID).First(&cat).Error; err != nil {
			response.BadRequest(c, "品牌分类不存在")
			return
		}
	}

	brand := model.Brand{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		Name:                 req.Name,
		Code:                 req.Code,
		Description:          req.Description,
		Image:                req.Image,
		CategoryID:           req.CategoryID,
		ApplyImage:           req.ApplyImage,
		Sort:                 req.Sort,
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

	if req.CategoryID > 0 {
		var cat model.BrandCategory
		if err := h.db.Where("id = ? AND company_id = ?", req.CategoryID, companyID).First(&cat).Error; err != nil {
			response.BadRequest(c, "品牌分类不存在")
			return
		}
	}

	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Code != "" {
		brand.Code = req.Code
	}
	brand.Description = req.Description
	brand.Image = req.Image
	brand.CategoryID = req.CategoryID
	brand.ApplyImage = req.ApplyImage
	brand.Sort = req.Sort
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

type SortBrandsReq struct {
	Items []struct {
		ID   uint `json:"id" binding:"required"`
		Sort int  `json:"sort"`
	} `json:"items" binding:"required"`
}

func (h *ProductHandler) SortBrands(c *gin.Context) {
	var req SortBrandsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			if err := tx.Model(&model.Brand{}).
				Where("id = ? AND company_id = ?", item.ID, companyID).
				Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("sort brands failed")
		response.ServerError(c, "排序失败")
		return
	}
	response.OkWithMessage(c, "排序成功", nil)
}

// ==================== 商品 ====================

type ProductListReq struct {
	Page          int    `form:"page,default=1"`
	PageSize      int    `form:"pageSize,default=20"`
	Keyword       string `form:"keyword"`
	CategoryID    uint   `form:"categoryId"`
	BrandID       uint   `form:"brandId"`
	TagID         uint   `form:"tagId"`
	Status        *int8  `form:"status"`
	HasImage      *int8  `form:"hasImage"`
	Name          string `form:"name"`
	TopicCategory string `form:"topicCategory"`
	StockStatus   string `form:"stockStatus"`
	OrderBy       string `form:"orderBy"`
	Order         string `form:"order"`
}

var productOrderColumns = map[string]string{
	"code":      "products.code",
	"barcode":   "products.barcode",
	"name":      "products.name",
	"createdAt": "products.created_at",
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

	Image                   string  `json:"image" binding:"max=255"`
	MallName                string  `json:"mallName" binding:"max=128"`
	PinyinCode              string  `json:"pinyinCode" binding:"max=128"`
	VideoURL                string  `json:"videoUrl" binding:"max=255"`
	DetailContent           string  `json:"detailContent"`
	TopicCategory           string  `json:"topicCategory" binding:"max=64"`
	MinOrderQty             float64 `json:"minOrderQty"`
	MaxOrderQty             float64 `json:"maxOrderQty"`
	OrderByMultiple         int8    `json:"orderByMultiple" binding:"oneof=0 1"`
	NoAvailableStockControl int8    `json:"noAvailableStockControl" binding:"oneof=0 1"`
	NoBookStockControl      int8    `json:"noBookStockControl" binding:"oneof=0 1"`
	ManualLevelPriceDefault int8    `json:"manualLevelPriceDefault" binding:"oneof=0 1"`
	AuxPriceReverseCalc     int8    `json:"auxPriceReverseCalc" binding:"oneof=0 1"`

	MinSalePrice        float64 `json:"minSalePrice"`
	MallSortWeight      int     `json:"mallSortWeight"`
	SearchKeywords      string  `json:"searchKeywords" binding:"max=255"`
	WarehouseID         uint    `json:"warehouseId"`
	SupplierID          uint    `json:"supplierId"`
	ForbidPurchaseUnits string  `json:"forbidPurchaseUnits" binding:"max=255"`
	// 辅助单位（可选）
	Units []ProductUnitReq `json:"units"`
}

type ProductUnitReq struct {
	Name            string  `json:"name" binding:"required,max=32"`
	Conversion      float64 `json:"conversion" binding:"required,gt=0"`
	IsDefault       bool    `json:"isDefault"`
	Barcode         string  `json:"barcode" binding:"max=64"`
	AllowSale       *int8   `json:"allowSale" binding:"omitempty,oneof=0 1"`
	DefaultSale     int8    `json:"defaultSale" binding:"oneof=0 1"`
	DefaultPurchase int8    `json:"defaultPurchase" binding:"oneof=0 1"`
	StorageType     int8    `json:"storageType" binding:"oneof=0 1 2"`
	RefPrice        float64 `json:"refPrice"`
	WholesalePrice  float64 `json:"wholesalePrice"`
	RetailPrice     float64 `json:"retailPrice"`
	MinSalePrice    float64 `json:"minSalePrice"`
	DefaultPrice    float64 `json:"defaultPrice"`
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

	Image                   string  `json:"image" binding:"max=255"`
	MallName                string  `json:"mallName" binding:"max=128"`
	PinyinCode              string  `json:"pinyinCode" binding:"max=128"`
	VideoURL                string  `json:"videoUrl" binding:"max=255"`
	DetailContent           string  `json:"detailContent"`
	TopicCategory           string  `json:"topicCategory" binding:"max=64"`
	MinOrderQty             float64 `json:"minOrderQty"`
	MaxOrderQty             float64 `json:"maxOrderQty"`
	OrderByMultiple         int8    `json:"orderByMultiple" binding:"oneof=0 1"`
	NoAvailableStockControl int8    `json:"noAvailableStockControl" binding:"oneof=0 1"`
	NoBookStockControl      int8    `json:"noBookStockControl" binding:"oneof=0 1"`
	ManualLevelPriceDefault int8    `json:"manualLevelPriceDefault" binding:"oneof=0 1"`
	AuxPriceReverseCalc     int8    `json:"auxPriceReverseCalc" binding:"oneof=0 1"`

	MinSalePrice        float64 `json:"minSalePrice"`
	MallSortWeight      int     `json:"mallSortWeight"`
	SearchKeywords      string  `json:"searchKeywords" binding:"max=255"`
	WarehouseID         uint    `json:"warehouseId"`
	SupplierID          uint    `json:"supplierId"`
	ForbidPurchaseUnits string  `json:"forbidPurchaseUnits" binding:"max=255"`
}

type ProductResp struct {
	model.Product
	CategoryName  string   `json:"categoryName"`
	BrandName     string   `json:"brandName"`
	WarehouseName string   `json:"warehouseName"`
	SupplierName  string   `json:"supplierName"`
	TotalStock    float64  `json:"totalStock"`
	SpecCount     int64    `json:"specCount"`
	DefaultPrice  float64  `json:"defaultPrice"`
	Tags          []string `json:"tags"`
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
		query = query.Where("products.name LIKE ? OR products.code LIKE ? OR products.barcode LIKE ? OR products.search_keywords LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.CategoryID > 0 {
		query = query.Where("products.category_id = ?", req.CategoryID)
	}
	if req.BrandID > 0 {
		query = query.Where("products.brand_id = ?", req.BrandID)
	}
	if req.TagID > 0 {
		query = query.Joins("JOIN product_tag_relations ON product_tag_relations.product_id = products.id AND product_tag_relations.deleted_at IS NULL AND product_tag_relations.tag_id = ?", req.TagID)
	}
	if req.Status != nil {
		query = query.Where("products.status = ?", *req.Status)
	}
	if req.HasImage != nil {
		if *req.HasImage == 1 {
			query = query.Where("products.image != ''")
		} else if *req.HasImage == 0 {
			query = query.Where("(products.image = '' OR products.image IS NULL)")
		}
	}
	if req.Name != "" {
		query = query.Where("products.name LIKE ?", "%"+req.Name+"%")
	}
	if req.TopicCategory != "" {
		query = query.Where("products.topic_category = ?", req.TopicCategory)
	}
	if req.StockStatus == "in" || req.StockStatus == "out" {
		stockSQL := "COALESCE((SELECT SUM(stocks.quantity) FROM stocks WHERE stocks.product_id = products.id AND stocks.company_id = ?), 0)"
		if req.StockStatus == "in" {
			query = query.Where(stockSQL+" > 0", companyID)
		} else {
			query = query.Where(stockSQL+" <= 0", companyID)
		}
	}

	var total int64
	query.Count(&total)

	orderClause := "products.created_at DESC"
	if col, ok := productOrderColumns[req.OrderBy]; ok {
		dir := "DESC"
		if req.Order == "asc" {
			dir = "ASC"
		}
		orderClause = col + " " + dir
	}

	var levelID uint
	h.db.Model(&model.PriceLevel{}).Where("company_id = ? AND is_default = ?", companyID, true).Select("id").Scan(&levelID)

	var list []ProductResp
	query.Select("products.*, product_categories.name as category_name, brands.name as brand_name, warehouses.name as warehouse_name, suppliers.name as supplier_name, (SELECT COALESCE(SUM(stocks.quantity),0) FROM stocks WHERE stocks.product_id = products.id) as total_stock, (SELECT COUNT(*) FROM product_spec_items WHERE product_spec_items.product_id = products.id AND product_spec_items.deleted_at IS NULL) as spec_count, COALESCE((SELECT product_prices.price FROM product_prices WHERE product_prices.product_id = products.id AND product_prices.unit_id = 0 AND product_prices.level_id = ? AND product_prices.customer_id = 0 LIMIT 1), 0) as default_price", levelID).
		Joins("LEFT JOIN product_categories ON product_categories.id = products.category_id").
		Joins("LEFT JOIN brands ON brands.id = products.brand_id").
		Joins("LEFT JOIN warehouses ON warehouses.id = products.warehouse_id").
		Joins("LEFT JOIN suppliers ON suppliers.id = products.supplier_id").
		Order(orderClause).
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list)

	productIDs := make([]uint, 0, len(list))
	for i := range list {
		productIDs = append(productIDs, list[i].ID)
		list[i].Tags = []string{}
	}
	if len(productIDs) > 0 {
		type tagRow struct {
			ProductID uint
			Name      string
		}
		var tagRows []tagRow
		h.db.Model(&model.ProductTagRelation{}).
			Select("product_tag_relations.product_id, product_tags.name").
			Joins("JOIN product_tags ON product_tags.id = product_tag_relations.tag_id AND product_tags.deleted_at IS NULL").
			Where("product_tag_relations.company_id = ? AND product_tag_relations.product_id IN ?", companyID, productIDs).
			Scan(&tagRows)
		tagMap := make(map[uint][]string, len(productIDs))
		for _, tr := range tagRows {
			tagMap[tr.ProductID] = append(tagMap[tr.ProductID], tr.Name)
		}
		for i := range list {
			if tags, ok := tagMap[list[i].ID]; ok {
				list[i].Tags = tags
			}
		}
	}

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

func (h *ProductHandler) GetProductTags(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	var relations []model.ProductTagRelation
	h.db.Where("company_id = ? AND product_id = ?", companyID, id).Find(&relations)
	tagIDs := make([]uint, 0, len(relations))
	for _, r := range relations {
		tagIDs = append(tagIDs, r.TagID)
	}
	response.Ok(c, gin.H{"tagIds": tagIDs})
}

func (h *ProductHandler) ReplaceProductTags(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		TagIDs []uint `json:"tagIds"`
	}
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
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ? AND product_id = ?", companyID, id).
			Delete(&model.ProductTagRelation{}).Error; err != nil {
			return err
		}
		for _, tagID := range req.TagIDs {
			rel := model.ProductTagRelation{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				ProductID:            product.ID,
				TagID:                tagID,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("replace product tags failed")
		response.ServerError(c, "保存失败")
		return
	}
	response.Ok(c, gin.H{"tagIds": req.TagIDs})
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

	// 获取规格行
	specItems := make([]model.ProductSpecItem, 0)
	h.db.Where("company_id = ? AND product_id = ?", companyID, id).Order("sort ASC, id ASC").Find(&specItems)

	// 获取标签
	var relations []model.ProductTagRelation
	h.db.Where("company_id = ? AND product_id = ?", companyID, id).Find(&relations)
	tagIDs := make([]uint, 0, len(relations))
	for _, r := range relations {
		tagIDs = append(tagIDs, r.TagID)
	}

	response.Ok(c, gin.H{
		"product":    product,
		"units":      units,
		"barcodes":   barcodes,
		"totalStock": totalStock,
		"specItems":  specItems,
		"tagIds":     tagIDs,
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
	// 验证出库仓库
	if req.WarehouseID > 0 {
		var warehouse model.Warehouse
		if err := h.db.Where("id = ? AND company_id = ?", req.WarehouseID, companyID).First(&warehouse).Error; err != nil {
			response.BadRequest(c, "仓库不存在")
			return
		}
	}
	// 验证默认供应商
	if req.SupplierID > 0 {
		var supplier model.Supplier
		if err := h.db.Where("id = ? AND company_id = ?", req.SupplierID, companyID).First(&supplier).Error; err != nil {
			response.BadRequest(c, "供应商不存在")
			return
		}
	}

	product := model.Product{
		BaseModelWithCompany:    model.BaseModelWithCompany{CompanyID: companyID},
		CategoryID:              req.CategoryID,
		BrandID:                 req.BrandID,
		SeriesID:                req.SeriesID,
		Name:                    req.Name,
		Code:                    req.Code,
		Barcode:                 req.Barcode,
		Specification:           req.Specification,
		Unit:                    req.Unit,
		PurchasePrice:           req.PurchasePrice,
		RetailPrice:             req.RetailPrice,
		WholesalePrice:          req.WholesalePrice,
		MinStock:                req.MinStock,
		MaxStock:                req.MaxStock,
		Description:             req.Description,
		Status:                  req.Status,
		Image:                   req.Image,
		MallName:                req.MallName,
		PinyinCode:              req.PinyinCode,
		VideoURL:                req.VideoURL,
		DetailContent:           req.DetailContent,
		TopicCategory:           req.TopicCategory,
		MinOrderQty:             req.MinOrderQty,
		MaxOrderQty:             req.MaxOrderQty,
		OrderByMultiple:         req.OrderByMultiple,
		NoAvailableStockControl: req.NoAvailableStockControl,
		NoBookStockControl:      req.NoBookStockControl,
		ManualLevelPriceDefault: req.ManualLevelPriceDefault,
		AuxPriceReverseCalc:     req.AuxPriceReverseCalc,
		MinSalePrice:            req.MinSalePrice,
		MallSortWeight:          req.MallSortWeight,
		SearchKeywords:          req.SearchKeywords,
		WarehouseID:             req.WarehouseID,
		SupplierID:              req.SupplierID,
		ForbidPurchaseUnits:     req.ForbidPurchaseUnits,
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if product.Code == "" {
			code, err := nextProductCode(tx, companyID)
			if err != nil {
				return err
			}
			product.Code = code
		}
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		if req.Status == 0 {
			if err := tx.Model(&product).Update("status", 0).Error; err != nil {
				return err
			}
		}
		// 创建辅助单位
		for _, u := range req.Units {
			unit := model.ProductUnit{
				ProductID:       product.ID,
				Name:            u.Name,
				Conversion:      u.Conversion,
				IsDefault:       u.IsDefault,
				Barcode:         u.Barcode,
				AllowSale:       1,
				DefaultSale:     u.DefaultSale,
				DefaultPurchase: u.DefaultPurchase,
				StorageType:     u.StorageType,
				RefPrice:        u.RefPrice,
				WholesalePrice:  u.WholesalePrice,
				RetailPrice:     u.RetailPrice,
				MinSalePrice:    u.MinSalePrice,
				DefaultPrice:    u.DefaultPrice,
				Status:          1,
			}
			if u.AllowSale != nil {
				unit.AllowSale = *u.AllowSale
			}
			if unit.StorageType == 0 {
				unit.StorageType = 1
			}
			if err := tx.Select("ProductID", "Name", "Conversion", "IsDefault", "Barcode", "AllowSale", "DefaultSale", "DefaultPurchase", "StorageType", "RefPrice", "WholesalePrice", "RetailPrice", "MinSalePrice", "DefaultPrice", "Status").Create(&unit).Error; err != nil {
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
	if req.WarehouseID > 0 {
		var warehouse model.Warehouse
		if err := h.db.Where("id = ? AND company_id = ?", req.WarehouseID, companyID).First(&warehouse).Error; err != nil {
			response.BadRequest(c, "仓库不存在")
			return
		}
	}
	if req.SupplierID > 0 {
		var supplier model.Supplier
		if err := h.db.Where("id = ? AND company_id = ?", req.SupplierID, companyID).First(&supplier).Error; err != nil {
			response.BadRequest(c, "供应商不存在")
			return
		}
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
	product.Image = req.Image
	product.MallName = req.MallName
	product.PinyinCode = req.PinyinCode
	product.VideoURL = req.VideoURL
	product.DetailContent = req.DetailContent
	product.TopicCategory = req.TopicCategory
	product.MinOrderQty = req.MinOrderQty
	product.MaxOrderQty = req.MaxOrderQty
	product.OrderByMultiple = req.OrderByMultiple
	product.NoAvailableStockControl = req.NoAvailableStockControl
	product.NoBookStockControl = req.NoBookStockControl
	product.ManualLevelPriceDefault = req.ManualLevelPriceDefault
	product.AuxPriceReverseCalc = req.AuxPriceReverseCalc
	product.MinSalePrice = req.MinSalePrice
	product.MallSortWeight = req.MallSortWeight
	product.SearchKeywords = req.SearchKeywords
	product.WarehouseID = req.WarehouseID
	product.SupplierID = req.SupplierID
	product.ForbidPurchaseUnits = req.ForbidPurchaseUnits

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
				ProductID:       product.ID,
				Name:            u.Name,
				Conversion:      u.Conversion,
				IsDefault:       u.IsDefault,
				Barcode:         u.Barcode,
				AllowSale:       1,
				DefaultSale:     u.DefaultSale,
				DefaultPurchase: u.DefaultPurchase,
				StorageType:     u.StorageType,
				RefPrice:        u.RefPrice,
				WholesalePrice:  u.WholesalePrice,
				RetailPrice:     u.RetailPrice,
				MinSalePrice:    u.MinSalePrice,
				DefaultPrice:    u.DefaultPrice,
				Status:          1,
			}
			if u.AllowSale != nil {
				unit.AllowSale = *u.AllowSale
			}
			if unit.StorageType == 0 {
				unit.StorageType = 1
			}
			if err := tx.Select("ProductID", "Name", "Conversion", "IsDefault", "Barcode", "AllowSale", "DefaultSale", "DefaultPurchase", "StorageType", "RefPrice", "WholesalePrice", "RetailPrice", "MinSalePrice", "DefaultPrice", "Status").Create(&unit).Error; err != nil {
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

// nextProductCode 生成商品编号：SP + yyMMdd + 4位递增序号
func nextProductCode(tx *gorm.DB, companyID uint) (string, error) {
	prefix := "SP" + time.Now().Format("060102")
	var maxCode string
	if err := tx.Model(&model.Product{}).
		Where("company_id = ? AND code LIKE ?", companyID, prefix+"%").
		Select("COALESCE(MAX(code), '')").
		Scan(&maxCode).Error; err != nil {
		return "", err
	}
	seq := 0
	if len(maxCode) == len(prefix)+4 {
		if n, err := strconv.Atoi(maxCode[len(prefix):]); err == nil {
			seq = n
		}
	}
	return fmt.Sprintf("%s%04d", prefix, seq+1), nil
}

// ReplaceProductSpecItems 全量替换商品规格行
func (h *ProductHandler) ReplaceProductSpecItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req struct {
		Items []struct {
			ID        uint    `json:"id"`
			SpecValue string  `json:"specValue" binding:"max=200"`
			Code      string  `json:"code" binding:"max=64"`
			Barcode   string  `json:"barcode" binding:"max=64"`
			Weight    float64 `json:"weight"`
			Volume    float64 `json:"volume"`
			Remark1   string  `json:"remark1" binding:"max=200"`
			Remark2   string  `json:"remark2" binding:"max=200"`
			OnShelf   *int8   `json:"onShelf" binding:"omitempty,oneof=0 1"`
			Sort      int     `json:"sort"`
		} `json:"items"`
	}
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

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ? AND product_id = ?", companyID, product.ID).
			Delete(&model.ProductSpecItem{}).Error; err != nil {
			return err
		}
		for _, item := range req.Items {
			row := model.ProductSpecItem{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				ProductID:            product.ID,
				SpecValue:            item.SpecValue,
				Code:                 item.Code,
				Barcode:              item.Barcode,
				Weight:               item.Weight,
				Volume:               item.Volume,
				Remark1:              item.Remark1,
				Remark2:              item.Remark2,
				OnShelf:              1,
				Sort:                 item.Sort,
			}
			if item.OnShelf != nil {
				row.OnShelf = *item.OnShelf
			}
			if err := tx.Select("CompanyID", "ProductID", "SpecValue", "Code", "Barcode", "Weight", "Volume", "Remark1", "Remark2", "OnShelf", "Sort").Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("replace product spec items failed")
		response.ServerError(c, "保存失败")
		return
	}

	items := make([]model.ProductSpecItem, 0)
	h.db.Where("company_id = ? AND product_id = ?", companyID, product.ID).
		Order("sort ASC, id ASC").Find(&items)
	response.Ok(c, gin.H{"items": items})
}

// CreateOpeningStock 期初库存录入（按公司+仓库+商品累加）
func (h *ProductHandler) CreateOpeningStock(c *gin.Context) {
	var req struct {
		ProductID uint `json:"productId" binding:"required"`
		Items     []struct {
			WarehouseID uint    `json:"warehouseId" binding:"required"`
			Quantity    float64 `json:"quantity"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var product model.Product
	if err := h.db.Where("id = ? AND company_id = ?", req.ProductID, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	processed := 0
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			var count int64
			if err := tx.Model(&model.Warehouse{}).
				Where("id = ? AND company_id = ?", item.WarehouseID, companyID).
				Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("仓库 %d 不存在", item.WarehouseID)
			}
			var stock model.Stock
			err := tx.Where("company_id = ? AND warehouse_id = ? AND product_id = ?",
				companyID, item.WarehouseID, product.ID).First(&stock).Error
			if err == gorm.ErrRecordNotFound {
				stock = model.Stock{
					BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
					WarehouseID:          item.WarehouseID,
					ProductID:            product.ID,
					Quantity:             item.Quantity,
				}
				if err := tx.Create(&stock).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				if err := tx.Model(&stock).Update("quantity", stock.Quantity+item.Quantity).Error; err != nil {
					return err
				}
			}
			processed++
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("create opening stock failed")
		response.ServerError(c, "期初库存保存失败: "+err.Error())
		return
	}
	response.Ok(c, gin.H{"processed": processed})
}
