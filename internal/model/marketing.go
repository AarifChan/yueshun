package model

import "time"

// Promotion 营销活动（限时特价/单品买赠/阶梯价/组合促销/优惠套餐/整单优惠）
type Promotion struct {
	BaseModelWithCompany
	Type       string    `json:"type" gorm:"size:20;not null;index;comment:seckill限时特价 gift单品买赠 tiered阶梯价 combo组合促销 package优惠套餐 whole_order整单优惠"`
	Name       string    `json:"name" gorm:"size:128;not null"`
	StartAt    time.Time `json:"startAt" gorm:"not null"`
	EndAt      time.Time `json:"endAt" gorm:"not null"`
	ProductIDs string    `json:"productIds" gorm:"size:2000;comment:参与商品ID,逗号分隔,空=全部商品"`
	GiftQty    float64   `json:"giftQty" gorm:"default:0;comment:活动赠品数"`
	Sort       int       `json:"sort" gorm:"default:0"`
	Status     int8      `json:"status" gorm:"default:1;comment:1启用 0停用"`
	Remark     string    `json:"remark" gorm:"size:500"`
}

func (Promotion) TableName() string { return "promotions" }

// Coupon 优惠券
type Coupon struct {
	BaseModelWithCompany
	Name       string     `json:"name" gorm:"size:128;not null"`
	FaceValue  float64    `json:"faceValue" gorm:"type:decimal(18,2);not null;comment:面值"`
	MinAmount  float64    `json:"minAmount" gorm:"type:decimal(18,2);default:0;comment:使用门槛(满多少可用)"`
	TotalQty   int        `json:"totalQty" gorm:"default:0;comment:发放总量,0=不限"`
	ValidStart *time.Time `json:"validStart" gorm:"type:date"`
	ValidEnd   *time.Time `json:"validEnd" gorm:"type:date"`
	Status     int8       `json:"status" gorm:"default:1;comment:1启用 0停用"`
	Sort       int        `json:"sort" gorm:"default:0"`
	Remark     string     `json:"remark" gorm:"size:500"`
}

func (Coupon) TableName() string { return "coupons" }

// CouponGrant 优惠券领用记录
type CouponGrant struct {
	BaseModelWithCompany
	CouponID   uint       `json:"couponId" gorm:"index;not null"`
	CustomerID uint       `json:"customerId" gorm:"index;not null"`
	Status     string     `json:"status" gorm:"size:20;default:unused;comment:unused未使用 used已使用"`
	ReceivedAt time.Time  `json:"receivedAt"`
	UsedAt     *time.Time `json:"usedAt"`
	RefBillNo  string     `json:"refBillNo" gorm:"size:64;comment:使用单号"`
}

func (CouponGrant) TableName() string { return "coupon_grants" }

// PointRule 积分赠送规则
type PointRule struct {
	BaseModelWithCompany
	Condition     string `json:"condition" gorm:"size:20;not null;comment:first_login首次登录 daily_login每日登录 order下单赠送"`
	Points        int    `json:"points" gorm:"default:0;comment:单笔赠送积分"`
	ProductScope  string `json:"productScope" gorm:"size:20;default:all;comment:all全部 part部分商品"`
	CustomerScope string `json:"customerScope" gorm:"size:20;default:all;comment:all全部客户 part部分客户"`
	CategoryID    uint   `json:"categoryId" gorm:"index;default:0"`
	BrandID       uint   `json:"brandId" gorm:"index;default:0"`
	SpecDate      string `json:"specDate" gorm:"size:20;comment:指定日期"`
	RuleText      string `json:"ruleText" gorm:"size:255;comment:赠送规则说明"`
	Enabled       bool   `json:"enabled" gorm:"default:true"`
	Sort          int    `json:"sort" gorm:"default:0"`
}

func (PointRule) TableName() string { return "point_rules" }

