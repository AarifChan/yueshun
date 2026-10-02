package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// MessageHandler 顶部消息中心（站内消息/语音提醒设置/近效期预警消息设置）
type MessageHandler struct {
	db *gorm.DB
}

func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

func (h *MessageHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/messages", h.MessageList)
	r.GET("/messages/unread-count", h.UnreadCount)
	r.PUT("/messages/read-all", h.ReadAll)
	r.PUT("/messages/:id/read", h.ReadOne)

	ms := r.Group("/message-settings")
	{
		ms.GET("/voice", h.VoiceSettingGet)
		ms.PUT("/voice", h.VoiceSettingSave)
		ms.GET("/expiry-warning", h.ExpiryWarningSettingGet)
		ms.PUT("/expiry-warning", h.ExpiryWarningSettingSave)
	}
}

// messageCategories 消息中心分类及各自筛选项（与 docs/功能文档/13-企业微信通知.md 一致，
// 各分类第一个筛选项为「全部」（单据消息为「全部订单」），不在此列出，仅列具体类型）
var messageCategories = map[string][]string{
	"bill":      {"order", "return_payment"},
	"customer":  {"new_audit", "contact_overdue", "trade_overdue", "orderer_audit"},
	"approval":  {"mine", "pending", "cc", "mention"},
	"report":    {"cc", "remind", "comment"},
	"system":    {"notice", "announcement"},
	"eco":       {"connection", "mall", "product", "bill"},
	"biz":       {"biz_report", "stock_alarm", "fund"},
	"print":     {"print"},
	"warehouse": {"shelving"},
}

func validMessageCategory(category string) bool {
	_, ok := messageCategories[category]
	return ok
}

// ==================== 站内消息 ====================

// MessageList 当前用户的消息列表，可按分类与筛选项过滤
func (h *MessageHandler) MessageList(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	category := c.Query("category")
	msgType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := h.db.Model(&model.Message{}).Where("company_id = ? AND user_id = ?", companyID, userID)
	if category != "" {
		if !validMessageCategory(category) {
			response.Fail(c, 4000, "未知消息分类")
			return
		}
		q = q.Where("category = ?", category)
	}
	if msgType != "" {
		q = q.Where("type = ?", msgType)
	}
	var total int64
	q.Count(&total)
	var list []model.Message
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	rows := make([]gin.H, 0, len(list))
	for _, x := range list {
		rows = append(rows, gin.H{
			"id": x.ID, "category": x.Category, "type": x.Type,
			"title": x.Title, "content": x.Content,
			"read": x.ReadAt != nil, "createdAt": x.CreatedAt,
		})
	}
	response.OkWithPage(c, rows, page, pageSize, int(total))
}

// UnreadCount 当前用户未读数（按分类分组）
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	type row struct {
		Category string
		Cnt      int64
	}
	var rows []row
	h.db.Model(&model.Message{}).
		Select("category, COUNT(*) AS cnt").
		Where("company_id = ? AND user_id = ? AND read_at IS NULL", companyID, userID).
		Group("category").Scan(&rows)
	byCategory := map[string]int64{}
	var total int64
	for _, r := range rows {
		byCategory[r.Category] = r.Cnt
		total += r.Cnt
	}
	response.Ok(c, gin.H{"total": total, "byCategory": byCategory})
}

