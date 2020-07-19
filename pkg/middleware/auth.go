package middleware

import (
	"context"
	"log"
	"net/http"
	"redditclone/pkg/session"
)

var (
	NoAuthUrls = map[string]struct{}{
		"/static/js/2.d59deea0.chunk.js":      {},
		"/static/js/main.32ebaf54.chunk.js":   {},
		"/static/css/main.74225161.chunk.css": {},

		"/":             {},
		"/api/register": {},
		"/api/login":    {},

		"/api/posts/":            {},
		"/api/posts/music":       {},
		"/api/posts/funny":       {},
		"/api/posts/videos":      {},
		"/api/posts/programming": {},
		"/api/posts/news":        {},
		"/api/posts/fashion":     {},
	}
)

func Auth(sm *session.SessionsManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := NoAuthUrls[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}
		sess, err := sm.Check(w, r)
		if err != nil {
			log.Println("didnt pass check")
			http.Redirect(w, r, "/", 302)
			return
		}
		ctx := context.WithValue(r.Context(), session.SessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
