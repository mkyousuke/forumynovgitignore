package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"forum/database"

	"golang.org/x/crypto/bcrypt"
)

func ConnexionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, err := template.ParseFiles("templates/connexion.html")
		if err != nil {
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage du template", http.StatusInternalServerError)
		}
	case http.MethodPost:
		identifier := r.FormValue("identifier")
		password := r.FormValue("password")
		if identifier == "" || password == "" {
			http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
			return
		}
		var user database.User
		var err error
		// Si l'identifiant contient un "@" on considère que c'est un email, sinon un nom d'utilisateur
		if strings.Contains(identifier, "@") {
			user, err = database.GetUserByEmail(identifier)
		} else {
			user, err = database.GetUserByUsername(identifier)
		}
		if err != nil {
			http.Error(w, "Utilisateur introuvable", http.StatusUnauthorized)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
			return
		}
		cookie := http.Cookie{
			Name:     "user_id",
			Value:    strconv.Itoa(user.ID),
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/index", http.StatusSeeOther)
	default:
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
	}
}
