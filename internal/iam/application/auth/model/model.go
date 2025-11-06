package model

import (
	userModel "tenant-crud/internal/iam/domain/user/model"
	"time"

	"github.com/google/uuid"
)

type AcessToken struct {
	UserUUID *uuid.UUID `gorm:"type:uuid;index"`
	Token    string     `gorm:"type:varchar(255);not null"`
	Expiry   time.Time  `gorm:"type:timestamp;not null;column:expire_date"`
}
type Login struct {
	User       userModel.User
	AcessToken AcessToken
}

func (AcessToken) TableName() string {
	return "users_acess_tokens"
}
