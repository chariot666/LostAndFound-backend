package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"lost-found-server/internal/config"
	"lost-found-server/internal/database"
	"lost-found-server/internal/response"
)

func main() {
	cfg := config.Load()
	if _, err := database.Open(cfg); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "ok",
		})
	})

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
