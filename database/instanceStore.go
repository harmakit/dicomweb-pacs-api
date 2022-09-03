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

func (store *InstanceStore) FindByFields(fields map[string]any, tx *pg.Tx) ([]*models.Instance, error) {
	db := store.GetOrm(tx)

	var result []*models.Instance
	query := db.Model(&result)
	for fieldName, fieldValue := range fields {
		structField := reflect.ValueOf(&models.Instance{}).Elem().FieldByName(fieldName)
		if !structField.IsValid() {
			return nil, fmt.Errorf("invalid field name: %s", fieldName)
		}
		columnName := utils.ToSnakeCase(fieldName)
		query.Where(fmt.Sprintf("%s = ?", columnName), fieldValue)
	}
	query.Relation("Series")

	err := query.Select()

	return result, err
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
func (store *InstanceStore) Update(instance *models.Instance) error {
	_, err := store.db.Model(instance).WherePK().Update()
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
