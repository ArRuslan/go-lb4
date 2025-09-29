package handlers

import (
	"go-lb4/db"
	"go-lb4/utils"
	"html/template"
	"log"
	"net/http"
)

type ProductWithCount struct {
	Product db.Product
	Count   int64
}

type ProductsAnalysisTmplContext struct {
	utils.BaseTmplContext

	MostOrdered       ProductWithCount
	LeastOrdered      ProductWithCount
	AverageOrderTotal float64
}

func ProductsAnalysisHandler(w http.ResponseWriter, _ *http.Request) {
	mostOrdered, mostCount, err := db.GetMostOrderedProduct()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	leastOrdered, leastCount, err := db.GetLeastOrderedProduct()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	averageTotal, err := db.GetOrdersAverageTotal()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	tmpl, _ := template.ParseFiles("templates/analysis.gohtml", "templates/layout.gohtml")
	err = tmpl.Execute(w, ProductsAnalysisTmplContext{
		BaseTmplContext: utils.BaseTmplContext{
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
