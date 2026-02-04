package user

import "time"

// User representa un usuario (propietario de una unidad o admin).
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	UnitID       string    `json:"unit_id"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateInput es el payload para crear un usuario (admin).
type CreateInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UnitID   string `json:"unit_id"`
}

// ChangePasswordInput es el payload para cambiar contraseña.
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
