package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型（所有实体嵌入）
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"createdAt" gorm:"index"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BaseModelWithCompany 带公司隔离的基础模型
type BaseModelWithCompany struct {
	BaseModel
	CompanyID uint `json:"companyId" gorm:"index;not null"`
}

// Employee 职员（系统用户）
type Employee struct {
	BaseModelWithCompany
	DeptID      uint       `json:"deptId" gorm:"index;not null"`
	RoleID      uint       `json:"roleId" gorm:"index;not null"`
	Username    string     `json:"username" gorm:"size:64;uniqueIndex:idx_emp_username_company"`
	Password    string     `json:"-" gorm:"size:128;not null"`
	Name        string     `json:"name" gorm:"size:64;not null"`
	Phone       string     `json:"phone" gorm:"size:20;index"`
	Email       string     `json:"email" gorm:"size:128"`
	Status      int8       `json:"status" gorm:"default:1;comment:1启用 0禁用"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	LastLoginIP string     `json:"lastLoginIp" gorm:"size:64"`
	WeComUserID string     `json:"wecomUserId" gorm:"size:64;index;comment:企业微信UserID"`
}

// Role 角色
type Role struct {
	BaseModelWithCompany
	Name        string `json:"name" gorm:"size:64;not null"`
	Code        string `json:"code" gorm:"size:64;index"`
	Description string `json:"description" gorm:"size:255"`
	Status      int8   `json:"status" gorm:"default:1"`
}

// Permission 权限
type Permission struct {
	BaseModel
	Name     string `json:"name" gorm:"size:64;not null"`
	Code     string `json:"code" gorm:"size:128;uniqueIndex"`
	Type     string `json:"type" gorm:"size:20;comment:menu|button|api|data"`
	ParentID uint   `json:"parentId" gorm:"default:0"`
	Path     string `json:"path" gorm:"size:255"`
	Sort     int    `json:"sort" gorm:"default:0"`
	Status   int8   `json:"status" gorm:"default:1"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

// Department 部门
type Department struct {
	BaseModelWithCompany
	ParentID  uint   `json:"parentId" gorm:"index;default:0"`
	Name      string `json:"name" gorm:"size:64;not null"`
	Code      string `json:"code" gorm:"size:64"`
	ManagerID uint   `json:"managerId"`
	Sort      int    `json:"sort" gorm:"default:0"`
	Status    int8   `json:"status" gorm:"default:1"`
}

// Company 公司/租户
type Company struct {
	BaseModel
	Name    string `json:"name" gorm:"size:128;not null"`
	Code    string `json:"code" gorm:"size:64;uniqueIndex"`
	Contact string `json:"contact" gorm:"size:64"`
	Phone   string `json:"phone" gorm:"size:20"`
	Address string `json:"address" gorm:"size:255"`
	Status  int8   `json:"status" gorm:"default:1"`
}

// DictType 字典类型
type DictType struct {
	BaseModelWithCompany
	Name   string `json:"name" gorm:"size:64;not null"`
	Code   string `json:"code" gorm:"size:64;index"`
	Status int8   `json:"status" gorm:"default:1"`
}

// DictItem 字典项
type DictItem struct {
	BaseModel
	DictTypeID uint   `json:"dictTypeId" gorm:"index;not null"`
	Label      string `json:"label" gorm:"size:64;not null"`
	Value      string `json:"value" gorm:"size:64;not null"`
	Sort       int    `json:"sort" gorm:"default:0"`
	Status     int8   `json:"status" gorm:"default:1"`
}

// ==================== 商品资料模块 ====================

// ProductCategory 商品分类（树形）
type ProductCategory struct {
	BaseModelWithCompany
	ParentID uint   `json:"parentId" gorm:"index;default:0"`
	Name     string `json:"name" gorm:"size:64;not null"`
	Code     string `json:"code" gorm:"size:64"`
	Image    string `json:"image" gorm:"size:255"`
	Sort     int    `json:"sort" gorm:"default:0"`
	Status   int8   `json:"status" gorm:"default:1"`
}

// Brand 品牌
type Brand struct {
	BaseModelWithCompany
	Name        string `json:"name" gorm:"size:64;not null"`
	Code        string `json:"code" gorm:"size:64;index"`
	Description string `json:"description" gorm:"size:255"`
	Image       string `json:"image" gorm:"size:255"`
	CategoryID  uint   `json:"categoryId" gorm:"index;default:0"` // 品牌分类 id，0=未分类
	ApplyImage  int8   `json:"applyImage" gorm:"default:0"`       // 应用为商品展示图 0否1是
	Sort        int    `json:"sort" gorm:"default:0"`
	Status      int8   `json:"status" gorm:"default:1"`
}

// BrandCategory 品牌分类（树形）
type BrandCategory struct {
	BaseModelWithCompany
	ParentID uint   `json:"parentId" gorm:"index;default:0"`
	Name     string `json:"name" gorm:"size:64;not null"`
	Sort     int    `json:"sort" gorm:"default:0"`
	Status   int8   `json:"status" gorm:"default:1"`
}

// ProductSeries 商品系列
type ProductSeries struct {
	BaseModelWithCompany
	BrandID uint   `json:"brandId" gorm:"index"`
	Name    string `json:"name" gorm:"size:64;not null"`
	Code    string `json:"code" gorm:"size:64;index"`
	Status  int8   `json:"status" gorm:"default:1"`
}

// Product 商品资料
type Product struct {
	BaseModelWithCompany
	CategoryID     uint    `json:"categoryId" gorm:"index;not null"`
	BrandID        uint    `json:"brandId" gorm:"index"`
	SeriesID       uint    `json:"seriesId" gorm:"index"`
	Name           string  `json:"name" gorm:"size:128;not null"`
	Code           string  `json:"code" gorm:"size:64;index"`
	Barcode        string  `json:"barcode" gorm:"size:64;index"`
	Specification  string  `json:"specification" gorm:"size:128"`
	Unit           string  `json:"unit" gorm:"size:32;not null"` // 主单位
	PurchasePrice  float64 `json:"purchasePrice" gorm:"type:decimal(18,4);default:0"`
	RetailPrice    float64 `json:"retailPrice" gorm:"type:decimal(18,4);default:0"`
	WholesalePrice float64 `json:"wholesalePrice" gorm:"type:decimal(18,4);default:0"`
	MinStock       float64 `json:"minStock" gorm:"type:decimal(18,4);default:0"`
	MaxStock       float64 `json:"maxStock" gorm:"type:decimal(18,4);default:0"`
	Description    string  `json:"description" gorm:"size:500"`
	Status         int8    `json:"status" gorm:"default:1"`

	Image                   string  `json:"image" gorm:"size:255"`                 // 商品主图
	MallName                string  `json:"mallName" gorm:"size:128"`              // 商城展示名称
	PinyinCode              string  `json:"pinyinCode" gorm:"size:128"`            // 拼音码
	VideoURL                string  `json:"videoUrl" gorm:"size:255"`
	DetailContent           string  `json:"detailContent" gorm:"type:text"`        // 商品详情 JSON 块
	TopicCategory           string  `json:"topicCategory" gorm:"size:64"`          // 专题分类
	MinOrderQty             float64 `json:"minOrderQty" gorm:"default:0"`          // 起订量
	MaxOrderQty             float64 `json:"maxOrderQty" gorm:"default:0"`          // 限订量
	OrderByMultiple         int8    `json:"orderByMultiple" gorm:"default:0"`      // 按起订量倍数订购
	NoAvailableStockControl int8    `json:"noAvailableStockControl" gorm:"default:0"`
	NoBookStockControl      int8    `json:"noBookStockControl" gorm:"default:0"`
	ManualLevelPriceDefault int8    `json:"manualLevelPriceDefault" gorm:"default:0"` // 手动设置的级别价取默认订货价
	AuxPriceReverseCalc     int8    `json:"auxPriceReverseCalc" gorm:"default:0"`     // 辅助单位价格变更后反算基本单位价格

	MinSalePrice        float64 `json:"minSalePrice" gorm:"type:decimal(18,4);default:0"` // 最低售价
	MallSortWeight      int     `json:"mallSortWeight" gorm:"default:0"`                  // 商城排序权重
	SearchKeywords      string  `json:"searchKeywords" gorm:"size:255"`                   // 搜索关键词
	WarehouseID         uint    `json:"warehouseId" gorm:"index;default:0"`               // 出库仓库
	SupplierID          uint    `json:"supplierId" gorm:"index;default:0"`                // 默认供应商
	ForbidPurchaseUnits string  `json:"forbidPurchaseUnits" gorm:"size:255"`              // 禁购单位(逗号分隔单位名)
}

// ProductSpecItem 商品规格行
type ProductSpecItem struct {
	BaseModelWithCompany
	ProductID uint    `json:"productId" gorm:"index;not null"`
	SpecValue string  `json:"specValue" gorm:"size:200"`
	Code      string  `json:"code" gorm:"size:64"`
	Barcode   string  `json:"barcode" gorm:"size:64"`
	Weight    float64 `json:"weight" gorm:"default:0"`
	Volume    float64 `json:"volume" gorm:"default:0"`
	Remark1   string  `json:"remark1" gorm:"size:200"`
	Remark2   string  `json:"remark2" gorm:"size:200"`
	OnShelf   int8    `json:"onShelf"` // 1上架 0下架（默认逻辑在 handler，gorm default 会吞掉显式 0）
	Sort      int     `json:"sort" gorm:"default:0"`
}

// ProductUnit 商品辅助单位
type ProductUnit struct {
	BaseModel
	ProductID       uint    `json:"productId" gorm:"index;not null"`
	Name            string  `json:"name" gorm:"size:32;not null"`                   // 单位名称
	Conversion      float64 `json:"conversion" gorm:"type:decimal(18,4);default:1"` // 换算系数（相对于主单位）
	IsDefault       bool    `json:"isDefault" gorm:"default:false"`                 // 是否默认单位
	Barcode         string  `json:"barcode" gorm:"size:64"`
	AllowSale       int8    `json:"allowSale"` // 1允许销售 0不允许（默认逻辑在 handler，gorm default 会吞掉显式 0）
	DefaultSale     int8    `json:"defaultSale" gorm:"default:0"`
	DefaultPurchase int8    `json:"defaultPurchase" gorm:"default:0"`
	StorageType     int8    `json:"storageType" gorm:"default:1"` // 1散货 2整件
	RefPrice        float64 `json:"refPrice" gorm:"type:decimal(18,4);default:0"`
	WholesalePrice  float64 `json:"wholesalePrice" gorm:"type:decimal(18,4);default:0"`
	RetailPrice     float64 `json:"retailPrice" gorm:"type:decimal(18,4);default:0"`
	MinSalePrice    float64 `json:"minSalePrice" gorm:"type:decimal(18,4);default:0"`
	DefaultPrice    float64 `json:"defaultPrice" gorm:"type:decimal(18,4);default:0"`
	Status          int8    `json:"status" gorm:"default:1"`
}

// ProductBarcode 商品条码（多单位条码）
type ProductBarcode struct {
	BaseModel
	ProductID uint   `json:"productId" gorm:"index;not null"`
	UnitID    uint   `json:"unitId" gorm:"index;default:0"` // 0=主单位
	Barcode   string `json:"barcode" gorm:"size:64;not null;index"`
	Status    int8   `json:"status" gorm:"default:1"`
}

// GoodsUnit 商品单位（字典）
type GoodsUnit struct {
	BaseModelWithCompany
	Name        string `json:"name" gorm:"size:32;not null"`
	Code        string `json:"code" gorm:"size:32;index"`
	Sort        int    `json:"sort" gorm:"default:0"`
	StorageType int8   `json:"storageType" gorm:"default:1"` // 1=散货，2=整件
	Status      int8   `json:"status" gorm:"default:1"`
}

// ProductSpec 商品规格（如颜色、尺寸）
type ProductSpec struct {
	BaseModelWithCompany
	Name   string `json:"name" gorm:"size:64;not null"`
	Values string `json:"values" gorm:"type:text"` // 可选值，逗号分隔
	Remark string `json:"remark" gorm:"size:255"`
	Sort   int    `json:"sort" gorm:"default:0"`
	Status int8   `json:"status" gorm:"default:1"`
}

// ProductTag 商品标签
type ProductTag struct {
	BaseModelWithCompany
	Name       string `json:"name" gorm:"size:64;not null"`
	Color      string `json:"color" gorm:"size:16;default:'#409EFF'"`
	Type       int8   `json:"type" gorm:"default:1;comment:1手动 2智能"`
	Rule       string `json:"rule" gorm:"size:4000"`
	Filterable int8   `json:"filterable" gorm:"default:1;comment:商城支持筛选"`
	ShowInList int8   `json:"showInList" gorm:"default:0;comment:商城列表展示"`
	ShowBadge  int8   `json:"showBadge" gorm:"default:0;comment:商品卡片角标"`
	Sort       int    `json:"sort" gorm:"default:0"`
	Status     int8   `json:"status" gorm:"default:1"`
}

// ProductTagRelation 商品-标签关联
type ProductTagRelation struct {
	BaseModelWithCompany
	ProductID uint `json:"productId" gorm:"index;not null"`
	TagID     uint `json:"tagId" gorm:"index;not null"`
}

// ProductSalesScope 商品销售范围
type ProductSalesScope struct {
	BaseModelWithCompany
	Name   string `json:"name" gorm:"size:64;not null"`
	Code   string `json:"code" gorm:"size:32;index"`
	Status int8   `json:"status" gorm:"default:1"`
}

// ProductSaleRule 商品销售范围规则
type ProductSaleRule struct {
	BaseModelWithCompany
	TargetType string `json:"targetType" gorm:"size:16;not null;index:idx_sale_rule"` // category/brand/supplier/product
	TargetID   uint   `json:"targetId" gorm:"not null;index:idx_sale_rule"`
	Rule       string `json:"rule" gorm:"size:4000"`
}

// ==================== 客户/供应商模块 ====================

// CustomerCategory 客户分类
type CustomerCategory struct {
	BaseModelWithCompany
	ParentID uint   `json:"parentId" gorm:"index;default:0"`
	Name     string `json:"name" gorm:"size:64;not null"`
	Code     string `json:"code" gorm:"size:64"`
	Sort     int    `json:"sort" gorm:"default:0"`
	Status   int8   `json:"status" gorm:"default:1"`
}

// Region 区域
type Region struct {
	BaseModelWithCompany
	ParentID uint   `json:"parentId" gorm:"index;default:0"`
	Name     string `json:"name" gorm:"size:64;not null"`
	Code     string `json:"code" gorm:"size:64"`
	Sort     int    `json:"sort" gorm:"default:0"`
	Status   int8   `json:"status" gorm:"default:1"`
}

// CustomerTag 客户标签
type CustomerTag struct {
	BaseModelWithCompany
	Name   string `json:"name" gorm:"size:64;not null"`
	Status int8   `json:"status" gorm:"default:1"`
}

// Supplier 供应商
type Supplier struct {
	BaseModelWithCompany
	Name    string `json:"name" gorm:"size:64;not null"`
	Code    string `json:"code" gorm:"size:64"`
	Contact string `json:"contact" gorm:"size:64"`
	Phone   string `json:"phone" gorm:"size:32"`
	Status  int8   `json:"status" gorm:"default:1"`
}

// CustomerLevel 客户等级
type CustomerLevel struct {
	BaseModelWithCompany
	Name        string  `json:"name" gorm:"size:64;not null"`
	Code        string  `json:"code" gorm:"size:64;index"`
	MinAmount   float64 `json:"minAmount" gorm:"type:decimal(18,4);default:0"`
	MaxAmount   float64 `json:"maxAmount" gorm:"type:decimal(18,4);default:0"`
	Discount    float64 `json:"discount" gorm:"type:decimal(4,2);default:100"` // 折扣率 0-100
	Description string  `json:"description" gorm:"size:255"`
	Status      int8    `json:"status" gorm:"default:1"`
}

// Customer 客户/供应商
type Customer struct {
	BaseModelWithCompany
	CategoryID  uint       `json:"categoryId" gorm:"index"`
	RegionID    uint       `json:"regionId" gorm:"index"`
	LevelID     uint       `json:"levelId" gorm:"index"`
	Name        string     `json:"name" gorm:"size:128;not null"`
	Code        string     `json:"code" gorm:"size:64;index"`
	Type        string     `json:"type" gorm:"size:20;default:customer;comment:customer客户 supplier供应商 both两者"`
	Contact     string     `json:"contact" gorm:"size:64"`
	Phone       string     `json:"phone" gorm:"size:20;index"`
	Email       string     `json:"email" gorm:"size:128"`
	Address     string     `json:"address" gorm:"size:255"`
	CreditLimit float64    `json:"creditLimit" gorm:"type:decimal(18,4);default:0"` // 信用额度
	CreditDays  int        `json:"creditDays" gorm:"default:0"`                     // 账期天数
	Balance     float64    `json:"balance" gorm:"type:decimal(18,4);default:0"`     // 欠款余额（正=欠款）
	LastOrderAt *time.Time `json:"lastOrderAt" gorm:"index"`                        // 最近下单时间
	TaxNo       string     `json:"taxNo" gorm:"size:64"`                            // 税号
	BankName    string     `json:"bankName" gorm:"size:128"`
	BankAccount string     `json:"bankAccount" gorm:"size:64"`
	Remark      string     `json:"remark" gorm:"size:500"`
	Status      int8       `json:"status" gorm:"default:1"`
}

// ==================== 仓库/资金模块 ====================

// Warehouse 仓库
type Warehouse struct {
	BaseModelWithCompany
	Name      string `json:"name" gorm:"size:64;not null"`
	Code      string `json:"code" gorm:"size:64;index"`
	Address   string `json:"address" gorm:"size:255"`
	ManagerID uint   `json:"managerId"`
	Status    int8   `json:"status" gorm:"default:1"`
}

// WarehousePosition 仓位
type WarehousePosition struct {
	BaseModel
	WarehouseID uint   `json:"warehouseId" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:64;not null"`
	Code        string `json:"code" gorm:"size:64"`
	Status      int8   `json:"status" gorm:"default:1"`
}

// Account 资金账户
type Account struct {
	BaseModelWithCompany
	Name        string  `json:"name" gorm:"size:64;not null"`
	Code        string  `json:"code" gorm:"size:64;index"`
	Type        string  `json:"type" gorm:"size:20;not null;comment:cash现金 bank银行 alipay支付宝 wechat微信"`
	Balance     float64 `json:"balance" gorm:"type:decimal(18,4);default:0"`
	Description string  `json:"description" gorm:"size:255"`
	Status      int8    `json:"status" gorm:"default:1"`
}

// IncomeExpenseItem 收支项目
type IncomeExpenseItem struct {
	BaseModelWithCompany
	Name   string `json:"name" gorm:"size:64;not null"`
	Code   string `json:"code" gorm:"size:64;index"`
	Type   string `json:"type" gorm:"size:10;not null;comment:income收入 expense支出"`
	Sort   int    `json:"sort" gorm:"default:0"`
	Status int8   `json:"status" gorm:"default:1"`
}

// AccountFlow 资金账户流水
type AccountFlow struct {
	BaseModelWithCompany
	AccountID uint    `json:"accountId" gorm:"index;not null"`
	Type      string  `json:"type" gorm:"size:10;not null;comment:income收入 expense支出"`
	Amount    float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	RefType   string  `json:"refType" gorm:"size:32;index"` // 业务类型 sales_receipt|purchase_payment 等
	RefID     uint    `json:"refId" gorm:"index;default:0"` // 业务单据ID
	Remark    string  `json:"remark" gorm:"size:255"`
}

// Stock 库存台账（公司+仓库+商品 维度）
type Stock struct {
	BaseModelWithCompany
	WarehouseID uint    `json:"warehouseId" gorm:"index;not null;uniqueIndex:uniq_stock_wh_product,priority:2"`
	ProductID   uint    `json:"productId" gorm:"index;not null;uniqueIndex:uniq_stock_wh_product,priority:3"`
	Quantity    float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
}

func (Stock) TableName() string {
	return "stocks"
}

// ==================== 价格体系模块 ====================

// PriceLevel 价格体系
type PriceLevel struct {
	BaseModelWithCompany
	Name        string `json:"name" gorm:"size:64;not null"`
	Code        string `json:"code" gorm:"size:64;index"`
	Description string `json:"description" gorm:"size:255"`
	IsDefault   bool   `json:"isDefault" gorm:"default:false"`
	Status      int8   `json:"status" gorm:"default:1"`
}

// ProductPrice 商品价格矩阵
type ProductPrice struct {
	BaseModel
	ProductID  uint    `json:"productId" gorm:"index;not null"`
	UnitID     uint    `json:"unitId" gorm:"index;default:0"`     // 0=主单位
	LevelID    uint    `json:"levelId" gorm:"index"`              // 价格体系ID
	CustomerID uint    `json:"customerId" gorm:"index;default:0"` // 0=非专属
	Price      float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Status     int8    `json:"status" gorm:"default:1"`
}

// ==================== 采购模块 ====================

// PurchaseOrder 采购订单
type PurchaseOrder struct {
	BaseModelWithCompany
	SupplierID   uint                `json:"supplierId" gorm:"index;not null"`
	WarehouseID  uint                `json:"warehouseId" gorm:"index;not null"`
	OrderNo      string              `json:"orderNo" gorm:"size:64;index;not null"`
	OrderDate    time.Time           `json:"orderDate" gorm:"type:date"`
	DeliveryDate time.Time           `json:"deliveryDate" gorm:"type:date"`
	Amount       float64             `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount     float64             `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TaxAmount    float64             `json:"taxAmount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount  float64             `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	PaidAmount   float64             `json:"paidAmount" gorm:"type:decimal(18,4);default:0"`
	Status       string              `json:"status" gorm:"size:20;default:draft;comment:draft草稿 confirmed已确认 partial部分入库 completed已完成 cancelled已取消"`
	OperatorID   uint                `json:"operatorId" gorm:"index"`
	Remark       string              `json:"remark" gorm:"size:500"`
	Items        []PurchaseOrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

// PurchaseOrderItem 采购订单明细
type PurchaseOrderItem struct {
	BaseModel
	OrderID     uint    `json:"orderId" gorm:"index;not null"`
	ProductID   uint    `json:"productId" gorm:"index;not null"`
	UnitID      uint    `json:"unitId" gorm:"index;default:0"`
	Quantity    float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price       float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount      float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount    float64 `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TaxRate     float64 `json:"taxRate" gorm:"type:decimal(8,4);default:0"`
	TaxAmount   float64 `json:"taxAmount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64 `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	ReceivedQty float64 `json:"receivedQty" gorm:"type:decimal(18,4);default:0"`
	Remark      string  `json:"remark" gorm:"size:255"`
}

// PurchaseInStock 采购入库单
type PurchaseInStock struct {
	BaseModelWithCompany
	SupplierID  uint                  `json:"supplierId" gorm:"index;not null"`
	WarehouseID uint                  `json:"warehouseId" gorm:"index;not null"`
	OrderID     uint                  `json:"orderId" gorm:"index;default:0"`
	BillNo      string                `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time             `json:"billDate" gorm:"type:date"`
	Amount      float64               `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount    float64               `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64               `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	PaidAmount  float64               `json:"paidAmount" gorm:"type:decimal(18,4);default:0"`
	Status      string                `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID  uint                  `json:"operatorId" gorm:"index"`
	Remark      string                `json:"remark" gorm:"size:500"`
	Items       []PurchaseInStockItem `json:"items" gorm:"foreignKey:InStockID"`
}

// PurchaseInStockItem 采购入库明细
type PurchaseInStockItem struct {
	BaseModel
	InStockID   uint      `json:"inStockId" gorm:"index;not null"`
	OrderItemID uint      `json:"orderItemId" gorm:"index;default:0"`
	ProductID   uint      `json:"productId" gorm:"index;not null"`
	UnitID      uint      `json:"unitId" gorm:"index;default:0"`
	Quantity    float64   `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price       float64   `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount      float64   `json:"amount" gorm:"type:decimal(18,4);default:0"`
	BatchNo     string    `json:"batchNo" gorm:"size:64"`
	ExpiryDate  time.Time `json:"expiryDate" gorm:"type:date"`
	PositionID  uint      `json:"positionId" gorm:"index;default:0"`
	Remark      string    `json:"remark" gorm:"size:255"`
}

// PurchaseReturn 采购退货单
type PurchaseReturn struct {
	BaseModelWithCompany
	SupplierID  uint                 `json:"supplierId" gorm:"index;not null"`
	WarehouseID uint                 `json:"warehouseId" gorm:"index;not null"`
	InStockID   uint                 `json:"inStockId" gorm:"index;default:0"`
	BillNo      string               `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time            `json:"billDate" gorm:"type:date"`
	Amount      float64              `json:"amount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64              `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	Status      string               `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID  uint                 `json:"operatorId" gorm:"index"`
	Remark      string               `json:"remark" gorm:"size:500"`
	Items       []PurchaseReturnItem `json:"items" gorm:"foreignKey:ReturnID"`
}

// PurchaseReturnItem 采购退货明细
type PurchaseReturnItem struct {
	BaseModel
	ReturnID      uint    `json:"returnId" gorm:"index;not null"`
	InStockItemID uint    `json:"inStockItemId" gorm:"index;default:0"`
	ProductID     uint    `json:"productId" gorm:"index;not null"`
	UnitID        uint    `json:"unitId" gorm:"index;default:0"`
	Quantity      float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price         float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount        float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Remark        string  `json:"remark" gorm:"size:255"`
}

// PurchasePayment 采购付款单
type PurchasePayment struct {
	BaseModelWithCompany
	SupplierID  uint                  `json:"supplierId" gorm:"index;not null"`
	AccountID   uint                  `json:"accountId" gorm:"index;not null"`
	BillNo      string                `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time             `json:"billDate" gorm:"type:date"`
	Amount      float64               `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount    float64               `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64               `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	Status      string                `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID  uint                  `json:"operatorId" gorm:"index"`
	Remark      string                `json:"remark" gorm:"size:500"`
	Items       []PurchasePaymentItem `json:"items" gorm:"foreignKey:PaymentID"`
}

// PurchasePaymentItem 采购付款明细
type PurchasePaymentItem struct {
	BaseModel
	PaymentID uint    `json:"paymentId" gorm:"index;not null"`
	OrderID   uint    `json:"orderId" gorm:"index;default:0"`
	InStockID uint    `json:"inStockId" gorm:"index;default:0"`
	Amount    float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount  float64 `json:"discount" gorm:"type:decimal(18,4);default:0"`
	Remark    string  `json:"remark" gorm:"size:255"`
}

// ==================== 销售模块 ====================

// SalesOrder 销售订单
type SalesOrder struct {
	BaseModelWithCompany
	CustomerID   uint             `json:"customerId" gorm:"index;not null"`
	WarehouseID  uint             `json:"warehouseId" gorm:"index;not null"`
	OrderNo      string           `json:"orderNo" gorm:"size:64;index;not null"`
	OrderDate    time.Time        `json:"orderDate" gorm:"type:date"`
	DeliveryDate time.Time        `json:"deliveryDate" gorm:"type:date"`
	Amount       float64          `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount     float64          `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TaxAmount    float64          `json:"taxAmount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount  float64          `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	PaidAmount   float64          `json:"paidAmount" gorm:"type:decimal(18,4);default:0"`
	Status       string           `json:"status" gorm:"size:20;default:draft;comment:draft草稿 confirmed已确认 partial部分出库 completed已完成 cancelled已取消"`
	OperatorID   uint             `json:"operatorId" gorm:"index"`
	Remark       string           `json:"remark" gorm:"size:500"`
	Items        []SalesOrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

// SalesOrderItem 销售订单明细
type SalesOrderItem struct {
	BaseModel
	OrderID      uint    `json:"orderId" gorm:"index;not null"`
	ProductID    uint    `json:"productId" gorm:"index;not null"`
	UnitID       uint    `json:"unitId" gorm:"index;default:0"`
	Quantity     float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price        float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount       float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount     float64 `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TaxRate      float64 `json:"taxRate" gorm:"type:decimal(8,4);default:0"`
	TaxAmount    float64 `json:"taxAmount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount  float64 `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	DeliveredQty float64 `json:"deliveredQty" gorm:"type:decimal(18,4);default:0"`
	Remark       string  `json:"remark" gorm:"size:255"`
}

// SalesOutStock 销售出库单
type SalesOutStock struct {
	BaseModelWithCompany
	CustomerID  uint                `json:"customerId" gorm:"index;not null"`
	WarehouseID uint                `json:"warehouseId" gorm:"index;not null"`
	OrderID     uint                `json:"orderId" gorm:"index;default:0"`
	BillNo      string              `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time           `json:"billDate" gorm:"type:date"`
	Amount      float64             `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount    float64             `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64             `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	PaidAmount  float64             `json:"paidAmount" gorm:"type:decimal(18,4);default:0"`
	Status      string              `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID  uint                `json:"operatorId" gorm:"index"`
	Remark      string              `json:"remark" gorm:"size:500"`
	Items       []SalesOutStockItem `json:"items" gorm:"foreignKey:OutStockID"`
}

// SalesOutStockItem 销售出库明细
type SalesOutStockItem struct {
	BaseModel
	OutStockID  uint    `json:"outStockId" gorm:"index;not null"`
	OrderItemID uint    `json:"orderItemId" gorm:"index;default:0"`
	ProductID   uint    `json:"productId" gorm:"index;not null"`
	UnitID      uint    `json:"unitId" gorm:"index;default:0"`
	Quantity    float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price       float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount      float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	BatchNo     string  `json:"batchNo" gorm:"size:64"`
	PositionID  uint    `json:"positionId" gorm:"index;default:0"`
	Remark      string  `json:"remark" gorm:"size:255"`
}

// SalesReturn 销售退货单
type SalesReturn struct {
	BaseModelWithCompany
	CustomerID  uint              `json:"customerId" gorm:"index;not null"`
	WarehouseID uint              `json:"warehouseId" gorm:"index;not null"`
	OutStockID  uint              `json:"outStockId" gorm:"index;default:0"`
	BillNo      string            `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time         `json:"billDate" gorm:"type:date"`
	Amount      float64           `json:"amount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64           `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	Status      string            `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID  uint              `json:"operatorId" gorm:"index"`
	Remark      string            `json:"remark" gorm:"size:500"`
	Items       []SalesReturnItem `json:"items" gorm:"foreignKey:ReturnID"`
}

// SalesReturnItem 销售退货明细
type SalesReturnItem struct {
	BaseModel
	ReturnID       uint    `json:"returnId" gorm:"index;not null"`
	OutStockItemID uint    `json:"outStockItemId" gorm:"index;default:0"`
	ProductID      uint    `json:"productId" gorm:"index;not null"`
	UnitID         uint    `json:"unitId" gorm:"index;default:0"`
	Quantity       float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price          float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount         float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Remark         string  `json:"remark" gorm:"size:255"`
}

// SalesReceipt 销售收款单
type SalesReceipt struct {
	BaseModelWithCompany
	CustomerID  uint               `json:"customerId" gorm:"index;not null"`
	AccountID   uint               `json:"accountId" gorm:"index;not null"`
	BillNo      string             `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time          `json:"billDate" gorm:"type:date"`
	Amount      float64            `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount    float64            `json:"discount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64            `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	Status      string             `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID  uint               `json:"operatorId" gorm:"index"`
	Remark      string             `json:"remark" gorm:"size:500"`
	Items       []SalesReceiptItem `json:"items" gorm:"foreignKey:ReceiptID"`
}

// SalesReceiptItem 销售收款明细
type SalesReceiptItem struct {
	BaseModel
	ReceiptID  uint    `json:"receiptId" gorm:"index;not null"`
	OrderID    uint    `json:"orderId" gorm:"index;default:0"`
	OutStockID uint    `json:"outStockId" gorm:"index;default:0"`
	Amount     float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Discount   float64 `json:"discount" gorm:"type:decimal(18,4);default:0"`
	Remark     string  `json:"remark" gorm:"size:255"`
}

// ==================== 库存模块 ====================

// InventoryCheck 库存盘点单
type InventoryCheck struct {
	BaseModelWithCompany
	WarehouseID uint                 `json:"warehouseId" gorm:"index;not null"`
	BillNo      string               `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time            `json:"billDate" gorm:"type:date"`
	Status      string               `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID  uint                 `json:"operatorId" gorm:"index"`
	Remark      string               `json:"remark" gorm:"size:500"`
	Items       []InventoryCheckItem `json:"items" gorm:"foreignKey:CheckID"`
}

// InventoryCheckItem 库存盘点明细
type InventoryCheckItem struct {
	BaseModel
	CheckID    uint    `json:"checkId" gorm:"index;not null"`
	ProductID  uint    `json:"productId" gorm:"index;not null"`
	PositionID uint    `json:"positionId" gorm:"index;default:0"`
	BookQty    float64 `json:"bookQty" gorm:"type:decimal(18,4);default:0"`
	ActualQty  float64 `json:"actualQty" gorm:"type:decimal(18,4);default:0"`
	DiffQty    float64 `json:"diffQty" gorm:"type:decimal(18,4);default:0"`
	Price      float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount     float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Remark     string  `json:"remark" gorm:"size:255"`
}

// InventoryTransfer 库存调拨单
type InventoryTransfer struct {
	BaseModelWithCompany
	FromWarehouseID uint                    `json:"fromWarehouseId" gorm:"index;not null"`
	ToWarehouseID   uint                    `json:"toWarehouseId" gorm:"index;not null"`
	BillNo          string                  `json:"billNo" gorm:"size:64;index;not null"`
	BillDate        time.Time               `json:"billDate" gorm:"type:date"`
	Amount          float64                 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Status          string                  `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	OperatorID      uint                    `json:"operatorId" gorm:"index"`
	Remark          string                  `json:"remark" gorm:"size:500"`
	Items           []InventoryTransferItem `json:"items" gorm:"foreignKey:TransferID"`
}

// InventoryTransferItem 库存调拨明细
type InventoryTransferItem struct {
	BaseModel
	TransferID     uint    `json:"transferId" gorm:"index;not null"`
	ProductID      uint    `json:"productId" gorm:"index;not null"`
	FromPositionID uint    `json:"fromPositionId" gorm:"index;default:0"`
	ToPositionID   uint    `json:"toPositionId" gorm:"index;default:0"`
	Quantity       float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price          float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount         float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Remark         string  `json:"remark" gorm:"size:255"`
}

// InventoryWarning 库存预警
type InventoryWarning struct {
	BaseModelWithCompany
	WarehouseID  uint    `json:"warehouseId" gorm:"index;not null"`
	ProductID    uint    `json:"productId" gorm:"index;not null"`
	MinStock     float64 `json:"minStock" gorm:"type:decimal(18,4);default:0"`
	MaxStock     float64 `json:"maxStock" gorm:"type:decimal(18,4);default:0"`
	CurrentStock float64 `json:"currentStock" gorm:"type:decimal(18,4);default:0"`
	WarningType  string  `json:"warningType" gorm:"size:20;comment:low低库存 high高库存"`
	Status       string  `json:"status" gorm:"size:20;default:active;comment:active预警中 resolved已解决"`
	Remark       string  `json:"remark" gorm:"size:255"`
}

// ==================== 商城模块 ====================

// MallProduct 商城商品展示
type MallProduct struct {
	BaseModelWithCompany
	ProductID   uint    `json:"productId" gorm:"index;not null"`
	Name        string  `json:"name" gorm:"size:128;not null"`
	Description string  `json:"description" gorm:"size:500"`
	ImageURL    string  `json:"imageUrl" gorm:"size:255"`
	Price       float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Status      int8    `json:"status" gorm:"default:1"`
}

// MallCart 购物车
type MallCart struct {
	BaseModelWithCompany
	CustomerID uint    `json:"customerId" gorm:"index;not null"`
	ProductID  uint    `json:"productId" gorm:"index;not null"`
	Quantity   float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price      float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
}

// MallOrder 商城订单
type MallOrder struct {
	BaseModelWithCompany
	CustomerID  uint            `json:"customerId" gorm:"index;not null"`
	OrderNo     string          `json:"orderNo" gorm:"size:64;index;not null"`
	OrderDate   time.Time       `json:"orderDate" gorm:"type:date"`
	Amount      float64         `json:"amount" gorm:"type:decimal(18,4);default:0"`
	TotalAmount float64         `json:"totalAmount" gorm:"type:decimal(18,4);default:0"`
	Status      string          `json:"status" gorm:"size:20;default:pending;comment:pending待处理 paid已付款 shipped已发货 completed已完成 cancelled已取消"`
	OperatorID  uint            `json:"operatorId" gorm:"index"`
	Remark      string          `json:"remark" gorm:"size:500"`
	Items       []MallOrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

// MallOrderItem 商城订单明细
type MallOrderItem struct {
	BaseModel
	OrderID   uint    `json:"orderId" gorm:"index;not null"`
	ProductID uint    `json:"productId" gorm:"index;not null"`
	Quantity  float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price     float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount    float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Remark    string  `json:"remark" gorm:"size:255"`
}

// ==================== CRM模块 ====================

// FollowUp 客户跟进
type FollowUp struct {
	BaseModelWithCompany
	CustomerID  uint      `json:"customerId" gorm:"index;not null"`
	ContactDate time.Time `json:"contactDate" gorm:"type:date"`
	Type        string    `json:"type" gorm:"size:20;comment:phone电话 visit拜访 email邮件 wechat微信"`
	Content     string    `json:"content" gorm:"size:500"`
	NextPlan    string    `json:"nextPlan" gorm:"size:500"`
	OperatorID  uint      `json:"operatorId" gorm:"index"`
	Status      int8      `json:"status" gorm:"default:1"`
}

// Opportunity 商机
type Opportunity struct {
	BaseModelWithCompany
	CustomerID   uint      `json:"customerId" gorm:"index;not null"`
	Name         string    `json:"name" gorm:"size:128;not null"`
	Amount       float64   `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Stage        string    `json:"stage" gorm:"size:20;default:lead;comment:lead线索 qualification资格审查 proposal方案 negotiation谈判 closed成交 lost丢失"`
	Probability  float64   `json:"probability" gorm:"type:decimal(5,2);default:0"`
	ExpectedDate time.Time `json:"expectedDate" gorm:"type:date"`
	OperatorID   uint      `json:"operatorId" gorm:"index"`
	Remark       string    `json:"remark" gorm:"size:500"`
	Status       int8      `json:"status" gorm:"default:1"`
}

// Contract 合同
type Contract struct {
	BaseModelWithCompany
	CustomerID    uint      `json:"customerId" gorm:"index;not null"`
	OpportunityID uint      `json:"opportunityId" gorm:"index;default:0"`
	ContractNo    string    `json:"contractNo" gorm:"size:64;index;not null"`
	ContractDate  time.Time `json:"contractDate" gorm:"type:date"`
	Amount        float64   `json:"amount" gorm:"type:decimal(18,4);default:0"`
	StartDate     time.Time `json:"startDate" gorm:"type:date"`
	EndDate       time.Time `json:"endDate" gorm:"type:date"`
	Status        string    `json:"status" gorm:"size:20;default:draft;comment:draft草稿 active生效 completed完成 terminated终止"`
	OperatorID    uint      `json:"operatorId" gorm:"index"`
	Remark        string    `json:"remark" gorm:"size:500"`
}

// ==================== 审批模块 ====================

// ApprovalProcess 审批流程
type ApprovalProcess struct {
	BaseModelWithCompany
	Name        string `json:"name" gorm:"size:128;not null"`
	Code        string `json:"code" gorm:"size:64;index;not null"`
	Type        string `json:"type" gorm:"size:20;comment:purchase采购 sale销售 transfer调拨 expense费用"`
	Description string `json:"description" gorm:"size:500"`
	Status      int8   `json:"status" gorm:"default:1"`
}

// ApprovalRecord 审批记录
type ApprovalRecord struct {
	BaseModelWithCompany
	ProcessID    uint      `json:"processId" gorm:"index;not null"`
	BusinessType string    `json:"businessType" gorm:"size:20;not null"`
	BusinessID   uint      `json:"businessId" gorm:"index;not null"`
	ApplicantID  uint      `json:"applicantId" gorm:"index"`
	ApproverID   uint      `json:"approverId" gorm:"index"`
	Action       string    `json:"action" gorm:"size:20;default:pending;comment:pending待审批 approved通过 rejected拒绝"`
	Comment      string    `json:"comment" gorm:"size:500"`
	ApproveDate  time.Time `json:"approveDate" gorm:"type:date"`
	Status       int8      `json:"status" gorm:"default:1"`
}

// ==================== 公司级设置 ====================

// CompanySetting 公司级键值设置
type CompanySetting struct {
	BaseModelWithCompany
	Key   string `json:"key" gorm:"size:64;not null;uniqueIndex:idx_company_setting"`
	Value string `json:"value" gorm:"size:2000"`
}

// ==================== 报表模块（使用查询实现） ====================
