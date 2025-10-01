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
	time "time"
)

type ProductsListTmplContext struct {
	utils.BaseTmplContext

	Products   []db.Product
	Pagination utils.PaginationInfo
}

func ProductsListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.GetPageAndSize(r)
	products, count, err := db.GetProducts(page, pageSize)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/products/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Funcs(utils.TmplPaginationFuncs).Execute(w, ProductsListTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "products",
		},
		Products: products,
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/products",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func ProductsSearchHandler(w http.ResponseWriter, r *http.Request) {
	var products []db.Product

	namePart := r.URL.Query().Get("model")
	if namePart != "" {
		_, pageSize := utils.GetPageAndSize(r)
		products, _ = db.SearchProducts(namePart, pageSize)
	}

	w.Header().Set("Content-Type", "application/json")

	if len(products) > 0 {
		productsJson, _ := json.Marshal(products)
		w.Write(productsJson)
	} else {
		w.Write([]byte("[]"))
	}
}

type CreateProductTmplContext struct {
	utils.BaseTmplContext

	Model        string
	Manufacturer string
	Price        string
	Quantity     string
	ImageUrl     string
	WarrantyDays string
	CategoryId   string

	Error string
}

func ProductCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateProductTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "products",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newProduct db.Product

		newProduct.Model = utils.GetFormStringNonEmpty(r, "model", &resp.Error, &allGood, &resp.Model)
		newProduct.Manufacturer = utils.GetFormStringNonEmpty(r, "manufacturer", &resp.Error, &allGood, &resp.Manufacturer)
		newProduct.Price = utils.GetFormDouble(r, "price", &resp.Error, &allGood, &resp.Price)
		newProduct.Quantity = utils.GetFormInt(r, "quantity", &resp.Error, &allGood, &resp.Quantity)
		newProduct.WarrantyDays = utils.GetFormInt(r, "warranty_days", &resp.Error, &allGood, &resp.WarrantyDays)
		newProduct.ImageUrl = utils.GetFormString(r, "image_url", &resp.Error, &allGood, &resp.ImageUrl)
		newProduct.Category.Id = utils.GetFormInt64(r, "category_id", &resp.Error, &allGood, &resp.CategoryId)

		if allGood {
			err := newProduct.DbSave(r.Context(), nil)
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/products", 301)
			return
		}
	}

	tmpl, _ := template.ParseFiles("templates/products/create.gohtml", "templates/layout.gohtml")
	err := tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type EditProductTmplContext struct {
	utils.BaseTmplContext

	Model        string
	Manufacturer string
	Price        string
	Quantity     string
	ImageUrl     string
	WarrantyDays string
	CategoryId   string
	CategoryName string

	BackLocation string
	Error        string
}

func ProductEditHandler(w http.ResponseWriter, r *http.Request) {
	backLocation := r.URL.Query().Get("back")
	productIdStr := r.PathValue("productId")
	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/products", 301)
		return
	}

	if backLocation == "product" {
		backLocation = "/products/" + productIdStr
	} else {
		backLocation = "/products"
	}

	product, err := db.GetProduct(productId)
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

	resp := EditProductTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "products",
		},
		Model:        product.Model,
		Manufacturer: product.Manufacturer,
		Price:        strconv.FormatFloat(product.Price, 'f', 2, 64),
		Quantity:     strconv.FormatInt(int64(product.Quantity), 10),
		ImageUrl:     product.ImageUrl,
		WarrantyDays: strconv.FormatInt(int64(product.WarrantyDays), 10),
		CategoryId:   strconv.FormatInt(product.Category.Id, 10),
		CategoryName: product.Category.Name,
		BackLocation: backLocation,
	}

	if r.Method == "POST" {
		allGood := true

		product.Model = utils.GetFormStringNonEmpty(r, "model", &resp.Error, &allGood, &resp.Model)
		product.Manufacturer = utils.GetFormStringNonEmpty(r, "manufacturer", &resp.Error, &allGood, &resp.Manufacturer)
		product.Price = utils.GetFormDouble(r, "price", &resp.Error, &allGood, &resp.Price)
		product.Quantity = utils.GetFormInt(r, "quantity", &resp.Error, &allGood, &resp.Quantity)
		product.WarrantyDays = utils.GetFormInt(r, "warranty_days", &resp.Error, &allGood, &resp.ImageUrl)
		product.ImageUrl = utils.GetFormString(r, "image_url", &resp.Error, &allGood, &resp.WarrantyDays)
		product.Category.Id = utils.GetFormInt64(r, "category_id", &resp.Error, &allGood, &resp.CategoryId)
		resp.CategoryName = r.FormValue("_category_name")

		if allGood {
			err = product.DbSave(r.Context(), nil)
			if err == nil {
				http.Redirect(w, r, backLocation, 301)
				return
			}

			log.Println(err)
			resp.Error += "Database error occurred. "
		}
	}

	tmpl, _ := template.ParseFiles("templates/products/edit.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type ProductTmplContext struct {
	utils.BaseTmplContext

	Product      db.Product
	BackLocation string
	Error        string
}

