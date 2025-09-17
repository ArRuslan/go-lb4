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

type CharacteristicsListTmplContext struct {
	Characteristics []Characteristic
	Count           int
}

func characteristicsListHandler(w http.ResponseWriter, r *http.Request) {
	characteristics, count, err := getCharacteristics(getPageAndSize(r))

	tmpl, _ := template.ParseFiles("templates/characteristics/list.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, CharacteristicsListTmplContext{
		Characteristics: characteristics,
		Count:           count,
	})
	if err != nil {
		log.Println(err)
	}
}

type CreateCharacteristicTmplContext struct {
	Name string
	Unit string

	Error string
}

func characteristicCreateHandler(w http.ResponseWriter, r *http.Request) {
	var resp CreateCharacteristicTmplContext

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
	Name string
	Unit string

	Error string
}

func characteristicEditHandler(w http.ResponseWriter, r *http.Request) {
	characteristicIdStr := r.PathValue("characteristicId")
	characteristicId, err := strconv.Atoi(characteristicIdStr)
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
	Characteristic Characteristic
	Error          string
}

func characteristicDeleteHandler(w http.ResponseWriter, r *http.Request) {
	characteristicIdStr := r.PathValue("characteristicId")
	characteristicId, err := strconv.Atoi(characteristicIdStr)
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
