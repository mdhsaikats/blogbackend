package main

import "time"

// ---------------- Users ----------------
type User struct {
	ID        uint64    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"` // omit in JSON response
	Role      string    `db:"role" json:"role"`  // admin, author, user
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ---------------- Posts ----------------
type Post struct {
	ID        uint64    `db:"id" json:"id"`
	UserID    uint64    `db:"user_id" json:"user_id"`
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	Thumbnail *string   `db:"thumbnail" json:"thumbnail,omitempty"` // pointer allows NULL
	Status    string    `db:"status" json:"status"`                 // draft, published
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ---------------- Categories ----------------
type Category struct {
	ID        uint64    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ---------------- PostCategories (junction) ----------------
type PostCategory struct {
	PostID     uint64 `db:"post_id" json:"post_id"`
	CategoryID uint64 `db:"category_id" json:"category_id"`
}

// ---------------- Comments ----------------
type Comment struct {
	ID        uint64    `db:"id" json:"id"`
	PostID    uint64    `db:"post_id" json:"post_id"`
	UserName  string    `db:"user_name" json:"user_name"`
	UserEmail string    `db:"user_email" json:"user_email"`
	Content   string    `db:"content" json:"content"`
	ParentID  *uint64   `db:"parent_id" json:"parent_id"` // pointer allows NULL
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
