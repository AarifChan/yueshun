package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

func setupAuthRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	v1 := r.Group("/api/v1")
	NewAuthHandler(db).RegisterRoutes(v1)
	return r
}

func TestLoginPasswordGate(t *testing.T) {
	db := setupTestDB(t)
	company, dept, superRole, staffRole := seedTenant(t, db)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	// 超管
	admin := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		DeptID:               dept.ID,
		RoleID:               superRole.ID,
		Username:             "admin",
		Password:             string(hashed),
		Name:                 "管理员",
		Phone:                "13800000001",
		Status:               1,
	}
	// 普通员工
	staff := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		DeptID:               dept.ID,
		RoleID:               staffRole.ID,
		Username:             "staff",
		Password:             string(hashed),
		Name:                 "小张",
		Phone:                "13800000002",
		Status:               1,
	}
	db.Create(&admin)
	db.Create(&staff)

	r := setupAuthRouter(db)

	// 超管密码登录 → 放行
	w := performRequest(r, "POST", "/api/v1/auth/login", map[string]string{"username": "admin", "password": "123456"}, nil)
	resp := parseBody(t, w)
	if resp["code"].(float64) != 200 {
		t.Fatalf("super admin login should succeed, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["accessToken"] == "" || data["refreshToken"] == "" {
		t.Fatal("expected token pair")
	}
	if data["accessExpiresIn"].(float64) != 7200 {
		t.Fatalf("expected accessExpiresIn 7200, got %v", data["accessExpiresIn"])
	}

	// 普通员工密码登录 → 4101
	w = performRequest(r, "POST", "/api/v1/auth/login", map[string]string{"username": "staff", "password": "123456"}, nil)
	resp = parseBody(t, w)
	if resp["code"].(float64) != response.CodeNeedWecom {
		t.Fatalf("expected 4101, got %v", resp)
	}
}
