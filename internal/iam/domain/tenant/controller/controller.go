package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	Create(c *gin.Context)
}
