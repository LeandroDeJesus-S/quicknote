package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Noter interface {
	List() ([]models.Note, error)
	ReadOne(id int) (*models.Note, error)
	Create(title, content, color string) (*models.Note, error)
	Update(id int, data map[string]any) (*models.Note, error)
	Delete(id int) error
}

type noteRepo struct {
	db *pgxpool.Pool
}

func NewNoteRepo(db *pgxpool.Pool) Noter {
	return &noteRepo{db: db}
}
func (nr noteRepo) List() ([]models.Note, error) {
	rows, err := nr.db.Query(
		context.Background(),
		"SELECT id, title, content, color, created_at, updated_at FROM notes",
	)
	if err != nil {
		return nil, errs.NewRepoError(err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err = rows.Scan(&note.ID, &note.Title, &note.Content, &note.Color, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, errs.NewRepoError(err)
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (nr noteRepo) ReadOne(id int) (*models.Note, error) {
	row := nr.db.QueryRow(
		context.Background(),
		"SELECT id, title, content, color, created_at, updated_at FROM notes WHERE id = $1",
		id,
	)
	var note models.Note
	err := row.Scan(&note.ID, &note.Title, &note.Content, &note.Color, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, errs.NewRepoError(err)
	}
	return &note, nil
}

func (nr noteRepo) Create(title, content, color string) (*models.Note, error) {
	var note models.Note

	note.Title = pgtype.Text{String: title, Valid: title != ""}
	note.Content = pgtype.Text{String: content, Valid: content != ""}
	note.Color = pgtype.Text{String: color, Valid: color != ""}
	note.CreatedAt = pgtype.Date{Time: time.Now(), Valid: true}
	note.UpdatedAt = pgtype.Date{Time: time.Now(), Valid: true}

	row := nr.db.QueryRow(
		context.Background(),
		"INSERT INTO notes (title, content, color, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		title, content, color, note.CreatedAt, note.UpdatedAt,
	)

	if err := row.Scan(&note.ID); err != nil {
		return nil, errs.NewRepoError(err)
	}

	return &note, nil
}

func (nr noteRepo) Update(id int, data map[string]any) (*models.Note, error) {
	if data == nil {
		return nil, errs.NewRepoError(fmt.Errorf("no data to update"))
	}

	var note models.Note

	var query strings.Builder
	query.WriteString("UPDATE notes SET ")
	args := make([]any, 0, len(data)+1)

	i := 0
	for field, value := range data {
		query.WriteString(fmt.Sprintf("%d = $%d ", i, i+1))
		args = append(args, field, value)

		i += 2
	}

	query.WriteString(fmt.Sprintf("WHERE id = $%d", i))
	args = append(args, id)

	query.WriteString("RETURNING id, title, content, color, created_at, updated_at")

	row := nr.db.QueryRow(
		context.Background(),
		query.String(),
		args...,
	)

	err := row.Scan(&note.ID, &note.Title, &note.Content, &note.Color, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, errs.NewRepoError(err)
	}

	return &note, nil
}

func (nr noteRepo) Delete(id int) error {
	_, err := nr.db.Exec(
		context.Background(),
		"DELETE FROM notes WHERE id = $1",
		id,
	)
	return err
}
