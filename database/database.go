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

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt string
	Photo     string
}

func CreateUser(username, email, password string) error {
	query := `INSERT INTO users (username, email, password, photo) VALUES (?, ?, ?, ?);`
	_, err := DB.Exec(query, username, email, password, "profil.png")
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

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

func CreatePost(userID int, title, content, imagePath string) error {
	query := `INSERT INTO posts (user_id, title, content, original_content, image_path) VALUES (?, ?, ?, ?, ?);`
	_, err := DB.Exec(query, userID, title, content, content, imagePath)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}

func GetAllPosts() ([]Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.original_content, p.image_path, p.created_at, p.modified_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
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
		parsedTime, err := time.Parse(time.RFC3339, createdAtStr.String)
		if err != nil {
			parsedTime, err = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
			if err != nil {
				parsedTime = time.Now()
			}
		}
		p.CreatedAt = parsedTime.Add(2 * time.Hour)
		if modifiedAtStr.Valid && modifiedAtStr.String != "" {
			modTime, err := time.Parse(time.RFC3339, modifiedAtStr.String)
			if err != nil {
				modTime, err = time.Parse("2006-01-02 15:04:05", modifiedAtStr.String)
				if err != nil {
					modTime = time.Time{}
				}
			}
			p.ModifiedAt = modTime.Add(2 * time.Hour)
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

func GetPostByID(id int) (Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.original_content, p.image_path, p.created_at, p.modified_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?;
	`
	row := DB.QueryRow(query, id)
	var p Post
	var createdAtStr, modifiedAtStr sql.NullString
	if err := row.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.OriginalContent, &p.ImagePath, &createdAtStr, &modifiedAtStr); err != nil {
		return p, fmt.Errorf("failed to get post by ID: %w", err)
	}
	parsedTime, err := time.Parse(time.RFC3339, createdAtStr.String)
	if err != nil {
		parsedTime, err = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
		if err != nil {
			parsedTime = time.Now()
		}
	}
	p.CreatedAt = parsedTime.Add(2 * time.Hour)
	if modifiedAtStr.Valid && modifiedAtStr.String != "" {
		modTime, err := time.Parse(time.RFC3339, modifiedAtStr.String)
		if err != nil {
			modTime, err = time.Parse("2006-01-02 15:04:05", modifiedAtStr.String)
			if err != nil {
				modTime = time.Time{}
			}
		}
		p.ModifiedAt = modTime.Add(2 * time.Hour)
	}
	p.Likes, _ = CountPostLikes(p.ID)
	p.Dislikes, _ = CountPostDislikes(p.ID)
	return p, nil
}

func DeletePost(postID int, userID int) error {
	query := "DELETE FROM posts WHERE id = ? AND user_id = ?"
	_, err := DB.Exec(query, postID, userID)
	return err
}

func UpdatePost(postID int, userID int, title, content, imagePath string) error {
	var oldContent string
	var oldOriginal sql.NullString
	err := DB.QueryRow("SELECT content, original_content FROM posts WHERE id = ? AND user_id = ?", postID, userID).Scan(&oldContent, &oldOriginal)
	if err != nil {
		return err
	}
	if !oldOriginal.Valid || oldOriginal.String == "" {
		oldOriginal.String = oldContent
	}
	query := "UPDATE posts SET title = ?, content = ?, image_path = ?, modified_at = CURRENT_TIMESTAMP, original_content = ? WHERE id = ? AND user_id = ?"
	_, err = DB.Exec(query, title, content, imagePath, oldOriginal.String, postID, userID)
	return err
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
	Photo     string // Champ ajouté pour stocker la photo de l'utilisateur
}

func CreateComment(postID int, userID int, content string) error {
	query := "INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)"
	_, err := DB.Exec(query, postID, userID, content)
	return err
}

func DeleteComment(commentID int, userID int) error {
	query := "DELETE FROM comments WHERE id = ? AND user_id = ?"
	_, err := DB.Exec(query, commentID, userID)
	return err
}

func GetCommentsByPostID(postID int) ([]Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, u.photo
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
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
		parsedTime, parseErr := time.Parse(time.RFC3339, createdAtStr)
		if parseErr != nil {
			parsedTime, parseErr = time.Parse("2006-01-02 15:04:05", createdAtStr)
			if parseErr != nil {
				parsedTime = time.Now()
			}
		}
		c.CreatedAt = parsedTime.Add(2 * time.Hour)
		c.Likes, _ = CountCommentLikes(c.ID)
		c.Dislikes, _ = CountCommentDislikes(c.ID)
		comments = append(comments, c)
	}
	return comments, nil
}

func GetUserStats(userID int) (int, int, error) {
	var postsLiked int
	var commentsCount int
	err := DB.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id IS NOT NULL AND value=1", userID).Scan(&postsLiked)
	if err != nil {
		return 0, 0, err
	}
	err = DB.QueryRow("SELECT COUNT(*) FROM comments WHERE user_id = ?", userID).Scan(&commentsCount)
	if err != nil {
		return 0, 0, err
	}
	return postsLiked, commentsCount, nil
}

type Notification struct {
	ID        int
	UserID    int
	Message   string
	PostID    int
	CommentID int
	CreatedAt time.Time
}

func CreateNotification(userID int, message string, postID, commentID int) error {
	query := `INSERT INTO notifications (user_id, message, post_id, comment_id) VALUES (?, ?, ?, ?)`
	_, err := DB.Exec(query, userID, message, postID, commentID)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func GetNotificationsByUserID(userID int) ([]Notification, error) {
	query := `
		SELECT id, user_id, message, post_id, comment_id, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
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
		parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if parseErr != nil {
			parsedTime, parseErr = time.Parse(time.RFC3339, createdAtStr)
			if parseErr != nil {
				parsedTime = time.Now()
			}
		}
		n.CreatedAt = parsedTime
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func DeleteNotificationsByUserID(userID int) error {
	query := "DELETE FROM notifications WHERE user_id = ?"
	_, err := DB.Exec(query, userID)
	return err
}

func GetUserByID(id int) (User, error) {
	var user User
	query := "SELECT id, username, email, created_at, photo FROM users WHERE id = ?"
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

func GetCommentByID(commentID int) (Comment, error) {
	var c Comment
	query := `
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, u.photo
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`
	row := DB.QueryRow(query, commentID)
	var createdAtStr string
	err := row.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &createdAtStr, &c.Photo)
	if err != nil {
		return c, err
	}
	parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if parseErr != nil {
		parsedTime, parseErr = time.Parse(time.RFC3339, createdAtStr)
		if parseErr != nil {
			parsedTime = time.Now()
		}
	}
	c.CreatedAt = parsedTime
	c.Likes, _ = CountCommentLikes(c.ID)
	c.Dislikes, _ = CountCommentDislikes(c.ID)
	return c, nil
}

func SetPostLike(userID int, postID int, value int) error {
	var id int
	query := "SELECT id FROM likes WHERE user_id = ? AND post_id = ?"
	err := DB.QueryRow(query, userID, postID).Scan(&id)
	if err == sql.ErrNoRows {
		insertQuery := "INSERT INTO likes (user_id, post_id, value) VALUES (?, ?, ?)"
		_, err := DB.Exec(insertQuery, userID, postID, value)
		return err
	} else if err != nil {
		return err
	}
	updateQuery := "UPDATE likes SET value = ? WHERE id = ?"
	_, err = DB.Exec(updateQuery, value, id)
	return err
}

func SetCommentLike(userID int, commentID int, value int) error {
	var id int
	query := "SELECT id FROM likes WHERE user_id = ? AND comment_id = ?"
	err := DB.QueryRow(query, userID, commentID).Scan(&id)
	if err == sql.ErrNoRows {
		insertQuery := "INSERT INTO likes (user_id, comment_id, value) VALUES (?, ?, ?)"
		_, err := DB.Exec(insertQuery, userID, commentID, value)
		return err
	} else if err != nil {
		return err
	}
	updateQuery := "UPDATE likes SET value = ? WHERE id = ?"
	_, err = DB.Exec(updateQuery, value, id)
	return err
}

func CountPostLikes(postID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = 1"
	var count int
	err := DB.QueryRow(query, postID).Scan(&count)
	return count, err
}

func CountPostDislikes(postID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = -1"
	var count int
	err := DB.QueryRow(query, postID).Scan(&count)
	return count, err
}

func CountCommentLikes(commentID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = 1"
	var count int
	err := DB.QueryRow(query, commentID).Scan(&count)
	return count, err
}

func CountCommentDislikes(commentID int) (int, error) {
	query := "SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = -1"
	var count int
	err := DB.QueryRow(query, commentID).Scan(&count)
	return count, err
}

func GetLastPostDate(userID int) (time.Time, error) {
	var lastPost string
	query := "SELECT MAX(created_at) FROM posts WHERE user_id = ?"
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

func GetLastCommentDate(userID int) (time.Time, error) {
	var lastComment string
	query := "SELECT MAX(created_at) FROM comments WHERE user_id = ?"
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

func GetLastLikeDate(userID int) (time.Time, error) {
	var lastLike string
	query := "SELECT MAX(created_at) FROM likes WHERE user_id = ?"
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

func GetLastConnection(userID int) (time.Time, error) {
	return time.Now(), nil
}