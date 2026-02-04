package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"redis_data/pkg/errors"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PageResponse 分页响应
type PageResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（带消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// SuccessPage 分页成功响应
func SuccessPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	if appErr, ok := errors.AsAppError(err); ok {
		c.JSON(appErr.HTTPStatus(), Response{
			Code:    int(appErr.Code),
			Message: appErr.Message,
			Error:   appErr.Error(),
		})
		return
	}

	// 未知错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:    int(errors.ErrCodeInternal),
		Message: "内部服务器错误",
		Error:   err.Error(),
	})
}

// ErrorWithCode 错误响应（指定错误码）
func ErrorWithCode(c *gin.Context, code errors.ErrorCode, message string) {
	appErr := errors.New(code, message)
	c.JSON(appErr.HTTPStatus(), Response{
		Code:    int(appErr.Code),
		Message: appErr.Message,
		Error:   appErr.Error(),
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	ErrorWithCode(c, errors.ErrCodeInvalidParam, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	ErrorWithCode(c, errors.ErrCodeNotFound, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	ErrorWithCode(c, errors.ErrCodeUnauthorized, message)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	ErrorWithCode(c, errors.ErrCodeForbidden, message)
}

// InternalError 500错误
func InternalError(c *gin.Context, err error) {
	Error(c, errors.Wrap(err, errors.ErrCodeInternal, "内部服务器错误"))
}
