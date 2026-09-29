package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// SaleScopeHandler 商品销售范围处理器
type SaleScopeHandler struct {
	db *gorm.DB
}

func NewSaleScopeHandler(db *gorm.DB) *SaleScopeHandler {
	return &SaleScopeHandler{db: db}
}

func (h *SaleScopeHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/product-settings/sale-rules")
	{
		g.GET("/dimension", h.GetDimension)
		g.POST("/dimension", h.SetDimension)
		g.GET("/tree", h.GetTree)
		g.GET("/products", h.ListProducts)
		g.PUT("", h.UpsertRule)
	}
}

const (
	settingKeySaleProductDim  = "saleScope.productDim"
	settingKeySaleCustomerDim = "saleScope.customerDim"
)

var validProductDims = map[string]bool{"category": true, "brand": true, "supplier": true}
var validCustomerDims = map[string]bool{"levelPrice": true, "customerCategory": true, "region": true, "customerTag": true}

// SaleRuleContent 规则内容（空数组=不限）
type SaleRuleContent struct {
	LevelPriceIDs       []uint `json:"levelPriceIds"`
	CustomerCategoryIDs []uint `json:"customerCategoryIds"`
	RegionIDs           []uint `json:"regionIds"`
	CustomerTagIDs      []uint `json:"customerTagIds"`
	CustomerIDs         []uint `json:"customerIds"`
}

func (r *SaleRuleContent) isEmpty() bool {
	return len(r.LevelPriceIDs) == 0 && len(r.CustomerCategoryIDs) == 0 &&
		len(r.RegionIDs) == 0 && len(r.CustomerTagIDs) == 0 && len(r.CustomerIDs) == 0
}

// SaleRuleSummary 规则摘要（id 已解析为名称）
type SaleRuleSummary struct {
	LevelPrices        string `json:"levelPrices"`
	CustomerCategories string `json:"customerCategories"`
	Regions            string `json:"regions"`
	CustomerTags       string `json:"customerTags"`
	Customers          string `json:"customers"`
}

// SaleRuleRow 树/列表行
type SaleRuleRow struct {
	TargetType  string           `json:"targetType"`
	TargetID    uint             `json:"targetId"`
	Name        string           `json:"name"`
	HasChildren bool             `json:"hasChildren"`
	Rule        *SaleRuleContent `json:"rule"`
	RuleSummary *SaleRuleSummary `json:"ruleSummary"`
	Children    []*SaleRuleRow   `json:"children,omitempty"`
}

func (h *SaleScopeHandler) companyID(c *gin.Context) uint {
	return middleware.GetCompanyID(c)
}

