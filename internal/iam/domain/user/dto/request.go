package dto

import "tenant-crud/internal/iam/domain/user/model"

type CreateUserRequest struct {
	Name     string         `json:"name" binding:"required"`
	Email    string         `json:"email" binding:"required,email"`
	Password string         `json:"password" binding:"required,min=8"`
	Role     model.UserRole `json:"role" binding:"required"`
}
