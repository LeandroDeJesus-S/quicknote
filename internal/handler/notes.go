package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"text/template"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
)

type noteHandler struct {
	noteRepo repo.Noter
}

func NewNoteHandler(noteRepo repo.Noter) *noteHandler {
	return &noteHandler{noteRepo: noteRepo}
}

func (nh noteHandler) ListNotes(w http.ResponseWriter, r *http.Request) error {
	notes, err := nh.noteRepo.List()
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error listing notes")
	}

	if err := json.NewEncoder(w).Encode(notes); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error encoding notes")
	}

	return nil
}

func (nr noteHandler) NotesDetail(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return errs.NewHTTPError(nil, http.StatusMethodNotAllowed, "method not allowed")
	}

	tpl, err := template.ParseFiles(
		"view/templates/base.html",
		"view/templates/pages/detail.html",
	)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "cannot parse templates")
	}

	noteId := r.URL.Query().Get("id")
	if noteId == "" {
		return errs.NewHTTPError(nil, http.StatusBadRequest, "id is required")
	}

	id, err := strconv.Atoi(noteId)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "id is invalid")
	}

	note, err := nr.noteRepo.ReadOne(id)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error reading note")
	}

	err = tpl.ExecuteTemplate(w, "base.html", map[string]string{"noteName": note.Title.String, "noteContent": note.Color.String})
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error executing template")
	}
	return nil
}

func (nh noteHandler) NotesCreate(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		tpl, err := template.ParseFiles(
			"view/templates/base.html",
			"view/templates/pages/create.html",
		)
		if err != nil {
			return errs.NewHTTPError(err, http.StatusInternalServerError, "cannot parse templates")
		}
		if err := tpl.ExecuteTemplate(w, "base.html", nil); err != nil {
			return errs.NewHTTPError(err, http.StatusInternalServerError, "error executing template")
		}

		slog.Debug("template for create notes was executed successfully")
		return nil
	}

	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "error parsing form")
	}
	defer r.Body.Close()

	newNote, err := nh.noteRepo.Create(
		r.PostForm.Get("title"),
		r.PostForm.Get("content"),
		r.PostForm.Get("color"),
	)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error creating note")
	}

	slog.Debug("note created successfully", "note_id", newNote.ID.Int)
	http.Redirect(w, r, fmt.Sprintf("/notes/detail?id=%d", newNote.ID.Int), http.StatusFound)
	return nil
}
