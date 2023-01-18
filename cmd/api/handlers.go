package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
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
	err := app.readJson(w, r, &requestPayload)
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

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.Auth.CookieName {
			claims := &claims{}
			refreshToken := cookie.Value

			// parse the token to get the claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
				return []byte(app.JwtSecret), nil
			})
			if err != nil {
				app.errorJson(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get the user ID from the claim
			userId, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJson(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.Db.GetUserById(userId)
			if err != nil {
				app.errorJson(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				Id: user.Id,
				FirstName: user.FirstName,
				LastName: user.LastName,
			}

			tokenPairs, err := app.Auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJson(w, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, app.Auth.getRefreshCookie(tokenPairs.RefreshToken))

			app.writeJson(w, http.StatusOK, tokenPairs)
		}
	}
}