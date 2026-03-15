package handler

import "github.com/gin-gonic/gin"

func GetPortfolio(c *gin.Context) {

	userID := c.GetString("userID")

	c.JSON(200, gin.H{
		"user": userID,
	})
}
