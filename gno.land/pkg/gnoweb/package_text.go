package gnoweb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"slices"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

// Bounds on servePackageText, so one request cannot turn into an unbounded
// number of node queries or response bytes.
const (
	packageTextMaxFiles = 100
	packageTextMaxBytes = 2 << 20
	packageTextFetchers = 8
	// packageTextTTL matches the max-age the response advertises.
	packageTextTTL = 60 * time.Second
	// packageTextCacheBytes caps the memory the cache holds.
	packageTextCacheBytes = 32 << 20
)

var errPackageTooLarge = errors.New("package too large")

// servePackageText serves every file of a package as one plain-text document,
// each under a "// file:" header, so it can be pasted into any assistant.
func (h *HTTPHandler) servePackageText(ctx context.Context, gnourl *weburl.GnoURL, w http.ResponseWriter) {
	text, err := h.packageText.get(gnourl.Path, func() ([]byte, error) {
		return h.buildPackageText(ctx, gnourl)
	})
	switch {
	case errors.Is(err, errPackageTooLarge):
		http.Error(w, "package too large", http.StatusRequestEntityTooLarge)
		return
	case err != nil:
		h.Logger.Error("unable to get package text", "path", gnourl.Path, "error", err)
		status, _ := GetClientErrorStatusView(gnourl, err, 0)
		http.Error(w, "not found", status)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(packageTextTTL.Seconds())))
	_, _ = w.Write(text)
}

// buildPackageText fetches every file of the package: one node query for
// the list, then one per file.
func (h *HTTPHandler) buildPackageText(ctx context.Context, gnourl *weburl.GnoURL) ([]byte, error) {
	files, err := h.Client.ListFiles(ctx, gnourl.Path, 0)
	files = slices.DeleteFunc(files, func(f string) bool { return f == "" })
	if err == nil && len(files) == 0 {
		err = ErrClientPackageNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(files) > packageTextMaxFiles {
		return nil, errPackageTooLarge
	}

	sources := make([][]byte, len(files))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(packageTextFetchers)
	for i, file := range files {
		g.Go(func() (err error) {
			sources[i], _, err = h.Client.File(gctx, gnourl.Path, file, 0)
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	fmt.Fprintf(&out, "// %s\n", path.Join(h.Static.Domain, gnourl.Path))
	for i, file := range files {
		fmt.Fprintf(&out, "\n// file: %s\n%s\n", file, sources[i])
	}
	if out.Len() > packageTextMaxBytes {
		return nil, errPackageTooLarge
	}
	return out.Bytes(), nil
}

// packageTextCache keeps each package text for packageTextTTL, so repeating
// a $download costs no node query, and coalesces concurrent builds of the
// same package into one. Errors are not cached: a missing package may be
// deployed the next block. The leader's context governs a shared build.
type packageTextCache struct {
	mu      sync.Mutex
	entries map[string]packageTextEntry
	size    int
	sf      singleflight.Group
	now     func() time.Time // for tests; time.Now when nil
}

type packageTextEntry struct {
	text    []byte
	expires time.Time
}

func (c *packageTextCache) get(key string, build func() ([]byte, error)) ([]byte, error) {
	if text, ok := c.lookup(key); ok {
		return text, nil
	}
	v, err, _ := c.sf.Do(key, func() (any, error) {
		if text, ok := c.lookup(key); ok {
			return text, nil
		}
		text, err := build()
		if err != nil {
			return nil, err
		}
		c.store(key, text)
		return text, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

func (c *packageTextCache) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *packageTextCache) lookup(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || !c.clock().Before(e.expires) {
		return nil, false
	}
	return e.text, true
}

// store keeps text unless the cache is full of live entries.
func (c *packageTextCache) store(key string, text []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	if c.entries == nil {
		c.entries = make(map[string]packageTextEntry)
	}
	for k, e := range c.entries {
		if k == key || !now.Before(e.expires) {
			c.size -= len(e.text)
			delete(c.entries, k)
		}
	}
	if c.size+len(text) > packageTextCacheBytes {
		return
	}
	c.entries[key] = packageTextEntry{text: text, expires: now.Add(packageTextTTL)}
	c.size += len(text)
}
