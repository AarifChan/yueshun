package model

// CustomerProductPrice 客户单独定价
type CustomerProductPrice struct {
	BaseModelWithCompany
	CustomerID uint    `json:"customerId" gorm:"index;not null;uniqueIndex:uniq_customer_product_price,priority:2"`
	ProductID  uint    `json:"productId" gorm:"index;not null;uniqueIndex:uniq_customer_product_price,priority:3"`
	Price      float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Remark     string  `json:"remark" gorm:"size:255"`
}

func (CustomerProductPrice) TableName() string { return "customer_product_prices" }
