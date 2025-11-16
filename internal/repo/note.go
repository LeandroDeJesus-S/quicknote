package repo

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var fieldNameRegex = regexp.MustCompile(`^[A-Za-z_]+$`)

type Noter interface {
	List(ctx context.Context, userID int64) ([]models.Note, error)
	ReadOne(ctx context.Context, tid int) (*models.Note, error)
	Create(ctx context.Context, userID int64, title, content, color string) (*models.Note, error)
	Update(ctx context.Context, id int, data map[string]any) (*models.Note, error)
	Delete(ctx context.Context, id int) error
}

type noteRepo struct {
	db *pgxpool.Pool
}

func NewNoteRepo(db *pgxpool.Pool) Noter {
	return &noteRepo{db: db}
}

func (nr noteRepo) List(ctx context.Context, userID int64) ([]models.Note, error) {
	rows, err := nr.db.Query(
		context.Background(),
		"SELECT id, title, content, color, created_at, updated_at FROM notes WHERE user_id = $1",
		userID,
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

func (nr noteRepo) ReadOne(ctx context.Context, id int) (*models.Note, error) {
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

func (nr noteRepo) Create(ctx context.Context, userID int64, title, content, color string) (*models.Note, error) {
	var note models.Note

	note.UserID = pgtype.Numeric{Int: big.NewInt(userID), Valid: true}
	note.Title = pgtype.Text{String: title, Valid: title != ""}
	note.Content = pgtype.Text{String: content, Valid: content != ""}
	note.Color = pgtype.Text{String: color, Valid: color != ""}
	note.CreatedAt = pgtype.Date{Time: time.Now(), Valid: true}
	note.UpdatedAt = pgtype.Date{Time: time.Now(), Valid: true}

	row := nr.db.QueryRow(
		context.Background(),
		"INSERT INTO notes (title, content, color, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		title, content, color, userID, note.CreatedAt, note.UpdatedAt,
	)

	if err := row.Scan(&note.ID); err != nil {
		return nil, errs.NewRepoError(err)
	}

	return &note, nil
}

func (nr noteRepo) Update(ctx context.Context, id int, data map[string]any) (*models.Note, error) {
	if data == nil {
		return nil, errs.NewRepoError(fmt.Errorf("no data to update"))
	}

	var note models.Note

	var query strings.Builder
	query.WriteString("UPDATE notes SET ")
	args := make([]any, 0, len(data)+1)

	i := 0
	nmap := len(data)

	for field, value := range data {
		if !fieldNameRegex.MatchString(field) {
			return nil, errs.NewRepoError(fmt.Errorf("invalid field name: %s", field))
		}

		if i == nmap-1 {
			query.WriteString(fmt.Sprintf("%s=$%d ", field, i+1))
			args = append(args, value)
		} else {
			query.WriteString(fmt.Sprintf("%s=$%d, ", field, i+1))
			args = append(args, value)
		}

		i++
	}

	query.WriteString(fmt.Sprintf("WHERE id=$%d ", i+1))
	args = append(args, id)

	query.WriteString("RETURNING id, title, content, color, created_at, updated_at")

	slog.Debug("updating note", "query", query.String(), "args", args)
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

func (nr noteRepo) Delete(ctx context.Context, id int) error {
	_, err := nr.db.Exec(
		context.Background(),
		"DELETE FROM notes WHERE id = $1",
		id,
	)
	return err
}
