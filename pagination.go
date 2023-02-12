package echoserver

import "github.com/aliworkshop/handlerlib"

type paginator struct {
	perPage int
	page    int
	sortBy  string
}

func (p *paginator) PerPage() int {
	return p.perPage
}

func NewPaginator() handlerlib.Pagination {
	return &paginator{}
}

func (p *paginator) Page() int {
	return p.page
}
func (p *paginator) SetPage(i int) {
	p.page = i
}
func (p *paginator) SetPerPage(i int) {
	p.perPage = i
}

func (p *paginator) SortBy() string {
	return p.sortBy
}
