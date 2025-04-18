// database/database.go
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initialise la connexion à la base de données et crée les tables.
func InitDB(dbFilePath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

// CloseDB ferme la connexion à la base.
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func createTables() error {
	sqlPath := filepath.Join("database", "SQL", "database.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}
	if _, err := DB.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute SQL commands: %w", err)
	}
	return nil
}

// User représente un utilisateur avec son rôle.
type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt string
	Photo     string
	Role      string // "user", "moderator" ou "admin"
}

// CreateUser insère un nouvel utilisateur avec rôle par défaut "user".
func CreateUser(username, email, password string) error {
	query := `INSERT INTO users (username, email, password, photo) VALUES (?, ?, ?, ?);`
	_, err := DB.Exec(query, username, email, password, "profil.png")
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByEmail récupère un utilisateur par email (sans le rôle).
func GetUserByEmail(email string) (User, error) {
	var user User
	query := `SELECT id, username, email, password, created_at, photo FROM users WHERE email = ?;`
	row := DB.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.Photo)
	if err != nil {
		return user, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// GetUserByUsername récupère un utilisateur par username (sans le rôle).
func GetUserByUsername(username string) (User, error) {
	var user User
	query := `SELECT id, username, email, password, created_at, photo FROM users WHERE username = ?;`
	row := DB.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.Photo)
	if err != nil {
		return user, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// GetUserWithRole récupère un utilisateur complet, y compris son rôle.
func GetUserWithRole(id int) (User, error) {
	var user User
	query := "SELECT id, username, email, password, created_at, photo, role FROM users WHERE id = ?;"
	row := DB.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.Photo, &user.Role)
	return user, err
}

// UpdateUserRole modifie le rôle d'un utilisateur.
func UpdateUserRole(userID int, role string) error {
	query := "UPDATE users SET role = ? WHERE id = ?;"
	_, err := DB.Exec(query, role, userID)
	return err
}

// GetModeratorsAndAdmins récupère tous les utilisateurs dont le rôle est "admin" ou "moderator".
func GetModeratorsAndAdmins() ([]User, error) {
	query := "SELECT id, username, email, password, created_at, photo, role FROM users WHERE role = 'admin' OR role = 'moderator';"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt, &u.Photo, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Post représente un post avec son statut de modération.
type Post struct {
	ID              int
	UserID          int
	Username        string
	Title           string
	Content         string
	OriginalContent string
	ImagePath       string
	CreatedAt       time.Time
	ModifiedAt      time.Time
	Likes           int
	Dislikes        int
}

// CreatePost insère un post selon le rôle de l'auteur.
func CreatePost(userID int, title, content, imagePath string) error {
	user, err := GetUserWithRole(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user role: %w", err)
	}
	status := "pending"
	if user.Role == "admin" || user.Role == "moderator" {
		status = "approved"
	}
	query := `INSERT INTO posts (user_id, title, content, original_content, image_path, moderation_status) VALUES (?, ?, ?, ?, ?, ?);`
	_, err = DB.Exec(query, userID, title, content, content, imagePath, status)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}

// GetAllPosts récupère tous les posts approuvés.
func GetAllPosts() ([]Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.original_content, p.image_path, p.created_at, p.modified_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.moderation_status = 'approved'
		ORDER BY p.created_at DESC;
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		var createdAtStr, modifiedAtStr sql.NullString
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.OriginalContent, &p.ImagePath, &createdAtStr, &modifiedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}
		parsed, perr := time.Parse(time.RFC3339, createdAtStr.String)
		if perr != nil {
			parsed, _ = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
		}
		p.CreatedAt = parsed.Add(2 * time.Hour)
		if modifiedAtStr.Valid && modifiedAtStr.String != "" {
			mparsed, merr := time.Parse(time.RFC3339, modifiedAtStr.String)
			if merr != nil {
				mparsed, _ = time.Parse("2006-01-02 15:04:05", modifiedAtStr.String)
			}
			p.ModifiedAt = mparsed.Add(2 * time.Hour)
		}
		p.Likes, _ = CountPostLikes(p.ID)
		p.Dislikes, _ = CountPostDislikes(p.ID)
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return posts, nil
}

// GetRecentPosts récupère les posts approuvés les plus récents.
func GetRecentPosts(limit int) ([]Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.original_content, p.image_path, p.created_at, p.modified_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.moderation_status = 'approved'
		ORDER BY p.created_at DESC
		LIMIT ?;
	`
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent posts: %w", err)
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		var createdAtStr, modifiedAtStr sql.NullString
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.OriginalContent, &p.ImagePath, &createdAtStr, &modifiedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan recent post row: %w", err)
		}
		parsed, perr := time.Parse(time.RFC3339, createdAtStr.String)
		if perr != nil {
			parsed, _ = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
		}
		p.CreatedAt = parsed.Add(2 * time.Hour)
		if modifiedAtStr.Valid && modifiedAtStr.String != "" {
			mparsed, merr := time.Parse(time.RFC3339, modifiedAtStr.String)
			if merr != nil {
				mparsed, _ = time.Parse("2006-01-02 15:04:05", modifiedAtStr.String)
			}
			p.ModifiedAt = mparsed.Add(2 * time.Hour)
		}
		p.Likes, _ = CountPostLikes(p.ID)
		p.Dislikes, _ = CountPostDislikes(p.ID)
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return posts, nil
}

// GetPostByID récupère un post approuvé par son ID.
func GetPostByID(id int) (Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.original_content, p.image_path, p.created_at, p.modified_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ? AND p.moderation_status = 'approved';
	`
	row := DB.QueryRow(query, id)
	var p Post
	var createdAtStr, modifiedAtStr sql.NullString
	if err := row.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.OriginalContent, &p.ImagePath, &createdAtStr, &modifiedAtStr); err != nil {
		return p, fmt.Errorf("failed to get post by ID: %w", err)
	}
	parsed, perr := time.Parse(time.RFC3339, createdAtStr.String)
	if perr != nil {
		parsed, _ = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
	}
	p.CreatedAt = parsed.Add(2 * time.Hour)
	if modifiedAtStr.Valid && modifiedAtStr.String != "" {
		mparsed, merr := time.Parse(time.RFC3339, modifiedAtStr.String)
		if merr != nil {
			mparsed, _ = time.Parse("2006-01-02 15:04:05", modifiedAtStr.String)
		}
		p.ModifiedAt = mparsed.Add(2 * time.Hour)
	}
	p.Likes, _ = CountPostLikes(p.ID)
	p.Dislikes, _ = CountPostDislikes(p.ID)
	return p, nil
}

// DeletePost supprime un post selon son ID et l'utilisateur propriétaire.
func DeletePost(postID int, userID int) error {
	query := "DELETE FROM posts WHERE id = ? AND user_id = ?;"
	_, err := DB.Exec(query, postID, userID)
	return err
}

// AdminDeletePost supprime un post sans vérifier l'appartenance (pour admin/modérateur).
func AdminDeletePost(postID int) error {
	query := "DELETE FROM posts WHERE id = ?;"
	_, err := DB.Exec(query, postID)
	return err
}

// UpdatePost met à jour un post existant.
func UpdatePost(postID int, userID int, title, content, imagePath string) error {
	var oldContent string
	var oldOriginal sql.NullString
	err := DB.QueryRow("SELECT content, original_content FROM posts WHERE id = ? AND user_id = ?;", postID, userID).Scan(&oldContent, &oldOriginal)
	if err != nil {
		return err
	}
	if !oldOriginal.Valid || oldOriginal.String == "" {
		oldOriginal.String = oldContent
	}
	query := "UPDATE posts SET title = ?, content = ?, image_path = ?, modified_at = CURRENT_TIMESTAMP, original_content = ? WHERE id = ? AND user_id = ?;"
	_, err = DB.Exec(query, title, content, imagePath, oldOriginal.String, postID, userID)
	return err
}

// GetLastPostForUser récupère le dernier post créé par un utilisateur.
func GetLastPostForUser(userID int) (Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.original_content, p.image_path, p.created_at, p.modified_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT 1;
	`
	var p Post
	var createdAtStr, modifiedAtStr sql.NullString
	row := DB.QueryRow(query, userID)
	err := row.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.OriginalContent, &p.ImagePath, &createdAtStr, &modifiedAtStr)
	if err != nil {
		return p, err
	}
	parsed, perr := time.Parse(time.RFC3339, createdAtStr.String)
	if perr != nil {
		parsed, _ = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
	}
	p.CreatedAt = parsed.Add(2 * time.Hour)
	if modifiedAtStr.Valid && modifiedAtStr.String != "" {
		mparsed, merr := time.Parse(time.RFC3339, modifiedAtStr.String)
		if merr != nil {
			mparsed, _ = time.Parse("2006-01-02 15:04:05", modifiedAtStr.String)
		}
		p.ModifiedAt = mparsed.Add(2 * time.Hour)
	}
	p.Likes, _ = CountPostLikes(p.ID)
	p.Dislikes, _ = CountPostDislikes(p.ID)
	return p, nil
}

// Comment représente un commentaire sur un post.
type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
	Photo     string // photo de l'utilisateur
}

// CreateComment insère un nouveau commentaire.
func CreateComment(postID int, userID int, content string) error {
	query := "INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?);"
	_, err := DB.Exec(query, postID, userID, content)
	return err
}

// DeleteComment supprime un commentaire selon son ID et l'utilisateur propriétaire.
func DeleteComment(commentID int, userID int) error {
	query := "DELETE FROM comments WHERE id = ? AND user_id = ?;"
	_, err := DB.Exec(query, commentID, userID)
	return err
}

// AdminDeleteComment supprime un commentaire sans vérifier l'appartenance.
func AdminDeleteComment(commentID int) error {
	query := "DELETE FROM comments WHERE id = ?;"
	_, err := DB.Exec(query, commentID)
	return err
}

// GetCommentsByPostID récupère les commentaires d'un post.
func GetCommentsByPostID(postID int) ([]Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, u.photo
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC;
	`
	rows, err := DB.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var c Comment
		var createdAtStr string
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &createdAtStr, &c.Photo); err != nil {
			return nil, err
		}
		parsed, perr := time.Parse(time.RFC3339, createdAtStr)
		if perr != nil {
			parsed, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		c.CreatedAt = parsed.Add(2 * time.Hour)
		c.Likes, _ = CountCommentLikes(c.ID)
		c.Dislikes, _ = CountCommentDislikes(c.ID)
		comments = append(comments, c)
	}
	return comments, nil
}

