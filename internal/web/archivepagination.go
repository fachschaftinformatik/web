package web

import (
	"net/url"
	"strconv"
)

const archivePageSize = 20

func (p *ArchivePage) paginate(raw string) {
	p.Total = len(p.Modules)
	p.Pages = max(1, (p.Total+archivePageSize-1)/archivePageSize)
	p.Page, _ = strconv.Atoi(raw)
	if p.Page < 1 || p.Page > p.Pages {
		p.Page = 1
	}
	start := (p.Page - 1) * archivePageSize
	p.Modules = p.Modules[start:min(start+archivePageSize, p.Total)]
}

func (p ArchivePage) PageURL(page int) string {
	q, _ := url.ParseQuery(p.querySuffix()[1:])
	if page > 1 {
		q.Set("seite", strconv.Itoa(page))
	}
	return "/archive?" + q.Encode()
}

func (p ArchivePage) Pagination() []PaginationItem {
	return paginationItems(p.Page, p.Pages, p.PageURL)
}

func (p ArchivePage) PreviousURL() string {
	if p.Page > 1 {
		return p.PageURL(p.Page - 1)
	}
	return ""
}

func (p ArchivePage) NextURL() string {
	if p.Page < p.Pages {
		return p.PageURL(p.Page + 1)
	}
	return ""
}
