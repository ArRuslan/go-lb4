package main

import (
	"html/template"
	"log"
	"net/http"
)

type ProductWithCount struct {
	Product Product
	Count   int64
}

type ProductsAnalysisTmplContext struct {
	BaseTmplContext

	MostOrdered       ProductWithCount
	LeastOrdered      ProductWithCount
	AverageOrderTotal float64
}

func productsAnalysisHandler(w http.ResponseWriter, _ *http.Request) {
	mostOrdered, mostCount, err := getMostOrderedProduct()
	if returnOnDatabaseError(err, w) {
		return
	}

	leastOrdered, leastCount, err := getLeastOrderedProduct()
	if returnOnDatabaseError(err, w) {
		return
	}

	averageTotal, err := getOrdersAverageTotal()
	if returnOnDatabaseError(err, w) {
		return
	}

	tmpl, _ := template.ParseFiles("templates/analysis.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, ProductsAnalysisTmplContext{
		BaseTmplContext: BaseTmplContext{
			Type: "analysis",
		},
		MostOrdered: ProductWithCount{
			Product: mostOrdered,
			Count:   mostCount,
		},
		LeastOrdered: ProductWithCount{
			Product: leastOrdered,
			Count:   leastCount,
		},
		AverageOrderTotal: averageTotal,
	})
	if err != nil {
		log.Println(err)
	}
}
