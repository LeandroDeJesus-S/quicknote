package handler

import (
	"fmt"

	"github.com/LeandroDeJesus-S/quicknote/internal/models"
)

type NoteDTO struct {
	ID      int
	Title   string
	Content string
	Color   string
}

func newNoteDTO(note models.Note) NoteDTO {
	return NoteDTO{
		ID:      int(note.ID.Int.Int64()),
		Title:   note.Title.String,
		Content: note.Content.String,
		Color:   note.Color.String,
	}
}

func newNoteDTOList(notes []models.Note) []NoteDTO {
	var dtos []NoteDTO
	for _, note := range notes {
		dtos = append(dtos, newNoteDTO(note))
	}
	return dtos
}

type NoteRequestDTO struct {
	ID      int
	Title   string
	Content string
	Color   string
	Colors  []string
}

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