// GetUserStats renvoie le nombre de posts likés et de commentaires d'un utilisateur.
func GetUserStats(userID int) (int, int, error) {
	var postsLiked int
	var commentsCount int
	err := DB.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id IS NOT NULL AND value = 1;", userID).Scan(&postsLiked)
	if err != nil {
		return 0, 0, err
	}
	err = DB.QueryRow("SELECT COUNT(*) FROM comments WHERE user_id = ?;", userID).Scan(&commentsCount)
	if err != nil {
		return 0, 0, err
	}
	return postsLiked, commentsCount, nil
}

// Notification représente une notification pour un utilisateur.
type Notification struct {
	ID        int
	UserID    int
	Message   string
	PostID    int
	CommentID int
	CreatedAt time.Time
}

// CreateNotification insère une notification.
func CreateNotification(userID int, message string, postID, commentID int) error {
	query := `INSERT INTO notifications (user_id, message, post_id, comment_id) VALUES (?, ?, ?, ?);`
	_, err := DB.Exec(query, userID, message, postID, commentID)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// GetNotificationsByUserID récupère les notifications d'un utilisateur.
func GetNotificationsByUserID(userID int) ([]Notification, error) {
	query := `
		SELECT id, user_id, message, post_id, comment_id, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC;
	`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notifs []Notification
	for rows.Next() {
		var n Notification
		var createdAtStr string
		if err := rows.Scan(&n.ID, &n.UserID, &n.Message, &n.PostID, &n.CommentID, &createdAtStr); err != nil {
			return nil, err
		}
		parsed, perr := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if perr != nil {
			parsed, _ = time.Parse(time.RFC3339, createdAtStr)
		}
		n.CreatedAt = parsed
		notifs = append(notifs, n)
	}
	return notifs, nil
}

// DeleteNotificationsByUserID supprime toutes les notifications d'un utilisateur.
func DeleteNotificationsByUserID(userID int) error {
	query := "DELETE FROM notifications WHERE user_id = ?;"
	_, err := DB.Exec(query, userID)
	return err
}

// GetUserByID récupère un utilisateur par son ID (sans le rôle).
func GetUserByID(id int) (User, error) {
	var user User
	query := "SELECT id, username, email, created_at, photo FROM users WHERE id = ?;"
	row := DB.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.Photo)
	if err != nil {
		return user, err
	}
	if user.Photo == "" {
		user.Photo = "profil.png"
	}
	return user, nil
}

