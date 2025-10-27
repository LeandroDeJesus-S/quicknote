// Package handler provides HTTP handlers.
package handler

import (
	"log/slog"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
)

// ErrorHandler is a custom HTTP handler that returns an error.
type ErrorHandler func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements the http.Handler interface.
// It executes the ErrorHandler and handles any errors returned.
func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}

	switch v := err.(type) {
	case errs.HTTPError:
		slog.Debug(v.Error(), "sourceErr", v.Unwrap())
		http.Error(w, v.Error(), v.Code())
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
