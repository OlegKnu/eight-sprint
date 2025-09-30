package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключение к БД
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// get
	stored, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, parcel.Client, stored.Client)
	assert.Equal(t, parcel.Status, stored.Status)
	assert.Equal(t, parcel.Address, stored.Address)

	// delete
	err = store.Delete(id)
	assert.NoError(t, err)

	_, err = store.Get(id)
	assert.Error(t, err) // посылка должна быть удалена
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	// add
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	assert.NoError(t, err)

	// check
	stored, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, stored.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	// add
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// set status
	err = store.SetStatus(id, ParcelStatusSent)
	assert.NoError(t, err)

	// check
	stored, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, stored.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		assert.NoError(t, err)
		assert.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	assert.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		assert.Equal(t, expected.Client, parcel.Client)
		assert.Equal(t, expected.Status, parcel.Status)
		assert.Equal(t, expected.Address, parcel.Address)
	}
}
