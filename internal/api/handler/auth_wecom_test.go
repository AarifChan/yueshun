package handler

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
	"zhizhang-server/internal/pkg/wecom"
	"zhizhang-server/internal/pkg/wxmp"
)

// --- mock 客户端 ---

type fakeWecom struct {
	byMobile map[string]string // mobile → userid（不存在则 ErrNotMember）
	byCode   map[string]string // auth code → userid
	details  map[string]*wecom.UserDetail
}

func (f *fakeWecom) GetUserIDByMobile(mobile string) (string, error) {
	if id, ok := f.byMobile[mobile]; ok {
		return id, nil
	}
	return "", wecom.ErrNotMember
}
func (f *fakeWecom) GetUserInfoByCode(code string) (string, error) {
	if id, ok := f.byCode[code]; ok {
		return id, nil
	}
	return "", wecom.ErrNotMember
}
func (f *fakeWecom) GetUserDetail(userID string) (*wecom.UserDetail, error) {
	if d, ok := f.details[userID]; ok {
		return d, nil
	}
	return nil, wecom.ErrNotMember
}

type fakeWxmp struct {
	byCode map[string]string // phone code → 手机号
}

func (f *fakeWxmp) GetPhoneNumber(code string) (string, error) {
	if p, ok := f.byCode[code]; ok {
		return p, nil
	}
	return "", wxmp.ErrInvalidCode
}

// --- 装配 ---

func setupWecomRouter(db *gorm.DB, wc wecom.Client, wx wxmp.Client) *gin.Engine {
	r := gin.New()
	v1 := r.Group("/api/v1")
	cfg := &wecom.Config{CorpID: "ww-test", AgentID: 1000002, InviteQRURL: "https://example.com/qr.png"}
	NewWecomAuthHandler(db, wc, wx, cfg).RegisterRoutes(v1)
	return r
}

// --- 用例 ---

func TestWecomConfig(t *testing.T) {
	db := setupTestDB(t)
	r := setupWecomRouter(db, &fakeWecom{}, &fakeWxmp{})
	w := performRequest(r, "GET", "/api/v1/auth/wecom/config", nil, nil)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]interface{})
	if data["corpid"] != "ww-test" || data["inviteQrUrl"] == "" {
		t.Fatalf("unexpected config resp: %v", resp)
	}
}

func TestMpLoginFirstBind(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	emp := seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000001")
	wc := &fakeWecom{byMobile: map[string]string{"13800000001": "wx_zhangsan"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13800000001"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	resp := parseBody(t, w)
	if resp["code"].(float64) != 200 {
		t.Fatalf("mp login should succeed, got %v", resp)
	}
	// 首次登录回写 wecom_user_id
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.WecomUserID != "wx_zhangsan" {
		t.Fatalf("expected wecom_user_id bound, got %q", reloaded.WecomUserID)
	}
}

func TestMpLoginBoundUserSkipsPhoneMatch(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	emp := seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000001")
	db.Model(&emp).Update("wecom_user_id", "wx_zhangsan")
	// 成员换了手机号：企业微信侧手机号变了，按 wecom_user_id 直接命中并更新 phone
	wc := &fakeWecom{byMobile: map[string]string{"13900000009": "wx_zhangsan"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13900000009"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("bound user should login by wecom_user_id, got %v", resp)
	}
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.Phone != "13900000009" {
		t.Fatalf("phone should sync from wecom, got %q", reloaded.Phone)
	}
}

func TestMpLoginNotMember(t *testing.T) {
	db := setupTestDB(t)
	wc := &fakeWecom{byMobile: map[string]string{}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "19900000000"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	resp := parseBody(t, w)
	if resp["code"].(float64) != response.CodeNotWecomMember {
		t.Fatalf("expected 4102, got %v", resp)
	}
	if resp["data"].(map[string]interface{})["inviteQrUrl"] == "" {
		t.Fatal("4102 should carry inviteQrUrl")
	}
}

func TestMpLoginNotRegistered(t *testing.T) {
	db := setupTestDB(t)
	seedTenant(t, db) // 有租户但无该手机号的员工
	wc := &fakeWecom{byMobile: map[string]string{"13800000003": "wx_lisi"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13800000003"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeEmpNotRegistered {
		t.Fatalf("expected 4103, got %v", resp)
	}
}

func TestMpLoginDisabled(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	emp := seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000004")
	db.Model(&emp).Update("status", 0)
	wc := &fakeWecom{byMobile: map[string]string{"13800000004": "wx_wangwu"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13800000004"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeForbidden {
		t.Fatalf("expected 403, got %v", resp)
	}
}

func TestMpLoginInvalidCode(t *testing.T) {
	db := setupTestDB(t)
	r := setupWecomRouter(db, &fakeWecom{}, &fakeWxmp{byCode: map[string]string{}})
	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "expired"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeBadRequest {
		t.Fatalf("expected 400, got %v", resp)
	}
}

func TestWebLoginByUserID(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	emp := seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000005")
	db.Model(&emp).Update("wecom_user_id", "wx_zhaoliu")
	wc := &fakeWecom{
		byCode:  map[string]string{"web-code": "wx_zhaoliu"},
		details: map[string]*wecom.UserDetail{"wx_zhaoliu": {UserID: "wx_zhaoliu", Name: "赵六", Mobile: "13800000005"}},
	}
	r := setupWecomRouter(db, wc, &fakeWxmp{})

	w := performRequest(r, "POST", "/api/v1/auth/wecom/web", map[string]string{"code": "web-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("web login should succeed, got %v", resp)
	}
}

func TestWebLoginFallbackBindByMobile(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	emp := seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000006")
	wc := &fakeWecom{
		byCode:  map[string]string{"web-code": "wx_qiqi"},
		details: map[string]*wecom.UserDetail{"wx_qiqi": {UserID: "wx_qiqi", Name: "七七", Mobile: "13800000006"}},
	}
	r := setupWecomRouter(db, wc, &fakeWxmp{})

	w := performRequest(r, "POST", "/api/v1/auth/wecom/web", map[string]string{"code": "web-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("web login fallback should succeed, got %v", resp)
	}
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.WecomUserID != "wx_qiqi" {
		t.Fatalf("expected wecom_user_id bound, got %q", reloaded.WecomUserID)
	}
}

func TestWebLoginExternalUser(t *testing.T) {
	db := setupTestDB(t)
	wc := &fakeWecom{byCode: map[string]string{}} // 任何 code 都 ErrNotMember
	r := setupWecomRouter(db, wc, &fakeWxmp{})
	w := performRequest(r, "POST", "/api/v1/auth/wecom/web", map[string]string{"code": "ext"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeNotWecomMember {
		t.Fatalf("expected 4102, got %v", resp)
	}
}

// 未配置时 config 端点返回 4001
func TestWecomConfigNotConfigured(t *testing.T) {
	db := setupTestDB(t)
	r := gin.New()
	v1 := r.Group("/api/v1")
	NewWecomAuthHandler(db, &fakeWecom{}, &fakeWxmp{}, &wecom.Config{}).RegisterRoutes(v1)
	w := performRequest(r, "GET", "/api/v1/auth/wecom/config", nil, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeBizError {
		t.Fatalf("expected 4001, got %v", resp)
	}
}

var _ = errors.Is // 保留给实现侧 stdlib errors 使用说明：注意与 internal/pkg/errors 区分
