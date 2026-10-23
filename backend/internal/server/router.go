package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter wires up all HTTP routes for the API.
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", health)

	return r
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
