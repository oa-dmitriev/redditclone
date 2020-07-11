package handlers

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/posts"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PostsHandler struct {
	Tmpl      *template.Template
	PostsRepo *posts.PostsRepo
	Logger    *zap.SugaredLogger
}

func (h *PostsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	posts, err := h.PostsRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "unknown payload", http.StatusBadRequest)
		return
	}

	user, err := getCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}

	pd := posts.CreateForm{}
	err = json.Unmarshal(body, &pd)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	post, err := h.PostsRepo.NewPost(user, &pd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) DelPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["POST_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	err = h.PostsRepo.DelPost(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"message":"success"}`))
}

func (h *PostsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["POST_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	post, err := h.PostsRepo.GetByID(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["USER_LOGIN"]
	post, err := h.PostsRepo.GetByUser(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) Comment(w http.ResponseWriter, r *http.Request) {
	user, err := getCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}

	message := posts.CommentMessage{}
	err = json.Unmarshal(body, &message)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["POST_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}

	post, err := h.PostsRepo.Comment(postID, user, message.Comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) DelComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["POST_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	commentID, err := strconv.ParseUint(vars["COMMENT_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}

	post, err := h.PostsRepo.DelComment(postID, commentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) Category(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryName := vars["CATEGORY_NAME"]
	posts, err := h.PostsRepo.GetByCategory(categoryName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["POST_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	user, err := getCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	post, err := h.PostsRepo.Upvote(user, postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["POST_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	user, err := getCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	post, err := h.PostsRepo.Downvote(user, postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *PostsHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.ParseUint(vars["POST_ID"], 10, 64)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	user, err := getCurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	post, err := h.PostsRepo.Unvote(user, postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(post)
	if err != nil {
		http.Error(w, Err("server error").Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}
