package main

import (
	"log"
	"net/http"
	"os"

	"github.com/UnstoppableMango/thecluster.lan/api/internal/server"
)

func main() {
	port := getenv("PORT", "8080")

	addr := ":" + port
	handler := server.New(staticDirs()...)

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

func staticDirs() []string {
	if value, ok := os.LookupEnv("STATIC_DIR"); ok && value != "" {
		return []string{value}
	}

	return []string{"../web/dist", "web/dist"}
}
