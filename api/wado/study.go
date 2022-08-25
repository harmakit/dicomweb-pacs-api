package wado

import (
	"context"
	"errors"
	"net/http"

	"dicom-store-api/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	validation "github.com/go-ozzo/ozzo-validation"
)

// ErrStudyValidation defines the list of error types returned from study resource.
var (
	ErrStudyValidation = errors.New("study validation error")
)

// StudyStore defines database operations for a study.
type StudyStore interface {
	Get(accountID int) (*models.Study, error)
	//FindByPatient(patient string) ([]*models.Study, error)
	Update(s *models.Study) error
}

// StudyResource implements study management handler.
type StudyResource struct {
	Store StudyStore
}

// NewStudyResource creates and returns a study resource.
func NewStudyResource(store StudyStore) *StudyResource {
	return &StudyResource{
		Store: store,
	}
}

func (rs *StudyResource) router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(rs.studyCtx)
	r.Get("/", rs.get)
	r.Put("/", rs.update)
	return r
}

func (rs *StudyResource) studyCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//claims := jwt.ClaimsFromCtx(r.Context())
		//s, err := rs.Store.Get(claims.ID)
		//if err != nil {
		//	log(r).WithField("studyCtx", claims.Sub).Error(err)
		//	render.Render(w, r, ErrInternalServerError)
		//	return
		//} todo: would be helpful for series and instances
		// todo: maybe i should add patient to context
		ctx := context.WithValue(r.Context(), ctxStudy, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type studyRequest struct {
	*models.Study
	ProtectedID int `json:"id"`
}

func (d *studyRequest) Bind(r *http.Request) error {
	return nil
}

type studyResponse struct {
	*models.Study
}

func newStudyResponse(s *models.Study) *studyResponse {
	return &studyResponse{
		Study: s,
	}
}

type studiesResponse struct {
	Studies []*models.Study
}

func newStudiesResponse(s []*models.Study) *studiesResponse {
	return &studiesResponse{
		Studies: s,
	}
}

func (rs *StudyResource) get(w http.ResponseWriter, r *http.Request) {
	//dataset, _ := dicom.ParseFile("testdata/1.dcm", nil)
	//
	//// Dataset will nicely print the DICOM dataset data out of the box.
	//fmt.Println(dataset)
	//
	//// Dataset is also JSON serializable out of the box.
	//j, _ := json.Marshal(dataset)
	//fmt.Println(j)

	studies := []*models.Study{}
	//studies, err := rs.Store.FindByPatient("patient1")
	//if err != nil {
	//	render.Render(w, r, ErrInternalServerError)
	//	return
	//}
	render.Respond(w, r, newStudiesResponse(studies))
}

func (rs *StudyResource) update(w http.ResponseWriter, r *http.Request) {
	s := r.Context().Value(ctxStudy).(*models.Study)
	data := &studyRequest{Study: s}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
	}

	if err := rs.Store.Update(s); err != nil {
		switch err.(type) {
		case validation.Errors:
			render.Render(w, r, ErrValidation(ErrStudyValidation, err.(validation.Errors)))
			return
		}
		render.Render(w, r, ErrRender(err))
		return
	}
	render.Respond(w, r, newStudyResponse(s))
}
