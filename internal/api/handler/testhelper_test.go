package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
)

func init() {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT(&middleware.JWTConfig{Secret: "test-secret", AccessTTL: 7200, RefreshTTL: 604800})
}

// setupTestDB 每个测试用独立的内存库名，避免 cache=shared 串数据
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:testdb-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&model.Company{}, &model.Department{}, &model.Role{}, &model.Employee{},
	); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_emp_phone_company ON employees (company_id, phone)").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

// seedTenant 建 公司/部门/两个角色，返回（公司, 部门, 超管角色, 普通角色）
func seedTenant(t *testing.T, db *gorm.DB) (model.Company, model.Department, model.Role, model.Role) {
	t.Helper()
	company := model.Company{Name: "测试公司", Code: "T"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatal(err)
	}
	dept := model.Department{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		Name:                 "总部", Code: "HQ",
	}
	if err := db.Create(&dept).Error; err != nil {
		t.Fatal(err)
	}
	superRole := model.Role{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		Name:                 "超级管理员", Code: "super_admin",
	}
	staffRole := model.Role{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		Name:                 "销售员", Code: "staff",
	}
	if err := db.Create(&superRole).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&staffRole).Error; err != nil {
		t.Fatal(err)
	}
	return company, dept, superRole, staffRole
}

// seedStaff 在指定租户下建员工档案
func seedStaff(t *testing.T, db *gorm.DB, companyID, deptID, roleID uint, phone string) model.Employee {
	t.Helper()
	emp := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		DeptID:               deptID,
		RoleID:               roleID,
		Username:             "staff-" + phone,
		Name:                 "员工" + phone[len(phone)-4:],
		Phone:                phone,
		Status:               1,
	}
	if err := db.Create(&emp).Error; err != nil {
		t.Fatal(err)
	}
	return emp
}

// performRequest 发 JSON 请求并返回 recorder
func performRequest(r http.Handler, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)
	return w
}

// parseBody 解析统一响应信封
func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response failed: %v, body=%s", err, w.Body.String())
	}
	return m
}
