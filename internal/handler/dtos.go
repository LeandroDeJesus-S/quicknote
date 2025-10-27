// Package handler provides HTTP handlers.
package handler

import (
	"fmt"

	"github.com/LeandroDeJesus-S/quicknote/internal/models"
)

// NoteDTO is a data transfer object for a note.
type NoteDTO struct {
	ID      int
	Title   string
	Content string
	Color   string
}

// newNoteDTO creates a new NoteDTO from a models.Note.
func newNoteDTO(note models.Note) NoteDTO {
	return NoteDTO{
		ID:      int(note.ID.Int.Int64()),
		Title:   note.Title.String,
		Content: note.Content.String,
		Color:   note.Color.String,
	}
}

// newNoteDTOList creates a new list of NoteDTOs from a list of models.Note.
func newNoteDTOList(notes []models.Note) []NoteDTO {
	var dtos []NoteDTO
	for _, note := range notes {
		dtos = append(dtos, newNoteDTO(note))
	}
	return dtos
}

// NoteRequestDTO is a data transfer object for a note request.
type NoteRequestDTO struct {
	ID      int
	Title   string
	Content string
	Color   string
	Colors  []string
}

// newNoteRequestDTO creates a new NoteRequestDTO with default values.
func newNoteRequestDTO() NoteRequestDTO {
	colors := []string{}
	for i := range 9 {
		colors = append(colors, fmt.Sprintf("color%d", i+1))
	}
	return NoteRequestDTO{
		Color:  "Color3",
		Colors: colors,
	}
}
