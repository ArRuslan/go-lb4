package db

import (
	"context"
	"database/sql"
	"time"
)

type Order struct {
	Id        int64
	Customer  Customer
	CreatedAt time.Time
	Address   string
	Status    string
}

func GetOrders(page, pageSize int) ([]Order, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				o.id, o.created_at, o.address, o.status,
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
				&order.Id, &order.CreatedAt, &order.Address, &order.Status,
				&order.Customer.Id, &order.Customer.FirstName, &order.Customer.LastName, &order.Customer.Email,
			)
			return order, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `orders`;")
		},
	)
}

func CreateOrder(ctx context.Context, order *Order, tx *sql.Tx) error {
	var dbExec func(context.Context, string, ...any) (sql.Result, error)

	if tx == nil {
		dbExec = database.ExecContext
	} else {
		dbExec = tx.ExecContext
	}

	if order.Customer.Email != "" {
		err := order.Customer.DbSave(ctx, tx)
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

	result, err := dbExec(
		ctx,
		"INSERT INTO orders (address, customer_id, status) VALUES (?, ?, ?);",
		order.Address, customerId, order.Status,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	order.Id = id
	return nil
}

func GetOrder(orderId int) (Order, error) {
	var order Order

	row := database.QueryRow(
		`SELECT 
    		o.id, o.created_at, o.address, o.status,
    		COALESCE(c.id, 0), COALESCE(c.first_name, ''), COALESCE(c.last_name, ''), COALESCE(c.email, '')
		FROM orders o 
		LEFT OUTER JOIN customers c ON o.customer_id = c.id
		WHERE o.id = ?;`,
		orderId,
	)
	err := row.Scan(
		&order.Id, &order.CreatedAt, &order.Address, &order.Status,
		&order.Customer.Id, &order.Customer.FirstName, &order.Customer.LastName, &order.Customer.Email,
	)

	return order, err
}

func (order *Order) DbSave(ctx context.Context, tx *sql.Tx) error {
	var dbExec func(context.Context, string, ...any) (sql.Result, error)

	if tx == nil {
		dbExec = database.ExecContext
	} else {
		dbExec = tx.ExecContext
	}

	if order.Id > 0 {
		var customerId sql.NullInt64
		if order.Customer.Id == 0 {
			customerId = sql.NullInt64{}
		} else {
			customerId = sql.NullInt64{Int64: order.Customer.Id, Valid: true}
		}

		_, err := dbExec(
			ctx,
			`UPDATE orders SET address=?, customer_id=?, status=? WHERE id=?;`,
			order.Address, customerId, order.Status, order.Id,
		)
		return err
	}

	return CreateOrder(ctx, order, tx)
}

func (order *Order) DbDelete() error {
	_, err := database.Exec("DELETE FROM `orders` WHERE `id`=?;", order.Id)
	return err
}
