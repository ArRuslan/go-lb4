package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type ProductsListTmplContext struct {
	Products []Product
	Count    int
}

func productsListHandler(w http.ResponseWriter, r *http.Request) {
	products, count, err := getProducts(getPageAndSize(r))

	tmpl, _ := template.ParseFiles("templates/products/list.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, ProductsListTmplContext{
		Products: products,
		Count:    count,
	})
	if err != nil {
		log.Println(err)
	}
}

type CreateProductTmplContext struct {
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
	var resp CreateProductTmplContext

	if r.Method == "POST" {
		allGood := true
		var newProduct Product

		newProduct.Model = getFormStringNonEmpty(r, "model", &resp.Error, &allGood, &resp.Model)
		newProduct.Manufacturer = getFormStringNonEmpty(r, "manufacturer", &resp.Error, &allGood, &resp.Manufacturer)
		newProduct.Price = getFormDouble(r, "price", &resp.Error, &allGood, &resp.Price)
		newProduct.Quantity = getFormInt(r, "quantity", &resp.Error, &allGood, &resp.Quantity)
		newProduct.WarrantyDays = getFormInt(r, "warranty_days", &resp.Error, &allGood, &resp.ImageUrl)
		newProduct.ImageUrl = getFormString(r, "image_url", &resp.Error, &allGood, &resp.WarrantyDays)
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
	Model        string
	Manufacturer string
	Price        string
	Quantity     string
	ImageUrl     string
	WarrantyDays string
	CategoryId   string

	Error string
}

func productEditHandler(w http.ResponseWriter, r *http.Request) {
	productIdStr := r.PathValue("productId")
	productId, err := strconv.Atoi(productIdStr)
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
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := EditProductTmplContext{
		Model:        product.Model,
		Manufacturer: product.Manufacturer,
		Price:        strconv.FormatFloat(product.Price, 'f', 2, 64),
		Quantity:     strconv.FormatInt(int64(product.Quantity), 10),
		ImageUrl:     product.ImageUrl,
		WarrantyDays: strconv.FormatInt(int64(product.WarrantyDays), 10),
		CategoryId:   strconv.FormatInt(product.Category.Id, 10),
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

		if allGood {
			err = product.dbSave()
			if err == nil {
				http.Redirect(w, r, "/products", 301)
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
	Product Product
	Error   string
}

func productDeleteHandler(w http.ResponseWriter, r *http.Request) {
	productIdStr := r.PathValue("productId")
	productId, err := strconv.Atoi(productIdStr)
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
			http.Redirect(w, r, "/products", 301)
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
