package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/cockroachdb/pebble"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gnolang/gno/contribs/gnodev/pkg/logger"
	"github.com/gnolang/tx-indexer/client"
	"github.com/gnolang/tx-indexer/events"
	"github.com/gnolang/tx-indexer/fetch"
	"github.com/gnolang/tx-indexer/serve"
	"github.com/gnolang/tx-indexer/serve/graph"
	"github.com/gnolang/tx-indexer/serve/health"
	"github.com/gnolang/tx-indexer/storage"
)

const (
	defaultRemote = "http://127.0.0.1:26657"
	defaultDBPath = "indexer-db"
)

const (
	JSONRPCName    = "json-rpc"
	HTTPServerName = "http-server"
	FetcherName    = "fetcher"
)

type FmtLoggerMessage struct {}

func getLoggerColor(name string) int {
	switch name {
	case JSONRPCName:
		return 4
	case HTTPServerName:
		return 5
	case FetcherName:
		return 6
	default:
		return 7
	}
}

func (_ FmtLoggerMessage) Fmt(message, name string) string {
	loggerName := fmt.Sprintf("\033[1;3%dm%s\033[0m", getLoggerColor(name), strings.ToUpper(name))
	return fmt.Sprintf("%s %s", loggerName, message)
}

type IndexerSettings struct {
	listenAddress string
	remote        string
	dbPath        string
	logLevel      zapcore.Level
	maxSlots      int
	maxChunkSize  int64
	rateLimit     int
}

func DefaultIndexerSettings() IndexerSettings {
	return IndexerSettings{
		listenAddress: serve.DefaultListenAddress,
		remote:        defaultRemote,
		dbPath:        defaultDBPath,
		logLevel:      zap.DebugLevel,
		maxSlots:      fetch.DefaultMaxSlots,
		maxChunkSize:  fetch.DefaultMaxChunkSize,
		rateLimit:     0,
	}
}

type Indexer struct {
	db       *storage.Pebble
	logger   *zap.Logger
	fetcher  *fetch.Fetcher
	server   *serve.HTTPServer
	settings IndexerSettings
	ctx      context.Context
	cancel   context.CancelFunc
	errCh    chan error
}

func (i Indexer) closeDB() error {
	if closeErr := i.db.Close(); closeErr != nil {
		return fmt.Errorf("unable to gracefully close DB, %s", closeErr.Error())
	}

	if deleteErr := os.RemoveAll(i.settings.dbPath); deleteErr != nil {
		return fmt.Errorf("unable to gracefully delete DB, %s", deleteErr.Error())
	}

	return nil
}

func (i *Indexer) run() {
	go func() {
		if err := i.fetcher.FetchChainData(i.ctx); err != nil {
			i.errCh <- err
		}
	}()

	go func() {
		if err := i.server.Serve(i.ctx); err != nil {
			i.errCh <- err
		}
	}()
}

func (i *Indexer) Start(ctx context.Context, wg *sync.WaitGroup) error {
	// used to wait for clean up work
	wg.Add(1)
	defer wg.Done()

	i.run()

	for {
		select {
		case <-ctx.Done():
			return i.stop()
		case <-i.errCh:
			return <-i.errCh
		}
	}
}

func (i *Indexer) Reload() error {
	i.cancel()

	if err := i.closeDB(); err != nil {
		return err
	}

	if err := i.setNewStorageDB(); err != nil {
		return err
	}

	i.ctx, i.cancel = context.WithCancel(context.Background())
	i.run()

	return nil
}

func (i *Indexer) stop() error {
	if err := i.closeDB(); err != nil {
		return err
	}

	i.cancel()
	return nil
}

func (i *Indexer) setNewStorageDB() (err error) {
	db, err := pebble.Open(i.settings.dbPath, nil)
	if err != nil {
		return fmt.Errorf("unable to create DB, %w", err)
	}

    defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred while setting new storage DB: %v", r)
		}
	}()

	// other option to solve this is to recreate all services that uses db
	val := reflect.ValueOf(i.db).Elem()
	field := val.FieldByName("db")
	field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()

	field.Set(reflect.ValueOf(db))
	return 
}

func NewIndexer(slogger *slog.Logger, settings IndexerSettings) (*Indexer, error) {
	logger := logger.NewZapLoggerWithSlog(slogger, FmtLoggerMessage{})

	// Create a DB instance
	db, err := storage.NewPebble(settings.dbPath)

	if err != nil {
		return nil, fmt.Errorf("unable to open storage DB, %w", err)
	}

	// Create an Event Manager instance
	em := events.NewManager()

	// Create a TM2 client
	tm2Client, err := client.NewClient(settings.remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create client, %w", err)
	}

	// Create the fetcher service
	f := fetch.New(
		db,
		tm2Client,
		em,
		fetch.WithLogger(
			logger.Named(FetcherName),
		),
		fetch.WithMaxSlots(settings.maxSlots),
		fetch.WithMaxChunkSize(settings.maxChunkSize),
	)

	// Create the JSON-RPC service
	j := setupJSONRPC(
		db,
		em,
		logger,
	)

	mux := chi.NewMux()

	if settings.rateLimit != 0 {
		logger.Info("rate-limit set", zap.Int("rate-limit", settings.rateLimit))
		mux.Use(httprate.Limit(
			settings.rateLimit,
			time.Minute,
			httprate.WithKeyFuncs(httprate.KeyByRealIP),
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				//nolint:errcheck // no need to handle error here, it had been checked before
				ip, _ := httprate.KeyByRealIP(r)
				logger.Debug("too many requests", zap.String("from", ip))

				// send a json response to give more info when using the graphQL explorer
				http.Error(w, `{"error": "too many requests"}`, http.StatusTooManyRequests)
			}),
		))
	}

	mux = j.SetupRoutes(mux)
	mux = graph.Setup(db, em, mux)
	mux = health.Setup(db, mux)

	// Create the HTTP server
	hs := serve.NewHTTPServer(mux, settings.listenAddress, logger.Named(HTTPServerName))

	ctx, cancel := context.WithCancel(context.Background())

	indexer := &Indexer{
		db:       db,
		logger:   logger,
		settings: settings,
		fetcher:  f,
		server:   hs,
		ctx:      ctx,
		cancel:   cancel,
		errCh:    make(chan error, 2),
	}

	return indexer, nil
}

// setupJSONRPC sets up the JSONRPC instance
func setupJSONRPC(
	db *storage.Pebble,
	em *events.Manager,
	logger *zap.Logger,
) *serve.JSONRPC {
	j := serve.NewJSONRPC(
		em,
		serve.WithLogger(
			logger.Named(JSONRPCName),
		),
	)

	// Transaction handlers
	j.RegisterTxEndpoints(db)

	// Block handlers
	j.RegisterBlockEndpoints(db)

	// Sub handlers
	j.RegisterSubEndpoints(db)

	return j
}

