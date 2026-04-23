package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/olivere/vite"
)

type pingResponse struct {
	Message string `json:"message"`
}

func New(staticDirs ...string) (http.Handler, error) {
	r := chi.NewRouter()
	r.Get("/ping", handlePing)

	if root := resolveStaticDir(staticDirs...); root != "" {
		vh, err := vite.NewHandler(vite.Config{
			FS:    os.DirFS(root),
			IsDev: false,
		})
		if err != nil {
			return nil, err
		}
		r.Handle("/*", vh)
	}

	return r, nil
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pingResponse{Message: "pong"})
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
