package web

import (
	"net/http"
	"net/url"
	"strconv"
)

type ListPagination struct {
	Page, Pages, Total int
	Path               string
	Query              url.Values
}

func requestedPage(r *http.Request) int {
	n, _ := strconv.Atoi(r.URL.Query().Get("page"))
	return max(1, n)
}
func listPagination(r *http.Request, total, size int) ListPagination {
	pages := max(1, (total+size-1)/size)
	return ListPagination{Page: min(requestedPage(r), pages), Pages: pages, Total: total, Path: r.URL.Path, Query: r.URL.Query()}
}
func (p ListPagination) PageURL(n int) string {
	q := url.Values{}
	for k, v := range p.Query {
		q[k] = append([]string(nil), v...)
	}
	q.Del("page")
	if n > 1 {
		q.Set("page", strconv.Itoa(n))
	}
	if len(q) == 0 {
		return p.Path
	}
	return p.Path + "?" + q.Encode()
}
func (p ListPagination) Items() []PaginationItem { return paginationItems(p.Page, p.Pages, p.PageURL) }
func (p ListPagination) PreviousURL() string {
	if p.Page > 1 {
		return p.PageURL(p.Page - 1)
	}
	return ""
}
func (p ListPagination) NextURL() string {
	if p.Page < p.Pages {
		return p.PageURL(p.Page + 1)
	}
	return ""
}
