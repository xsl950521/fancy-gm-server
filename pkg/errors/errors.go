package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 通用错误 1000-1999
	ErrCodeInvalidParam ErrorCode = 1001 // 参数错误
	ErrCodeNotFound     ErrorCode = 1002 // 资源不存在
	ErrCodeInternal     ErrorCode = 1003 // 内部错误
	ErrCodeUnauthorized ErrorCode = 1004 // 未授权
	ErrCodeForbidden    ErrorCode = 1005 // 禁止访问
	ErrCodeConflict     ErrorCode = 1006 // 资源冲突

	// 业务错误 2000-2999
	ErrCodeJobNotFound      ErrorCode = 2001 // 任务不存在
	ErrCodeJobAlreadyExists ErrorCode = 2002 // 任务已存在
	ErrCodeJobProcessing    ErrorCode = 2003 // 任务处理中
	ErrCodeFileNotFound     ErrorCode = 2004 // 文件不存在
	ErrCodeFileTooLarge     ErrorCode = 2005 // 文件过大
	ErrCodeInvalidFileType  ErrorCode = 2006 // 无效文件类型
)

// AppError 应用错误
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Err     error     `json:"-"` // 原始错误，不序列化
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus 返回HTTP状态码
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidParam:
		return http.StatusBadRequest
	case ErrCodeNotFound, ErrCodeJobNotFound, ErrCodeFileNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict, ErrCodeJobAlreadyExists:
		return http.StatusConflict
	case ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// New 创建新错误
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建格式化错误
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap 包装错误
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf 包装错误（格式化）
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// IsAppError 判断是否为AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError 转换为AppError
func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
