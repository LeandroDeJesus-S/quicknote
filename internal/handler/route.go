package handler

import (
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/render"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
	"github.com/LeandroDeJesus-S/quicknote/internal/support/authutil"
	"github.com/alexedwards/scs/v2"
)

type Mux struct {
	*http.ServeMux

	noteRepo repo.Noter
	userRepo repo.UserRepository

	pwHasher   authutil.PasswordHasher
	sessionMng *scs.SessionManager
}

func NewMux(noteRepo repo.Noter, userRepo repo.UserRepository, pwHasher authutil.PasswordHasher, sessionMng *scs.SessionManager) *Mux {
	mux := &Mux{ServeMux: http.NewServeMux(), noteRepo: noteRepo, userRepo: userRepo, pwHasher: pwHasher, sessionMng: sessionMng}

	renderer := render.NewTemplateRender(sessionMng)
	renderer.WithGlobalTag("isAuthenticated", authutil.TagIsAuthenticated(sessionMng)).
		WithGlobalTag("csrfField", authutil.TagCSRFField)

	staticHandle := http.FileServer(http.Dir("view/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandle))

	homeHandler := NewHomeHandler(renderer)
	noteHandler := NewNoteHandler(noteRepo, renderer, sessionMng)
	userHandler := NewUserHandler(userRepo, mux.pwHasher, mux.sessionMng, renderer)

	authMiddleware := authutil.NewAuthMiddleware(mux.sessionMng)

	mux.Handle("/", ErrorHandler(homeHandler.Home))
	mux.Handle("GET /notes", authMiddleware.RequireAuth(ErrorHandler(noteHandler.ListNotes)))
	mux.Handle("GET /notes/{id}", authMiddleware.RequireAuth(ErrorHandler(noteHandler.NotesDetail)))
	mux.Handle("GET /notes/create", authMiddleware.RequireAuth(ErrorHandler(noteHandler.NotesCreate)))
	mux.Handle("DELETE /notes/{id}", authMiddleware.RequireAuth(ErrorHandler(noteHandler.NotesDelete)))
	mux.Handle("POST /notes", authMiddleware.RequireAuth(ErrorHandler(noteHandler.Save)))
	mux.Handle("GET /notes/{id}/edit", authMiddleware.RequireAuth(ErrorHandler(noteHandler.NotesUpdate)))

	mux.Handle("GET /users/signup", ErrorHandler(userHandler.SignUp))
	mux.Handle("POST /users/signup", ErrorHandler(userHandler.SignUpPost))

	mux.Handle("GET /users/signin", ErrorHandler(userHandler.SignIn))
	mux.Handle("POST /users/signin", ErrorHandler(userHandler.SignInPost))

	mux.Handle("GET /users/confirm/{token}", ErrorHandler(userHandler.Confirm))
	mux.Handle("GET /users/signout", authMiddleware.RequireAuth(ErrorHandler(userHandler.SignOut)))
	mux.Handle("GET /users/me", authMiddleware.RequireAuth(ErrorHandler(userHandler.Me)))

	return mux
}

func (m *Mux) WithMiddleware(mw ...func(http.Handler) http.Handler) http.Handler {
	var wm http.Handler = m.ServeMux
	for i := len(mw) - 1; i >= 0; i-- {
		wm = mw[i](wm)
	}
	return wm
}
