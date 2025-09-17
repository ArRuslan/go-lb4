package main

import (
	"database/sql"
	"fmt"
	"log"
)

var database *sql.DB

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

func getProducts(page, pageSize int) ([]Product, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				p.id, p.model, p.manufacturer, p.price, p.quantity, COALESCE(p.image_url, ''), p.warranty_days,
    				COALESCE(c.id, 0), COALESCE(c.name, ''), COALESCE(c.description, '')
				FROM products p 
				LEFT OUTER JOIN categories c ON p.category_id = c.id
				ORDER BY p.id LIMIT ? OFFSET ?;`,
				pageSize, (page-1)*pageSize,
			)
		},
		func(rows *sql.Rows) (Product, error) {
			product := Product{}
			err := rows.Scan(
				&product.Id, &product.Model, &product.Manufacturer, &product.Price, &product.Quantity, &product.ImageUrl, &product.WarrantyDays,
				&product.Category.Id, &product.Category.Name, &product.Category.Description,
			)
			return product, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `products`;")
		},
	)
}

func createProduct(product Product) error {
	var imageUrl sql.NullString
	if product.ImageUrl == "" {
		imageUrl = sql.NullString{}
	} else {
		imageUrl = sql.NullString{String: product.ImageUrl, Valid: true}
	}

	var categoryId sql.NullInt64
	if product.Category.Id == 0 {
		categoryId = sql.NullInt64{}
	} else {
		categoryId = sql.NullInt64{Int64: product.Category.Id, Valid: true}
	}

	_, err := database.Exec(
		`INSERT INTO products (model, manufacturer, price, quantity, image_url, warranty_days, category_id) 
		VALUES (?, ?, ?, ?, ?, ?, ?);`,
		product.Model, product.Manufacturer, product.Price, product.Quantity, imageUrl, product.WarrantyDays, categoryId,
	)
	return err
}

func getProduct(productId int) (Product, error) {
	var product Product

	row := database.QueryRow(
		`SELECT 
    		p.id, p.model, p.manufacturer, p.price, p.quantity, COALESCE(p.image_url, ''), p.warranty_days,
    		COALESCE(c.id, 0), COALESCE(c.name, ''), COALESCE(c.description, '')
		FROM products p 
		LEFT OUTER JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?;`,
		productId,
	)
	err := row.Scan(
		&product.Id, &product.Model, &product.Manufacturer, &product.Price, &product.Quantity, &product.ImageUrl, &product.WarrantyDays,
		&product.Category.Id, &product.Category.Name, &product.Category.Description,
	)

	return product, err
}

func (product *Product) dbSave() error {
	if product.Id > 0 {
		var imageUrl sql.NullString
		if product.ImageUrl == "" {
			imageUrl = sql.NullString{}
		} else {
			imageUrl = sql.NullString{String: product.ImageUrl, Valid: true}
		}

		var categoryId sql.NullInt64
		if product.Category.Id == 0 {
			categoryId = sql.NullInt64{}
		} else {
			categoryId = sql.NullInt64{Int64: product.Category.Id, Valid: true}
		}

		_, err := database.Exec(
			`UPDATE products 
			SET model=?, manufacturer=?, price=?, quantity=?, image_url=?, warranty_days=?, category_id=?
			WHERE id=?;`,
			product.Model, product.Manufacturer, product.Price, product.Quantity, imageUrl, product.WarrantyDays, categoryId, product.Id,
		)
		return err
	}

	return createProduct(*product)
}

func (product *Product) dbDelete() error {
	_, err := database.Exec("DELETE FROM `products` WHERE `id`=?;", product.Id)
	return err
}
