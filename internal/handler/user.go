// Package handler provides HTTP handlers.
package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/mail"
	"github.com/LeandroDeJesus-S/quicknote/internal/render"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
	"github.com/LeandroDeJesus-S/quicknote/internal/support"

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
	mailer mail.Mailer

	appDomain string
}

// NewUserHandler creates a new userHandler.
func NewUserHandler(repo repo.UserRepository, pwHasher authutil.PasswordHasher, sesMng *scs.SessionManager, render render.TemplateRender, mailer mail.Mailer, appDomain string) *userHandler {
	uh := &userHandler{repo: repo, pwHasher: pwHasher, sesMng: sesMng, render: render, mailer: mailer, appDomain: appDomain}
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

	// testar se a funcionalidade de reenvio de token está ativada
	if !usr.Active.Bool {
		validator.AddError("email", "your account is not active")
		tok, err := h.repo.UserPendingToken(r.Context(), usr.ID.Int.Int64())
		if err != nil {
			return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to get user pending token")
		}
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-signin.html").WithData(map[string]any{
				"FieldErrors":    validator.FieldErrors(),
				"FormData":       map[string]string{"email": r.PostForm.Get("email")},
				"askResendToken": true,
				"userToken":      tok.Token.String,
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

	tokURL := fmt.Sprintf("%s/users/confirm/%s", h.appDomain, token.Token.String)
	body, err := h.render.Mail("confirmation.html", tokURL)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to render confirmation email")
	}

	if err := h.mailer.Send(mail.Message{
		To:      []string{usr.Email.String},
		Subject: "Your confirmation token",
		Body:    body,
		IsHTML:  true,
	}); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to send confirmation email")
	}

	slog.Debug("user created", "id", usr.ID, "email", usr.Email, "created_at", usr.CreatedAt, "tokURL", tokURL)
	http.Redirect(w, r, "/users/signup-success", http.StatusSeeOther)
	return nil
}

func (h *userHandler) SignUpSuccess(w http.ResponseWriter, r *http.Request) error {
	return h.render.Page(w, r, render.NewOpts().WithPage("user-signup-success.html"))
}

func (h *userHandler) Confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")

	if err := h.repo.ConfirmUserWithToken(r.Context(), token); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to confirm user")
	}

	msg, typ := "your registration was successfuly cofirmed", support.FlashMsgSuccess
	support.SendFlashMessage(h.sesMng, r, typ, msg)
	return h.render.Page(w, r, render.NewOpts().WithPage("generic-message.html"))
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

func (h *userHandler) EmailForm(w http.ResponseWriter, r *http.Request) error {
	data := make(map[string]any)

	switch r.URL.Query().Get("sub") {
	case "forgot-password":
		data["Title"] = "Forgot Password"
		data["Action"] = "/users/forgot-password"
	case "resend-token":
		data["Title"] = "Resend Token"
		data["Action"] = "/users/resend-token"
	default:
		return errs.NewHTTPError(errors.New("invalid sub"), http.StatusBadRequest, "invalid sub")
	}

	return h.render.Page(w, r, render.NewOpts().WithPage("user-email-form.html").WithData(data))
}

func (h *userHandler) ForgotPasswordPost(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "failed to parse form")
	}

	validator := validation.NewFormValidator()
	validator.AddValidator(
		[]string{"email"},
		validation.ValidateStringNotEmpty,
	)
	validator.AddValidator([]string{"email"}, validation.ValidateEmailPattern)

	validator.ValidateForm(r.PostForm)
	if !validator.Ok() {
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-email-form.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	usr, err := h.repo.FindByEmail(r.Context(), r.PostForm.Get("email"))
	if err != nil {
		validator.AddError("email", "invalid email")
		slog.Error("failed to find user", "error", err)
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-email-form.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": r.PostForm.Get("email")},
			}),
		)
	}

	tok, err := h.repo.CreateUserToken(r.Context(), usr.ID.Int.Int64(), authutil.GenerateToken())
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to create user token")
	}

	tokURL := fmt.Sprintf("%s/users/reset-password/%s", h.appDomain, tok.Token.String)
	body, err := h.render.Mail("forgot-password.html", tokURL)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to render confirmation email")
	}
	if err := h.mailer.Send(mail.Message{
		To:      []string{usr.Email.String},
		Subject: "Reset your password",
		Body:    body,
		IsHTML:  true,
	}); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to send confirmation email")
	}

	support.SendFlashMessage(h.sesMng, r, support.FlashMsgSuccess, "Almost there, check your email to reset your password")
	http.Redirect(w, r, "/users/signin", http.StatusSeeOther)
	return nil
}

