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

type CharacteristicsListTmplContext struct {
	BaseTmplContext

	Characteristics []Characteristic
	Pagination      PaginationInfo
}

func characteristicsListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := getPageAndSize(r)
	characteristics, count, err := getCharacteristics(page, pageSize)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(tmplPaginationFuncs).ParseFiles("templates/characteristics/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, CharacteristicsListTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "characteristics",
		},
		Characteristics: characteristics,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			urlPath:  "/characteristics",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func characteristicsSearchHandler(w http.ResponseWriter, r *http.Request) {
	var characteristics []Characteristic

	namePart := r.URL.Query().Get("name")
	if namePart != "" {
		_, pageSize := getPageAndSize(r)
		characteristics, _ = searchCharacteristics(namePart, pageSize)
	}

	w.Header().Set("Content-Type", "application/json")

	if len(characteristics) > 0 {
		characteristicsJson, _ := json.Marshal(characteristics)
		w.Write(characteristicsJson)
	} else {
		w.Write([]byte("[]"))
	}
}

type CreateCharacteristicTmplContext struct {
	BaseTmplContext

	Name string
	Unit string

	Error string
}

func characteristicCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateCharacteristicTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "characteristics",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newCharacteristic Characteristic

		newCharacteristic.Name = getFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		newCharacteristic.Unit = getFormString(r, "measurement_unit", &resp.Error, &allGood, &resp.Unit)

		if allGood {
			err := newCharacteristic.dbSave()
			if err != nil {
				log.Println(err)
			}

			http.Redirect(w, r, "/characteristics", 301)
			return
		}
	}

	tmpl, _ := template.ParseFiles("templates/characteristics/create.gohtml", "templates/layout.gohtml")
	err := tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type EditCharacteristicTmplContext struct {
	BaseTmplContext

	Name string
	Unit string

	Error string
}

func characteristicEditHandler(w http.ResponseWriter, r *http.Request) {
	characteristicIdStr := r.PathValue("characteristicId")
	characteristicId, err := strconv.ParseInt(characteristicIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/characteristics", 301)
		return
	}

	characteristic, err := getCharacteristic(characteristicId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown characteristic!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := EditCharacteristicTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "characteristics",
		},
		Name: characteristic.Name,
		Unit: characteristic.Unit,
	}

	if r.Method == "POST" {
		allGood := true

		characteristic.Name = getFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		characteristic.Unit = getFormString(r, "measurement_unit", &resp.Error, &allGood, &resp.Unit)

		if allGood {
			err = characteristic.dbSave()
			if err == nil {
				http.Redirect(w, r, "/characteristics", 301)
				return
			}

			log.Println(err)
			resp.Error += "Database error occurred. "
		}
	}

	tmpl, _ := template.ParseFiles("templates/characteristics/edit.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}

type CharacteristicTmplContext struct {
	BaseTmplContext
	Characteristic Characteristic
	Error          string
}

func characteristicDeleteHandler(w http.ResponseWriter, r *http.Request) {
	characteristicIdStr := r.PathValue("characteristicId")
	characteristicId, err := strconv.ParseInt(characteristicIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/characteristics", 301)
		return
	}

	characteristic, err := getCharacteristic(characteristicId)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(404)
		w.Write([]byte("Unknown characteristic!"))
		return
	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database error occurred!"))
		return
	}

	resp := CharacteristicTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "characteristics",
		},
		Characteristic: characteristic,
		Error:          "",
	}

	if r.Method == "POST" {
		err = characteristic.dbDelete()
		if err == nil {
			http.Redirect(w, r, "/characteristics", 301)
			return
		}

		log.Println(err)
		resp.Error += "Database error occurred. "
	}

	tmpl, _ := template.ParseFiles("templates/characteristics/delete.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}
