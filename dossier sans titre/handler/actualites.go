package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Structures pour parser la réponse de NewsAPI
type NewsAPIResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type Article struct {
	Source struct {
		Name string `json:"name"`
	} `json:"source"`
	Author      string `json:"author"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	URLToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
	Content     string `json:"content"`
}

// ActualitesHandler interroge NewsAPI et affiche le template actualites.html
func ActualitesHandler(w http.ResponseWriter, r *http.Request) {
	// Récupération de la clé API depuis les variables d'environnement
	apiKey := os.Getenv("NEWSAPI_KEY")
	if apiKey == "" {
		http.Error(w, "NEWSAPI_KEY non défini", http.StatusInternalServerError)
		return
	}
	log.Println("Clé NewsAPI =", apiKey)

	// Construction de la requête à NewsAPI avec filtrage sur les mots-clés et en français
	baseURL := "https://newsapi.org/v2/everything"
	// La requête recherche des actualités contenant "netflix", "séries", "cinéma" ou "films"
	q := "netflix OR séries OR cinéma OR films"
	params := url.Values{}
	params.Add("q", q)
	params.Add("language", "fr")
	params.Add("apiKey", apiKey)

	apiURL := baseURL + "?" + params.Encode()
	log.Println("[NewsAPI] Requête :", apiURL)

	// Requête HTTP vers NewsAPI
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Erreur lors de la requête vers NewsAPI:", err)
		http.Error(w, "Erreur lors de la récupération des actualités", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var newsResp NewsAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&newsResp)
	if err != nil {
		log.Println("Erreur lors du décodage JSON:", err)
		http.Error(w, "Erreur lors du traitement des données", http.StatusInternalServerError)
		return
	}

	// Chargement et exécution du template
	tmpl, err := template.ParseFiles("templates/actualites.html")
	if err != nil {
		log.Println("Erreur lors du parsing du template:", err)
		http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, newsResp)
	if err != nil {
		log.Println("Erreur lors de l'exécution du template:", err)
		http.Error(w, "Erreur lors de l'affichage des actualités", http.StatusInternalServerError)
		return
	}
}
