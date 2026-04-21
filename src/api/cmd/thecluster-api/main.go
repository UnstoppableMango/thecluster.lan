package main

import (
	"log"
	"net/http"
	"os"

	"github.com/UnstoppableMango/thecluster.lan/src/api/internal/server"
)

func main() {
	port := getenv("PORT", "8080")
	staticDir := getenv("STATIC_DIR", "src/web/dist")

	addr := ":" + port
	handler := server.New(staticDir)

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func getenv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}
