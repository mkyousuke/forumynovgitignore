package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"forum/database"
)

type AdminUsersData struct {
	Users []database.User
	Admin database.User
}

// AdminUsersHandler affiche la liste de tous les utilisateurs avec
// des boutons pour promouvoir/démouvoir.
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	adminID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	admin, err := database.GetUserWithRole(adminID)
	if err != nil || admin.Role != "admin" {
		http.Error(w, "Accès réservé aux administrateurs", http.StatusForbidden)
		return
	}

	users, err := getAllUsers()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("templates/admin_users.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}
	data := AdminUsersData{Users: users, Admin: admin}
	if err := t.Execute(w, data); err != nil {
		fmt.Println("Erreur template admin_users:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// AdminUsersUpdateHandler traite la promotion ou la rétrogradation.
func AdminUsersUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	adminID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	admin, err := database.GetUserWithRole(adminID)
	if err != nil || admin.Role != "admin" {
		http.Error(w, "Accès réservé aux administrateurs", http.StatusForbidden)
		return
	}

	targetIDStr := r.FormValue("user_id")
	action := r.FormValue("action") // "promote" ou "demote"
	if targetIDStr == "" || (action != "promote" && action != "demote") {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	var newRole string
	if action == "promote" {
		newRole = "moderator"
	} else {
		newRole = "user"
	}
	if err := database.UpdateUserRole(targetID, newRole); err != nil {
		http.Error(w, "Erreur lors de la mise à jour du rôle", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// getAllUsers récupère tous les utilisateurs en base.
func getAllUsers() ([]database.User, error) {
	rows, err := database.DB.Query(
		"SELECT id, username, email, password, created_at, photo, role FROM users",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []database.User
	for rows.Next() {
		var u database.User
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Password,
			&u.CreatedAt,
			&u.Photo,
			&u.Role,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
