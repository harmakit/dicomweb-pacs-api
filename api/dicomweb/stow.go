package dicomweb

import (
	"bytes"
	"dicom-store-api/fs"
	"dicom-store-api/models"
	"dicom-store-api/utils"
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/suyashkumar/dicom"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

// STOWResource implements management handler.
type STOWResource struct {
	DB            *pg.DB
	StudyStore    StudyStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

// NewSTOWResource creates and returns a STOWResource.
func NewSTOWResource(db *pg.DB, studyStore StudyStore, seriesStore SeriesStore, instanceStore InstanceStore) *STOWResource {
	return &STOWResource{
		DB:            db,
		StudyStore:    studyStore,
		SeriesStore:   seriesStore,
		InstanceStore: instanceStore,
	}
}

func (rs *STOWResource) save(w http.ResponseWriter, r *http.Request) {
	const MaxUploadSize = 128 << 20
	if r.ContentLength > MaxUploadSize {
		http.Error(w, fmt.Sprintf("request exceeds max upload size of %d bytes", MaxUploadSize), http.StatusBadRequest)
		return
	}

	contentType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var files [][]byte
	if contentType == "application/dicom" {
		bodyReader := http.MaxBytesReader(w, r.Body, MaxUploadSize)

		defer bodyReader.Close()

		body, err := ioutil.ReadAll(bodyReader)
		if err != nil || len(body) == 0 {
			http.Error(w, "Wrong request body", http.StatusBadRequest)
			return
		}
		files = append(files, body)
	} else {
		if !strings.HasPrefix(contentType, "multipart/") {
			http.Error(w, "expecting a multipart message", http.StatusBadRequest)
			return
		}
		multipartReader := multipart.NewReader(r.Body, params["boundary"])

		for {
			part, err := multipartReader.NextPart()
			if err != nil {
				break
			}

			if part.Header.Get("Content-Type") != "application/dicom" {
				http.Error(w, "expecting a multipart message of application/dicom content", http.StatusBadRequest)
				return
			}

			fileBytes, err := ioutil.ReadAll(part)
			if err != nil {
				part.Close()
				http.Error(w, "failed to read content of the part", http.StatusInternalServerError)
				return
			}
			files = append(files, fileBytes)

			part.Close()
		}
	}

	if len(files) == 0 {
		http.Error(w, "no files found in the request", http.StatusBadRequest)
		return
	}

	var successfullySavedFiles = make(map[int]bool)
	for index, fileBytes := range files {
		successfullySavedFiles[index] = false
		dataset, _ := dicom.Parse(bytes.NewReader(fileBytes), MaxUploadSize, nil)

		study := &models.Study{}
		utils.ExtractDicomObjectFromDataset(dataset, study)

		series := &models.Series{Study: study}
		utils.ExtractDicomObjectFromDataset(dataset, series)

		instance := &models.Instance{Series: series}
		utils.ExtractDicomObjectFromDataset(dataset, instance)

		tx, err := rs.DB.Begin()

		studyList, err := rs.StudyStore.FindBy(map[string]any{
			"StudyInstanceUID": study.StudyInstanceUID,
		}, nil, nil)
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

		seriesList, err := rs.SeriesStore.FindBy(map[string]any{
			"SeriesInstanceUID": series.SeriesInstanceUID,
		}, nil, nil)
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

		instanceList, err := rs.InstanceStore.FindBy(map[string]any{
			"SOPInstanceUID": instance.SOPInstanceUID,
		}, nil, nil)
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
		if err = fs.Save(path, fileBytes); err != nil {
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
		successfullySavedFiles[index] = true
	}

	render.JSON(w, r, successfullySavedFiles)
}
