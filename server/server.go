
package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"forum/database"
	"forum/handler"
	"forum/middleware"

	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/twitter"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/bcrypt"
)

func seedDefaultUsers() {
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
			hash, _ := bcrypt.GenerateFromPassword([]byte(u.pwd), bcrypt.DefaultCost)
			_ = database.CreateUser(u.username, u.email, string(hash))
			user, _ = database.GetUserByUsername(u.username)
			fmt.Printf("⚙️  Utilisateur %q créé\n", u.username)
		}
		if err := database.UpdateUserRole(user.ID, u.role); err != nil {
			fmt.Printf("❌ Impossible de définir le rôle de %q: %v\n", u.username, err)
		} else {
			fmt.Printf("✅ Rôle de %q défini sur %q\n", u.username, u.role)
		}
	}
}

// StartServer initialise la config et démarre le serveur HTTP/HTTPS.
func StartServer() {
	// Charger .env si présent
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Pas de fichier .env trouvé, on compte sur les vars d'environnement")
	}

	// ----- Configuration OAuth -----
	domain := os.Getenv("DOMAIN")
	scheme := "https"
	if os.Getenv("CERT_FILE") == "" && domain == "" {
		scheme = "http"
		domain = "localhost:2020"
	} else if domain == "" {
		domain = "localhost"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, domain)

	// Clés OAuth
	googleKey := os.Getenv("GOOGLE_KEY")
	googleSecret := os.Getenv("GOOGLE_SECRET")
	facebookKey := os.Getenv("FACEBOOK_KEY")
	facebookSecret := os.Getenv("FACEBOOK_SECRET")
	githubKey := os.Getenv("GITHUB_KEY")
	githubSecret := os.Getenv("GITHUB_SECRET")
	twitterKey := os.Getenv("TWITTER_KEY")
	twitterSecret := os.Getenv("TWITTER_SECRET")

	goth.UseProviders(
		google.New(googleKey, googleSecret, baseURL+"/auth/google/callback", "email", "profile"),
		facebook.New(facebookKey, facebookSecret, baseURL+"/auth/facebook/callback", "email", "public_profile"),
		github.New(githubKey, githubSecret, baseURL+"/auth/github/callback", "user", "user:email"),
		twitter.New(twitterKey, twitterSecret, baseURL+"/auth/twitter/callback"),
	)
	// --------------------------------

	// Init DB
	if err := database.InitDB("./forum.db"); err != nil {
		log.Fatal(err)
	}
	defer database.CloseDB()

	// Seed users
	seedDefaultUsers()

	// Création du mux
	mux := http.NewServeMux()

	// Statiques
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes...
	mux.HandleFunc("/", handler.RedirectToIndex)
	mux.HandleFunc("/index", handler.IndexHandler)
	mux.HandleFunc("/inscription", handler.InscriptionHandler)
	mux.HandleFunc("/connexion", handler.ConnexionHandler)
	mux.HandleFunc("/deconnexion", handler.DeconnexionHandler)
	mux.HandleFunc("/profil", handler.ProfilHandler)
	mux.HandleFunc("/modify-profil", handler.ModifyProfileHandler)
	mux.HandleFunc("/api-tmdb", handler.TmdbHandler)
	mux.HandleFunc("/actualites", handler.ActualitesHandler)
	mux.HandleFunc("/theories-spoilers", handler.TheoriesSpoilersHandler)
	mux.HandleFunc("/nouveau-post", handler.NewPostHandler)
	mux.HandleFunc("/posts", handler.PostsHandler)
	mux.HandleFunc("/post", handler.PostDetailHandler)
	mux.HandleFunc("/delete-post", handler.DeletePostHandler)
	mux.HandleFunc("/edit-post", handler.EditPostHandler)
	mux.HandleFunc("/add-comment", handler.AddCommentHandler)
	mux.HandleFunc("/delete-comment", handler.DeleteCommentHandler)
	mux.HandleFunc("/notifications", handler.NotificationsHandler)
	mux.HandleFunc("/notifications-page", handler.NotificationsPageHandler)
	mux.HandleFunc("/notifications/mark-read", handler.MarkNotificationsAsReadHandler)
	mux.HandleFunc("/like-post", handler.LikePostHandler)
	mux.HandleFunc("/dislike-post", handler.DislikePostHandler)
	mux.HandleFunc("/like-comment", handler.LikeCommentHandler)
	mux.HandleFunc("/dislike-comment", handler.DislikeCommentHandler)
	mux.HandleFunc("/auth/google", handler.GoogleAuthHandler)
	mux.HandleFunc("/auth/google/callback", handler.GoogleCallbackHandler)
	mux.HandleFunc("/auth/facebook", handler.FacebookAuthHandler)
	mux.HandleFunc("/auth/facebook/callback", handler.FacebookCallbackHandler)
	mux.HandleFunc("/auth/github", handler.GithubAuthHandler)
	mux.HandleFunc("/auth/github/callback", handler.GithubCallbackHandler)
	mux.HandleFunc("/auth/twitter", handler.TwitterAuthHandler)
	mux.HandleFunc("/auth/twitter/callback", handler.TwitterCallbackHandler)
	mux.HandleFunc("/moderation", handler.ModerationDashboardHandler)
	mux.HandleFunc("/moderation/approve", handler.ApprovePostHandler)
	mux.HandleFunc("/moderation/reject", handler.RejectPostHandler)
	mux.HandleFunc("/admin/promote", handler.PromoteUserHandler)
	mux.HandleFunc("/admin/demote", handler.DemoteUserHandler)
	mux.HandleFunc("/admin/users", handler.AdminUsersHandler)
	mux.HandleFunc("/admin/users/update", handler.AdminUsersUpdateHandler)
	mux.HandleFunc("/report-post", handler.ReportPostHandler)
	mux.HandleFunc("/admin/reports", handler.AdminReportsHandler)
	mux.HandleFunc("/admin/reports/respond", handler.RespondReportHandler)

	// Gemini Chat Routes
	mux.HandleFunc("/gemini-chat", handler.GeminiChatPage)
	mux.HandleFunc("/api/gemini-chat", handler.GeminiChatAPI)

	// Rate Limiter
	handlerWithRate := middleware.RateLimit(middleware.GzipAndCacheMiddleware(mux))

	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	// HTTPS local (PEM)
	if certFile != "" && keyFile != "" {
		url := "https://localhost"
		fmt.Printf("✅ Serveur HTTPS démarré. Ouvrez ce lien dans votre navigateur :\n")
		fmt.Printf("\x1b]8;;%s\x07%s\x1b]8;;\x07\n", url, url)
		log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, handlerWithRate))
	}

	// HTTP fallback
	if domain == "localhost:2020" {
		url := "http://localhost:2020"
		fmt.Println("⚠️  Pas de DOMAIN défini, démarrage sur HTTP :2020")
		fmt.Printf("✅ Accédez à votre site :\n")
		fmt.Printf("\x1b]8;;%s\x07%s\x1b]8;;\x07\n", url, url)
		log.Fatal(http.ListenAndServe(":2020", handlerWithRate))
	}

	// Let's Encrypt prod
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("cert-cache"),
	}
	tlsConfig := &tls.Config{
		GetCertificate: m.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	}
	server := &http.Server{
		Addr:      ":443",
		Handler:   handlerWithRate,
		TLSConfig: tlsConfig,
	}
	go func() {
		httpSrv := &http.Server{
			Addr:    ":80",
			Handler: m.HTTPHandler(mux),
		}
		log.Fatal(httpSrv.ListenAndServe())
	}()

	// HTTPS prod
	fmt.Printf("✅ Serveur HTTPS démarré sur %s\n", baseURL)
	fmt.Printf("✅ Accédez à votre site :\n")
	fmt.Printf("\x1b]8;;%s\x07%s\x1b]8;;\x07\n", baseURL, baseURL)
	log.Fatal(server.ListenAndServeTLS("", ""))
}

