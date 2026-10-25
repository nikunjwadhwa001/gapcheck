package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/analysis"
	"github.com/nikunjwadhwa001/gapcheck/backend/internal/gemini"
	"github.com/nikunjwadhwa001/gapcheck/backend/internal/server"
)

// shutdownGrace is how long in-flight requests get to finish after a stop
// signal. Render waits 30s before force-killing, so stay under that.
const shutdownGrace = 25 * time.Second

func main() {
	// Local development reads backend/.env. In production (Render) the file
	// doesn't exist and real environment variables are used instead.
	_ = godotenv.Load()

	port := getenv("PORT", "8080")
	model := getenv("GEMINI_MODEL", "gemini-3.5-flash-lite")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set (copy .env.example to .env and add your key)")
	}

	timeout, err := time.ParseDuration(getenv("LLM_TIMEOUT", "30s"))
	if err != nil {
		log.Fatalf("LLM_TIMEOUT is not a valid duration: %v", err)
	}

	svc := analysis.NewService(gemini.New(apiKey, model))
	r := server.NewRouter(server.Config{
		Analyzer:       svc,
		LLMTimeout:     timeout,
		AllowedOrigins: strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
	})

	// An explicit http.Server instead of r.Run, so we can set timeouts and
	// shut down gracefully.
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second, // slow-header (Slowloris) protection
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      timeout + 15*time.Second, // an analysis may take up to LLM_TIMEOUT
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("gapcheck api listening on :%s (model %s)", port, model)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Wait for Ctrl+C locally, or SIGTERM when Render redeploys or stops us.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Printf("shutting down, giving requests up to %s to finish", shutdownGrace)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
