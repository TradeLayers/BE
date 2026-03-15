package middleware

import (
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
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

		c.Set("userID", token.UID)

		c.Next()
	}
}
