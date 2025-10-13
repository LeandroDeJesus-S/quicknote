package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
	notes, err := nh.noteRepo.List(r.Context())
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error listing notes")
	}

	return render(
		w,
		newRenderOpts().WithPage("list.html").WithData(newNoteDTOList(notes)),
	)
}

func (nr noteHandler) NotesDetail(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return errs.NewHTTPError(nil, http.StatusMethodNotAllowed, "method not allowed")
	}

	noteId := r.URL.Query().Get("id")
	if noteId == "" {
		return errs.NewHTTPError(nil, http.StatusBadRequest, "id is required")
	}

	id, err := strconv.Atoi(noteId)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "id is invalid")
	}

	note, err := nr.noteRepo.ReadOne(r.Context(), id)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error reading note")
	}

	return render(
		w,
		newRenderOpts().WithPage("detail.html").WithData(
			map[string]string{"noteName": note.Title.String, "noteContent": note.Content.String},
		),
	)
}

func (nh noteHandler) NotesCreate(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return render(
			w,
			newRenderOpts().WithPage("create.html").WithData(newNoteRequestDTO()),
		)
	}

	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "error parsing form")
	}
	defer r.Body.Close()

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
	http.Redirect(w, r, fmt.Sprintf("/notes/detail?id=%d", newNote.ID.Int), http.StatusFound)
	return nil
}

func (nh noteHandler) NotesDelete(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		r.Header.Set("Allow", http.MethodDelete)
		return errs.NewHTTPError(nil, http.StatusMethodNotAllowed, "Method not allowed")
	}

	rawNoteID := r.URL.Query().Get("id")
	if rawNoteID == "" {
		return errs.NewHTTPError(nil, http.StatusBadRequest, "id is required")
	}

	numID, err := strconv.Atoi(rawNoteID)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "invalid note id")
	}

	if err := nh.noteRepo.Delete(r.Context(), numID); err != nil {
		slog.Error("Failed to delete note", "err", err, "note_id", numID)
	}

	return nil
}

func (nh noteHandler) NotesUpdate(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return render(
			w,
			newRenderOpts().WithPage("note-edit.html"),
		)
	}
	return nil
}
