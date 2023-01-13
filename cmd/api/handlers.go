package main

import (
	"fmt"
	"net/http"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "running",
		Version: "1.0.0",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.Db.AllMovies()
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movies)
}