// GetCommentByID récupère un commentaire par son ID.
func GetCommentByID(commentID int) (Comment, error) {
	var c Comment
	query := `
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, u.photo
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?;
	`
	row := DB.QueryRow(query, commentID)
	var createdAtStr string
	err := row.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &createdAtStr, &c.Photo)
	if err != nil {
		return c, err
	}
	parsed, perr := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if perr != nil {
		parsed, _ = time.Parse(time.RFC3339, createdAtStr)
	}
	c.CreatedAt = parsed
	c.Likes, _ = CountCommentLikes(c.ID)
	c.Dislikes, _ = CountCommentDislikes(c.ID)
	return c, nil
}

// SetPostLike gère les likes/dislikes sur un post.
func SetPostLike(userID int, postID int, value int) error {
	var id int
	query := "SELECT id FROM likes WHERE user_id = ? AND post_id = ?;"
	err := DB.QueryRow(query, userID, postID).Scan(&id)
	if err == sql.ErrNoRows {
		insertQuery := "INSERT INTO likes (user_id, post_id, value) VALUES (?, ?, ?);"
		_, err := DB.Exec(insertQuery, userID, postID, value)
		return err
	} else if err != nil {
		return err
	}
	updateQuery := "UPDATE likes SET value = ? WHERE id = ?;"
	_, err = DB.Exec(updateQuery, value, id)
	return err
}

