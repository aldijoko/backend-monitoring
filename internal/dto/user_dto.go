package dto

import "monitoring-cctv-be/internal/model"

type UserListResponse struct {
	Items  []model.User `json:"items"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

type CreateUserRequest struct {
	Username string         `json:"username" binding:"required,min=3"`
	Email    string         `json:"email" binding:"required,email"`
	Role     model.UserRole `json:"role" binding:"required"`
	Password string         `json:"password" binding:"required,min=8"`
	IsActive *bool          `json:"is_active"`
}

type UpdateUserRequest struct {
	Email    string         `json:"email" binding:"required,email"`
	Role     model.UserRole `json:"role" binding:"required"`
	IsActive bool           `json:"is_active"`
}

type SetActiveRequest struct {
	IsActive bool `json:"is_active"`
}

type ResetPasswordResponse struct {
	TemporaryPassword string `json:"temporary_password"`
}
