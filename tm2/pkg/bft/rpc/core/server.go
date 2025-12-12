package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/config"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	h      http.Handler
	logger *slog.Logger

	config *config.Config
}

func New(h http.Handler, config *config.Config, logger *slog.Logger) *Server {
	return &Server{
		h:      h,
		config: config,
		logger: logger,
	}
}

// Serve serves the JSON-RPC server
func (s *Server) Serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.config.ListenAddress,
		Handler:           s.h,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      60 * time.Second,
	}

	group, gCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		defer s.logger.Info("RPC server shut down")

		ln, err := net.Listen("tcp", srv.Addr)
		if err != nil {
			return err
		}

		s.logger.Info(
			"RPC server started",
			"address", ln.Addr().String(),
		)

		if err = srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	group.Go(func() error {
		<-gCtx.Done()

		s.logger.Info("RPC server to be shut down")

		wsCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		return srv.Shutdown(wsCtx)
	})

	return group.Wait()
}
