package validator

import "regexp"

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Validator 类型中存放校验错误
type Validator struct {
	Errors map[string]string
}

// New 构造函数，返回新的 Validator 实例
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Valid 函数在 errors 为空时返回 true
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError map 中新增一条错误信息
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check 在校验未通过时增加一条错误消息
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Matches 在传入值匹配正则的时候返回 true

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique 在传入的列表没有重复值时返回 true
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// In 当值在指定的列表中时返回 true
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}

	return false
}
