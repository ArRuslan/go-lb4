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
			return database.Query("SELECT `id`, `model`, `company`, `price` FROM `products` ORDER BY `id` LIMIT ? OFFSET ?;", pageSize, (page-1)*pageSize)
		},
		func(rows *sql.Rows) (Product, error) {
			product := Product{}
			err := rows.Scan(&product.Id, &product.Model, &product.Company, &product.Price)
			return product, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `products`;")
		},
	)
}

func createProduct(product Product) error {
	_, err := database.Exec("INSERT INTO products (`model`, `company`, `price`) VALUES (?, ?, ?);", product.Model, product.Company, product.Price)
	return err
}

func getProduct(productId int) (Product, error) {
	var product Product

	row := database.QueryRow("SELECT `id`, `model`, `company`, `price` FROM products WHERE `id`=?;", productId)
	err := row.Scan(&product.Id, &product.Model, &product.Company, &product.Price)

	return product, err
}

func (product *Product) dbSave() error {
	if product.Id > 0 {
		_, err := database.Exec("UPDATE `products` SET `model`=?, `company`=?, `price`=? WHERE `id`=?;", product.Model, product.Company, product.Price, product.Id)
		return err
	} else {
		return createProduct(*product)
	}
}

func (product *Product) dbDelete() error {
	_, err := database.Exec("DELETE FROM `products` WHERE `id`=?;", product.Id)
	return err
}
