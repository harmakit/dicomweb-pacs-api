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

type StudiesStore interface {
	FindBy(s *models.Study, fields map[string]any, tx *pg.Tx) error
	Create(s *models.Study, tx *pg.Tx) error
}
type SeriesStore interface {
	Create(s *models.Series, tx *pg.Tx) error
}
type InstanceStore interface {
	Create(s *models.Instance, tx *pg.Tx) error
}

// StudiesResource implements management handler.
type StudiesResource struct {
	DB            *pg.DB
	StudyStore    StudiesStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

// NewStudiesResource creates and returns a study resource.
func NewStudiesResource(db *pg.DB, studiesStore StudiesStore, seriesStore SeriesStore, instanceStore InstanceStore) *StudiesResource {
	return &StudiesResource{
		DB:            db,
		StudyStore:    studiesStore,
		SeriesStore:   seriesStore,
		InstanceStore: instanceStore,
	}
}

func (rs *StudiesResource) router() *chi.Mux {
	r := chi.NewRouter()
	//r.Use(rs.studiesCtx)
	r.Post("/", rs.save)
	return r
}

//func (rs *StudiesResource) studiesCtx(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		ctx := context.WithValue(r.Context(), ctxStudy, nil)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	})
//}

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
	//s := r.Context().Value(ctxStudy).(*models.Study)
	//data := &studiesSaveRequest{}
	//if err := render.Bind(r, data); err != nil {
	//	render.Render(w, r, ErrInvalidRequest(err))
	//}
	//file, _, err := r.FormFile("receipt") // r is *http.Request
	//switch err {
	//case nil:
	//case http.ErrMissingFile:
	//	fmt.Println("no file")
	//	render.Render(w, r, ErrInternalServerError)
	//	return
	//default:
	//	fmt.Println(err)
	//	render.Render(w, r, ErrInternalServerError)
	//	return
	//}
	//var buff bytes.Buffer
	//fileSize, err := buff.ReadFrom(file)
	//if err != nil {
	//	//log(r).WithField("profileCtx", claims.Sub).Error(err)
	//	render.Render(w, r, ErrInternalServerError)
	//	return
	//}
	//fmt.Println(fileSize) // this will return you a file size.

	const MaxUploadSize = 10 << 20 // 10MB
	//if r.ContentLength > MaxUploadSize {
	//	http.Error(w, "The uploaded image is too big. Please use an image less than 10MB in size", http.StatusBadRequest)
	//	return
	//}
	bodyReader := http.MaxBytesReader(w, r.Body, MaxUploadSize)

	defer bodyReader.Close()

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dataset, _ := dicom.Parse(bytes.NewReader(body), MaxUploadSize, nil)

	fmt.Println(dataset)

	study := models.Study{}
	ExtractDicomObjectFromDataset(dataset, &study)

	series := models.Series{Study: &study}
	ExtractDicomObjectFromDataset(dataset, &series)

	instance := models.Instance{Series: &series}
	ExtractDicomObjectFromDataset(dataset, &instance)

	// todo check if study exists...
	studies := rs.StudyStore.FindBy(&study, map[string]any{
		"StudyInstanceUID": study.StudyInstanceUID,
	}, nil)
	fmt.Println(studies)

	tx, err := rs.DB.Begin()

	if err = rs.StudyStore.Create(&study, tx); err != nil {
		tx.Rollback()
		render.Render(w, r, ErrInternalServerError)
		return
	}

	series.StudyId = study.ID
	series.Study = &study
	study.Series = append(study.Series, &series)
	// todo check if series exists...
	if err = rs.SeriesStore.Create(&series, tx); err != nil {
		tx.Rollback()
		render.Render(w, r, ErrInternalServerError)
		return
	}

	instance.SeriesId = series.ID
	instance.Series = &series
	series.Instances = append(series.Instances, &instance)
	// todo check if instance exists...
	if err = rs.InstanceStore.Create(&instance, tx); err != nil {
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

	render.Respond(w, r, newStudiesSaveResponse(&study))
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
