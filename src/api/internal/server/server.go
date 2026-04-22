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

func New(staticDirs ...string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", handlePing)

	fileServer := newStaticHandler(staticDirs...)
	if fileServer != nil {
		mux.Handle("/", fileServer)
	}

	return mux
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pingResponse{Message: "pong"})
}

func newStaticHandler(staticDirs ...string) http.Handler {
	root := resolveStaticDir(staticDirs...)
	if root == "" {
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		path, ok := resolveStaticPath(root, r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

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

		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		indexPath := filepath.Join(absPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return absPath
		}
	}

	return ""
}

func resolveStaticPath(root string, requestPath string) (string, bool) {
	target := filepath.Clean(strings.TrimPrefix(requestPath, "/"))
	if target == "." || target == "" {
		return filepath.Join(root, "index.html"), true
	}

	path := filepath.Join(root, target)
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}

	return path, true
}
