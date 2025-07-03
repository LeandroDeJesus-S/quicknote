package handler

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
)

type noteHandler struct {}

func NewNoteHandler() *noteHandler {
	return new(noteHandler)
}

func (noteHandler) ListNotes(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("teste", "123")
	w.Header().Set("teste", "456")
	fmt.Fprint(w, "List Notes")
	return nil
}

func (noteHandler) NotesDetail(w http.ResponseWriter, r *http.Request) error {
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
	
	err = tpl.ExecuteTemplate(w, "base.html", map[string]string{"noteName": noteId, "noteContent": "-"})
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error executing template")
	}
	return nil
}

func (noteHandler) NotesCreate(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		tpl, err := template.ParseFiles(
			"view/templates/base.html",
			"view/templates/create.html",
		)
		if err != nil {
			return errs.NewHTTPError(err, http.StatusInternalServerError, "cannot parse templates")
		}
		if err := tpl.ExecuteTemplate(w, "base.html", nil); err != nil {
			return errs.NewHTTPError(err, http.StatusInternalServerError, "error executing template")
		}
	}
	fmt.Fprint(w, "Notes Create")
	return nil
}
