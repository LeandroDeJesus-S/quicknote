package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/config"
	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/handler"
)

func homeHandler(w http.ResponseWriter, r *http.Request) error {
	tpl, err := template.ParseFiles(
		"view/templates/base.html",
		"view/templates/pages/home.html",
	)
	if err != nil {
		slog.Error("cannot parse templates", "err", err.Error())
		return errs.NewHTTPError(err, http.StatusInternalServerError, "cannot parse templates")
	}

	err = tpl.ExecuteTemplate(w, "base.html", nil)
	if err != nil {
		slog.Error("error executing template", "err", err.Error())
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error executing template")
	}
	return nil
}

func main() {
	conf := config.MustLoadConfig()

	slog.SetDefault(config.NewLogger(conf.LoggerOut(), conf.LoggerLevel()))
	slog.Info("configurations loaded successfully", "server_host", conf.ServerHost, "server_port", conf.ServerPort)

	mux := http.NewServeMux()

	noteHandler := handler.NewNoteHandler()

	staticHandle := http.FileServer(http.Dir("view/static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandle))

	mux.Handle("/", handler.ErrorHandler(homeHandler))
	mux.Handle("/notes", handler.ErrorHandler(noteHandler.ListNotes))
	mux.Handle("/notes/detail", handler.ErrorHandler(noteHandler.NotesDetail))
	mux.Handle("/notes/create", handler.ErrorHandler(noteHandler.NotesCreate))

	http.ListenAndServe(fmt.Sprintf("%s:%s", conf.ServerHost, conf.ServerPort), mux)
}
