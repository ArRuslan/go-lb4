package handlers

import (
	"database/sql"
	"errors"
	"go-lb4/db"
	"go-lb4/utils"
	"html/template"
	"log"
	"net/http"
)

type CatalogTmplContext struct {
	Products []db.Product
	Category db.Category
	Query    string

	Pagination utils.PaginationInfo
	ThisUrl    string
}

func ProductCatalogHandler(w http.ResponseWriter, r *http.Request) {
	allGood := true

	var category db.Category

	query := utils.GetFormString(r, "query", nil, &allGood, nil)
	category.Id = utils.GetFormInt64(r, "category_id", nil, &allGood, nil)

	category, err := db.GetCategory(category.Id)
	if errors.Is(err, sql.ErrNoRows) {
		category = db.Category{}
	} else if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	page, pageSize := utils.GetPageAndSize(r)
	products, count, err := db.SearchProductsCatalog(page, pageSize, category, query)
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	tmpl := template.New("catalog.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/catalog.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Funcs(utils.TmplPaginationFuncs).Execute(w, CatalogTmplContext{
		Products: products,
		Category: category,
		Query:    query,
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/catalog",
			Query:    r.URL.RawQuery,
		},
		ThisUrl: r.RequestURI,
	})
	if err != nil {
		log.Println(err)
	}
}
