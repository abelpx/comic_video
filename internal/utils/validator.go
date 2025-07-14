package utils

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateErrors 验证错误处理
func ValidateErrors(err error) []ValidationError {
	var errors []ValidationError
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			message := getErrorMessage(e.Tag(), e.Param())
			
			errors = append(errors, ValidationError{
				Field:   field,
				Message: message,
			})
		}
	}
	
	return errors
}

// getErrorMessage 获取错误消息
func getErrorMessage(tag string, param string) string {
	switch tag {
	case "required":
		return "此字段为必填项"
	case "email":
		return "请输入有效的邮箱地址"
	case "min":
		return "长度不能少于 " + param + " 个字符"
	case "max":
		return "长度不能超过 " + param + " 个字符"
	case "numeric":
		return "请输入数字"
	case "alpha":
		return "只能包含字母"
	case "alphanum":
		return "只能包含字母和数字"
	case "url":
		return "请输入有效的URL"
	case "uuid":
		return "请输入有效的UUID格式"
	default:
		return "验证失败"
	}
}

// IsEmpty 检查值是否为空
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	validate := validator.New()
	return validate.Var(email, "email") == nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) (bool, string) {
	if len(password) < 6 {
		return false, "密码长度不能少于6个字符"
	}
	
	if len(password) > 50 {
		return false, "密码长度不能超过50个字符"
	}
	
	// 可以添加更多密码强度验证规则
	// 例如：必须包含大小写字母、数字、特殊字符等
	
	return true, ""
} 