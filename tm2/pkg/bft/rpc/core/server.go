package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/config"
)

type Server struct {
	h      http.Handler
	logger *slog.Logger
	config *config.Config

	srv      *http.Server
	ln       net.Listener
	errCh    chan error
	stopOnce sync.Once
}

func New(h http.Handler, config *config.Config, logger *slog.Logger) *Server {
	return &Server{
		h:      h,
		config: config,
		logger: logger,
		errCh:  make(chan error, 1),
	}
}

// Start starts the server asynchronously
func (s *Server) Start() error {
	s.srv = &http.Server{
		Addr:              s.config.ListenAddress,
		Handler:           s.h,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      60 * time.Second,
	}

	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf(
			"unable to listen on address %s: %w",
			s.srv.Addr,
			err,
		)
	}

	s.ln = ln

	s.logger.Info(
		"RPC server started",
		"address", ln.Addr().String(),
	)

	// Start serving async
	go func() {
		if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.errCh <- err
		}

		close(s.errCh)
	}()

	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	var shutdownErr error

	s.stopOnce.Do(func() {
		s.logger.Info("RPC server shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if s.srv != nil {
			shutdownErr = s.srv.Shutdown(ctx)
		}

		s.logger.Info("RPC server shut down")
	})

	return shutdownErr
}
