package model

import "time"

// SupplierCategory 供应商分类
type SupplierCategory struct {
	BaseModelWithCompany
	ParentID uint   `json:"parentId" gorm:"index;default:0"`
	Name     string `json:"name" gorm:"size:64;not null"`
	Code     string `json:"code" gorm:"size:64"`
	Sort     int    `json:"sort" gorm:"default:0"`
	Status   int8   `json:"status" gorm:"default:1"`
}

func (SupplierCategory) TableName() string { return "supplier_categories" }

// WorkReport 工作汇报（日志/周报/月报）
type WorkReport struct {
	BaseModelWithCompany
	Type       string    `json:"type" gorm:"size:10;not null;comment:log日志 week周报 month月报"`
	ReportDate time.Time `json:"reportDate" gorm:"type:date"`
	Content    string    `json:"content" gorm:"type:text"`
	AuthorID   uint      `json:"authorId" gorm:"index;not null"`
	CcIDs      string    `json:"ccIds" gorm:"size:1000;comment:抄送人ID,逗号分隔"`
	TemplateID uint      `json:"templateId" gorm:"index;default:0"`
}

func (WorkReport) TableName() string { return "work_reports" }

// ReportTemplate 汇报模板
type ReportTemplate struct {
	BaseModelWithCompany
	Category    string `json:"category" gorm:"size:10;not null;comment:log日志 week周报 month月报"`
	Name        string `json:"name" gorm:"size:128;not null"`
	Description string `json:"description" gorm:"size:500"`
	UserIDs     string `json:"userIds" gorm:"size:1000;comment:可使用人ID,逗号分隔,空=全部"`
	Status      int8   `json:"status" gorm:"default:1;comment:1启用 0停用"`
	Sort        int    `json:"sort" gorm:"default:0"`
}

func (ReportTemplate) TableName() string { return "report_templates" }

// OpportunityTemplate 商机设置模板
type OpportunityTemplate struct {
	BaseModelWithCompany
	Name        string `json:"name" gorm:"size:128;not null"`
	Stages      string `json:"stages" gorm:"size:1000;comment:商机阶段,逗号分隔"`
	UserIDs     string `json:"userIds" gorm:"size:1000;comment:可使用人ID,逗号分隔,空=全部"`
	Description string `json:"description" gorm:"size:500"`
	Status      int8   `json:"status" gorm:"default:1"`
	Sort        int    `json:"sort" gorm:"default:0"`
}

func (OpportunityTemplate) TableName() string { return "opportunity_templates" }

// CustomFieldDef 客户自定义字段
type CustomFieldDef struct {
	BaseModelWithCompany
	Name         string `json:"name" gorm:"size:64;not null"`
	FieldType    string `json:"fieldType" gorm:"size:20;not null;default:text;comment:text文本 number数字 date日期 select下拉"`
	Options      string `json:"options" gorm:"size:500;comment:下拉选项,逗号分隔"`
	Enabled      bool   `json:"enabled" gorm:"default:true"`
	Required     bool   `json:"required" gorm:"default:false"`
	DefaultValue string `json:"defaultValue" gorm:"size:255"`
	Sort         int    `json:"sort" gorm:"default:0"`
}

func (CustomFieldDef) TableName() string { return "custom_field_defs" }

// SalesPlan 年度销售计划（业务驾驶舱）
type SalesPlan struct {
	BaseModelWithCompany
	Year       int     `json:"year" gorm:"index;not null;uniqueIndex:uniq_sales_plan,priority:2"`
	EmployeeID uint    `json:"employeeId" gorm:"index;not null;uniqueIndex:uniq_sales_plan,priority:3"`
	M1         float64 `json:"m1" gorm:"type:decimal(18,4);default:0"`
	M2         float64 `json:"m2" gorm:"type:decimal(18,4);default:0"`
	M3         float64 `json:"m3" gorm:"type:decimal(18,4);default:0"`
	M4         float64 `json:"m4" gorm:"type:decimal(18,4);default:0"`
	M5         float64 `json:"m5" gorm:"type:decimal(18,4);default:0"`
	M6         float64 `json:"m6" gorm:"type:decimal(18,4);default:0"`
	M7         float64 `json:"m7" gorm:"type:decimal(18,4);default:0"`
	M8         float64 `json:"m8" gorm:"type:decimal(18,4);default:0"`
	M9         float64 `json:"m9" gorm:"type:decimal(18,4);default:0"`
	M10        float64 `json:"m10" gorm:"type:decimal(18,4);default:0"`
	M11        float64 `json:"m11" gorm:"type:decimal(18,4);default:0"`
	M12        float64 `json:"m12" gorm:"type:decimal(18,4);default:0"`
}

func (SalesPlan) TableName() string { return "sales_plans" }
