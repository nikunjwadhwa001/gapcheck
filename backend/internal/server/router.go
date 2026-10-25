package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/analysis"
)

// Analyzer is what the analyze handler needs. *analysis.Service satisfies
// it; tests pass a fake so they never call a real LLM.
type Analyzer interface {
	Analyze(ctx context.Context, req analysis.Request) (analysis.Result, error)
}

// Config holds the router's dependencies.
type Config struct {
	Analyzer   Analyzer
	LLMTimeout time.Duration // max time for one analysis
}

// NewRouter wires up all HTTP routes for the API.
func NewRouter(cfg Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", health)

	api := r.Group("/api")
	api.POST("/analyze", analyzeHandler(cfg.Analyzer, cfg.LLMTimeout))

	return r
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
