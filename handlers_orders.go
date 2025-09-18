package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type OrdersListTmplContext struct {
	Orders []Order
	Count  int
}

func ordersListHandler(w http.ResponseWriter, r *http.Request) {
	orders, count, err := getOrders(getPageAndSize(r))

	tmpl, _ := template.ParseFiles("templates/orders/list.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, OrdersListTmplContext{
		Orders: orders,
		Count:  count,
	})
	if err != nil {
		log.Println(err)
	}
}

func customersSearchHandler(w http.ResponseWriter, r *http.Request) {
	var customers []Customer

	emailPart := r.URL.Query().Get("email")
	if emailPart != "" {
		_, pageSize := getPageAndSize(r)
		customers, _ = searchCustomersByEmail(emailPart, pageSize)
	}

	w.Header().Set("Content-Type", "application/json")

	if len(customers) > 0 {
		customersJson, _ := json.Marshal(customers)
		w.Write(customersJson)
	} else {
		w.Write([]byte("[]"))
	}
}

type CreateOrderTmplContext struct {
	CustomerEmail     string
	CustomerFirstName string
	CustomerLastName  string
	Address           string

	Error string
}

func orderCreateHandler(w http.ResponseWriter, r *http.Request) {
	var resp CreateOrderTmplContext

	if r.Method == "POST" {
		allGood := true
		var newOrder Order

		newOrder.Customer.Email = getFormStringNonEmpty(r, "customer_email", &resp.Error, &allGood, &resp.CustomerEmail)
		newOrder.Customer.FirstName = getFormStringNonEmpty(r, "customer_first_name", &resp.Error, &allGood, &resp.CustomerFirstName)
		newOrder.Customer.LastName = getFormStringNonEmpty(r, "customer_last_name", &resp.Error, &allGood, &resp.CustomerLastName)
		newOrder.Address = getFormStringNonEmpty(r, "address", &resp.Error, &allGood, &resp.Address)

		if allGood {
			err := newOrder.dbSave()
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
	CustomerEmailReadonly     string
	CustomerFirstNameReadonly string
	CustomerLastNameReadonly  string
	Address                   string

	Error string
}

func orderEditHandler(w http.ResponseWriter, r *http.Request) {
	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/orders", 301)
		return
	}

	order, err := getOrder(orderId)
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
		CustomerEmailReadonly:     order.Customer.Email,
		CustomerFirstNameReadonly: order.Customer.FirstName,
		CustomerLastNameReadonly:  order.Customer.LastName,
		Address:                   order.Address,
	}

	if r.Method == "POST" {
		allGood := true

		order.Address = getFormStringNonEmpty(r, "address", &resp.Error, &allGood, &resp.Address)

		if allGood {
			err = order.dbSave()
			if err == nil {
				http.Redirect(w, r, "/orders", 301)
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
	Order Order
	Error string
}

func orderDeleteHandler(w http.ResponseWriter, r *http.Request) {
	orderIdStr := r.PathValue("orderId")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/orders", 301)
		return
	}

	order, err := getOrder(orderId)
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
		Order: order,
		Error: "",
	}

	if r.Method == "POST" {
		err = order.dbDelete()
		if err == nil {
			http.Redirect(w, r, "/orders", 301)
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
