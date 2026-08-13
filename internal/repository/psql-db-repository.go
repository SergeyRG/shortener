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

// generate:reset
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
	query := "Insert into urls (id, url, user_id, deleted_flag) VALUES ($1, $2, $3, false)"
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

func (r *PSQLDBRepositoryURL) GetByID(ctx context.Context, id string) (model.ShortenModel, error) {
	query := "SELECT url, user_id, deleted_flag from urls where id = $1"
	row := r.stor.QueryRowContext(ctx, query, id)
	var data model.ShortenModel
	if err := row.Scan(&data.OriginURL, &data.UserID, &data.DeletedFlag); err != nil {
		return model.ShortenModel{}, ErrURLNotFound
	}
	data.ID = id
	return data, nil
}

func (r *PSQLDBRepositoryURL) GetURLSCount(ctx context.Context) (int, error) {
	query := "SELECT count(*) from urls"
	row := r.stor.QueryRowContext(ctx, query)
	var urlsCount int
	if err := row.Scan(&urlsCount); err != nil {
		return 0, err
	}

	return urlsCount, nil
}

func (r *PSQLDBRepositoryURL) GetUsersCount(ctx context.Context) (int, error) {
	query := "SELECT count(DISTINCT user_id) AS unique_users_count FROM urls;"
	row := r.stor.QueryRowContext(ctx, query)
	var usersCount int
	if err := row.Scan(&usersCount); err != nil {
		return 0, err
	}

	return usersCount, nil
}

func (r *PSQLDBRepositoryURL) GetByUserID(ctx context.Context, userID string) ([]model.ShortenModel, error) {
	query := "SELECT id, url, user_id, deleted_flag from urls where user_id = $1"
	rows, err := r.stor.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userURL []model.ShortenModel

	for rows.Next() {
		var data model.ShortenModel
		if err := rows.Scan(&data.ID, &data.OriginURL, &data.UserID, &data.DeletedFlag); err != nil {
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
	tx, err := r.stor.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO urls (id, url, user_id, deleted_flag) VALUES($1,$2,$3, false)`
	stmt, err := tx.PrepareContext(ctx, query)
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
	return tx.Commit()
}

func (r *PSQLDBRepositoryURL) DeleteBatch(data []model.DeleteTaskDto) error {
	for _, d := range data {
		query := `Update urls SET deleted_flag=true WHERE user_id = $1 
          	  AND id = ANY($2) 
              AND deleted_flag = false;`

		_, err := r.stor.Exec(query, d.UserID, d.IDs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PSQLDBRepositoryURL) Close() error {
	return r.stor.Close()
}
