package session

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Session struct {
	ID     string
	UserID string
}

func NewSession(userID string, userLogin string) (*Session, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": userLogin,
			"id":       userID,
		},
		"exp": time.Now().Add(90 * 24 * time.Hour),
	})
	tokenString, err := token.SignedString(ExampleTokenSecret)
	if err != nil {
		return nil, err
	}
	session := &Session{tokenString, userID}
	return session, nil
}

var (
	ErrNoAuth = errors.New("No session found")
)
