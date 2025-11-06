package dto

import (
	"tenant-crud/internal/iam/domain/user/dto"
	"time"
)

type LoginResponse struct {
	User   dto.UserResponse `json:"user"`
	Token  string           `json:"token"`
	Expire time.Time        `json:"expire"`
}
