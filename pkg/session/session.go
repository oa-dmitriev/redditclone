package session

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type sessKey string

var (
	ErrNoAuth          error   = errors.New("No session found")
	SessionKey         sessKey = "sessionKey"
	ExampleTokenSecret         = []byte("cfdsdf")
)

type Session struct {
	Sid      string
	UserID   primitive.ObjectID
	Username string
}

func NewSession(userID primitive.ObjectID, userLogin string) (*Session, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": userLogin,
			"id":       userID,
		},
		"exp": time.Now().Add(10 * time.Minute),
	})
	tokenString, err := token.SignedString(ExampleTokenSecret)
	if err != nil {
		return nil, err
	}
	session := &Session{tokenString, userID, userLogin}
	return session, nil
}

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(SessionKey).(*Session)
	if !ok || sess == nil {
		return nil, ErrNoAuth
	}
	return sess, nil
}
