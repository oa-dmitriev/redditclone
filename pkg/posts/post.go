package posts

import (
	"redditclone/pkg/user"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vote struct {
	UserID primitive.ObjectID `json:"user,string"`
	Value  int                `json:"vote"`
}

type CommentMessage struct {
	Comment string `json:"comment"`
}

type Comment struct {
	Created time.Time          `json:"created"`
	Author  *user.User         `json:"author"`
	Body    string             `json:"body"`
	ID      primitive.ObjectID `json:"_id,string" bson:"_id,omitempty"`
}

type Post struct {
	Score            int        `json:"score" bson:"score"`
	Views            uint64     `json:"views,string" bson:"views"`
	Type             string     `json:"type" bson:"type"`
	Title            string     `json:"title" bson:"title"`
	Author           *user.User `json:"author" bson:"author"`
	Category         string     `json:"category" bson:"category"`
	Text             string     `json:"text,omitempty" bson:"text"`
	URL              string     `json:"url,omitempty" bson:"url"`
	Votes            []*Vote    `json:"votes" bson:"votes"`
	Comments         []*Comment `json:"comments" bson:"comments,omitempty"`
	lastCommentID    uint64
	Created          time.Time          `json:"created" bson:"created"`
	UpvotePercentage uint64             `json:"upvotePercentage,string" bson:"upvotePercentage"`
	ID               primitive.ObjectID `json:"id,string" bson:"_id"`
}

type CreateForm struct {
	Category string `json:"category"`
	Text     string `json:"text"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}
