package database

import (
	"dicom-store-api/models"
	"dicom-store-api/utils"
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"reflect"
)

// InstanceStore implements database operations for instance management.
type InstanceStore struct {
	db *pg.DB
}

// NewInstanceStore returns a InstanceStore implementation.
func NewInstanceStore(db *pg.DB) *InstanceStore {
	return &InstanceStore{
		db: db,
	}
}

func (store *InstanceStore) FindBy(fields map[string]any, options *SelectQueryOptions, tx *pg.Tx) ([]*models.Instance, error) {
	db := store.GetOrm(tx)
	tableName := (&models.Instance{}).GetTableName()

	var result []*models.Instance
	query := db.Model(&result)
	for fieldName, fieldValue := range fields {
		structField := reflect.ValueOf(&models.Instance{}).Elem().FieldByName(fieldName)
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
	query.Relation("Series")
	query.Relation("Series.Study")
	options.Apply(query)

	err := query.Select()

	return result, err
}

func (store *InstanceStore) CountBy(fields map[string]any, tx *pg.Tx) (int, error) {
	db := store.GetOrm(tx)
	tableName := (&models.Instance{}).GetTableName()

	var count int
	query := db.Model(&models.Instance{}).ColumnExpr("count(*)")
	for fieldName, fieldValue := range fields {
		structField := reflect.ValueOf(&models.Instance{}).Elem().FieldByName(fieldName)
		if !structField.IsValid() {
			return 0, fmt.Errorf("invalid field name: %s", fieldName)
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
	_, err := query.SelectAndCount(&count)

	return count, err
}

// Get gets an instance by instance ID.
func (store *InstanceStore) Get(instanceID int) (*models.Instance, error) {
	instance := models.Instance{ID: instanceID}
	err := store.db.Model(&instance).
		Where("id = ?", instanceID).
		Select()

	return &instance, err
}

// Update updates instance.
func (store *InstanceStore) Update(instance *models.Instance, tx *pg.Tx) error {
	db := store.GetOrm(tx)
	_, err := db.Model(instance).WherePK().Update()
	return err
}

// Create creates a new instance.
func (store *InstanceStore) Create(instance *models.Instance, tx *pg.Tx) error {
	db := store.GetOrm(tx)
	_, err := db.Model(instance).Insert()
	return err
}

func (store *InstanceStore) GetOrm(tx *pg.Tx) orm.DB {
	if tx != nil {
		return tx
	} else {
		return store.db
	}
}
