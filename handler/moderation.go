package handler

import (
	"forum/database"
	"html/template"
	"net/http"
	"strconv"
)

// ModerationDashboardHandler affiche la liste des posts en attente de modération.
func ModerationDashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	user, err := database.GetUserWithRole(userID)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}
	if user.Role != "moderator" && user.Role != "admin" {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}
	pendingPosts, err := database.GetPendingPosts()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts en attente", http.StatusInternalServerError)
		return
	}
	t, err := template.ParseFiles("templates/moderation.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}
	data := struct {
		PendingPosts []database.Post
		User         database.User
	}{
		PendingPosts: pendingPosts,
		User:         user,
	}
	t.Execute(w, data)
}

// ApprovePostHandler permet à un modérateur d'approuver un post.
func ApprovePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	user, err := database.GetUserWithRole(userID)
	if err != nil || (user.Role != "moderator" && user.Role != "admin") {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}
	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	err = database.SetPostModerationStatus(postID, "approved")
	if err != nil {
		http.Error(w, "Erreur lors de l'approbation du post", http.StatusInternalServerError)
		return
	}
	// Envoi de la notification à l'auteur
	post, err := database.GetPostByID(postID)
	if err == nil {
		msg := "Votre post \"" + post.Title + "\" a été approuvé."
		_ = database.CreateNotification(post.UserID, msg, post.ID, 0)
	}
	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

// RejectPostHandler permet à un modérateur de rejeter un post.
func RejectPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	user, err := database.GetUserWithRole(userID)
	if err != nil || (user.Role != "moderator" && user.Role != "admin") {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}
	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	err = database.SetPostModerationStatus(postID, "rejected")
	if err != nil {
		http.Error(w, "Erreur lors du rejet du post", http.StatusInternalServerError)
		return
	}
	// Envoi de la notification à l'auteur pour indiquer que son post a été rejeté.
	post, err := database.GetPostByID(postID)
	if err == nil {
		msg := "Votre post \"" + post.Title + "\" a été rejeté."
		_ = database.CreateNotification(post.UserID, msg, post.ID, 0)
	}
	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

// PromoteUserHandler permet à un administrateur de promouvoir un utilisateur en modérateur.
func PromoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	adminID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	admin, err := database.GetUserWithRole(adminID)
	if err != nil || admin.Role != "admin" {
		http.Error(w, "Accès réservé aux administrateurs", http.StatusForbidden)
		return
	}
	targetUserIDStr := r.FormValue("user_id")
	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		http.Error(w, "ID d'utilisateur invalide", http.StatusBadRequest)
		return
	}
	err = database.UpdateUserRole(targetUserID, "moderator")
	if err != nil {
		http.Error(w, "Erreur lors de la promotion de l'utilisateur", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// DemoteUserHandler permet à un administrateur de rétrograder un modérateur vers un utilisateur classique.
func DemoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	adminID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	admin, err := database.GetUserWithRole(adminID)
	if err != nil || admin.Role != "admin" {
		http.Error(w, "Accès réservé aux administrateurs", http.StatusForbidden)
		return
	}
	targetUserIDStr := r.FormValue("user_id")
	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		http.Error(w, "ID d'utilisateur invalide", http.StatusBadRequest)
		return
	}
	err = database.UpdateUserRole(targetUserID, "user")
	if err != nil {
		http.Error(w, "Erreur lors de la rétrogradation de l'utilisateur", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
