// Package app ties together application resources and handlers.
package wado

import (
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
