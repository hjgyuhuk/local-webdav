package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/user/local-webdav/internal/config"
)

func newTestHandler(t *testing.T) (*ShareHandler, string) {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(dir+"/test.txt", []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	shares := []config.Share{
		{Name: "docs", Path: dir, Username: "user", Password: "pass"},
	}
	return New(shares, logger), dir
}

func TestServeHTTP_GetFile(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/docs/test.txt", nil)
	req.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "hello" {
		t.Errorf("body = %q, want %q", w.Body.String(), "hello")
	}
}

func TestServeHTTP_NoAuth(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/docs/test.txt", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestServeHTTP_WrongAuth(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/docs/test.txt", nil)
	req.SetBasicAuth("user", "wrong")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestServeHTTP_UnknownShare(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/unknown/test.txt", nil)
	req.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_EmptyPath(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_PathTraversal(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/docs/../../../etc/passwd", nil)
	req.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (path traversal should not expose files outside share)", w.Code, http.StatusNotFound)
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input     string
		wantShare string
		wantSub   string
	}{
		{"/docs/file.txt", "docs", "file.txt"},
		{"/docs/", "docs", ""},
		{"/docs", "docs", ""},
		{"/", "", ""},
		{"", "", ""},
		{"/docs/sub/dir/file.txt", "docs", "sub/dir/file.txt"},
	}

	for _, tt := range tests {
		share, sub := splitPath(tt.input)
		if share != tt.wantShare || sub != tt.wantSub {
			t.Errorf("splitPath(%q) = (%q, %q), want (%q, %q)",
				tt.input, share, sub, tt.wantShare, tt.wantSub)
		}
	}
}

func TestServeHTTP_NoCredentialsAllowed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/open.txt", []byte("open"), 0644); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	shares := []config.Share{
		{Name: "public", Path: dir},
	}
	h := New(shares, logger)

	req := httptest.NewRequest("GET", "/public/open.txt", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "open" {
		t.Errorf("body = %q, want %q", w.Body.String(), "open")
	}
}
