package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusBoardUnknownIsAnAnswer covers the absent case.
//
// A path nobody submitted and a path the oracle has not reached yet must both
// come back as a verdict rather than an error, or a caller cannot tell "we have
// nothing on this" from "the request failed".
func TestStatusBoardUnknownIsAnAnswer(t *testing.T) {
	b := newStatusBoard()
	got := b.get("gno.land/r/nobody/here")
	assert.Equal(t, statusUnknown, got.Status)
	assert.Equal(t, "gno.land/r/nobody/here", got.Path)
	assert.Empty(t, b.list())
}

// TestStatusBoardKeepsTheLatestVerdict pins that a path's entry is replaced,
// not appended. A package is re-verified when its bytes change, and a stale
// "rejected" left beside a later "approved" is worse than no record.
func TestStatusBoardKeepsTheLatestVerdict(t *testing.T) {
	b := newStatusBoard()
	const path = "gno.land/r/x/y"

	b.record(path, statusRejected, "does not compile", 0)
	require.Equal(t, statusRejected, b.get(path).Status)

	b.record(path, statusApproved, "", 0)
	assert.Equal(t, statusApproved, b.get(path).Status)
	assert.Empty(t, b.get(path).Reason, "the old reason must not survive the new verdict")
	assert.Len(t, b.list(), 1, "one path, one entry")
}

// TestStatusBoardIsSafeAcrossGoroutines is the reason the board exists as its
// own structure: the verifier writes while an HTTP handler reads, which the
// oracle's other maps cannot survive because they carry no lock.
func TestStatusBoardIsSafeAcrossGoroutines(t *testing.T) {
	b := newStatusBoard()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) { defer wg.Done(); b.record("gno.land/r/x/y", statusPending, "retrying", i) }(i)
		go func() { defer wg.Done(); _ = b.list(); _ = b.get("gno.land/r/x/y") }()
	}
	wg.Wait()
}

// TestStatusHandlerServesOneAndAll covers both routes, including that a package
// path with slashes survives the URL.
func TestStatusHandlerServesOneAndAll(t *testing.T) {
	b := newStatusBoard()
	b.record("gno.land/r/b/two", statusApproved, "", 0)
	b.record("gno.land/r/a/one", statusRejected, "undefined: Foo", 0)
	srv := httptest.NewServer(b.statusHandler())
	defer srv.Close()

	get := func(t *testing.T, path string, into any) {
		t.Helper()
		res, err := http.Get(srv.URL + path)
		require.NoError(t, err)
		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.NoError(t, json.NewDecoder(res.Body).Decode(into))
	}

	var all []pkgStatus
	get(t, "/status", &all)
	require.Len(t, all, 2)
	assert.Equal(t, "gno.land/r/a/one", all[0].Path, "sorted, so the output is stable")

	var one pkgStatus
	get(t, "/status/gno.land/r/a/one", &one)
	assert.Equal(t, statusRejected, one.Status)
	assert.Contains(t, one.Reason, "undefined: Foo",
		"the reason is the whole point; a status alone says nothing to act on")

	var missing pkgStatus
	get(t, "/status/gno.land/r/never/submitted", &missing)
	assert.Equal(t, statusUnknown, missing.Status)
}

// TestStatusBoardNilIsInert lets the oracle run without a board in tests that
// construct it directly, rather than making every one of them wire one up.
func TestStatusBoardNilIsInert(t *testing.T) {
	var b *statusBoard
	assert.NotPanics(t, func() {
		b.record("gno.land/r/x/y", statusPending, "", 0)
		assert.Equal(t, statusUnknown, b.get("gno.land/r/x/y").Status)
		assert.Nil(t, b.list())
	})
}
