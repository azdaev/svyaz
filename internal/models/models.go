package models

import "time"

type Role struct {
	ID    int64
	Slug  string
	Name  string
	Count int
}

type User struct {
	ID         int64
	TgID       int64
	TgUsername string
	Name       string
	Bio        string
	Experience string
	Skills     []string
	PhotoURL   string
	TgChatID   int64
	Onboarded  bool
	IsAdmin    bool
	IsBanned   bool
	Roles      []Role
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Project struct {
	ID          int64
	Slug        string
	AuthorID    int64
	Title       string
	Description string
	Status      string
	Stack       []string
	Roles       []Role
	Author      *User
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AdminStats struct {
	UserCount       int
	ProjectTotal    int
	ProjectPending  int
	ProjectActive   int
	ProjectHidden   int
	ResponseCount   int
}

type Response struct {
	ID        int64
	ProjectID int64
	UserID    int64
	Status    string
	User      *User
	Project   *Project
	CreatedAt time.Time
}

type Notification struct {
	ID        int64
	UserID    int64
	Type      string
	Payload   map[string]interface{}
	Read      bool
	CreatedAt time.Time
}