func ProductDeleteHandler(w http.ResponseWriter, r *http.Request) {
	backLocation := r.URL.Query().Get("back")

	productIdStr := r.PathValue("productId")
	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/products", 301)
		return
	}

	if backLocation == "product" {
		backLocation = "/products/" + productIdStr
	} else {
		backLocation = "/products"
	}

	product, err := db.GetProduct(productId)
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
		BaseTmplContext: utils.BaseTmplContext{
			Type: "products",
		},
		Product:      product,
		BackLocation: backLocation,
		Error:        "",
	}

	if r.Method == "POST" {
		err = product.DbDelete()
		if err == nil {
			http.Redirect(w, r, backLocation, 301)
			return
		}

		log.Println(err)
		resp.Error += "Database error occurred. "
	}

	tmpl, _ := template.ParseFiles("templates/products/delete.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type ProductWithCharacteristicsTmplContext struct {
	utils.BaseTmplContext

	Product         db.Product
	Characteristics []db.ProductCharacteristic
}

func ProductPageHandler(w http.ResponseWriter, r *http.Request) {
	productIdStr := r.PathValue("productId")
	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/products", 301)
		return
	}

	product, err := db.GetProduct(productId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown product!"))
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	var characteristics []db.ProductCharacteristic
	characteristics, _, err = db.GetProductCharacteristics(product.Id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := ProductWithCharacteristicsTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "products",
		},
		Product:         product,
		Characteristics: characteristics,
	}

	tmpl, _ := template.ParseFiles("templates/products/product.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

func ProductAddCharacteristicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("Method is not allowed!"))
		return
	}

	productIdStr := r.PathValue("productId")
	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/products", 301)
		return
	}

	allGood := true
	charId := utils.GetFormInt64(r, "characteristic_id", nil, &allGood, nil)
	charValue := utils.GetFormStringNonEmpty(r, "value", nil, &allGood, nil)

	if !allGood {
		http.Redirect(w, r, "/products/"+productIdStr, 301)
		return
	}

	product, err := db.GetProduct(productId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown product!"))
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	var characteristic db.Characteristic
	characteristic, err = db.GetCharacteristic(charId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown characteristic!"))
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	productChar := db.ProductCharacteristic{
		Id:             0,
		ProductId:      product.Id,
		Characteristic: characteristic,
		Value:          charValue,
	}

	err = productChar.DbSave()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	http.Redirect(w, r, "/products/"+productIdStr, 301)
}

func ProductDeleteCharacteristicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("Method is not allowed!"))
		return
	}

	productIdStr := r.PathValue("productId")
	productId, err := strconv.Atoi(productIdStr)
	if err != nil {
		http.Redirect(w, r, "/products", 301)
		return
	}

	characteristicIdStr := r.PathValue("characteristicId")
	characteristicId, err := strconv.Atoi(characteristicIdStr)
	if err != nil {
		http.Redirect(w, r, "/products/"+productIdStr, 301)
		return
	}

	characteristic, err := db.GetProductCharacteristic(characteristicId, productId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown characteristic!"))
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	err = characteristic.DbDelete()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	http.Redirect(w, r, "/products/"+productIdStr, 301)
}

func ProductAddToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write([]byte("Method is not allowed!"))
		return
	}

	productIdStr := r.PathValue("productId")
	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/products", 301)
		return
	}

	product, err := db.GetProduct(productId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown product!"))
		return
	}
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	cartId := utils.GetCartId(r)
	http.SetCookie(w, &http.Cookie{Name: "cartId", Value: cartId.String(), Path: "/", HttpOnly: true, MaxAge: 86400})
	cart, err := db.GetOrCreateCart(cartId)
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	cart.LastAccessTime = time.Now()
	if utils.ReturnOnDatabaseError(cart.DbSave(), w) {
		return
	}

	cartProduct, err := db.GetCartProductByProductId(product.Id, cart.Id)
	if errors.Is(err, sql.ErrNoRows) {
		cartProduct = db.CartProduct{
			Id:       0,
			CartId:   cartId,
			Product:  product,
			Quantity: 1,
		}
	} else if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	if utils.ReturnOnDatabaseError(cartProduct.DbSave(r.Context(), nil), w) {
		return
	}

	http.Redirect(w, r, "/products", 301)
}
