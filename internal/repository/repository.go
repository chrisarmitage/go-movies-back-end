package repository

import (
	"backend/internal/models"
	"database/sql"
)

type DatabaseRepo interface {
	Connection() *sql.DB

	AllMovies() ([]*models.Movie, error)
	OneMovie(id int) (*models.Movie, error)
	OneMovieForEdit(id int) (*models.Movie, []*models.Genre, error)
	InsertMovie(movie models.Movie) (int, error)
	UpdateMovie(movie models.Movie) error
	UpdateMovieGenres(id int, genreIds []int) error

	AllGenres() ([]*models.Genre, error)

	GetUserByEmail(email string) (*models.User, error)
	GetUserById(id int) (*models.User, error)
}
