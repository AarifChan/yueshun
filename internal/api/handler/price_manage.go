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

// PriceManageHandler 商品物价管理处理器
type PriceManageHandler struct {
	db *gorm.DB
}

func NewPriceManageHandler(db *gorm.DB) *PriceManageHandler {
	return &PriceManageHandler{db: db}
}

func (h *PriceManageHandler) RegisterRoutes(r *gin.RouterGroup) {
	price := r.Group("/prices/manage")
	{
		price.GET("", h.ListPriceManage)
		price.PUT("/:id", h.UpdatePriceManage)
		price.POST("/import", h.ImportPriceManage)
	}
}

type PriceManageListReq struct {
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=30"`
	Keyword     string `form:"keyword"`
	CategoryID  uint   `form:"categoryId"`
	Status      *int8  `form:"status"`
	StockStatus string `form:"stockStatus"` // in有库存 out无库存
	PriceType   string `form:"priceType"`   // default/wholesale/retail
	PriceOp     string `form:"priceOp"`     // gt/lt
	PriceValue  string `form:"priceValue"`
}

type PriceManageItem struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Code           string  `json:"code"`
	Unit           string  `json:"unit"`
	DefaultPrice   float64 `json:"defaultPrice"`
	WholesalePrice float64 `json:"wholesalePrice"`
	RetailPrice    float64 `json:"retailPrice"`
	TotalStock     float64 `json:"totalStock"`
	Status         int8    `json:"status"`
}

func (h *PriceManageHandler) defaultLevelID(companyID uint) uint {
	var level model.PriceLevel
	if err := h.db.Where("company_id = ? AND is_default = ?", companyID, true).First(&level).Error; err != nil {
		return 0
	}
	return level.ID
}

func (h *PriceManageHandler) ListPriceManage(c *gin.Context) {
	var req PriceManageListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 200 {
		req.PageSize = 30
	}

	levelID := h.defaultLevelID(companyID)
	defaultPriceSQL := "COALESCE((SELECT product_prices.price FROM product_prices WHERE product_prices.product_id = products.id AND product_prices.unit_id = 0 AND product_prices.level_id = ? AND product_prices.customer_id = 0 LIMIT 1), 0)"
	totalStockSQL := "COALESCE((SELECT SUM(stocks.quantity) FROM stocks WHERE stocks.product_id = products.id AND stocks.company_id = ?), 0)"

	query := h.db.Model(&model.Product{}).Where("products.company_id = ?", companyID)
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("products.name LIKE ? OR products.code LIKE ? OR products.barcode LIKE ?", like, like, like)
	}
	if req.CategoryID > 0 {
		query = query.Where("products.category_id = ?", req.CategoryID)
	}
	if req.Status != nil {
		query = query.Where("products.status = ?", *req.Status)
	}
	if req.StockStatus == "in" {
		query = query.Where(totalStockSQL+" > 0", companyID)
	} else if req.StockStatus == "out" {
		query = query.Where(totalStockSQL+" <= 0", companyID)
	}
	if req.PriceValue != "" {
		if priceValue, err := strconv.ParseFloat(req.PriceValue, 64); err == nil {
			op := ">"
			if req.PriceOp == "lt" {
				op = "<"
			}
			switch req.PriceType {
			case "wholesale":
				query = query.Where("products.wholesale_price "+op+" ?", priceValue)
			case "retail":
				query = query.Where("products.retail_price "+op+" ?", priceValue)
			default:
				query = query.Where(defaultPriceSQL+" "+op+" ?", levelID, priceValue)
			}
		}
	}

	var total int64
	query.Count(&total)

	var list []PriceManageItem
	query.Select("products.id, products.name, products.code, products.unit, products.wholesale_price, products.retail_price, products.status, "+defaultPriceSQL+" AS default_price, "+totalStockSQL+" AS total_stock", levelID, companyID).
		Order("products.id ASC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Scan(&list)

	response.OkWithPage(c, list, req.Page, req.PageSize, int(total))
}

type PriceManageUpdateReq struct {
	WholesalePrice *float64 `json:"wholesalePrice"`
	RetailPrice    *float64 `json:"retailPrice"`
	DefaultPrice   *float64 `json:"defaultPrice"`
}

