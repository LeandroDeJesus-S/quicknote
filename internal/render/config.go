package render

import (
	"fmt"
	"html/template"
	"net/http"
)

// renderOpts represents the options for rendering a template.
type renderOpts struct {
	status    int
	page      string
	data      any
	csrfField template.HTML
	tags      map[string]DynamicTag
}

// NewOpts creates a new renderOpts with default values.
func NewOpts() *renderOpts {
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

func (ro *renderOpts) WithTag(name string, tag DynamicTag) *renderOpts {
	ro.tags[name] = tag
	return ro
}

// DynamicTag wraps a common template tag function for tags witch depends on http.Request context.
// It must be a function that returns another function
type (
	DynamicTag func(*http.Request) any
)
