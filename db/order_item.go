package db

import (
	"context"
	"database/sql"
)

type OrderItem struct {
	Id           int64
	OrderId      int64
	Product      Product
	Quantity     int
	PricePerItem float64
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

func CreateOrderItem(ctx context.Context, item OrderItem, tx *sql.Tx) error {
	var dbExec func(context.Context, string, ...any) (sql.Result, error)

	if tx == nil {
		dbExec = database.ExecContext
	} else {
		dbExec = tx.ExecContext
	}

	_, err := dbExec(
		ctx,
		`INSERT INTO order_items (order_id, product_id, quantity, price_per_item) 
		VALUES (?, ?, ?, ?);`,
		item.OrderId, item.Product.Id, item.Quantity, item.PricePerItem,
	)
	return err
}

func (item *OrderItem) DbSave(ctx context.Context, tx *sql.Tx) error {
	var dbExec func(context.Context, string, ...any) (sql.Result, error)

	if tx == nil {
		dbExec = database.ExecContext
	} else {
		dbExec = tx.ExecContext
	}

	if item.Id > 0 {
		_, err := dbExec(
			ctx,
			`UPDATE order_items SET quantity=? WHERE id=?;`,
			item.Quantity, item.Id,
		)
		return err
	}

	return CreateOrderItem(ctx, *item, tx)
}

func (item *OrderItem) DbDelete() error {
	_, err := database.Exec("DELETE FROM `order_items` WHERE `id`=?;", item.Id)
	return err
}
