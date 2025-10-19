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
	slog.Debug("fetching note details")
	noteId := r.PathValue("id")

	id, err := strconv.Atoi(noteId)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "id is invalid")
	}

	note, err := nr.noteRepo.ReadOne(r.Context(), id)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error reading note")
	}

	slog.Debug("rendering note detail", "note_id", id)
	return render(
		w,
		newRenderOpts().WithPage("detail.html").WithData(
			map[string]any{"ID": id, "noteName": note.Title.String, "noteContent": note.Content.String},
		),
	)
}

func (nh noteHandler) NotesCreate(w http.ResponseWriter, r *http.Request) error {
	return render(
		w,
		newRenderOpts().WithPage("create.html").WithData(newNoteRequestDTO()),
	)
}

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

func (nh noteHandler) NotesUpdate(w http.ResponseWriter, r *http.Request) error {
	noteId := r.PathValue("id")

	id, err := strconv.Atoi(noteId)
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
		newRenderOpts().WithPage("note-edit.html").WithData(
			noteR,
		),
	)
}

func (nh noteHandler) Save(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return errs.NewHTTPError(err, http.StatusBadRequest, "error parsing form")
	}
	defer r.Body.Close()

	id := r.PostForm.Get("id")
	if id != "" {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return errs.NewHTTPError(err, http.StatusBadRequest, "id is invalid")
		}

		note, err := nh.noteRepo.Update(r.Context(), idInt, map[string]any{
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
