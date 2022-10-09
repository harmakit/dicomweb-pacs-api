package dicomweb

import (
	"bufio"
	"bytes"
	"context"
	"dicom-store-api/database"
	"dicom-store-api/fs"
	"dicom-store-api/models"
	"dicom-store-api/utils"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
)

type RequestType int

const (
	requestTypeDefault RequestType = iota
	requestTypeMetadata
	requestTypeRendered
)

const (
	ctxRequestType = iota
)

func (rs *WADOResource) ctx(next http.Handler) http.Handler {
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

		instanceUID := chi.URLParam(r, "instanceUID")
		if instanceUID != "" {
			instanceUIDTagInfo, _ := tag.Find((&models.Instance{}).GetObjectIdFieldTag())
			fields := map[string]any{instanceUIDTagInfo.Name: instanceUID}

			series := ctx.Value(ctxSeries).(*models.Series)
			if series == nil {
				render.Render(w, r, ErrInternalServerError)
				return
			}
			fields["SeriesId"] = series.ID

			instanceList, err := rs.InstanceStore.FindBy(fields, &database.SelectQueryOptions{Limit: 1}, nil)
			if err != nil || len(instanceList) != 1 {
				render.Render(w, r, ErrNotFound)
				return
			}
			ctx = context.WithValue(ctx, ctxInstance, instanceList[0])
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rs *WADOResource) ctxDefaultRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxRequestType, requestTypeDefault)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rs *WADOResource) ctxMetadataRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxRequestType, requestTypeMetadata)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rs *WADOResource) ctxRenderedRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxRequestType, requestTypeRendered)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WADOResource implements management handler.
type WADOResource struct {
	DB            *pg.DB
	StudyStore    StudyStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

// NewWADOResource creates and returns a WADOResource.
func NewWADOResource(db *pg.DB, studyStore StudyStore, seriesStore SeriesStore, instanceStore InstanceStore) *WADOResource {
	return &WADOResource{
		DB:            db,
		StudyStore:    studyStore,
		SeriesStore:   seriesStore,
		InstanceStore: instanceStore,
	}
}

// Writes a multipart response from a list of paths to dicom files
func writeWADORSResponse(w http.ResponseWriter, r *http.Request, paths []string) error {
	if len(paths) == 0 {
		render.Render(w, r, ErrNotFound)
		return nil
	}

	requestType := r.Context().Value(ctxRequestType).(RequestType)
	switch requestType {
	case requestTypeMetadata:
		var responseData []any
		for _, path := range paths {
			var formatted = map[string]any{}
			dataset, _ := dicom.ParseFile(path, nil)
			for _, element := range dataset.Elements {
				if element.ValueRepresentation == tag.VRPixelData {
					continue
				}
				tagInfo, err := tag.Find(element.Tag)
				if err != nil {
					continue
				}

				fieldKey := fmt.Sprintf("%04x%04x", tagInfo.Tag.Group, tagInfo.Tag.Element)

				formatted[fieldKey] = map[string]interface{}{
					"vr":    tagInfo.VR,
					"Value": element.Value.GetValue(),
				}

			}
			responseData = append(responseData, formatted)
		}
		render.Respond(w, r, responseData)
		return nil
	case requestTypeRendered:
		path := paths[0]
		dataset, _ := dicom.ParseFile(path, nil)
		viewportParameters := utils.RenderImageViewportParameters{}
		viewportParameters.ViewportHeight = 0
		viewportParameters.ViewportHeight = 0
		windowParameters := utils.RenderImageWindowParameters{}
		windowParameters.WindowWidth = 255
		windowParameters.WindowCenter = 127
		windowParameters.Function = utils.Sigmoid

		img, err := utils.RenderImage(dataset, viewportParameters, windowParameters)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, img, nil); err != nil {
			return err
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(buffer.Bytes())
	case requestTypeDefault:
		mw := multipart.NewWriter(w)
		w.Header().Set("Content-Type", fmt.Sprintf("multipart/related; type=\"application/dicom\"; boundary=%s", mw.Boundary()))

		partHeaders := textproto.MIMEHeader{}
		partHeaders.Set("Content-Type", "application/dicom")

		for _, path := range paths {
			partWriter, err := mw.CreatePart(partHeaders)
			if err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				if _, err := partWriter.Write(scanner.Bytes()); err != nil {
					return err
				}
			}

			if err := file.Close(); err != nil {
				return err
			}

			if err := scanner.Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rs *WADOResource) study(w http.ResponseWriter, r *http.Request) {
	var paths []string

	study := r.Context().Value(ctxStudy).(*models.Study)
	seriesList, err := rs.SeriesStore.FindBy(map[string]any{"StudyId": study.ID}, nil, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}
	for _, series := range seriesList {
		instanceList, err := rs.InstanceStore.FindBy(map[string]any{"SeriesId": series.ID}, nil, nil)
		if err != nil {
			render.Render(w, r, ErrInternalServerError)
			return
		}

		for _, instance := range instanceList {
			path := fs.GetDicomPath(study, series, instance)
			paths = append(paths, path)
		}
	}

	err = writeWADORSResponse(w, r, paths)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	return
}

func (rs *WADOResource) series(w http.ResponseWriter, r *http.Request) {
	var paths []string

	study := r.Context().Value(ctxStudy).(*models.Study)
	series := r.Context().Value(ctxSeries).(*models.Series)

	instanceList, err := rs.InstanceStore.FindBy(map[string]any{"SeriesId": series.ID}, nil, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	for _, instance := range instanceList {
		path := fs.GetDicomPath(study, series, instance)
		paths = append(paths, path)
	}

	err = writeWADORSResponse(w, r, paths)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	return
}

func (rs *WADOResource) instance(w http.ResponseWriter, r *http.Request) {
	var paths []string

	study := r.Context().Value(ctxStudy).(*models.Study)
	series := r.Context().Value(ctxSeries).(*models.Series)
	instance := r.Context().Value(ctxInstance).(*models.Instance)

	path := fs.GetDicomPath(study, series, instance)
	paths = append(paths, path)

	err := writeWADORSResponse(w, r, paths)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	return
}
