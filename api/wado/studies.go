package wado

import (
	"bytes"
	"dicom-store-api/fs"
	"dicom-store-api/models"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"io/ioutil"
	"net/http"
	"reflect"
)

// ErrStudiesValidation defines the list of error types returned from study resource.
var (
	ErrStudiesValidation = errors.New("studies validation error")
)

// StudiesResource implements management handler.
type StudiesResource struct {
	DB            *pg.DB
	StudyStore    StudyStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

// NewStudiesResource creates and returns a study resource.
func NewStudiesResource(db *pg.DB, studyStore StudyStore, seriesStore SeriesStore, instanceStore InstanceStore) *StudiesResource {
	return &StudiesResource{
		DB:            db,
		StudyStore:    studyStore,
		SeriesStore:   seriesStore,
		InstanceStore: instanceStore,
	}
}

func (rs *StudiesResource) router() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", rs.save)
	return r
}

type studiesSaveRequest struct {
	*models.Study
	ProtectedID int `json:"id"`
}

func (d *studiesSaveRequest) Bind(r *http.Request) error {
	return nil
}

type studiesSaveResponse struct {
	Study *models.Study
}

func newStudiesSaveResponse(s *models.Study) *studiesSaveResponse {
	return &studiesSaveResponse{
		Study: s,
	}
}

func (rs *StudiesResource) save(w http.ResponseWriter, r *http.Request) {
	const MaxUploadSize = 10 << 20 // 10MB
	if r.ContentLength > MaxUploadSize {
		http.Error(w, "The uploaded image is too big. Please use an image less than 10MB in size", http.StatusBadRequest)
		return
	}
	bodyReader := http.MaxBytesReader(w, r.Body, MaxUploadSize)

	defer bodyReader.Close()

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dataset, _ := dicom.Parse(bytes.NewReader(body), MaxUploadSize, nil)

	fmt.Println(dataset)

	study := &models.Study{}
	ExtractDicomObjectFromDataset(dataset, study)

	series := &models.Series{Study: study}
	ExtractDicomObjectFromDataset(dataset, series)

	instance := &models.Instance{Series: series}
	ExtractDicomObjectFromDataset(dataset, instance)

	tx, err := rs.DB.Begin()

	studyList, err := rs.StudyStore.FindByFields(map[string]any{
		"StudyInstanceUID": study.StudyInstanceUID,
	}, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	if len(studyList) == 1 {
		study = studyList[0]
	} else {
		if err = rs.StudyStore.Create(study, tx); err != nil {
			tx.Rollback()
			render.Render(w, r, ErrInternalServerError)
			return
		}
	}

	seriesList, err := rs.SeriesStore.FindByFields(map[string]any{
		"SeriesInstanceUID": series.SeriesInstanceUID,
	}, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	if len(seriesList) == 1 {
		series = seriesList[0]
	} else {
		series.StudyId = study.ID
		series.Study = study
		if err = rs.SeriesStore.Create(series, tx); err != nil {
			tx.Rollback()
			render.Render(w, r, ErrInternalServerError)
			return
		}
	}

	instanceList, err := rs.InstanceStore.FindByFields(map[string]any{
		"SOPInstanceUID": instance.SOPInstanceUID,
	}, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	if len(instanceList) == 1 {
		render.Render(w, r, ErrBadRequest) // instance already exists
		return
	}

	instance.SeriesId = series.ID
	instance.Series = series
	if err = rs.InstanceStore.Create(instance, tx); err != nil {
		tx.Rollback()
		render.Render(w, r, ErrInternalServerError)
		return
	}

	path := fs.GetDicomPath(study, series, instance)
	if err = fs.Save(path, body); err != nil {
		tx.Rollback()
		render.Render(w, r, ErrInternalServerError)
		return
	}

	tx.Commit()

	render.Respond(w, r, newStudiesSaveResponse(study))
}

func ExtractDicomObjectFromDataset(dataset dicom.Dataset, object models.DicomObject) {
	reflection := reflect.TypeOf(object).Elem()

	for i := 0; i < reflection.NumField(); i++ {
		field := reflection.Field(i)
		tagInfo, err := tag.FindByName(field.Tag.Get("dicom"))
		if err != nil {
			continue
		}
		element, _ := dataset.FindElementByTag(tagInfo.Tag)
		if element == nil {
			continue
		}
		if element.Value.ValueType() != 0 {
			panic(fmt.Sprintf("field %s is not a string type", field.Name))
		}

		var stringDatasetValue string
		if tagInfo.VR == "SQ" {
			stringDatasetValue = element.Value.String()
		} else {
			stringDatasetValue = element.Value.GetValue().([]string)[0]
		}
		reflect.ValueOf(object).Elem().FieldByIndex(field.Index).SetString(stringDatasetValue)
	}
}
