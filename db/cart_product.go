package db

import (
	"database/sql"

	"github.com/google/uuid"
)

type CartProduct struct {
	Id       int64
	CartId   uuid.UUID
	Product  Product
	Quantity int
}

func GetCartProducts(cartId uuid.UUID) ([]CartProduct, int, error) {
	return getRowsAndCount(
		1,
		0,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT 
    				i.id, i.cart_id, i.quantity,
    				p.id, p.model, p.manufacturer, p.price, p.quantity, p.warranty_days, COALESCE(p.image_url, '')
				FROM cart_products i 
				LEFT OUTER JOIN products p ON i.product_id = p.id
				WHERE i.cart_id = ?
				ORDER BY i.id;`,
				cartId,
			)
		},
		func(rows *sql.Rows) (CartProduct, error) {
			item := CartProduct{}
			err := rows.Scan(
				&item.Id, &item.CartId, &item.Quantity,
				&item.Product.Id, &item.Product.Model, &item.Product.Manufacturer, &item.Product.Price, &item.Product.Quantity, &item.Product.WarrantyDays, &item.Product.ImageUrl,
			)
			return item, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `cart_products` WHERE cart_id=?;", cartId)
		},
	)
}

func GetCartProduct(itemId int64, cartId uuid.UUID) (CartProduct, error) {
	var item CartProduct

	row := database.QueryRow(
		`SELECT 
    		i.id, i.cart_id, i.quantity,
    		p.id, p.model, p.manufacturer, p.price, p.quantity, p.warranty_days, COALESCE(p.image_url, '')
		FROM cart_products i
		LEFT OUTER JOIN products p ON i.product_id = p.id
		WHERE i.id = ? AND i.cart_id = ?;`,
		itemId, cartId,
	)
	err := row.Scan(
		&item.Id, &item.CartId, &item.Quantity,
		&item.Product.Id, &item.Product.Model, &item.Product.Manufacturer, &item.Product.Price, &item.Product.Quantity, &item.Product.WarrantyDays, &item.Product.ImageUrl,
	)

	return item, err
}

func GetCartProductByProductId(productId int64, cartId uuid.UUID) (CartProduct, error) {
	var item CartProduct

	row := database.QueryRow(
		`SELECT 
    		i.id, i.cart_id, i.quantity,
    		p.id, p.model, p.manufacturer, p.price, p.quantity, p.warranty_days, COALESCE(p.image_url, '')
		FROM cart_products i
		LEFT OUTER JOIN products p ON i.product_id = p.id
		WHERE i.product_id = ? AND i.cart_id = ?;`,
		productId, cartId,
	)
	err := row.Scan(
		&item.Id, &item.CartId, &item.Quantity,
		&item.Product.Id, &item.Product.Model, &item.Product.Manufacturer, &item.Product.Price, &item.Product.Quantity, &item.Product.WarrantyDays, &item.Product.ImageUrl,
	)

	return item, err
}

func CreateCartProduct(item CartProduct) error {
	_, err := database.Exec(
		`INSERT INTO cart_products (cart_id, product_id, quantity) 
		VALUES (?, ?, ?);`,
		item.CartId, item.Product.Id, item.Quantity,
	)
	return err
}

func (item *CartProduct) DbSave() error {
	if item.Id > 0 {
		_, err := database.Exec(
			`UPDATE cart_products SET quantity=? WHERE id=?;`,
			item.Quantity, item.Id,
		)
		return err
	}

	return CreateCartProduct(*item)
}

func (item *CartProduct) DbDelete() error {
	_, err := database.Exec("DELETE FROM `cart_products` WHERE `id`=?;", item.Id)
	return err
}
