// Package handler provides HTTP handlers.
package handler

import (
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/render"
)

type homeHandler struct {
	render render.TemplateRender
}

// NewHomeHandler creates a new homeHandler.
func NewHomeHandler(renderer render.TemplateRender) *homeHandler {
	h := &homeHandler{render: renderer}
	return h
}

// HomeHandler handles the home page.
func (h homeHandler) Home(w http.ResponseWriter, r *http.Request) error {
	return h.render.Page(
		w,
		r,
		render.NewOpts().WithPage("home.html"),
	)
}
