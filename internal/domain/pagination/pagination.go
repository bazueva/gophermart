package pagination

const (
	perPage20 = 20
	perPage50 = 50
)

var perPagesAllow = map[int64]struct{}{
	perPage20: {},
	perPage50: {},
}

type Pagination struct {
	page       int64
	perPage    int64
	totalCount int64
}

func (p *Pagination) TotalCount() int64 {
	return p.totalCount
}

func (p *Pagination) setPage(page int64) {
	if page <= 0 {
		p.page = 1

		return
	}

	p.page = page
}

func (p *Pagination) setPerPage(perPage int64) {
	if _, ok := perPagesAllow[perPage]; ok {
		p.perPage = perPage

		return
	}

	p.perPage = perPage20
}

func (p *Pagination) SetTotalCount(totalCount int64) {
	p.totalCount = totalCount
}

func (p *Pagination) GetPage() int64 {
	if p.totalCount == 0 {
		return 1
	}

	totalPages := p.GetTotalPages()

	if p.page > totalPages {
		return totalPages
	}

	return p.page
}

func (p *Pagination) GetPerPage() int64 {
	return p.perPage
}

func (p *Pagination) GetOffset() int64 {
	return (p.GetPage() - 1) * p.perPage
}

func (p *Pagination) GetTotalPages() int64 {
	if p.totalCount == 0 {
		return 0
	}

	totalPages := p.totalCount / p.perPage
	if p.totalCount%p.perPage != 0 {
		totalPages++
	}

	return totalPages
}

func NewPagination(
	page int64,
	perPage int64,
) *Pagination {
	pagin := &Pagination{}
	pagin.setPage(page)
	pagin.setPerPage(perPage)

	return pagin
}
