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

type ProductsListTmplContext struct {
	BaseTmplContext

	Products []Product
	Count    int
}

func productsListHandler(w http.ResponseWriter, r *http.Request) {
	products, count, err := getProducts(getPageAndSize(r))

	tmpl, _ := template.ParseFiles("templates/products/list.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, ProductsListTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "products",
		},
		Products: products,
		Count:    count,
	})
	if err != nil {
		log.Println(err)
	}
}

func productsSearchHandler(w http.ResponseWriter, r *http.Request) {
	var products []Product

	namePart := r.URL.Query().Get("model")
	if namePart != "" {
		_, pageSize := getPageAndSize(r)
		products, _ = searchProducts(namePart, pageSize)
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
	BaseTmplContext

	Model        string
	Manufacturer string
	Price        string
	Quantity     string
	ImageUrl     string
	WarrantyDays string
	CategoryId   string

	Error string
}

func productCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateProductTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "products",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newProduct Product

		newProduct.Model = getFormStringNonEmpty(r, "model", &resp.Error, &allGood, &resp.Model)
		newProduct.Manufacturer = getFormStringNonEmpty(r, "manufacturer", &resp.Error, &allGood, &resp.Manufacturer)
		newProduct.Price = getFormDouble(r, "price", &resp.Error, &allGood, &resp.Price)
		newProduct.Quantity = getFormInt(r, "quantity", &resp.Error, &allGood, &resp.Quantity)
		newProduct.WarrantyDays = getFormInt(r, "warranty_days", &resp.Error, &allGood, &resp.WarrantyDays)
		newProduct.ImageUrl = getFormString(r, "image_url", &resp.Error, &allGood, &resp.ImageUrl)
		newProduct.Category.Id = getFormInt64(r, "category_id", &resp.Error, &allGood, &resp.CategoryId)

		if allGood {
			err := newProduct.dbSave()
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
	BaseTmplContext

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

func productEditHandler(w http.ResponseWriter, r *http.Request) {
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

	resp := EditProductTmplContext{
		BaseTmplContext: BaseTmplContext{
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

		product.Model = getFormStringNonEmpty(r, "model", &resp.Error, &allGood, &resp.Model)
		product.Manufacturer = getFormStringNonEmpty(r, "manufacturer", &resp.Error, &allGood, &resp.Manufacturer)
		product.Price = getFormDouble(r, "price", &resp.Error, &allGood, &resp.Price)
		product.Quantity = getFormInt(r, "quantity", &resp.Error, &allGood, &resp.Quantity)
		product.WarrantyDays = getFormInt(r, "warranty_days", &resp.Error, &allGood, &resp.ImageUrl)
		product.ImageUrl = getFormString(r, "image_url", &resp.Error, &allGood, &resp.WarrantyDays)
		product.Category.Id = getFormInt64(r, "category_id", &resp.Error, &allGood, &resp.CategoryId)
		resp.CategoryName = r.FormValue("_category_name")

		if allGood {
			err = product.dbSave()
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
	BaseTmplContext

	Product      Product
	BackLocation string
	Error        string
}

func productDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
		BaseTmplContext: BaseTmplContext{
			Type: "products",
		},
		Product:      product,
		BackLocation: backLocation,
		Error:        "",
	}

	if r.Method == "POST" {
		err = product.dbDelete()
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
	BaseTmplContext

	Product         Product
	Characteristics []ProductCharacteristic
}

func productPageHandler(w http.ResponseWriter, r *http.Request) {
	productIdStr := r.PathValue("productId")
	productId, err := strconv.ParseInt(productIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/products", 301)
		return
	}

	product, err := getProduct(productId)
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

	var characteristics []ProductCharacteristic
	characteristics, _, err = getProductCharacteristics(product.Id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := ProductWithCharacteristicsTmplContext{
		BaseTmplContext: BaseTmplContext{
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

func productAddCharacteristicHandler(w http.ResponseWriter, r *http.Request) {
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
	charId := getFormInt64(r, "characteristic_id", nil, &allGood, nil)
	charValue := getFormStringNonEmpty(r, "value", nil, &allGood, nil)

	if !allGood {
		http.Redirect(w, r, "/products/"+productIdStr, 301)
		return
	}

	product, err := getProduct(productId)
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

	var characteristic Characteristic
	characteristic, err = getCharacteristic(charId)
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

	productChar := ProductCharacteristic{
		Id:             0,
		ProductId:      product.Id,
		Characteristic: characteristic,
		Value:          charValue,
	}

	err = productChar.dbSave()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	http.Redirect(w, r, "/products/"+productIdStr, 301)
}

func productDeleteCharacteristicHandler(w http.ResponseWriter, r *http.Request) {
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

	characteristic, err := getProductCharacteristic(characteristicId, productId)
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

	err = characteristic.dbDelete()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	http.Redirect(w, r, "/products/"+productIdStr, 301)
}
