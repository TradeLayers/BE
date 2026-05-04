package middleware

import (
	"context"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/requestlog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/TradeLayers/BE/internal/model"
)

type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

func FirebaseAuth(authClient TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := requestlog.FromContext(c.Request.Context())

		header := c.GetHeader("Authorization")
		if header == "" {
			log.Warn("missing authorization header")
			appErrors.ReturnError(c, appErrors.ErrNoAuthenticationHeader)
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		token, err := authClient.VerifyIDToken(c.Request.Context(), tokenString)
		if err != nil {
			log.Warn("failed to verify id token", zap.Error(err))
			appErrors.ReturnError(c, appErrors.ErrJwtExpired)
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
