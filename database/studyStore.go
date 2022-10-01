package database

import (
	"dicom-store-api/models"
	"dicom-store-api/utils"
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"reflect"
)

// StudyStore implements database operations for study management.
type StudyStore struct {
	db *pg.DB
}

// NewStudyStore returns a StudyStore implementation.
func NewStudyStore(db *pg.DB) *StudyStore {
	return &StudyStore{
		db: db,
	}
}

func (store *StudyStore) FindBy(fields map[string]any, options *SelectQueryOptions, tx *pg.Tx) ([]*models.Study, error) {
	db := store.GetOrm(tx)
	tableName := (&models.Study{}).GetTableName()

	var result []*models.Study
	query := db.Model(&result)
	for fieldName, fieldValue := range fields {
		structField := reflect.ValueOf(&models.Study{}).Elem().FieldByName(fieldName)
		if !structField.IsValid() {
			return nil, fmt.Errorf("invalid field name: %s", fieldName)
		}
		columnName := utils.ToSnakeCase(fieldName)

		rt := reflect.TypeOf(fieldValue)
		if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
			var values []interface{}
			for i := 0; i < reflect.ValueOf(fieldValue).Len(); i++ {
				values = append(values, reflect.ValueOf(fieldValue).Index(i).Interface())
			}
			query.WhereIn(fmt.Sprintf("%s.%s IN (?)", tableName, columnName), values...)
		} else {
			query.Where(fmt.Sprintf("%s.%s = ?", tableName, columnName), fieldValue)
		}
	}
	options.Apply(query)

	err := query.Select()
	return result, err
}

// Get gets a study by study ID.
func (store *StudyStore) Get(studyID int) (*models.Study, error) {
	study := models.Study{ID: studyID}
	err := store.db.Model(&study).
		Where("id = ?", studyID).
		Select()

	return &study, err
}

// Update updates study.
func (store *StudyStore) Update(study *models.Study, tx *pg.Tx) error {
	db := store.GetOrm(tx)
	_, err := db.Model(study).WherePK().Update()
	return err
}

// Create creates a new study.
func (store *StudyStore) Create(study *models.Study, tx *pg.Tx) error {
	db := store.GetOrm(tx)
	_, err := db.Model(study).Insert()
	return err
}

func (store *StudyStore) GetOrm(tx *pg.Tx) orm.DB {
	if tx != nil {
		return tx
	} else {
		return store.db
	}
}
