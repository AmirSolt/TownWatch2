// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package authmodels

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Otp struct {
	ID        pgtype.UUID
	CreatedAt pgtype.Timestamptz
	ExpiresAt pgtype.Timestamptz
	IsActive  bool
	UserID    pgtype.UUID
}

type User struct {
	ID        pgtype.UUID
	CreatedAt pgtype.Timestamptz
	Email     string
}
