package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func corsRequest(method, origin string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := NewRouter(Config{AllowedOrigins: []string{"http://localhost:5000", " https://meetoova.me "}})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/health", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestCORSAllowedOrigin(t *testing.T) {
	w := corsRequest(http.MethodGet, "http://localhost:5000")
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5000" {
		t.Fatalf("Allow-Origin = %q", got)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestCORSTrimsConfiguredOrigins(t *testing.T) {
	w := corsRequest(http.MethodGet, "https://meetoova.me")
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://meetoova.me" {
		t.Fatalf("Allow-Origin = %q", got)
	}
}

func TestCORSUnknownOriginGetsNoHeader(t *testing.T) {
	w := corsRequest(http.MethodGet, "https://evil.example")
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Allow-Origin = %q, want empty", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary = %q, want Origin", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	w := corsRequest(http.MethodOptions, "http://localhost:5000")
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("Allow-Headers = %q", got)
	}
}
