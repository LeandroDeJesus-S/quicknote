// Package handler provides HTTP handlers.
package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
)

// renderOpts represents the options for rendering a template.
type renderOpts struct {
	status int
	page   string
	data   any
}

// newRenderOpts creates a new renderOpts with default values.
func newRenderOpts() *renderOpts {
	std := &renderOpts{
		status: http.StatusOK,
		page:   "home.html",
		data:   nil,
	}
	return std
}

// WithStatus sets the HTTP status code for the response.
func (ro *renderOpts) WithStatus(s int) *renderOpts {
	ro.status = s
	return ro
}

// WithPage sets the page to be rendered.
func (ro *renderOpts) WithPage(p string) *renderOpts {
	ro.page = fmt.Sprintf("view/templates/pages/%s", p)
	return ro
}

// WithData sets the data to be passed to the template.
func (ro *renderOpts) WithData(d any) *renderOpts {
	ro.data = d
	return ro
}

// render renders a template to the given http.ResponseWriter.
func render(w http.ResponseWriter, opts *renderOpts) error {
	if opts == nil {
		opts = newRenderOpts()
	}
	tpl, err := template.ParseFiles(
		"view/templates/base.html",
		opts.page,
	)
	if err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "cannot parse templates")
	}

	// write to the buffer first to avoid susperfulous writing to the response writer
	buff := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(buff, "base.html", opts.data); err != nil {
		return errs.NewHTTPError(err, http.StatusInternalServerError, "error executing template")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buff.WriteTo(w)
	return nil
}