// SetCommentLike gère les likes/dislikes sur un commentaire.
func SetCommentLike(userID int, commentID int, value int) error {
	var id int
	query := "SELECT id FROM likes WHERE user_id = ? AND comment_id = ?;"
	err := DB.QueryRow(query, userID, commentID).Scan(&id)
	if err == sql.ErrNoRows {
		insertQuery := "INSERT INTO likes (user_id, comment_id, value) VALUES (?, ?, ?);"
		_, err := DB.Exec(insertQuery, userID, commentID, value)
		return err
	} else if err != nil {
		return err
	}
	updateQuery := "UPDATE likes SET value = ? WHERE id = ?;"
	_, err = DB.Exec(updateQuery, value, id)
	return err
}

// CountPostLikes compte les likes d'un post.
func CountPostLikes(postID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = 1;"
	var count int
	err := DB.QueryRow(query, postID).Scan(&count)
	return count, err
}

// CountPostDislikes compte les dislikes d'un post.
func CountPostDislikes(postID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = -1;"
	var count int
	err := DB.QueryRow(query, postID).Scan(&count)
	return count, err
}

// CountCommentLikes compte les likes d'un commentaire.
func CountCommentLikes(commentID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = 1;"
	var count int
	err := DB.QueryRow(query, commentID).Scan(&count)
	return count, err
}

// CountCommentDislikes compte les dislikes d'un commentaire.
func CountCommentDislikes(commentID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = -1;"
	var count int
	err := DB.QueryRow(query, commentID).Scan(&count)
	return count, err
}

// GetLastPostDate récupère la date du dernier post d'un utilisateur.
func GetLastPostDate(userID int) (time.Time, error) {
	var lastPost string
	query := "SELECT MAX(created_at) FROM posts WHERE user_id = ?;"
	err := DB.QueryRow(query, userID).Scan(&lastPost)
	if err != nil || lastPost == "" {
		return time.Time{}, err
	}
	t, err := time.Parse("2006-01-02 15:04:05", lastPost)
	if err != nil {
		t, err = time.Parse(time.RFC3339, lastPost)
	}
	return t, err
}

// GetLastCommentDate récupère la date du dernier commentaire d'un utilisateur.
func GetLastCommentDate(userID int) (time.Time, error) {
	var lastComment string
	query := "SELECT MAX(created_at) FROM comments WHERE user_id = ?;"
	err := DB.QueryRow(query, userID).Scan(&lastComment)
	if err != nil || lastComment == "" {
		return time.Time{}, err
	}
	t, err := time.Parse("2006-01-02 15:04:05", lastComment)
	if err != nil {
		t, err = time.Parse(time.RFC3339, lastComment)
	}
	return t, err
}

