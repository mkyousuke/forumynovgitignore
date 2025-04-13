package handler

import (
	"net/http"
)

func DeconnexionHandler(w http.ResponseWriter, r *http.Request) {
	// Supprimer le cookie "user_id" en le configurant avec MaxAge négatif
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
