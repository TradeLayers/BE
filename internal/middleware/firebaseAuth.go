package middleware

import (
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"

	"github.com/TradeLayers/BE/internal/model"
)

func FirebaseAuth(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatus(401)
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		token, err := authClient.VerifyIDToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}

		authUser := model.UserContext{
			FirebaseId: token.UID,
			Email:      token.Claims["email"].(string),
			Name:       token.Claims["name"].(string),
		}

		c.Set("userContext", authUser)
		c.Next()
	}
}
