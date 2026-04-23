package server

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/olivere/vite"
)

//go:embed 404.html
var notFoundPage []byte

type pingResponse struct {
	Message string `json:"message"`
}

func New(staticDirs ...string) (http.Handler, error) {
	r := chi.NewRouter()
	r.Get("/ping", handlePing)

	if root := resolveStaticDir(staticDirs...); root != "" {
		fsys := os.DirFS(root)
		vh, err := vite.NewHandler(vite.Config{
			FS:    fsys,
			IsDev: false,
		})
		if err != nil {
			return nil, err
		}
		r.Get("/", vh.ServeHTTP)
		r.Handle("/assets/*", vh)
		r.NotFound(notFoundHandler)
	}

	return r, nil
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pingResponse{Message: "pong"})
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(notFoundPage)
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

		if _, err := os.Stat(filepath.Join(absPath, "index.html")); err == nil {
			return absPath
		}
	}

	return ""
}
