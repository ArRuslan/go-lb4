package main

import (
	"html/template"
)

type PaginationInfo struct {
	Page     int
	PageSize int
	Count    int

	urlPath string
}

type PaginationResult struct {
	PrevDisabled string
	PrevPage     int
	Pages        []int
	NextDisabled string
	NextPage     int

	UrlPath string
}

var tmplPaginationFuncs = template.FuncMap{
	"calculatePagination": func(pagination PaginationInfo) PaginationResult {
		result := PaginationResult{
			UrlPath: pagination.urlPath,
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
}
