package middleware

import (
	"net/http"
	"redditclone/pkg/session"
)

var (
	AuthUrls = map[string]struct{}{
		"/createpost": struct{}{},
	}
)

func Auth(sm *session.SessionsManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := AuthUrls[r.URL.Path]; !ok {
			next.ServeHTTP(w, r)
			return
		}
		_, err := sm.Check(r)
		if err != nil {
			http.Redirect(w, r, "/", 302)
			return
		}
		next.ServeHTTP(w, r)
	})
}
