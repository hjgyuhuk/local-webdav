package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/user/local-webdav/internal/config"
	"github.com/user/local-webdav/internal/handler"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *Server {
	h := handler.New(cfg.Shares, logger)

	mux := http.NewServeMux()
	mux.Handle("/", h)

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Server.Address,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("starting webdav server", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down webdav server")
	return s.httpServer.Shutdown(ctx)
}
