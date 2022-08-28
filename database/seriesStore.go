package database

import (
	"dicom-store-api/models"
	"github.com/go-pg/pg"
	DicomTag "github.com/suyashkumar/dicom/pkg/tag"
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

func (store *SeriesStore) FindByTags(tags []*DicomTag.Tag) ([]*models.Series, error) {
	//sTag := models.Series{}.GetObjectIdFieldTag()
	//info, _ := tag.Find(sTag)
	//info.Name

	//val := reflect.ValueOf(models.Series{})
	//for i := 0; i < val.Type().NumField(); i++ {
	//	fmt.Println(val.Type().Field(i).Tag.Get("json"))
	//}

	var series []*models.Series
	err := store.db.Model(&series).
		//Where("patient = ?", patient).
		Select()

	return series, err
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
func (store *SeriesStore) Update(series *models.Series) error {
	_, err := store.db.Model(series).WherePK().Update()
	return err
}

// Create creates a new series.
func (store *SeriesStore) Create(series *models.Series) error {
	_, err := store.db.Model(series).Insert()
	return err
}