func (h *PriceManageHandler) applyPriceUpdate(tx *gorm.DB, companyID, productID uint, req *PriceManageUpdateReq) error {
	updates := map[string]interface{}{}
	if req.WholesalePrice != nil {
		updates["wholesale_price"] = *req.WholesalePrice
	}
	if req.RetailPrice != nil {
		updates["retail_price"] = *req.RetailPrice
	}
	if len(updates) > 0 {
		if err := tx.Model(&model.Product{}).Where("id = ? AND company_id = ?", productID, companyID).Updates(updates).Error; err != nil {
			return err
		}
	}
	if req.DefaultPrice != nil {
		var levelID uint
		if err := tx.Model(&model.PriceLevel{}).Where("company_id = ? AND is_default = ?", companyID, true).Select("id").Scan(&levelID).Error; err != nil {
			return err
		}
		if levelID > 0 {
			var pp model.ProductPrice
			err := tx.Where("product_id = ? AND unit_id = 0 AND level_id = ? AND customer_id = 0", productID, levelID).First(&pp).Error
			if err == nil {
				if err := tx.Model(&pp).Update("price", *req.DefaultPrice).Error; err != nil {
					return err
				}
			} else if err == gorm.ErrRecordNotFound {
				pp = model.ProductPrice{
					ProductID: productID,
					UnitID:    0,
					LevelID:   levelID,
					Price:     *req.DefaultPrice,
					Status:    1,
				}
				if err := tx.Create(&pp).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func (h *PriceManageHandler) UpdatePriceManage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 格式错误")
		return
	}
	var req PriceManageUpdateReq
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
	if err := h.db.Where("id = ? AND company_id = ?", id, companyID).First(&product).Error; err != nil {
		response.NotFound(c, "商品不存在")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		return h.applyPriceUpdate(tx, companyID, product.ID, &req)
	}); err != nil {
		log.Error().Err(err).Msg("update price manage failed")
		response.ServerError(c, "更新失败")
		return
	}
	response.OkWithMessage(c, "价格更新成功", nil)
}

type PriceImportError struct {
	Row    int    `json:"row"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

func parsePriceCell(cell string) (*float64, error) {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(cell, 64)
	if err != nil || v < 0 {
		return nil, fmt.Errorf("价格格式错误")
	}
	return &v, nil
}

func (h *PriceManageHandler) ImportPriceManage(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}
	if file.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小不能超过 10MB")
		return
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".xlsx" {
		response.BadRequest(c, "仅支持 .xlsx 文件")
		return
	}
	src, err := file.Open()
	if err != nil {
		response.ServerError(c, "读取文件失败")
		return
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		response.BadRequest(c, "Excel 文件解析失败")
		return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil || len(rows) < 2 {
		response.BadRequest(c, "文件内容为空")
		return
	}

	var levelID uint
	h.db.Model(&model.PriceLevel{}).Where("company_id = ? AND is_default = ?", companyID, true).Select("id").Scan(&levelID)

	updated := 0
	failed := 0
	errors := make([]PriceImportError, 0, 20)

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		rowNo := i + 1
		getCell := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		code := getCell(0)
		if code == "" {
			continue
		}
		fail := func(reason string) {
			failed++
			if len(errors) < 20 {
				errors = append(errors, PriceImportError{Row: rowNo, Code: code, Reason: reason})
			}
		}

		wholesale, err := parsePriceCell(getCell(1))
		if err != nil {
			fail("批发价" + err.Error())
			continue
		}
		retail, err := parsePriceCell(getCell(2))
		if err != nil {
			fail("零售价" + err.Error())
			continue
		}
		defaultPrice, err := parsePriceCell(getCell(3))
		if err != nil {
			fail("默认订货价" + err.Error())
			continue
		}

		var product model.Product
		if err := h.db.Where("code = ? AND company_id = ?", code, companyID).First(&product).Error; err != nil {
			fail("商品编号不存在")
			continue
		}

		req := PriceManageUpdateReq{WholesalePrice: wholesale, RetailPrice: retail, DefaultPrice: defaultPrice}
		if err := h.db.Transaction(func(tx *gorm.DB) error {
			return h.applyPriceUpdate(tx, companyID, product.ID, &req)
		}); err != nil {
			fail("更新失败")
			continue
		}
		updated++
	}

	response.Ok(c, gin.H{
		"updated": updated,
		"failed":  failed,
		"errors":  errors,
	})
}
