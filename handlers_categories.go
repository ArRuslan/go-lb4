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

type CategoriesListTmplContext struct {
	Categories []Category
	Count      int
}

func categoriesListHandler(w http.ResponseWriter, r *http.Request) {
	categories, count, err := getCategories(getPageAndSize(r))

	tmpl, _ := template.ParseFiles("templates/categories/list.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, CategoriesListTmplContext{
		Categories: categories,
		Count:      count,
	})
	if err != nil {
		log.Println(err)
	}
}

func categoriesSearchHandler(w http.ResponseWriter, r *http.Request) {
	var categories []Category

	namePart := r.URL.Query().Get("name")
	if namePart != "" {
		_, pageSize := getPageAndSize(r)
		categories, _ = searchCategories(namePart, pageSize)
	}

	w.Header().Set("Content-Type", "application/json")

	if len(categories) > 0 {
		categoriesJson, _ := json.Marshal(categories)
		w.Write(categoriesJson)
	} else {
		w.Write([]byte("[]"))
	}
}

type CreateCategoryTmplContext struct {
	Name        string
	Description string

	Error string
}

func categoryCreateHandler(w http.ResponseWriter, r *http.Request) {
	var resp CreateCategoryTmplContext

	if r.Method == "POST" {
		allGood := true
		var newCategory Category

		newCategory.Name = getFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		newCategory.Description = getFormString(r, "description", &resp.Error, &allGood, &resp.Description)

		if allGood {
			err := newCategory.dbSave()
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/categories", 301)
			return
		}
	}

	tmpl, _ := template.ParseFiles("templates/categories/create.gohtml", "templates/layout.gohtml")
	err := tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type EditCategoryTmplContext struct {
	Name        string
	Description string

	Error string
}

func categoryEditHandler(w http.ResponseWriter, r *http.Request) {
	categoryIdStr := r.PathValue("categoryId")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		http.Redirect(w, r, "/categories", 301)
		return
	}

	category, err := getCategory(categoryId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown category!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := EditCategoryTmplContext{
		Name:        category.Name,
		Description: category.Description,
	}

	if r.Method == "POST" {
		allGood := true

		category.Name = getFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		category.Description = getFormString(r, "description", &resp.Error, &allGood, &resp.Description)

		if allGood {
			err = category.dbSave()
			if err == nil {
				http.Redirect(w, r, "/categories", 301)
				return
			}

			log.Println(err)
			resp.Error += "Database error occurred. "
		}
	}

	tmpl, _ := template.ParseFiles("templates/categories/edit.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type CategoryTmplContext struct {
	Category Category
	Error    string
}

func categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	categoryIdStr := r.PathValue("categoryId")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		http.Redirect(w, r, "/categories", 301)
		return
	}

	category, err := getCategory(categoryId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown category!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := CategoryTmplContext{
		Category: category,
		Error:    "",
	}

	if r.Method == "POST" {
		err = category.dbDelete()
		if err == nil {
			http.Redirect(w, r, "/categories", 301)
			return
		}

		log.Println(err)
		resp.Error += "Database error occurred. "
	}

	tmpl, _ := template.ParseFiles("templates/categories/delete.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}
