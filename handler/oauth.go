package handler

import (
	"net/http"
	"strconv"
	"time"

	"forum/database"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
)

func GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=google"
	gothic.BeginAuthHandler(w, r)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
		return
	}
	dbUser, err := database.GetUserByEmail(user.Email)
	if err != nil {
		_ = database.CreateUser(user.Name, user.Email, "oauth")
		dbUser, _ = database.GetUserByEmail(user.Email)
	}
	// Création session
	sessionID := uuid.NewString()
	expiry := time.Now().Add(24 * time.Hour)
	_ = database.CreateSession(sessionID, dbUser.ID, expiry)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiry,
	})
	// Compatibilité user_id
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/profil", http.StatusTemporaryRedirect)
}

// Même logique pour Facebook, Github et Twitter :

func FacebookAuthHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=facebook"
	gothic.BeginAuthHandler(w, r)
}

func FacebookCallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
		return
	}
	dbUser, err := database.GetUserByEmail(user.Email)
	if err != nil {
		_ = database.CreateUser(user.Name, user.Email, "oauth")
		dbUser, _ = database.GetUserByEmail(user.Email)
	}
	sessionID := uuid.NewString()
	expiry := time.Now().Add(24 * time.Hour)
	_ = database.CreateSession(sessionID, dbUser.ID, expiry)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiry,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/profil", http.StatusTemporaryRedirect)
}

func GithubAuthHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=github"
	gothic.BeginAuthHandler(w, r)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
		return
	}
	dbUser, err := database.GetUserByEmail(user.Email)
	if err != nil {
		_ = database.CreateUser(user.Name, user.Email, "oauth")
		dbUser, _ = database.GetUserByEmail(user.Email)
	}
	sessionID := uuid.NewString()
	expiry := time.Now().Add(24 * time.Hour)
	_ = database.CreateSession(sessionID, dbUser.ID, expiry)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiry,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/profil", http.StatusTemporaryRedirect)
}

func TwitterAuthHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=twitter"
	gothic.BeginAuthHandler(w, r)
}

func TwitterCallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
		return
	}
	dbUser, err := database.GetUserByEmail(user.Email)
	if err != nil {
		_ = database.CreateUser(user.Name, user.Email, "oauth")
		dbUser, _ = database.GetUserByEmail(user.Email)
	}
	sessionID := uuid.NewString()
	expiry := time.Now().Add(24 * time.Hour)
	_ = database.CreateSession(sessionID, dbUser.ID, expiry)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiry,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/profil", http.StatusTemporaryRedirect)
}
