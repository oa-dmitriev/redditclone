package session

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SessionsManager struct {
	DB          *sql.DB
	maxlifetime int64
}

func NewSessionsMem(db *sql.DB) *SessionsManager {
	return &SessionsManager{DB: db}
}

func (sm *SessionsManager) Check(w http.ResponseWriter, r *http.Request) (*Session, error) {
	session_id, err := r.Cookie("token")
	if err == http.ErrNoCookie {
		log.Println("Check: NO COOKIE")
		return nil, ErrNoAuth
	}
	sess := Session{}
	var temp string
	err = sm.DB.
		QueryRow("SELECT id, user_id, username FROM sessions WHERE id = ?", session_id.Value).
		Scan(&sess.Sid, &temp, &sess.Username)
	if err != nil {
		log.Println("Check: QUERY, ", err.Error())
		return nil, ErrNoAuth
	}
	sess.UserID, err = primitive.ObjectIDFromHex(temp)
	if err != nil {
		log.Println("couldn't convert")
		return nil, fmt.Errorf("server error")
	}
	return &sess, nil
}

func (sm *SessionsManager) Create(w http.ResponseWriter, userID primitive.ObjectID, userLogin string) (*Session, error) {
	sess, err := NewSession(userID, userLogin)
	if err != nil {
		fmt.Println("Create: newSession", err.Error())
		return nil, err
	}
	_, err = sm.DB.Exec(
		"INSERT INTO sessions (id, user_id, username) VALUES (?, ?, ?)",
		sess.Sid, primitive.ObjectID(sess.UserID).Hex(), sess.Username,
	)
	if err != nil {
		fmt.Println("Create: insert ", err.Error())
		return nil, err
	}
	cookie := &http.Cookie{
		Name:    "token",
		Value:   sess.Sid,
		Expires: time.Now().Add(10 * time.Minute),
		Path:    "/",
	}
	http.SetCookie(w, cookie)
	return sess, nil
}

func (sm *SessionsManager) DestroyCurrent(w http.ResponseWriter, r *http.Request) error {
	sess, err := SessionFromContext(r.Context())
	if err != nil {
		return err
	}
	_, err = sm.DB.Exec(
		"DELETE FROM sessions WHERE id=?",
		sess.Sid,
	)
	if err != nil {
		return err
	}
	cookie := http.Cookie{
		Name:    "token",
		Expires: time.Now().AddDate(0, 0, -1),
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
	return nil
}
