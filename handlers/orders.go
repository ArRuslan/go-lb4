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
)

type OrdersListTmplContext struct {
	utils.BaseTmplContext
	Orders     []db.Order
	Pagination utils.PaginationInfo
}

func OrdersListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.GetPageAndSize(r)
	orders, count, err := db.GetOrders(page, pageSize)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/orders/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, OrdersListTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "orders",
		},
		Orders: orders,
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/orders",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

type CreateOrderTmplContext struct {
	utils.BaseTmplContext

	CustomerEmail     string
	CustomerFirstName string
	CustomerLastName  string
	Address           string

	Error string
}

func OrderCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateOrderTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "orders",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newOrder db.Order

		newOrder.Customer.Email = utils.GetFormStringNonEmpty(r, "customer_email", &resp.Error, &allGood, &resp.CustomerEmail)
		newOrder.Customer.FirstName = utils.GetFormStringNonEmpty(r, "customer_first_name", &resp.Error, &allGood, &resp.CustomerFirstName)
		newOrder.Customer.LastName = utils.GetFormStringNonEmpty(r, "customer_last_name", &resp.Error, &allGood, &resp.CustomerLastName)
		newOrder.Address = utils.GetFormStringNonEmpty(r, "address", &resp.Error, &allGood, &resp.Address)

		if allGood {
			err := newOrder.DbSave(r.Context(), nil)
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/orders", 301)
			return
		}
	}

	tmpl, _ := template.ParseFiles("templates/orders/create.gohtml", "templates/layout.gohtml")
	err := tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type EditOrderTmplContext struct {
	utils.BaseTmplContext

	CustomerEmailReadonly     string
	CustomerFirstNameReadonly string
	CustomerLastNameReadonly  string
	Address                   string

	BackLocation string
	Error        string
}

func OrderEditHandler(w http.ResponseWriter, r *http.Request) {
	backLocation := r.URL.Query().Get("back")
	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/orders", 301)
		return
	}

	if backLocation == "order" {
		backLocation = "/orders/" + orderIdStr
	} else {
		backLocation = "/orders"
	}

	order, err := db.GetOrder(orderId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown order!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := EditOrderTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "orders",
		},
		CustomerEmailReadonly:     order.Customer.Email,
		CustomerFirstNameReadonly: order.Customer.FirstName,
		CustomerLastNameReadonly:  order.Customer.LastName,
		Address:                   order.Address,
		BackLocation:              backLocation,
	}

	if r.Method == "POST" {
		allGood := true

		order.Address = utils.GetFormStringNonEmpty(r, "address", &resp.Error, &allGood, &resp.Address)

		if allGood {
			err = order.DbSave(r.Context(), nil)
			if err == nil {
				http.Redirect(w, r, backLocation, 301)
				return
			}

			log.Println(err)
			resp.Error += "Database error occurred. "
		}
	}

	tmpl, _ := template.ParseFiles("templates/orders/edit.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type OrderTmplContext struct {
	utils.BaseTmplContext

	Order        db.Order
	BackLocation string
	Error        string
}

func OrderDeleteHandler(w http.ResponseWriter, r *http.Request) {
	backLocation := r.URL.Query().Get("back")
	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/orders", 301)
		return
	}

	if backLocation == "order" {
		backLocation = "/orders/" + orderIdStr
	} else {
		backLocation = "/orders"
	}

	order, err := db.GetOrder(orderId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown order!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := OrderTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "orders",
		},
		Order:        order,
		BackLocation: backLocation,
		Error:        "",
	}

	if r.Method == "POST" {
		err = order.DbDelete()
		if err == nil {
			http.Redirect(w, r, backLocation, 301)
			return
		}

		log.Println(err)
		resp.Error += "Database error occurred. "
	}

	tmpl, _ := template.ParseFiles("templates/orders/delete.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type OrderWithProductsTmplContext struct {
	utils.BaseTmplContext

	Order    db.Order
	Products []db.OrderItem
}

func OrderPageHandler(w http.ResponseWriter, r *http.Request) {
	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		http.Redirect(w, r, "/orders", 301)
		return
	}

	order, err := db.GetOrder(orderId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown order!"))
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	var items []db.OrderItem
	items, _, err = db.GetOrderItems(order.Id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := OrderWithProductsTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "orders",
		},
		Order:    order,
		Products: items,
	}

	tmpl, _ := template.ParseFiles("templates/orders/order.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

func OrderAddProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("Method is not allowed!"))
		return
	}

	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		http.Redirect(w, r, "/orders", 301)
		return
	}

	allGood := true
	prodId := utils.GetFormInt64(r, "product_id", nil, &allGood, nil)
	prodQuantity := utils.GetFormInt(r, "quantity", nil, &allGood, nil)

	if !allGood {
		http.Redirect(w, r, "/orders/"+orderIdStr, 301)
		return
	}

	var order db.Order
	order, err = db.GetOrder(orderId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown order!"))
		return
	}
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}
	if order.Status != "created" {
		http.Redirect(w, r, "/orders/"+orderIdStr, 301)
		return
	}

	product, err := db.GetProduct(prodId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown product!"))
		return
	}
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	if prodQuantity <= 0 || prodQuantity > product.Quantity {
		http.Redirect(w, r, "/orders/"+orderIdStr, 301)
		return
	}

	orderItem := db.OrderItem{
		Id:           0,
		OrderId:      order.Id,
		Product:      product,
		Quantity:     prodQuantity,
		PricePerItem: product.Price,
	}

	product.Quantity -= prodQuantity
	err = product.DbSave(r.Context(), nil)
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	err = orderItem.DbSave(r.Context(), nil)
	if err != nil {
		product.Quantity += prodQuantity
		err = product.DbSave(r.Context(), nil)
		if utils.ReturnOnDatabaseError(err, w) {
			return
		}
	}

	http.Redirect(w, r, "/orders/"+orderIdStr, 301)
}

func OrderDeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("Method is not allowed!"))
		return
	}

	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		http.Redirect(w, r, "/orders", 301)
		return
	}

	itemIdStr := r.PathValue("itemId")
	itemId, err := strconv.Atoi(itemIdStr)
	if err != nil {
		http.Redirect(w, r, "/orders/"+orderIdStr, 301)
		return
	}

	item, err := db.GetOrderItem(itemId, orderId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown order item!"))
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	err = item.DbDelete()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	product := item.Product
	product.Quantity += item.Quantity
	err = product.DbSave(r.Context(), nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	http.Redirect(w, r, "/orders/"+orderIdStr, 301)
}

func OrderFinishPaymentHandler(w http.ResponseWriter, r *http.Request) {
	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		http.Redirect(w, r, "/orders", 301)
		return
	}

	var order db.Order
	order, err = db.GetOrder(orderId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown order!"))
		return
	}
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}
	if order.Status != "payment" {
		http.Redirect(w, r, "/orders/"+orderIdStr, 301)
		return
	}

	completed, err := payPal.CheckOrderCompleted(order.PayPalId)
	if err == nil && completed {
		order.Status = "complete"
		if utils.ReturnOnDatabaseError(order.DbSave(r.Context(), nil), w) {
			return
		}
		http.Redirect(w, r, "/orders/"+orderIdStr, 301)
		return
	}

	w.Header().Set("Refresh", "5")

	tmpl, _ := template.ParseFiles("templates/orders/finish-payment.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, utils.BaseTmplContext{
		Type: "orders",
	})
	if err != nil {
		log.Println(err)
	}
}
