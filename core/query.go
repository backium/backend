package core

type DateFilter struct {
	Gte int64
	Lte int64
}

type SortOrder string

const (
	SortNone       SortOrder = ""
	SortAscending  SortOrder = "asc"
	SortDescending SortOrder = "desc"
)
