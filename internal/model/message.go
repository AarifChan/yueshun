package model

import "time"

// Message 站内消息（顶部消息中心）
type Message struct {
	BaseModelWithCompany
	UserID   uint       `json:"userId" gorm:"index;not null;comment:接收人"`
	Category string     `json:"category" gorm:"size:20;index;not null;comment:消息分类"`
	Type     string     `json:"type" gorm:"size:40;comment:分类下的筛选类型"`
	Title    string     `json:"title" gorm:"size:200;not null"`
	Content  string     `json:"content" gorm:"size:1000"`
	ReadAt   *time.Time `json:"readAt"`
}

func (Message) TableName() string { return "messages" }

// MessageUserSetting 消息中心个人设置（语音提醒）
// 注意：不加 gorm default tag——bool 零值 false 会被 default 覆盖，导致存不进「关」。
// 默认值（新订单开、新审批关）由 handler 在无记录时返回。
type MessageUserSetting struct {
	BaseModelWithCompany
	UserID           uint `json:"userId" gorm:"not null;uniqueIndex:uniq_msg_user_setting,priority:2"`
	NewOrderVoice    bool `json:"newOrderVoice" gorm:"comment:新订单语音提醒"`
	NewApprovalVoice bool `json:"newApprovalVoice" gorm:"comment:新审批语音提醒"`
}

func (MessageUserSetting) TableName() string { return "message_user_settings" }
