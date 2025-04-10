package handler

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"forum/database"
)

// Fonction existante pour rendre un template
func renderTemplate(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	templatePath := filepath.Join("templates", templateName)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		http.Error(w, "Template "+templateName+" introuvable", http.StatusNotFound)
		return
	}
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution du template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func CritiquesAvisHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "critiques-avis.html", nil)
}

func DiscussionsHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "discussions.html", nil)
}

// --- Modification de IndexHandler ---
// Ajout de la gestion du cookie user_id pour transmettre
// les données IsLoggedIn et UserPhoto au template index.html.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	type IndexData struct {
		IsLoggedIn bool
		UserPhoto  string
	}
	// Par défaut, on considère que l'utilisateur n'est pas connecté
	data := IndexData{
		IsLoggedIn: false,
		UserPhoto:  "profil.png",
	}
	cookie, err := r.Cookie("user_id")
	if err == nil && cookie.Value != "" {
		userID, err := strconv.Atoi(cookie.Value)
		if err == nil {
			user, err := database.GetUserByID(userID)
			if err == nil {
				data.IsLoggedIn = true
				if user.Photo != "" {
					data.UserPhoto = user.Photo
				}
			}
		}
	}
	renderTemplate(w, r, "index.html", data)
}

func TheoriesSpoilersHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "theoriesSpoilers.html", nil)
}

func AnimationsHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "animations.html", nil)
}

func CritiquesAnalysesHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "critiquesAnalyses.html", nil)
}

func FilmsHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "films.html", nil)
}

func SeriesHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "series.html", nil)
}

func RedirectToIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}