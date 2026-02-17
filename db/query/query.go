package query

type Pagination struct {
	Offset int32
	Limit  int32
	// 顺序：asc/desc
	Order  string
	SortBy string
}
