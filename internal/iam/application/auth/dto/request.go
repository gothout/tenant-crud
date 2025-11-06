package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
