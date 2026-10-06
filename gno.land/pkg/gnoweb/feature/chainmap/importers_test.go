package chainmap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

const avl = "gno.land/p/nt/avl/v0"

func newImporterHandler(f *fakeIndexer, imps fakeImports, lim Limiter) *Handler {
	return New(Deps{Indexer: f, Imports: imps, Domain: "gno.land", Limiter: lim})
}

// A batched deploy matches as a whole: only the chain can say which of its
// packages imports the path.
func TestImportersChecksEveryCandidateOnChain(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: 10, deploys: []indexer.Tx{
		deploy("gno.land/r/user"),
		deploy("gno.land/p/batch/one", "gno.land/p/batch/two"), // only one imports avl
		deploy(avl),                   // the package itself
		deploy("gno.land/r/user"),     // redeployed: listed once
		deploy("gno.land/r/gone"),     // no longer live
		deploy("gno.land/r/testonly"), // quotes avl in a test only
	}}
	imps := fakeImports{imports: map[string][]string{
		"gno.land/r/user":      {avl, "chain"},
		"gno.land/p/batch/one": {avl},
		"gno.land/p/batch/two": {"strings"},
		"gno.land/r/testonly":  {"strings"},
	}}

	got, err := newImporterHandler(f, imps, nil).Importers(context.Background(), nil, avl)
	if err != nil {
		t.Fatalf("Importers: %v", err)
	}
	want := []string{"gno.land/p/batch/one", "gno.land/r/user"}
	if !slices.Equal(got.Paths, want) {
		t.Fatalf("importers = %v, want %v", got.Paths, want)
	}
	if got.AtLeast {
		t.Error("every candidate was checked: the answer is complete")
	}
	if got.AsOf != 10 {
		t.Errorf("as of = %d, want the indexer tip 10", got.AsOf)
	}
	if len(f.quoted) == 0 || f.quoted[0] != avl {
		t.Errorf("searched %q, want the import path", f.quoted)
	}
}

// A candidate the node could not read is unknown, not "imports nothing".
func TestImportersSaysAtLeastWhenACheckFails(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: 10, deploys: []indexer.Tx{deploy("gno.land/r/a"), deploy("gno.land/r/b")}}
	imps := fakeImports{
		imports: map[string][]string{"gno.land/r/a": {avl}},
		fail:    map[string]bool{"gno.land/r/b": true},
	}
	got, err := newImporterHandler(f, imps, nil).Importers(context.Background(), nil, avl)
	if err != nil {
		t.Fatalf("Importers: %v", err)
	}
	if !got.AtLeast {
		t.Error("an unread candidate must turn the answer into a lower bound")
	}
	if !slices.Equal(got.Paths, []string{"gno.land/r/a"}) {
		t.Errorf("importers = %v", got.Paths)
	}
}

func TestImportersSaysAtLeastWhenTheIndexerCaps(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{
		tip:        10,
		deploys:    []indexer.Tx{deploy("gno.land/r/a")},
		deploysErr: fmt.Errorf("%w: max elements per query", indexer.ErrTooLarge),
	}
	imps := fakeImports{imports: map[string][]string{"gno.land/r/a": {avl}}}
	got, err := newImporterHandler(f, imps, nil).Importers(context.Background(), nil, avl)
	if err != nil {
		t.Fatalf("Importers: %v", err)
	}
	if !got.AtLeast || len(got.Paths) != 1 {
		t.Fatalf("got %+v, want the kept row checked and the answer marked at least", got)
	}
}

func TestImportersCapsTheChecks(t *testing.T) {
	t.Parallel()

	var deploys []indexer.Tx
	imports := make(map[string][]string)
	for i := range maxCandidates + 5 {
		p := fmt.Sprintf("gno.land/r/u%04d", i)
		deploys = append(deploys, deploy(p))
		imports[p] = []string{avl}
	}
	got, err := newImporterHandler(&fakeIndexer{tip: 1, deploys: deploys}, fakeImports{imports: imports}, nil).
		Importers(context.Background(), nil, avl)
	if err != nil {
		t.Fatalf("Importers: %v", err)
	}
	if len(got.Paths) != maxCandidates || !got.AtLeast {
		t.Fatalf("checked %d (at least: %v), want %d and a lower bound", len(got.Paths), got.AtLeast, maxCandidates)
	}
}

