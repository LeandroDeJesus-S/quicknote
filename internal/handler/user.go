// Package handler provides HTTP handlers.
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
	"github.com/LeandroDeJesus-S/quicknote/internal/support/authutil"
	"github.com/LeandroDeJesus-S/quicknote/internal/validation"
)

// userHandler handles HTTP requests for users.
type userHandler struct {
	repo     repo.UserRepository
	pwHasher authutil.PasswordHasher
}

// NewUserHandler creates a new userHandler.
func NewUserHandler(repo repo.UserRepository, pwHasher authutil.PasswordHasher) *userHandler {
	uh := &userHandler{repo: repo, pwHasher: pwHasher}
	return uh
}

// SignIn handles the request to show the sign-in page.
func (h *userHandler) SignIn(w http.ResponseWriter, r *http.Request) error {
	return render(w, newRenderOpts().WithPage("user-signin.html"))
}

// SignInPost handles the request to signin a user.
func (h *userHandler) SignInPost(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "failed to parse form")
	}

	validator := validation.NewFormValidator()
	validator.AddValidator(
		[]string{"email", "password"},
		validation.ValidateStringNotEmpty,
	)
	validator.AddValidator([]string{"email"}, validation.ValidateEmailPattern)

	validator.ValidateForm(r.PostForm)
	if !validator.Ok() {
		return render(
			w,
			newRenderOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	usr, err := h.repo.FindByEmail(r.Context(), r.PostForm.Get("email"))
	if err != nil {
		if !errors.Is(err, repo.ErrUserNotFound) {
			return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to verify credentials")
		}

		validator.AddError("email", "invalid credentials")
		return render(
			w,
			newRenderOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	if !usr.Active.Bool {
		validator.AddError("email", "your account is not active")
		return render(
			w,
			newRenderOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	if ok, err := h.pwHasher.CheckPassword(r.PostForm.Get("password"), usr.Password.String); !ok {
		validator.AddError("email", "invalid credentials")
		slog.Error("failed to verify credentials", "error", err)
		return render(
			w,
			newRenderOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
}

// SignUp handles the request to show the sign-up page.
func (h *userHandler) SignUp(w http.ResponseWriter, r *http.Request) error {
	return render(w, newRenderOpts().WithPage("user-signup.html"))
}

// SignUpPost handles the request to create a new user.
func (h *userHandler) SignUpPost(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "failed to parse form")
	}

	validator := validation.NewFormValidator()
	validator.AddValidator(
		[]string{"email", "password"},
		validation.ValidateStringNotEmpty,
	)
	validator.AddValidator([]string{"email"}, validation.ValidateEmailPattern)
	validator.AddValidator([]string{"password"}, validation.ValidateMinMaxLen(6, 20))

	validator.ValidateForm(r.PostForm)
	if !validator.Ok() {
		return render(
			w,
			newRenderOpts().WithPage("user-signup.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	hashedPw, err := h.pwHasher.HashPassword(r.PostForm.Get("password"))
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to hash password")
	}

	usr, err := h.repo.Create(r.Context(), r.PostForm.Get("email"), hashedPw)

	if errors.Is(err, repo.ErrDuplicatedEmail) {
		validator.AddError("email", "email not available")
		return render(
			w,
			newRenderOpts().WithPage("user-signup.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to create user")
	}

	token, err := h.repo.CreateUserToken(r.Context(), usr.ID.Int.Int64(), authutil.GenerateToken())
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to create user token")
	}

	slog.Debug("user created", "id", usr.ID, "email", usr.Email, "created_at", usr.CreatedAt)
	return render(
		w,
		newRenderOpts().WithPage("user-signup-success.html").WithData(map[string]any{"user_token": token.Token.String}),
	)
}

func (h *userHandler) Confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")
	if err := h.repo.ConfirmUserWithToken(r.Context(), token); err != nil {
		slog.Error("failed to confirm user token", "error", err)
		return errs.NewHTTPError(err, http.StatusBadRequest, "failed to confirm user")
	}
	return render(
		w,
		newRenderOpts().WithPage(
			"user-confirm.html",
		).WithData(
			"your regestration was successfuly cofirmed",
		),
	)
}
