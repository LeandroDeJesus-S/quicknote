package support

import (
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/internal/render"
	"github.com/alexedwards/scs/v2"
)

const (
	FlashMsgKey = "flashMsg"
	FlashTypKey = "flashTyp"

	FlashMsgInfo    = "info"
	FlashMsgWarn    = "warn"
	FlashMsgError   = "error"
	FlashMsgSuccess = "success"
)

type (
	flashMessageType string
	FlashMessage     struct {
		Message string
		Typ     string
	}
)

// TernaryIf returns t if the condition is true, f otherwise
func TernaryIf[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

// SendFlashMessage sends a flash message to the user
func SendFlashMessage(ses *scs.SessionManager, r *http.Request, typ, message string) {
	ses.Put(r.Context(), FlashMsgKey, message)
	ses.Put(r.Context(), FlashTypKey, typ)
}

// GetFlashMessage returns the flash message and type
func GetFlashMessage(ses *scs.SessionManager, r *http.Request) (string, string) {
	msg := ses.PopString(r.Context(), FlashMsgKey)
	typ := ses.PopString(r.Context(), FlashTypKey)
	return msg, typ
}

// TagFlashMessage returns a dynamic tag that renders the flash message
// the tag returns a map with the message and type accessed by the key "message" and "typ"
func TagFlashMessage(ses *scs.SessionManager) render.DynamicTag {
	return func(r *http.Request) any {
		msg, typ := GetFlashMessage(ses, r)
		return func() *FlashMessage {
			if msg == "" {
				return nil
			}
			return &FlashMessage{
				Message: msg,
				Typ:     typ,
			}
		}
	}
}
