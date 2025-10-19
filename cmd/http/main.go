package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/config"
	"github.com/LeandroDeJesus-S/quicknote/config/db"
	"github.com/LeandroDeJesus-S/quicknote/internal/handler"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
)

func main() {
	conf := config.MustLoadConfig()

	pool := db.MustConnect(context.Background(), conf.DatabaseURL)
	defer pool.Close()

	slog.SetDefault(config.NewLogger(conf.LoggerOut(), conf.LoggerLevel()))
	slog.Info("logger level set to", "level", conf.LoggerLevel())

	slog.Info("configurations loaded successfully", "server_host", conf.ServerHost, "server_port", conf.ServerPort)

	noteRepo := repo.NewNoteRepo(pool)

	mux := http.NewServeMux()

	noteHandler := handler.NewNoteHandler(noteRepo)

	staticHandle := http.FileServer(http.Dir("view/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandle))

	mux.Handle("/", handler.ErrorHandler(handler.HomeHandler))
	mux.Handle("GET /notes", handler.ErrorHandler(noteHandler.ListNotes))
	mux.Handle("GET /notes/{id}", handler.ErrorHandler(noteHandler.NotesDetail))
	mux.Handle("GET /notes/create", handler.ErrorHandler(noteHandler.NotesCreate))
	mux.Handle("DELETE /notes/{id}", handler.ErrorHandler(noteHandler.NotesDelete))
	mux.Handle("POST /notes", handler.ErrorHandler(noteHandler.Save))
	mux.Handle("GET /notes/{id}/edit", handler.ErrorHandler(noteHandler.NotesUpdate))

	lh := func(h http.Handler) http.Handler {
		return http.HandlerFunc((func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("", "method", r.Method, "pattern", r.URL.Path)
			h.ServeHTTP(w, r)
		}))
	}
	http.ListenAndServe(fmt.Sprintf("%s:%s", conf.ServerHost, conf.ServerPort), lh(mux))
}
