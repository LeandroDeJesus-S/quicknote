package handler

import (
	"fmt"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/config"
	"github.com/LeandroDeJesus-S/quicknote/internal/mail"
	"github.com/LeandroDeJesus-S/quicknote/internal/render"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
	"github.com/LeandroDeJesus-S/quicknote/internal/support"
	"github.com/LeandroDeJesus-S/quicknote/internal/support/authutil"
	"github.com/alexedwards/scs/v2"
)

type Mux struct {
	*http.ServeMux
}

func NewMux(noteRepo repo.Noter, userRepo repo.UserRepository, pwHasher authutil.PasswordHasher, sessionMng *scs.SessionManager, mailer mail.Mailer, conf *config.Config) *Mux {
	mux := &Mux{ServeMux: http.NewServeMux()}

	renderer := render.NewTemplateRender(sessionMng)
	renderer.WithGlobalTag("isAuthenticated", authutil.TagIsAuthenticated(sessionMng)).
		WithGlobalTag("csrfField", authutil.TagCSRFField).
		WithGlobalTag("flashMessage", support.TagFlashMessage(sessionMng))

	errH := ErrorHandler{Render: renderer, Sess: sessionMng}

	staticHandle := http.FileServer(http.Dir("view/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandle))

	homeHandler := NewHomeHandler(renderer)
	noteHandler := NewNoteHandler(noteRepo, renderer, sessionMng)
	userHandler := NewUserHandler(
		userRepo,
		pwHasher,
		sessionMng,
		renderer,
		mailer,
		mountAppDomain(conf),
	)

	authMiddleware := authutil.NewAuthMiddleware(sessionMng)

	mux.Handle("/", errH.Wrap(homeHandler.Home))
	mux.Handle("GET /notes", authMiddleware.RequireAuth(errH.Wrap(noteHandler.ListNotes)))
	mux.Handle("GET /notes/{id}", authMiddleware.RequireAuth(errH.Wrap(noteHandler.NotesDetail)))
	mux.Handle("GET /notes/create", authMiddleware.RequireAuth(errH.Wrap(noteHandler.NotesCreate)))
	mux.Handle("DELETE /notes/{id}", authMiddleware.RequireAuth(errH.Wrap(noteHandler.NotesDelete)))
	mux.Handle("POST /notes", authMiddleware.RequireAuth(errH.Wrap(noteHandler.Save)))
	mux.Handle("GET /notes/{id}/edit", authMiddleware.RequireAuth(errH.Wrap(noteHandler.NotesUpdate)))

	mux.Handle("GET /users/signup", errH.Wrap(userHandler.SignUp))
	mux.Handle("POST /users/signup", errH.Wrap(userHandler.SignUpPost))
	mux.Handle("GET /users/signup-success", errH.Wrap(userHandler.SignUpSuccess))

	mux.Handle("GET /users/signin", errH.Wrap(userHandler.SignIn))
	mux.Handle("POST /users/signin", errH.Wrap(userHandler.SignInPost))

	mux.Handle("GET /users/confirm/{token}", errH.Wrap(userHandler.Confirm))
	mux.Handle("GET /users/signout", authMiddleware.RequireAuth(errH.Wrap(userHandler.SignOut)))
	mux.Handle("GET /users/me", authMiddleware.RequireAuth(errH.Wrap(userHandler.Me)))
	mux.Handle("GET /users/email-form", errH.Wrap(userHandler.EmailForm))
	mux.Handle("POST /users/forgot-password", errH.Wrap(userHandler.ForgotPasswordPost))
	mux.Handle("GET /users/reset-password/{token}", errH.Wrap(userHandler.ResetPassword))
	mux.Handle("POST /users/reset-password", errH.Wrap(userHandler.ResetPasswordPost))
	mux.Handle("POST /users/resend-token", errH.Wrap(userHandler.ResendToken))

	return mux
}

func (m *Mux) WithMiddleware(mw ...func(http.Handler) http.Handler) http.Handler {
	var wm http.Handler = m.ServeMux
	for i := len(mw) - 1; i >= 0; i-- {
		wm = mw[i](wm)
	}
	return wm
}

func mountAppDomain(conf *config.Config) string {
	scheme := "https"
	if conf.DebugMode() {
		scheme = "http"
	}

	host := conf.ServerHost
	if conf.DebugMode() {
		host = "localhost"
	}

	return fmt.Sprintf("%s://%s:%s", scheme, host, conf.ServerPort)
}
