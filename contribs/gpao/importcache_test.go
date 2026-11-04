package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	bfttypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// A public package added at genesis is write-once, so one fetched for a
// verification is the same package for every later one, transitive imports
// included. chainPackage carries no [addpkg] section, which reads as genesis.
func TestImportCacheServesLaterVerifications(t *testing.T) {
	low := chainPackage("gno.land/p/test/low", "package low\n\nfunc One() int { return 1 }\n")
	mid := chainPackage("gno.land/p/test/mid",
		"package mid\n\nimport \"gno.land/p/test/low\"\n\nfunc Two() int { return low.One() + 1 }\n")
	first := chainPackage("gno.land/r/test/first",
		"package first\n\nimport \"gno.land/p/test/mid\"\n\nfunc F(cur realm) int { return mid.Two() }\n")
	second := chainPackage("gno.land/r/test/second",
		"package second\n\nimport \"gno.land/p/test/mid\"\n\nfunc F(cur realm) int { return mid.Two() }\n")

	cache := importCache{}
	found, _ := cache.closure(&rpcGetter{cache: map[string]*std.MemPackage{}, qfile: fakeQFile(mid, low)}, first)
	require.ElementsMatch(t, []string{mid.Path, low.Path}, pkgPaths(found))

	queried := 0
	gone := &rpcGetter{
		cache: map[string]*std.MemPackage{},
		qfile: func(string) ([]byte, error) {
			queried++
			return nil, errors.New("package is not available")
		},
	}
	found, _ = cache.closure(gone, second)
	assert.ElementsMatch(t, []string{mid.Path, low.Path}, pkgPaths(found))
	assert.Zero(t, queried, "every import was fetched for the first verification")
}

// A package the cache does not keep is fetched again every verification. A
// live private package can be redeployed under the same path, so what was
// fetched once may not be what the chain holds next time. A package added
// after genesis is one anyone can get approved and then imported, so keeping
// it would let a submitter grow the cache without bound.
func TestImportCacheRefetchesWhatItDoesNotKeep(t *testing.T) {
	for name, gnomod := range map[string]func(path string) string{
		"private package": privateGnoMod,
		"package added after genesis": func(path string) string {
			return gno.GenGnoModLatest(path) + "\n[addpkg]\nheight = 42\n"
		},
	} {
		t.Run(name, func(t *testing.T) {
			const depPath = "gno.land/r/test/dep"
			dep := chainPackage(depPath, "package dep\n\nfunc V(cur realm) string { return \"v1\" }\n")
			dep.SetFile("gnomod.toml", gnomod(depPath))
			mpkg := chainPackage("gno.land/r/test/user",
				"package user\n\nimport \""+depPath+"\"\n\nfunc F(cur realm) string { return dep.V(cross) }\n")

			cache := importCache{}
			queried := 0
			serve := fakeQFile(dep)
			g := func() *rpcGetter {
				return &rpcGetter{
					cache: map[string]*std.MemPackage{},
					qfile: func(fp string) ([]byte, error) {
						queried++
						return serve(fp)
					},
				}
			}

			found, _ := cache.closure(g(), mpkg)
			require.Equal(t, []string{depPath}, pkgPaths(found))
			afterFirst := queried
			found, _ = cache.closure(g(), mpkg)
			assert.Equal(t, []string{depPath}, pkgPaths(found))
			assert.Greater(t, queried, afterFirst, "it must be asked for again")
		})
	}
}

// An absence answered for one verification says nothing about the next: the
// package may have been enabled in between.
func TestImportCacheRefetchesAbsentPackages(t *testing.T) {
	dep := chainPackage("gno.land/p/test/dep", "package dep\n\nfunc Add(a, b int) int { return a + b }\n")
	mpkg := chainPackage("gno.land/r/test/user",
		"package user\n\nimport \"gno.land/p/test/dep\"\n\nfunc F(cur realm) int { return dep.Add(1, 2) }\n")

	cache := importCache{}
	found, absent := cache.closure(&rpcGetter{cache: map[string]*std.MemPackage{}, qfile: fakeQFile()}, mpkg)
	require.Empty(t, found)
	require.Equal(t, []string{dep.Path}, absent)

	found, absent = cache.closure(&rpcGetter{cache: map[string]*std.MemPackage{}, qfile: fakeQFile(dep)}, mpkg)
	assert.Equal(t, []string{dep.Path}, pkgPaths(found), "the package went live after the first verification")
	assert.Empty(t, absent)
}

// The parent's "not found" crosses into the child, which classifies the path
// without asking vm/qfile for it a second time.
func TestChildTakesTheParentsAbsences(t *testing.T) {
	const parked = "gno.land/p/test/parked"
	mpkg := chainPackage("gno.land/r/test/user",
		"package user\n\nimport \""+parked+"\"\n\nfunc F(cur realm) int { return parked.Answer() }\n")

	found, absent := importCache{}.closure(&rpcGetter{cache: map[string]*std.MemPackage{}, qfile: fakeQFile()}, mpkg)

	v := newTestVerifier(t)
	queried := 0
	v.rpc = &rpcGetter{
		cache: map[string]*std.MemPackage{},
		qfile: func(string) ([]byte, error) {
			queried++
			return nil, errors.New("package is not available")
		},
		qmeta: func(path string) (vm.PackageMeta, error) {
			return vm.PackageMeta{Path: path, Status: vm.PackageStatusInert}, nil
		},
	}
	v.rpc.seed(found, absent)

	require.NoError(t, v.prepare(mpkg))
	assert.Zero(t, queried)
	assert.True(t, v.parked[parked], "the absence is still classified")
}

