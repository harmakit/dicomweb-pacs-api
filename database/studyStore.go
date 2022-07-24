package database

import (
	"dicom-store-api/models"
	"github.com/go-pg/pg"
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

func (s *StudyStore) FindByPatient(patient string) ([]*models.Study, error) {
	var studies []*models.Study
	err := s.db.Model(&studies).
		Where("patient = ?", patient).
		Select()

	return studies, err
}

// Get gets a study by study ID.
func (s *StudyStore) Get(studyID int) (*models.Study, error) {
	study := models.Study{ID: studyID}
	err := s.db.Model(&study).
		Where("id = ?", studyID).
		Select()

	return &study, err
}

// Update updates study.
func (s *StudyStore) Update(study *models.Study) error {
	err := s.db.Update(study)
	return err
}
