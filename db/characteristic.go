package db

import (
	"database/sql"
	"strings"
)

type Characteristic struct {
	Id   int64
	Name string
	Unit string
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
