package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
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

func GetProducts(page, pageSize int) ([]Product, int, error) {
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

func SearchProducts(model string, limit int) ([]Product, error) {
	products, _, err := getRowsAndCount(
		1,
		limit,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				p.id, p.model, p.manufacturer, p.price, p.quantity, COALESCE(p.image_url, ''), p.warranty_days,
    				COALESCE(c.id, 0), COALESCE(c.name, ''), COALESCE(c.description, '')
				FROM products p 
				LEFT OUTER JOIN categories c ON p.category_id = c.id
				WHERE LOWER(p.model) LIKE ?
				ORDER BY p.id LIMIT ?;`,
				"%"+model+"%", pageSize,
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
			return database.QueryRow("SELECT 0;")
		},
	)
	return products, err
}

func CreateProduct(product Product) error {
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

func GetProduct(productId int64) (Product, error) {
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

func (product *Product) DbSave() error {
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

	return CreateProduct(*product)
}

func (product *Product) DbDelete() error {
	_, err := database.Exec("DELETE FROM `products` WHERE `id`=?;", product.Id)
	return err
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

func GetCharacteristics(page, pageSize int) ([]Characteristic, int, error) {
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

func SearchCharacteristics(namePart string, limit int) ([]Characteristic, error) {
	characteristics, _, err := getRowsAndCount(
		1,
		limit,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				c.id, c.name, COALESCE(c.measurement_unit, '')
				FROM characteristics c
				WHERE LOWER(c.name) LIKE ?
				ORDER BY c.id LIMIT ?;`,
				"%"+strings.ToLower(namePart)+"%", pageSize,
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
			return database.QueryRow("SELECT 0;")
		},
	)

	return characteristics, err
}

func CreateCharacteristic(characteristic Characteristic) error {
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

func GetCharacteristic(characteristicId int64) (Characteristic, error) {
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

func (characteristic *Characteristic) DbSave() error {
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

	return CreateCharacteristic(*characteristic)
}

func (characteristic *Characteristic) DbDelete() error {
	_, err := database.Exec("DELETE FROM `characteristics` WHERE `id`=?;", characteristic.Id)
	return err
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

func GetOrders(page, pageSize int) ([]Order, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				o.id, o.created_at, o.address,
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
				&order.Id, &order.CreatedAt, &order.Address,
				&order.Customer.Id, &order.Customer.FirstName, &order.Customer.LastName, &order.Customer.Email,
			)
			return order, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `orders`;")
		},
	)
}

