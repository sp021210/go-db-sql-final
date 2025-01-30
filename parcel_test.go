package main

import (
	"database/sql"
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
	require.NoError(t, err)

	defer db.Close() // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	parcelId, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelId)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel

	storedParcels, err := store.Get(parcelId)
	require.NoError(t, err)

	parcel.Number = parcelId
	assert.Equal(t, storedParcels, parcel)
	assert.Equal(t, storedParcels.Client, parcel.Client)
	assert.Equal(t, storedParcels.Status, parcel.Status)
	assert.Equal(t, storedParcels.Address, parcel.Address)
	assert.Equal(t, storedParcels.CreatedAt, parcel.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД

	err = store.Delete(parcelId)
	require.NoError(t, err)

	_, err = store.Get(parcelId)
	require.Error(t, err)
	assert.ErrorIs(t, sql.ErrNoRows, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err) // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	parcelId, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelId)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcelId, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился

	storedParcels, err := store.Get(parcelId)
	require.NoError(t, err)
	assert.Equal(t, storedParcels.Address, parcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err) // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	parcelId, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelId)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки

	err = store.SetStatus(parcelId, ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился

	storedParcels, err := store.Get(parcelId)
	require.NoError(t, err)
	assert.Equal(t, storedParcels.Status, ParcelStatusSent)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err) // настройте подключение к БД

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
		require.NotEmpty(t, id) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err)                         // убедитесь в отсутствии ошибки
	assert.Len(t, storedParcels, len(parcels))      // убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		assert.Equal(t, parcel, parcelMap[parcel.Number])
		assert.Equal(t, parcel.Number, parcelMap[parcel.Number].Number)
		assert.Equal(t, parcel.Client, parcelMap[parcel.Number].Client)   // в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		assert.Equal(t, parcel.Status, parcelMap[parcel.Number].Status)   // убедитесь, что все посылки из storedParcels есть в parcelMap
		assert.Equal(t, parcel.Address, parcelMap[parcel.Number].Address) // убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, parcel.CreatedAt, parcelMap[parcel.Number].CreatedAt)
	}
}
