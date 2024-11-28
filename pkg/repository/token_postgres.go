package repository

import (
	"database/sql"
	"fmt"
	"time"
	"todo"

	"github.com/jmoiron/sqlx"
)


type TokenPostgres struct {
	db *sqlx.DB
}

func NewTokenPostgres(db *sqlx.DB) *TokenPostgres {
	return &TokenPostgres{db: db,}
}

func (r *TokenPostgres) Create(userID int, token todo.RefreshToken) error {
	query := fmt.Sprintf("INSERT INTO %s (user_id, token, expires_date) VALUES ($1, $2, $3)", refreshTokens)
	_, err := r.db.Exec(query, userID, token.Token, token.ExpiresDate)
	return err
}

func (r *TokenPostgres) Get(token string) (todo.RefreshToken, error) {
	var t todo.RefreshToken
	query := fmt.Sprintf("SELECT id, user_id, token, expires_date FROM %s WHERE token = $1", refreshTokens)
	row := r.db.QueryRow(query, token)
	if err := row.Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresDate); err != nil {
		if err == sql.ErrNoRows {
			return t, fmt.Errorf("token not found: %w", err)
		}
		return t, fmt.Errorf("failed to get token: %w", err)
	}
	return t, nil
}

func (r *TokenPostgres) Update(token todo.RefreshToken, id int) error {
	query := fmt.Sprintf("UPDATE %s SET token = $1, expires_date = $2 WHERE id = $3", refreshTokens)
	_, err := r.db.Exec(query, token.Token, token.ExpiresDate, id)
	return err
}

func (r *TokenPostgres) DeleteByUserId(userId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE user_id = $1", refreshTokens)
	_, err := r.db.Exec(query, userId)
	return err
}

func (r *TokenPostgres) DeleteExpired(currentDate time.Time) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE expires_date < $1", refreshTokens)
	_, err := r.db.Exec(query, currentDate)
	return err
}