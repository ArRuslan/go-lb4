package main

import (
	"net/http"
	"strconv"
)

func getPageAndSize(req *http.Request) (int, int) {
	pageMaybe := req.URL.Query().Get("page")
	page, err := strconv.Atoi(pageMaybe)
	if err != nil {
		page = 1
	}

	pageSizeMaybe := req.URL.Query().Get("pageSize")
	pageSize, err := strconv.Atoi(pageSizeMaybe)
	if err != nil {
		pageSize = 10
	}

	return page, pageSize
}
