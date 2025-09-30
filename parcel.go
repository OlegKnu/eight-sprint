package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add — добавление новой посылки в таблицу
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`,
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get — получение посылки по её номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`,
		number,
	)
	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

// GetByClient — получение всех посылок клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at FROM parcel WHERE client = ? ORDER BY number`,
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return parcels, nil
}

// SetStatus — обновление статуса посылки
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		`UPDATE parcel SET status = ? WHERE number = ?`,
		status, number,
	)
	return err
}

// SetAddress — изменение адреса доставки (только если статус = registered)
func (s ParcelStore) SetAddress(number int, address string) error {
	res, err := s.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ? AND status = ?`,
		address, number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("посылка не найдена")
	}
	return nil
}

// Delete — удаление посылки (только если статус = registered)
func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec(
		`DELETE FROM parcel WHERE number = ? AND status = ?`,
		number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("посылка не найдена")
	}
	return nil
}
