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

type CustomersListTmplContext struct {
	BaseTmplContext

	Customers  []Customer
	Pagination PaginationInfo
}

func customersListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := getPageAndSize(r)
	customers, count, err := getCustomers(page, pageSize)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(tmplPaginationFuncs).ParseFiles("templates/customers/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, CustomersListTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "customers",
		},
		Customers: customers,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			urlPath:  "/customers",
		},
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

type CreateCustomerTmplContext struct {
	BaseTmplContext

	FirstName string
	LastName  string
	Email     string

	Error string
}

func customerCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateCustomerTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "customers",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newCustomer Customer

		newCustomer.FirstName = getFormStringNonEmpty(r, "first_name", &resp.Error, &allGood, &resp.FirstName)
		newCustomer.LastName = getFormStringNonEmpty(r, "last_name", &resp.Error, &allGood, &resp.LastName)
		newCustomer.Email = getFormStringNonEmpty(r, "email", &resp.Error, &allGood, &resp.Email)

		if allGood {
			err := newCustomer.dbSave()
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/customers", 301)
			return
		}
	}

	tmpl, _ := template.ParseFiles("templates/customers/create.gohtml", "templates/layout.gohtml")
	err := tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type EditCustomerTmplContext struct {
	BaseTmplContext

	FirstName string
	LastName  string
	Email     string

	Error string
}

func customerEditHandler(w http.ResponseWriter, r *http.Request) {
	customerIdStr := r.PathValue("customerId")
	customerId, err := strconv.Atoi(customerIdStr)
	if err != nil {
		http.Redirect(w, r, "/customers", 301)
		return
	}

	customer, err := getCustomer(customerId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown customer!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := EditCustomerTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "customers",
		},
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
	}

	if r.Method == "POST" {
		allGood := true

		customer.FirstName = getFormStringNonEmpty(r, "first_name", &resp.Error, &allGood, &resp.FirstName)
		customer.LastName = getFormStringNonEmpty(r, "last_name", &resp.Error, &allGood, &resp.LastName)
		customer.Email = getFormStringNonEmpty(r, "email", &resp.Error, &allGood, &resp.Email)

		if allGood {
			err = customer.dbSave()
			if err == nil {
				http.Redirect(w, r, "/customers", 301)
				return
			}

			log.Println(err)
			resp.Error += "Database error occurred. "
		}
	}

	tmpl, _ := template.ParseFiles("templates/customers/edit.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type CustomerTmplContext struct {
	BaseTmplContext
	Customer Customer
	Error    string
}

func customerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	customerIdStr := r.PathValue("customerId")
	customerId, err := strconv.Atoi(customerIdStr)
	if err != nil {
		http.Redirect(w, r, "/customers", 301)
		return
	}

	customer, err := getCustomer(customerId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown customer!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := CustomerTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "customers",
		},
		Customer: customer,
		Error:    "",
	}

	if r.Method == "POST" {
		err = customer.dbDelete()
		if err == nil {
			http.Redirect(w, r, "/customers", 301)
			return
		}

		log.Println(err)
		resp.Error += "Database error occurred. "
	}

	tmpl, _ := template.ParseFiles("templates/customers/delete.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}
