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

func (r *PSQLDBRepositoryURL) Add(ctx context.Context, url string, id string, userID string) error {
	query := "Insert into urls (id, url, user_id) VALUES ($1, $2, $3)"
	_, err := r.stor.ExecContext(ctx, query, id, url, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExist
		}
		return fmt.Errorf("%w:%w", ErrUnexpected, err)
	}
	return nil
}

func (r *PSQLDBRepositoryURL) GetByID(ctx context.Context, id string) ([]string, error) {
	query := "SELECT url, user_id from urls where id = $1"
	row := r.stor.QueryRowContext(ctx, query, id)
	var url string
	var userID string
	if err := row.Scan(&url, &userID); err != nil {
		return nil, ErrURLNotFound
	}

	return []string{url, userID}, nil
}

func (r *PSQLDBRepositoryURL) GetByUserID(ctx context.Context, userID string) ([]model.ShortenData, error) {
	query := "SELECT id, url from urls where user_id = $1"
	rows, err := r.stor.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userURL []model.ShortenData

	for rows.Next() {
		var data model.ShortenData
		if err := rows.Scan(&data.ID, &data.OriginURL); err != nil {
			return nil, err
		}
		userURL = append(userURL, data)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return userURL, nil
}

func (r *PSQLDBRepositoryURL) Delete(ctx context.Context, id string) error {
	query := "Delete from urls where id = $1"
	_, err := r.stor.ExecContext(ctx, query, id)
	return err
}

func (r *PSQLDBRepositoryURL) AddBatch(ctx context.Context, data []model.ShortenData, userID string) error {
	TX, err := r.stor.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer TX.Rollback()

	query := `INSERT INTO urls (id, url, user_id) VALUES($1,$2,$3)`
	stmt, err := TX.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range data {
		_, sqlErr := stmt.ExecContext(ctx, v.ID, v.OriginURL, userID)
		if sqlErr != nil {
			return sqlErr
		}
	}
	return TX.Commit()
}
