package model

import "time"

// ==================== 库存扩展模块（03-库存） =====================

// CostAdjust 成本调价单
type CostAdjust struct {
	BaseModelWithCompany
	WarehouseID uint             `json:"warehouseId" gorm:"index;not null"`
	BillNo      string           `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time        `json:"billDate" gorm:"type:date"`
	Amount      float64          `json:"amount" gorm:"type:decimal(18,4);default:0;comment:调价差额合计"`
	Status      string           `json:"status" gorm:"size:20;default:draft;comment:draft草稿 completed已过账 cancelled已撤销"`
	HandlerID   uint             `json:"handlerId" gorm:"index"`
	OperatorID  uint             `json:"operatorId" gorm:"index"`
	RefBillNo   string           `json:"refBillNo" gorm:"size:64"`
	Remark      string           `json:"remark" gorm:"size:500"`
	Items       []CostAdjustItem `json:"items" gorm:"foreignKey:AdjustID"`
}

func (CostAdjust) TableName() string { return "cost_adjusts" }

// CostAdjustItem 成本调价单明细
type CostAdjustItem struct {
	BaseModel
	AdjustID   uint    `json:"adjustId" gorm:"index;not null"`
	ProductID  uint    `json:"productId" gorm:"index;not null"`
	Quantity   float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	OldPrice   float64 `json:"oldPrice" gorm:"type:decimal(18,4);default:0"`
	NewPrice   float64 `json:"newPrice" gorm:"type:decimal(18,4);default:0"`
	DiffAmount float64 `json:"diffAmount" gorm:"type:decimal(18,4);default:0"`
	Remark     string  `json:"remark" gorm:"size:255"`
	Product    Product `json:"product" gorm:"foreignKey:ProductID"`
}

func (CostAdjustItem) TableName() string { return "cost_adjust_items" }

// TransferApply 调拨申请单
type TransferApply struct {
	BaseModelWithCompany
	FromWarehouseID uint                `json:"fromWarehouseId" gorm:"index;not null"`
	ToWarehouseID   uint                `json:"toWarehouseId" gorm:"index;not null"`
	BillNo          string              `json:"billNo" gorm:"size:64;index;not null"`
	BillDate        time.Time           `json:"billDate" gorm:"type:date"`
	TotalQty        float64             `json:"totalQty" gorm:"type:decimal(18,4);default:0"`
	Status          string              `json:"status" gorm:"size:20;default:draft;comment:draft草稿 pending待审核 approved待出库 shipped待入库 completed已完成 cancelled已取消"`
	HandlerID       uint                `json:"handlerId" gorm:"index"`
	OperatorID      uint                `json:"operatorId" gorm:"index"`
	Remark          string              `json:"remark" gorm:"size:500"`
	Items           []TransferApplyItem `json:"items" gorm:"foreignKey:ApplyID"`
}

func (TransferApply) TableName() string { return "transfer_applies" }

// TransferApplyItem 调拨申请单明细
type TransferApplyItem struct {
	BaseModel
	ApplyID  uint    `json:"applyId" gorm:"index;not null"`
	ProductID uint   `json:"productId" gorm:"index;not null"`
	Quantity float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	ShippedQty float64 `json:"shippedQty" gorm:"type:decimal(18,4);default:0"`
	ReceivedQty float64 `json:"receivedQty" gorm:"type:decimal(18,4);default:0"`
	Remark   string  `json:"remark" gorm:"size:255"`
	Product  Product `json:"product" gorm:"foreignKey:ProductID"`
}

func (TransferApplyItem) TableName() string { return "transfer_apply_items" }

// TransferOut 调拨出库单
type TransferOut struct {
	BaseModelWithCompany
	ApplyID         uint              `json:"applyId" gorm:"index;default:0"`
	FromWarehouseID uint              `json:"fromWarehouseId" gorm:"index;not null"`
	ToWarehouseID   uint              `json:"toWarehouseId" gorm:"index;not null"`
	BillNo          string            `json:"billNo" gorm:"size:64;index;not null"`
	BillDate        time.Time         `json:"billDate" gorm:"type:date"`
	TransferType    string            `json:"transferType" gorm:"size:20;default:same;comment:same同价 diff变价"`
	TotalQty        float64           `json:"totalQty" gorm:"type:decimal(18,4);default:0"`
	Amount          float64           `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Status          string            `json:"status" gorm:"size:20;default:draft;comment:draft草稿 completed已过账 cancelled已撤销"`
	HandlerID       uint              `json:"handlerId" gorm:"index"`
	OperatorID      uint              `json:"operatorId" gorm:"index"`
	Remark          string            `json:"remark" gorm:"size:500"`
	Items           []TransferOutItem `json:"items" gorm:"foreignKey:OutID"`
}

func (TransferOut) TableName() string { return "transfer_outs" }

// TransferOutItem 调拨出库单明细
type TransferOutItem struct {
	BaseModel
	OutID       uint    `json:"outId" gorm:"index;not null"`
	ProductID   uint    `json:"productId" gorm:"index;not null"`
	Quantity    float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price       float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount      float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	ReceivedQty float64 `json:"receivedQty" gorm:"type:decimal(18,4);default:0"`
	DiffStatus  string  `json:"diffStatus" gorm:"size:20;default:none;comment:none无差异 pending待处理 handled已处理"`
	DiffResult  string  `json:"diffResult" gorm:"size:64"`
	Remark      string  `json:"remark" gorm:"size:255"`
	Product     Product `json:"product" gorm:"foreignKey:ProductID"`
}

