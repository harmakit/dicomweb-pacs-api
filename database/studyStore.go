package database

import (
	"dicom-store-api/models"
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

func (store *StudyStore) FindBy(s *models.Study, fields map[string]any, tx *pg.Tx) ([]*models.Study, error) {
	for fieldName := range fields {
		studyField := reflect.ValueOf(s).Elem().FieldByName(fieldName)
		if !studyField.IsValid() {
			return nil, fmt.Errorf("invalid field name: %s", fieldName)
		}
	}

	db := store.GetOrm(tx)

	var studies []*models.Study
	query := db.Model(&studies)
	for fieldName, fieldValue := range fields {
		query = query.Where(fmt.Sprintf("%s = ?", fieldName), fieldValue)
	}
	err := query.Select()

	return studies, err
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
func (store *StudyStore) Update(study *models.Study) error {
	_, err := store.db.Model(study).WherePK().Update()
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
