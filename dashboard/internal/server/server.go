package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/HomericIntelligence/atlas/internal/config"
	"github.com/HomericIntelligence/atlas/internal/events"
	"github.com/HomericIntelligence/atlas/internal/handlers"
)

type Server struct {
	cfg *config.Config
	srv *http.Server
	bus *events.Bus
	sse *handlers.SSE
}

func New(cfg *config.Config, bus *events.Bus) *Server {
	s := &Server{
		cfg: cfg,
		bus: bus,
		sse: handlers.NewSSE(bus),
	}
	s.srv = &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      s.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // SSE connections are long-lived; disable write timeout.
		IdleTimeout:  60 * time.Second,
	}
	return s
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.srv.Shutdown(shutCtx)
	}
}
