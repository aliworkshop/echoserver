package echoserver

import "github.com/aliworkshop/handlerlib"

type paginator struct {
	perPage int
	page    int
	sortBy  string
}

func NewPaginator() handlerlib.Pagination {
	return &paginator{}
}

func (p *paginator) Page() int {
	if p.page == 0 {
		p.page = 1
	}
	return p.page
}

func (p *paginator) PerPage() int {
	if p.perPage == 0 {
		p.perPage = 10
	}
	return p.perPage
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
