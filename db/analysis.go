package db

import (
	"database/sql"
	"time"
)

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
		`SELECT p.id, SUM(oi.quantity) AS total_bought
		FROM order_items oi
			JOIN products p ON p.id = oi.product_id
			JOIN orders o ON oi.order_id = o.id
		WHERE o.created_at >= NOW() - INTERVAL 30 DAY
		GROUP BY p.id
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
		`SELECT AVG(totals.total) FROM (
			SELECT SUM(i.quantity * i.price_per_item) AS total
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

type CountPerDay struct {
	Day   time.Time
	Count int
}

func GetCustomersCountPerDay() ([]CountPerDay, error) {
	rows, err := database.Query(
		`SELECT DATE(created_at) as date, COUNT(DISTINCT customer_id) as num_customers
		FROM orders
		WHERE created_at >= CURDATE() - INTERVAL 30 DAY
		GROUP BY date;`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []CountPerDay

	for rows.Next() {
		var row CountPerDay
		err = rows.Scan(&row.Day, &row.Count)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, nil
}

func GetDayWithMinOrderCount() (time.Time, error) {
	row := database.QueryRow(
		`SELECT day, order_count
		FROM (
			SELECT DATE(created_at) AS day, COUNT(*) AS order_count
			FROM orders
			WHERE created_at >= NOW() - INTERVAL 30 DAY
			GROUP BY day
		) t
		WHERE order_count = (SELECT MIN(order_count) FROM (
			SELECT COUNT(*) AS order_count
			FROM orders
			WHERE created_at >= NOW() - INTERVAL 30 DAY
			GROUP BY DATE(created_at)
		) x)
		LIMIT 1;`,
	)

	var minDay time.Time

	err := row.Scan(&minDay)
	if err != nil {
		return minDay, err
	}

	return minDay, nil
}

func GetDayWithMaxOrderCount() (time.Time, error) {
	row := database.QueryRow(
		`SELECT day, order_count
		FROM (
			SELECT DATE(created_at) AS day, COUNT(*) AS order_count
			FROM orders
			WHERE created_at >= NOW() - INTERVAL 30 DAY
			GROUP BY day
		) t
		WHERE order_count = (SELECT MAX(order_count) FROM (
			SELECT COUNT(*) AS order_count
			FROM orders
			WHERE created_at >= NOW() - INTERVAL 30 DAY
			GROUP BY DATE(created_at)
		) x)
		LIMIT 1;`,
	)

	var minDay time.Time

	err := row.Scan(&minDay)
	if err != nil {
		return minDay, err
	}

	return minDay, nil
}

type FloatStatPerDay struct {
	Day   time.Time
	Value float64
}

func GetAverageOrderTotalPerDay() ([]FloatStatPerDay, error) {
	rows, err := database.Query(
		`SELECT DATE(o.created_at) AS day, AVG(order_total) AS avg_order_total
		FROM (
			SELECT oi.order_id, SUM(oi.price_per_item * oi.quantity) AS order_total
			FROM order_items oi
			GROUP BY oi.order_id
		) totals
			JOIN orders o ON o.id = totals.order_id
		WHERE o.created_at >= NOW() - INTERVAL 30 DAY
		GROUP BY day
		ORDER BY day;`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []FloatStatPerDay

	for rows.Next() {
		var row FloatStatPerDay
		err = rows.Scan(&row.Day, &row.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, nil
}

func GetMedianOrderTotalPerDay() ([]FloatStatPerDay, error) {
	rows, err := database.Query(
		`SELECT day, MEDIAN(order_total) OVER (PARTITION BY day) AS median_order_total
		FROM (
			SELECT DATE(o.created_at) AS day, SUM(oi.price_per_item * oi.quantity) AS order_total
			FROM order_items oi
				JOIN orders o ON o.id = oi.order_id
			WHERE o.created_at >= NOW() - INTERVAL 30 DAY
			GROUP BY oi.order_id, day
		) t
		GROUP BY day
		ORDER BY day;`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []FloatStatPerDay

	for rows.Next() {
		var row FloatStatPerDay
		err = rows.Scan(&row.Day, &row.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, nil
}

func GetMostOrderedProductWithThis(product Product) (Product, int64, error) {
	return getMostLeastOrderedProduct(database.QueryRow(
		`SELECT oi2.product_id, COUNT(oi2.product_id) AS together_count
		FROM order_items oi1
			JOIN order_items oi2 ON oi1.order_id = oi2.order_id AND oi1.product_id <> oi2.product_id
			JOIN orders o ON o.id = oi2.order_id
		WHERE oi1.product_id = ? AND o.created_at >= NOW() - INTERVAL 30 DAY
		GROUP BY oi2.product_id
		ORDER BY together_count DESC
		LIMIT 1;`,
		product.Id,
	))
}

type OrderedProductPair struct {
	Products [2]Product
	Count    int64
}

func getMostLeastOrderedProductPairs(rows *sql.Rows) ([]OrderedProductPair, error) {
	var err error

	var result []OrderedProductPair

	for rows.Next() {
		var row OrderedProductPair
		err = rows.Scan(&row.Products[0].Id, &row.Products[1].Id, &row.Count)
		if err != nil {
			return nil, err
		}

		row.Products[0], err = GetProduct(row.Products[0].Id)
		if err != nil {
			return nil, err
		}

		row.Products[1], err = GetProduct(row.Products[1].Id)
		if err != nil {
			return nil, err
		}

		result = append(result, row)
	}

	return result, nil
}

func GetMostOrderedProductPairs(limit int) ([]OrderedProductPair, error) {
	rows, err := database.Query(
		`SELECT oi1.product_id, oi2.product_id, COUNT(oi2.product_id) AS together_count
		FROM order_items oi1
			JOIN order_items oi2 ON oi1.order_id = oi2.order_id AND oi1.product_id < oi2.product_id
			JOIN orders o ON o.id = oi2.order_id
		WHERE o.created_at >= NOW() - INTERVAL 30 DAY
		GROUP BY oi1.product_id, oi2.product_id
		HAVING together_count > 0
		ORDER BY together_count DESC
		LIMIT ?;`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return getMostLeastOrderedProductPairs(rows)
}

func GetLeastOrderedProductPairs(limit int) ([]OrderedProductPair, error) {
	rows, err := database.Query(
		`SELECT oi1.product_id, oi2.product_id, COUNT(oi2.product_id) AS together_count
		FROM order_items oi1
			JOIN order_items oi2 ON oi1.order_id = oi2.order_id AND oi1.product_id < oi2.product_id
			JOIN orders o ON o.id = oi2.order_id
		WHERE o.created_at >= NOW() - INTERVAL 30 DAY
		GROUP BY oi1.product_id, oi2.product_id
		HAVING together_count > 0
		ORDER BY together_count
		LIMIT ?;`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return getMostLeastOrderedProductPairs(rows)
}
