package wado

import (
	"dicom-store-api/database"
	"dicom-store-api/models"
	"dicom-store-api/utils"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/suyashkumar/dicom/pkg/tag"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// QIDOResource implements management handler.
type QIDOResource struct {
	DB            *pg.DB
	StudyStore    StudyStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

// NewQIDOResource creates and returns a study resource.
func NewQIDOResource(db *pg.DB, studyStore StudyStore, seriesStore SeriesStore, instanceStore InstanceStore) *QIDOResource {
	return &QIDOResource{
		DB:            db,
		StudyStore:    studyStore,
		SeriesStore:   seriesStore,
		InstanceStore: instanceStore,
	}
}

type QIDOResponse []interface{}

func newQIDOResponse(objects []models.DicomObject, rd *QIDORequest) *QIDOResponse {
	var s = make([]interface{}, len(objects))
	for _, object := range objects {
		var formatted = map[string]any{}

		reflection := reflect.TypeOf(object).Elem()

		for i := 0; i < reflection.NumField(); i++ {
			field := reflection.Field(i)
			tagInfo, err := tag.FindByName(field.Tag.Get("dicom"))
			if err != nil {
				continue
			}

			_, isInIncludedFieldsMap := rd.IncludedFields[tagInfo.Tag]
			if rd.IncludeAllFields == false && !isInIncludedFieldsMap {
				continue
			}

			fieldKey := fmt.Sprintf("%04x%04x", tagInfo.Tag.Group, tagInfo.Tag.Element)
			formatted[fieldKey] = reflection.Field(i)
		}
		s = append(s, formatted)
	}

	response := QIDOResponse(s)
	return &response
}

type QIDORequest struct {
	Limit            int
	Offset           int
	IncludedFields   map[tag.Tag]bool
	IncludeAllFields bool
	Filters          map[tag.Tag][]string
}

func getQIDORequest(r *http.Request) *QIDORequest {
	data := &QIDORequest{
		Limit:            10,
		Offset:           0,
		IncludedFields:   map[tag.Tag]bool{},
		IncludeAllFields: false,
		Filters:          map[tag.Tag][]string{},
	}

	values := r.URL.Query()
	for key, value := range values {
		switch key {
		case "limit":
			limit, err := strconv.Atoi(value[0])
			if err != nil {
				data.Limit = limit
			}
			break
		case "offset":
			offset, err := strconv.Atoi(value[0])
			if err != nil {
				data.Offset = offset
			}
			break
		case "includefield":
			for _, field := range value {
				if field == "all" {
					data.IncludeAllFields = true
					data.IncludedFields = map[tag.Tag]bool{}
				}
				if data.IncludeAllFields == true {
					continue
				}
				fieldTag, err := utils.GetTagByNameOrCode(field)
				if err != nil {
					continue
				}
				data.IncludedFields[fieldTag] = true
			}
			break
		default:
			fieldTag, err := utils.GetTagByNameOrCode(key)
			if err != nil {
				continue
			}
			values := strings.Split(value[0], ",")
			data.Filters[fieldTag] = append(data.Filters[fieldTag], values...)
		}
	}

	if len(data.IncludedFields) == 0 {
		data.IncludeAllFields = true
	}

	return data
}

func (rs *QIDOResource) studies(w http.ResponseWriter, r *http.Request) {
	requestData := getQIDORequest(r)

	options := &database.SelectQueryOptions{
		Limit:  requestData.Limit,
		Offset: requestData.Offset,
	}

	fields := map[string]any{}
	if requestData.Filters != nil {
		for key, value := range requestData.Filters {
			if len(value) == 0 {
				continue
			}
			tagInfo, _ := tag.Find(key)
			if len(value) == 1 {
				fields[tagInfo.Name] = value[0]
			} else {
				fields[tagInfo.Name] = value
			}
		}
	}

	studyList, err := rs.StudyStore.FindBy(fields, options, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	dicomObjectsList := make([]models.DicomObject, len(studyList))
	for i, study := range studyList {
		dicomObjectsList[i] = study
	}
	render.Respond(w, r, newQIDOResponse(dicomObjectsList, requestData))
}

func (rs *QIDOResource) series(w http.ResponseWriter, r *http.Request) {
	requestData := getQIDORequest(r)
	studyUIDTag := (&models.Study{}).GetObjectIdFieldTag()
	requestData.Filters[studyUIDTag] = append(requestData.Filters[studyUIDTag], chi.URLParam(r, "studyUID"))
	_ = requestData
}

func (rs *QIDOResource) instances(w http.ResponseWriter, r *http.Request) {
	requestData := getQIDORequest(r)
	studyUIDTag := (&models.Study{}).GetObjectIdFieldTag()
	requestData.Filters[studyUIDTag] = append(requestData.Filters[studyUIDTag], chi.URLParam(r, "studyUID"))
	seriesUIDTag := (&models.Series{}).GetObjectIdFieldTag()
	requestData.Filters[seriesUIDTag] = append(requestData.Filters[seriesUIDTag], chi.URLParam(r, "seriesUID"))
}

//
//type QIDOSaveResponse struct {
//	Study *models.Study
//}
//
//func newQIDOSaveResponse(s *models.Study) *QIDOSaveResponse {
//	return &QIDOSaveResponse{
//		Study: s,
//	}
//}
//
//func (rs *QIDOResource) save(w http.ResponseWriter, r *http.Request) {
//	const MaxUploadSize = 10 << 20 // 10MB
//	if r.ContentLength > MaxUploadSize {
//		http.Error(w, "The uploaded image is too big. Please use an image less than 10MB in size", http.StatusBadRequest)
//		return
//	}
//	bodyReader := http.MaxBytesReader(w, r.Body, MaxUploadSize)
//
//	defer bodyReader.Close()
//
//	body, err := ioutil.ReadAll(bodyReader)
//	if err != nil || len(body) == 0 {
//		http.Error(w, "Wrong request body", http.StatusBadRequest)
//		return
//	}
//
//	dataset, _ := dicom.Parse(bytes.NewReader(body), MaxUploadSize, nil)
//
//	study := &models.Study{}
//	utils.ExtractDicomObjectFromDataset(dataset, study)
//
//	series := &models.Series{Study: study}
//	utils.ExtractDicomObjectFromDataset(dataset, series)
//
//	instance := &models.Instance{Series: series}
//	utils.ExtractDicomObjectFromDataset(dataset, instance)
//
//	tx, err := rs.DB.Begin()
//
//	studyList, err := rs.StudyStore.FindBy(map[string]any{
//		"StudyInstanceUID": study.StudyInstanceUID,
//	}, nil)
//	if err != nil {
//		render.Render(w, r, ErrInternalServerError)
//		return
//	}
//
//	if len(studyList) == 1 {
//		study = studyList[0]
//		if err = rs.StudyStore.Update(study, tx); err != nil {
//			tx.Rollback()
//			render.Render(w, r, ErrInternalServerError)
//			return
//		}
//	} else {
//		if err = rs.StudyStore.Create(study, tx); err != nil {
//			tx.Rollback()
//			render.Render(w, r, ErrInternalServerError)
//			return
//		}
//	}
//
//	seriesList, err := rs.SeriesStore.FindBy(map[string]any{
//		"SeriesInstanceUID": series.SeriesInstanceUID,
//	}, nil)
//	if err != nil {
//		render.Render(w, r, ErrInternalServerError)
//		return
//	}
//
//	if len(seriesList) == 1 {
//		series = seriesList[0]
//		if err = rs.SeriesStore.Update(series, tx); err != nil {
//			tx.Rollback()
//			render.Render(w, r, ErrInternalServerError)
//			return
//		}
//	} else {
//		series.StudyId = study.ID
//		series.Study = study
//		if err = rs.SeriesStore.Create(series, tx); err != nil {
//			tx.Rollback()
//			render.Render(w, r, ErrInternalServerError)
//			return
//		}
//	}
//
//	instanceList, err := rs.InstanceStore.FindBy(map[string]any{
//		"SOPInstanceUID": instance.SOPInstanceUID,
//	}, nil)
//	if err != nil {
//		render.Render(w, r, ErrInternalServerError)
//		return
//	}
//
//	if len(instanceList) == 1 {
//		instance = instanceList[0]
//		if err = rs.InstanceStore.Update(instance, tx); err != nil {
//			tx.Rollback()
//			render.Render(w, r, ErrInternalServerError)
//			return
//		}
//	} else {
//		instance.SeriesId = series.ID
//		instance.Series = series
//		if err = rs.InstanceStore.Create(instance, tx); err != nil {
//			tx.Rollback()
//			render.Render(w, r, ErrInternalServerError)
//			return
//		}
//	}
//
//	path := fs.GetDicomPath(study, series, instance)
//	if err = fs.Save(path, body); err != nil {
//		tx.Rollback()
//		render.Render(w, r, ErrInternalServerError)
//		return
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		tx.Rollback()
//		render.Render(w, r, ErrInternalServerError)
//		return
//	}
//
//	render.Respond(w, r, newQIDOSaveResponse(study))
//}
