package handler

import (
	"net/http"

	"forum/database"
)

func DeconnexionHandler(w http.ResponseWriter, r *http.Request) {
	// Supprimer la session côté serveur
	if cookie, err := r.Cookie("session_id"); err == nil {
		_ = database.DeleteSession(cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}
	// Supprimer cookie user_id
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
