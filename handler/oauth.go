package handler

import (
	"net/http"
	"strconv"

	"forum/database"
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
		err = database.CreateUser(user.Name, user.Email, "oauth")
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
		dbUser, err = database.GetUserByEmail(user.Email)
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
	}
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/profil", http.StatusTemporaryRedirect)
}

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
		err = database.CreateUser(user.Name, user.Email, "oauth")
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
		dbUser, err = database.GetUserByEmail(user.Email)
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
	}
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
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
		err = database.CreateUser(user.Name, user.Email, "oauth")
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
		dbUser, err = database.GetUserByEmail(user.Email)
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
	}
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/profil", http.StatusTemporaryRedirect)
}

func TwitterAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Pour Twitter, il faut préciser le provider dans la query
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
		// Si l'utilisateur n'existe pas, on le crée
		err = database.CreateUser(user.Name, user.Email, "oauth")
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
		dbUser, err = database.GetUserByEmail(user.Email)
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusTemporaryRedirect)
			return
		}
	}
	
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(dbUser.ID),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/profil", http.StatusTemporaryRedirect)
}
