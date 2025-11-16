package authutil

import (
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

const (
	DefaultUserIDKey = "userId"
	defaultLoginURL  = "/users/signin"
)

type authMiddleware struct {
	sessionMng *scs.SessionManager

	userIDkey string
	loginURL  string
}

func NewAuthMiddleware(sesMng *scs.SessionManager) *authMiddleware {
	return &authMiddleware{sessionMng: sesMng, userIDkey: DefaultUserIDKey, loginURL: defaultLoginURL}
}

func (am *authMiddleware) WithUserIDKey(key string) {
	am.userIDkey = key
}

func (am *authMiddleware) WithLoginURL(url string) {
	am.loginURL = url
}

func (am *authMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("[authMiddleware] middleware start")
		defer slog.Debug("[authMiddleware] middleware end")
		uID := am.sessionMng.GetInt64(r.Context(), am.userIDkey)

		if uID <= 0 {
			slog.Debug("[authMiddleware] user not logged in", "user_id", uID)
			http.Redirect(w, r, am.loginURL, http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
