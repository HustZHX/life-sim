package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GateRequest struct {
	Code string `json:"code" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type AuthUserResponse struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type AuthStatusResponse struct {
	Enabled     bool `json:"enabled"`
	GatePassed  bool `json:"gate_passed"`
	LoggedIn    bool `json:"logged_in"`
}
