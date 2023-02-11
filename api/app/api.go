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
}

type InstanceStore interface {
	FindBy(fields map[string]any, options *database.SelectQueryOptions, tx *pg.Tx) ([]*models.Instance, error)
	Update(s *models.Instance, tx *pg.Tx) error
}

func NewAPI(db *pg.DB) (*API, error) {
	instanceStore := database.NewInstanceStore(db)

	instanceResource := NewInstanceResource(db, instanceStore)

	api := &API{
		instanceResource,
	}
	return api, nil
}

func (a *API) Router() *chi.Mux {
	r := chi.NewRouter()

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
