package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	Id             uuid.UUID
	LastAccessTime time.Time
}

// TODO: Do i really need this ?
func GetCarts(page, pageSize int) ([]Cart, int, error) {
	return getRowsAndCount(
		page,
		pageSize,
		func(page, pageSize int) (*sql.Rows, error) {
			return database.Query(
				`SELECT o.id, o.last_access_time
				FROM carts o 
				ORDER BY o.id LIMIT ? OFFSET ?;`,
				pageSize, (page-1)*pageSize,
			)
		},
		func(rows *sql.Rows) (Cart, error) {
			cart := Cart{}
			err := rows.Scan(&cart.Id, &cart.LastAccessTime)
			return cart, err
		},
		func() *sql.Row {
			return database.QueryRow("SELECT COUNT(*) FROM `carts`;")
		},
	)
}

func GetCart(cartId uuid.UUID) (Cart, error) {
	var cart Cart

	row := database.QueryRow(
		`SELECT c.id, c.last_access_time
		FROM carts c 
		WHERE c.id = ?;`,
		cartId,
	)
	err := row.Scan(&cart.Id, &cart.LastAccessTime)

	return cart, err
}

func GetOrCreateCart(cartId uuid.UUID) (Cart, error) {
	cart, err := GetCart(cartId)
	if errors.Is(err, sql.ErrNoRows) {
		return Cart{
			Id:             cartId,
			LastAccessTime: time.Now(),
		}, nil
	}

	if err != nil {
		return Cart{}, err
	}

	return cart, nil
}

func (cart *Cart) DbSave() error {
	_, err := database.Exec(
		`INSERT INTO carts (id, last_access_time) VALUES (?, ?) ON DUPLICATE KEY UPDATE last_access_time=?;`,
		cart.Id, cart.LastAccessTime, cart.LastAccessTime,
	)
	return err
}

func (cart *Cart) DbDelete() error {
	_, err := database.Exec("DELETE FROM `carts` WHERE `id`=?;", cart.Id)
	return err
}
