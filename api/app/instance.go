package app

import (
	"context"
	"dicom-store-api/database"
	"dicom-store-api/models"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"github.com/suyashkumar/dicom/pkg/tag"
	"net/http"
)

type InstanceResource struct {
	DB            *pg.DB
	InstanceStore InstanceStore
}

func NewInstanceResource(db *pg.DB, instanceStore InstanceStore) *InstanceResource {
	return &InstanceResource{
		DB:            db,
		InstanceStore: instanceStore,
	}
}

func (rs *InstanceResource) ctx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		instanceUID := chi.URLParam(r, "instanceUID")
		if instanceUID != "" {
			instanceUIDTagInfo, _ := tag.Find((&models.Instance{}).GetObjectIdFieldTag())
			fields := map[string]any{instanceUIDTagInfo.Name: instanceUID}

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

type ToolsResponse []interface{}

type UpdateToolsRequest struct {
	FreehandRoiTool []interface{}
}

func (rs *InstanceResource) loadToolsData(w http.ResponseWriter, r *http.Request) {
	instance, ok := r.Context().Value(ctxInstance).(*models.Instance)
	if !ok {
		render.Render(w, r, ErrNotFound)
		return
	}

	var toolsData interface{}
	err := json.Unmarshal([]byte(instance.ToolsData), &toolsData)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}
	render.JSON(w, r, toolsData)
}

func (rs *InstanceResource) updateToolsData(w http.ResponseWriter, r *http.Request) {
	instance, ok := r.Context().Value(ctxInstance).(*models.Instance)
	if !ok {
		render.Render(w, r, ErrNotFound)
		return
	}

	var req UpdateToolsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	var toolsData = map[string]interface{}{
		"FreehandRoiTool": req.FreehandRoiTool,
	}

	toolsDataJson, err := json.Marshal(toolsData)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	instance.ToolsData = string(toolsDataJson)
	err = rs.InstanceStore.Update(instance, nil)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}

	render.JSON(w, r, toolsData)
}
