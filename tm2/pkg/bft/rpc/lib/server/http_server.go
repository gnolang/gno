// Commons for HTTP handling
package rpcserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/netutil"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Config is a RPC server configuration.
type Config struct {
	// see netutil.LimitListener
	MaxOpenConnections int
	// mirrors http.Server#ReadTimeout
	ReadTimeout time.Duration
	// mirrors http.Server#WriteTimeout
	WriteTimeout time.Duration
	// mirrors http.Server#IdleTimeout: zero falls back to ReadTimeout,
	// per net/http.
	IdleTimeout time.Duration
	// MaxBodyBytes controls the maximum number of bytes the
	// server will read parsing the request body.
	MaxBodyBytes int64
	// mirrors http.Server#MaxHeaderBytes
	MaxHeaderBytes int
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		MaxOpenConnections: 0, // unlimited
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       30 * time.Second,
		IdleTimeout:        0,              // net/http: fall back to ReadTimeout
		MaxBodyBytes:       int64(5000000), // 5MB
		MaxHeaderBytes:     1 << 20,        // same as the net/http default
	}
}

// StartHTTPServer takes a listener and starts an HTTP server with the given handler.
// It wraps handler with RecoverAndLogHandler.
// NOTE: This function blocks - you may want to call it in a go-routine.
func StartHTTPServer(listener net.Listener, handler http.Handler, logger *slog.Logger, config *Config) error {
	logger.Info(fmt.Sprintf("Starting RPC HTTP server on %s", listener.Addr()))
	s := &http.Server{
		Handler:           RecoverAndLogHandler(maxBytesHandler{h: handler, n: config.MaxBodyBytes}, logger),
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}
	err := s.Serve(listener)
	logger.Info("RPC HTTP server stopped", "err", err)
	return err
}

// StartHTTPAndTLSServer takes a listener and starts an HTTPS server with the given handler.
// It wraps handler with RecoverAndLogHandler.
// NOTE: This function blocks - you may want to call it in a go-routine.
func StartHTTPAndTLSServer(
	listener net.Listener,
	handler http.Handler,
	certFile, keyFile string,
	logger *slog.Logger,
	config *Config,
) error {
	logger.Info(fmt.Sprintf("Starting RPC HTTPS server on %s (cert: %q, key: %q)",
		listener.Addr(), certFile, keyFile))

	s := &http.Server{
		Handler:           RecoverAndLogHandler(maxBytesHandler{h: handler, n: config.MaxBodyBytes}, logger),
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}
	err := s.ServeTLS(listener, certFile, keyFile)

	logger.Error("RPC HTTPS server stopped", "err", err)
	return err
}

func WriteRPCResponseHTTPError(
	w http.ResponseWriter,
	httpCode int,
	res types.RPCResponse,
) error {
	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	if _, err := w.Write(jsonBytes); err != nil {
		return err
	}
	return nil
}

func WriteRPCResponseHTTP(w http.ResponseWriter, res types.RPCResponse) error {
	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if _, err := w.Write(jsonBytes); err != nil {
		return err
	}
	return nil
}

// WriteStreamingRPCResponseHTTP writes a successful JSON-RPC response whose
// result body is produced incrementally by result.StreamJSON. The envelope
// (jsonrpc / id) is written around the streamed body without buffering the
// whole result in memory. Returns the first write error encountered (no
// panics), mirroring the contract of WriteRPCResponseHTTP.
//
// ctx is forwarded to result.StreamJSON so that long streams can be aborted
// when the client disconnects.
func WriteStreamingRPCResponseHTTP(ctx context.Context, w http.ResponseWriter, id types.JSONRPCID, result types.StreamableResult) error {
	idBytes, err := json.Marshal(id)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON-RPC id: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	if _, err := fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":`, idBytes); err != nil {
		return err
	}
	if err := result.StreamJSON(ctx, w); err != nil {
		return err
	}
	if _, err := w.Write([]byte(`}`)); err != nil {
		return err
	}
	return nil
}

