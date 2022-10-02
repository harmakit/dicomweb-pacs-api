package dicomweb

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
	ctxSeries
	ctxInstance
)

// API provides application resources and handlers.
type API struct {
	QIDO *QIDOResource
	STOW *STOWResource
	WADO *WADOResource
}

type StudyStore interface {
	FindBy(fields map[string]any, options *database.SelectQueryOptions, tx *pg.Tx) ([]*models.Study, error)
	Create(s *models.Study, tx *pg.Tx) error
	Update(s *models.Study, tx *pg.Tx) error
}
type SeriesStore interface {
	FindBy(fields map[string]any, options *database.SelectQueryOptions, tx *pg.Tx) ([]*models.Series, error)
	Create(s *models.Series, tx *pg.Tx) error
	Update(s *models.Series, tx *pg.Tx) error
}
type InstanceStore interface {
	FindBy(fields map[string]any, options *database.SelectQueryOptions, tx *pg.Tx) ([]*models.Instance, error)
	Create(s *models.Instance, tx *pg.Tx) error
	Update(s *models.Instance, tx *pg.Tx) error
}

// NewAPI configures and returns application API.
func NewAPI(db *pg.DB) (*API, error) {
	studyStore := database.NewStudyStore(db)
	seriesStore := database.NewSeriesStore(db)
	instanceStore := database.NewInstanceStore(db)

	QIDO := NewQIDOResource(db, studyStore, seriesStore, instanceStore)
	STOW := NewSTOWResource(db, studyStore, seriesStore, instanceStore)
	WADO := NewWADOResource(db, studyStore, seriesStore, instanceStore)

	api := &API{
		QIDO,
		STOW,
		WADO,
	}
	return api, nil
}

// Router provides application routes.
func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()

	// QIDO group
	r.Group(func(r chi.Router) {
		r.Use(a.QIDO.ctx)
		r.Get("/studies", a.QIDO.studies)
		r.Get("/studies/{studyUID}/series", a.QIDO.series)
		r.Get("/studies/{studyUID}/series/{seriesUID}/instances", a.QIDO.instances)
	})

	// WADO group
	r.Group(func(r chi.Router) {
		r.Use(a.WADO.ctx)
		r.Get("/studies/{studyUID}", a.WADO.study)
		r.Get("/studies/{studyUID}/series/{seriesUID}", a.WADO.series)
		r.Get("/studies/{studyUID}/series/{seriesUID}/instances/{instanceUID}", a.WADO.instance)

		// todo metadata, generated, wado-uri
	})

	// STOW group
	r.Group(func(r chi.Router) {
		r.Post("/studies", a.STOW.save)
	})

	return r
}

func log(r *http.Request) logrus.FieldLogger {
	return logging.GetLogEntry(r)
}
