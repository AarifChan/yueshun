package handler

import (
	_ "embed"
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
)

//go:embed seeddata/inventory_seed.csv
var inventorySeedCSV string

const inventorySeedMarkerKey = "inventory_seed_v1"

// SeedInventoryStatusData 库存状况表数据种子（幂等，按公司只做一次）：
// 以导出表格（库存状况列表）为准，按商品名称匹配已有商品并更新
// 规格/单位/成本价/库存上下限/排序权重，同名多规格时创建缺失的商品行，
// 最后把库存数量写入默认仓库的库存台账。
func SeedInventoryStatusData(db *gorm.DB) error {
	var company model.Company
	if err := db.First(&company).Error; err != nil {
		return err
	}

	var marker model.CompanySetting
	err := db.Where("company_id = ? AND key = ?", company.ID, inventorySeedMarkerKey).First(&marker).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	records, err := csv.NewReader(strings.NewReader(inventorySeedCSV)).ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 新建商品用的默认分类
		var category model.ProductCategory
		if err := tx.Where("company_id = ? AND name = ?", company.ID, "默认分类").First(&category).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			category = model.ProductCategory{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
				Name:                 "默认分类",
				Code:                 "DEFAULT",
				Status:               1,
			}
			if err := tx.Create(&category).Error; err != nil {
				return err
			}
		}

		// 库存写入的默认仓库：优先复用已有仓库
		var warehouse model.Warehouse
		if err := tx.Where("company_id = ?", company.ID).Order("id").First(&warehouse).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			warehouse = model.Warehouse{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
				Name:                 "默认仓库",
				Code:                 "DEFAULT",
				Status:               1,
			}
			if err := tx.Create(&warehouse).Error; err != nil {
				return err
			}
		}

		type seedRow struct {
			name, spec, unit           string
			qty, cost, maxStock, min   float64
			sortWeight                 int
		}
		parseFloat := func(s string) float64 {
			f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
			return f
		}

		// 按名称分组（保持表格顺序，同名多规格为多行变体）
		order := []string{}
		byName := map[string][]seedRow{}
		for _, rec := range records[1:] {
			if len(rec) < 9 || strings.TrimSpace(rec[0]) == "" {
				continue
			}
			unit := strings.TrimSpace(rec[3])
			if unit == "" {
				unit = "个"
			}
			sortWeight, _ := strconv.Atoi(strings.TrimSpace(rec[8]))
			row := seedRow{
				name: strings.TrimSpace(rec[0]), spec: strings.TrimSpace(rec[1]), unit: unit,
				qty: parseFloat(rec[2]), cost: parseFloat(rec[4]),
				maxStock: parseFloat(rec[5]), min: parseFloat(rec[6]),
				sortWeight: sortWeight,
			}
			if _, ok := byName[row.name]; !ok {
				order = append(order, row.name)
			}
			byName[row.name] = append(byName[row.name], row)
		}

		updated, created, stocked := 0, 0, 0
		for _, name := range order {
			rows := byName[name]
			var existing []model.Product
			if err := tx.Where("company_id = ? AND name = ?", company.ID, name).Order("id").Find(&existing).Error; err != nil {
				return err
			}
			for i, row := range rows {
				var p model.Product
				if i < len(existing) {
					p = existing[i]
					if err := tx.Model(&model.Product{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
						"specification":    row.spec,
						"unit":             row.unit,
						"purchase_price":   row.cost,
						"min_stock":        row.min,
						"max_stock":        row.maxStock,
						"mall_sort_weight": row.sortWeight,
					}).Error; err != nil {
						return err
					}
					updated++
				} else {
					p = model.Product{
						BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
						CategoryID:           category.ID,
						Name:                 row.name,
						Specification:        row.spec,
						Unit:                 row.unit,
						PurchasePrice:        row.cost,
						MinStock:             row.min,
						MaxStock:             row.maxStock,
						Status:               1,
						MallSortWeight:       row.sortWeight,
					}
					if err := tx.Create(&p).Error; err != nil {
						return err
					}
					created++
				}

				res := tx.Model(&model.Stock{}).
					Where("company_id = ? AND warehouse_id = ? AND product_id = ?", company.ID, warehouse.ID, p.ID).
					Update("quantity", row.qty)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 && row.qty != 0 {
					if err := tx.Create(&model.Stock{
						BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
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

		if err := tx.Create(&model.CompanySetting{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
			Key:                  inventorySeedMarkerKey,
			Value:                "done",
		}).Error; err != nil {
			return err
		}

		log.Info().Int("updated", updated).Int("created", created).Int("stocks", stocked).Msg("inventory seed data imported")
		return nil
	})
}
