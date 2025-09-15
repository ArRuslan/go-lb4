package main

import (
	"database/sql"
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

	tmpl, _ := template.ParseFiles("templates/index.gohtml")
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

	tmpl, _ := template.ParseFiles("templates/create.gohtml")
	err := tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	db, err := sql.Open("mysql", "nure_golang_pz3:123456789@tcp(127.0.0.1:3306)/nure_golang_pz3")
	if err != nil {
		panic(err)
	}

	defer db.Close()
	database = db

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)

	fmt.Println("Server is listening on port 8081")
	err = http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		panic(err)
	}
}
