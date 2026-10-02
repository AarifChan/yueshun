package model

import "time"

// SupplierConnection 供应商连接（上游企业）
type SupplierConnection struct {
	BaseModelWithCompany
	CorpName   string `json:"corpName" gorm:"size:128;not null;comment:上游企业微信名称"`
	CorpCode   string `json:"corpCode" gorm:"size:64;comment:企业编号"`
	SupplierID uint   `json:"supplierId" gorm:"index;default:0;comment:关联供应商"`
	Status     string `json:"status" gorm:"size:20;default:pending;comment:pending待连接 connected已连接 disconnected已断开"`
	Remark     string `json:"remark" gorm:"size:500"`
}

func (SupplierConnection) TableName() string { return "supplier_connections" }

// ReceivedGoods 商品接收（上游推送的商品）
type ReceivedGoods struct {
	BaseModelWithCompany
	ConnectionID  uint    `json:"connectionId" gorm:"index;default:0"`
	SupplierID    uint    `json:"supplierId" gorm:"index;default:0"`
	Code          string  `json:"code" gorm:"size:64"`
	Barcode       string  `json:"barcode" gorm:"size:64"`
	Name          string  `json:"name" gorm:"size:128;not null"`
	Spec          string  `json:"spec" gorm:"size:128"`
	Brand         string  `json:"brand" gorm:"size:64"`
	Image         string  `json:"image" gorm:"size:255"`
	Unit          string  `json:"unit" gorm:"size:32;default:个"`
	PurchasePrice float64 `json:"purchasePrice" gorm:"type:decimal(18,4);default:0"`
	Status        string  `json:"status" gorm:"size:20;default:pending;comment:pending自动接收(待接收) received已接收 closed已关闭"`
	ProductID     uint    `json:"productId" gorm:"index;default:0;comment:接收后生成的商品"`
}

func (ReceivedGoods) TableName() string { return "received_goods" }

// ReceivedDecoration 商城装修接收
type ReceivedDecoration struct {
	BaseModelWithCompany
	Title      string     `json:"title" gorm:"size:128;not null"`
	SourceName string     `json:"sourceName" gorm:"size:128;comment:来源(上游企业/平台)"`
	Content    string     `json:"content" gorm:"type:text;comment:装修配置JSON"`
	Status     string     `json:"status" gorm:"size:20;default:pending;comment:pending自动接收(待接收) received已接收 closed已关闭"`
	ReceivedAt *time.Time `json:"receivedAt"`
}

func (ReceivedDecoration) TableName() string { return "received_decorations" }
