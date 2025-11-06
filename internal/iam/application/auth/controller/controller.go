package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	Login(c *gin.Context)
}
