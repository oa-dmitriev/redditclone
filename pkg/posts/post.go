package posts

import (
	"redditclone/pkg/user"
	"time"
)

type Vote struct {
	UserID string `json:"user"`
	Value  int    `json:"vote"`
}

type CommentMessage struct {
	Comment string `json:"comment"`
}

type Comment struct {
	Created time.Time  `json:"created"`
	Author  *user.User `json:"author"`
	Body    string     `json:"body"`
	ID      uint64     `json:"id,string"`
}

type Post struct {
	Score            int        `json:"score"`
	Views            uint64     `json:"views,string"`
	Type             string     `json:"type"`
	Title            string     `json:"title"`
	Author           *user.User `json:"author"`
	Category         string     `json:"category"`
	Text             string     `json:"text,omitempty"`
	URL              string     `json:"url,omitempty"`
	Votes            []*Vote    `json:"votes"`
	Comments         []*Comment `json:"comments"`
	lastCommentID    uint64
	Created          time.Time `json:"created"`
	UpvotePercentage uint64    `json:"upvotePercentage,string"`
	ID               uint64    `json:"id,string"`
}

type CreateForm struct {
	Category string `json:"category"`
	Text     string `json:"text"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}
