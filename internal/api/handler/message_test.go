package handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
)

func setupMessageRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	v1 := r.Group("/api/v1")
	authorized := v1.Group("")
	authorized.Use(middleware.JWTMiddleware())
	NewMessageHandler(db).RegisterRoutes(authorized)
	return r
}

func setupMessageDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(&model.Message{}, &model.MessageUserSetting{}, &model.CompanySetting{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedMessage(t *testing.T, db *gorm.DB, companyID, userID uint, category, msgType, title string) model.Message {
	t.Helper()
	m := model.Message{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
		UserID:               userID,
		Category:             category, Type: msgType, Title: title,
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMessageListFilterAndUnread(t *testing.T) {
	db := setupMessageDB(t)
	company, _, _, _ := seedTenant(t, db)
	r := setupMessageRouter(db)
	header := authHeader(t, company.ID, 42)

	seedMessage(t, db, company.ID, 42, "approval", "pending", "待审批单")
	seedMessage(t, db, company.ID, 42, "approval", "cc", "抄送你的审批")
	seedMessage(t, db, company.ID, 42, "system", "notice", "系统提示")
	// 其他用户的消息不应出现
	seedMessage(t, db, company.ID, 7, "approval", "pending", "别人的消息")

	// 分类过滤
	w := performRequest(r, "GET", "/api/v1/messages?category=approval", nil, header)
	resp := parseBody(t, w)
	if resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 2 {
		t.Fatalf("expected 2 approval messages, got %v", data["total"])
	}

	// 筛选项过滤
	w = performRequest(r, "GET", "/api/v1/messages?category=approval&type=cc", nil, header)
	resp = parseBody(t, w)
	data = resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Fatalf("expected 1 cc message, got %v", data["total"])
	}

	// 未读数
	w = performRequest(r, "GET", "/api/v1/messages/unread-count", nil, header)
	resp = parseBody(t, w)
	data = resp["data"].(map[string]interface{})
	if data["total"].(float64) != 3 {
		t.Fatalf("expected 3 unread, got %v", data["total"])
	}
	byCategory := data["byCategory"].(map[string]interface{})
	if byCategory["approval"].(float64) != 2 || byCategory["system"].(float64) != 1 {
		t.Fatalf("unexpected byCategory %v", byCategory)
	}

	// 未知分类 → 4000
	w = performRequest(r, "GET", "/api/v1/messages?category=nope", nil, header)
	if resp := parseBody(t, w); resp["code"].(float64) != 4000 {
		t.Fatalf("expected 4000, got %v", resp)
	}
}

func TestMessageReadOneAndReadAll(t *testing.T) {
	db := setupMessageDB(t)
	company, _, _, _ := seedTenant(t, db)
	r := setupMessageRouter(db)
	header := authHeader(t, company.ID, 42)

	m1 := seedMessage(t, db, company.ID, 42, "approval", "pending", "待审批单")
	seedMessage(t, db, company.ID, 42, "approval", "cc", "抄送你的审批")
	seedMessage(t, db, company.ID, 42, "system", "notice", "系统提示")

	// 单条已读
	w := performRequest(r, "PUT", fmt.Sprintf("/api/v1/messages/%d/read", m1.ID), nil, header)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
	var reloaded model.Message
	db.First(&reloaded, m1.ID)
	if reloaded.ReadAt == nil {
		t.Fatal("expected read_at set")
	}

	// 全部已读（当前分类）→ 只剩 system 未读
	w = performRequest(r, "PUT", "/api/v1/messages/read-all?category=approval", nil, header)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
	w = performRequest(r, "GET", "/api/v1/messages/unread-count", nil, header)
	data := parseBody(t, w)["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Fatalf("expected 1 unread after category read-all, got %v", data["total"])
	}

	// 一键已读 → 全部已读
	w = performRequest(r, "PUT", "/api/v1/messages/read-all", nil, header)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
	w = performRequest(r, "GET", "/api/v1/messages/unread-count", nil, header)
	data = parseBody(t, w)["data"].(map[string]interface{})
	if data["total"].(float64) != 0 {
		t.Fatalf("expected 0 unread after read-all, got %v", data["total"])
	}
}

func TestVoiceSettingDefaultAndSave(t *testing.T) {
	db := setupMessageDB(t)
	company, _, _, _ := seedTenant(t, db)
	r := setupMessageRouter(db)
	header := authHeader(t, company.ID, 42)

	// 文档默认值：新订单语音提醒开、新审批语音提醒关
	w := performRequest(r, "GET", "/api/v1/message-settings/voice", nil, header)
	data := parseBody(t, w)["data"].(map[string]interface{})
	if data["newOrderVoice"].(bool) != true || data["newApprovalVoice"].(bool) != false {
		t.Fatalf("unexpected defaults %v", data)
	}

	// 部分更新：只关新订单
	w = performRequest(r, "PUT", "/api/v1/message-settings/voice", map[string]interface{}{
		"newOrderVoice": false,
	}, header)
	data = parseBody(t, w)["data"].(map[string]interface{})
	if data["newOrderVoice"].(bool) != false || data["newApprovalVoice"].(bool) != false {
		t.Fatalf("unexpected after save %v", data)
	}

	// 再查应持久化
	w = performRequest(r, "GET", "/api/v1/message-settings/voice", nil, header)
	data = parseBody(t, w)["data"].(map[string]interface{})
	if data["newOrderVoice"].(bool) != false {
		t.Fatalf("expected persisted, got %v", data)
	}

	// 另一用户互不影响，仍是默认值
	w = performRequest(r, "GET", "/api/v1/message-settings/voice", nil, authHeader(t, company.ID, 7))
	data = parseBody(t, w)["data"].(map[string]interface{})
	if data["newOrderVoice"].(bool) != true {
		t.Fatalf("expected default for other user, got %v", data)
	}
}

func TestExpiryWarningSettingDefaultAndSave(t *testing.T) {
	db := setupMessageDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	emp := seedStaff(t, db, company.ID, dept.ID, staffRole.ID, "13800000020")
	r := setupMessageRouter(db)
	header := authHeader(t, company.ID, 42)

	// 默认：两类通知关、发送时间 8 时
	w := performRequest(r, "GET", "/api/v1/message-settings/expiry-warning", nil, header)
	data := parseBody(t, w)["data"].(map[string]interface{})
	if data["warnEnabled"].(bool) != false || data["expiredEnabled"].(bool) != false {
		t.Fatalf("unexpected defaults %v", data)
	}
	if data["sendHour"].(float64) != 8 {
		t.Fatalf("expected default hour 8, got %v", data["sendHour"])
	}

	// 保存：含一个不存在的接收人应被过滤
	w = performRequest(r, "PUT", "/api/v1/message-settings/expiry-warning", map[string]interface{}{
		"warnEnabled":      true,
		"warnReceivers":    []uint{emp.ID, 99999},
		"expiredEnabled":   true,
		"expiredReceivers": []uint{emp.ID},
		"sendHour":         9,
	}, header)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}

	w = performRequest(r, "GET", "/api/v1/message-settings/expiry-warning", nil, header)
	data = parseBody(t, w)["data"].(map[string]interface{})
	if data["warnEnabled"].(bool) != true || data["expiredEnabled"].(bool) != true {
		t.Fatalf("expected enabled, got %v", data)
	}
	if data["sendHour"].(float64) != 9 {
		t.Fatalf("expected hour 9, got %v", data["sendHour"])
	}
	receivers := data["warnReceivers"].([]interface{})
	if len(receivers) != 1 || receivers[0].(float64) != float64(emp.ID) {
		t.Fatalf("expected receivers filtered to [%d], got %v", emp.ID, receivers)
	}

	// 发送时间非法 → 4000
	w = performRequest(r, "PUT", "/api/v1/message-settings/expiry-warning", map[string]interface{}{
		"sendHour": 24,
	}, header)
	if resp := parseBody(t, w); resp["code"].(float64) != 4000 {
		t.Fatalf("expected 4000, got %v", resp)
	}
}

func TestMessageReadAtKeepsOrdering(t *testing.T) {
	db := setupMessageDB(t)
	company, _, _, _ := seedTenant(t, db)
	r := setupMessageRouter(db)
	header := authHeader(t, company.ID, 42)

	now := time.Now()
	m := model.Message{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		UserID:               42, Category: "print", Type: "print", Title: "打印完成",
		ReadAt:               &now,
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatal(err)
	}

	w := performRequest(r, "GET", "/api/v1/messages?category=print", nil, header)
	data := parseBody(t, w)["data"].(map[string]interface{})
	list := data["list"].([]interface{})
	if len(list) != 1 {
		t.Fatalf("expected 1 message, got %v", len(list))
	}
	row := list[0].(map[string]interface{})
	if row["read"].(bool) != true {
		t.Fatalf("expected read=true, got %v", row)
	}
}
