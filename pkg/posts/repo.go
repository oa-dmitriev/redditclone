package posts

import (
	"context"
	"fmt"
	"redditclone/pkg/user"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostsRepo struct {
	data *mongo.Collection
}

func NewRepo(collection *mongo.Collection) *PostsRepo {
	return &PostsRepo{data: collection}
}

func (repo *PostsRepo) GetAll() ([]*Post, error) {
	posts := []*Post{}
	curs, err := repo.data.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	err = curs.All(context.TODO(), &posts)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (repo *PostsRepo) GetByCategory(category string) ([]*Post, error) {
	posts := []*Post{}
	curs, err := repo.data.Find(context.TODO(), bson.D{primitive.E{Key: "category", Value: category}})
	if err != nil {
		return nil, err
	}
	curs.All(context.TODO(), &posts)
	return posts, nil
}

func (repo *PostsRepo) NewPost(u *user.User, pd *CreateForm) (*Post, error) {
	post := Post{
		Score:    1,
		Type:     pd.Type,
		Title:    pd.Title,
		Author:   u,
		Category: pd.Category,
		Text:     pd.Text,
		URL:      pd.URL,
		Votes: []*Vote{
			{
				UserID: u.ID,
				Value:  1,
			},
		},
		Comments:         make([]*Comment, 0, 10),
		Created:          time.Now(),
		UpvotePercentage: 100,
		ID:               primitive.NewObjectID(),
	}
	_, err := repo.data.InsertOne(context.TODO(), &post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (repo *PostsRepo) DelPost(ID primitive.ObjectID) error {
	_, err := repo.data.DeleteOne(context.TODO(), bson.D{primitive.E{Key: "_id", Value: ID}})
	return err
}

func (repo *PostsRepo) GetByID(ID primitive.ObjectID) (*Post, error) {
	post := Post{}
	update := bson.D{primitive.E{Key: "$inc", Value: bson.D{primitive.E{Key: "views", Value: 1}}}}
	filter := bson.D{primitive.E{Key: "_id", Value: ID}}
	opts := options.FindOneAndUpdate().SetReturnDocument(1)
	err := repo.data.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&post)
	if err != nil {
		return nil, Err("no post with that id")
	}
	return &post, nil
}

func (repo *PostsRepo) GetByUser(username string) ([]*Post, error) {
	posts := []*Post{}
	filter := bson.D{primitive.E{Key: "author.username", Value: username}}
	curs, err := repo.data.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	curs.All(context.TODO(), &posts)
	return posts, nil
}

func (repo *PostsRepo) Comment(ID primitive.ObjectID, user *user.User, message string) (*Post, error) {
	post := Post{}
	comment := &Comment{
		Created: time.Now(),
		Author:  user,
		Body:    message,
		ID:      primitive.NewObjectID(),
	}

	update := bson.D{primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "comments", Value: comment}}}}
	filter := bson.D{primitive.E{Key: "_id", Value: ID}}
	opts := options.FindOneAndUpdate().SetReturnDocument(1)

	err := repo.data.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&post)
	if err != nil {
		return nil, Err("no post with that id")
	}
	return &post, nil
}

func (repo *PostsRepo) DelComment(postID, commentID primitive.ObjectID) (*Post, error) {
	update := bson.D{primitive.E{Key: "$pull",
		Value: bson.D{primitive.E{Key: "comments",
			Value: bson.D{primitive.E{Key: "_id", Value: commentID}}}}}}
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: postID,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(1)
	post := Post{}

	err := repo.data.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&post)
	if err != nil {
		return nil, Err("no comment found")
	}
	return &post, nil
}

func (repo *PostsRepo) Upvote(user *user.User, postID primitive.ObjectID) (*Post, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: postID}}
	opts := options.FindOneAndUpdate().SetReturnDocument(1)

	post := Post{}
	err := repo.data.FindOne(context.TODO(), filter).Decode(&post)
	if err != nil {
		return nil, Err("bd: couldn't fetch query")
	}
	for _, val := range post.Votes {
		if val.UserID == user.ID {
			post.Score -= val.Value - 1
			val.Value = 1
			_, err = repo.data.ReplaceOne(context.TODO(), filter, &post)
			if err != nil {
				return nil, Err("bd: couldn't execute query")
			}
			return &post, nil
		}
	}
	vote := Vote{user.ID, 1}
	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "votes", Value: &vote}}},
		primitive.E{Key: "$inc", Value: bson.D{primitive.E{Key: "score", Value: 1}}},
	}
	err = repo.data.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&post)
	if err != nil {
		return nil, Err("db error")
	}
	return &post, nil
}

func (repo *PostsRepo) Downvote(user *user.User, postID primitive.ObjectID) (*Post, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: postID}}
	opts := options.FindOneAndUpdate().SetReturnDocument(1)

	post := Post{}
	err := repo.data.FindOne(context.TODO(), filter).Decode(&post)
	if err != nil {
		return nil, Err("bd: couldn't fetch query")
	}
	for _, val := range post.Votes {
		if val.UserID == user.ID {
			post.Score -= val.Value + 1
			val.Value = -1
			_, err = repo.data.ReplaceOne(context.TODO(), filter, &post)
			if err != nil {
				return nil, Err("bd: couldn't execute query")
			}
			return &post, nil
		}
	}
	vote := Vote{user.ID, -1}
	update := bson.D{
		primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "votes", Value: &vote}}},
		primitive.E{Key: "$inc", Value: bson.D{primitive.E{Key: "score", Value: -1}}},
	}
	err = repo.data.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&post)
	if err != nil {
		return nil, Err("db error")
	}
	return &post, nil
}

func (repo *PostsRepo) Unvote(user *user.User, postID primitive.ObjectID) (*Post, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: postID}}
	post := Post{}
	err := repo.data.FindOne(context.TODO(), filter).Decode(&post)
	if err != nil {
		return nil, Err("bd: couldn't fetch query")
	}
	for i, val := range post.Votes {
		if val.UserID == user.ID {
			post.Score -= val.Value
			post.Votes[i] = post.Votes[len(post.Votes)-1]
			post.Votes = post.Votes[:len(post.Votes)-1]
			_, err = repo.data.ReplaceOne(context.TODO(), filter, &post)
			if err != nil {
				return nil, Err("bd: couldn't execute query")
			}
			return &post, nil
		}
	}
	return nil, Err("you haven't voted yet")
}

func Err(msg string) error {
	return fmt.Errorf(`{"message":"%s"}`, msg)
}
