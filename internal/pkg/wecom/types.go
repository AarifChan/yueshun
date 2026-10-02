package wecom

import (
	"errors"
	"fmt"
)

// Config 企业微信自建应用配置（secret 仅服务端持有）
type Config struct {
	CorpID       string
	AgentID      int
	Secret       string
	InviteQRURL  string // 企业邀请二维码图片地址，用于「未加入企业」引导页
	RedirectHost string // Web 扫码回调域名（须为应用可信域名），为空时前端用 location.origin
}

// UserDetail 企业微信成员详情（本系统关心的子集）
type UserDetail struct {
	UserID string
	Name   string
	Mobile string
}

var (
	// ErrNotConfigured corpid/secret 未配置
	ErrNotConfigured = errors.New("企业微信未配置")
	// ErrNotMember 手机号/授权码对应的用户不是企业成员
	ErrNotMember = errors.New("非企业成员")
	// ErrInvalidCode 授权码无效或已过期
	ErrInvalidCode = errors.New("授权码无效或已过期")
)

// APIError 企业微信 API 返回的业务错误（errcode != 0 且无法归类时）
type APIError struct {
	Code int
	Msg  string
}

func (e *APIError) Error() string { return fmt.Sprintf("wecom api error %d: %s", e.Code, e.Msg) }
