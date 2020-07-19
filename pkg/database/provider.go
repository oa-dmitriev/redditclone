package database

import (
	"context"
	"database/sql"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func GetMysql() (*sql.DB, error) {
	// dsn := "root:admin@tcp(localhost:3306)/redditclone?"
	dsn := "root:admin@tcp(db)/redditclone?"

	dsn += "charset=utf8"
	dsn += "&interpolateParams=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println("Ping ERROR")
		return nil, err
	}
	db.SetMaxOpenConns(10)
	_, err = db.Exec(`
	DROP TABLE IF EXISTS users
	`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users(
		id VARCHAR(255) NOT NULL UNIQUE, 
		username VARCHAR(255) NOT NULL UNIQUE, 
		password VARCHAR(255))
	`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
	DROP TABLE IF EXISTS sessions
	`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions(
		id VARCHAR(255) NOT NULL UNIQUE, 
		user_id VARCHAR(255), 
		username VARCHAR(255), 
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetCollection() (*mongo.Collection, error) {
	// client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongodb"))

	if err != nil {
		return nil, err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		return nil, err
	}
	collection := client.Database("redditclone").Collection("posts")
	return collection, nil
}