// PointFlow 积分流水
type PointFlow struct {
	BaseModelWithCompany
	CustomerID uint   `json:"customerId" gorm:"index;not null"`
	Change     int    `json:"change" gorm:"not null;comment:正为增加负为扣减"`
	Type       string `json:"type" gorm:"size:20;not null;comment:grant赠送 deduct抵扣 exchange兑换 clear清零 manual手工调整"`
	Remark     string `json:"remark" gorm:"size:255"`
	RefID      uint   `json:"refId" gorm:"index;default:0"`
}

func (PointFlow) TableName() string { return "point_flows" }

// PointExchange 积分商品兑换设置
type PointExchange struct {
	BaseModelWithCompany
	ProductID uint `json:"productId" gorm:"index;not null"`
	Points    int  `json:"points" gorm:"not null;comment:兑换所需积分"`
	Enabled   bool `json:"enabled" gorm:"default:true"`
	Sort      int  `json:"sort" gorm:"default:0"`
}

func (PointExchange) TableName() string { return "point_exchanges" }

// Distributor 分销商
type Distributor struct {
	BaseModelWithCompany
	Name       string     `json:"name" gorm:"size:128;not null"`
	Phone      string     `json:"phone" gorm:"size:32"`
	ManagerID  uint       `json:"managerId" gorm:"index;default:0;comment:业务经理"`
	CustomerID uint       `json:"customerId" gorm:"index;default:0;comment:关联客户"`
	Status     string     `json:"status" gorm:"size:20;default:pending;comment:pending待审核 active已通过 rejected已拒绝 disabled已停用"`
	AppliedAt  time.Time  `json:"appliedAt"`
	ApprovedAt *time.Time `json:"approvedAt"`
	Remark     string     `json:"remark" gorm:"size:500"`
}

func (Distributor) TableName() string { return "distributors" }

// DistributorWithdrawal 佣金提现
type DistributorWithdrawal struct {
	BaseModelWithCompany
	DistributorID uint       `json:"distributorId" gorm:"index;not null"`
	Amount        float64    `json:"amount" gorm:"type:decimal(18,2);not null"`
	Status        string     `json:"status" gorm:"size:20;default:pending;comment:pending提现中 paid已提现 rejected已拒绝"`
	AppliedAt     time.Time  `json:"appliedAt"`
	PaidAt        *time.Time `json:"paidAt"`
	Remark        string     `json:"remark" gorm:"size:500"`
}

func (DistributorWithdrawal) TableName() string { return "distributor_withdrawals" }

// StoredValueRule 储值规则
type StoredValueRule struct {
	BaseModelWithCompany
	Name       string  `json:"name" gorm:"size:128;not null"`
	Amount     float64 `json:"amount" gorm:"type:decimal(18,2);not null;comment:储值金额"`
	GiftAmount float64 `json:"giftAmount" gorm:"type:decimal(18,2);default:0;comment:赠送金额"`
	Status     int8    `json:"status" gorm:"default:1"`
	Sort       int     `json:"sort" gorm:"default:0"`
	Remark     string  `json:"remark" gorm:"size:500"`
}

func (StoredValueRule) TableName() string { return "stored_value_rules" }

// StoredValueRecord 储值记录
type StoredValueRecord struct {
	BaseModelWithCompany
	CustomerID uint    `json:"customerId" gorm:"index;not null"`
	RuleID     uint    `json:"ruleId" gorm:"index;default:0"`
	Amount     float64 `json:"amount" gorm:"type:decimal(18,2);not null"`
	GiftAmount float64 `json:"giftAmount" gorm:"type:decimal(18,2);default:0"`
	AccountID  uint    `json:"accountId" gorm:"index;default:0;comment:收款账户"`
	OperatorID uint    `json:"operatorId" gorm:"index;default:0"`
	Remark     string  `json:"remark" gorm:"size:500"`
}

func (StoredValueRecord) TableName() string { return "stored_value_records" }
