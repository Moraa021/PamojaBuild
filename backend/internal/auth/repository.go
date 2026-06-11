package auth

import (
	"database/sql"
	"pamoja-build/backend/internal/db"
	"pamoja-build/backend/internal/models"
)

type Repository struct{}

func (r *Repository) CreateUser(phone, passwordHash string) (*models.User, error) {
	result, err := db.DB.Exec(
		"INSERT INTO users (phone, password_hash) VALUES (?, ?)",
		phone, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetUserByID(userID)
}

func (r *Repository) GetUserByPhone(phone string) (*models.User, error) {
	var user models.User
	var passwordHash string

	err := db.DB.QueryRow(
		"SELECT id, phone, password_hash, role, created_at FROM users WHERE phone = ?",
		phone,
	).Scan(&user.ID, &user.Phone, &passwordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	user.PasswordHash = passwordHash
	return &user, nil
}

func (r *Repository) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	var passwordHash string

	err := db.DB.QueryRow(
		"SELECT id, phone, password_hash, role, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Phone, &passwordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	user.PasswordHash = passwordHash
	return &user, nil
}

func (r *Repository) UpdateUserRole(userID int64, role string) error {
	_, err := db.DB.Exec("UPDATE users SET role = ? WHERE id = ?", role, userID)
	return err
}
