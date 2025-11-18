package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var jwtSecret = []byte("bangladesh2025")

type ctxKey string

// CORS middleware to allow frontend access
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var id int
	var passwordHash string

	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", creds.Email).Scan(&id, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"token": tokenString, "user_id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "name, email and password required", http.StatusBadRequest)
		return
	}
	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	// insert user - match your schema: (name, email, password)
	res, err := db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", req.Name, req.Email, string(hash))
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	id64, _ := res.LastInsertId()
	id := int(id64)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "name": req.Name, "email": req.Email})
}

func CreatePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// get user id from context (set by JWT middleware)
	userID := 0
	if v := r.Context().Value(ctxKey("user_id")); v != nil {
		if uid, ok := v.(int); ok {
			userID = uid
		}
	}
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// basic trimming and validation
	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)
	if post.Thumbnail != nil {
		trimmed := strings.TrimSpace(*post.Thumbnail)
		post.Thumbnail = &trimmed
	}
	post.Status = strings.TrimSpace(post.Status)
	if post.Title == "" || post.Content == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)
		return
	}
	// limit thumbnail size to avoid excessively long strings
	if post.Thumbnail != nil && len(*post.Thumbnail) > 2048 {
		http.Error(w, "thumbnail too long", http.StatusBadRequest)
		return
	}
	if post.Status == "" {
		post.Status = "draft"
	} else if post.Status != "draft" && post.Status != "published" {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	// include user_id to satisfy NOT NULL and FK constraint in your schema
	res, err := db.Exec("INSERT INTO posts (user_id, title, content, thumbnail, status) VALUES (?, ?, ?, ?, ?)", userID, post.Title, post.Content, post.Thumbnail, post.Status)
	if err != nil {
		http.Error(w, "Failed to save post", http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	post.ID = uint64(id)
	post.UserID = uint64(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// jwtAuth is middleware that validates a Bearer JWT and sets user_id in context
func jwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}
		// extract user_id as float64 then convert to int
		if uidVal, ok := claims["user_id"]; ok {
			switch v := uidVal.(type) {
			case float64:
				ctx := context.WithValue(r.Context(), ctxKey("user_id"), int(v))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			case int:
				ctx := context.WithValue(r.Context(), ctxKey("user_id"), v)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		http.Error(w, "user_id not found in token", http.StatusUnauthorized)
	})
}

func GetAllPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.Query("SELECT id, user_id, title, content, thumbnail, status, created_at, updated_at FROM posts")
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Thumbnail, &post.Status, &post.CreatedAt, &post.UpdatedAt); err != nil {
			fmt.Println("Scan error:", err)
			http.Error(w, "Failed to scan post: "+err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func GetPostByID(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	if postID == "" {
		http.Error(w, "post id required", http.StatusBadRequest)
		return
	}

	var post Post
	err := db.QueryRow("SELECT id, user_id, title, content, thumbnail, status, created_at, updated_at FROM posts WHERE id = ?", postID).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Content, &post.Thumbnail, &post.Status, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	if postID == "" {
		http.Error(w, "post id required", http.StatusBadRequest)
		return
	}

	userID := 0
	if v := r.Context().Value(ctxKey("user_id")); v != nil {
		if uid, ok := v.(int); ok {
			userID = uid
		}
	}
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)
	if post.Thumbnail != nil {
		trimmed := strings.TrimSpace(*post.Thumbnail)
		post.Thumbnail = &trimmed
	}
	post.Status = strings.TrimSpace(post.Status)

	if post.Title == "" || post.Content == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)
		return
	}

	res, err := db.Exec("UPDATE posts SET title = ?, content = ?, thumbnail = ?, status = ?, updated_at = NOW() WHERE id = ? AND user_id = ?",
		post.Title, post.Content, post.Thumbnail, post.Status, postID, userID)
	if err != nil {
		http.Error(w, "failed to update post", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "post not found or unauthorized", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "post updated successfully"})
}

func DeletePosts(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	if postID == "" {
		http.Error(w, "post id required", http.StatusBadRequest)
		return
	}

	userID := 0
	if v := r.Context().Value(ctxKey("user_id")); v != nil {
		if uid, ok := v.(int); ok {
			userID = uid
		}
	}
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := db.Exec("DELETE FROM posts WHERE id = ? AND user_id = ?", postID, userID)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No post found or unauthorized", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetMyPosts(w http.ResponseWriter, r *http.Request) {
	userID := 0
	if v := r.Context().Value(ctxKey("user_id")); v != nil {
		if uid, ok := v.(int); ok {
			userID = uid
		}
	}
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.Query("SELECT id, user_id, title, content, thumbnail, status, created_at, updated_at FROM posts WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Thumbnail, &post.Status, &post.CreatedAt, &post.UpdatedAt); err != nil {
			http.Error(w, "Failed to scan post", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:29112003@tcp(127.0.0.1:3306)/blogdb?parseTime=true")
	if err != nil {
		fmt.Println("Failed to open database:", err)
		return
	}

	if err = db.Ping(); err != nil {
		fmt.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(corsMiddleware)

	// API routes
	r.Post("/register", registerHandler)
	r.Post("/login", loginHandler)

	// Post routes
	r.Get("/posts", GetAllPosts)
	r.Get("/posts/{id}", GetPostByID)
	r.With(jwtAuth).Post("/posts", CreatePosts)
	r.With(jwtAuth).Put("/posts/{id}", UpdatePost)
	r.With(jwtAuth).Delete("/posts/{id}", DeletePosts)
	r.With(jwtAuth).Get("/my-posts", GetMyPosts)

	// Serve frontend files
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/index.html")
	})
	r.Get("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/style.css")
	})
	r.Get("/script.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/script.js")
	})

	fmt.Println("Server is running on http://localhost:3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		fmt.Println("Server error:", err)
	}
}
