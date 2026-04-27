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
