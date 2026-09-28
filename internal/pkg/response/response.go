package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess      = 200
	CodeError        = 500
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeConflict     = 409

	// 业务错误码段 4000-4999
	CodeValidationError = 4000
	CodeBizError        = 4001
	CodeDuplicate       = 4002
	CodeNotExist        = 4003
	CodeNoPermission    = 4004
)

// Success 成功响应
func Success(data interface{}) Response {
	return Response{Code: CodeSuccess, Message: "success", Data: data}
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(message string, data interface{}) Response {
	return Response{Code: CodeSuccess, Message: message, Data: data}
}

// Error 错误响应
func Error(code int, message string) Response {
	return Response{Code: code, Message: message}
}

// PageResult 分页响应
func PageResult(list interface{}, total int64, page, pageSize int) Response {
	// nil slice 序列化为 null，前端组件（如 el-table）要求数组，统一转为空数组
	if list == nil {
		list = []interface{}{}
	}
	return Success(map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// --- Gin 快捷方法 ---

// JSON 返回 JSON 响应
func JSON(c *gin.Context, code int, resp Response) {
	c.JSON(code, resp)
}

// Ok 返回成功
func Ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Success(data))
}

// OkWithMessage 返回成功（自定义消息）
func OkWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessWithMessage(message, data))
}

// Fail 返回错误（HTTP 200，业务错误码）
func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Error(code, message))
}

// BadRequest 返回 400
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Error(CodeBadRequest, message))
}

// Unauthorized 返回 401
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Error(CodeUnauthorized, message))
}

// Forbidden 返回 403
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Error(CodeForbidden, message))
}

// NotFound 返回 404
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Error(CodeNotFound, message))
}

// ServerError 返回 500
func ServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Error(CodeError, message))
}

// Page 分页快捷方法
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, PageResult(list, total, page, pageSize))
}

// OkWithPage 成功响应加分页
func OkWithPage(c *gin.Context, list interface{}, page, pageSize, total int) {
	c.JSON(http.StatusOK, PageResult(list, int64(total), page, pageSize))
}
