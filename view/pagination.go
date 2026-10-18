package view

import (
	"net/url"
	"strconv"
)

type PaginationItem struct {
	Number            int
	URL               string
	Current, Ellipsis bool
}

func paginationItems(page, pages int, pageURL func(int) string) []PaginationItem {
	var result []PaginationItem
	var previous int
	for number := 1; number <= pages; number++ {
		if number != 1 && number != pages && (number < page-1 || number > page+1) {
			continue
		}
		if number-previous == 2 {
			skipped := previous + 1
			result = append(result, PaginationItem{Number: skipped, URL: pageURL(skipped)})
		} else if number-previous > 2 {
			result = append(result, PaginationItem{Ellipsis: true})
		}
		result = append(result, PaginationItem{Number: number, URL: pageURL(number), Current: number == page})
		previous = number
	}
	return result
}

const ExamsPageSize = 20

// Zero marks an ellipsis between the current neighbourhood and an end page.
func MediaPageNumbers(page, pages int) []int {
	var numbers []int
	for number := 1; number <= pages; number++ {
		if pages <= 7 || number == 1 || number == pages || (number >= page-1 && number <= page+1) || (page <= 4 && number < page) || (page >= pages-3 && number > page) {
			if len(numbers) > 0 && number-numbers[len(numbers)-1] > 1 {
				numbers = append(numbers, 0)
			}
			numbers = append(numbers, number)
		}
	}
	return numbers
}

type ListPagination struct {
	Page, Pages, Total int
	Path               string
	Query              url.Values
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