func (h *SaleScopeHandler) getSetting(companyID uint, key string) string {
	var setting model.CompanySetting
	if err := h.db.Where("company_id = ? AND key = ?", companyID, key).First(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}

func (h *SaleScopeHandler) getDimensions(companyID uint) (string, string) {
	productDim := h.getSetting(companyID, settingKeySaleProductDim)
	if !validProductDims[productDim] {
		productDim = "category"
	}
	customerDim := h.getSetting(companyID, settingKeySaleCustomerDim)
	if !validCustomerDims[customerDim] {
		customerDim = "levelPrice"
	}
	return productDim, customerDim
}

// GetDimension 获取当前设置维度
func (h *SaleScopeHandler) GetDimension(c *gin.Context) {
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	productDim, customerDim := h.getDimensions(companyID)
	response.Ok(c, gin.H{"productDim": productDim, "customerDim": customerDim})
}

// SetDimension 切换设置维度（清空现有规则）
func (h *SaleScopeHandler) SetDimension(c *gin.Context) {
	var req struct {
		ProductDim  string `json:"productDim" binding:"required"`
		CustomerDim string `json:"customerDim" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if !validProductDims[req.ProductDim] {
		response.BadRequest(c, "无效的商品维度")
		return
	}
	if !validCustomerDims[req.CustomerDim] {
		response.BadRequest(c, "无效的客户维度")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range map[string]string{
			settingKeySaleProductDim:  req.ProductDim,
			settingKeySaleCustomerDim: req.CustomerDim,
		} {
			var setting model.CompanySetting
			err := tx.Where("company_id = ? AND key = ?", companyID, key).First(&setting).Error
			if err == nil {
				if err := tx.Model(&setting).Update("value", value).Error; err != nil {
					return err
				}
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			setting = model.CompanySetting{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				Key:                  key,
				Value:                value,
			}
			if err := tx.Create(&setting).Error; err != nil {
				return err
			}
		}
		return tx.Where("company_id = ?", companyID).Delete(&model.ProductSaleRule{}).Error
	})
	if err != nil {
		log.Error().Err(err).Msg("set sale scope dimension failed")
		response.ServerError(c, "保存失败")
		return
	}
	response.OkWithMessage(c, "设置成功", nil)
}

func (h *SaleScopeHandler) loadRules(companyID uint) map[string]*SaleRuleContent {
	var list []model.ProductSaleRule
	h.db.Where("company_id = ?", companyID).Find(&list)
	result := make(map[string]*SaleRuleContent, len(list))
	for _, item := range list {
		var content SaleRuleContent
		if err := json.Unmarshal([]byte(item.Rule), &content); err != nil {
			continue
		}
		result[item.TargetType+":"+strconv.FormatUint(uint64(item.TargetID), 10)] = &content
	}
	return result
}

func summarizeNames(names []string) string {
	if len(names) == 0 {
		return "不限"
	}
	if len(names) > 3 {
		return strings.Join(names[:3], "、") + "…"
	}
	return strings.Join(names, "、")
}

func (h *SaleScopeHandler) buildSummary(companyID uint, rule *SaleRuleContent) *SaleRuleSummary {
	summary := &SaleRuleSummary{
		LevelPrices:        "不限",
		CustomerCategories: "不限",
		Regions:            "不限",
		CustomerTags:       "不限",
		Customers:          "不限",
	}
	if len(rule.LevelPriceIDs) > 0 {
		var items []model.PriceLevel
		h.db.Where("company_id = ? AND id IN ?", companyID, rule.LevelPriceIDs).Find(&items)
		names := make([]string, 0, len(items))
		for _, item := range items {
			names = append(names, item.Name)
		}
		summary.LevelPrices = summarizeNames(names)
	}
	if len(rule.CustomerCategoryIDs) > 0 {
		var items []model.CustomerCategory
		h.db.Where("company_id = ? AND id IN ?", companyID, rule.CustomerCategoryIDs).Find(&items)
		names := make([]string, 0, len(items))
		for _, item := range items {
			names = append(names, item.Name)
		}
		summary.CustomerCategories = summarizeNames(names)
	}
	if len(rule.RegionIDs) > 0 {
		var items []model.Region
		h.db.Where("company_id = ? AND id IN ?", companyID, rule.RegionIDs).Find(&items)
		names := make([]string, 0, len(items))
		for _, item := range items {
			names = append(names, item.Name)
		}
		summary.Regions = summarizeNames(names)
	}
	if len(rule.CustomerTagIDs) > 0 {
		var items []model.CustomerTag
		h.db.Where("company_id = ? AND id IN ?", companyID, rule.CustomerTagIDs).Find(&items)
		names := make([]string, 0, len(items))
		for _, item := range items {
			names = append(names, item.Name)
		}
		summary.CustomerTags = summarizeNames(names)
	}
	if len(rule.CustomerIDs) > 0 {
		var items []model.Customer
		h.db.Where("company_id = ? AND id IN ?", companyID, rule.CustomerIDs).Find(&items)
		names := make([]string, 0, len(items))
		for _, item := range items {
			names = append(names, item.Name)
		}
		summary.Customers = summarizeNames(names)
	}
	return summary
}

func (h *SaleScopeHandler) attachRule(companyID uint, rules map[string]*SaleRuleContent, row *SaleRuleRow) {
	if rule, ok := rules[row.TargetType+":"+strconv.FormatUint(uint64(row.TargetID), 10)]; ok {
		row.Rule = rule
		row.RuleSummary = h.buildSummary(companyID, rule)
	}
}

// GetTree 按当前商品维度返回行
func (h *SaleScopeHandler) GetTree(c *gin.Context) {
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	productDim, _ := h.getDimensions(companyID)
	rules := h.loadRules(companyID)
	rows := make([]*SaleRuleRow, 0)

	switch productDim {
	case "brand":
		var brands []model.Brand
		query := h.db.Where("company_id = ?", companyID)
		if keyword != "" {
			query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		query.Order("sort ASC, created_at DESC").Find(&brands)
		for _, b := range brands {
			row := &SaleRuleRow{TargetType: "brand", TargetID: b.ID, Name: b.Name}
			h.attachRule(companyID, rules, row)
			rows = append(rows, row)
		}
	case "supplier":
		var suppliers []model.Supplier
		query := h.db.Where("company_id = ?", companyID)
		if keyword != "" {
			query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		query.Order("created_at DESC").Find(&suppliers)
		for _, s := range suppliers {
			row := &SaleRuleRow{TargetType: "supplier", TargetID: s.ID, Name: s.Name}
			h.attachRule(companyID, rules, row)
			rows = append(rows, row)
		}
	default:
		parentID, _ := strconv.ParseUint(c.Query("parentId"), 10, 64)
		var cats []model.ProductCategory
		h.db.Where("company_id = ?", companyID).Order("sort ASC, created_at DESC").Find(&cats)
		type catCount struct {
			CategoryID uint
			Cnt        int64
		}
		var counts []catCount
		h.db.Model(&model.Product{}).
			Select("category_id, COUNT(*) AS cnt").
			Where("company_id = ?", companyID).
			Group("category_id").Scan(&counts)
		countMap := make(map[uint]int64, len(counts))
		for _, item := range counts {
			countMap[item.CategoryID] = item.Cnt
		}
		childCount := make(map[uint]int64, len(cats))
		for _, cat := range cats {
			childCount[cat.ParentID]++
		}
		if keyword != "" {
			for _, cat := range cats {
				if strings.Contains(cat.Name, keyword) || strings.Contains(cat.Code, keyword) {
					row := &SaleRuleRow{
						TargetType:  "category",
						TargetID:    cat.ID,
						Name:        cat.Name,
						HasChildren: countMap[cat.ID] > 0 || childCount[cat.ID] > 0,
					}
					h.attachRule(companyID, rules, row)
					rows = append(rows, row)
				}
			}
			break
		}
		for _, cat := range cats {
			if cat.ParentID != uint(parentID) {
				continue
			}
			row := &SaleRuleRow{
				TargetType:  "category",
				TargetID:    cat.ID,
				Name:        cat.Name,
				HasChildren: countMap[cat.ID] > 0 || childCount[cat.ID] > 0,
			}
			h.attachRule(companyID, rules, row)
			rows = append(rows, row)
		}
	}
	if rows == nil {
		rows = []*SaleRuleRow{}
	}
	response.Ok(c, rows)
}

// ListProducts 商品行（分类懒加载 / 关键字平铺）
func (h *SaleScopeHandler) ListProducts(c *gin.Context) {
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	categoryID, _ := strconv.ParseUint(c.Query("categoryId"), 10, 64)
	keyword := strings.TrimSpace(c.Query("keyword"))

	query := h.db.Model(&model.Product{}).Where("company_id = ?", companyID)
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var products []model.Product
	query.Order("created_at DESC").Limit(500).Find(&products)

	rules := h.loadRules(companyID)
	rows := make([]*SaleRuleRow, 0, len(products))
	for _, p := range products {
		row := &SaleRuleRow{TargetType: "product", TargetID: p.ID, Name: p.Name}
		h.attachRule(companyID, rules, row)
		rows = append(rows, row)
	}
	response.Ok(c, rows)
}

// UpsertRule 保存/清除某目标行的规则
func (h *SaleScopeHandler) UpsertRule(c *gin.Context) {
	var req struct {
		TargetType string          `json:"targetType" binding:"required"`
		TargetID   uint            `json:"targetId" binding:"required"`
		Rule       SaleRuleContent `json:"rule"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	if !map[string]bool{"category": true, "brand": true, "supplier": true, "product": true}[req.TargetType] {
		response.BadRequest(c, "无效的目标类型")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	query := h.db.Where("company_id = ? AND target_type = ? AND target_id = ?", companyID, req.TargetType, req.TargetID)
	if req.Rule.isEmpty() {
		if err := query.Delete(&model.ProductSaleRule{}).Error; err != nil {
			log.Error().Err(err).Msg("delete sale rule failed")
			response.ServerError(c, "保存失败")
			return
		}
		response.OkWithMessage(c, "已恢复不限", nil)
		return
	}
	raw, err := json.Marshal(req.Rule)
	if err != nil {
		response.BadRequest(c, "规则格式错误")
		return
	}
	var item model.ProductSaleRule
	err = query.First(&item).Error
	if err == nil {
		item.Rule = string(raw)
		if err := h.db.Save(&item).Error; err != nil {
			log.Error().Err(err).Msg("update sale rule failed")
			response.ServerError(c, "保存失败")
			return
		}
	} else if err == gorm.ErrRecordNotFound {
		item = model.ProductSaleRule{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			TargetType:           req.TargetType,
			TargetID:             req.TargetID,
			Rule:                 string(raw),
		}
		if err := h.db.Create(&item).Error; err != nil {
			log.Error().Err(err).Msg("create sale rule failed")
			response.ServerError(c, "保存失败")
			return
		}
	} else {
		log.Error().Err(err).Msg("query sale rule failed")
		response.ServerError(c, "保存失败")
		return
	}
	response.Ok(c, gin.H{"rule": req.Rule, "ruleSummary": h.buildSummary(companyID, &req.Rule)})
}
