package main

import (
	"fmt"
	"go-lb4/db"
	"go-lb4/handlers"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

/*
База даних має містити щонайменше 5 таблиць.
Веб-додаток повинен надавати можливості редагування, додавання та видалення даних, а також містити блок аналізу даних (наприклад, товар, що найчастіше/рідко купується).
Параметри аналізу даних визначити самостійно.
Можна створювати БД за індивідуальним завданням, або взяти будь-яку свою БД, яка вже була створена на іншій дисципліні (курсовій роботі)

Інтернет магазин
Вимоги:
  - Механізм додавання нового товару
  - Механізм видалення вже існуючого товару
  - Каталог товарів
  - Кошик
  - механізм оплати вибраних товарів

Аналітичний модуль:
Необхідне графічне представлення отриманих результатів у вигляді графіків/гістограм/діаграм/таблиць залежно від переваг розробника.
Необхідно мати завантажений або сформований випадковим чином набір даних про покупки (щонайменше 100 позицій)
Необхідно відображати наступну базову інформацію:
  - Кількість покупців по днях,
  - День з мінімальною та з максимальною кількістю скоєних покупок,
  - середній чек,
  - медіанний чек,
  - Найбільш часто купується товар,
  - які товари найчастіше купують разом з найчастіше купованим,
  - комбінації товарів, що найчастіше зустрічаються,
  - найрідше зустрічаються комбінації товарів
*/
func main() {
	db.InitDatabase("mysql", "nure_golang_pz3:123456789@tcp(127.0.0.1:3306)/nure_golang_pz3?parseTime=true")
	defer db.CloseDatabase()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/products", 301)
	})

	http.HandleFunc("/products", handlers.ProductsListHandler)
	http.HandleFunc("/products/create", handlers.ProductCreateHandler)
	http.HandleFunc("/products/search", handlers.ProductsSearchHandler)
	http.HandleFunc("/products/{productId}/edit", handlers.ProductEditHandler)
	http.HandleFunc("/products/{productId}/delete", handlers.ProductDeleteHandler)
	http.HandleFunc("/products/{productId}", handlers.ProductPageHandler)
	http.HandleFunc("/products/{productId}/characteristics", handlers.ProductAddCharacteristicHandler)
	http.HandleFunc("/products/{productId}/characteristics/{characteristicId}/delete", handlers.ProductDeleteCharacteristicHandler)
	http.HandleFunc("/products/{productId}/add-to-cart", handlers.ProductAddToCartHandler)

	http.HandleFunc("/categories", handlers.CategoriesListHandler)
	http.HandleFunc("/categories/create", handlers.CategoryCreateHandler)
	http.HandleFunc("/categories/{categoryId}/edit", handlers.CategoryEditHandler)
	http.HandleFunc("/categories/{categoryId}/delete", handlers.CategoryDeleteHandler)
	http.HandleFunc("/categories/search", handlers.CategoriesSearchHandler)

	http.HandleFunc("/characteristics", handlers.CharacteristicsListHandler)
	http.HandleFunc("/characteristics/create", handlers.CharacteristicCreateHandler)
	http.HandleFunc("/characteristics/{characteristicId}/edit", handlers.CharacteristicEditHandler)
	http.HandleFunc("/characteristics/{characteristicId}/delete", handlers.CharacteristicDeleteHandler)
	http.HandleFunc("/characteristics/search", handlers.CharacteristicsSearchHandler)

	http.HandleFunc("/customers", handlers.CustomersListHandler)
	http.HandleFunc("/customers/create", handlers.CustomerCreateHandler)
	http.HandleFunc("/customers/{customerId}/edit", handlers.CustomerEditHandler)
	http.HandleFunc("/customers/{customerId}/delete", handlers.CustomerDeleteHandler)
	http.HandleFunc("/customers/search", handlers.CustomersSearchHandler)

	http.HandleFunc("/orders", handlers.OrdersListHandler)
	http.HandleFunc("/orders/create", handlers.OrderCreateHandler)
	http.HandleFunc("/orders/{orderId}/edit", handlers.OrderEditHandler)
	http.HandleFunc("/orders/{orderId}/delete", handlers.OrderDeleteHandler)
	http.HandleFunc("/orders/{orderId}", handlers.OrderPageHandler)
	http.HandleFunc("/orders/{orderId}/products", handlers.OrderAddProductHandler)
	http.HandleFunc("/orders/{orderId}/products/{itemId}/delete", handlers.OrderDeleteProductHandler)

	http.HandleFunc("/analysis", handlers.ProductsAnalysisHandler)

	http.HandleFunc("/cart", handlers.CartProductsListHandler)
	http.HandleFunc("/cart/{itemId}/edit", handlers.CartProductEditHandler)
	http.HandleFunc("/cart/{itemId}/delete", handlers.CartProductDeleteHandler)

	fmt.Println("Server is listening on port 8081 (http://127.0.0.1:8081)")
	err := http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		panic(err)
	}
}
