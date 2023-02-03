package dicomweb

import (
	"context"
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

type EmptyParentEntitiesListError struct{}

func (e EmptyParentEntitiesListError) Error() string {
	return "empty parent entities list"
}

var ErrEmptyParentEntitiesList = &EmptyParentEntitiesListError{}

func (rs *QIDOResource) ctx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		studyUID := chi.URLParam(r, "studyUID")
		if studyUID != "" {
			studyUIDTagInfo, _ := tag.Find((&models.Study{}).GetObjectIdFieldTag())
			fields := map[string]any{studyUIDTagInfo.Name: studyUID}
			studyList, err := rs.StudyStore.FindBy(fields, &database.SelectQueryOptions{Limit: 1}, nil)
			if err != nil || len(studyList) != 1 {
				render.Render(w, r, ErrNotFound)
				return
			}
			ctx = context.WithValue(ctx, ctxStudy, studyList[0])
		}

		seriesUID := chi.URLParam(r, "seriesUID")
		if seriesUID != "" {
			seriesUIDTagInfo, _ := tag.Find((&models.Series{}).GetObjectIdFieldTag())
			fields := map[string]any{seriesUIDTagInfo.Name: seriesUID}

			study := ctx.Value(ctxStudy).(*models.Study)
			if study == nil {
				render.Render(w, r, ErrInternalServerError)
				return
			}
			fields["StudyId"] = study.ID

			seriesList, err := rs.SeriesStore.FindBy(fields, &database.SelectQueryOptions{Limit: 1}, nil)
			if err != nil || len(seriesList) != 1 {
				render.Render(w, r, ErrNotFound)
				return
			}
			ctx = context.WithValue(ctx, ctxSeries, seriesList[0])
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// QIDOResource implements management handler.
type QIDOResource struct {
	DB            *pg.DB
	StudyStore    StudyStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

// NewQIDOResource creates and returns a QIDOResource.
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
	for objectIndex, object := range objects {
		var formattedDataMaps []map[string]any

		formattedDataMaps = append(formattedDataMaps, formatDicomObject(object, rd.IncludedFields, rd.IncludeAllFields))
		if _, ok := object.(*models.Series); ok {
			formattedDataMaps = append(formattedDataMaps, formatDicomObject(object.(*models.Series).Study, rd.IncludedFields, rd.IncludeAllFields))
		}
		if _, ok := object.(*models.Instance); ok {
			formattedDataMaps = append(formattedDataMaps, formatDicomObject(object.(*models.Instance).Series, rd.IncludedFields, rd.IncludeAllFields))
			formattedDataMaps = append(formattedDataMaps, formatDicomObject(object.(*models.Instance).Series.Study, rd.IncludedFields, rd.IncludeAllFields))
		}

		mergedMaps := map[string]any{}
		for _, dataMap := range formattedDataMaps {
			for k, v := range dataMap {
				mergedMaps[k] = v
			}
		}
		s[objectIndex] = mergedMaps
	}

	response := QIDOResponse(s)
	return &response
}

func formatDicomObject(object models.DicomObject, includedFields map[tag.Tag]bool, includeAllFields bool) map[string]any {
	var formatted = map[string]any{}

	reflection := reflect.TypeOf(object).Elem()
	for fieldIndex := 0; fieldIndex < reflection.NumField(); fieldIndex++ {
		field := reflection.Field(fieldIndex)
		tagInfo, err := tag.FindByName(field.Tag.Get("dicom"))
		if err != nil {
			continue
		}

		_, isInIncludedFieldsMap := includedFields[tagInfo.Tag]
		if includeAllFields == false && !isInIncludedFieldsMap {
			continue
		}

		fieldKey := fmt.Sprintf("%04X%04X", tagInfo.Tag.Group, tagInfo.Tag.Element)

		value, err := utils.FormatStringValueForResponse(tagInfo, reflect.ValueOf(object).Elem().Field(fieldIndex).String())
		if err != nil {
			panic(err)
		}
		formatted[fieldKey] = map[string]interface{}{
			"vr":    tagInfo.VR,
			"Value": value,
		}
	}
	return formatted
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
				continue
			}
			data.Limit = limit
			break
		case "offset":
			offset, err := strconv.Atoi(value[0])
			if err != nil {
				continue
			}
			data.Offset = offset
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

	fields := requestData.getFieldsForStoreRequest()

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

	options := &database.SelectQueryOptions{
		Limit:  requestData.Limit,
		Offset: requestData.Offset,
	}

	fields := requestData.getFieldsForStoreRequest()
	fields, err := transformFieldsForObject(rs, fields, &models.Series{})
	if err != nil {
		if err == ErrEmptyParentEntitiesList {
			render.Respond(w, r, newQIDOResponse([]models.DicomObject{}, requestData))
			return
		}
		render.Render(w, r, ErrInternalServerError)
		return
	}

	study, ok := r.Context().Value(ctxStudy).(*models.Study)
	if ok {
		fields["StudyId"] = study.ID
	}

	seriesList, err := rs.SeriesStore.FindBy(fields, options, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	dicomObjectsList := make([]models.DicomObject, len(seriesList))
	for i, series := range seriesList {
		dicomObjectsList[i] = series
	}
	render.Respond(w, r, newQIDOResponse(dicomObjectsList, requestData))
}

func (rs *QIDOResource) instances(w http.ResponseWriter, r *http.Request) {
	requestData := getQIDORequest(r)

	options := &database.SelectQueryOptions{
		Limit:  requestData.Limit,
		Offset: requestData.Offset,
	}

	fields := requestData.getFieldsForStoreRequest()
	fields, err := transformFieldsForObject(rs, fields, &models.Instance{})
	if err != nil {
		if err == ErrEmptyParentEntitiesList {
			render.Respond(w, r, newQIDOResponse([]models.DicomObject{}, requestData))
			return
		}
		render.Render(w, r, ErrInternalServerError)
		return
	}

	series, ok := r.Context().Value(ctxSeries).(*models.Series)
	if ok {
		fields["SeriesId"] = series.ID
	}

	instanceList, err := rs.InstanceStore.FindBy(fields, options, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	dicomObjectsList := make([]models.DicomObject, len(instanceList))
	for i, study := range instanceList {
		dicomObjectsList[i] = study
	}
	render.Respond(w, r, newQIDOResponse(dicomObjectsList, requestData))
}

func transformFieldsForObject(rs *QIDOResource, fields map[string]any, dicomObject models.DicomObject) (map[string]any, error) {
	_, isStudy := dicomObject.(*models.Study)
	_, isSeries := dicomObject.(*models.Series)
	_, isInstance := dicomObject.(*models.Instance)

	if isStudy {
		return fields, nil
	}

	if isSeries || isInstance {
		studyFields := make(map[string]any)
		for fieldName, fieldValue := range fields {
			studyField := reflect.ValueOf(&models.Study{}).Elem().FieldByName(fieldName)
			if studyField.IsValid() {
				studyFields[fieldName] = fieldValue
				delete(fields, fieldName)
			}
		}
		if len(studyFields) > 0 {
			studyList, err := rs.StudyStore.FindBy(studyFields, nil, nil)
			if err != nil {
				return nil, err
			}
			if len(studyList) == 0 {
				return nil, ErrEmptyParentEntitiesList
			}
			studyIds := make([]int, len(studyList))
			for i, study := range studyList {
				studyIds[i] = study.ID
			}
			fields["StudyId"] = studyIds
		}
	}

	if isInstance {
		seriesFields := make(map[string]any)
		for fieldName, fieldValue := range fields {
			seriesField := reflect.ValueOf(&models.Series{}).Elem().FieldByName(fieldName)
			if seriesField.IsValid() {
				seriesFields[fieldName] = fieldValue
				delete(fields, fieldName)
			}
		}
		if len(seriesFields) > 0 {
			seriesList, err := rs.SeriesStore.FindBy(seriesFields, nil, nil)
			if err != nil {
				return nil, err
			}
			if len(seriesList) == 0 {
				return nil, ErrEmptyParentEntitiesList
			}
			seriesIds := make([]int, len(seriesList))
			for i, series := range seriesList {
				seriesIds[i] = series.ID
			}
			fields["SeriesId"] = seriesIds
		}
	}

	return fields, nil
}

func (requestData *QIDORequest) getFieldsForStoreRequest() map[string]any {
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
	return fields
}
