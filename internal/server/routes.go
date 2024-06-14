package server

import (
	"net/http"

	_ "city-tags-api/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (s *Api) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Logger)

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/test", testHandler)

	return r
}

type ErrorResp struct {
	Message string `json:"message" example:"internal server error"`
}

// @Summary		Test handler
// @Description	Test handler
// @Accept			json
// @Produce		text/plain
// @Success		200
// @Failure      500  {object} ErrorResp
// @Router			/test [get]
func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
