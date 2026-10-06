package chainmap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"

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
	if len(f.quoted) == 0 || f.quoted[0] != `"`+avl+`"` {
		t.Errorf("searched %q, want the quoted import path", f.quoted)
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
