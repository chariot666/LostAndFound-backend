package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"lost-found-server/internal/config"
	"lost-found-server/internal/database"
	"lost-found-server/internal/handler"
	"lost-found-server/internal/response"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "ok",
		})
	})

	authHandler := handler.NewAuthHandler(db)
	api := router.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
