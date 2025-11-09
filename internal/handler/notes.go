// Package handler provides HTTP handlers.
package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/repo"
	"github.com/LeandroDeJesus-S/quicknote/internal/support"
	"github.com/LeandroDeJesus-S/quicknote/internal/validation"
)

// noteHandler handles HTTP requests for notes.
type noteHandler struct {
	noteRepo repo.Noter
}

// NewNoteHandler creates a new noteHandler.
func NewNoteHandler(noteRepo repo.Noter) *noteHandler {
	return &noteHandler{noteRepo: noteRepo}
}

// ListNotes handles the request to list all notes.
func (nh noteHandler) ListNotes(w http.ResponseWriter, r *http.Request) error {
	notes, err := nh.noteRepo.List(r.Context())
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error listing notes")
	}

	return render(
		w,
		r,
		newRenderOpts().WithPage("list.html").WithData(newNoteDTOList(notes)),
	)
}

// NotesDetail handles the request to show the details of a note.
func (nh noteHandler) NotesDetail(w http.ResponseWriter, r *http.Request) error {
	slog.Debug("fetching note details")
	noteID := r.PathValue("id")

	id, err := strconv.Atoi(noteID)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "id is invalid")
	}

	note, err := nh.noteRepo.ReadOne(r.Context(), id)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error reading note")
	}

	slog.Debug("rendering note detail", "note_id", id)
	return render(
		w,
		r,
		newRenderOpts().WithPage("detail.html").WithData(
			map[string]any{"ID": id, "noteName": note.Title.String, "noteContent": note.Content.String},
		),
	)
}

// NotesCreate handles the request to show the create note page.
func (nh noteHandler) NotesCreate(w http.ResponseWriter, r *http.Request) error {
	return render(
		w,
		r,
		newRenderOpts().WithPage("create.html").WithData(
			map[string]any{"note": newNoteRequestDTO()},
		),
	)
}

// NotesDelete handles the request to delete a note.
func (nh noteHandler) NotesDelete(w http.ResponseWriter, r *http.Request) error {
	rawNoteID := r.PathValue("id")

	numID, err := strconv.Atoi(rawNoteID)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "invalid note id")
	}

	if err := nh.noteRepo.Delete(r.Context(), numID); err != nil {
		slog.Error("Failed to delete note", "err", err, "note_id", numID)
	}

	return nil
}

// NotesUpdate handles the request to show the update note page.
func (nh noteHandler) NotesUpdate(w http.ResponseWriter, r *http.Request) error {
	noteID := r.PathValue("id")

	id, err := strconv.Atoi(noteID)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "id is invalid")
	}

	note, err := nh.noteRepo.ReadOne(r.Context(), id)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error reading note")
	}
	noteR := newNoteRequestDTO()
	noteR.ID = id
	noteR.Title = note.Title.String
	noteR.Content = note.Content.String
	noteR.Color = note.Color.String

	return render(
		w,
		r,
		newRenderOpts().WithPage("note-edit.html").WithData(
			map[string]any{"note": noteR},
		),
	)
}

// Save handles the request to save a note.
func (nh noteHandler) Save(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "error parsing form")
	}
	defer r.Body.Close()

	rawID := r.PostForm.Get("id")
	id, err := strconv.Atoi(rawID)
	if err != nil && rawID != "" {
		return errs.NewHTTPError(err, http.StatusBadRequest, "id is invalid")
	}

	noteR := newNoteRequestDTO()
	noteR.ID = id
	noteR.Title = r.PostForm.Get("title")
	noteR.Content = r.PostForm.Get("content")
	noteR.Color = r.PostForm.Get("color")

	validator := validation.NewFormValidator()
	validator.AddValidator([]string{"title", "content", "color"}, validation.ValidateStringNotEmpty)

	validator.ValidateForm(r.PostForm)
	if !validator.Ok() {
		page := support.TernaryIf(id > 0, "note-edit.html", "create.html")
		render(
			w,
			r,
			newRenderOpts().WithPage(page).WithData(map[string]any{
				"FieldErrors": validator.FieldErrors(),
				"note":        noteR,
			}),
		)
		return nil
	}

	if id > 0 {
		note, err := nh.noteRepo.Update(r.Context(), id, map[string]any{
			"title":   r.PostForm.Get("title"),
			"content": r.PostForm.Get("content"),
			"color":   r.PostForm.Get("color"),
		})
		if err != nil {
			return errs.NewHTTPError(err, http.StatusInternalServerError, "error updating note")
		}

		http.Redirect(w, r, fmt.Sprintf("/notes/%d", note.ID.Int), http.StatusFound)
		return nil
	}

	newNote, err := nh.noteRepo.Create(
		r.Context(),
		r.PostForm.Get("title"),
		r.PostForm.Get("content"),
		r.PostForm.Get("color"),
	)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error creating note")
	}

	slog.Debug("note created successfully", "note_id", newNote.ID.Int)
	http.Redirect(w, r, fmt.Sprintf("/notes/%d", newNote.ID.Int), http.StatusFound)
	return nil
}
