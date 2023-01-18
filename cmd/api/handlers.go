package main

import (
	"errors"
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
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movies)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// Read the JSON payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJson(w, r, requestPayload)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	// Validate the user against DB
	user, err := app.Db.GetUserByEmail(requestPayload.Email)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, errors.New("invalid credentials"))
		return
	}

	// Check password
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		fmt.Println(err)
		app.errorJson(w, errors.New("invalid credentials"))
		return
	}

	// Create JwtUser
	u := jwtUser {
		Id: user.Id,
		FirstName: user.FirstName,
		LastName: user.LastName,
	}

	// Generate tokens
	tokens, err := app.Auth.GenerateTokenPair(&u)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	// log.Println(tokens.Token)
	refreshCookie := app.Auth.getRefreshCookie(tokens.RefreshToken)

	http.SetCookie(w, refreshCookie)

	app.writeJson(w, http.StatusAccepted, tokens)
}