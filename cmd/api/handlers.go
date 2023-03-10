package main

import (
	"backend/internal/graph"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
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

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.Auth.getExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

func (app *application) movieCatalog(w http.ResponseWriter, r *http.Request) {
	movies, err := app.Db.AllMovies()
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movies)
}

func (app *application) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieId, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	movie, err := app.Db.OneMovie(movieId)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movie)
}

func (app *application) GetMovieForEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieId, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	movie, allGenres, err := app.Db.OneMovieForEdit(movieId)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	var payload = struct {
		Movie *models.Movie `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		movie,
		allGenres,
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app *application) AllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.Db.AllGenres()
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, genres)
}

func (app *application) InsertMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie

	err := app.readJson(w, r, &movie)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	// try to get image
	movie = app.getPoster(movie)

	// Insert movie
	newId, err := app.Db.InsertMovie(movie)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	// handle genres
	err = app.Db.UpdateMovieGenres(newId, movie.GenresArray)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	resp := JsonResponse{
		Error: false,
		Message: "movie updated",
	}

	_ = app.writeJson(w, http.StatusOK, resp)
}

func (app *application) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	var payload models.Movie

	err := app.readJson(w, r, &payload)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	movie, err := app.Db.OneMovie(payload.Id)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	movie.Title = payload.Title
	movie.ReleaseDate = payload.ReleaseDate
	movie.Description = payload.Description
	movie.MpaaRating = payload.MpaaRating
	movie.RunTime = payload.RunTime
	movie.UpdatedAt = time.Now()

	err = app.Db.UpdateMovie(*movie)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	err = app.Db.UpdateMovieGenres(movie.Id, payload.GenresArray)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	resp := JsonResponse{
		Error: false,
		Message: "movie updated",
	}
	app.writeJson(w, http.StatusAccepted, resp)
}

func (app *application) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieId, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	err = app.Db.DeleteMovie(movieId)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	resp := JsonResponse{
		Error: false,
		Message: "movie deleted",
	}
	app.writeJson(w, http.StatusAccepted, resp)
}

func (app *application) getPoster(movie models.Movie) models.Movie {
	type theMovieDb struct {
		Page int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
	}

	client := &http.Client{}
	endpointUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.TmdbApiKey)

	req, err := http.NewRequest(
		"GET", 
		endpointUrl + "&query=" + url.QueryEscape(movie.Title),
		nil,
 	)
	if err != nil {
		fmt.Println(err)
		return movie
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return movie
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return movie
	}

	var responseObject theMovieDb
	err = json.Unmarshal(bodyBytes, &responseObject)
	if err != nil {
		fmt.Println(err)
		return movie
	}

	if len(responseObject.Results) > 0 {
		movie.Image = responseObject.Results[0].PosterPath
	}

	return movie
}

func (app *application) AllMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	genreId, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}
	
	movies, err := app.Db.AllMovies(genreId)
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	_ = app.writeJson(w, http.StatusOK, movies)
}

func (app *application) MoviesGraphQl(w http.ResponseWriter, r *http.Request) {
	// Populate the graph type with the movies
	movies, err := app.Db.AllMovies()
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	// Get query from request
	q, _ := io.ReadAll(r.Body)
	query := string(q)
	
	// Create new var of graph.Graph
	g := graph.New(movies)

	// Set query string on var
	g.QueryString = query

	// Perform the query
	response, err := g.Query()
	if err != nil {
		fmt.Println(err)
		app.errorJson(w, err)
		return
	}

	// Send the response
	j, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)

}