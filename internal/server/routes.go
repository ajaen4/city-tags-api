package server

import (
	_ "city-tags-api/docs"
	"log"
	"net/http"

	"city-tags-api/internal/api_errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (api *Api) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Logger)
	r.Use(jwtauth.Verifier(api.tokenAuth))
	r.Use(Authenticator(api.tokenAuth))

	r.Route("/v0", func(r chi.Router) {
		r.Get("/swagger/*", httpSwagger.WrapHandler)

		r.Get("/cities/{cityId}", NewHandler(api.getCity))
		r.Get("/cities", NewHandler(api.getCities))
	})

	return r
}

type CustomHandler func(w http.ResponseWriter, request *http.Request) error

func NewHandler(customHandler CustomHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := customHandler(w, r)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			if clientErr, ok := err.(*api_errors.ClientErr); ok {
				respondWithJSON(w, clientErr.HttpCode, clientErr)
			} else {
				respondWithJSON(w, http.StatusInternalServerError,
					api_errors.InternalErr{
						HttpCode: http.StatusInternalServerError,
						Message:  "internal server error",
					},
				)
			}
		}
	}
}
