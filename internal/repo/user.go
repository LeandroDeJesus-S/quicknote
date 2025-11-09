package repo

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDuplicatedEmail           = errs.NewRepoError(errors.New("email not available"))
	ErrConfirmationTokenNotFound = errs.NewRepoError(errors.New("confirmation token not found"))
	ErrUserNotFound              = errs.NewRepoError(errors.New("user not found"))
)

type UserRepository interface {
	Create(ctx context.Context, email, password string) (*models.User, error)
	CreateUserToken(ctx context.Context, userID int64, token string) (*models.UserConfirmationToken, error)
	ConfirmUserWithToken(ctx context.Context, token string) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, email, password string) (*models.User, error) {
	var u models.User
	u.Email = pgtype.Text{String: email, Valid: email != ""}
	u.Password = pgtype.Text{String: password, Valid: password != ""}

	query := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, created_at;"
	if err := r.db.QueryRow(ctx, query, u.Email, u.Password).Scan(&u.ID, &u.CreatedAt); err != nil {
		if strings.Contains(err.Error(), "violates unique constraint") {
			return nil, ErrDuplicatedEmail
		}
		return nil, errs.NewRepoError(err)
	}
	return &u, nil
}

func (r *UserRepo) CreateUserToken(ctx context.Context, userID int64, token string) (*models.UserConfirmationToken, error) {
	var u models.UserConfirmationToken

	u.Token = pgtype.Text{String: token, Valid: true}
	u.UserID = pgtype.Numeric{Int: big.NewInt(userID), Valid: true}
	query := "INSERT INTO user_tokens (user_id, token) VALUES ($1, $2) RETURNING id, created_at, updated_at;"
	if err := r.db.QueryRow(ctx, query, u.UserID, u.Token).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, errs.NewRepoError(err)
	}
	return &u, nil
}

func (r *UserRepo) ConfirmUserWithToken(ctx context.Context, token string) error {
	query := "select u.id, t.id from users u inner join user_tokens t on u.id = t.user_id where t.confirmed = false and t.token = $1 and u.active = false"
	var userID, tokenID pgtype.Numeric
	if err := r.db.QueryRow(ctx, query, token).Scan(&userID, &tokenID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConfirmationTokenNotFound
		}
		return errs.NewRepoError(err)
	}

	if _, err := r.db.Exec(ctx, "UPDATE users SET active = TRUE, updated_at = now() WHERE id = $1", userID); err != nil {
		return errs.NewRepoError(err)
	}
	if _, err := r.db.Exec(ctx, "UPDATE user_tokens SET confirmed = TRUE, updated_at = now() WHERE id = $1", tokenID); err != nil {
		return errs.NewRepoError(err)
	}
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	u.Email = pgtype.Text{String: email, Valid: true}
	query := "SELECT id, password, active, created_at, updated_at FROM users WHERE email = $1"
	if err := r.db.QueryRow(ctx, query, u.Email).Scan(&u.ID, &u.Password, &u.Active, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, errs.NewRepoError(err)
	}
	return &u, nil
}
