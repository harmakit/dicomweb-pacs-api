package wado

import (
	"bytes"
	"dicom-store-api/fs"
	"dicom-store-api/models"
	"dicom-store-api/utils"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/suyashkumar/dicom"
	"io/ioutil"
	"net/http"
)

// STOWResource implements management handler.
type STOWResource struct {
	DB            *pg.DB
	StudyStore    StudyStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

// NewSTOWResource creates and returns a study resource.
func NewSTOWResource(db *pg.DB, studyStore StudyStore, seriesStore SeriesStore, instanceStore InstanceStore) *STOWResource {
	return &STOWResource{
		DB:            db,
		StudyStore:    studyStore,
		SeriesStore:   seriesStore,
		InstanceStore: instanceStore,
	}
}

type STOWSaveResponse struct {
	Study *models.Study
}

func newSTOWSaveResponse(s *models.Study) *STOWSaveResponse {
	return &STOWSaveResponse{
		Study: s,
	}
}

func (rs *STOWResource) save(w http.ResponseWriter, r *http.Request) {
	const MaxUploadSize = 10 << 20 // 10MB
	if r.ContentLength > MaxUploadSize {
		http.Error(w, "The uploaded image is too big. Please use an image less than 10MB in size", http.StatusBadRequest)
		return
	}
	bodyReader := http.MaxBytesReader(w, r.Body, MaxUploadSize)

	defer bodyReader.Close()

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil || len(body) == 0 {
		http.Error(w, "Wrong request body", http.StatusBadRequest)
		return
	}

	dataset, _ := dicom.Parse(bytes.NewReader(body), MaxUploadSize, nil)

	study := &models.Study{}
	utils.ExtractDicomObjectFromDataset(dataset, study)

	series := &models.Series{Study: study}
	utils.ExtractDicomObjectFromDataset(dataset, series)

	instance := &models.Instance{Series: series}
	utils.ExtractDicomObjectFromDataset(dataset, instance)

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
		if err = rs.StudyStore.Update(study, tx); err != nil {
			tx.Rollback()
			render.Render(w, r, ErrInternalServerError)
			return
		}
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
		if err = rs.SeriesStore.Update(series, tx); err != nil {
			tx.Rollback()
			render.Render(w, r, ErrInternalServerError)
			return
		}
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
		instance = instanceList[0]
		if err = rs.InstanceStore.Update(instance, tx); err != nil {
			tx.Rollback()
			render.Render(w, r, ErrInternalServerError)
			return
		}
	} else {
		instance.SeriesId = series.ID
		instance.Series = series
		if err = rs.InstanceStore.Create(instance, tx); err != nil {
			tx.Rollback()
			render.Render(w, r, ErrInternalServerError)
			return
		}
	}

	path := fs.GetDicomPath(study, series, instance)
	if err = fs.Save(path, body); err != nil {
		tx.Rollback()
		render.Render(w, r, ErrInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		render.Render(w, r, ErrInternalServerError)
		return
	}

	render.Respond(w, r, newSTOWSaveResponse(study))
}
