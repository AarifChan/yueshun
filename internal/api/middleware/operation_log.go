package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
)

// OperationLogger 记录写操作（POST/PUT/DELETE/PATCH）到操作日志
func OperationLogger(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			c.Next()
			return
		}
		c.Next()

		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/v1/") {
			return
		}
		// 跳过高频/无意义路径
		if strings.Contains(path, "/ping") || strings.Contains(path, "/upload") {
			return
		}
		companyID := GetCompanyID(c)
		if companyID == 0 {
			return
		}
		objectType := objectTypeOf(path)
		detail := method + " " + path
		if c.Request.URL.RawQuery != "" {
			detail += "?" + c.Request.URL.RawQuery
		}
		if len(detail) > 480 {
			detail = detail[:480]
		}
		entry := model.OperationLog{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			UserID:               GetUserID(c),
			ObjectType:           objectType,
			Action:               method,
			Source:               "web",
			IP:                   c.ClientIP(),
			Detail:               detail,
		}
		// 异步失败不影响主流程
		go func() { _ = db.Create(&entry).Error }()
	}
}

// objectTypeOf 从路径提取操作对象（模块名）
func objectTypeOf(path string) string {
	p := strings.TrimPrefix(path, "/api/v1/")
	segs := strings.Split(p, "/")
	if len(segs) == 0 || segs[0] == "" {
		return "其他"
	}
	if name, ok := objectTypeNames[segs[0]]; ok {
		return name
	}
	return segs[0]
}

var objectTypeNames = map[string]string{
	"products": "商品", "customers": "客户", "suppliers": "供应商", "warehouses": "仓库",
	"purchase-orders": "采购订单", "purchase-in-stocks": "采购入库单", "purchase-returns": "采购退货单",
	"sales-orders": "销售订单", "sales-out-stocks": "销售出库单", "sales-returns": "销售退货单",
	"employees": "职员", "departments": "部门", "roles": "角色", "settings": "系统设置",
	"promotions": "营销活动", "coupons": "优惠券", "distributors": "分销商",
	"expense-bills": "费用单", "other-income-bills": "其他收入单",
	"inventory-checks": "盘点单", "inventory-transfers": "调拨单",
	"other-in-stocks": "其他入库单", "other-out-stocks": "其他出库单",
	"mall-orders": "商城订单", "accounts": "现金银行账户",
	"printers": "打印机", "subject-categories": "专题分类", "product-relations": "关联商品",
	"employee-permissions": "职员权限", "company-info": "企业信息", "operation-logs": "操作日志",
}
