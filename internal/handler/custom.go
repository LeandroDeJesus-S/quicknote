// Package handler provides HTTP handlers.
package handler

import (
	"log/slog"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/internal/render"
	"github.com/LeandroDeJesus-S/quicknote/internal/support"
	"github.com/alexedwards/scs/v2"
)

// ErrorHandler is a custom HTTP handler that returns an error.
type ErrorHandler struct {
	f      func(w http.ResponseWriter, r *http.Request) error
	Render render.TemplateRender
	Sess   *scs.SessionManager
}

func (h *ErrorHandler) Wrap(hf func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	h.f = hf
	hc := *h
	return &hc
}

// ServeHTTP implements the http.Handler interface.
// It executes the ErrorHandler and handles any errors returned.
func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.f(w, r)
	if err == nil {
		return
	}

	switch v := err.(type) {
	case errs.HTTPError:
		slog.Debug("[ErrorHandler] "+v.Error(), "sourceErr", v.Unwrap())
		support.SendFlashMessage(h.Sess, r, support.FlashMsgError, v.Message())
		if err := h.Render.Page(w, r, render.NewOpts().WithPage("generic-message.html").WithStatus(v.Code())); err != nil {
			slog.Error("failed to render generic message", "error", err)
		}

	default:
		slog.Debug("[ErrorHandler]", "err", v)
		support.SendFlashMessage(h.Sess, r, support.FlashMsgError, "seomthing went wrong, please try again later")
		h.Render.Page(w, r, render.NewOpts().WithPage("generic-message.html").WithStatus(http.StatusInternalServerError))
	}
}
