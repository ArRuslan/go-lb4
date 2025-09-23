package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type BaseTmplContext struct {
	Type string
}

/*
База даних має містити щонайменше 5 таблиць.
Веб-додаток повинен надавати можливості редагування, додавання та видалення даних, а також містити блок аналізу даних (наприклад, товар, що найчастіше/рідко купується).
Параметри аналізу даних визначити самостійно.
Можна створювати БД за індивідуальним завданням, або взяти будь-яку свою БД, яка вже була створена на іншій дисципліні (курсовій роботі)
*/
func main() {
	db, err := sql.Open("mysql", "nure_golang_pz3:123456789@tcp(127.0.0.1:3306)/nure_golang_pz3?parseTime=true")
	if err != nil {
		panic(err)
	}

	defer db.Close()
	database = db

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/products", 301)
	})

	http.HandleFunc("/products", productsListHandler)
	http.HandleFunc("/products/create", productCreateHandler)
	http.HandleFunc("/products/search", productsSearchHandler)
	http.HandleFunc("/products/{productId}/edit", productEditHandler)
	http.HandleFunc("/products/{productId}/delete", productDeleteHandler)
	http.HandleFunc("/products/{productId}", productPageHandler)
	http.HandleFunc("/products/{productId}/characteristics", productAddCharacteristicHandler)
	http.HandleFunc("/products/{productId}/characteristics/{characteristicId}/delete", productDeleteCharacteristicHandler)

	http.HandleFunc("/categories", categoriesListHandler)
	http.HandleFunc("/categories/create", categoryCreateHandler)
	http.HandleFunc("/categories/{categoryId}/edit", categoryEditHandler)
	http.HandleFunc("/categories/{categoryId}/delete", categoryDeleteHandler)
	http.HandleFunc("/categories/search", categoriesSearchHandler)

	http.HandleFunc("/characteristics", characteristicsListHandler)
	http.HandleFunc("/characteristics/create", characteristicCreateHandler)
	http.HandleFunc("/characteristics/{characteristicId}/edit", characteristicEditHandler)
	http.HandleFunc("/characteristics/{characteristicId}/delete", characteristicDeleteHandler)
	http.HandleFunc("/characteristics/search", characteristicsSearchHandler)

	http.HandleFunc("/customers", customersListHandler)
	http.HandleFunc("/customers/create", customerCreateHandler)
	http.HandleFunc("/customers/{customerId}/edit", customerEditHandler)
	http.HandleFunc("/customers/{customerId}/delete", customerDeleteHandler)
	http.HandleFunc("/customers/search", customersSearchHandler)

	http.HandleFunc("/orders", ordersListHandler)
	http.HandleFunc("/orders/create", orderCreateHandler)
	http.HandleFunc("/orders/{orderId}/edit", orderEditHandler)
	http.HandleFunc("/orders/{orderId}/delete", orderDeleteHandler)
	http.HandleFunc("/orders/{orderId}", orderPageHandler)
	http.HandleFunc("/orders/{orderId}/products", orderAddProductHandler)
	http.HandleFunc("/orders/{orderId}/products/{itemId}/delete", orderDeleteProductHandler)

	http.HandleFunc("/analysis", productsAnalysisHandler)

	fmt.Println("Server is listening on port 8081 (http://127.0.0.1:8081)")
	err = http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		panic(err)
	}
}
