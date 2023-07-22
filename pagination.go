package echoserver

import "github.com/aliworkshop/gateway/v2"

type paginator struct {
	limit  int
	page   int
	sortBy string
	total  uint64
}

func NewPaginator() gateway.Paginator {
	return &paginator{}
}

func (p *paginator) Page() int {
	if p.page == 0 {
		p.page = 1
	}
	return p.page
}

func (p *paginator) PerPage() int {
	if p.limit == 0 {
		p.limit = 10
	}
	return p.limit
}

func (p *paginator) SetPage(i int) {
	p.page = i
}
func (p *paginator) SetPerPage(i int) {
	p.limit = i
}

func (p *paginator) SortBy() string {
	return p.sortBy
}

func (p *paginator) Total() uint64 {
	return p.total
}

func (p *paginator) SetTotal(total uint64) {
	p.total = total
}
