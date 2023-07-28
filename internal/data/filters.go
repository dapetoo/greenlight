package data

import (
	"github.com/dapetoo/greenlight/internal/validator"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	//Check that the page and page_size parameters are valid
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	//Check that the sort parameter matches a value in the safelist
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

// Check that the provided Sort field matches one of the entries in the SafeList and extract the column name from the sort
// field by stripping the leading hypen character(if one exists)
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// Return the sort direction(ASC or DESC) depending on the prefix character of the sort field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
