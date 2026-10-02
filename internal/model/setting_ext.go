package model

// OperationLog 操作日志
type OperationLog struct {
	BaseModelWithCompany
	UserID     uint   `json:"userId" gorm:"index;default:0"`
	ObjectType string `json:"objectType" gorm:"size:64;index;comment:操作对象(模块)"`
	Action     string `json:"action" gorm:"size:10;comment:POST/PUT/DELETE"`
	Source     string `json:"source" gorm:"size:20;default:web;comment:操作来源"`
	IP         string `json:"ip" gorm:"size:64"`
	Detail     string `json:"detail" gorm:"size:500"`
}

func (OperationLog) TableName() string { return "operation_logs" }

// SubjectCategory 商城专题分类
type SubjectCategory struct {
	BaseModelWithCompany
	Name         string `json:"name" gorm:"size:64;not null"`
	Image        string `json:"image" gorm:"size:255"`
	VisibleScope string `json:"visibleScope" gorm:"size:20;default:all;comment:all全部客户 part部分客户"`
	ShowInMall   bool   `json:"showInMall" gorm:"default:true;comment:显示到小程序商城"`
	ProductIDs   string `json:"productIds" gorm:"size:2000;comment:绑定商品ID,逗号分隔"`
	Sort         int    `json:"sort" gorm:"default:0"`
	Status       int8   `json:"status" gorm:"default:1"`
}

func (SubjectCategory) TableName() string { return "subject_categories" }

// Printer 打印机设置
type Printer struct {
	BaseModelWithCompany
	Name     string `json:"name" gorm:"size:64;not null"`
	Type     string `json:"type" gorm:"size:20;default:remote;comment:remote远程打印 cloud_box云盒子 cloud云打印机"`
	DeviceNo string `json:"deviceNo" gorm:"size:64"`
	IsDefault bool  `json:"isDefault" gorm:"default:false"`
	Status   int8   `json:"status" gorm:"default:1"`
	Remark   string `json:"remark" gorm:"size:255"`
}

func (Printer) TableName() string { return "printers" }

// ProductRelation 商城关联商品（推荐搭配）
type ProductRelation struct {
	BaseModelWithCompany
	ProductID        uint `json:"productId" gorm:"index;not null;uniqueIndex:uniq_product_relation,priority:2"`
	RelatedProductID uint `json:"relatedProductId" gorm:"index;not null;uniqueIndex:uniq_product_relation,priority:3"`
	Sort             int  `json:"sort" gorm:"default:0"`
}

func (ProductRelation) TableName() string { return "product_relations" }

// EmployeePermission 职员权限（管辖范围）
type EmployeePermission struct {
	BaseModelWithCompany
	EmployeeID     uint   `json:"employeeId" gorm:"index;not null;uniqueIndex:uniq_emp_perm,priority:2"`
	AppScopes      string `json:"appScopes" gorm:"size:255;comment:已授权应用,逗号分隔"`
	CustomerScope  string `json:"customerScope" gorm:"size:20;default:all;comment:all全部 self仅自己 dept本部门"`
	EmployeeScope  string `json:"employeeScope" gorm:"size:20;default:all"`
	WarehouseScope string `json:"warehouseScope" gorm:"size:255;comment:仓库管辖范围,逗号分隔,空=全部"`
	SupplierScope  string `json:"supplierScope" gorm:"size:20;default:all"`
	ProductScope   string `json:"productScope" gorm:"size:20;default:all"`
	AccountScope   string `json:"accountScope" gorm:"size:255;comment:现金银行账户使用范围,逗号分隔,空=全部"`
	LoginTimeRange string `json:"loginTimeRange" gorm:"size:32;comment:允许登录时间,如 08:00-22:00"`
	LicenseType    string `json:"licenseType" gorm:"size:20;default:normal;comment:normal普通 wecom企微许可"`
}

func (EmployeePermission) TableName() string { return "employee_permissions" }
