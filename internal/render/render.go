// Package render provides rendering utilities.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/errs"
	"github.com/LeandroDeJesus-S/quicknote/view"
	"github.com/alexedwards/scs/v2"
)

type TemplateRender interface {
	Page(w http.ResponseWriter, r *http.Request, opts *renderOpts) error
	Mail(tplName string, data any) ([]byte, error)
	WithGlobalTag(name string, tag DynamicTag) *templateRender
}

type templateRender struct {
	globalTags map[string]DynamicTag
	useEmbedFS bool
}

func NewTemplateRender(sesMsg *scs.SessionManager) *templateRender {
	return &templateRender{
		globalTags: make(map[string]DynamicTag),
	}
}

// WithTag adds a new tag globally availabe to templates
func (tr *templateRender) WithGlobalTag(name string, tag DynamicTag) *templateRender {
	tr.globalTags[name] = tag
	return tr
}

// WithEmbedFS sets the template render to use the embedded filesystem from [view.Assets].
func (tr *templateRender) WithEmbedFS(enabled bool) *templateRender {
	tr.useEmbedFS = enabled
	return tr
}

// ParseAssets parses the given templates using the embedded filesystem or the filesystem.
// The patterns are relative to the view/templates/ directory.
func (tr *templateRender) ParseAssets(t *template.Template, patterns ...string) (*template.Template, error) {
	if tr.useEmbedFS {
		return t.ParseFS(view.Assets, patterns...)
	}
	patts := make([]string, len(patterns))
	for i, p := range patterns {
		patts[i] = fmt.Sprintf("view/%s", p)
	}
	return t.ParseFiles(patts...)
}

// render renders a template to the given http.ResponseWriter.
func (tr *templateRender) Page(w http.ResponseWriter, r *http.Request, opts *renderOpts) error {
	if opts == nil {
		opts = NewOpts()
	}

	tags := make(template.FuncMap)
	for name, tag := range tr.globalTags {
		tags[name] = tag(r)
	}
	for name, tag := range opts.tags {
		tags[name] = tag(r)
	}

	tpl := template.New("").Funcs(tags)
	tpl, err := tr.ParseAssets(
		tpl,
		"templates/base.html",
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

	w.WriteHeader(opts.status)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buff.WriteTo(w)
	return nil
}

// Mail renders a HTML template to a email content.
//
// tplName is the name of the template to render (relative to the view/templates/mail/ directory
// data is the data to pass to the template.
func (tr *templateRender) Mail(tplName string, data any) ([]byte, error) {
	buf := new(bytes.Buffer)
	t, err := tr.ParseAssets(template.New(""), fmt.Sprintf("templates/mail/%s", tplName))
	if err != nil {
		return nil, err
	}

	if err := t.Execute(buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
