package router

import (
	"os"

	"firebase.google.com/go/auth"
	appErrors "github.com/TradeLayers/BE/internal/appErrors"
	appLogger "github.com/TradeLayers/BE/internal/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/TradeLayers/BE/internal/handler"
	"github.com/TradeLayers/BE/internal/middleware"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/service"
)

func Setup(db *gorm.DB, log *zap.Logger, authClient *auth.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		appErrors.ReturnError(c, appErrors.ErrInternal)
	}))
	r.Use(appLogger.HTTPMiddleware(log))

	frontendUrl := os.Getenv("FRONTEND_URL")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendUrl},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	healthHanlder := handler.NewHealthHandler(db)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	api := r.Group("/api")
	{
		api.GET("/health", healthHanlder.Health)

		protected := api.Group("")

		protected.Use(middleware.FirebaseAuth(authClient))

		protected.POST("/user", userHandler.CreateOrFetchUser)
		protected.PATCH("/user", userHandler.UpdateFields)
		protected.DELETE("/user", userHandler.DeleteUserAccount)
	}

	return r
}
