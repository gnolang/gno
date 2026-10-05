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
	// packageTextBuildTimeout bounds a shared build, which outlives the
	// request that started it.
	packageTextBuildTimeout = 30 * time.Second
)

var errPackageTooLarge = errors.New("package too large")

// servePackageText serves every file of a package as one plain-text document,
// each under a "// file:" header, so it can be pasted into any assistant.
func (h *HTTPHandler) servePackageText(ctx context.Context, gnourl *weburl.GnoURL, w http.ResponseWriter) {
	text, err := h.packageText.get(gnourl.Path, func() ([]byte, error) {
		// Concurrent requests share this build: one client hanging up must
		// not fail the others.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), packageTextBuildTimeout)
		defer cancel()
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
	return bytes.Clone(out.Bytes()), nil
}

// packageTextCache keeps each package text for packageTextTTL, so repeating
// a $download costs no node query, and coalesces concurrent builds of the
// same package into one. errPackageTooLarge is kept too, since a package
// does not change once deployed; other errors are not: a missing package
// may be deployed the next block.
type packageTextCache struct {
	mu      sync.Mutex
	entries map[string]packageTextEntry
	size    int
	sf      singleflight.Group
	now     func() time.Time // for tests; time.Now when nil
}

type packageTextEntry struct {
	text    []byte
	err     error // errPackageTooLarge, or nil
	expires time.Time
}

func (c *packageTextCache) get(key string, build func() ([]byte, error)) ([]byte, error) {
	if e, ok := c.lookup(key); ok {
		return e.text, e.err
	}
	v, err, _ := c.sf.Do(key, func() (any, error) {
		if e, ok := c.lookup(key); ok {
			return e.text, e.err
		}
		text, err := build()
		if err == nil || errors.Is(err, errPackageTooLarge) {
			c.store(key, text, err)
		}
		return text, err
	})
	text, _ := v.([]byte)
	return text, err
}

func (c *packageTextCache) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *packageTextCache) lookup(key string) (packageTextEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || !c.clock().Before(e.expires) {
		return packageTextEntry{}, false
	}
	return e, true
}

// store keeps the result unless the cache is full of live entries.
func (c *packageTextCache) store(key string, text []byte, err error) {
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
	c.entries[key] = packageTextEntry{text: text, err: err, expires: now.Add(packageTextTTL)}
	c.size += len(text)
}
