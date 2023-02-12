package app

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
	ctxInstance ctxKey = iota
)

type API struct {
	instanceResource *InstanceResource
	summaryResource  *SummaryResource
}

type StudyStore interface {
	FindBy(fields map[string]any, options *database.SelectQueryOptions, tx *pg.Tx) ([]*models.Study, error)
	Create(s *models.Study, tx *pg.Tx) error
	Update(s *models.Study, tx *pg.Tx) error
	CountBy(fields map[string]any, tx *pg.Tx) (int, error)
}
type SeriesStore interface {
	FindBy(fields map[string]any, options *database.SelectQueryOptions, tx *pg.Tx) ([]*models.Series, error)
	Create(s *models.Series, tx *pg.Tx) error
	Update(s *models.Series, tx *pg.Tx) error
	CountBy(fields map[string]any, tx *pg.Tx) (int, error)
}
type InstanceStore interface {
	FindBy(fields map[string]any, options *database.SelectQueryOptions, tx *pg.Tx) ([]*models.Instance, error)
	Create(s *models.Instance, tx *pg.Tx) error
	Update(s *models.Instance, tx *pg.Tx) error
	CountBy(fields map[string]any, tx *pg.Tx) (int, error)
}

func NewAPI(db *pg.DB) (*API, error) {
	studyStore := database.NewStudyStore(db)
	seriesStore := database.NewSeriesStore(db)
	instanceStore := database.NewInstanceStore(db)

	instanceResource := NewInstanceResource(db, instanceStore)
	summaryResource := NewSummaryResource(db, studyStore, seriesStore, instanceStore)

	api := &API{
		instanceResource,
		summaryResource,
	}
	return api, nil
}

func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/summary", func(r chi.Router) {
		r.Get("/", a.summaryResource.getSummary)
	})

	r.Route("/instance/{instanceUID}", func(r chi.Router) {
		r.Use(a.instanceResource.ctx)
		r.Get("/tools", a.instanceResource.loadToolsData)
		r.Put("/tools", a.instanceResource.updateToolsData)
	})

	return r
}

func log(r *http.Request) logrus.FieldLogger {
	return logging.GetLogEntry(r)
}
