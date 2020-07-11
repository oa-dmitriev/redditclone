package posts

import (
	"fmt"
	"redditclone/pkg/user"
	"sync"
	"sync/atomic"
	"time"
)

type PostsRepo struct {
	lastID uint64
	data   map[uint64]*Post
	mux    *sync.RWMutex
}

func NewRepo() *PostsRepo {
	return &PostsRepo{
		lastID: 0,
		data:   make(map[uint64]*Post),
		mux:    &sync.RWMutex{},
	}
}

func (repo *PostsRepo) GetAll() ([]*Post, error) {
	posts := make([]*Post, 0, 10)

	repo.mux.Lock()
	for _, val := range repo.data {
		posts = append(posts, val)
	}
	repo.mux.Unlock()

	return posts, nil
}

func (repo *PostsRepo) GetByCategory(category string) ([]*Post, error) {
	posts := make([]*Post, 0, 10)

	repo.mux.Lock()
	for _, val := range repo.data {
		if val.Category == category {
			posts = append(posts, val)
		}
	}
	repo.mux.Unlock()

	return posts, nil
}

func (repo *PostsRepo) NewPost(u *user.User, pd *CreateForm) (*Post, error) {
	repo.mux.Lock()
	post := Post{
		Score:    1,
		Type:     pd.Type,
		Title:    pd.Title,
		Author:   u,
		Category: pd.Category,
		Text:     pd.Text,
		URL:      pd.URL,
		Votes: []*Vote{
			&Vote{
				UserID: u.ID,
				Value:  1,
			},
		},
		Created:          time.Now(),
		UpvotePercentage: 100,
		ID:               repo.lastID,
	}
	repo.data[post.ID] = &post
	repo.lastID++
	repo.mux.Unlock()
	return &post, nil
}

func (repo *PostsRepo) DelPost(ID uint64) error {

	repo.mux.RLock()
	_, ok := repo.data[ID]
	repo.mux.RUnlock()

	if ok {
		repo.mux.Lock()
		delete(repo.data, ID)
		repo.mux.Unlock()
		return nil
	}
	return Err("no post to delete")
}

func (repo *PostsRepo) GetByID(ID uint64) (*Post, error) {
	repo.mux.RLock()
	post, ok := repo.data[ID]
	repo.mux.RUnlock()

	if ok {
		atomic.AddUint64(&post.Views, 1)
		return post, nil
	}
	return nil, Err("no post with that id")
}

func (repo *PostsRepo) GetByUser(username string) ([]*Post, error) {
	posts := make([]*Post, 0, 10)
	repo.mux.Lock()
	for _, v := range repo.data {
		if username == v.Author.Username {
			posts = append(posts, v)
		}
	}
	repo.mux.Unlock()
	return posts, nil
}

func (repo *PostsRepo) Comment(ID uint64, user *user.User, message string) (*Post, error) {
	repo.mux.RLock()
	post, ok := repo.data[ID]
	repo.mux.RUnlock()

	if ok {
		repo.mux.Lock()
		comment := &Comment{
			Created: time.Now(),
			Author:  user,
			Body:    message,
			ID:      post.lastCommentID,
		}
		post.lastCommentID++
		post.Comments = append(post.Comments, comment)
		repo.mux.Unlock()
		return post, nil
	}
	return nil, Err("no post with that id")
}

func (repo *PostsRepo) DelComment(postID, commentID uint64) (*Post, error) {

	post, err := repo.GetByID(postID)
	if err != nil {
		Err("no post found")
	}

	repo.mux.Lock()
	for i, c := range post.Comments {
		if c.ID == commentID {
			post.Comments[i] = post.Comments[len(post.Comments)-1]
			post.Comments = post.Comments[:len(post.Comments)-1]
			return post, nil
		}
	}
	repo.mux.Unlock()
	return nil, Err("no comment found")
}

func (repo *PostsRepo) Upvote(user *user.User, postID uint64) (*Post, error) {
	post, err := repo.GetByID(postID)
	if err != nil {
		return nil, Err("no post found")
	}

	repo.mux.Lock()
	for _, v := range post.Votes {
		if v.UserID == user.ID {
			post.Score -= v.Value
			post.Score++
			v.Value = 1
			repo.mux.Unlock()
			return post, nil
		}
	}
	post.Votes = append(post.Votes, &Vote{user.ID, 1})
	post.Score += 1
	repo.mux.Unlock()

	return post, nil
}

func (repo *PostsRepo) Downvote(user *user.User, postID uint64) (*Post, error) {
	post, err := repo.GetByID(postID)
	if err != nil {
		return nil, Err("no such post")
	}

	repo.mux.Lock()
	for _, v := range post.Votes {
		if v.UserID == user.ID {
			post.Score -= v.Value
			post.Score--
			v.Value = -1
			repo.mux.Unlock()
			return post, nil
		}
	}
	post.Votes = append(post.Votes, &Vote{user.ID, -1})
	post.Score -= 1
	repo.mux.Unlock()
	return post, nil
}

func (repo *PostsRepo) Unvote(user *user.User, postID uint64) (*Post, error) {
	post, err := repo.GetByID(postID)
	if err != nil {
		return nil, Err("no such post")
	}

	repo.mux.Lock()
	for i, v := range post.Votes {
		if v.UserID == user.ID {
			post.Score -= v.Value
			post.Votes[i] = post.Votes[len(post.Votes)-1]
			post.Votes = post.Votes[:len(post.Votes)-1]
			repo.mux.Unlock()
			return post, nil
		}
	}
	repo.mux.Unlock()
	return post, nil
}

func Err(msg string) error {
	return fmt.Errorf(`{"message":"%s"}`, msg)
}
