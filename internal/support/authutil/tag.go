package authutil

import (
	"html/template"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
)

func TagIsAuthenticated(ses *scs.SessionManager) render.DynamicTag {
	return func(r *http.Request) any {
		return func() bool { return ses.GetInt64(r.Context(), DefaultUserIDKey) > 0 }
	}
}

func TagCSRFField(r *http.Request) any {
	return func() template.HTML { return csrf.TemplateField(r) }
}
