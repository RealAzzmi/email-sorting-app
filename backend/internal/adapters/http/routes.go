package http

import (
	"github.com/email-sorting-app/internal/adapters/http/handlers"
	"github.com/email-sorting-app/internal/adapters/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	authHandler *handlers.AuthHandler,
	accountHandler *handlers.AccountHandler,
	categoryHandler *handlers.CategoryHandler,
	emailHandler *handlers.EmailHandler,
) *gin.Engine {
	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORS())

	// Auth routes
	router.GET("/auth/login", authHandler.Login)
	router.GET("/auth/callback", authHandler.Callback)
	router.POST("/auth/logout", authHandler.Logout)

	// Account routes
	router.GET("/accounts", accountHandler.GetAccounts)
	router.DELETE("/accounts/:id", accountHandler.DeleteAccount)

	// Category routes
	router.GET("/accounts/:id/categories", categoryHandler.GetAccountCategories)

	// Email routes
	router.GET("/accounts/:id/emails", emailHandler.GetAccountEmails)
	router.POST("/accounts/:id/emails/refresh", emailHandler.RefreshAccountEmails)

	return router
}