type denyAll struct{ asked int }

func (d *denyAll) AllowRequest(*http.Request) bool { d.asked++; return false }

// The limiter guards the expensive miss only: a cached answer is free.
func TestImportersLimitsOnlyCacheMisses(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: 10}
	lim := &denyAll{}
	h := newImporterHandler(f, fakeImports{}, lim)

	if _, err := h.Importers(context.Background(), nil, avl); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited on a miss", err)
	}
	h.importers.answers.store(avl, &Importers{AsOf: 7}, nil)
	got, err := h.Importers(context.Background(), nil, avl)
	if err != nil || got.AsOf != 7 {
		t.Fatalf("got %+v, %v; want the cached answer", got, err)
	}
	if lim.asked != 1 {
		t.Errorf("limiter asked %d times, want once: a cache hit costs nothing", lim.asked)
	}
}

func TestImportersWithoutAnIndexer(t *testing.T) {
	t.Parallel()

	h := New(Deps{Imports: fakeImports{}})
	if h.HasIndexer() {
		t.Fatal("no indexer configured, yet HasIndexer")
	}
	if _, err := h.Importers(context.Background(), nil, avl); err == nil {
		t.Fatal("Importers must refuse without an indexer")
	}
}

// The first failing band stops the scan: later bands would only fail too, and
// each failure counts against the shared indexer client's breaker.
func TestScanStopsAfterAFailingBand(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: 20 * scanBand, deploysErr: errors.New("indexer down")}
	h := newImporterHandler(f, fakeImports{}, nil)
	if _, err := h.Importers(context.Background(), nil, avl); err == nil {
		t.Fatal("a failing scan must fail the lookup")
	}
	if n := len(f.quoted); n > 2*scanConcurrency {
		t.Errorf("%d bands queried after the first failure, want at most a concurrent handful", n)
	}
}

// With every slot taken a lookup is refused at once, and the refusal is not
// remembered: the next reader may find a slot.
func TestImportersBusyIsNotRemembered(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: 10, deploys: []indexer.Tx{deploy("gno.land/r/a")}}
	h := newImporterHandler(f, fakeImports{imports: map[string][]string{"gno.land/r/a": {avl}}}, nil)
	for range maxLookups {
		h.importers.slots <- struct{}{}
	}
	if _, err := h.Importers(context.Background(), nil, avl); !errors.Is(err, ErrBusy) {
		t.Fatalf("err = %v, want ErrBusy", err)
	}
	for range maxLookups {
		<-h.importers.slots
	}
	got, err := h.Importers(context.Background(), nil, avl)
	if err != nil || len(got.Paths) != 1 {
		t.Fatalf("got %+v, %v; a busy refusal must not be cached", got, err)
	}
}

// A reader joining a lookup already running is not charged: reloading as the
// page asks must not end in "too many lookups".
func TestImportersJoiningIsFree(t *testing.T) {
	t.Parallel()

	f := &fakeIndexer{tip: 10, block: make(chan struct{})}
	lim := &countingLimiter{allow: 1}
	h := newImporterHandler(f, fakeImports{}, lim)

	short := func() context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		t.Cleanup(cancel)
		return ctx
	}
	if _, err := h.Importers(short(), nil, avl); !errors.Is(err, ErrPending) {
		t.Fatalf("first reader: err = %v, want ErrPending", err)
	}
	for range 3 {
		if _, err := h.Importers(short(), nil, avl); !errors.Is(err, ErrPending) {
			t.Fatalf("joining reader: err = %v, want ErrPending, not a rate limit", err)
		}
	}
	close(f.block)
	if lim.asked != 1 {
		t.Errorf("limiter asked %d times, want once: only the reader who started the lookup pays", lim.asked)
	}
}

type countingLimiter struct {
	mu           sync.Mutex
	allow, asked int
}

func (c *countingLimiter) AllowRequest(*http.Request) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.asked++
	return c.asked <= c.allow
}
