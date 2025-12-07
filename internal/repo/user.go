package repo

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

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
	ErrTokenExpired              = errs.NewRepoError(errors.New("token expired"))
	ErrTokenAlreadyConfirmed     = errs.NewRepoError(errors.New("token already confirmed"))

	pwTokenTTL = 1 * time.Minute
)

type UserRepository interface {
	Create(ctx context.Context, email, password string) (*models.User, error)                               // creates a new user
	CreateUserToken(ctx context.Context, userID int64, token string) (*models.UserConfirmationToken, error) // creates a new token for a user
	ConfirmUserWithToken(ctx context.Context, token string) error                                           // fetches a user's token where it's neither confirmed nor expired then marks it as confirmed
	FindByEmail(ctx context.Context, email string) (*models.User, error)                                    // finds a user by its email
	CheckResetToken(ctx context.Context, token string) error                                                // returns [ErrConfirmationTokenNotFound] error if the token was not found, [ErrTokenAlreadyConfirmed] if it was already confirmed, and [ErrTokenExpired] if it's expired
	UpdatePasswordByToken(ctx context.Context, token, newPassword string) (string, error)                   // set the new password for the token owner and returns its email
	UpdateUserToken(ctx context.Context, oldTokID int64, newTok string) error                               // updates the token for the new one
	UserEmailByToken(ctx context.Context, token string) (string, error)                                     // returns the user's email by the token
	UserPendingToken(ctx context.Context, userID int64) (*models.UserConfirmationToken, error)              // returns the user's pending token
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
	query := "SELECT u.id, t.id FROM users u INNER JOIN user_tokens t ON u.id = t.user_id WHERE t.confirmed = false AND t.token = $1 AND u.active = false"
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

func (r *UserRepo) CheckResetToken(ctx context.Context, token string) error {
	q := `SELECT confirmed, created_at FROM user_tokens WHERE token = $1`
	var (
		createdAt pgtype.Date
		confirmed bool
	)
	err := r.db.QueryRow(ctx, q, token).Scan(&confirmed, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrConfirmationTokenNotFound
	}
	if err := err; err != nil {
		return errs.NewRepoError(err)
	}

	if confirmed {
		return ErrTokenAlreadyConfirmed
	}

	if time.Since(createdAt.Time) > pwTokenTTL {
		return ErrTokenExpired
	}

	return nil
}

func (r *UserRepo) UpdatePasswordByToken(ctx context.Context, token, newPassword string) (string, error) {
	q := `UPDATE users SET password = $1, updated_at = now() WHERE id = (SELECT user_id FROM user_tokens WHERE token = $2) RETURNING email`
	var email string
	err := r.db.QueryRow(ctx, q, newPassword, token).Scan(&email)
	if err != nil {
		return "", errs.NewRepoError(err)
	}
	return email, nil
}

func (r *UserRepo) UpdateUserToken(ctx context.Context, oldTokID int64, newTok string) error {
	q := `UPDATE user_tokens SET updated_at = now(), token = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, q, newTok, pgtype.Numeric{Int: big.NewInt(oldTokID), Valid: true})
	if err != nil {
		return errs.NewRepoError(err)
	}
	return nil
}

func (r *UserRepo) UserEmailByToken(ctx context.Context, token string) (string, error) {
	q := `SELECT u.email FROM users u INNER JOIN user_tokens ut ON u.id = ut.user_id AND token = $1`
	var email string
	err := r.db.QueryRow(ctx, q, token).Scan(&email)
	if err != nil {
		return "", errs.NewRepoError(err)
	}
	return email, nil
}

func (r *UserRepo) UserPendingToken(ctx context.Context, userID int64) (*models.UserConfirmationToken, error) {
	var u models.UserConfirmationToken
	u.UserID = pgtype.Numeric{Int: big.NewInt(userID), Valid: true}
	query := "SELECT user_id, token, confirmed, created_at FROM user_tokens WHERE user_id = $1 AND confirmed = false"
	if err := r.db.QueryRow(ctx, query, u.UserID).Scan(&u.UserID, &u.Token, &u.Confirmed, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConfirmationTokenNotFound
		}
		return nil, errs.NewRepoError(err)
	}
	return &u, nil
}
