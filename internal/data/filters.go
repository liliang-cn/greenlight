package data

import (
	"strings"

	"github.com/liliang-cn/greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// ValidateFilters 校验过滤条件
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// sortColumn 检查客户端提供的排序条件是否在允许排序的列表中
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection 返回排序正向或者反向
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

// limit 返回分页大小
func (f Filters) limit() int {
	return f.PageSize
}

// offset 返回页数
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}
