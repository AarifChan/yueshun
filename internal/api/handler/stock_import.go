package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// errWarehouseNotFound 导入时指定的目标仓库不存在
var errWarehouseNotFound = errors.New("warehouse not found")

// cleanHeaderCell 表头单元格归一化：去换行/回车后 TrimSpace
func cleanHeaderCell(s string) string {
	return strings.TrimSpace(strings.NewReplacer("\n", "", "\r", "").Replace(s))
}

// ImportStockStatus 库存状况 Excel 导入
// @Summary 库存状况 Excel 导入
// @Description 导入老系统库存状况导出文件（.xls/.xlsx），按名称匹配/新建商品并把库存总量写入目标仓库
// @Tags 库存
// @Accept multipart/form-data
// @Param file formData file true "库存状况导出文件（.xls/.xlsx，≤10MB）"
// @Param warehouseId formData int false "目标仓库ID，为空取公司第一个仓库"
// @Success 200 {object} response.Response
// @Router /stocks/import [post]
func (h *WarehouseHandler) ImportStockStatus(c *gin.Context) {
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
	src, err := file.Open()
	if err != nil {
		response.ServerError(c, "读取文件失败")
		return
	}
	defer src.Close()

	rows, err := openSpreadsheetRows(src, file.Filename)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 表头/数据单元格统一去换行
	for i := range rows {
		for j := range rows[i] {
			rows[i][j] = cleanHeaderCell(rows[i][j])
		}
	}

	headerIdx := findHeaderRow(rows, 5, "商品名称", "规格")
	if headerIdx < 0 {
		response.BadRequest(c, "未找到表头行（需包含“商品名称”或“规格”列）")
		return
	}

	// 按表头名定位列，找不到时用老系统导出文件的固定列序兜底
	colDefaults := map[string]int{
		"商品名称": 0, "规格": 1, "库存总量": 2, "单位": 3, "成本均价": 4,
		"库存上限": 9, "库存下限": 10, "上下架状态": 11, "排序权重": 12,
	}
	colIdx := map[string]int{}
	for name, def := range colDefaults {
		colIdx[name] = def
	}
	for j, cell := range rows[headerIdx] {
		if _, ok := colDefaults[cell]; ok {
			colIdx[cell] = j
		}
	}

	var warehouseID uint
	if v := strings.TrimSpace(c.PostForm("warehouseId")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil || id == 0 {
			response.BadRequest(c, "warehouseId 参数错误")
			return
		}
		warehouseID = uint(id)
	}

	type importRow struct {
		name, spec, unit     string
		qty, cost            float64
		maxStock, minStock   float64
		status               int8
		sortWeight           int
	}
	parseFloat := func(s string) float64 {
		f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
		return f
	}
	getCell := func(raw []string, col string) string {
		idx := colIdx[col]
		if idx < len(raw) {
			return strings.TrimSpace(raw[idx])
		}
		return ""
	}

	// 按名称分组（保持表格顺序，同名多规格为多行变体）
	order := []string{}
	byName := map[string][]importRow{}
	skipped := 0
	for i := headerIdx + 1; i < len(rows); i++ {
		raw := rows[i]
		name := getCell(raw, "商品名称")
		if name == "" {
			if strings.Join(raw, "") != "" {
				skipped++
			}
			continue
		}
		unit := getCell(raw, "单位")
		if unit == "" {
			unit = "个"
		}
		status := int8(1)
		if strings.Contains(getCell(raw, "上下架状态"), "下架") {
			status = 0
		}
		sortWeight, _ := strconv.Atoi(getCell(raw, "排序权重"))
		row := importRow{
			name: name, spec: getCell(raw, "规格"), unit: unit,
			qty: parseFloat(getCell(raw, "库存总量")), cost: parseFloat(getCell(raw, "成本均价")),
			maxStock: parseFloat(getCell(raw, "库存上限")), minStock: parseFloat(getCell(raw, "库存下限")),
			status: status, sortWeight: sortWeight,
		}
		if _, ok := byName[name]; !ok {
			order = append(order, name)
		}
		byName[name] = append(byName[name], row)
	}
	if len(order) == 0 {
		response.BadRequest(c, "文件中没有可导入的数据行")
		return
	}

	warehouseName := ""
	updated, created, stocked := 0, 0, 0

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 目标仓库：指定 → 校验归属；未指定 → 公司第一个仓库，没有则建“默认仓库”
		var warehouse model.Warehouse
		if warehouseID > 0 {
			if err := tx.Where("company_id = ? AND id = ?", companyID, warehouseID).First(&warehouse).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return errWarehouseNotFound
				}
				return err
			}
		} else {
			if err := tx.Where("company_id = ?", companyID).Order("id").First(&warehouse).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
				warehouse = model.Warehouse{
					BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
					Name:                 "默认仓库",
					Code:                 "DEFAULT",
					Status:               1,
				}
				if err := tx.Create(&warehouse).Error; err != nil {
					return err
				}
			}
		}
		warehouseName = warehouse.Name

		// 新建商品用的默认分类
		var category model.ProductCategory
		if err := tx.Where("company_id = ? AND name = ?", companyID, "默认分类").First(&category).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			category = model.ProductCategory{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				Name:                 "默认分类",
				Code:                 "DEFAULT",
				Status:               1,
			}
			if err := tx.Create(&category).Error; err != nil {
				return err
			}
		}

		for _, name := range order {
			nameRows := byName[name]
			var existing []model.Product
			if err := tx.Where("company_id = ? AND name = ?", companyID, name).Order("id").Find(&existing).Error; err != nil {
				return err
			}
			for i, row := range nameRows {
				var p model.Product
				if i < len(existing) {
					p = existing[i]
					if err := tx.Model(&model.Product{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
						"specification":    row.spec,
						"unit":             row.unit,
						"purchase_price":   row.cost,
						"min_stock":        row.minStock,
						"max_stock":        row.maxStock,
						"status":           row.status,
						"mall_sort_weight": row.sortWeight,
					}).Error; err != nil {
						return err
					}
					updated++
				} else {
					code, err := nextProductCode(tx, companyID)
					if err != nil {
						return err
					}
					p = model.Product{
						BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
						CategoryID:           category.ID,
						Name:                 row.name,
						Code:                 code,
						Specification:        row.spec,
						Unit:                 row.unit,
						PurchasePrice:        row.cost,
						MinStock:             row.minStock,
						MaxStock:             row.maxStock,
						Status:               row.status,
						MallSortWeight:       row.sortWeight,
					}
					if err := tx.Create(&p).Error; err != nil {
						return err
					}
					created++
				}

				res := tx.Model(&model.Stock{}).
					Where("company_id = ? AND warehouse_id = ? AND product_id = ?", companyID, warehouse.ID, p.ID).
					Update("quantity", row.qty)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 && row.qty != 0 {
					if err := tx.Create(&model.Stock{
						BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
						WarehouseID:          warehouse.ID,
						ProductID:            p.ID,
						Quantity:             row.qty,
					}).Error; err != nil {
						return err
					}
				}
				stocked++
			}
		}
		return nil
	})
	if err != nil {
		if err == errWarehouseNotFound {
			response.BadRequest(c, "目标仓库不存在")
			return
		}
		log.Error().Err(err).Msg("stock status import failed")
		response.ServerError(c, "导入失败")
		return
	}

	response.Ok(c, gin.H{
		"warehouse": warehouseName,
		"updated":   updated,
		"created":   created,
		"stocked":   stocked,
		"skipped":   skipped,
	})
}
