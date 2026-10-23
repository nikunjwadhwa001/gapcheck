package main

import (
	"log"
	"os"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := server.NewRouter()

	log.Printf("gapcheck api listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
