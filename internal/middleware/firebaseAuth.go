package middleware

import (
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"

	"github.com/TradeLayers/BE/internal/model"
)

func FirebaseAuth(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		token, err := authClient.VerifyIDToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		email, ok := token.Claims["email"].(string)

		if !ok {
			email = ""
		}

		name, ok := token.Claims["name"].(string)

		if !ok {
			name = ""
		}

		authUser := model.UserContext{
			FirebaseId: token.UID,
			Email:      email,
			Name:       name,
		}

		c.Set("userContext", authUser)
		c.Next()
	}
}
