package dto

type CreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Document string `json:"document" binding:"required"`
	Live     bool   `json:"live" binding:"required"`
}
