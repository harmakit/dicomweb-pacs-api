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

// Get gets a study by study ID.
func (s *StudyStore) Get(studyID int) (*models.Study, error) {
	p := models.Study{StudyId: studyID}
	_, err := s.db.Model(&p).
		Where("study_id = ?", studyID).
		SelectOrInsert()

	return &p, err
}

// Update updates study.
func (s *StudyStore) Update(p *models.Study) error {
	err := s.db.Update(p)
	return err
}
