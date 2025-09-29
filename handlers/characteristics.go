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

type CharacteristicsListTmplContext struct {
	utils.BaseTmplContext

	Characteristics []db.Characteristic
	Pagination      utils.PaginationInfo
}

func CharacteristicsListHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.GetPageAndSize(r)
	characteristics, count, err := db.GetCharacteristics(page, pageSize)

	tmpl := template.New("list.gohtml")
	_, err = tmpl.Funcs(utils.TmplPaginationFuncs).ParseFiles("templates/characteristics/list.gohtml", "templates/layout.gohtml", "templates/pagination.gohtml")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, CharacteristicsListTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "characteristics",
		},
		Characteristics: characteristics,
		Pagination: utils.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Count:    count,
			UrlPath:  "/characteristics",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func CharacteristicsSearchHandler(w http.ResponseWriter, r *http.Request) {
	var characteristics []db.Characteristic

	namePart := r.URL.Query().Get("name")
	if namePart != "" {
		_, pageSize := utils.GetPageAndSize(r)
		characteristics, _ = db.SearchCharacteristics(namePart, pageSize)
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
	utils.BaseTmplContext

	Name string
	Unit string

	Error string
}

func CharacteristicCreateHandler(w http.ResponseWriter, r *http.Request) {
	resp := CreateCharacteristicTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
			Type: "characteristics",
		},
	}

	if r.Method == "POST" {
		allGood := true
		var newCharacteristic db.Characteristic

		newCharacteristic.Name = utils.GetFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		newCharacteristic.Unit = utils.GetFormString(r, "measurement_unit", &resp.Error, &allGood, &resp.Unit)

		if allGood {
			err := newCharacteristic.DbSave()
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
	utils.BaseTmplContext

	Name string
	Unit string

	Error string
}

func CharacteristicEditHandler(w http.ResponseWriter, r *http.Request) {
	characteristicIdStr := r.PathValue("characteristicId")
	characteristicId, err := strconv.ParseInt(characteristicIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/characteristics", 301)
		return
	}

	characteristic, err := db.GetCharacteristic(characteristicId)
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
		BaseTmplContext: utils.BaseTmplContext{
			Type: "characteristics",
		},
		Name: characteristic.Name,
		Unit: characteristic.Unit,
	}

	if r.Method == "POST" {
		allGood := true

		characteristic.Name = utils.GetFormStringNonEmpty(r, "name", &resp.Error, &allGood, &resp.Name)
		characteristic.Unit = utils.GetFormString(r, "measurement_unit", &resp.Error, &allGood, &resp.Unit)

		if allGood {
			err = characteristic.DbSave()
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
	utils.BaseTmplContext
	Characteristic db.Characteristic
	Error          string
}

func CharacteristicDeleteHandler(w http.ResponseWriter, r *http.Request) {
	characteristicIdStr := r.PathValue("characteristicId")
	characteristicId, err := strconv.ParseInt(characteristicIdStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/characteristics", 301)
		return
	}

	characteristic, err := db.GetCharacteristic(characteristicId)
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
		BaseTmplContext: utils.BaseTmplContext{
			Type: "characteristics",
		},
		Characteristic: characteristic,
		Error:          "",
	}

	if r.Method == "POST" {
		err = characteristic.DbDelete()
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
