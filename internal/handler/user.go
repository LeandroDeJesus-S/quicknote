// Package handler provides HTTP handlers.
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/render"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"

	"github.com/LeandroDeJesus-S/quicknote/internal/support/authutil"
	"github.com/LeandroDeJesus-S/quicknote/internal/validation"
	"github.com/alexedwards/scs/v2"
)

// userHandler handles HTTP requests for users.
type userHandler struct {
	sesMng   *scs.SessionManager
	repo     repo.UserRepository
	pwHasher authutil.PasswordHasher

	render render.TemplateRender
}

// NewUserHandler creates a new userHandler.
func NewUserHandler(repo repo.UserRepository, pwHasher authutil.PasswordHasher, sesMng *scs.SessionManager, render render.TemplateRender) *userHandler {
	uh := &userHandler{repo: repo, pwHasher: pwHasher, sesMng: sesMng, render: render}
	return uh
}

// SignIn handles the request to show the sign-in page.
func (h *userHandler) SignIn(w http.ResponseWriter, r *http.Request) error {
	return h.render.Page(w, r, render.NewOpts().WithPage("user-signin.html"))
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
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-signin.html").WithData(map[string]any{
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
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	if !usr.Active.Bool {
		validator.AddError("email", "your account is not active")
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	if ok, err := h.pwHasher.CheckPassword(r.PostForm.Get("password"), usr.Password.String); !ok {
		validator.AddError("email", "invalid credentials")
		slog.Error("failed to verify credentials", "error", err)
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	if err := h.sesMng.RenewToken(r.Context()); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to renew session token")
	}

	h.sesMng.Put(r.Context(), authutil.DefaultUserIDKey, usr.ID.Int.Int64())

	if _, _, err := h.sesMng.Commit(r.Context()); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to commit session")
	}

	slog.Debug("session after commit",
		"exists", h.sesMng.Exists(r.Context(), "userId"),
		"direct", h.sesMng.GetInt64(r.Context(), "userId"),
	)
	http.Redirect(w, r, "/notes", http.StatusSeeOther)
	return nil
}

// SignUp handles the request to show the sign-up page.
func (h *userHandler) SignUp(w http.ResponseWriter, r *http.Request) error {
	return h.render.Page(w, r, render.NewOpts().WithPage("user-signup.html"))
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
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-signup.html").WithData(map[string]any{
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
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-signup.html").WithData(map[string]any{
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
	return h.render.Page(
		w,
		r,
		render.NewOpts().WithPage("user-signup-success.html").WithData(map[string]any{"user_token": token.Token.String}),
	)
}

func (h *userHandler) Confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")
	if err := h.repo.ConfirmUserWithToken(r.Context(), token); err != nil {
		slog.Error("failed to confirm user token", "error", err)
		return errs.NewHTTPError(err, http.StatusBadRequest, "failed to confirm user")
	}
	return h.render.Page(
		w,
		r,
		render.NewOpts().WithPage(
			"user-confirm.html",
		).WithData(
			"your regestration was successfuly cofirmed",
		),
	)
}

func (h *userHandler) SignOut(w http.ResponseWriter, r *http.Request) error {
	if err := h.sesMng.Destroy(r.Context()); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to destroy session")
	}
	http.Redirect(w, r, "/users/signin", http.StatusSeeOther)
	return nil
}

func (h *userHandler) Me(w http.ResponseWriter, r *http.Request) error {
	return nil
}
