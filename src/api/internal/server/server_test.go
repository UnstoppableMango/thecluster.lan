package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPing(t *testing.T) {
	handler := New()

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

func TestResolveStaticPathRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	writeStaticFile(t, root, "index.html", "index")

	if _, ok := resolveStaticPath(root, "/../etc/passwd"); ok {
		t.Fatal("expected traversal path to be rejected")
	}
}

func TestStaticHandlerFallsBackToIndex(t *testing.T) {
	root := t.TempDir()
	writeStaticFile(t, root, "index.html", "index")

	handler := New(root)

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if got := recorder.Body.String(); got != "index" {
		t.Fatalf("expected index response, got %q", got)
	}
}

func writeStaticFile(t *testing.T, root string, name string, contents string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", name, err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
