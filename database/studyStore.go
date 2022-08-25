package database

import (
	"dicom-store-api/models"
	"github.com/go-pg/pg"
	DicomTag "github.com/suyashkumar/dicom/pkg/tag"
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

func (s *StudyStore) FindByTags(tags []*DicomTag.Tag) ([]*models.Study, error) {
	//sTag := models.Study{}.GetObjectIdFieldTag()
	//info, _ := tag.Find(sTag)
	//info.Name

	//val := reflect.ValueOf(models.Study{})
	//for i := 0; i < val.Type().NumField(); i++ {
	//	fmt.Println(val.Type().Field(i).Tag.Get("json"))
	//}

	var studies []*models.Study
	err := s.db.Model(&studies).
		//Where("patient = ?", patient).
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
	_, err := s.db.Model(study).WherePK().Update()
	return err
}
