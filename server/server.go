package server

import (
	"fmt"
	"log"
	"net/http"

	"forum/database"
	"forum/handler"
)

func StartServer() {
	err := database.InitDB("./forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDB()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handler.RedirectToIndex)
	http.HandleFunc("/index", handler.IndexHandler)
	http.HandleFunc("/inscription", handler.InscriptionHandler)
	http.HandleFunc("/connexion", handler.ConnexionHandler)
	http.HandleFunc("/profil", handler.ProfilHandler)
	http.HandleFunc("/modify-profil", handler.ModifyProfileHandler)
	http.HandleFunc("/api-tmdb", handler.TmdbHandler)
	http.HandleFunc("/deconnexion", handler.DeconnexionHandler)
	http.HandleFunc("/actualites", handler.ActualitesHandler)
	http.HandleFunc("/theories-spoilers", handler.TheoriesSpoilersHandler)
	http.HandleFunc("/discussions", handler.DiscussionsHandler)
	http.HandleFunc("/nouveau-post", handler.NewPostHandler)
	http.HandleFunc("/posts", handler.PostsHandler)
	http.HandleFunc("/post", handler.PostDetailHandler)
	http.HandleFunc("/delete-post", handler.DeletePostHandler)
	http.HandleFunc("/edit-post", handler.EditPostHandler)
	http.HandleFunc("/add-comment", handler.AddCommentHandler)
	http.HandleFunc("/delete-comment", handler.DeleteCommentHandler)
	http.HandleFunc("/notifications", handler.NotificationsHandler)
	http.HandleFunc("/notifications-page", handler.NotificationsPageHandler)
	http.HandleFunc("/notifications/mark-read", handler.MarkNotificationsAsReadHandler)

	// Routes pour le like/dislike
	http.HandleFunc("/like-post", handler.LikePostHandler)
	http.HandleFunc("/dislike-post", handler.DislikePostHandler)
	http.HandleFunc("/like-comment", handler.LikeCommentHandler)
	http.HandleFunc("/dislike-comment", handler.DislikeCommentHandler)

	http.HandleFunc("/auth/google", handler.GoogleAuthHandler)
	http.HandleFunc("/auth/google/callback", handler.GoogleCallbackHandler)
	http.HandleFunc("/auth/facebook", handler.FacebookAuthHandler)
	http.HandleFunc("/auth/facebook/callback", handler.FacebookCallbackHandler)
	http.HandleFunc("/auth/github", handler.GithubAuthHandler)
	http.HandleFunc("/auth/github/callback", handler.GithubCallbackHandler)
	http.HandleFunc("/auth/twitter", handler.TwitterAuthHandler)
	http.HandleFunc("/auth/twitter/callback", handler.TwitterCallbackHandler)

	fmt.Println("✅ Serveur actif sur http://localhost:2020")
	log.Fatal(http.ListenAndServe(":2020", nil))
}