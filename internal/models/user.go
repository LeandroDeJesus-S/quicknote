package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID        pgtype.Numeric `json:"id"`
	Email     pgtype.Text    `json:"email"`
	Password  pgtype.Text    `json:"password"`
	Active    pgtype.Bool    `json:"active"`
	CreatedAt pgtype.Date    `json:"created_at"`
	UpdatedAt pgtype.Date    `json:"updated_at"`
}

type UserConfirmationToken struct {
	ID        pgtype.Numeric `json:"id"`
	UserID    pgtype.Numeric `json:"user_id"`
	Token     pgtype.Text    `json:"token"`
	Confirmed pgtype.Bool    `json:"confirmed"`
	CreatedAt pgtype.Date    `json:"created_at"`
	UpdatedAt pgtype.Date    `json:"updated_at"`
}
