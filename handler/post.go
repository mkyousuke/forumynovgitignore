package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"forum/database"
)

// Helper qui calcule la différence simple entre original et modifié (non utilisé dans la nouvelle version)
func diffContent(original, modified string) (string, string, string) {
	i := 0
	for i < len(original) && i < len(modified) && original[i] == modified[i] {
		i++
	}
	j := 0
	for j < len(original)-i && j < len(modified)-i && original[len(original)-1-j] == modified[len(modified)-1-j] {
		j++
	}
	prefix := modified[:i]
	diff := modified[i : len(modified)-j]
	suffix := modified[len(modified)-j:]
	return prefix, diff, suffix
}

func NewPostHandler(w http.ResponseWriter, r *http.Request) {
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
	switch r.Method {
	case http.MethodGet:
		t, err := template.ParseFiles(filepath.Join("templates", "new_post.html"))
		if err != nil {
			http.Error(w, "Erreur interne du serveur (template)", http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	case http.MethodPost:
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Erreur lors du traitement du formulaire", http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		content := r.FormValue("content")
		if title == "" || content == "" {
			http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
			return
		}
		var imagePath string
		file, fileHeader, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			uploadDir := filepath.Join("static", "uploads")
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				os.MkdirAll(uploadDir, os.ModePerm)
			}
			imagePath = filepath.Join("static", "uploads", fileHeader.Filename)
			dst, err := os.Create(imagePath)
			if err != nil {
				http.Error(w, "Erreur lors de l'enregistrement de l'image", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, "Erreur lors de l'enregistrement de l'image", http.StatusInternalServerError)
				return
			}
		}
		err = database.CreatePost(userID, title, content, imagePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erreur lors de la création du post: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	default:
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
	}
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := database.GetAllPosts()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Posts []database.Post
	}{
		Posts: posts,
	}
	t, err := template.ParseFiles(filepath.Join("templates", "posts.html"))
	if err != nil {
		http.Error(w, "Erreur interne du serveur (template): "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Erreur lors de l'affichage des posts: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func PostDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	post, err := database.GetPostByID(id)
	if err != nil {
		http.Error(w, "Post introuvable: "+err.Error(), http.StatusNotFound)
		return
	}
	var editable bool
	var userPhoto string = "profil.png"
	cookie, err := r.Cookie("user_id")
	if err == nil {
		userID, errConv := strconv.Atoi(cookie.Value)
		if errConv == nil {
			if userID == post.UserID {
				editable = true
			}
			user, errUser := database.GetUserByID(userID)
			if errUser == nil && user.Photo != "" {
				userPhoto = user.Photo
			}
		}
	}
	comments, err := database.GetCommentsByPostID(post.ID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des commentaires: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Si le post a été modifié et que le contenu diffère de l'original,
	// nous affichons l'ancienne version et la version modifiée.
	modified := false
	if !post.ModifiedAt.IsZero() && post.OriginalContent != "" && post.OriginalContent != post.Content {
		modified = true
	}
	data := struct {
		Post      database.Post
		Editable  bool
		Comments  []database.Comment
		UserPhoto string
		Modified  bool
	}{
		Post:      post,
		Editable:  editable,
		Comments:  comments,
		UserPhoto: userPhoto,
		Modified:  modified,
	}
	t, err := template.ParseFiles(filepath.Join("templates", "post_detail.html"))
	if err != nil {
		http.Error(w, "Erreur interne du serveur (template): "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Erreur lors de l'affichage du post: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
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
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	err = database.DeletePost(postID, userID)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression du post: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
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
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet {
		post, err := database.GetPostByID(postID)
		if err != nil {
			http.Error(w, "Post introuvable: "+err.Error(), http.StatusNotFound)
			return
		}
		if post.UserID != userID {
			http.Error(w, "Non autorisé", http.StatusForbidden)
			return
		}
		t, err := template.ParseFiles(filepath.Join("templates", "edit_post.html"))
		if err != nil {
			http.Error(w, "Erreur interne du serveur (template)", http.StatusInternalServerError)
			return
		}
		t.Execute(w, post)
	} else if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Erreur lors du traitement du formulaire", http.StatusBadRequest)
			return
		}
		// Titre non modifiable : si vide, on récupère le titre existant
		title := r.FormValue("title")
		if title == "" {
			existingPost, err := database.GetPostByID(postID)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération du titre existant", http.StatusInternalServerError)
				return
			}
			title = existingPost.Title
		}
		content := r.FormValue("content")
		if content == "" {
			http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
			return
		}
		var imagePath string
		file, fileHeader, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			uploadDir := filepath.Join("static", "uploads")
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				os.MkdirAll(uploadDir, os.ModePerm)
			}
			imagePath = filepath.Join("static", "uploads", fileHeader.Filename)
			dst, err := os.Create(imagePath)
			if err != nil {
				http.Error(w, "Erreur lors de l'enregistrement de l'image", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, "Erreur lors de l'enregistrement de l'image", http.StatusInternalServerError)
				return
			}
		} else {
			existingPost, err := database.GetPostByID(postID)
			if err == nil {
				imagePath = existingPost.ImagePath
			}
		}
		// Mise à jour du post avec modified_at et conservation de l'original_content
		err = database.UpdatePost(postID, userID, title, content, imagePath)
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour du post: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
	}
}