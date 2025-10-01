package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

var database *sql.DB

func InitDatabase(driver, dsn string) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic(err)
	}
	database = db
}

func CloseDatabase() {
	database.Close()
}

func getRowsAndCount[T any](page, pageSize int, getRows func(int, int) (*sql.Rows, error), scanRow func(*sql.Rows) (T, error), getCount func() *sql.Row) ([]T, int, error) {
	var objects []T

	rows, err := getRows(page, pageSize)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		log.Println(err)
		return objects, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var object T
		object, err = scanRow(rows)
		if err != nil {
			fmt.Println(err)
			continue
		}
		objects = append(objects, object)
	}

	var count int
	row := getCount()
	err = row.Scan(&count)
	if err != nil {
		log.Println(err)
		return objects, 0, err
	}

	return objects, count, nil
}

func getMostLeastOrderedProduct(row *sql.Row) (Product, int64, error) {
	var productId, count int64

	err := row.Scan(&productId, &count)
	if err != nil {
		return Product{}, 0, err
	}

	var product Product
	product, err = GetProduct(productId)
	if err != nil {
		return Product{}, 0, err
	}

	return product, count, nil
}

func GetMostOrderedProduct() (Product, int64, error) {
	return getMostLeastOrderedProduct(database.QueryRow(
		`SELECT p.id, sum(i.quantity) AS total_bought
		FROM products p
			INNER JOIN order_items i ON p.id = i.product_id
		GROUP BY p.id
		HAVING total_bought > 0
		ORDER BY total_bought DESC
		LIMIT 1;`,
	))
}

func GetLeastOrderedProduct() (Product, int64, error) {
	return getMostLeastOrderedProduct(database.QueryRow(
		`SELECT p.id, sum(i.quantity) AS total_bought
		FROM products p
			INNER JOIN order_items i ON p.id = i.product_id
		GROUP BY p.id
		HAVING total_bought > 0
		ORDER BY total_bought
		LIMIT 1;`,
	))
}

func GetOrdersAverageTotal() (float64, error) {
	row := database.QueryRow(
		`SELECT avg(totals.total) FROM (
			SELECT o.id, sum(i.quantity * i.price_per_item) AS total
			FROM orders o
				INNER JOIN order_items i ON o.id = i.order_id
			GROUP BY o.id
		) AS totals;`,
	)

	var averageTotal float64

	err := row.Scan(&averageTotal)
	if err != nil {
		return 0, err
	}

	return averageTotal, nil
}

func BeginTx(ctx context.Context) (*sql.Tx, error) {
	return database.BeginTx(ctx, nil)
}