// The parent hands the child every genesis import it holds, so a second
// verification of a package importing one queries the node for nothing, in the
// parent or in the child; it reads block 1 alone, to tell the chain is the same.
// The same package added in a later block is fetched again, with the keeper's
// own height stamp deciding.
func TestOnlyGenesisImportsOutliveAVerification(t *testing.T) {
	const dep = "gno.land/p/nt/ufmt/v0"
	cases := map[string]struct {
		node    func(t *testing.T, pkgPath string) string
		fetched bool
	}{
		"added at genesis":    {node: nodeWithGenesisPackage, fetched: false},
		"added after genesis": {node: nodeServing, fetched: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var queries atomic.Int64
			counted := httptest.NewServer(countingBody(t, tc.node(t, dep), `"method":"abci_query"`, &queries))
			t.Cleanup(counted.Close)

			o := newTestOracleAt(t, counted.URL)
			o.cfg.verifyBudget = time.Minute

			require.NoError(t, o.verify(context.Background(), packageImporting(dep)))
			afterFirst := queries.Load()
			require.Positive(t, afterFirst, "premise: the first verification must fetch the import")

			require.NoError(t, o.verify(context.Background(), packageImporting(dep)))
			assert.Equal(t, tc.fetched, queries.Load() > afterFirst)
		})
	}
}

// TestRestartedChainIsVerifiedAgainstItsOwnGenesis: a chain started again under
// the same ID has a genesis of its own, so what the cache holds from the last
// one must not decide a verification on it, whatever the tip did meanwhile.
func TestRestartedChainIsVerifiedAgainstItsOwnGenesis(t *testing.T) {
	const lib = "gno.land/p/test/lib"
	first := nodeWithGenesis(t, chainPackage(lib, "package lib\n\nfunc One() int { return 1 }\n"))
	second := nodeWithGenesis(t, chainPackage(lib,
		"package lib\n\nfunc One() int { return 1 }\n\nfunc Two() int { return 2 }\n"))
	var node atomic.Pointer[httputil.ReverseProxy]
	node.Store(nodeProxy(t, first))
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		node.Load().ServeHTTP(w, r)
	}))
	t.Cleanup(proxy.Close)
	calling := func(fn string) *std.MemPackage {
		return chainPackage("gno.land/r/test/user",
			"package user\n\nimport \""+lib+"\"\n\nfunc N(cur realm) int { return lib."+fn+"() }\n")
	}

	o := newTestOracleAt(t, proxy.URL)
	o.cfg.verifyBudget = time.Minute
	require.NoError(t, o.verify(t.Context(), calling("One")))
	require.Contains(t, o.genesisImports(), lib, "premise: the first chain's lib is cached")

	node.Store(nodeProxy(t, second))
	assert.NoError(t, o.verify(t.Context(), calling("Two")))
}

// nodeWithGenesisPackage starts a node whose genesis deploys pkgPath from
// examples/, as a chain deploys its own packages, and returns its RPC address.
func nodeWithGenesisPackage(t *testing.T, pkgPath string) string {
	t.Helper()
	mpkg, err := gno.ReadMemPackage(filepath.Join(gnoenv.RootDir(), "examples", pkgPath), pkgPath, gno.MPUserAll)
	require.NoError(t, err)
	return nodeWithGenesis(t, mpkg)
}

// nodeWithGenesis starts a node whose genesis deploys mpkg and returns its RPC
// address.
func nodeWithGenesis(t *testing.T, mpkg *std.MemPackage) string {
	t.Helper()
	cfg := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	cfg.SkipGenesisSigVerification = true

	tx, err := gnoland.LoadPackage(mpkg, crypto.MustAddressFromString(integration.DefaultAccount_Address),
		std.NewFee(50_000_000, std.MustParseCoin("1000000ugnot")), nil)
	require.NoError(t, err)
	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Txs = append(ggs.Txs, gnoland.TxWithMetadata{Tx: tx})
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	t.Cleanup(func() { node.Stop() })
	return remote
}

func pkgPaths(pkgs []*std.MemPackage) []string {
	paths := make([]string, 0, len(pkgs))
	for _, mpkg := range pkgs {
		paths = append(paths, mpkg.Path)
	}
	return paths
}

// firstBlockRPC is a node whose block 1 has hash, or that will not give it
// when hash is nil. The embedded interface is nil on purpose, as in stubRPC.
type firstBlockRPC struct {
	rpcclient.Client
	hash []byte
}

func (s firstBlockRPC) BlockchainInfo(context.Context, int64, int64) (*ctypes.ResultBlockchainInfo, error) {
	if s.hash == nil {
		return nil, errors.New("blockchain refused")
	}
	return &ctypes.ResultBlockchainInfo{
		BlockMetas: []*bfttypes.BlockMeta{{BlockID: bfttypes.BlockID{Hash: s.hash}}},
	}, nil
}

// TestImportsFollowBlockOne: the cache belongs to the chain whose block 1 it
// was filled under. It is kept while block 1 stays the same, and dropped when
// block 1 changes or the node will not say what it is.
func TestImportsFollowBlockOne(t *testing.T) {
	for name, tc := range map[string]struct {
		first []byte
		kept  int
	}{
		"same chain":          {first: []byte("a"), kept: 1},
		"chain started again": {first: []byte("b"), kept: 0},
		"block 1 refused":     {first: nil, kept: 0},
	} {
		t.Run(name, func(t *testing.T) {
			o := newStubOracle(firstBlockRPC{hash: tc.first})
			o.imports = importCache{"gno.land/p/test/lib": chainPackage("gno.land/p/test/lib", "package lib\n")}
			o.importsChain = []byte("a")

			o.dropImportsFromAnotherChain(t.Context())
			assert.Len(t, o.genesisImports(), tc.kept)
		})
	}
}
