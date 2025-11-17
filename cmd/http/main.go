package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/LeandroDeJesus-S/quicknote/config"
	"github.com/LeandroDeJesus-S/quicknote/config/db"
	"github.com/LeandroDeJesus-S/quicknote/internal/handler"
	"github.com/LeandroDeJesus-S/quicknote/internal/mail"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
	"github.com/LeandroDeJesus-S/quicknote/internal/support/authutil"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
)

func main() {
	conf := config.MustLoadConfig()

	pool := db.MustConnect(context.Background(), conf.DatabaseURL)
	defer pool.Close()

	slog.SetDefault(config.NewLogger(conf.LoggerOut(), conf.LoggerLevel()))
	slog.Info("logger level set to", "level", conf.LoggerLevel())

	slog.Info("configurations loaded successfully", "server_host", conf.ServerHost, "server_port", conf.ServerPort)

	mailer := mail.NewGOMailV2Mailer(
		mail.NewConfig().
			WithServer(conf.MailServer).
			WithPort(conf.MailPortInt()).
			// WithUsername(conf.MailUsername).
			// WithPassword(conf.MailPassword).
			WithDefaultFrom(conf.MailDefaultFrom),
	)

	noteRepo := repo.NewNoteRepo(pool)
	userRepo := repo.NewUserRepo(pool)

	pwHasher := authutil.NewBcryptHasher()
	sessionMng := scs.New()
	sessionMng.Lifetime = 1 * time.Hour
	sessionMng.Store = pgxstore.New(pool)
	pgxstore.NewWithCleanupInterval(pool, 12*time.Hour)

	mux := handler.NewMux(noteRepo, userRepo, pwHasher, sessionMng, mailer, conf)
	muxH := mux.WithMiddleware(
		sessionMng.LoadAndSave,
		csrf.Protect(
			[]byte(conf.SecretKey),
			csrf.TrustedOrigins([]string{"localhost:8000", "127.0.0.1:8000"}),
		),
		func(h http.Handler) http.Handler {
			return http.HandlerFunc((func(w http.ResponseWriter, r *http.Request) {
				slog.Debug("[log]", "method", r.Method, "pattern", r.URL.Path)
				h.ServeHTTP(w, r)
			}))
		},
	)

	http.ListenAndServe(
		fmt.Sprintf("%s:%s", conf.ServerHost, conf.ServerPort),
		muxH,
	)
}
