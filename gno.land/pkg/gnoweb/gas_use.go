package gnoweb

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

const (
	// gasUseTxs is how many of a package's recent transactions are read; one
	// read covers every function of the package.
	gasUseTxs = 100
	// gasUseCalls is how many recent calls of a function are measured.
	gasUseCalls = 20
	// gasUseTTL is how long a package's gas use is served before it is read
	// again. Gas use changes with a redeploy or a chain upgrade, not per block.
	gasUseTTL = 5 * time.Minute
	// gasUseWait bounds how long a page waits for a package's first read,
	// like gasPriceTimeout for the price. The read itself may run longer and
	// serves the next page.
	gasUseWait = gasPriceTimeout
	// gasUseTimeout bounds the read, long enough for the indexer to walk
	// back past a quiet stretch of blocks.
	gasUseTimeout = 10 * time.Second
	// maxGasUsePkgs bounds the cache.
	maxGasUsePkgs = 1024
)

// gasUses returns, per function of pkgPath, the most gas any of its last
// gasUseCalls successful calls used. txs are newest first. Only a
// transaction made of a single call is measured, and each is checked again:
// the indexer's filter has let unrelated transactions through (#6411).
func gasUses(txs []indexer.Tx, pkgPath string) map[string]components.GasUse {
	uses := make(map[string]components.GasUse)
	for _, tx := range txs {
		if !tx.Success || len(tx.Messages) != 1 {
			continue
		}
		m := tx.Messages[0].Value
		if m.Type != "MsgCall" || m.PkgPath != pkgPath || m.Func == "" {
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

// gasUseCache keeps each package's gas use for gasUseTTL, then serves it
// stale while one background read refreshes it. A failed read keeps the
// last answer and is retried after gasUseTTL.
type gasUseCache struct {
	// recent reads a package's recent transactions; nil without an indexer.
	recent func(ctx context.Context, pkgPath string, limit int) ([]indexer.Tx, error)
	logger *slog.Logger

	mu      sync.Mutex
	entries map[string]*gasUseEntry
}

type gasUseEntry struct {
	uses  map[string]components.GasUse // replaced, never modified
	at    time.Time                    // when the last read started
	busy  bool                         // a read is in flight
	ready chan struct{}                // closed when the first read ends
}

// get returns pkgPath's gas use by function, or nil when there is none yet.
// Until the first read ends, callers wait for it until gasUseWait after it
// started, so a page with several commands waits once.
func (c *gasUseCache) get(ctx context.Context, pkgPath string) map[string]components.GasUse {
	if c.recent == nil {
		return nil
	}
	c.mu.Lock()
	e := c.entries[pkgPath]
	if e == nil {
		if c.entries == nil {
			c.entries = make(map[string]*gasUseEntry)
		}
		for k := range c.entries {
			if len(c.entries) < maxGasUsePkgs {
				break
			}
			delete(c.entries, k) // any one: the bound matters, not which
		}
		e = &gasUseEntry{ready: make(chan struct{})}
		c.entries[pkgPath] = e
	}
	if !e.busy && time.Since(e.at) >= gasUseTTL {
		e.busy, e.at = true, time.Now()
		go c.read(ctx, e, pkgPath)
	}
	wait := time.Until(e.at.Add(gasUseWait))
	c.mu.Unlock()

	if wait > 0 {
		t := time.NewTimer(wait)
		select {
		case <-e.ready:
		case <-t.C:
		case <-ctx.Done():
		}
		t.Stop()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return e.uses
}

// read refreshes e. It is detached from the caller's cancellation: its
// answer serves every page that reads the package after it.
func (c *gasUseCache) read(ctx context.Context, e *gasUseEntry, pkgPath string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), gasUseTimeout)
	defer cancel()
	txs, err := c.recent(ctx, pkgPath, gasUseTxs)
	if errors.Is(err, indexer.ErrPartial) {
		err = nil // the newest rows, which is what is measured
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.logger.Debug("gas use unavailable", "pkgpath", pkgPath, "error", err)
	} else {
		e.uses = gasUses(txs, pkgPath)
	}
	e.busy = false
	select {
	case <-e.ready:
	default:
		close(e.ready)
	}
}
