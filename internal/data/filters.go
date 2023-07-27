package data

import "github.com/dapetoo/greenlight/internal/validator"

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
