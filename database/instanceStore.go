package database

import (
	"dicom-store-api/models"
	"github.com/go-pg/pg"
	DicomTag "github.com/suyashkumar/dicom/pkg/tag"
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

func (store *InstanceStore) FindByTags(tags []*DicomTag.Tag) ([]*models.Instance, error) {
	//sTag := models.Instance{}.GetObjectIdFieldTag()
	//info, _ := tag.Find(sTag)
	//info.Name

	//val := reflect.ValueOf(models.Instance{})
	//for store := 0; store < val.Type().NumField(); store++ {
	//	fmt.Println(val.Type().Field(store).Tag.Get("json"))
	//}

	var studies []*models.Instance
	err := store.db.Model(&studies).
		//Where("patient = ?", patient).
		Select()

	return studies, err
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
func (store *InstanceStore) Create(instance *models.Instance) error {
	_, err := store.db.Model(instance).Insert()
	return err
}
