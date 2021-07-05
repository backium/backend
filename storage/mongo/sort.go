package mongo

import "github.com/backium/backend/core"

func sortOrder(o core.SortOrder) int {
	switch o {
	case core.SortAscending:
		return 1
	case core.SortDescending:
		return -1
	default:
		return 1
	}
}