// ResetPassword renders the reset password page, which the user is redirected by clicking the link in the email
func (h *userHandler) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	token := r.PathValue("token")
	if strings.TrimSpace(token) == "" {
		return errs.NewHTTPError(errors.New("pw token not sent"), http.StatusBadRequest, "invalid token")
	}

	if err := h.repo.CheckResetToken(r.Context(), token); err != nil {
		slog.Error("failed to check pw token", "error", err)
		if errors.Is(err, repo.ErrTokenExpired) {
			support.SendFlashMessage(h.sesMng, r, support.FlashMsgError, "your token has expired, please try again")
			http.Redirect(w, r, "/users/forgot-password", http.StatusSeeOther)
		}
		return errs.NewHTTPError(err, http.StatusBadRequest, err.Error())
	}

	return h.render.Page(
		w,
		r,
		render.NewOpts().
			WithPage("user-reset-pw.html").
			WithData(map[string]any{"token": token}),
	)
}

func (h *userHandler) ResetPasswordPost(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "failed to parse form")
	}

	validator := validation.NewFormValidator()
	validator.AddValidator(
		[]string{"new_password", "password_confirm", "token"},
		validation.ValidateStringNotEmpty,
	)
	validator.AddValidator([]string{"new_password"}, validation.ValidateMinMaxLen(6, 20))

	if r.PostForm.Get("new_password") != r.PostForm.Get("password_confirm") {
		validator.AddError("password_confirm", "passwords do not match")
	}

	validator.ValidateForm(r.PostForm)
	if !validator.Ok() {
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-reset-pw.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"token":       r.PostForm.Get("token"),
			}),
		)
	}

	token := r.PostForm.Get("token")
	if err := h.repo.CheckResetToken(r.Context(), token); err != nil {
		slog.Error("failed to check pw token", "error", err)
		return errs.NewHTTPError(err, http.StatusBadRequest, err.Error())
	}

	newPW, err := h.pwHasher.HashPassword(r.PostForm.Get("new_password"))
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to hash password")
	}

	usrMail, err := h.repo.UpdatePasswordByToken(r.Context(), token, newPW)
	if err != nil {
		slog.Error("failed to update password", "error", err)
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to update password")
	}

	if err := h.mailer.Send(mail.Message{
		To:      []string{usrMail},
		Subject: "Password changed",
		Body:    []byte("Your password was successfully changed"),
	}); err != nil {
		slog.Error("failed to send confirmation email", "error", err)
	}

	support.SendFlashMessage(h.sesMng, r, support.FlashMsgSuccess, "your password was successfully changed, you can now sign in")
	http.Redirect(w, r, "/users/signin", http.StatusSeeOther)
	return nil
}

func (h *userHandler) ResendToken(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "failed to parse form")
	}

	email := r.PostForm.Get("email")
	validator := validation.NewFormValidator()
	validator.AddValidator(
		[]string{"email"},
		validation.ValidateStringNotEmpty,
	)
	validator.AddValidator([]string{"email"}, validation.ValidateEmailPattern)

	validator.ValidateForm(r.PostForm)
	if !validator.Ok() {
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-email-form.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": email},
			}),
		)
	}

	usr, err := h.repo.FindByEmail(r.Context(), email)
	if err != nil {
		validator.AddError("email", "invalid email")
		slog.Error("failed to find user", "error", err)
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-email-form.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": email},
			}),
		)
	}

	// checks for pending token
	pendingTok, err := h.repo.UserPendingToken(r.Context(), usr.ID.Int.Int64())
	if err != nil && !errors.Is(err, repo.ErrConfirmationTokenNotFound) {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to get user pending token")
	}
	if err != nil {
		validator.AddError("email", "invalid email")
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-email-form.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": email},
			}),
		)
	}
	err = h.repo.CheckResetToken(r.Context(), pendingTok.Token.String)
	if err != nil && !errors.Is(err, repo.ErrTokenExpired) {
		validator.AddError("email", "invalid email")
		return h.render.Page(
			w,
			r,
			render.NewOpts().WithPage("user-email-form.html").WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"FormData":    map[string]string{"email": email},
			}),
		)
	}

	newTok := authutil.GenerateToken()
	if err := h.repo.UpdateUserToken(r.Context(), pendingTok.ID.Int.Int64(), newTok); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to update token")
	}

	tokURL := fmt.Sprintf("%s/users/confirm/%s", h.appDomain, newTok)
	body, err := h.render.Mail("confirmation.html", tokURL)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to render confirmation email")
	}

	if err := h.mailer.Send(mail.Message{
		To:      []string{email},
		Subject: "Your new confirmation token",
		Body:    body,
		IsHTML:  true,
	}); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "failed to send confirmation email")
	}

	support.SendFlashMessage(h.sesMng, r, support.FlashMsgSuccess, "your token was successfully sent")
	http.Redirect(w, r, "/users/signup-success", http.StatusSeeOther)
	return nil
}
