package db

import "database/sql"

type ProductCharacteristic struct {
	Id             int64
	ProductId      int64
	Characteristic Characteristic
	Value          string
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
