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

func BeginTx(ctx context.Context) (*sql.Tx, error) {
	return database.BeginTx(ctx, nil)
}
