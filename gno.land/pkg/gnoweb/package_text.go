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
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

// Bounds on packageFiles, so one request cannot turn into an unbounded
// number of node queries or response bytes. The byte cap is on the files'
// sources, read by $download and by a playground copy alike.
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

// packageFile is one file of a package, as the node serves it.
type packageFile struct {
	name   string
	source []byte
}

// servePackageText serves every file of a package as one plain-text document,
// each under a "// file:" header, so it can be pasted into any assistant.
func (h *HTTPHandler) servePackageText(ctx context.Context, gnourl *weburl.GnoURL, w http.ResponseWriter) {
	files, err := h.packageFiles(ctx, gnourl.Path)
	switch {
	case errors.Is(err, errPackageTooLarge):
		http.Error(w, "package too large", http.StatusRequestEntityTooLarge)
		return
	case err != nil:
		h.Logger.Error("unable to get package text", "path", gnourl.Path, "error", err)
		status, msg := clientErrorMessage(err, 0)
		http.Error(w, msg, status)
		return
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "// %s\n", path.Join(h.Static.Domain, gnourl.Path))
	for _, f := range files {
		fmt.Fprintf(&out, "\n// file: %s\n%s\n", f.name, f.source)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(packageTextTTL.Seconds())))
	_, _ = w.Write(out.Bytes())
}

// packageFiles reads every file of the package at pkgPath, under the bounds
// above, from the cache while a recent read holds it. $download and the
// playground's copies share it. The slice is the cache's: read, never write.
func (h *HTTPHandler) packageFiles(ctx context.Context, pkgPath string) ([]packageFile, error) {
	files, err := h.pkgFiles.get(ctx, pkgPath, func() ([]packageFile, error) {
		// Concurrent requests share this read: one client hanging up must
		// not fail the others.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), packageTextBuildTimeout)
		defer cancel()
		return h.fetchPackageFiles(ctx, pkgPath)
	})
	if err != nil && ctx.Err() != nil {
		// The request ran out of time before the shared read was done,
		// which goes on for the next request.
		return nil, fmt.Errorf("%w: %w", ErrClientTimeout, ctx.Err())
	}
	return files, err
}

// fetchPackageFiles fetches every file of the package: one node query for
// the list, then one per file. Passing the byte cap stops the fetches still
// to come.
func (h *HTTPHandler) fetchPackageFiles(ctx context.Context, pkgPath string) ([]packageFile, error) {
	names, err := h.Client.ListFiles(ctx, pkgPath, 0)
	names = slices.DeleteFunc(names, func(f string) bool { return f == "" })
	if err == nil && len(names) == 0 {
		err = ErrClientPackageNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(names) > packageTextMaxFiles {
		return nil, errPackageTooLarge
	}

	files := make([]packageFile, len(names))
	var total atomic.Int64
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(packageTextFetchers)
	for i, name := range names {
		g.Go(func() (err error) {
			defer recoverAsError(&err) // errgroup does not
			if err := gctx.Err(); err != nil {
				return err // a fetch failed or passed the cap: ask no more
			}
			source, _, err := h.Client.File(gctx, pkgPath, name, 0)
			if err != nil {
				return err
			}
			if total.Add(int64(len(source))) > packageTextMaxBytes {
				return errPackageTooLarge
			}
			files[i] = packageFile{name: name, source: source}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return files, nil
}

// packageFileCache keeps each package's files for packageTextTTL, so
// repeating a $download or a copy costs no node query, and coalesces
// concurrent reads of the same package into one. errPackageTooLarge is
// kept too, since a package does not change once deployed; other errors
// are not: a missing package may be deployed the next block.
type packageFileCache struct {
	mu      sync.Mutex
	entries map[string]packageFileEntry
	size    int
	sf      singleflight.Group
	now     func() time.Time // for tests; time.Now when nil
}

type packageFileEntry struct {
	files   []packageFile
	size    int
	err     error // errPackageTooLarge, or nil
	expires time.Time
}

// get waits for the build no longer than ctx allows; the build itself goes
// on, and stores its result for the callers after.
func (c *packageFileCache) get(ctx context.Context, key string, build func() ([]packageFile, error)) ([]packageFile, error) {
	if e, ok := c.lookup(key); ok {
		return e.files, e.err
	}
	ch := c.sf.DoChan(key, func() (_ any, err error) {
		defer recoverAsError(&err) // DoChan would re-raise it where nothing can recover
		if e, ok := c.lookup(key); ok {
			return e.files, e.err
		}
		files, err := build()
		if err == nil || errors.Is(err, errPackageTooLarge) {
			c.store(key, files, err)
		}
		return files, err
	})
	select {
	case r := <-ch:
		files, _ := r.Val.([]packageFile)
		return files, r.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// recoverAsError turns a panic into *errp. Deferred on a goroutine of its
// own, which net/http's recovery does not cover, it keeps one bad read from
// stopping gnoweb.
func recoverAsError(errp *error) {
	if p := recover(); p != nil {
		*errp = fmt.Errorf("panic: %v", p)
	}
}

func (c *packageFileCache) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *packageFileCache) lookup(key string) (packageFileEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || !c.clock().Before(e.expires) {
		return packageFileEntry{}, false
	}
	return e, true
}

// store keeps the result unless the cache is full of live entries.
func (c *packageFileCache) store(key string, files []packageFile, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	if c.entries == nil {
		c.entries = make(map[string]packageFileEntry)
	}
	for k, e := range c.entries {
		if k == key || !now.Before(e.expires) {
			c.size -= e.size
			delete(c.entries, k)
		}
	}
	size := 0
	for _, f := range files {
		size += len(f.name) + len(f.source)
	}
	if c.size+size > packageTextCacheBytes {
		return
	}
	c.entries[key] = packageFileEntry{files: files, size: size, err: err, expires: now.Add(packageTextTTL)}
	c.size += size
}
