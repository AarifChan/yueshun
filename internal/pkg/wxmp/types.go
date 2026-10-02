package wxmp

import (
	"errors"
	"fmt"
)

// Config 微信小程序配置（secret 仅服务端持有）
type Config struct {
	AppID  string
	Secret string
}

var (
	ErrNotConfigured = errors.New("微信小程序未配置")
	ErrInvalidCode   = errors.New("授权码无效或已过期")
)

// APIError 微信 API 返回的业务错误
type APIError struct {
	Code int
	Msg  string
}

func (e *APIError) Error() string { return fmt.Sprintf("wxmp api error %d: %s", e.Code, e.Msg) }
