// Package app ties together application resources and handlers.
package wado

import (
	"dicom-store-api/models"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-pg/pg"
	"github.com/sirupsen/logrus"

	"dicom-store-api/database"
	"dicom-store-api/logging"
)

type ctxKey int

const (
	ctxStudy ctxKey = iota
)

// API provides application resources and handlers.
type API struct {
	Study   *StudyResource
	Studies *StudiesResource
}

type StudyStore interface {
	FindByFields(fields map[string]any, tx *pg.Tx) ([]*models.Study, error)
	Create(s *models.Study, tx *pg.Tx) error
	Update(s *models.Study, tx *pg.Tx) error
}
type SeriesStore interface {
	FindByFields(fields map[string]any, tx *pg.Tx) ([]*models.Series, error)
	Create(s *models.Series, tx *pg.Tx) error
	Update(s *models.Series, tx *pg.Tx) error
}
type InstanceStore interface {
	FindByFields(fields map[string]any, tx *pg.Tx) ([]*models.Instance, error)
	Create(s *models.Instance, tx *pg.Tx) error
	Update(s *models.Instance, tx *pg.Tx) error
}

// NewAPI configures and returns application API.
func NewAPI(db *pg.DB) (*API, error) {
	studyStore := database.NewStudyStore(db)
	seriesStore := database.NewSeriesStore(db)
	instanceStore := database.NewInstanceStore(db)
	study := NewStudyResource(studyStore)
	studies := NewStudiesResource(db, studyStore, seriesStore, instanceStore)

	api := &API{
		Study:   study,
		Studies: studies,
	}
	return api, nil
}

// Router provides application routes.
func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Mount("/study", a.Study.router())
	r.Mount("/studies", a.Studies.router())

	return r
}

func log(r *http.Request) logrus.FieldLogger {
	return logging.GetLogEntry(r)
}
