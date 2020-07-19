package main

import (
	"html/template"
	"log"
	"net/http"
	"redditclone/pkg/database"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	time.Sleep(30 * time.Second)

	db, err := database.GetMysql()
	if err != nil {
		log.Fatal(err.Error())
	}
	collection, err := database.GetCollection()
	if err != nil {
		log.Fatal(err.Error())
	}

	/*---------------------------------*/

	templates := template.Must(template.ParseGlob("./template/index.html"))

	sm := session.NewSessionsMem(db)
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	userRepo := user.NewUserRepo(db)
	postsRepo := posts.NewRepo(collection)

	userHandler := &handlers.UserHandler{
		Tmpl:     templates,
		UserRepo: userRepo,
		Sessions: sm,
	}

	handlers := &handlers.PostsHandler{
		Tmpl:      templates,
		PostsRepo: postsRepo,
	}
	appMux := mux.NewRouter()

	appMux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./template/static"))))

	appMux.HandleFunc("/", userHandler.Index).Methods("GET")
	appMux.HandleFunc("/api/register", userHandler.Register).Methods("POST")
	appMux.HandleFunc("/api/login", userHandler.Login).Methods("POST")

	appMux.HandleFunc("/api/posts/", handlers.GetAll).Methods("GET")
	appMux.HandleFunc("/api/posts", handlers.NewPost).Methods("POST")

	appMux.HandleFunc("/api/posts/{CATEGORY_NAME}", handlers.Category).Methods("GET")
	appMux.HandleFunc("/api/post/{POST_ID}", handlers.GetByID).Methods("GET")
	appMux.HandleFunc("/api/post/{POST_ID}", handlers.Comment).Methods("POST")
	appMux.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", handlers.DelComment).Methods("DELETE")

	appMux.HandleFunc("/api/post/{POST_ID}/upvote", handlers.Upvote).Methods("GET")
	appMux.HandleFunc("/api/post/{POST_ID}/downvote", handlers.Downvote).Methods("GET")
	appMux.HandleFunc("/api/post/{POST_ID}/unvote", handlers.Unvote).Methods("GET")

	appMux.HandleFunc("/api/post/{POST_ID}", handlers.DelPost).Methods("DELETE")
	appMux.HandleFunc("/api/user/{USER_LOGIN}", handlers.GetByUser).Methods("GET")

	mux := middleware.Auth(sm, appMux)
	mux = middleware.AccessLog(logger, mux)
	mux = middleware.Panic(mux)

	addr := "8080"
	log.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(":"+addr, mux))
}
