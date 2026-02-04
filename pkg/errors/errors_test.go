package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(ErrCodeInvalidParam, "参数错误")
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeInvalidParam, err.Code)
	assert.Equal(t, "参数错误", err.Message)
	assert.Nil(t, err.Err)
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("原始错误")
	err := Wrap(originalErr, ErrCodeInternal, "包装错误")
	
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeInternal, err.Code)
	assert.Equal(t, "包装错误", err.Message)
	assert.Equal(t, originalErr, err.Err)
}

func TestAsAppError(t *testing.T) {
	appErr := New(ErrCodeNotFound, "未找到")
	
	// 测试是AppError的情况
	result, ok := AsAppError(appErr)
	assert.True(t, ok)
	assert.Equal(t, appErr, result)

	// 测试不是AppError的情况
	normalErr := errors.New("普通错误")
	result, ok = AsAppError(normalErr)
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestAppErrorHTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeInvalidParam, 400},
		{ErrCodeNotFound, 404},
		{ErrCodeUnauthorized, 401},
		{ErrCodeForbidden, 403},
		{ErrCodeInternal, 500},
		{ErrorCode(9999), 500}, // 未知错误码
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			err := New(tt.code, "test")
			// 通过AppError的HTTPStatus方法测试
			status := err.HTTPStatus()
			assert.Equal(t, tt.expected, status)
		})
	}
}
