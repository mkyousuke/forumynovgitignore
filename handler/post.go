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
		if err := r.ParseMultipartForm(10 << 20); err != nil {
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
		if file, fileHeader, err := r.FormFile("image"); err == nil {
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
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Erreur lors de l'enregistrement de l'image", http.StatusInternalServerError)
				return
			}
		}
		if err := database.CreatePost(userID, title, content, imagePath); err != nil {
			http.Error(w, fmt.Sprintf("Erreur lors de la création du post: %v", err), http.StatusInternalServerError)
			return
		}

		// Notifications
		if user, errUser := database.GetUserWithRole(userID); errUser == nil {
			if lastPost, errLast := database.GetLastPostForUser(userID); errLast == nil {
				if user.Role == "admin" || user.Role == "moderator" {
					msg := fmt.Sprintf("Votre post \"%s\" a bien été publié.", title)
					_ = database.CreateNotification(userID, msg, lastPost.ID, 0)
				} else {
					msg := fmt.Sprintf("Votre post \"%s\" a été soumis à vérification.", title)
					_ = database.CreateNotification(userID, msg, lastPost.ID, 0)
					if mods, errMods := database.GetModeratorsAndAdmins(); errMods == nil {
						for _, mod := range mods {
							msgMod := fmt.Sprintf("Nouveau post \"%s\" en attente de vérification.", title)
							_ = database.CreateNotification(mod.ID, msgMod, lastPost.ID, 0)
						}
					}
				}
			}
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
	}{Posts: posts}
	t, err := template.ParseFiles(filepath.Join("templates", "posts.html"))
	if err != nil {
		http.Error(w, "Erreur interne du serveur (template): "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'affichage des posts: "+err.Error(), http.StatusInternalServerError)
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
	userPhoto := "profil.png"
	if cookie, err := r.Cookie("user_id"); err == nil {
		if uid, errConv := strconv.Atoi(cookie.Value); errConv == nil {
			if uwr, errUwr := database.GetUserWithRole(uid); errUwr == nil {
				if uid == post.UserID || uwr.Role == "admin" || uwr.Role == "moderator" {
					editable = true
				}
				if uwr.Photo != "" {
					userPhoto = uwr.Photo
				}
			} else if u, errU := database.GetUserByID(uid); errU == nil && u.Photo != "" {
				userPhoto = u.Photo
			}
		}
	}

	comments, err := database.GetCommentsByPostID(post.ID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des commentaires: "+err.Error(), http.StatusInternalServerError)
		return
	}

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
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'affichage du post: "+err.Error(), http.StatusInternalServerError)
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

	user, err := database.GetUserWithRole(userID)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}

	if user.Role == "admin" || user.Role == "moderator" {
		err = database.AdminDeletePost(postID)
	} else {
		err = database.DeletePost(postID, userID)
	}

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
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Erreur lors du traitement du formulaire", http.StatusBadRequest)
			return
		}
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
		if file, fileHeader, err := r.FormFile("image"); err == nil {
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
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Erreur lors de l'enregistrement de l'image", http.StatusInternalServerError)
				return
			}
		} else {
			if existingPost, err := database.GetPostByID(postID); err == nil {
				imagePath = existingPost.ImagePath
			}
		}
		if err := database.UpdatePost(postID, userID, title, content, imagePath); err != nil {
			http.Error(w, "Erreur lors de la mise à jour du post: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
	}
}
