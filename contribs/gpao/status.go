package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// States a package can be in from the oracle's point of view.
//
// These say what gpao decided, which is not what the chain knows. The chain can
// report that a package is parked and why no enable could succeed right now --
// no approvers configured, policy moved. It cannot report that the code failed
// to type-check, or that the enable was simulated and would fail: that is the
// oracle's alone, and without somewhere to put it the only record is the
// operator's stderr.
const (
	StatusRejected = "rejected" // verification says the code is bad
	StatusPending  = "pending"  // no verdict yet; will be retried
	StatusGaveUp   = "gave_up"  // retried to the cap, needs a human
	StatusBlocked  = "blocked"  // nothing wrong with the package; the oracle cannot proceed
	StatusApproved = "approved" // enabled on-chain by this oracle
	StatusUnknown  = "unknown"  // never seen
)

// PkgStatus is one package's last verdict.
type PkgStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	// Reason is free text from the failure itself. Present for anything a
	// submitter would need in order to act.
	Reason string `json:"reason,omitempty"`
	// Attempt counts tries so far, for the states that retry.
	Attempt int `json:"attempt,omitempty"`
}

// statusBoard records the oracle's verdicts for reading from outside.
//
// Deliberately separate from seen/overBudget/failedEnable rather than exposing
// those. They are owned by the verifier goroutine and carry no lock, which is
// what keeps the hot path cheap; serving them over HTTP would read them from
// another goroutine and race. This is the only structure that crosses that
// boundary, so it is the only one that needs a mutex.
//
// Keyed by path, while the retry counters are keyed by content hash. That is
// the right split: a retry is about a specific set of bytes, but somebody
// asking after a package wants the latest word on the path they submitted.
type statusBoard struct {
	mu sync.RWMutex
	by map[string]PkgStatus
}

func newStatusBoard() *statusBoard {
	return &statusBoard{by: make(map[string]PkgStatus)}
}

// record stores the latest verdict for a path, replacing any earlier one.
func (b *statusBoard) record(path, status, reason string, attempt int) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.by[path] = PkgStatus{Path: path, Status: status, Reason: reason, Attempt: attempt}
}

// get returns the verdict for a path. A path the oracle never reached is
// StatusUnknown rather than an error: "we have nothing on this" is an answer,
// and the caller cannot tell it from a failed lookup otherwise.
func (b *statusBoard) get(path string) PkgStatus {
	if b == nil {
		return PkgStatus{Path: path, Status: StatusUnknown}
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if s, ok := b.by[path]; ok {
		return s
	}
	return PkgStatus{Path: path, Status: StatusUnknown}
}

// list returns every verdict, sorted by path so the output is stable.
func (b *statusBoard) list() []PkgStatus {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]PkgStatus, 0, len(b.by))
	for _, s := range b.by {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// statusHandler serves the board read-only over HTTP.
//
//	GET /status            every verdict
//	GET /status/<pkgpath>  one verdict, StatusUnknown if there is none
//
// Read-only and unauthenticated by design: every verdict here concerns a
// package that was submitted in a public transaction, and the point is for its
// submitter to be able to read it without the operator relaying stderr.
func (b *statusBoard) statusHandler() http.Handler {
	mux := http.NewServeMux()
	write := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		write(w, b.list())
	})
	mux.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/status/")
		if path == "" {
			write(w, b.list())
			return
		}
		write(w, b.get(path))
	})
	return mux
}
