package handlers

import (
	"go-lb4/db"
	"go-lb4/utils"
	"html/template"
	"log"
	"net/http"
	"time"
)

type ProductWithCount struct {
	Product db.Product
	Count   int64
}

type ProductsAnalysisTmplContext struct {
	utils.BaseTmplContext

	MostOrdered                      ProductWithCount
	LeastOrdered                     ProductWithCount
	AverageOrderTotal                float64
	CustomersPerDay                  []db.FloatStatPerDay
	AvgTotalPerDay                   []db.FloatStatPerDay
	MedTotalPerDay                   []db.FloatStatPerDay
	MinOrdersDay                     db.FloatStatPerDay
	MaxOrdersDay                     db.FloatStatPerDay
	ProductMostCommonWithMostOrdered ProductWithCount
	MostOrderedProductPairs          []db.OrderedProductPair
	LeastOrderedProductPairs         []db.OrderedProductPair
}

func fillMissingDates(counts []db.FloatStatPerDay) []db.FloatStatPerDay {
	now := time.Now().Truncate(24 * time.Hour).UTC()
	start := now.AddDate(0, 0, -30)

	dateSet := make(map[time.Time]float64)
	for _, c := range counts {
		day := c.Day.Truncate(24 * time.Hour)
		dateSet[day] = c.Value
	}

	var filled []db.FloatStatPerDay
	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		if value, exists := dateSet[d]; !exists {
			filled = append(filled, db.FloatStatPerDay{Day: d, Value: 0})
		} else {
			filled = append(filled, db.FloatStatPerDay{Day: d, Value: value})
		}
	}

	return filled
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

	customersPerDay, err := db.GetCustomersCountPerDay()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	avgOrderTotalPerDay, err := db.GetAverageOrderTotalPerDay()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	medOrderTotalPerDay, err := db.GetMedianOrderTotalPerDay()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	minOrdersDay, minOrderCount, err := db.GetDayWithMinOrderCount()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	maxOrdersDay, maxOrderCount, err := db.GetDayWithMaxOrderCount()
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	mostCommonWithMostOrdered, countOrdered, err := db.GetMostOrderedProductWithThis(mostOrdered)
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	mostOrderedPairs, err := db.GetMostOrderedProductPairs(5)
	if utils.ReturnOnDatabaseError(err, w) {
		return
	}

	leastOrderedPairs, err := db.GetLeastOrderedProductPairs(5)
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
		AverageOrderTotal:                averageTotal,
		CustomersPerDay:                  fillMissingDates(customersPerDay),
		AvgTotalPerDay:                   fillMissingDates(avgOrderTotalPerDay),
		MedTotalPerDay:                   fillMissingDates(medOrderTotalPerDay),
		MinOrdersDay:                     db.FloatStatPerDay{Day: minOrdersDay, Value: float64(minOrderCount)},
		MaxOrdersDay:                     db.FloatStatPerDay{Day: maxOrdersDay, Value: float64(maxOrderCount)},
		ProductMostCommonWithMostOrdered: ProductWithCount{Product: mostCommonWithMostOrdered, Count: countOrdered},
		MostOrderedProductPairs:          mostOrderedPairs,
		LeastOrderedProductPairs:         leastOrderedPairs,
	})
	if err != nil {
		log.Println(err)
	}
}
