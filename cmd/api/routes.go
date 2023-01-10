package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// create router mux
	mux := chi.NewRouter()

	// configure the middleware
	//   handles 500 errors
	mux.Use(middleware.Recoverer)

	// add routes
	mux.Get("/", app.Home)

	return mux
}