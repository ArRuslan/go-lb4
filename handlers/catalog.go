package handlers

import (
	"database/sql"
	"errors"
	"go-lb4/db"
	"go-lb4/utils"
	"html/template"
	"log"
	"net/http"
	"time"
)

type CatalogTmplContext struct {
	Products       []db.Product
	Category       db.Category
	Query          string
	CartItemsCount int

	Pagination utils.PaginationInfo
	ThisUrl    string
}

func ProductCatalogHandler(w http.ResponseWriter, r *http.Request) {
	cartId := utils.GetCartId(r)
	http.SetCookie(w, &http.Cookie{Name: "cartId", Value: cartId.String(), Path: "/", HttpOnly: true, MaxAge: 86400})
	cart, err := db.GetOrCreateCart(cartId)
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	cart.LastAccessTime = time.Now()
	if utils.ReturnOnDatabaseError(cart.DbSave(), w) {
		return
	}

	cartCount, err := db.GetCartProductsCount(cart.Id)

	allGood := true

	var category db.Category

	query := utils.GetFormString(r, "query", nil, &allGood, nil)
	category.Id = utils.GetFormInt64(r, "category_id", nil, &allGood, nil)

	category, err = db.GetCategory(category.Id)
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
		Products:       products,
		Category:       category,
		Query:          query,
		CartItemsCount: cartCount,
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
