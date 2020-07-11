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
	u, err := h.UserRepo.Register(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := h.Sessions.Create(w, u.ID, u.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(map[string]interface{}{
		"token": sess.ID,
	})
	if err != nil {
		http.Error(w, Err("bad request").Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]interface{}{
		"token": sess.ID,
	})
	if err != nil {
		http.Error(w, Err("bad request").Error(), http.StatusInternalServerError)
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
	inToken, err := r.Cookie("token")
	if err == http.ErrNoCookie {
		return nil, err
	}
	user := &user.User{}
	claims, err := getClaims(inToken.Value)
	_, ok := claims["user"]
	if !ok {
		return nil, Err("no user in claims")
	}
	userClaims, ok := claims["user"].(map[string]interface{})
	if !ok {
		return nil, Err("cant unpack payload")
	}
	user.ID, ok = userClaims["id"].(string)
	if !ok {
		return nil, Err("user id expected to be string")
	}
	user.Username, ok = userClaims["username"].(string)
	if !ok {
		return nil, Err("username expected to be string")
	}
	return user, nil
}

func Err(msg string) error {
	return fmt.Errorf(`{"message":"%s"}`, msg)
}
