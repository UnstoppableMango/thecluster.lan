package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type pingResponse struct {
	Message string `json:"message"`
}

func New(staticDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", handlePing)

	fileServer := newStaticHandler(staticDir)
	if fileServer != nil {
		mux.Handle("/", fileServer)
	}

	return mux
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pingResponse{Message: "pong"})
}

func newStaticHandler(staticDir string) http.Handler {
	root := resolveStaticDir(staticDir)
	if root == "" {
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		target := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if target == "." || target == "" {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}

		path := filepath.Join(root, target)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}

		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}

func resolveStaticDir(paths ...string) string {
	for _, path := range paths {
		if path == "" {
			continue
		}

		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return path
		}
	}

	return ""
}
