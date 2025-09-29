package db

import (
	"database/sql"
	"errors"
	"strings"
)

type Customer struct {
	Id        int64
	FirstName string
	LastName  string
	Email     string
}

func GetCustomers(page, pageSize int) ([]Customer, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				c.id, c.first_name, c.last_name, c.email
				FROM customers c
				ORDER BY c.id LIMIT ? OFFSET ?;`,
				pageSize, (page-1)*pageSize,
			)
		},
		func(rows *sql.Rows) (Customer, error) {
			customer := Customer{}
			err := rows.Scan(&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email)
			return customer, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `customers`;")
		},
	)
}

func SearchCustomersByEmail(emailPart string, limit int) ([]Customer, error) {
	customers, _, err := getRowsAndCount(
		1,
		limit,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				c.id, c.first_name, c.last_name, c.email
				FROM customers c
				WHERE LOWER(c.email) like ?
				ORDER BY c.id LIMIT ?;`,
				"%"+strings.ToLower(emailPart)+"%", pageSize,
			)
		},
		func(rows *sql.Rows) (Customer, error) {
			customer := Customer{}
			err := rows.Scan(&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email)
			return customer, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT 0;")
		},
	)

	return customers, err
}

func CreateCustomer(customer Customer) error {
	_, err := database.Exec(
		"INSERT INTO customers (first_name, last_name, email) VALUES (?, ?, ?);",
		customer.FirstName, customer.LastName, customer.Email,
	)
	return err
}

func GetCustomer(customerId int) (Customer, error) {
	var customer Customer

	row := database.QueryRow(
		"SELECT c.id, c.first_name, c.last_name, c.email FROM customers c WHERE c.id = ?;",
		customerId,
	)
	err := row.Scan(
		&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email,
	)

	return customer, err
}

func GetCustomerByEmail(email string) (Customer, error) {
	var customer Customer

	row := database.QueryRow(
		"SELECT c.id, c.first_name, c.last_name, c.email FROM customers c WHERE c.email = ?;",
		email,
	)
	err := row.Scan(
		&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email,
	)

	return customer, err
}

func (customer *Customer) DbSave() error {
	if customer.Id == 0 {
		row := database.QueryRow("SELECT c.id FROM customers c WHERE c.email = ?;", customer.Email)
		err := row.Scan(&customer.Id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	if customer.Id > 0 {
		_, err := database.Exec(
			"UPDATE customers SET first_name=?, last_name=?, email=? WHERE id=?;",
			customer.FirstName, customer.LastName, customer.Email, customer.Id,
		)
		return err
	}

	return CreateCustomer(*customer)
}

func (customer *Customer) DbDelete() error {
	_, err := database.Exec("DELETE FROM `customers` WHERE `id`=?;", customer.Id)
	return err
}
