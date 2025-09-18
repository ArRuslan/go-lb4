package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
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

func getCategories(page, pageSize int) ([]Category, int, error) {
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

func createCategory(category Category) error {
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

func getCategory(categoryId int) (Category, error) {
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

func searchCategories(namePart string, limit int) ([]Category, error) {
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

func (category *Category) dbSave() error {
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

	return createCategory(*category)
}

func (category *Category) dbDelete() error {
	_, err := database.Exec("DELETE FROM `categories` WHERE `id`=?;", category.Id)
	return err
}

func getCharacteristics(page, pageSize int) ([]Characteristic, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				c.id, c.name, COALESCE(c.measurement_unit, '')
				FROM characteristics c
				ORDER BY c.id LIMIT ? OFFSET ?;`,
				pageSize, (page-1)*pageSize,
			)
		},
		func(rows *sql.Rows) (Characteristic, error) {
			characteristic := Characteristic{}
			err := rows.Scan(
				&characteristic.Id, &characteristic.Name, &characteristic.Unit,
			)
			return characteristic, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `characteristics`;")
		},
	)
}

func createCharacteristic(characteristic Characteristic) error {
	var unit sql.NullString
	if characteristic.Unit == "" {
		unit = sql.NullString{}
	} else {
		unit = sql.NullString{String: characteristic.Unit, Valid: true}
	}

	_, err := database.Exec(
		"INSERT INTO characteristics (name, measurement_unit) VALUES (?, ?);",
		characteristic.Name, unit,
	)
	return err
}

func getCharacteristic(characteristicId int) (Characteristic, error) {
	var characteristic Characteristic

	row := database.QueryRow(
		"SELECT c.id, c.name, COALESCE(c.measurement_unit, '') FROM characteristics c WHERE c.id = ?;",
		characteristicId,
	)
	err := row.Scan(
		&characteristic.Id, &characteristic.Name, &characteristic.Unit,
	)

	return characteristic, err
}

func (characteristic *Characteristic) dbSave() error {
	if characteristic.Id > 0 {
		var unit sql.NullString
		if characteristic.Unit == "" {
			unit = sql.NullString{}
		} else {
			unit = sql.NullString{String: characteristic.Unit, Valid: true}
		}

		_, err := database.Exec(
			"UPDATE characteristics SET name=?, measurement_unit=? WHERE id=?;",
			characteristic.Name, unit, characteristic.Id,
		)
		return err
	}

	return createCharacteristic(*characteristic)
}

func (characteristic *Characteristic) dbDelete() error {
	_, err := database.Exec("DELETE FROM `characteristics` WHERE `id`=?;", characteristic.Id)
	return err
}

func getCustomers(page, pageSize int) ([]Customer, int, error) {
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
			err := rows.Scan(
				&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email,
			)
			return customer, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `customers`;")
		},
	)
}

func createCustomer(customer Customer) error {
	_, err := database.Exec(
		"INSERT INTO customers (first_name, last_name, email) VALUES (?, ?, ?);",
		customer.FirstName, customer.LastName, customer.Email,
	)
	return err
}

func getCustomer(customerId int) (Customer, error) {
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

func getCustomerByEmail(email string) (Customer, error) {
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

func (customer *Customer) dbSave() error {
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

	return createCustomer(*customer)
}

func (customer *Customer) dbDelete() error {
	_, err := database.Exec("DELETE FROM `customers` WHERE `id`=?;", customer.Id)
	return err
}

func getOrders(page, pageSize int) ([]Order, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				o.id, o.address,
    				COALESCE(c.id, 0), COALESCE(c.first_name, ''), COALESCE(c.last_name, ''), COALESCE(c.email, '')
				FROM orders o 
				LEFT OUTER JOIN customers c ON o.customer_id = c.id
				ORDER BY o.id LIMIT ? OFFSET ?;`,
				pageSize, (page-1)*pageSize,
			)
		},
		func(rows *sql.Rows) (Order, error) {
			order := Order{}
			err := rows.Scan(
				&order.Id, &order.Address,
				&order.Customer.Id, &order.Customer.FirstName, &order.Customer.LastName, &order.Customer.Email,
			)
			return order, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `orders`;")
		},
	)
}

func createOrder(order Order) error {
	if order.Customer.Email != "" {
		err := order.Customer.dbSave()
		if err != nil {
			return err
		}
		if order.Customer.Id == 0 {
			customerByEmail, err := getCustomerByEmail(order.Customer.Email)
			if err != nil {
				return err
			}
			order.Customer = customerByEmail
		}
	}

	var customerId sql.NullInt64
	if order.Customer.Id == 0 {
		customerId = sql.NullInt64{}
	} else {
		customerId = sql.NullInt64{Int64: order.Customer.Id, Valid: true}
	}

	_, err := database.Exec(
		"INSERT INTO orders (address, customer_id) VALUES (?, ?);",
		order.Address, customerId,
	)
	return err
}

func getOrder(orderId int) (Order, error) {
	var order Order

	row := database.QueryRow(
		`SELECT 
    		o.id, o.address,
    		COALESCE(c.id, 0), COALESCE(c.first_name, ''), COALESCE(c.last_name, ''), COALESCE(c.email, '')
		FROM orders o 
		LEFT OUTER JOIN customers c ON o.customer_id = c.id
		WHERE o.id = ?;`,
		orderId,
	)
	err := row.Scan(
		&order.Id, &order.Address,
		&order.Customer.Id, &order.Customer.FirstName, &order.Customer.LastName, &order.Customer.Email,
	)

	return order, err
}

func (order *Order) dbSave() error {
	if order.Id > 0 {
		var customerId sql.NullInt64
		if order.Customer.Id == 0 {
			customerId = sql.NullInt64{}
		} else {
			customerId = sql.NullInt64{Int64: order.Customer.Id, Valid: true}
		}

		_, err := database.Exec(
			`UPDATE orders 
			SET address=?, customer_id=?
			WHERE id=?;`,
			order.Address, customerId, order.Id,
		)
		return err
	}

	return createOrder(*order)
}

func (order *Order) dbDelete() error {
	_, err := database.Exec("DELETE FROM `orders` WHERE `id`=?;", order.Id)
	return err
}
