package models

type User struct {
	ID           int64  `json:"id"`
	Phone        string `json:"phone"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	CreatedAt    string `json:"created_at"`
}

type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleKeyholder UserRole = "keyholder"
)