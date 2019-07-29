package ttt

import (
	"github.com/gin-gonic/gin"
)

func Abs(code int,message string,c *gin.Context)  {
	c.JSON(code,gin.H{
		"message": message,
	})
}
