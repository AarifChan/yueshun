package model

import "time"

// ExpenseBill 费用单
type ExpenseBill struct {
	BaseModelWithCompany
	BillNo      string    `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time `json:"billDate" gorm:"type:date"`
	AccountID   uint      `json:"accountId" gorm:"index;not null"`
	ItemID      uint      `json:"itemId" gorm:"index;not null;comment:费用项目"`
	Amount      float64   `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Counterpart string    `json:"counterpart" gorm:"size:128;comment:往来单位"`
	HandlerID   uint      `json:"handlerId" gorm:"index;comment:经手人"`
	DeptID      uint      `json:"deptId" gorm:"index;default:0"`
	Status      string    `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	Remark      string    `json:"remark" gorm:"size:500"`
}

func (ExpenseBill) TableName() string { return "expense_bills" }

// OtherIncomeBill 其他收入单
type OtherIncomeBill struct {
	BaseModelWithCompany
	BillNo      string    `json:"billNo" gorm:"size:64;index;not null"`
	BillDate    time.Time `json:"billDate" gorm:"type:date"`
	AccountID   uint      `json:"accountId" gorm:"index;not null"`
	ItemID      uint      `json:"itemId" gorm:"index;not null;comment:收入项目"`
	Amount      float64   `json:"amount" gorm:"type:decimal(18,4);default:0"`
	Counterpart string    `json:"counterpart" gorm:"size:128;comment:往来单位"`
	HandlerID   uint      `json:"handlerId" gorm:"index;comment:经手人"`
	DeptID      uint      `json:"deptId" gorm:"index;default:0"`
	Status      string    `json:"status" gorm:"size:20;default:pending;comment:pending待审核 completed已完成"`
	Remark      string    `json:"remark" gorm:"size:500"`
}

func (OtherIncomeBill) TableName() string { return "other_income_bills" }
