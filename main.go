package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type ProductsListTmplContext struct {
	Products []Product
	Count    int
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	products, count, err := getProducts(getPageAndSize(r))

	tmpl, _ := template.ParseFiles("templates/index.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, ProductsListTmplContext{
		Products: products,
		Count:    count,
	})
	if err != nil {
		log.Println(err)
	}
}

type CreateProductTmplContext struct {
	Model   string
	Company string
	Price   string

	Error string
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var resp CreateProductTmplContext

	if r.Method == "POST" {
		model := r.FormValue("model")
		company := r.FormValue("company")
		price := r.FormValue("price")

		priceInt, err := strconv.Atoi(price)
		if model != "" && company != "" && err == nil {
			newProduct := Product{
				Id:      0,
				Model:   model,
				Company: company,
				Price:   priceInt,
			}
			err = newProduct.dbSave()
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/", 301)
			return
		}

		resp.Model = model
		resp.Company = company
		resp.Price = price

		if model == "" {
			resp.Error += "Model is empty or invalid. "
		}
		if company == "" {
			resp.Error += "Company is empty or invalid. "
		}
		if err != nil {
			resp.Error += "Price is empty or invalid. "
		}
	}

	tmpl, _ := template.ParseFiles("templates/create.gohtml", "templates/layout.gohtml")
	err := tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type ProductTmplContext struct {
	Product Product
	Error   string
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	productIdStr := r.PathValue("productId")
	productId, err := strconv.Atoi(productIdStr)
	if err != nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	product, err := getProduct(productId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown product!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := ProductTmplContext{
		Product: product,
		Error:   "",
	}

	if r.Method == "POST" {
		model := r.FormValue("model")
		company := r.FormValue("company")
		price := r.FormValue("price")
		priceInt, atoiErr := strconv.Atoi(price)

		resp.Product.Model = model
		resp.Product.Company = company
		resp.Product.Price = priceInt

		if model != "" && company != "" && atoiErr == nil {
			err = resp.Product.dbSave()
			if err == nil {
				http.Redirect(w, r, "/", 301)
				return
			}

			log.Println(err)
			resp.Error += "Database error occurred. "
		}

		if model == "" {
			resp.Error += "Model is empty or invalid. "
		}
		if company == "" {
			resp.Error += "Company is empty or invalid. "
		}
		if atoiErr != nil {
			resp.Error += "Price is empty or invalid. "
		}
	}

	tmpl, _ := template.ParseFiles("templates/edit.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	productIdStr := r.PathValue("productId")
	productId, err := strconv.Atoi(productIdStr)
	if err != nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	product, err := getProduct(productId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown product!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := ProductTmplContext{
		Product: product,
		Error:   "",
	}

	if r.Method == "POST" {
		err = product.dbDelete()
		if err == nil {
			http.Redirect(w, r, "/", 301)
			return
		}

		log.Println(err)
		resp.Error += "Database error occurred. "
	}

	tmpl, _ := template.ParseFiles("templates/delete.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

/*
База даних має містити щонайменше 5 таблиць.
Веб-додаток повинен надавати можливості редагування, додавання та видалення даних, а також містити блок аналізу даних (наприклад, товар, що найчастіше/рідко купується).
Параметри аналізу даних визначити самостійно.
Можна створювати БД за індивідуальним завданням, або взяти будь-яку свою БД, яка вже була створена на іншій дисципліні (курсовій роботі)
*/
func main() {
	db, err := sql.Open("mysql", "nure_golang_pz3:123456789@tcp(127.0.0.1:3306)/nure_golang_pz3")
	if err != nil {
		panic(err)
	}

	defer db.Close()
	database = db

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/products/{productId}/edit", editHandler)
	http.HandleFunc("/products/{productId}/delete", deleteHandler)

	fmt.Println("Server is listening on port 8081")
	err = http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		panic(err)
	}
}