// GetLastLikeDate récupère la date du dernier like d'un utilisateur.
func GetLastLikeDate(userID int) (time.Time, error) {
	var lastLike string
	query := "SELECT MAX(created_at) FROM likes WHERE user_id = ?;"
	err := DB.QueryRow(query, userID).Scan(&lastLike)
	if err != nil || lastLike == "" {
		return time.Time{}, err
	}
	t, err := time.Parse("2006-01-02 15:04:05", lastLike)
	if err != nil {
		t, err = time.Parse(time.RFC3339, lastLike)
	}
	return t, err
}

// GetLastActivityDate renvoie la date de la dernière activité (commentaire ou like).
func GetLastActivityDate(userID int) (time.Time, error) {
	lastComment, err1 := GetLastCommentDate(userID)
	lastLike, err2 := GetLastLikeDate(userID)
	if err1 != nil && err2 != nil {
		return time.Time{}, fmt.Errorf("no activity found")
	}
	if lastComment.After(lastLike) {
		return lastComment, nil
	}
	return lastLike, nil
}

// GetLastConnection renvoie la date de dernière connexion (simulée ici).
func GetLastConnection(userID int) (time.Time, error) {
	return time.Now(), nil
}

// SetPostModerationStatus met à jour le statut de modération d'un post.
func SetPostModerationStatus(postID int, status string) error {
	query := "UPDATE posts SET moderation_status = ? WHERE id = ?;"
	_, err := DB.Exec(query, status, postID)
	return err
}

// GetPendingPosts récupère les posts en attente de modération.
func GetPendingPosts() ([]Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.original_content, p.image_path, p.created_at, p.modified_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.moderation_status = 'pending'
		ORDER BY p.created_at DESC;
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending posts: %w", err)
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		var createdAtStr, modifiedAtStr sql.NullString
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.OriginalContent, &p.ImagePath, &createdAtStr, &modifiedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan pending post row: %w", err)
		}
		parsed, perr := time.Parse(time.RFC3339, createdAtStr.String)
		if perr != nil {
			parsed, _ = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
		}
		p.CreatedAt = parsed.Add(2 * time.Hour)
		if modifiedAtStr.Valid && modifiedAtStr.String != "" {
			mparsed, merr := time.Parse(time.RFC3339, modifiedAtStr.String)
			if merr != nil {
				mparsed, _ = time.Parse("2006-01-02 15:04:05", modifiedAtStr.String)
			}
			p.ModifiedAt = mparsed.Add(2 * time.Hour)
		}
		p.Likes, _ = CountPostLikes(p.ID)
		p.Dislikes, _ = CountPostDislikes(p.ID)
		posts = append(posts, p)
	}
	return posts, nil
}

// --- Gestion des sessions ---
  
// CreateSession insère une session serveur pour un utilisateur.
func CreateSession(sessionID string, userID int, expiresAt time.Time) error {
	query := `INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?);`
	_, err := DB.Exec(query, sessionID, userID, expiresAt.Format("2006-01-02 15:04:05"))
	return err
}

// GetUserIDBySession retourne l'ID utilisateur lié à une session, ou une erreur si expirée/inexistante.
func GetUserIDBySession(sessionID string) (int, error) {
	var userID int
	var expiresStr string
	query := `SELECT user_id, expires_at FROM sessions WHERE session_id = ?;`
	err := DB.QueryRow(query, sessionID).Scan(&userID, &expiresStr)
	if err != nil {
		return 0, err
	}
	exp, err := time.Parse("2006-01-02 15:04:05", expiresStr)
	if err != nil {
		return 0, err
	}
	if time.Now().After(exp) {
		_ = DeleteSession(sessionID)
		return 0, fmt.Errorf("session expirée")
	}
	return userID, nil
}

// DeleteSession supprime une session côté serveur.
func DeleteSession(sessionID string) error {
	_, err := DB.Exec(`DELETE FROM sessions WHERE session_id = ?;`, sessionID)
	return err
}
