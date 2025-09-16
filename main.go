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

type Product struct {
	Id      int
	Model   string
	Company string
	Price   int
}

var database *sql.DB

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("SELECT `id`, `model`, `company`, `price` FROM products;")
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
	var products []Product

	for rows.Next() {
		product := Product{}
		err = rows.Scan(&product.Id, &product.Model, &product.Company, &product.Price)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, product)
	}

	tmpl, _ := template.ParseFiles("templates/index.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, products)
	if err != nil {
		log.Println(err)
	}
}

type CreateTmplResponse struct {
	Model   string
	Company string
	Price   string

	Error string
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var resp CreateTmplResponse

	if r.Method == "POST" {
		model := r.FormValue("model")
		company := r.FormValue("company")
		price := r.FormValue("price")

		priceInt, err := strconv.Atoi(price)
		if model != "" && company != "" && err == nil {
			_, err = database.Exec("INSERT INTO products (`model`, `company`, `price`) VALUES (?, ?, ?);", model, company, priceInt)
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
	var product Product

	productIdStr := r.PathValue("productId")
	productId, err := strconv.Atoi(productIdStr)
	if err != nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	row := database.QueryRow("SELECT `id`, `model`, `company`, `price` FROM products WHERE `id`=?;", productId)
	err = row.Scan(&product.Id, &product.Model, &product.Company, &product.Price)
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
		if model != "" && company != "" && atoiErr == nil {
			_, err = database.Exec("UPDATE `products` SET `model`=?, `company`=?, `price`=? WHERE `id`=?;", model, company, priceInt, productId)
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
	var product Product

	productIdStr := r.PathValue("productId")
	productId, err := strconv.Atoi(productIdStr)
	if err != nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	row := database.QueryRow("SELECT `id`, `model`, `company`, `price` FROM products WHERE `id`=?;", productId)
	err = row.Scan(&product.Id, &product.Model, &product.Company, &product.Price)
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
		_, err = database.Exec("DELETE FROM `products` WHERE `id`=?;", productId)
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
