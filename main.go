package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/products", 301)
	})
	http.HandleFunc("/products", indexHandler)
	http.HandleFunc("/products/create", createHandler)
	http.HandleFunc("/products/{productId}/edit", editHandler)
	http.HandleFunc("/products/{productId}/delete", deleteHandler)

	fmt.Println("Server is listening on port 8081")
	err = http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		panic(err)
	}
}
