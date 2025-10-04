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
)

type CategoriesListTmplContext struct {
	utils.BaseTmplContext

	Categories []db.Category
	Pagination utils.PaginationInfo
}

func CategoriesListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.GetPageAndSize(r)
	categories, count, err := db.GetCategories(page, pageSize)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/categories/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, CategoriesListTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "categories",
		},
		Categories: categories,
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/categories",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func CategoriesSearchHandler(w http.ResponseWriter, r *http.Request) {
	var categories []db.Category

	namePart := r.URL.Query().Get("name")
	if namePart != "" {
		_, pageSize := utils.GetPageAndSize(r)
		categories, _ = db.SearchCategories(namePart, pageSize)
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
	utils.BaseTmplContext

	Name        string
	Description string

	Error string
}

func CategoryCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateCategoryTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "categories",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newCategory db.Category

		newCategory.Name = utils.GetFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		newCategory.Description = utils.GetFormString(r, "description", &resp.Error, &allGood, &resp.Description)

		if allGood {
			err := newCategory.DbSave()
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
	utils.BaseTmplContext

	Name        string
	Description string

	Error string
}

func CategoryEditHandler(w http.ResponseWriter, r *http.Request) {
	categoryIdStr := r.PathValue("categoryId")
	categoryId, err := strconv.ParseInt(categoryIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/categories", 301)
		return
	}

	category, err := db.GetCategory(categoryId)
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
		BaseTmplContext: utils.BaseTmplContext{
			Type: "categories",
		},
		Name:        category.Name,
		Description: category.Description,
	}

	if r.Method == "POST" {
		allGood := true

		category.Name = utils.GetFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		category.Description = utils.GetFormString(r, "description", &resp.Error, &allGood, &resp.Description)

		if allGood {
			err = category.DbSave()
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
	utils.BaseTmplContext

	Category db.Category
	Error    string
}

func CategoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	categoryIdStr := r.PathValue("categoryId")
	categoryId, err := strconv.ParseInt(categoryIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/categories", 301)
		return
	}

	category, err := db.GetCategory(categoryId)
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
		BaseTmplContext: utils.BaseTmplContext{
			Type: "categories",
		},
		Category: category,
		Error:    "",
	}

	if r.Method == "POST" {
		err = category.DbDelete()
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
