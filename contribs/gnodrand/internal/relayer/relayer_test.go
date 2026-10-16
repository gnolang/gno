package relayer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/contribs/gnodrand/internal/evmnet"
)

// Real evmnet beacons.
var known = map[uint64]string{
	1:    "11f812d738a36b2210dc88c2d635ad8039588205f42445d6de09e6530165c3462a23aca348c84badcf8df5321ac24577b7963d5b0d780bc4626baedb45cde373",
	2:    "1dbfe644485f31e2a7978fb551b50933b0001e64397fd2b7dc4022e8e2fcdbc325798cea5506deec489c8db3b4697f6b1bb6be8934b95bd9c0db88923487d9ad",
	1000: "06fd5996329504d3a56b482d9222bf7205857d0a9559ddd216ca31a286f6a8cc0a120f021aac2f13553fb164f62bc3a5ca32c76dea88a777b39bcf3cac5fdbd6",
}

type fakeSource map[uint64]string

func (f fakeSource) Beacon(_ context.Context, round uint64) (Beacon, error) {
	sig, ok := f[round]
	if !ok {
		return Beacon{}, errors.New("not found")
	}
	return Beacon{Round: round, Signature: sig}, nil
}

type fakeChain struct {
	pending   []uint64
	submitted [][]Beacon
	err       error
}

func (c *fakeChain) PendingRounds(_ context.Context, limit int) ([]uint64, error) {
	if limit < len(c.pending) {
		return c.pending[:limit], nil
	}
	return c.pending, nil
}

func (c *fakeChain) Submit(_ context.Context, b []Beacon) error {
	if c.err != nil {
		return c.err
	}
	c.submitted = append(c.submitted, b)
	return nil
}

func newRelayer(src Source, chain Chain, now time.Time) *Relayer {
	return &Relayer{
		Source:   src,
		Chain:    chain,
		Log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		MaxBatch: 10,
		Now:      func() time.Time { return now },
	}
}

func rounds(bs []Beacon) []uint64 {
	var out []uint64
	for _, b := range bs {
		out = append(out, b.Round)
	}
	return out
}

func TestTickSubmitsPublishedRounds(t *testing.T) {
	chain := &fakeChain{pending: []uint64{1, 2, 1000, 5000}}
	// At the publication time of round 1000: 5000 is still in the future.
	r := newRelayer(fakeSource(known), chain, time.Unix(evmnet.TimeOf(1000), 0))

	n, err := r.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 || len(chain.submitted) != 1 {
		t.Fatalf("submitted %d beacons in %d txs, want 3 in 1", n, len(chain.submitted))
	}
	if got := rounds(chain.submitted[0]); !reflect.DeepEqual(got, []uint64{1, 2, 1000}) {
		t.Fatalf("rounds = %v", got)
	}
}

func TestTickSkipsBadBeacons(t *testing.T) {
	src := fakeSource{
		1: known[1],
		2: known[1000], // valid signature, wrong round
	}
	chain := &fakeChain{pending: []uint64{1, 2, 3}} // 3: fetch fails
	r := newRelayer(src, chain, time.Unix(evmnet.TimeOf(1000), 0))

	n, err := r.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 || !reflect.DeepEqual(rounds(chain.submitted[0]), []uint64{1}) {
		t.Fatalf("submitted %v, want only round 1", chain.submitted)
	}
}

func TestTickNothingToDo(t *testing.T) {
	chain := &fakeChain{pending: []uint64{1000}}
	r := newRelayer(fakeSource(known), chain, time.Unix(evmnet.TimeOf(999), 0))
	if n, err := r.Tick(context.Background()); n != 0 || err != nil || len(chain.submitted) != 0 {
		t.Fatalf("n=%d err=%v submitted=%v", n, err, chain.submitted)
	}
}

func TestTickRespectsMaxBatch(t *testing.T) {
	chain := &fakeChain{pending: []uint64{1, 2, 1000}}
	r := newRelayer(fakeSource(known), chain, time.Unix(evmnet.TimeOf(1000), 0))
	r.MaxBatch = 2
	if n, _ := r.Tick(context.Background()); n != 2 {
		t.Fatalf("n = %d, want 2", n)
	}
}

func TestTickSubmitError(t *testing.T) {
	chain := &fakeChain{pending: []uint64{1}, err: errors.New("boom")}
	r := newRelayer(fakeSource(known), chain, time.Unix(evmnet.TimeOf(1000), 0))
	if _, err := r.Tick(context.Background()); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v", err)
	}
}

func TestHTTPSourceFallsBack(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer down.Close()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		want := "/" + evmnet.ChainHash + "/public/1000"
		if req.URL.Path != want {
			http.NotFound(w, req)
			return
		}
		_, _ = fmt.Fprintf(w, `{"round":1000,"randomness":"x","signature":%q}`, known[1000])
	}))
	defer up.Close()

	src := HTTPSource{Mirrors: []string{down.URL, up.URL + "/"}, Client: up.Client()}
	b, err := src.Beacon(context.Background(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	if b.Round != 1000 || b.Signature != known[1000] {
		t.Fatalf("beacon = %+v", b)
	}

	src.Mirrors = []string{down.URL}
	if _, err := src.Beacon(context.Background(), 1000); err == nil {
		t.Fatal("expected error when every mirror fails")
	}
}

func TestParsePendingRounds(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []uint64
		err  bool
	}{
		{`("" string)`, nil, false},
		{"(\"5\" string)\n", []uint64{5}, false},
		{`("12,15,20" string)`, []uint64{12, 15, 20}, false},
		{`(12 uint64)`, nil, true},
		{`("1,x" string)`, nil, true},
	} {
		got, err := parsePendingRounds(tc.in)
		if (err != nil) != tc.err || !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parse(%q) = %v, %v", tc.in, got, err)
		}
	}
}
