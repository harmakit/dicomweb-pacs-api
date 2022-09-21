package database

import (
	"dicom-store-api/models"
	"dicom-store-api/utils"
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"reflect"
)

// SeriesStore implements database operations for series management.
type SeriesStore struct {
	db *pg.DB
}

// NewSeriesStore returns a SeriesStore implementation.
func NewSeriesStore(db *pg.DB) *SeriesStore {
	return &SeriesStore{
		db: db,
	}
}

func (store *SeriesStore) FindByFields(fields map[string]any, tx *pg.Tx) ([]*models.Series, error) {
	db := store.GetOrm(tx)

	var result []*models.Series
	query := db.Model(&result)
	for fieldName, fieldValue := range fields {
		structField := reflect.ValueOf(&models.Series{}).Elem().FieldByName(fieldName)
		if !structField.IsValid() {
			return nil, fmt.Errorf("invalid field name: %s", fieldName)
		}
		columnName := utils.ToSnakeCase(fieldName)
		query.Where(fmt.Sprintf("%s = ?", columnName), fieldValue)
	}
	query.Relation("Study")

	err := query.Select()

	return result, err
}

// Get gets a series by series ID.
func (store *SeriesStore) Get(seriesID int) (*models.Series, error) {
	series := models.Series{ID: seriesID}
	err := store.db.Model(&series).
		Where("id = ?", seriesID).
		Select()

	return &series, err
}

// Update updates series.
func (store *SeriesStore) Update(series *models.Series, tx *pg.Tx) error {
	db := store.GetOrm(tx)
	_, err := db.Model(series).WherePK().Update()
	return err
}

// Create creates a new series.
func (store *SeriesStore) Create(series *models.Series, tx *pg.Tx) error {
	db := store.GetOrm(tx)
	_, err := db.Model(series).Insert()
	return err
}

func (store *SeriesStore) GetOrm(tx *pg.Tx) orm.DB {
	if tx != nil {
		return tx
	} else {
		return store.db
	}
}
