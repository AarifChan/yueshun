package handler

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

func setupEmpRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	v1 := r.Group("/api/v1")
	authorized := v1.Group("")
	authorized.Use(middleware.JWTMiddleware())
	NewEmployeeHandler(db).RegisterRoutes(authorized)
	return r
}

func authHeader(t *testing.T, companyID, userID uint) map[string]string {
	t.Helper()
	token, err := middleware.GenerateToken(userID, "admin", 1, 1, companyID)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]string{"Authorization": "Bearer " + token}
}

func TestCreateEmployeePhoneRequired(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	r := setupEmpRouter(db)

	// 缺手机号 → 400
	w := performRequest(r, "POST", "/api/v1/employees", map[string]interface{}{
		"deptId": dept.ID, "roleId": staffRole.ID, "username": "nophone", "name": "无号", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeBadRequest {
		t.Fatalf("expected 400 for missing phone, got %v", resp)
	}

	// 不带密码 + 带手机号 → 200
	w = performRequest(r, "POST", "/api/v1/employees", map[string]interface{}{
		"deptId": dept.ID, "roleId": staffRole.ID, "username": "withphone", "name": "有号",
		"phone": "13800000010", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
}

func TestCreateEmployeeDuplicatePhone(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000011")
	r := setupEmpRouter(db)

	w := performRequest(r, "POST", "/api/v1/employees", map[string]interface{}{
		"deptId": dept.ID, "roleId": staffRole.ID, "username": "another", "name": "重复",
		"phone": "13800000011", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeDuplicate {
		t.Fatalf("expected 4002, got %v", resp)
	}
}

func TestUpdateEmployeeUnbindWecom(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	emp := seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000012")
	db.Model(&emp).Update("wecom_user_id", "wx_abc")
	r := setupEmpRouter(db)

	// 显式传空串 → 解绑
	w := performRequest(r, "PUT", fmt.Sprintf("/api/v1/employees/%d", emp.ID), map[string]interface{}{
		"wecomUserId": "", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.WecomUserID != "" {
		t.Fatalf("expected unbound, got %q", reloaded.WecomUserID)
	}
}
