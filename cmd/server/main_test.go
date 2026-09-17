package main

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
)

// 验证 autoMigrate 后：employees 有 wecom_user_id 列，且 (company_id, phone) 复合唯一索引生效
func TestAutoMigrateEmployee(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:wecom-migrate-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := autoMigrate(db); err != nil {
		t.Fatal(err)
	}

	// 密码可空：不传密码应能建员工
	emp := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: 1},
		DeptID:               1,
		RoleID:               1,
		Username:             "u1",
		Name:                 "甲",
		Phone:                "13800000001",
		Status:               1,
	}
	if err := db.Create(&emp).Error; err != nil {
		t.Fatalf("create without password should succeed: %v", err)
	}

	// 同公司同手机号 → 唯一索引冲突
	dup := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: 1},
		DeptID:               1,
		RoleID:               1,
		Username:             "u2",
		Name:                 "乙",
		Phone:                "13800000001",
		Status:               1,
	}
	if err := db.Create(&dup).Error; err == nil {
		t.Fatal("expected unique index violation on (company_id, phone), got nil")
	}

	// 不同公司同手机号 → 允许
	other := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: 2},
		DeptID:               1,
		RoleID:               1,
		Username:             "u3",
		Name:                 "丙",
		Phone:                "13800000001",
		Status:               1,
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("same phone in another company should be allowed: %v", err)
	}

	// wecom_user_id 可写入
	if err := db.Model(&emp).Update("wecom_user_id", "zhangsan").Error; err != nil {
		t.Fatal(err)
	}
}
