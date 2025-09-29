package db

import (
	"database/sql"
	"strings"
)

type Category struct {
	Id          int64
	Name        string
	Description string
}

func GetCategories(page, pageSize int) ([]Category, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				c.id, c.name, COALESCE(c.description, '')
				FROM categories c
				ORDER BY c.id LIMIT ? OFFSET ?;`,
				pageSize, (page-1)*pageSize,
			)
		},
		func(rows *sql.Rows) (Category, error) {
			category := Category{}
			err := rows.Scan(
				&category.Id, &category.Name, &category.Description,
			)
			return category, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `categories`;")
		},
	)
}

func CreateCategory(category Category) error {
	var description sql.NullString
	if category.Description == "" {
		description = sql.NullString{}
	} else {
		description = sql.NullString{String: category.Description, Valid: true}
	}

	_, err := database.Exec(
		"INSERT INTO categories (name, description) VALUES (?, ?);",
		category.Name, description,
	)
	return err
}

func GetCategory(categoryId int) (Category, error) {
	var category Category

	row := database.QueryRow(
		"SELECT c.id, c.name, COALESCE(c.description, '') FROM categories c WHERE c.id = ?;",
		categoryId,
	)
	err := row.Scan(
		&category.Id, &category.Name, &category.Description,
	)

	return category, err
}

func SearchCategories(namePart string, limit int) ([]Category, error) {
	categories, _, err := getRowsAndCount(
		1,
		limit,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				c.id, c.name, COALESCE(c.description, '')
				FROM categories c
				WHERE LOWER(c.name) like ?
				ORDER BY c.id LIMIT ?;`,
				"%"+strings.ToLower(namePart)+"%", pageSize,
			)
		},
		func(rows *sql.Rows) (Category, error) {
			category := Category{}
			err := rows.Scan(&category.Id, &category.Name, &category.Description)
			return category, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT 0;")
		},
	)

	return categories, err
}

func (category *Category) DbSave() error {
	if category.Id > 0 {
		var description sql.NullString
		if category.Description == "" {
			description = sql.NullString{}
		} else {
			description = sql.NullString{String: category.Description, Valid: true}
		}

		_, err := database.Exec(
			"UPDATE categories SET name=?, description=? WHERE id=?;",
			category.Name, description, category.Id,
		)
		return err
	}

	return CreateCategory(*category)
}

func (category *Category) DbDelete() error {
	_, err := database.Exec("DELETE FROM `categories` WHERE `id`=?;", category.Id)
	return err
}
