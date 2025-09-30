package handlers

import (
	"database/sql"
	"errors"
	"go-lb4/db"
	"go-lb4/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CartProductsListTmplContext struct {
	utils.BaseTmplContext

	Products   []db.CartProduct
	Pagination utils.PaginationInfo
}

func CartProductsListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.GetPageAndSize(r)

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

	products, count, err := db.GetCartProducts(cart.Id)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/cart/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Funcs(utils.TmplPaginationFuncs).Execute(w, CartProductsListTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "cart",
		},
		Products: products,
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/cart",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func CartProductEditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("Method is not allowed!"))
		return
	}

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

	itemIdStr := r.PathValue("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/cart", 301)
		return
	}

	product, err := db.GetCartProduct(itemId, cartId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown cart item!"))
		return
	}
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	allGood := true

	product.Quantity = utils.GetFormInt(r, "quantity", nil, &allGood, nil)
	if product.Quantity < 1 || product.Quantity > product.Product.Quantity {
		allGood = false
	}

	if allGood {
		err = product.DbSave()
		if utils.ReturnOnDatabaseError(err, w) {
			return
		}
	}

	http.Redirect(w, r, "/cart", 301)
}

type CartProductTmplContext struct {
	utils.BaseTmplContext

	Product      db.Product
	BackLocation string
	Error        string
}

func CartProductDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("Method is not allowed!"))
		return
	}

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

	itemIdStr := r.PathValue("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/cart", 301)
		return
	}

	product, err := db.GetCartProduct(itemId, cartId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown cart item!"))
		return
	}
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	if utils.ReturnOnDatabaseError(product.DbDelete(), w) {
		return
	}

	http.Redirect(w, r, "/cart", 301)
}
