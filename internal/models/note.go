package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Note struct {
	ID        pgtype.Numeric `json:"id"`
	Title     pgtype.Text    `json:"title"`
	Content   pgtype.Text    `json:"content"`
	Color     pgtype.Text    `json:"color"`
	CreatedAt pgtype.Date    `json:"created_at"`
	UpdatedAt pgtype.Date    `json:"updated_at"`
	UserID    pgtype.Numeric `json:"user_id"`
}