// WriteRPCResponseArrayHTTP will do the same as WriteRPCResponseHTTP, except it
// can write arrays of responses for batched request/response interactions via
// the JSON RPC.
func WriteRPCResponseArrayHTTP(w http.ResponseWriter, res types.RPCResponses) error {
	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if _, err := w.Write(jsonBytes); err != nil {
		return err
	}
	return nil
}

// -----------------------------------------------------------------------------

// RecoverAndLogHandler wraps an HTTP handler, adding error logging.
// If the inner function panics, the outer function recovers, logs, sends an
// HTTP 500 error response.
func RecoverAndLogHandler(handler http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter to remember the status
		rww := &ResponseWriterWrapper{-1, w}
		begin := time.Now()
		ctx, span := traces.Tracer().Start(r.Context(), "rpcserver")
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.Path),
			attribute.String("remoteAddr", r.RemoteAddr),
		)
		defer span.End()
		logger.Warn("started Span", "span", span.SpanContext().TraceID().String())

		r = r.WithContext(ctx)

		rww.Header().Set("X-Server-Time", fmt.Sprintf("%v", begin.Unix()))

		defer func() {
			// Send a 500 error if a panic happens during a handler.
			// Without this, Chrome & Firefox were retrying aborted ajax requests,
			// at least to my localhost.
			if e := recover(); e != nil {
				switch e := e.(type) {
				case types.RPCResponse:
					if werr := WriteRPCResponseHTTP(rww, e); werr != nil {
						logger.Error("failed to write RPC response", "err", werr)
					}

				case error:
					logger.Error(
						"Panic in RPC HTTP handler", "err", e, "stack",
						string(debug.Stack()),
					)
					if werr := WriteRPCResponseHTTPError(rww, http.StatusInternalServerError,
						types.RPCInternalError(types.JSONRPCStringID(""), e)); werr != nil {
						logger.Error("failed to write RPC response", "err", werr)
					}

				default: // handle string type and any other types
					logger.Error(
						"Panic in RPC HTTP handler", "err", e, "stack",
						string(debug.Stack()),
					)
					if werr := WriteRPCResponseHTTPError(rww, http.StatusInternalServerError,
						types.RPCInternalError(types.JSONRPCStringID(""), fmt.Errorf("%v", e))); werr != nil {
						logger.Error("failed to write RPC response", "err", werr)
					}
				}
			}

			// Finally, log.
			durationMS := time.Since(begin).Nanoseconds() / 1000000
			if rww.Status == -1 {
				rww.Status = 200
			}
			logger.Debug("Served RPC HTTP response",
				"method", r.Method, "url", r.URL,
				"status", rww.Status, "duration", durationMS,
				"remoteAddr", r.RemoteAddr,
			)
		}()

		handler.ServeHTTP(rww, r)
	})
}

// Remember the status for logging
type ResponseWriterWrapper struct {
	Status int
	http.ResponseWriter
}

func (w *ResponseWriterWrapper) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

// implements http.Hijacker
func (w *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

type maxBytesHandler struct {
	h http.Handler
	n int64
}

func (h maxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.n)
	h.h.ServeHTTP(w, r)
}

// Listen starts a new net.Listener on the given address.
// It returns an error if the address is invalid or the call to Listen() fails.
func Listen(addr string, config *Config) (listener net.Listener, err error) {
	parts := strings.SplitN(addr, "://", 2)
	if len(parts) != 2 {
		return nil, errors.New(
			"invalid listening address %s (use fully formed addresses, including the tcp:// or unix:// prefix)",
			addr,
		)
	}
	proto, addr := parts[0], parts[1]
	listener, err = net.Listen(proto, addr)
	if err != nil {
		return nil, errors.New("failed to listen on %v: %v", addr, err)
	}
	if config.MaxOpenConnections > 0 {
		listener = netutil.LimitListener(listener, config.MaxOpenConnections)
	}

	return listener, nil
}
