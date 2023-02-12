package app

import (
	"dicom-store-api/models"
	"github.com/go-chi/render"
	"github.com/go-pg/pg"
	"net/http"
)

type SummaryResource struct {
	DB            *pg.DB
	StudyStore    StudyStore
	SeriesStore   SeriesStore
	InstanceStore InstanceStore
}

func NewSummaryResource(db *pg.DB, studyStore StudyStore, seriesStore SeriesStore, instanceStore InstanceStore) *SummaryResource {
	return &SummaryResource{
		DB:            db,
		StudyStore:    studyStore,
		SeriesStore:   seriesStore,
		InstanceStore: instanceStore,
	}
}

type ModalitiesCount struct {
	Modality string `json:"modality"`
	Count    int    `json:"count"`
}

type SummaryResponse struct {
	StudyCount       int               `json:"studyCount"`
	SeriesCount      int               `json:"seriesCount"`
	InstanceCount    int               `json:"instanceCount"`
	PatientsCount    int               `json:"patientsCount"`
	ModalitiesCounts []ModalitiesCount `json:"modalitiesCounts"`
}

func (rs *SummaryResource) getSummary(w http.ResponseWriter, r *http.Request) {
	studiesCount, err := rs.StudyStore.CountBy(nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	seriesCount, err := rs.SeriesStore.CountBy(nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	instancesCount, err := rs.InstanceStore.CountBy(nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var uniquePatientsStudiesIds []*models.Study
	query := rs.DB.Model(&uniquePatientsStudiesIds)
	query = query.ColumnExpr("DISTINCT ON (patient_id) id")
	err = query.Select()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	patientsCount := len(uniquePatientsStudiesIds)

	var modalitiesCounts []ModalitiesCount
	stringQuery := "SELECT modality, COUNT(*) FROM " + (&models.Series{}).GetTableName() + " GROUP BY modality"
	_, err = rs.DB.Query(&modalitiesCounts, stringQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summary := SummaryResponse{
		studiesCount,
		seriesCount,
		instancesCount,
		patientsCount,
		modalitiesCounts,
	}

	render.JSON(w, r, summary)
}
