package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"lost-found-server/internal/config"
	"lost-found-server/internal/database"
	"lost-found-server/internal/handler"
	"lost-found-server/internal/middleware"
	"lost-found-server/internal/model"
	"lost-found-server/internal/response"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Static("/uploads", cfg.UploadDir)

	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "ok",
		})
	})

	authHandler := handler.NewAuthHandler(db, cfg.JWTSecret)
	itemHandler := handler.NewItemHandler(db)
	claimHandler := handler.NewClaimHandler(db)
	adminHandler := handler.NewAdminHandler(db)
	announcementHandler := handler.NewAnnouncementHandler(db)
	uploadHandler := handler.NewUploadHandler(cfg.UploadDir)
	authMiddleware := middleware.NewAuthMiddleware(db, cfg.JWTSecret)

	api := router.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.GET("/me", authMiddleware.RequireAuth(), authHandler.Me)

	api.GET("/items", authMiddleware.OptionalAuth(), itemHandler.List)
	api.GET("/items/:id", authMiddleware.OptionalAuth(), itemHandler.Detail)
	api.GET("/announcements", announcementHandler.List)

	protected := api.Group("")
	protected.Use(authMiddleware.RequireAuth())
	protected.POST("/items", itemHandler.Create)
	protected.PUT("/items/:id", itemHandler.Update)
	protected.DELETE("/items/:id", itemHandler.Delete)
	protected.GET("/me/items", itemHandler.MyItems)
	protected.POST("/items/:id/claims", claimHandler.Create)
	protected.GET("/me/claims", claimHandler.Mine)
	protected.POST("/upload", uploadHandler.Upload)

	admin := api.Group("/admin")
	admin.Use(
		authMiddleware.RequireAuth(),
		authMiddleware.RequireRoles(model.RoleItemAdmin, model.RoleSystemAdmin),
	)
	admin.GET("/items", adminHandler.Items)
	admin.PUT("/items/:id", adminHandler.ReviewItem)
	admin.GET("/claims", adminHandler.Claims)
	admin.PUT("/claims/:id", adminHandler.ReviewClaim)
	admin.GET("/statistics", adminHandler.Statistics)

	systemAdmin := api.Group("/admin")
	systemAdmin.Use(
		authMiddleware.RequireAuth(),
		authMiddleware.RequireRoles(model.RoleSystemAdmin),
	)
	systemAdmin.GET("/users", adminHandler.Users)
	systemAdmin.PUT("/users/:id", adminHandler.UpdateUser)
	systemAdmin.GET("/announcements", announcementHandler.AdminList)
	systemAdmin.POST("/announcements", announcementHandler.Create)
	systemAdmin.PUT("/announcements/:id", announcementHandler.Update)
	systemAdmin.DELETE("/announcements/:id", announcementHandler.Delete)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
