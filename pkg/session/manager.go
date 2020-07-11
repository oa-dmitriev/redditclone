package session

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SessionsManager struct {
	data map[string]*Session
	mu   *sync.RWMutex
}

func NewSessionsMem() *SessionsManager {
	return &SessionsManager{
		data: make(map[string]*Session, 10),
		mu:   &sync.RWMutex{},
	}
}

func (sm *SessionsManager) Check(r *http.Request) (*Session, error) {
	inToken, err := r.Cookie("token")
	if err == http.ErrNoCookie {
		return nil, ErrNoAuth
	}
	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return ExampleTokenSecret, nil
	}
	token, err := jwt.Parse(inToken.Value, hashSecretGetter)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("bad token")
	}

	sm.mu.RLock()
	sess, ok := sm.data[inToken.Value]
	sm.mu.RUnlock()

	if !ok {
		return nil, ErrNoAuth
	}

	return sess, nil
}

var (
	ExampleTokenSecret = []byte("cfdsdf")
)

func (sm *SessionsManager) Create(w http.ResponseWriter, userID string, userLogin string) (*Session, error) {
	sess, err := NewSession(userID, userLogin)
	if err != nil {
		return nil, err
	}
	sm.mu.Lock()
	sm.data[sess.ID] = sess
	sm.mu.Unlock()

	cookie := &http.Cookie{
		Name:    "token",
		Value:   sess.ID,
		Expires: time.Now().Add(90 * 24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)
	return sess, nil
}
