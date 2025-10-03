package handlers

import (
	"go-lb4/db"
	"go-lb4/utils"
	"html/template"
	"log"
	"net/http"
)

type CatalogTmplContext struct {
	Products []db.Product

	Pagination utils.PaginationInfo
}

func ProductCatalogHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.GetPageAndSize(r)
	products, count, err := db.GetProducts(page, pageSize)
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
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/products",
		},
	})
	if err != nil {
		log.Println(err)
	}
}
