package txindexer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var _ Manager = (*Service)(nil)

// Service provides functionality to start, stop, and reload the tx-indexer
// process using exec.Cmd constructs. The methods are meant to be
// idempotent and should be safe to call multiple times without causing
// side effects.
type Service struct {
	logger  *slog.Logger
	dbPath  string
	listen  string
	process *process
}

// NewService returns an instantiated tx-indexer service or an error
// due to missing dependencies. Please refer to the README for an
// example usage of the tx-indexer service.
func NewService(logger *slog.Logger, cfg Config) (*Service, error) {
	s := Service{
		logger:  logger,
		dbPath:  cfg.DBPath,
		listen:  cfg.ListenAddress,
		process: newProcess(logger.WithGroup("process"), cfg),
	}
	if err := s.validate(); err != nil {
		return nil, err
	}

	return &s, nil
}

// Start starts the tx-indexer process.
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("starting tx-indexer", "db_path", s.dbPath, "listen", s.listen)

	if err := s.process.start(ctx); err != nil {
		const msg = "failed to start tx-indexer"
		s.logger.Error(msg, "error", err)
		return fmt.Errorf(msg+": %w", err)
	}

	return nil
}

// Reload performs three steps to reload the tx-indexer:
// 1. Stop the tx-indexer process
// 2. Remove the tx-indexer database
// 3. Start the tx-indexer process again
func (s *Service) Reload(ctx context.Context) error {
	s.logger.Info("reloading tx-indexer")

	if err := s.Stop(ctx); err != nil {
		return err
	}

	if err := os.RemoveAll(s.dbPath); err != nil {
		const msg = "failed to remove tx-indexer database"
		s.logger.Error(msg, "error", err)
		return fmt.Errorf(msg+": %w", err)
	}
	s.logger.Info("tx-indexer database removed", "db_path", s.dbPath)

	if err := s.Start(ctx); err != nil {
		return err
	}

	return nil
}

// Stop stops the tx-indexer process.
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("stopping tx-indexer")

	if err := s.process.stop(ctx); err != nil {
		const msg = "failed to stop tx-indexer"
		s.logger.Error(msg, "error", err)
		return fmt.Errorf(msg+": %w", err)
	}

	return nil
}

func (s *Service) validate() error {
	var missing []string

	for dep, chk := range map[string]func() bool{
		"process": func() bool { return s.process != nil },
		"dbPath":  func() bool { return s.dbPath != "" },
		"logger":  func() bool { return s.logger != nil },
	} {
		if !chk() {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("unable to initialize service due to missing dependencies: %s", strings.Join(missing, ", "))
	}

	return nil
}
