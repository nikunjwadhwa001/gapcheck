package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/analysis"
)

// maxBodyBytes caps the request body. Two 20,000-character texts fit with
// room to spare even with multi-byte characters.
const maxBodyBytes = 200 << 10 // 200 KB

// apiError is the one error shape every endpoint returns:
// {"error": {"code": "...", "message": "...", "field": "..."}}
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func writeError(c *gin.Context, status int, e apiError) {
	c.AbortWithStatusJSON(status, gin.H{"error": e})
}

func analyzeHandler(an Analyzer, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

		var req analysis.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				writeError(c, http.StatusRequestEntityTooLarge, apiError{
					Code:    "body_too_large",
					Message: "Request is too large. Keep the resume and job description under 20,000 characters each.",
				})
				return
			}
			writeError(c, http.StatusBadRequest, apiError{
				Code:    "invalid_json",
				Message: `Send JSON like {"resume_text": "...", "jd_text": "..."}.`,
			})
			return
		}

		// The analysis stops when the timeout hits or the client disconnects.
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		res, err := an.Analyze(ctx, req)
		if err != nil {
			writeAnalyzeError(c, err)
			return
		}
		c.JSON(http.StatusOK, res)
	}
}

// writeAnalyzeError maps internal errors to HTTP statuses. Internal details
// are logged for us, never sent to the user.
func writeAnalyzeError(c *gin.Context, err error) {
	var ve *analysis.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(c, http.StatusBadRequest, apiError{Code: "invalid_input", Message: ve.Error(), Field: ve.Field})
	case errors.Is(err, context.DeadlineExceeded):
		writeError(c, http.StatusGatewayTimeout, apiError{Code: "timeout", Message: "The analysis took too long. Please try again."})
	case errors.Is(err, analysis.ErrRateLimited):
		writeError(c, http.StatusTooManyRequests, apiError{Code: "rate_limited", Message: "Too many analyses right now. Please wait a minute and try again."})
	default:
		log.Printf("analyze failed: %v", err)
		writeError(c, http.StatusBadGateway, apiError{Code: "upstream_error", Message: "The AI service had a problem. Please try again."})
	}
}
