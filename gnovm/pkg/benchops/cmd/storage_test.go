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

	dir := "../gno"
	bstore := benchmarkDiskStore()

	// load  stdlibs
	loadStdlibs(bstore)
	avlPkgDir := filepath.Join(dir, "avl")
	addPackage(bstore, avlPkgDir, "gno.land/p/demo/avl")

	storagePkgDir := filepath.Join(dir, "storage")
	pv := addPackage(bstore, storagePkgDir, storagePkgPath)
	benchStoreSet(bstore, pv)
	// verify the post content from all three boards
	for i := 0; i < 3; i++ {
		for j := 0; j < rounds; j++ {
			cx := gno.Call("GetPost", gno.X(0), gno.X(0))
			res := callFunc(bstore, pv, cx)
			parts := strings.Split(res[0].V.String(), ",")
			p := strings.Trim(parts[1], `\"`)
			expected := strings.Repeat("a", 1024)
			assert.Equal(p, expected, "it should be 1 KB of character a")
		}
	}
}
