package handler

import (
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/user/local-webdav/internal/config"

	"golang.org/x/net/webdav"
)

type ShareHandler struct {
	shares   map[string]*shareEntry
	debounce *debounce
	logger   *slog.Logger
}

type shareEntry struct {
	cfg     config.Share
	handler *webdav.Handler
}

func New(shares []config.Share, logger *slog.Logger) *ShareHandler {
	h := &ShareHandler{
		shares:   make(map[string]*shareEntry, len(shares)),
		debounce: newDebounce(),
		logger:   logger,
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
							"remote_addr", r.RemoteAddr,
							"err", err,
						)
					} else {
						logger.Info("webdav operation",
							"method", r.Method,
							"path", r.URL.Path,
							"remote_addr", r.RemoteAddr,
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
		h.logger.Warn("share not found", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	entry, ok := h.shares[shareName]
	if !ok {
		h.logger.Warn("share not found", "share", shareName, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !checkAuth(r, entry.cfg.Username, entry.cfg.Password) {
		h.logger.Warn("unauthorized access", "share", shareName, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
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

	if isMutating(r.Method) {
		keys := debounceKeys(r.Method, shareName, cleanPath, r.Header.Get("Destination"), r.Host)
		release, ok := h.debounce.waitAll(r.Context(), keys)
		if !ok {
			h.logger.Info("debounced duplicate request",
				"method", r.Method,
				"share", shareName,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)
			w.WriteHeader(http.StatusOK)
			return
		}
		defer release()
	}

	entry.handler.ServeHTTP(w, &r2)
}

func isMutating(method string) bool {
	switch method {
	case "PUT", "DELETE", "MKCOL", "MOVE", "COPY", "PROPPATCH":
		return true
	default:
		return false
	}
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

func debounceKeys(method, shareName, cleanPath, destHeader, requestHost string) []string {
	src := shareName + ":" + cleanPath

	if method != "MOVE" && method != "COPY" {
		return []string{src}
	}

	dst := parseDestination(destHeader, requestHost)
	if dst == "" {
		return []string{src}
	}

	if method == "COPY" {
		return []string{dst}
	}

	return []string{src, dst}
}

func parseDestination(header, requestHost string) string {
	if header == "" {
		return ""
	}

	u, err := url.Parse(header)
	if err != nil {
		return ""
	}
	if u.Host != "" && u.Host != requestHost {
		return ""
	}

	share, sub := splitPath(u.Path)
	if share == "" {
		return ""
	}

	destPath := path.Clean("/" + sub)
	if destPath == "." {
		destPath = "/"
	}
	return share + ":" + destPath
}
