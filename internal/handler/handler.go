package handler

import (
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/user/local-webdav/internal/config"

	"golang.org/x/net/webdav"
)

type ShareHandler struct {
	shares map[string]*shareEntry
	logger *slog.Logger
}

type shareEntry struct {
	cfg     config.Share
	handler *webdav.Handler
}

func New(shares []config.Share, logger *slog.Logger) *ShareHandler {
	h := &ShareHandler{
		shares: make(map[string]*shareEntry, len(shares)),
		logger: logger,
	}

	for _, s := range shares {
		h.shares[s.Name] = &shareEntry{
			cfg: s,
			handler: &webdav.Handler{
				FileSystem: webdav.Dir(s.Path),
				LockSystem: webdav.NewMemLS(),
				Logger: func(r *http.Request, err error) {
					if err != nil {
						logger.Error("webdav error",
							"method", r.Method,
							"path", r.URL.Path,
							"err", err,
						)
					}
				},
			},
		}
	}

	return h
}

func (h *ShareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shareName, subPath := splitPath(r.URL.Path)

	if shareName == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	entry, ok := h.shares[shareName]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !checkAuth(r, entry.cfg.Username, entry.cfg.Password) {
		w.Header().Set("WWW-Authenticate", `Basic realm="WebDAV"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cleanPath := path.Clean("/" + subPath)
	if cleanPath == "." {
		cleanPath = "/"
	}

	r2 := *r
	r2.URL.Path = cleanPath

	entry.handler.ServeHTTP(w, &r2)
}

func splitPath(urlPath string) (shareName, subPath string) {
	cleaned := path.Clean(urlPath)
	cleaned = strings.TrimPrefix(cleaned, "/")

	if cleaned == "" || cleaned == "." {
		return "", ""
	}

	parts := strings.SplitN(cleaned, "/", 2)
	shareName = parts[0]
	if len(parts) > 1 {
		subPath = parts[1]
	}

	return shareName, subPath
}

func checkAuth(r *http.Request, username, password string) bool {
	if username == "" && password == "" {
		return true
	}
	u, p, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return u == username && p == password
}
