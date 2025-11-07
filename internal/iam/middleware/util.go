package middleware

import (
	userModel "tenant-crud/internal/iam/domain/user/model"
	"time"

	"github.com/gin-gonic/gin"
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

const UserContextKey = "AuthenticatedUserKey"

func SetAuthenticatedUser(c *gin.Context, userLogin *Login) {
	if userLogin != nil {
		c.Set(UserContextKey, userLogin)
	}
}
func GetAuthenticatedUser(c *gin.Context) (*Login, bool) {
	value, exists := c.Get(UserContextKey)

	if !exists {
		return nil, false
	}
	userLogin, ok := value.(*Login)

	if !ok {
		return nil, false
	}

	return userLogin, true
}
