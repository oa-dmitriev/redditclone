package user

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	ID       string
	Username string
	password string
}

type UserRepo struct {
	data   map[string]*User
	lastID uint32
	mux    *sync.RWMutex
}

type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		data: map[string]*User{
			"oleg": &User{
				ID:       "0",
				Username: "oleg",
				password: "adminadmin",
			},
		},
		lastID: 0,
		mux:    &sync.RWMutex{},
	}
}

func (repo *UserRepo) Register(r *http.Request) (*User, error) {
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

	user := User{
		ID:       strconv.Itoa(int(repo.lastID)),
		Username: fd.Username,
		password: fd.Password,
	}

	repo.mux.Lock()
	repo.data[fd.Username] = &user
	repo.mux.Unlock()

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

	repo.mux.RLock()
	user, exist := repo.data[fd.Username]
	if !exist || fd.Password != user.password {
		return nil, Err("invalid password")
	}
	repo.mux.RUnlock()

	return user, nil
}

func Err(msg string) error {
	return fmt.Errorf(`{"message":"%s"}`, msg)
}
