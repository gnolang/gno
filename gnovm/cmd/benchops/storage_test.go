package main

import (
	"path/filepath"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
)

func TestBenchStoreSet(t *testing.T) {
	assert := assert.New(t)

	dir := "../../pkg/benchops/gno"
	bstore := benchmarkDiskStore()
	t.Cleanup(func() { bstore.Delete() })
	gstore := bstore.gnoStore

	// load  stdlibs
	loadStdlibs(bstore)
	avlPkgDir := filepath.Join(dir, "avl")
	addPackage(gstore, avlPkgDir, "gno.land/p/nt/avl")

	storagePkgDir := filepath.Join(dir, "storage")
	pv := addPackage(gstore, storagePkgDir, storagePkgPath)
	benchStoreSet(bstore, pv)
	// verify the post content from all three boards
	for range 3 {
		for range rounds {
			cx := gno.Call("GetPost", gno.X(0), gno.X(0))
			res := callFunc(gstore, pv, cx)
			parts := strings.Split(res[0].V.String(), ",")
			p := strings.Trim(parts[1], `\"`)
			expected := strings.Repeat("a", 1024)
			assert.Equal(p, expected, "it should be 1 KB of character a")
		}
	}
}