func CreateOrder(order Order) error {
	if order.Customer.Email != "" {
		err := order.Customer.DbSave()
		if err != nil {
			return err
		}
		if order.Customer.Id == 0 {
			customerByEmail, err := GetCustomerByEmail(order.Customer.Email)
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

func GetOrder(orderId int) (Order, error) {
	var order Order

	row := database.QueryRow(
		`SELECT 
    		o.id, o.created_at, o.address,
    		COALESCE(c.id, 0), COALESCE(c.first_name, ''), COALESCE(c.last_name, ''), COALESCE(c.email, '')
		FROM orders o 
		LEFT OUTER JOIN customers c ON o.customer_id = c.id
		WHERE o.id = ?;`,
		orderId,
	)
	err := row.Scan(
		&order.Id, &order.CreatedAt, &order.Address,
		&order.Customer.Id, &order.Customer.FirstName, &order.Customer.LastName, &order.Customer.Email,
	)

	return order, err
}

func (order *Order) DbSave() error {
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

	return CreateOrder(*order)
}

func (order *Order) DbDelete() error {
	_, err := database.Exec("DELETE FROM `orders` WHERE `id`=?;", order.Id)
	return err
}

func GetProductCharacteristics(productId int64) ([]ProductCharacteristic, int, error) {
	return getRowsAndCount(
		1,
		0,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				p.id, p.product_id, p.value,
    				c.id, c.name, COALESCE(c.measurement_unit, '')
				FROM product_characteristics p 
				LEFT OUTER JOIN characteristics c ON p.characteristic_id = c.id
				WHERE p.product_id = ?
				ORDER BY p.id;`,
				productId,
			)
		},
		func(rows *sql.Rows) (ProductCharacteristic, error) {
			char := ProductCharacteristic{}
			err := rows.Scan(
				&char.Id, &char.ProductId, &char.Value,
				&char.Characteristic.Id, &char.Characteristic.Name, &char.Characteristic.Unit,
			)
			return char, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `product_characteristics` WHERE product_id=?;", productId)
		},
	)
}

func GetProductCharacteristic(characteristicId, productId int) (ProductCharacteristic, error) {
	var char ProductCharacteristic

	row := database.QueryRow(
		`SELECT 
    		p.id, p.product_id, p.value,
    		c.id, c.name, COALESCE(c.measurement_unit, '')
		FROM product_characteristics p 
		LEFT OUTER JOIN characteristics c ON p.characteristic_id = c.id
		WHERE p.id = ? AND p.product_id = ?;`,
		characteristicId, productId,
	)
	err := row.Scan(
		&char.Id, &char.ProductId, &char.Value,
		&char.Characteristic.Id, &char.Characteristic.Name, &char.Characteristic.Unit,
	)

	return char, err
}

func CreateProductCharacteristic(char ProductCharacteristic) error {
	_, err := database.Exec(
		`INSERT INTO product_characteristics (product_id, characteristic_id, value) 
		VALUES (?, ?, ?);`,
		char.ProductId, char.Characteristic.Id, char.Value,
	)
	return err
}

func (char *ProductCharacteristic) DbSave() error {
	if char.Id > 0 {
		_, err := database.Exec(
			`UPDATE product_characteristics 
			SET value=?
			WHERE id=?;`,
			char.Value, char.Id,
		)
		return err
	}

	return CreateProductCharacteristic(*char)
}

func (char *ProductCharacteristic) DbDelete() error {
	_, err := database.Exec("DELETE FROM `product_characteristics` WHERE `id`=?;", char.Id)
	return err
}

func GetOrderItems(orderId int64) ([]OrderItem, int, error) {
	return getRowsAndCount(
		1,
		0,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				i.id, i.order_id, i.quantity, i.price_per_item,
    				p.id, p.model, p.manufacturer, p.price, p.quantity, p.warranty_days, COALESCE(p.image_url, '')
				FROM order_items i 
				LEFT OUTER JOIN products p ON i.product_id = p.id
				WHERE i.order_id = ?
				ORDER BY i.id;`,
				orderId,
			)
		},
		func(rows *sql.Rows) (OrderItem, error) {
			item := OrderItem{}
			err := rows.Scan(
				&item.Id, &item.OrderId, &item.Quantity, &item.PricePerItem,
				&item.Product.Id, &item.Product.Model, &item.Product.Manufacturer, &item.Product.Price, &item.Product.Quantity, &item.Product.WarrantyDays, &item.Product.ImageUrl,
			)
			return item, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `order_items` WHERE order_id=?;", orderId)
		},
	)
}

func GetOrderItem(itemId, orderId int) (OrderItem, error) {
	var item OrderItem

	row := database.QueryRow(
		`SELECT 
    		i.id, i.order_id, i.quantity, i.price_per_item,
    		p.id, p.model, p.manufacturer, p.price, p.quantity, p.warranty_days, COALESCE(p.image_url, '')
		FROM order_items i
		LEFT OUTER JOIN products p ON i.product_id = p.id
		WHERE i.id = ? AND i.order_id = ?;`,
		itemId, orderId,
	)
	err := row.Scan(
		&item.Id, &item.OrderId, &item.Quantity, &item.PricePerItem,
		&item.Product.Id, &item.Product.Model, &item.Product.Manufacturer, &item.Product.Price, &item.Product.Quantity, &item.Product.WarrantyDays, &item.Product.ImageUrl,
	)

	return item, err
}

func CreateOrderItem(item OrderItem) error {
	_, err := database.Exec(
		`INSERT INTO order_items (order_id, product_id, quantity, price_per_item) 
		VALUES (?, ?, ?, ?);`,
		item.OrderId, item.Product.Id, item.Quantity, item.PricePerItem,
	)
	return err
}

func (item *OrderItem) DbSave() error {
	if item.Id > 0 {
		_, err := database.Exec(
			`UPDATE order_items SET quantity=? WHERE id=?;`,
			item.Quantity, item.Id,
		)
		return err
	}

	return CreateOrderItem(*item)
}

func (item *OrderItem) DbDelete() error {
	_, err := database.Exec("DELETE FROM `order_items` WHERE `id`=?;", item.Id)
	return err
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
