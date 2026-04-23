package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	handler, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Message != "pong" {
		t.Fatalf("expected pong, got %q", response.Message)
	}
}

func TestStaticIndex(t *testing.T) {
	dist := makeTestDist(t)
	handler, err := New(dist)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNotFound(t *testing.T) {
	dist := makeTestDist(t)
	handler, err := New(dist)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-such-page", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "404") {
		t.Fatalf("expected 404 page content, got: %s", rec.Body.String())
	}
}

func TestPathTraversal(t *testing.T) {
	dist := makeTestDist(t)
	handler, err := New(dist)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cases := []string{
		"/../etc/passwd",
		"/assets/../../etc/passwd",
		"/%2e%2e/etc/passwd",
	}
	for _, path := range cases {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code == http.StatusOK {
			t.Errorf("path %q: expected non-200, got 200", path)
		}
	}
}

func TestAssetServing(t *testing.T) {
	dist := makeTestDist(t)
	handler, err := New(dist)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/main.js", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for asset, got %d", rec.Code)
	}
}

// makeTestDist creates a minimal Vite dist directory for testing.
func makeTestDist(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "index.html"), "<html><body>app</body></html>")

	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	writeFile(t, filepath.Join(assetsDir, "main.js"), "console.log('hello')")

	viteDir := filepath.Join(dir, ".vite")
	if err := os.MkdirAll(viteDir, 0755); err != nil {
		t.Fatalf("mkdir .vite: %v", err)
	}
	manifest := `{"src/main.ts":{"file":"assets/main.js","isEntry":true,"src":"src/main.ts"}}`
	writeFile(t, filepath.Join(viteDir, "manifest.json"), manifest)

	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
