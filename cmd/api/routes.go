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
	mux.Use(app.enableCors)

	// add routes
	mux.Get("/", app.Home)
	mux.Get("/movies", app.AllMovies)

	mux.Post("/authenticate", app.authenticate)
	mux.Get("/refresh", app.refreshToken)

	return mux
}