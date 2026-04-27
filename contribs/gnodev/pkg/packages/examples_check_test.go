package packages

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/stretchr/testify/assert"
)

func TestCheckMissingExampleImports(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "myrealm")
	writePkg(t, pkgDir, "gno.land/r/me/myrealm",
		`package myrealm
import "gno.land/r/demo/boards"
var _ = boards.Render
`)

	// Empty fetcher + no examples + no extra root → demo/boards is unresolvable.
	l := New(Config{
		Examples: false,
		Fetcher:  pkgdownload.NewInMemoryFetcher(),
		Logger:   testLogger(),
	})

	missing := CheckMissingExampleImports(l, root)
	assert.Equal(t, []string{"gno.land/r/demo/boards"}, missing)
}

func TestCheckMissingExampleImports_AllResolved(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "alone")
	writePkg(t, pkgDir, "gno.land/p/me/alone", "package alone\n")
	_ = pkgDir

	l := New(Config{
		Examples: false,
		Fetcher:  pkgdownload.NewInMemoryFetcher(),
		Logger:   testLogger(),
	})

	missing := CheckMissingExampleImports(l, root)
	assert.Empty(t, missing)
}

func TestCheckMissingExampleImports_StdlibIgnored(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "uses-chain")
	writePkg(t, pkgDir, "gno.land/p/me/usechain",
		`package usechain
import "chain"
var _ = chain.ChainDomain
`)
	_ = pkgDir

	l := New(Config{
		Examples: false,
		Fetcher:  pkgdownload.NewInMemoryFetcher(),
		Logger:   testLogger(),
	})

	missing := CheckMissingExampleImports(l, root)
	assert.Empty(t, missing, "stdlib imports must be ignored")
}

func TestCheckMissingExampleImports_EmptyWorkspace(t *testing.T) {
	l := New(Config{Logger: testLogger()})
	assert.Nil(t, CheckMissingExampleImports(l, ""))
}

// TestCheckMissingExampleImports_NoMutation locks in the contract that the
// helper does not write to l.index or l.tracked. A revert to l.Resolve(imp)
// would silently restore the blocking-RPC + state-pollution bug fixed in
// 2f21494b91 — the FS-resolved import would land in both maps on hit.
//
// The workspace imports a package reachable via an extra root, so a revert
// to Resolve would mutate l.index/l.tracked on the FS hit. LookupFS must not.
func TestCheckMissingExampleImports_NoMutation(t *testing.T) {
	root := t.TempDir()
	consumerDir := filepath.Join(root, "consumer")
	writePkg(t, consumerDir, "gno.land/r/me/consumer",
		`package consumer
import "gno.land/p/demo/dep"
var _ = dep.X
`)

	extra := t.TempDir()
	depDir := filepath.Join(extra, "dep")
	writePkg(t, depDir, "gno.land/p/demo/dep", "package dep\nvar X = 1\n")

	l := New(Config{
		Examples:   false,
		ExtraRoots: []string{extra},
		Fetcher:    pkgdownload.NewInMemoryFetcher(),
		Logger:     testLogger(),
	})

	missing := CheckMissingExampleImports(l, root)
	// Sanity: the dep is reachable via extra root, so the helper should not
	// flag it as missing. This guarantees we exercise the FS-hit code path.
	assert.Empty(t, missing, "dep is reachable via extra root")

	l.mu.RLock()
	defer l.mu.RUnlock()
	assert.Empty(t, l.index, "CheckMissingExampleImports must not insert into l.index")
	assert.Empty(t, l.tracked, "CheckMissingExampleImports must not insert into l.tracked")
}
