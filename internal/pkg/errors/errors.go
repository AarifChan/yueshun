package errors

import "errors"

// 通用业务错误
var (
	ErrNotFound      = errors.New("记录不存在")
	ErrDuplicate     = errors.New("记录已存在")
	ErrUnauthorized  = errors.New("未授权")
	ErrForbidden     = errors.New("无权限")
	ErrValidation    = errors.New("参数验证失败")
	ErrInternal      = errors.New("内部错误")
	ErrInvalidToken  = errors.New("无效的令牌")
	ErrExpiredToken  = errors.New("令牌已过期")
	ErrPasswordWrong = errors.New("密码错误")
	ErrUserDisabled  = errors.New("用户已禁用")
)

// BizError 业务错误
type BizError struct {
	Code    int
	Message string
	Err     error
}

func (e *BizError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func NewBizError(code int, message string) *BizError {
	return &BizError{Code: code, Message: message}
}

func WrapBizError(code int, message string, err error) *BizError {
	return &BizError{Code: code, Message: message, Err: err}
}
