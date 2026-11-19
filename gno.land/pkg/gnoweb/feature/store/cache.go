package store

import (
	"log/slog"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/feature/internal/flight"
)

const (
	// cacheTTL bounds how stale a store page can be. The realm's shelves
	// move by block height, so seconds of staleness are invisible.
	cacheTTL = 30 * time.Second

	// errorTTL remembers a failure briefly, so a store realm that is down
	// costs one timeout per key every few seconds, not one per reader.
	errorTTL = 5 * time.Second

	// maxPages bounds the category and list pages kept; the front page and
	// the build lens have flights of their own, so a crawler evicts neither.
	maxPages = 512
)

// responses memoises the decoded, sanitised answers, and failures, per
// endpoint: a burst of readers costs one RPC and one decode. Answers are
// shared: callers must treat them as read-only.
type responses struct {
	home  *flight.Flight[*homeResponse]
	build *flight.Flight[*buildResponse]
	pages *flight.Flight[*pageResponse]
}

func newResponses(logger *slog.Logger) responses {
	return responses{
		home:  newFlight[*homeResponse](1, logger),
		build: newFlight[*buildResponse](1, logger),
		pages: newFlight[*pageResponse](maxPages, logger),
	}
}

func newFlight[T any](size int, logger *slog.Logger) *flight.Flight[T] {
	return &flight.Flight[T]{TTL: cacheTTL, ErrTTL: errorTTL, Timeout: fetchTimeout, Max: size, Logger: logger}
}
