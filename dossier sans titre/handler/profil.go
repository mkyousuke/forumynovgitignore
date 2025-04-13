package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"forum/database"
)

type ProfileData struct {
	database.User
	PostsLiked         int
	CommentsCount      int
	LastPostDate       string
	LastActivityDate   string
	LastConnectionDate string
}

func ProfilHandler(w http.ResponseWriter, r *http.Request) {
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
	user, err := getUserByID(userID)
	if err != nil {
		fmt.Println("Erreur dans getUserByID:", err)
		http.Error(w, "Erreur lors de la récupération du profil", http.StatusInternalServerError)
		return
	}
	postsLiked, commentsCount, err := database.GetUserStats(userID)
	if err != nil {
		postsLiked = 0
		commentsCount = 0
	}
	lastPostTime, err := database.GetLastPostDate(userID)
	lastPostStr := "Aucun post"
	if err == nil && !lastPostTime.IsZero() {
		lastPostStr = lastPostTime.Format("02/01/2006 15:04:05")
	}
	lastActivityTime, err := database.GetLastActivityDate(userID)
	lastActivityStr := "Aucune activité"
	if err == nil && !lastActivityTime.IsZero() {
		lastActivityStr = lastActivityTime.Format("02/01/2006 15:04:05")
	}
	lastConnectionTime, err := database.GetLastConnection(userID)
	lastConnectionStr := "Inconnue"
	if err == nil && !lastConnectionTime.IsZero() {
		lastConnectionStr = lastConnectionTime.Format("02/01/2006 15:04:05")
	}
	data := ProfileData{
		User:               user,
		PostsLiked:         postsLiked,
		CommentsCount:      commentsCount,
		LastPostDate:       lastPostStr,
		LastActivityDate:   lastActivityStr,
		LastConnectionDate: lastConnectionStr,
	}
	t, err := template.ParseFiles("templates/profil.html")
	if err != nil {
		fmt.Println("Erreur lors du parsing du template:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		fmt.Println("Erreur lors de l'exécution du template:", err)
		http.Error(w, "Erreur lors de l'affichage du profil", http.StatusInternalServerError)
		return
	}
}

func getUserByID(id int) (database.User, error) {
	var user database.User
	query := "SELECT id, username, email, created_at, photo FROM users WHERE id = ?"
	row := database.DB.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.Photo)
	if err != nil {
		fmt.Println("Erreur dans row.Scan de getUserByID:", err)
		return user, err
	}
	if user.Photo == "" {
		user.Photo = "profil.png"
	}
	return user, nil
}

func ModifyProfileHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		user, err := getUserByID(userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération du profil", http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("templates/modify_profil.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
			return
		}
		t.Execute(w, user)
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Erreur lors de la soumission du formulaire", http.StatusBadRequest)
			return
		}
		newUsername := r.FormValue("username")
		newPhoto := r.FormValue("photo")
		removePhoto := r.FormValue("remove_photo")

		if removePhoto == "true" {
			newPhoto = "profil.png"
		}
		if newUsername == "" {
			http.Error(w, "Le nom d'utilisateur est requis", http.StatusBadRequest)
			return
		}
		updateQuery := "UPDATE users SET username = ?, photo = ? WHERE id = ?"
		_, err = database.DB.Exec(updateQuery, newUsername, newPhoto, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour du profil", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profil", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
	}
}