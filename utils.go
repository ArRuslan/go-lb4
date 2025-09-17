package main

import (
	"fmt"
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

func getFormStringNonEmpty(req *http.Request, name string, errorText *string, valid *bool, out *string) string {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	if value == "" {
		*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		*valid = false

		return value
	}

	return value
}

func getFormString(req *http.Request, name string, _ *string, _ *bool, out *string) string {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	return value
}

func getFormInt(req *http.Request, name string, errorText *string, valid *bool, out *string) int {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		*valid = false

		return 0
	}

	return valueInt
}

func getFormInt64(req *http.Request, name string, errorText *string, valid *bool, out *string) int64 {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		*valid = false

		return 0
	}

	return valueInt
}

func getFormDouble(req *http.Request, name string, errorText *string, valid *bool, out *string) float64 {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	valueDouble, err := strconv.ParseFloat(value, 64)
	if err != nil {
		*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		*valid = false

		return 0
	}

	return valueDouble
}