// ReadOne 标记单条已读
func (h *MessageHandler) ReadOne(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Model(&model.Message{}).
		Where("id = ? AND company_id = ? AND user_id = ? AND read_at IS NULL", id, companyID, userID).
		Update("read_at", time.Now()).Error; err != nil {
		response.Fail(c, 4000, "操作失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

// ReadAll 全部已读（带 category 时仅该分类）/一键已读（不带 category 时全部分类）
func (h *MessageHandler) ReadAll(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	category := c.Query("category")
	q := h.db.Model(&model.Message{}).
		Where("company_id = ? AND user_id = ? AND read_at IS NULL", companyID, userID)
	if category != "" {
		if !validMessageCategory(category) {
			response.Fail(c, 4000, "未知消息分类")
			return
		}
		q = q.Where("category = ?", category)
	}
	if err := q.Update("read_at", time.Now()).Error; err != nil {
		response.Fail(c, 4000, "操作失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}

// ==================== 语音提醒设置 ====================

// voiceSetting 读取个人语音提醒设置；未保存过时返回文档默认值（新订单开、新审批关）
func (h *MessageHandler) voiceSetting(companyID, userID uint) model.MessageUserSetting {
	var s model.MessageUserSetting
	if err := h.db.Where("company_id = ? AND user_id = ?", companyID, userID).First(&s).Error; err != nil {
		return model.MessageUserSetting{NewOrderVoice: true, NewApprovalVoice: false}
	}
	return s
}

func (h *MessageHandler) VoiceSettingGet(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	s := h.voiceSetting(companyID, userID)
	response.Ok(c, gin.H{
		"newOrderVoice":    s.NewOrderVoice,
		"newApprovalVoice": s.NewApprovalVoice,
	})
}

func (h *MessageHandler) VoiceSettingSave(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	var req struct {
		NewOrderVoice    *bool `json:"newOrderVoice"`
		NewApprovalVoice *bool `json:"newApprovalVoice"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	s := h.voiceSetting(companyID, userID)
	s.CompanyID = companyID
	s.UserID = userID
	if req.NewOrderVoice != nil {
		s.NewOrderVoice = *req.NewOrderVoice
	}
	if req.NewApprovalVoice != nil {
		s.NewApprovalVoice = *req.NewApprovalVoice
	}
	var err error
	if s.ID == 0 {
		err = h.db.Create(&s).Error
	} else {
		err = h.db.Save(&s).Error
	}
	if err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, gin.H{
		"newOrderVoice":    s.NewOrderVoice,
		"newApprovalVoice": s.NewApprovalVoice,
	})
}

// ==================== 商品近效期预警-消息设置 ====================

// 存公司级键值设置（company_settings 表）
const (
	expiryKeyWarnEnabled      = "expiry.warn.enabled"
	expiryKeyWarnReceivers    = "expiry.warn.receivers"
	expiryKeyExpiredEnabled   = "expiry.expired.enabled"
	expiryKeyExpiredReceivers = "expiry.expired.receivers"
	expiryKeySendHour         = "expiry.sendHour"
)

func boolSetting(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func joinIDList(ids []uint) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, strconv.Itoa(int(id)))
		}
	}
	return strings.Join(parts, ",")
}

func upsertCompanySetting(tx *gorm.DB, companyID uint, key, value string) error {
	var s model.CompanySetting
	err := tx.Where("company_id = ? AND key = ?", companyID, key).First(&s).Error
	if err != nil {
		return tx.Create(&model.CompanySetting{
			BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
			Key:                  key, Value: value,
		}).Error
	}
	return tx.Model(&s).Update("value", value).Error
}

// ExpiryWarningSettingGet 近效期预警消息设置（管理员和系统管理员默认收到，无需配置）
func (h *MessageHandler) ExpiryWarningSettingGet(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var settings []model.CompanySetting
	h.db.Where("company_id = ? AND key LIKE ?", companyID, "expiry.%").Find(&settings)
	m := map[string]string{}
	for _, s := range settings {
		m[s.Key] = s.Value
	}
	hour := 8
	if v, err := strconv.Atoi(m[expiryKeySendHour]); err == nil && v >= 0 && v <= 23 {
		hour = v
	}
	response.Ok(c, gin.H{
		"warnEnabled":      m[expiryKeyWarnEnabled] == "1",
		"warnReceivers":    parseIDList(m[expiryKeyWarnReceivers]),
		"expiredEnabled":   m[expiryKeyExpiredEnabled] == "1",
		"expiredReceivers": parseIDList(m[expiryKeyExpiredReceivers]),
		"sendHour":         hour,
	})
}

func (h *MessageHandler) ExpiryWarningSettingSave(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		WarnEnabled      bool   `json:"warnEnabled"`
		WarnReceivers    []uint `json:"warnReceivers"`
		ExpiredEnabled   bool   `json:"expiredEnabled"`
		ExpiredReceivers []uint `json:"expiredReceivers"`
		SendHour         int    `json:"sendHour"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 4000, "参数错误")
		return
	}
	if req.SendHour < 0 || req.SendHour > 23 {
		response.Fail(c, 4000, "发送时间需为 0-23 时")
		return
	}

	// 接收人必须是本企业职员
	validIDs := map[uint]bool{}
	var emps []model.Employee
	h.db.Where("company_id = ?", companyID).Find(&emps)
	for _, e := range emps {
		validIDs[e.ID] = true
	}
	filterReceivers := func(ids []uint) []uint {
		out := make([]uint, 0, len(ids))
		for _, id := range ids {
			if validIDs[id] {
				out = append(out, id)
			}
		}
		return out
	}

	kv := map[string]string{
		expiryKeyWarnEnabled:      boolSetting(req.WarnEnabled),
		expiryKeyWarnReceivers:    joinIDList(filterReceivers(req.WarnReceivers)),
		expiryKeyExpiredEnabled:   boolSetting(req.ExpiredEnabled),
		expiryKeyExpiredReceivers: joinIDList(filterReceivers(req.ExpiredReceivers)),
		expiryKeySendHour:         strconv.Itoa(req.SendHour),
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		for k, v := range kv {
			if err := upsertCompanySetting(tx, companyID, k, v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Fail(c, 4000, "保存失败: "+err.Error())
		return
	}
	response.Ok(c, nil)
}
