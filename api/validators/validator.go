package validators

import (
	"redis_data/pkg/errors"
	"redis_data/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct 验证结构体并返回友好的错误信息
func ValidateStruct(c *gin.Context, s interface{}) bool {
	if err := validate.Struct(s); err != nil {
		// 提取验证错误
		var errMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, getErrorMessage(err))
		}

		// 返回统一的错误响应
		appErr := errors.New(
			errors.ErrCodeInvalidParam,
			"参数验证失败: "+errMsgs[0],
		)
		response.BadRequest(c, appErr.Message)
		return false
	}
	return true
}

// getErrorMessage 获取友好的错误信息
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return field + " 是必填字段"
	case "email":
		return field + " 必须是有效的邮箱地址"
	case "min":
		return field + " 的最小值为 " + err.Param()
	case "max":
		return field + " 的最大值为 " + err.Param()
	case "len":
		return field + " 的长度必须为 " + err.Param()
	case "oneof":
		return field + " 必须是以下值之一: " + err.Param()
	case "url":
		return field + " 必须是有效的URL"
	case "numeric":
		return field + " 必须是数字"
	case "alphanum":
		return field + " 只能包含字母和数字"
	default:
		return field + " 验证失败"
	}
}

// ValidateQuery 验证查询参数
func ValidateQuery(c *gin.Context, s interface{}) bool {
	if err := c.ShouldBindQuery(s); err != nil {
		appErr := errors.New(
			errors.ErrCodeInvalidParam,
			"查询参数错误: "+err.Error(),
		)
		response.BadRequest(c, appErr.Message)
		return false
	}
	return ValidateStruct(c, s)
}

// ValidateJSON 验证JSON请求体
func ValidateJSON(c *gin.Context, s interface{}) bool {
	if err := c.ShouldBindJSON(s); err != nil {
		appErr := errors.New(
			errors.ErrCodeInvalidParam,
			"请求体格式错误: "+err.Error(),
		)
		response.BadRequest(c, appErr.Message)
		return false
	}
	return ValidateStruct(c, s)
}

// ValidateURI 验证URI参数
func ValidateURI(c *gin.Context, s interface{}) bool {
	if err := c.ShouldBindUri(s); err != nil {
		appErr := errors.New(
			errors.ErrCodeInvalidParam,
			"URI参数错误: "+err.Error(),
		)
		response.BadRequest(c, appErr.Message)
		return false
	}
	return ValidateStruct(c, s)
}
