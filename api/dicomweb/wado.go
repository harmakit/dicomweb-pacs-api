package dicomweb

import (
	"bufio"
	"context"
	"dicom-store-api/database"
	"dicom-store-api/fs"
	"dicom-store-api/models"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/suyashkumar/dicom/pkg/tag"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
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
func writeWADORSResponse(w http.ResponseWriter, paths []string) error {
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

	err = writeWADORSResponse(w, paths)
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

	err = writeWADORSResponse(w, paths)
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

	err := writeWADORSResponse(w, paths)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	return
}
