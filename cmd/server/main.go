package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"lost-found-server/internal/response"
)

func main() {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "ok",
		})
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
