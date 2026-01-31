package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	gh "github.com/BuildGuard-Test-Lab/pbom/internal/github"
)

// Config holds webhook server configuration.
type Config struct {
	Addr          string
	WebhookSecret string
	GitHubToken   string
	StorageDir    string
}

// Server is the webhook HTTP server.
type Server struct {
	cfg      Config
	ghClient *gh.Client
	enricher *Enricher
	logger   *slog.Logger
	mux      *http.ServeMux

	eventsProcessed atomic.Int64
	lastEventAt     atomic.Value // time.Time
}

// NewServer creates a configured webhook server.
func NewServer(cfg Config, logger *slog.Logger) *Server {
	ghClient := gh.NewClient(cfg.GitHubToken)
	enricher := NewEnricher(ghClient, cfg.StorageDir, logger)

	s := &Server{
		cfg:      cfg,
		ghClient: ghClient,
		enricher: enricher,
		logger:   logger,
		mux:      http.NewServeMux(),
	}

	s.mux.HandleFunc("/webhook", s.handleWebhook)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/status", s.handleStatus)

	return s
}

// Start begins listening for webhook events. Blocks until context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:         s.cfg.Addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("webhook listener starting",
			"addr", s.cfg.Addr,
			"storage_dir", s.cfg.StorageDir,
		)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.logger.Info("shutting down webhook listener")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"events_processed": s.eventsProcessed.Load(),
	}
	if t, ok := s.lastEventAt.Load().(time.Time); ok {
		status["last_event_at"] = t.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
