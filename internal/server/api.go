package server

import (
	"fmt"
	"net/http"
	"time"

	"city-tags-api/internal/database"
)

type Api struct {
	db database.Service
}

func NewServer(port int) *http.Server {
	api := &Api{
		db: database.New(),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      api.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
