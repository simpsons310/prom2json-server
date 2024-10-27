package p2jsvr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	ForceKillServerTimeout = 5 * time.Second
)

type ServerConfig struct {
	Port int `yaml:"port"`
}

type Server struct {
	addr   string
	port   int
	config *ServerConfig
}

func NewServer(cfg *ServerConfig) *Server {
	return &Server{
		addr:   fmt.Sprintf(":%d", cfg.Port),
		port:   cfg.Port,
		config: cfg,
	}
}

func (s *Server) Start(ctx context.Context, handler http.Handler) error {
	svr := &http.Server{
		Addr:    s.addr,
		Handler: handler,
	}
	logger := LoggerFromContext(ctx)

	// Handle grateful shutdown
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		logger.Info().Msg("Context closed, shutting down server")
		shutdownCtx, done := context.WithTimeout(context.Background(), ForceKillServerTimeout)
		defer done()
		errCh <- svr.Shutdown(shutdownCtx)
	}()

	// If server serve failed with error other than http.ErrServerClosed, return the error
	// Otherwise, wait for the server to shutdown gracefully
	logger.Info().Str("address", s.addr).Msgf("Starting HTTP server")
	if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	logger.Info().Msg("Server stopped serving, waiting for graceful shutdown")

	// Wait for the server to shutdown gracefully & return error if any
	if err := <-errCh; err != nil {
		return err
	}
	return nil
}
