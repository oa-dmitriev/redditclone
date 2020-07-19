package user

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"id"`
	Username string             `bson:"username"`
	password string             `bson:"password"`
}

type UserRepo struct {
	DB *sql.DB
}

type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (repo *UserRepo) Register(r *http.Request) (*User, error) {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err != nil {
		return nil, Err("error reading the body")
	}

	fd := AuthForm{}
	err = json.Unmarshal(body, &fd)
	if err != nil {
		return nil, Err("cant unpack payload")
	}

	user := User{
		ID:       primitive.NewObjectID(),
		Username: fd.Username,
		password: fd.Password,
	}

	_, err = repo.DB.Exec(
		"INSERT INTO users (id, username, password) VALUES (?, ?, ?)",
		primitive.ObjectID(user.ID).Hex(), user.Username, user.password,
	)
	if err != nil {
		return nil, Err(err.Error())
	}
	return &user, nil
}

func (repo *UserRepo) Authorize(r *http.Request) (*User, error) {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return nil, Err("error reading the body")
	}

	fd := AuthForm{}
	err = json.Unmarshal(body, &fd)
	if err != nil {
		return nil, Err("cant unpack payload")
	}
	user := User{}

	var id string
	err = repo.DB.
		QueryRow("SELECT id, username, password FROM users WHERE username=? AND password=?", fd.Username, fd.Password).
		Scan(&id, &user.Username, &user.password)
	if err != nil {
		return nil, Err("invalid password")
	}
	user.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, Err("server error")
	}
	return &user, nil
}

func Err(msg string) error {
	return fmt.Errorf(`{"message":"%s"}`, msg)
}
