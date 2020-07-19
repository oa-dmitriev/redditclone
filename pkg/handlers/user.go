package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

type UserHandler struct {
	Tmpl     *template.Template
	UserRepo *user.UserRepo
	Sessions *session.SessionsManager
	Logger   *zap.SugaredLogger
}

func (h *UserHandler) Index(w http.ResponseWriter, r *http.Request) {
	err := h.Tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "unknown payload", http.StatusBadRequest)
		return
	}
	_, err := h.UserRepo.Register(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Login(w, r)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "unknown payload", http.StatusBadRequest)
		return
	}
	u, err := h.UserRepo.Authorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	sess, err := h.Sessions.Create(w, u.ID, u.Username)
	if err != nil {
		fmt.Println("CREATE FAILED MAINAIANIANA")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]interface{}{
		"token": sess.Sid,
	})
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func getClaims(inToken string) (jwt.MapClaims, error) {
	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, Err("bad sign method")
		}
		return session.ExampleTokenSecret, nil
	}
	token, err := jwt.Parse(inToken, hashSecretGetter)
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, Err("token invalid")
	}
	return claims, nil
}

func getCurrentUser(r *http.Request) (*user.User, error) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		return nil, Err("no session")
	}
	user := &user.User{ID: sess.UserID, Username: sess.Username}
	return user, nil
}

func Err(msg string) error {
	return fmt.Errorf(`{"message":"%s"}`, msg)
}
