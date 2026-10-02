package model

// ==================== 资料模块（10-资料） =====================

// StockOpType 出入库类型
type StockOpType struct {
	BaseModelWithCompany
	Name   string `json:"name" gorm:"size:64;not null"`
	Code   string `json:"code" gorm:"size:64;index"`
	Type   string `json:"type" gorm:"size:10;not null;comment:in入库 out出库"`
	Sort   int    `json:"sort" gorm:"default:0"`
	Status int8   `json:"status" gorm:"default:1"`
}

func (StockOpType) TableName() string { return "stock_op_types" }

// DeliveryMethod 发货方式
type DeliveryMethod struct {
	BaseModelWithCompany
	Name         string `json:"name" gorm:"size:64;not null"`
	NeedLogistic bool   `json:"needLogistic" gorm:"default:false;comment:是否需要物流"`
	Sort         int    `json:"sort" gorm:"default:0"`
	Status       int8   `json:"status" gorm:"default:1"`
}

func (DeliveryMethod) TableName() string { return "delivery_methods" }

// LogisticsCompany 物流公司
type LogisticsCompany struct {
	BaseModelWithCompany
	Name      string `json:"name" gorm:"size:64;not null"`
	Code      string `json:"code" gorm:"size:64;index"`
	Tracking  bool   `json:"tracking" gorm:"default:false;comment:支持物流跟踪"`
	IsCustom  bool   `json:"isCustom" gorm:"default:false;comment:其他快递/物流"`
	Sort      int    `json:"sort" gorm:"default:0"`
	Status    int8   `json:"status" gorm:"default:1"`
}

func (LogisticsCompany) TableName() string { return "logistics_companies" }

// Material 素材库
type Material struct {
	BaseModelWithCompany
	Name string `json:"name" gorm:"size:128;not null"`
	Type string `json:"type" gorm:"size:20;not null;comment:image图片 video视频 file附件"`
	URL  string `json:"url" gorm:"size:512;not null"`
	Size int64  `json:"size" gorm:"default:0"`
	Sort int    `json:"sort" gorm:"default:0"`
}

func (Material) TableName() string { return "materials" }

// InitialStock 商品库存期初
type InitialStock struct {
	BaseModelWithCompany
	ProductID   uint    `json:"productId" gorm:"index;not null;uniqueIndex:uniq_init_stock,priority:2"`
	WarehouseID uint    `json:"warehouseId" gorm:"index;not null;uniqueIndex:uniq_init_stock,priority:3"`
	Quantity    float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	CostPrice   float64 `json:"costPrice" gorm:"type:decimal(18,4);default:0"`
	CostAmount  float64 `json:"costAmount" gorm:"type:decimal(18,4);default:0"`
}

func (InitialStock) TableName() string { return "initial_stocks" }

// InitialBalance 往来期初（应收预收/应付预付）
type InitialBalance struct {
	BaseModelWithCompany
	BizType    string  `json:"bizType" gorm:"size:10;not null;uniqueIndex:uniq_init_balance,priority:2;comment:customer客户 supplier供应商"`
	TargetID   uint    `json:"targetId" gorm:"index;not null;uniqueIndex:uniq_init_balance,priority:3"`
	Receivable float64 `json:"receivable" gorm:"type:decimal(18,4);default:0;comment:应收/应付期初"`
	Advance    float64 `json:"advance" gorm:"type:decimal(18,4);default:0;comment:预收/预付期初"`
}

func (InitialBalance) TableName() string { return "initial_balances" }

// InitialAccountBalance 现金银行期初
type InitialAccountBalance struct {
	BaseModelWithCompany
	AccountID uint    `json:"accountId" gorm:"index;not null;uniqueIndex:uniq_init_account,priority:2"`
	Amount    float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
}

func (InitialAccountBalance) TableName() string { return "initial_account_balances" }
