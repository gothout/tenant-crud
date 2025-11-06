package model

import (
	userModel "tenant-crud/internal/iam/domain/user/model"
	"time"
)

type Login struct {
	User  userModel.User
	Token Token
}

type Token struct {
	Token      string
	Expiration time.Time
}
