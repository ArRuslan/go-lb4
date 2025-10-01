package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-lb4/db"
	"go-lb4/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type CustomersListTmplContext struct {
	utils.BaseTmplContext

	Customers  []db.Customer
	Pagination utils.PaginationInfo
}

func CustomersListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.GetPageAndSize(r)
	customers, count, err := db.GetCustomers(page, pageSize)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/customers/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, CustomersListTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "customers",
		},
		Customers: customers,
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/customers",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func CustomersSearchHandler(w http.ResponseWriter, r *http.Request) {
	var customers []db.Customer

	emailPart := r.URL.Query().Get("email")
	if emailPart != "" {
		_, pageSize := utils.GetPageAndSize(r)
		customers, _ = db.SearchCustomersByEmail(emailPart, pageSize)
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
	utils.BaseTmplContext

	FirstName string
	LastName  string
	Email     string

	Error string
}

func CustomerCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateCustomerTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "customers",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newCustomer db.Customer

		newCustomer.FirstName = utils.GetFormStringNonEmpty(r, "first_name", &resp.Error, &allGood, &resp.FirstName)
		newCustomer.LastName = utils.GetFormStringNonEmpty(r, "last_name", &resp.Error, &allGood, &resp.LastName)
		newCustomer.Email = utils.GetFormStringNonEmpty(r, "email", &resp.Error, &allGood, &resp.Email)

		if allGood {
			err := newCustomer.DbSave(r.Context(), nil)
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
	utils.BaseTmplContext

	FirstName string
	LastName  string
	Email     string

	Error string
}

func CustomerEditHandler(w http.ResponseWriter, r *http.Request) {
	customerIdStr := r.PathValue("customerId")
	customerId, err := strconv.Atoi(customerIdStr)
	if err != nil {
		http.Redirect(w, r, "/customers", 301)
		return
	}

	customer, err := db.GetCustomer(customerId)
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
		BaseTmplContext: utils.BaseTmplContext{
			Type: "customers",
		},
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
	}

	if r.Method == "POST" {
		allGood := true

		customer.FirstName = utils.GetFormStringNonEmpty(r, "first_name", &resp.Error, &allGood, &resp.FirstName)
		customer.LastName = utils.GetFormStringNonEmpty(r, "last_name", &resp.Error, &allGood, &resp.LastName)
		customer.Email = utils.GetFormStringNonEmpty(r, "email", &resp.Error, &allGood, &resp.Email)

		if allGood {
			err = customer.DbSave(r.Context(), nil)
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
	utils.BaseTmplContext
	Customer db.Customer
	Error    string
}

func CustomerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	customerIdStr := r.PathValue("customerId")
	customerId, err := strconv.Atoi(customerIdStr)
	if err != nil {
		http.Redirect(w, r, "/customers", 301)
		return
	}

	customer, err := db.GetCustomer(customerId)
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
		BaseTmplContext: utils.BaseTmplContext{
			Type: "customers",
		},
		Customer: customer,
		Error:    "",
	}

	if r.Method == "POST" {
		err = customer.DbDelete()
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
