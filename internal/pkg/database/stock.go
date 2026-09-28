package database

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"zhizhang-server/internal/model"
)

// ErrStockNotEnough 库存不足
var ErrStockNotEnough = errors.New("库存不足")

// ChangeStock 在事务内增减库存台账（company + 仓库 + 商品 维度）。
// delta 为正表示入库，为负表示出库；结果数量为负时返回 ErrStockNotEnough。
func ChangeStock(tx *gorm.DB, companyID, warehouseID, productID uint, delta float64) error {
	var stock model.Stock
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("company_id = ? AND warehouse_id = ? AND product_id = ?", companyID, warehouseID, productID).
		First(&stock).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if delta < 0 {
			return ErrStockNotEnough
		}
		stock = model.Stock{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			WarehouseID:          warehouseID,
			ProductID:            productID,
			Quantity:             delta,
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "company_id"}, {Name: "warehouse_id"}, {Name: "product_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"quantity": gorm.Expr("stocks.quantity + ?", delta),
			}),
		}).Create(&stock).Error
	}
	if err != nil {
		return err
	}

	newQty := stock.Quantity + delta
	if newQty < 0 {
		return ErrStockNotEnough
	}
	return tx.Model(&model.Stock{}).
		Where("id = ?", stock.ID).
		UpdateColumn("quantity", newQty).Error
}
