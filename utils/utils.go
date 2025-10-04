package utils

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func GetPageAndSize(req *http.Request) (int, int) {
	pageMaybe := req.URL.Query().Get("page")
	page, err := strconv.Atoi(pageMaybe)
	if err != nil {
		page = 1
	}

	pageSizeMaybe := req.URL.Query().Get("pageSize")
	pageSize, err := strconv.Atoi(pageSizeMaybe)
	if err != nil {
		pageSize = 10
	}

	return page, min(pageSize, 100)
}

func GetFormStringNonEmpty(req *http.Request, name string, errorText *string, valid *bool, out *string) string {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	if value == "" {
		if errorText != nil {
			*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		}
		*valid = false

		return value
	}

	return value
}

func GetFormString(req *http.Request, name string, _ *string, valid *bool, out *string) string {
	if err := req.ParseForm(); err != nil {
		log.Printf("Failed to parse form: %s\n", err)
		*valid = false
		return ""
	}

	if !req.PostForm.Has(name) && !req.URL.Query().Has(name) {
		*valid = false
		return ""
	}

	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	return value
}

func GetFormInt(req *http.Request, name string, errorText *string, valid *bool, out *string) int {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		if errorText != nil {
			*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		}
		*valid = false

		return 0
	}

	return valueInt
}

func GetFormInt64(req *http.Request, name string, errorText *string, valid *bool, out *string) int64 {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		if errorText != nil {
			*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		}
		*valid = false

		return 0
	}

	return valueInt
}

func GetFormDouble(req *http.Request, name string, errorText *string, valid *bool, out *string) float64 {
	value := req.FormValue(name)

	if out != nil {
		*out = value
	}

	valueDouble, err := strconv.ParseFloat(value, 64)
	if err != nil {
		if errorText != nil {
			*errorText += fmt.Sprintf("\"%s\" is empty or invalid. ", name)
		}
		*valid = false

		return 0
	}

	return valueDouble
}

func ReturnOnDatabaseError(err error, w http.ResponseWriter) bool {
	if err == nil {
		return false
	}

	log.Println(err)
	w.WriteHeader(500)
	w.Write([]byte("Database error occurred!"))
	return true
}

func GetCartId(r *http.Request) uuid.UUID {
	cartIdCookie, err := r.Cookie("cartId")
	var cartId uuid.UUID
	if err != nil {
		cartId = uuid.New()
	} else {
		cartId, err = uuid.Parse(cartIdCookie.Value)
		if err != nil {
			cartId = uuid.New()
		}
	}

	return cartId
}
