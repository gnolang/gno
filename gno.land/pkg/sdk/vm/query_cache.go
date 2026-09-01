package vm

import (
	"sync"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

// Bounds on what the query cache will hold. A cached entry is only worth
// keeping while it is cheaper than recomputing it, and a query result can be
// as large as the export size guard allows, so both the per-entry size and the
// total are capped. Reaching either stops new entries being added for the rest
// of the height; nothing is evicted early, because the whole map is dropped as
// soon as the chain advances anyway.
const (
	maxQueryCacheEntries = 1024
	maxQueryCacheBytes   = 32 << 20 // 32 MB total
	maxQueryCacheEntry   = 1 << 20  // 1 MB per entry
)

// queryCache memoises read-only VM query results.
//
// Two things decide a query's answer: the state it reads, chosen by
// req.Height, and the context it runs in, which handleQueryCustom builds from
// the *latest* block header regardless of req.Height. So an entry is only
// valid while the latest height is unchanged, and entries are keyed by the
// height they asked for.
//
// The whole map is therefore dropped when the chain advances, rather than
// entries being invalidated one by one. That is what makes this safe to reason
// about: a hit can only ever be returned for the same chain tip that produced
// it.
type queryCache struct {
	mu sync.Mutex

	// tipHeight is the latest block height the entries were produced under.
	// Entries are dropped when it moves.
	tipHeight int64
	entries   map[string]abci.ResponseQuery
	bytes     int
}

func newQueryCache() *queryCache {
	return &queryCache{entries: make(map[string]abci.ResponseQuery)}
}

// key identifies a query by everything that can change its answer: which
// endpoint, the argument, and which state version it reads. The tip is not in
// the key because it clears the map instead.
//
// Both variable-length parts are length-prefixed rather than separated by a
// byte. A separator would be ambiguous, since the path is caller-supplied and
// may contain that byte: with a NUL separator, path "vm/q" with data "\x00x"
// and path "vm/q\x00" with data "x" produce the same key, and one query would
// be answered with another's result.
func queryCacheKey(req abci.RequestQuery) string {
	var b []byte
	putUint64 := func(v uint64) {
		for i := 56; i >= 0; i -= 8 {
			b = append(b, byte(v>>uint(i)))
		}
	}
	putUint64(uint64(req.Height))
	putUint64(uint64(len(req.Path)))
	b = append(b, req.Path...)
	putUint64(uint64(len(req.Data)))
	b = append(b, req.Data...)
	return string(b)
}

// get returns a cached response for req, if one was produced under the same
// chain tip. A changed tip drops everything before answering.
func (c *queryCache) get(tipHeight int64, req abci.RequestQuery) (abci.ResponseQuery, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tipHeight != tipHeight {
		c.reset(tipHeight)
		return abci.ResponseQuery{}, false
	}
	res, ok := c.entries[queryCacheKey(req)]
	if !ok {
		return abci.ResponseQuery{}, false
	}
	return cloneQueryResponse(res), true
}

// put records a response. It is a no-op once either bound is reached, and for
// responses too large to be worth holding.
func (c *queryCache) put(tipHeight int64, req abci.RequestQuery, res abci.ResponseQuery) {
	if len(res.Data) > maxQueryCacheEntry {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tipHeight != tipHeight {
		c.reset(tipHeight)
	}
	if len(c.entries) >= maxQueryCacheEntries || c.bytes+len(res.Data) > maxQueryCacheBytes {
		return
	}
	key := queryCacheKey(req)
	if _, exists := c.entries[key]; exists {
		return
	}
	c.entries[key] = cloneQueryResponse(res)
	c.bytes += len(res.Data)
}

// reset drops every entry and adopts the new tip. Caller holds the lock.
func (c *queryCache) reset(tipHeight int64) {
	c.tipHeight = tipHeight
	c.entries = make(map[string]abci.ResponseQuery)
	c.bytes = 0
}

// cloneQueryResponse copies the byte fields, so neither the caller nor a later
// hit can see another's edits to a shared array.
func cloneQueryResponse(res abci.ResponseQuery) abci.ResponseQuery {
	out := res
	if res.Data != nil {
		out.Data = append([]byte(nil), res.Data...)
	}
	if res.Value != nil {
		out.Value = append([]byte(nil), res.Value...)
	}
	if res.Key != nil {
		out.Key = append([]byte(nil), res.Key...)
	}
	return out
}
