package utils

import (
	"html/template"
	"strings"
)

type PaginationInfo struct {
	Page     int
	PageSize int
	Count    int

	UrlPath string
	Query   string
}

type PaginationResult struct {
	PrevDisabled string
	PrevPage     int
	Pages        []int
	NextDisabled string
	NextPage     int

	UrlPath string
	Query   template.URL
}

var TmplPaginationFuncs = template.FuncMap{
	"calculatePagination": func(pagination PaginationInfo) PaginationResult {
		var queryParams []string
		for _, param := range strings.Split(pagination.Query, "&") {
			if strings.HasPrefix(param, "page=") || strings.HasPrefix(param, "pageSize=") {
				continue
			}
			queryParams = append(queryParams, param)
		}

		result := PaginationResult{
			UrlPath: pagination.UrlPath,
			Query:   template.URL(strings.Join(queryParams, "&")),
		}

		totalPages := (pagination.Count + pagination.PageSize - 1) / pagination.PageSize

		if pagination.Page == 1 {
			result.PrevDisabled = "disabled"
			result.PrevPage = pagination.Page
		} else {
			result.PrevPage = pagination.Page - 1
		}

		if pagination.Page == totalPages {
			result.NextDisabled = "disabled"
			result.NextPage = pagination.Page
		} else {
			result.NextPage = pagination.Page + 1
		}

		minPage := max(1, pagination.Page-2)
		maxPage := min(totalPages, pagination.Page+2)

		result.Pages = make([]int, maxPage-minPage+1)
		for i, page := 0, minPage; page <= maxPage; page++ {
			result.Pages[i] = page
			i++
		}

		return result
	},
	"unescape": func(s string) template.HTML {
		return template.HTML(s)
	},
}
