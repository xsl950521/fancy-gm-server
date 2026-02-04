package validators

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidateStruct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		struct_ interface{}
		wantErr bool
	}{
		{
			name: "valid struct",
			struct_: struct {
				Name string `validate:"required"`
			}{Name: "test"},
			wantErr: false,
		},
		{
			name: "invalid struct - missing required field",
			struct_: struct {
				Name string `validate:"required"`
			}{},
			wantErr: true,
		},
		{
			name: "valid struct with min",
			struct_: struct {
				Age int `validate:"min=18"`
			}{Age: 20},
			wantErr: false,
		},
		{
			name: "invalid struct - below min",
			struct_: struct {
				Age int `validate:"min=18"`
			}{Age: 15},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			result := ValidateStruct(c, tt.struct_)
			if tt.wantErr {
				assert.False(t, result, "Expected validation to fail")
				assert.Equal(t, 400, w.Code)
			} else {
				assert.True(t, result, "Expected validation to pass")
			}
		})
	}
}

func TestGetErrorMessage(t *testing.T) {
	// 这个测试需要创建一个validator.FieldError
	// 由于validator.FieldError是接口，我们通过实际验证来测试
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Email string `validate:"required,email"`
		Name  string `validate:"required,min=3"`
		Age   int    `validate:"min=18"`
	}

	// 测试required错误
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	invalidStruct := TestStruct{}
	result := ValidateStruct(c, invalidStruct)
	assert.False(t, result, "Should fail validation")
	assert.Equal(t, 400, w.Code)

	// 测试email错误
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	invalidEmail := TestStruct{
		Email: "invalid-email",
		Name:  "Test",
		Age:   20,
	}
	result = ValidateStruct(c, invalidEmail)
	assert.False(t, result, "Should fail email validation")

	// 测试min错误
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	invalidMin := TestStruct{
		Email: "test@example.com",
		Name:  "AB", // 小于min=3
		Age:   15,   // 小于min=18
	}
	result = ValidateStruct(c, invalidMin)
	assert.False(t, result, "Should fail min validation")
}