func (TransferOutItem) TableName() string { return "transfer_out_items" }

// TransferIn 调拨入库单
type TransferIn struct {
	BaseModelWithCompany
	OutID           uint             `json:"outId" gorm:"index;default:0"`
	OutBillNo       string           `json:"outBillNo" gorm:"size:64;index"`
	FromWarehouseID uint             `json:"fromWarehouseId" gorm:"index;not null"`
	ToWarehouseID   uint             `json:"toWarehouseId" gorm:"index;not null"`
	BillNo          string           `json:"billNo" gorm:"size:64;index;not null"`
	BillDate        time.Time        `json:"billDate" gorm:"type:date"`
	TransferType    string           `json:"transferType" gorm:"size:20;default:same;comment:same同价 diff变价"`
	TotalQty        float64          `json:"totalQty" gorm:"type:decimal(18,4);default:0"`
	Amount          float64          `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Status          string           `json:"status" gorm:"size:20;default:draft;comment:draft草稿 completed已过账 cancelled已撤销"`
	HandlerID       uint             `json:"handlerId" gorm:"index"`
	OperatorID      uint             `json:"operatorId" gorm:"index"`
	Remark          string           `json:"remark" gorm:"size:500"`
	Items           []TransferInItem `json:"items" gorm:"foreignKey:InID"`
}

func (TransferIn) TableName() string { return "transfer_ins" }

// TransferInItem 调拨入库单明细
type TransferInItem struct {
	BaseModel
	InID      uint    `json:"inId" gorm:"index;not null"`
	ProductID uint    `json:"productId" gorm:"index;not null"`
	Quantity  float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price     float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount    float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Remark    string  `json:"remark" gorm:"size:255"`
	Product   Product `json:"product" gorm:"foreignKey:ProductID"`
}

func (TransferInItem) TableName() string { return "transfer_in_items" }

// AssemblyTemplate 拆装模板
type AssemblyTemplate struct {
	BaseModelWithCompany
	Name     string `json:"name" gorm:"size:128;not null"`
	OutItems string `json:"outItems" gorm:"type:text;comment:出库商品JSON [{productId,quantity}]"`
	InItems  string `json:"inItems" gorm:"type:text;comment:入库商品JSON [{productId,quantity}]"`
	Remark   string `json:"remark" gorm:"size:500"`
	Status   int8   `json:"status" gorm:"default:1"`
}

func (AssemblyTemplate) TableName() string { return "assembly_templates" }

// AssemblyOrder 组装拆装单
type AssemblyOrder struct {
	BaseModelWithCompany
	TemplateID     uint                `json:"templateId" gorm:"index;default:0"`
	OutWarehouseID uint                `json:"outWarehouseId" gorm:"index;not null"`
	InWarehouseID  uint                `json:"inWarehouseId" gorm:"index;not null"`
	BillNo         string              `json:"billNo" gorm:"size:64;index;not null"`
	BillDate       time.Time           `json:"billDate" gorm:"type:date"`
	Fee            float64             `json:"fee" gorm:"type:decimal(18,4);default:0;comment:加工费用"`
	OutAmount      float64             `json:"outAmount" gorm:"type:decimal(18,4);default:0"`
	InAmount       float64             `json:"inAmount" gorm:"type:decimal(18,4);default:0"`
	Status         string              `json:"status" gorm:"size:20;default:draft;comment:draft草稿 completed已过账 cancelled已撤销"`
	HandlerID      uint                `json:"handlerId" gorm:"index"`
	OperatorID     uint                `json:"operatorId" gorm:"index"`
	Remark         string              `json:"remark" gorm:"size:500"`
	Items          []AssemblyOrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

func (AssemblyOrder) TableName() string { return "assembly_orders" }

// AssemblyOrderItem 组装拆装单明细（direction 区分出库/入库行）
type AssemblyOrderItem struct {
	BaseModel
	OrderID   uint    `json:"orderId" gorm:"index;not null"`
	Direction string  `json:"direction" gorm:"size:5;not null;comment:out出库 in入库"`
	ProductID uint    `json:"productId" gorm:"index;not null"`
	Quantity  float64 `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	Price     float64 `json:"price" gorm:"type:decimal(18,4);default:0"`
	Amount    float64 `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Remark    string  `json:"remark" gorm:"size:255"`
	Product   Product `json:"product" gorm:"foreignKey:ProductID"`
}

func (AssemblyOrderItem) TableName() string { return "assembly_order_items" }

// StockBatch 库存批次
type StockBatch struct {
	BaseModelWithCompany
	ProductID     uint       `json:"productId" gorm:"index;not null"`
	WarehouseID   uint       `json:"warehouseId" gorm:"index;not null"`
	BatchNo       string     `json:"batchNo" gorm:"size:64;index;not null"`
	Quantity      float64    `json:"quantity" gorm:"type:decimal(18,4);default:0"`
	CostPrice     float64    `json:"costPrice" gorm:"type:decimal(18,4);default:0"`
	ProduceDate   *time.Time `json:"produceDate" gorm:"type:date"`
	ExpiryDate    *time.Time `json:"expiryDate" gorm:"type:date"`
	ShelfLifeDays int        `json:"shelfLifeDays" gorm:"default:0"`
	InDate        *time.Time `json:"inDate" gorm:"type:date"`
	SupplierID    uint       `json:"supplierId" gorm:"index;default:0"`
}

func (StockBatch) TableName() string { return "stock_batches" }
