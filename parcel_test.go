package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляется новая посылка в БД
	id, err := store.Add(parcel)
	// отсутствии ошибки
	require.NoError(t, err)
	// наличие идентификатора
	assert.NotEmpty(t, id)

	// get
	// получение только что добавленной посылки
	obj, err := store.Get(id)
	// отсутствии ошибки
	require.NoError(t, err)
	// проверка, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	assert.Equal(t, obj.Client, parcel.Client)
	assert.Equal(t, obj.Status, parcel.Status)
	assert.Equal(t, obj.Address, parcel.Address)
	assert.Equal(t, obj.CreatedAt, parcel.CreatedAt)

	// delete
	// удаление добавленной посылки
	err = store.Delete(id)
	// отсутствии ошибки
	require.NoError(t, err)
	// проверка, что посылку больше нельзя получить из БД
	obj, err = store.Get(id)
	require.Error(t, err)
	require.Empty(t, obj)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавляется новая посылка в БД
	id, err := store.Add(parcel)
	// отсутствие ошибки
	require.NoError(t, err)
	// наличие идентификатора
	assert.NotEmpty(t, id)

	// set address
	// обновляется адрес
	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)
	// отсутствии ошибки
	require.NoError(t, err)

	// check
	// получает добавленную посылку
	obj, err := store.Get(id)
	require.NoError(t, err)
	// адрес обновился
	assert.Equal(t, obj.Address, newAddress)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляется новая посылка в БД
	id, err := store.Add(parcel)
	// отсутствие ошибки
	require.NoError(t, err)
	// наличии идентификатора
	assert.NotEmpty(t, id)
	// set status
	// обновяется статус
	err = store.SetStatus(id, ParcelStatusSent)
	// отсутствии ошибки
	require.NoError(t, err)

	// check
	// получает добавленную посылку
	obj, err := store.Get(id)
	require.NoError(t, err)
	// статус обновился
	assert.Equal(t, obj.Status, ParcelStatusSent)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
	}
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
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // список посылок по идентификатору клиента, сохранённого в переменной client
	// отсутствии ошибки
	require.NoError(t, err)
	// количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(parcels), len(storedParcels))
	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// все посылки из storedParcels есть в parcelMap
		// значения полей полученных посылок заполнены верно
		val, ok := parcelMap[parcel.Number]
		if ok {
			assert.Equal(t, val, parcel)
		}
	}
}
