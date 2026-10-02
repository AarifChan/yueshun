package model

import "time"

// CommissionPlan 销售提成方案
type CommissionPlan struct {
	BaseModelWithCompany
	Name        string     `json:"name" gorm:"size:128;not null"`
	StartDate   time.Time  `json:"startDate" gorm:"type:date"`
	EndDate     *time.Time `json:"endDate" gorm:"type:date"`
	EmployeeIDs string     `json:"employeeIds" gorm:"size:1000;comment:逗号分隔职员ID,空=全部职员"`
	CalcType    string     `json:"calcType" gorm:"size:32;not null;default:amount_percent;comment:amount_percent按销售金额比例 profit_percent按销售毛利比例 qty_fixed按销售数量定额"`
	Rate        float64    `json:"rate" gorm:"type:decimal(18,4);default:0;comment:比例(%)或每单位定额"`
	Status      int8       `json:"status" gorm:"default:1;comment:1启用 0停用"`
	Remark      string     `json:"remark" gorm:"size:500"`
}

func (CommissionPlan) TableName() string { return "commission_plans" }
