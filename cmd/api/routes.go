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
	mux.Get("/movies/{id}", app.GetMovie)

	mux.Get("/genres", app.AllGenres)

	mux.Post("/authenticate", app.authenticate)
	mux.Get("/refresh", app.refreshToken)
	mux.Get("/logout", app.logout)

	// Route group. All routes in here will have authRequired active
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.authRequired)

		mux.Get("/movies", app.movieCatalog)
		mux.Get("/movies/{id}", app.GetMovieForEdit)
		mux.Put("/movies/0", app.InsertMovie)
		mux.Patch("/movies/{id}", app.UpdateMovie)
	})

	return mux
}