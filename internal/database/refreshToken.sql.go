// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: refreshToken.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const getRefreshToken = `-- name: GetRefreshToken :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at FROM refresh_tokens
WHERE token = $1
`

func (q *Queries) GetRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getRefreshToken, token)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const insertRefreshToken = `-- name: InsertRefreshToken :exec
INSERT INTO
  refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES
  ($1, NOW(), NOW(), $2, $3, $4)
`

type InsertRefreshTokenParams struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt time.Time
}

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) error {
	_, err := q.db.ExecContext(ctx, insertRefreshToken,
		arg.Token,
		arg.UserID,
		arg.ExpiresAt,
		arg.RevokedAt,
	)
	return err
}
