package dbrepo

import (
	"backend/internal/models"
	"database/sql"
)

type PostgresDbRepo struct {
	Db *sql.DB
}

func (r *PostgresDbRepo) AllMovies() ([]*models.Movie, error) {
	var movies []*models.Movie

	return movies, nil
}