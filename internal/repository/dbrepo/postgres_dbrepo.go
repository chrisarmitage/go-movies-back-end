package dbrepo

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"time"
)

type PostgresDbRepo struct {
	Db *sql.DB
}

const dbTimeout = time.Second * 3

func (r *PostgresDbRepo) Connection() *sql.DB {
	return r.Db
}

func (r *PostgresDbRepo) AllMovies() ([]*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT
			id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		FROM 
			movies
		ORDER BY title
	`

	rows, err := r.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*models.Movie

	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(
			&movie.Id,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.RunTime,
			&movie.MpaaRating,
			&movie.Description,
			&movie.Image,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (r *PostgresDbRepo) OneMovie(id int) (*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT
			id, title, release_date, runtime, mpaa_rating, description, COALESCE(image, ''), created_at, updated_at
		FROM
			movies
		WHERE
			id = $1
	`
	var movie models.Movie

	row := r.Db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&movie.Id,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MpaaRating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// get genres
	query = `
		SELECT
			g.id, g.genre
		FROM
			movies_genres AS mg
		LEFT JOIN genres AS g on (mg.genre_id = g.id)
		WHERE
			mg.movie_id = $1
		ORDER BY g.genre
	`

	rows, err := r.Db.QueryContext(ctx, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.Id,
			&g.Genre,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	movie.Genres = genres

	return &movie, nil
}

func (r *PostgresDbRepo) OneMovieForEdit(id int) (*models.Movie, []*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT
			id, title, release_date, runtime, mpaa_rating, description, COALESCE(image, ''), created_at, updated_at
		FROM
			movies
		WHERE
			id = $1
	`
	var movie models.Movie

	row := r.Db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&movie.Id,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MpaaRating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)

	if err != nil {
		return nil, nil, err
	}

	// get selected genres
	query = `
		SELECT
			g.id, g.genre
		FROM
			movies_genres AS mg
		LEFT JOIN genres AS g on (mg.genre_id = g.id)
		WHERE
			mg.movie_id = $1
		ORDER BY g.genre
	`

	rows, err := r.Db.QueryContext(ctx, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	var genresArray []int

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.Id,
			&g.Genre,
		)
		if err != nil {
			return nil, nil, err
		}

		genres = append(genres, &g)
		genresArray = append(genresArray, g.Id)
	}

	movie.Genres = genres
	movie.GenresArray = genresArray

	// get all genres
	query = `
		SELECT
			id, genre
		FROM genres
		ORDER BY genre
	`
	var allGenres []*models.Genre
	rows, err = r.Db.QueryContext(ctx, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.Id,
			&g.Genre,
		)
		if err != nil {
			return nil, nil, err
		}

		allGenres = append(allGenres, &g)
	}

	return &movie, allGenres, nil
}

func (r *PostgresDbRepo) AllGenres() ([]*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// get all genres
	query := `
		SELECT
			id, genre, created_at, updated_at
		FROM genres
		ORDER BY genre
	`
	var genres []*models.Genre
	rows, err := r.Db.QueryContext(ctx, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.Id,
			&g.Genre,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	return genres, nil
}

func (r *PostgresDbRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT
			id, email, first_name, last_name, password, created_at, updated_at
		FROM
			users
		WHERE
			email = $1
	`

	var user models.User

	row := r.Db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresDbRepo) GetUserById(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT
			id, email, first_name, last_name, password, created_at, updated_at
		FROM
			users
		WHERE
			id = $1
	`

	var user models.User

	row := r.Db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}