package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/analysis"
	"github.com/nikunjwadhwa001/gapcheck/backend/internal/gemini"
	"github.com/nikunjwadhwa001/gapcheck/backend/internal/server"
)

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
	r := server.NewRouter(server.Config{Analyzer: svc, LLMTimeout: timeout})

	log.Printf("gapcheck api listening on :%s (model %s)", port, model)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
