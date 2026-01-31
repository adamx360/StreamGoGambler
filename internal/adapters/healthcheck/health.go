package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"streamgogambler/internal/adapters/logging"
	"streamgogambler/internal/ports"
)

type HealthServer struct {
	port     int
	provider ports.StatsProvider
	logger   *logging.Logger
	server   *http.Server
}

func NewHealthServer(port int, provider ports.StatsProvider, logger *logging.Logger) *HealthServer {
	return &HealthServer{
		port:     port,
		provider: provider,
		logger:   logger,
	}
}

func (s *HealthServer) Start(ctx context.Context) error {
	if s.port <= 0 {
		return nil // Disabled
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		s.logger.Infof(ctx, "Health server started on port %d", s.port)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Errorf(ctx, "Health server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Errorf(shutdownCtx, "Health server shutdown error: %v", err)
		}
	}()

	return nil
}

func (s *HealthServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	stats := s.provider.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.Errorf(r.Context(), "JSON encoding error in /health: %v", err)
	}
}

func (s *HealthServer) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
