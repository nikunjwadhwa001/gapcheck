package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// cors lets browser apps on the allowed origins call this API. Browsers
// block cross-origin requests unless the server opts in with these headers.
func cors(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = true
		}
	}

	return func(c *gin.Context) {
		h := c.Writer.Header()
		// Responses differ per Origin, so caches must not share them.
		h.Add("Vary", "Origin")

		if origin := c.GetHeader("Origin"); allowed[origin] {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type")
			h.Set("Access-Control-Max-Age", "600") // cache the preflight for 10 minutes
		}

		// The browser's preflight check: answer it here, never reach a handler.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
