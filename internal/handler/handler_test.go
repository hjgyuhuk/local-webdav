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

func TestDebounceKeys(t *testing.T) {
	tests := []struct {
		method string
		share  string
		path   string
		dest   string
		host   string
		want   []string
	}{
		{"PUT", "docs", "/file.txt", "", "", []string{"docs:/file.txt"}},
		{"DELETE", "docs", "/file.txt", "", "", []string{"docs:/file.txt"}},
		{"MKCOL", "docs", "/dir", "", "", []string{"docs:/dir"}},
		{"MOVE", "docs", "/a.txt", "http://localhost/docs/b.txt", "localhost", []string{"docs:/a.txt", "docs:/b.txt"}},
		{"MOVE", "docs", "/a.txt", "http://localhost/other/b.txt", "localhost", []string{"docs:/a.txt", "other:/b.txt"}},
		{"COPY", "docs", "/a.txt", "http://localhost/docs/b.txt", "localhost", []string{"docs:/b.txt"}},
		{"MOVE", "docs", "/a.txt", "", "", []string{"docs:/a.txt"}},
		{"MOVE", "docs", "/a.txt", "/docs/b.txt", "localhost", []string{"docs:/a.txt", "docs:/b.txt"}},
		{"COPY", "docs", "/a.txt", "/other/b.txt", "localhost", []string{"other:/b.txt"}},
		{"MOVE", "docs", "/a.txt", "http://other-host/docs/b.txt", "localhost", []string{"docs:/a.txt"}},
	}

	for _, tt := range tests {
		got := debounceKeys(tt.method, tt.share, tt.path, tt.dest, tt.host)
		if len(got) != len(tt.want) {
			t.Errorf("debounceKeys(%q, %q, %q, %q, %q) = %v, want %v",
				tt.method, tt.share, tt.path, tt.dest, tt.host, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("debounceKeys(%q, %q, %q, %q, %q)[%d] = %q, want %q",
					tt.method, tt.share, tt.path, tt.dest, tt.host, i, got[i], tt.want[i])
			}
		}
	}
}

func TestDebounceKeys_MoveSameDestination(t *testing.T) {
	k1 := debounceKeys("MOVE", "docs", "/a.txt", "http://localhost/docs/x.txt", "localhost")
	k2 := debounceKeys("MOVE", "docs", "/b.txt", "http://localhost/docs/x.txt", "localhost")

	hasDst := func(keys []string) bool {
		for _, k := range keys {
			if k == "docs:/x.txt" {
				return true
			}
		}
		return false
	}

	if !hasDst(k1) || !hasDst(k2) {
		t.Fatalf("both moves must lock destination: k1=%v k2=%v", k1, k2)
	}
}

func TestParseDestination(t *testing.T) {
	tests := []struct {
		header string
		host   string
		want   string
	}{
		{"http://localhost/docs/b.txt", "localhost", "docs:/b.txt"},
		{"http://localhost/other/deep/file.txt", "localhost", "other:/deep/file.txt"},
		{"", "", ""},
		{"bad-uri%", "", ""},
		{"http://localhost/docs/", "localhost", "docs:/"},
		{"/docs/b.txt", "localhost", "docs:/b.txt"},
		{"/other/deep/file.txt", "localhost", "other:/deep/file.txt"},
		{"/docs/b.txt", "", "docs:/b.txt"},
		{"http://other-host/docs/b.txt", "localhost", ""},
	}

	for _, tt := range tests {
		got := parseDestination(tt.header, tt.host)
		if got != tt.want {
			t.Errorf("parseDestination(%q, %q) = %q, want %q", tt.header, tt.host, got, tt.want)
		}
	}
}
