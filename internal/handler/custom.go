package handler

import (
	"log/slog"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request) error

func (h ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}

	switch v := err.(type) {
	case errs.HTTPError:
		slog.Debug(v.Error(), "sourceErr", v.Unwrap().Error())
		http.Error(w, v.Error(), v.Code())
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
