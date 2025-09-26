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
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("address can be changed only for registered parcels")
	}
	_, err = s.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ?`,
		address, number,
	)
	return err
}

// Delete — удаление посылки (только если статус = registered)
func (s ParcelStore) Delete(number int) error {
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("only registered parcels can be deleted")
	}
	_, err = s.db.Exec(
		`DELETE FROM parcel WHERE number = ?`,
		number,
	)
	return err
}
