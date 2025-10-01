package handlers

import (
	"database/sql"
	"errors"
	"fmt"
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

	Products []db.CartProduct
}

func CartProductsListHandler(w http.ResponseWriter, r *http.Request) {
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

	products, _, err := db.GetCartProducts(cart.Id)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/cart/list.gohtml", "templates/layout.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Funcs(utils.TmplPaginationFuncs).Execute(w, CartProductsListTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "cart",
		},
		Products: products,
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
		err = product.DbSave(r.Context(), nil)
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

type CartPaymentTmplContext struct {
	utils.BaseTmplContext

	CustomerEmail     string
	CustomerFirstName string
	CustomerLastName  string
	Address           string

	Products      []db.CartProduct
	ProductsCount int
	CartTotal     float64

	Error string
}

func CartPaymentHandler(w http.ResponseWriter, r *http.Request) {
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

	products, _, err := db.GetCartProducts(cart.Id)

	total := 0.
	allProductsCount := 0

	ctx := r.Context()

	tx, err := db.BeginTx(ctx)
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}
	defer tx.Rollback()
	for _, product := range products {
		if product.Quantity > product.Product.Quantity {
			product.Quantity = product.Product.Quantity
			if utils.ReturnOnDatabaseError(product.DbSave(ctx, tx), w) {
				return
			}
		}

		total += float64(product.Quantity) * product.Product.Price
		allProductsCount += product.Quantity
	}

	if utils.ReturnOnDatabaseError(tx.Commit(), w) {
		return
	}

	resp := CartPaymentTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "cart",
		},
		Products:      products,
		ProductsCount: allProductsCount,
		CartTotal:     total,
	}

	if r.Method == "POST" {
		tx, err = db.BeginTx(ctx)
		if utils.ReturnOnDatabaseError(err, w) {
			return
		}
		defer tx.Rollback()

		order := db.Order{
			Id:        0,
			Customer:  db.Customer{},
			CreatedAt: time.Now(),
			Address:   "",
			Status:    "created",
		}

		allGood := true

		order.Customer.Email = utils.GetFormStringNonEmpty(r, "email", &resp.Error, &allGood, &resp.CustomerEmail)
		order.Customer.FirstName = utils.GetFormStringNonEmpty(r, "first_name", &resp.Error, &allGood, &resp.CustomerFirstName)
		order.Customer.LastName = utils.GetFormStringNonEmpty(r, "last_name", &resp.Error, &allGood, &resp.CustomerLastName)
		order.Address = utils.GetFormStringNonEmpty(r, "address", &resp.Error, &allGood, &resp.Address)

		if allGood {
			if utils.ReturnOnDatabaseError(order.DbSave(ctx, tx), w) {
				return
			}

			for _, product := range products {
				if utils.ReturnOnDatabaseError(product.Product.SubtractQuantity(ctx, product.Quantity, tx), w) {
					return
				}

				item := db.OrderItem{
					Id:           0,
					OrderId:      order.Id,
					Product:      product.Product,
					Quantity:     product.Quantity,
					PricePerItem: product.Product.Price,
				}

				if utils.ReturnOnDatabaseError(item.DbSave(ctx, tx), w) {
					return
				}
			}

			if utils.ReturnOnDatabaseError(tx.Commit(), w) {
				return
			}

			http.Redirect(w, r, fmt.Sprintf("/orders/%d", order.Id), 301)
			return
		}
	}

	tmpl := template.New("payment.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/cart/payment.gohtml", "templates/layout.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Funcs(utils.TmplPaginationFuncs).Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}
