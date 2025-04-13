package server

import (
	"fmt"
	"log"
	"net/http"

	"forum/database"
	"forum/handler"

	"golang.org/x/crypto/bcrypt"
)

// seedDefaultUsers crée ou met à jour les comptes admin et modérateur
// pour qu'ils aient toujours le bon rôle.
func seedDefaultUsers() {
	// Définitions des comptes par défaut
	defaults := []struct {
		username string
		email    string
		pwd      string
		role     string
	}{
		{"admin", "admin@example.com", "admin", "admin"},
		{"moderateur", "mod@example.com", "moderateur", "moderator"},
	}

	for _, u := range defaults {
		user, err := database.GetUserByUsername(u.username)
		if err != nil {
			// Utilisateur introuvable → on le crée
			hash, _ := bcrypt.GenerateFromPassword([]byte(u.pwd), bcrypt.DefaultCost)
			_ = database.CreateUser(u.username, u.email, string(hash))
			user, _ = database.GetUserByUsername(u.username)
			fmt.Printf("⚙️  Utilisateur %q créé\n", u.username)
		}
		// On met (ou remet) toujours le rôle correct
		if err := database.UpdateUserRole(user.ID, u.role); err != nil {
			fmt.Printf("❌ Impossible de définir le rôle de %q: %v\n", u.username, err)
		} else {
			fmt.Printf("✅ Rôle de %q défini sur %q\n", u.username, u.role)
		}
	}
}

func StartServer() {
	// 1. Initialisation de la DB
	if err := database.InitDB("./forum.db"); err != nil {
		log.Fatal(err)
	}
	defer database.CloseDB()

	// 2. Création / mise à jour des comptes admin et moderateur
	seedDefaultUsers()

	// 3. Fichiers statiques
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 4. Routes publiques
	http.HandleFunc("/", handler.RedirectToIndex)
	http.HandleFunc("/index", handler.IndexHandler)
	http.HandleFunc("/inscription", handler.InscriptionHandler)
	http.HandleFunc("/connexion", handler.ConnexionHandler)
	http.HandleFunc("/deconnexion", handler.DeconnexionHandler)

	// 5. Profil
	http.HandleFunc("/profil", handler.ProfilHandler)
	http.HandleFunc("/modify-profil", handler.ModifyProfileHandler)

	// 6. TMDB & actualités
	http.HandleFunc("/api-tmdb", handler.TmdbHandler)
	http.HandleFunc("/actualites", handler.ActualitesHandler)
	http.HandleFunc("/theories-spoilers", handler.TheoriesSpoilersHandler)
	http.HandleFunc("/discussions", handler.DiscussionsHandler)

	// 7. Posts & commentaires
	http.HandleFunc("/nouveau-post", handler.NewPostHandler)
	http.HandleFunc("/posts", handler.PostsHandler)
	http.HandleFunc("/post", handler.PostDetailHandler)
	http.HandleFunc("/delete-post", handler.DeletePostHandler)
	http.HandleFunc("/edit-post", handler.EditPostHandler)
	http.HandleFunc("/add-comment", handler.AddCommentHandler)
	http.HandleFunc("/delete-comment", handler.DeleteCommentHandler)

	// 8. Notifications
	http.HandleFunc("/notifications", handler.NotificationsHandler)
	http.HandleFunc("/notifications-page", handler.NotificationsPageHandler)
	http.HandleFunc("/notifications/mark-read", handler.MarkNotificationsAsReadHandler)

	// 9. Like / dislike
	http.HandleFunc("/like-post", handler.LikePostHandler)
	http.HandleFunc("/dislike-post", handler.DislikePostHandler)
	http.HandleFunc("/like-comment", handler.LikeCommentHandler)
	http.HandleFunc("/dislike-comment", handler.DislikeCommentHandler)

	// 10. OAuth
	http.HandleFunc("/auth/google", handler.GoogleAuthHandler)
	http.HandleFunc("/auth/google/callback", handler.GoogleCallbackHandler)
	http.HandleFunc("/auth/facebook", handler.FacebookAuthHandler)
	http.HandleFunc("/auth/facebook/callback", handler.FacebookCallbackHandler)
	http.HandleFunc("/auth/github", handler.GithubAuthHandler)
	http.HandleFunc("/auth/github/callback", handler.GithubCallbackHandler)
	http.HandleFunc("/auth/twitter", handler.TwitterAuthHandler)
	http.HandleFunc("/auth/twitter/callback", handler.TwitterCallbackHandler)

	// 11. Modération & administration
	http.HandleFunc("/moderation", handler.ModerationDashboardHandler)
	http.HandleFunc("/moderation/approve", handler.ApprovePostHandler)
	http.HandleFunc("/moderation/reject", handler.RejectPostHandler)
	http.HandleFunc("/admin/promote", handler.PromoteUserHandler)
	http.HandleFunc("/admin/demote", handler.DemoteUserHandler)

	fmt.Println("✅ Serveur actif sur http://localhost:2020")
	log.Fatal(http.ListenAndServe(":2020", nil))
}
