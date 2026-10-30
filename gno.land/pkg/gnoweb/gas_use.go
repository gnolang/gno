package gnoweb

import (
	"context"
	"errors"
	"go/token"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
	"golang.org/x/sync/singleflight"
)

const (
	// gasUseTxs is how many of a package's recent transactions are read; one
	// read covers every function of the package.
	gasUseTxs = 100
	// gasUseCalls is how many recent calls of a function are measured.
	gasUseCalls = 20
	// gasUseTTL is how long a package's gas use is served before it is read
	// again. It changes as the realm's state grows or with a chain upgrade,
	// not per block.
	gasUseTTL = 5 * time.Minute
	// gasUseRetry is how long a failed read is kept: the indexer client's
	// breaker reopens after 30s.
	gasUseRetry = 30 * time.Second
	// gasUseTimeout bounds the read, long enough for the indexer to walk
	// back past a quiet stretch of blocks.
	gasUseTimeout = 10 * time.Second
	// maxGasUseReads caps the reads in flight. Past it a page skips its read
	// and a later page retries.
	maxGasUseReads = 4
	// maxGasUsePkgs bounds the cache; a full one is cleared.
	maxGasUsePkgs = 1024
	// maxGasUseFunc is the longest function name kept from the indexer.
	maxGasUseFunc = 64
)

// gasUses returns, per function of pkgPath, the most gas any of its last
// gasUseCalls successful calls used. txs are newest first. Only a
// transaction made of a single call is measured, and each is checked again:
// the indexer's filter has let unrelated transactions through (#6411).
func gasUses(txs []indexer.Tx, pkgPath string) map[string]components.GasUse {
	uses := make(map[string]components.GasUse)
	for _, tx := range txs {
		if !tx.Success || tx.GasUsed <= 0 || len(tx.Messages) != 1 {
			continue
		}
		m := tx.Messages[0].Value
		if m.Type != "MsgCall" || m.PkgPath != pkgPath ||
			len(m.Func) > maxGasUseFunc || !token.IsIdentifier(m.Func) {
			continue
		}
		u := uses[m.Func]
		if u.Calls < gasUseCalls {
			u.Calls++
			u.Max = max(u.Max, int64(tx.GasUsed))
			uses[m.Func] = u
		}
	}
	return uses
}

// gasUseCache keeps each package's gas use for gasUseTTL. A page never waits
// on a read: it gets what is cached, nil at first, while one background read
// per package fills the cache for the next page.
type gasUseCache struct {
	// recent reads a package's recent transactions; nil without an indexer.
	recent func(ctx context.Context, pkgPath string, limit int) ([]indexer.Tx, error)
	logger *slog.Logger

	reads   atomic.Int32 // reads in flight
	sf      singleflight.Group
	mu      sync.Mutex
	entries map[string]gasUseEntry
	now     func() time.Time // for tests; time.Now when nil
}

type gasUseEntry struct {
	uses    map[string]components.GasUse // nil when the read failed
	expires time.Time
}

// get returns pkgPath's gas use by function, or nil when there is none yet,
// and starts a read when the cached one is missing or expired.
func (c *gasUseCache) get(ctx context.Context, pkgPath string) map[string]components.GasUse {
	if c.recent == nil {
		return nil
	}
	e, ok := c.lookup(pkgPath)
	if !ok {
		// Detached: the read outlives the page and serves the ones after it.
		ctx := context.WithoutCancel(ctx)
		c.sf.DoChan(pkgPath, func() (any, error) {
			c.read(ctx, pkgPath)
			return nil, nil
		})
	}
	return e.uses
}

// read reads pkgPath's gas use into the cache, unless maxGasUseReads are in
// flight or another read has just stored it.
func (c *gasUseCache) read(ctx context.Context, pkgPath string) {
	if _, ok := c.lookup(pkgPath); ok {
		return
	}
	if c.reads.Add(1) > maxGasUseReads {
		c.reads.Add(-1)
		return
	}
	defer c.reads.Add(-1)

	ctx, cancel := context.WithTimeout(ctx, gasUseTimeout)
	defer cancel()
	txs, err := c.recent(ctx, pkgPath, gasUseTxs)
	var e gasUseEntry
	if err == nil || errors.Is(err, indexer.ErrPartial) { // the newest rows, which is what is measured
		e = gasUseEntry{uses: gasUses(txs, pkgPath), expires: c.clock().Add(gasUseTTL)}
	} else {
		c.logger.Debug("gas use unavailable", "pkgpath", pkgPath, "error", err)
		e.expires = c.clock().Add(gasUseRetry)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = make(map[string]gasUseEntry)
	}
	if _, ok := c.entries[pkgPath]; !ok && len(c.entries) >= maxGasUsePkgs {
		clear(c.entries)
	}
	c.entries[pkgPath] = e
}

// lookup returns pkgPath's entry, and whether it is still fresh.
func (c *gasUseCache) lookup(pkgPath string) (gasUseEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[pkgPath]
	return e, ok && c.clock().Before(e.expires)
}

func (c *gasUseCache) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}
