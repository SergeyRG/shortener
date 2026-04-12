package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SergeyRG/shortener/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNilDB = errors.New("DB cant be nil")

type PSQLDBRepositoryURL struct {
	stor *sql.DB
}

func NewPSQLDBRepositoryURL(db *sql.DB) (*PSQLDBRepositoryURL, error) {
	if db != nil {
		return &PSQLDBRepositoryURL{
			stor: db,
		}, nil
	}
	return nil, ErrNilDB
}

func (r *PSQLDBRepositoryURL) Add(ctx context.Context, url string, id string) error {
	query := "Insert into urls (id, url) VALUES ($1, $2)"
	_, err := r.stor.ExecContext(ctx, query, id, url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlredyExist
		}
		return fmt.Errorf("%w:%w", ErrUnexpected, err)
	}
	return nil
}

func (r *PSQLDBRepositoryURL) GetByID(ctx context.Context, id string) (string, error) {
	query := "SELECT url from urls where id = $1"
	row := r.stor.QueryRowContext(ctx, query, id)
	var url string
	if err := row.Scan(&url); err != nil {
		return "", ErrURLNotFound
	}
	return url, nil
}

func (r *PSQLDBRepositoryURL) Delete(ctx context.Context, id string) error {
	query := "Delete from urls where id = $1"
	_, err := r.stor.ExecContext(ctx, query, id)
	return err
}

func (r *PSQLDBRepositoryURL) AddBatch(ctx context.Context, data []model.ShortenData) error {
	TX, err := r.stor.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer TX.Rollback()

	query := `INSERT INTO urls (id, url) VALUES($1,$2)`
	stmt, err := TX.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range data {
		_, sqlErr := stmt.ExecContext(ctx, v.ID, v.OriginURL)
		if sqlErr != nil {
			return sqlErr
		}
	}
	return TX.Commit()
}
