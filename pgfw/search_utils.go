package pgfw

import (
	"slices"

	"github.com/doug-martin/goqu/v9"
	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis/models/utility/golang/dmutil"
)

// TODO: rename to ColSortKey
type ColSortOption struct {
	ColName    string
	SortOption *dmutil.SortKey
}

func GenOrderBy(options []ColSortOption) []interface{} {
	ext.FilterIn(&options, func(opt ColSortOption) bool {
		return opt.SortOption != nil
	})
	slices.SortFunc(options, func(opt1 ColSortOption, opt2 ColSortOption) int {
		if opt1.SortOption.Index < opt2.SortOption.Index {
			return -1
		}
		if opt1.SortOption.Index == opt2.SortOption.Index {
			return 0
		}
		return 1
	})
	var orderBy []interface{}
	for _, opt := range options {
		col := goqu.L(opt.ColName)
		if opt.SortOption.Order == dmutil.SO_Asc {
			orderBy = append(orderBy, col.Asc())
		} else {
			orderBy = append(orderBy, col.Desc())
		}
	}
	return orderBy
}
