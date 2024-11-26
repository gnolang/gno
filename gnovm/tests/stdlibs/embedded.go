package stdlibs

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
)

// embeddedSources embeds the testing stdlibs.
// Be careful to remove transpile artifacts before building release binaries or they will be included
//
//go:embed */*
var embeddedSources embed.FS

// EmbeddedMemPackage returns a slice of [gnovm.MemPackage] generated from embedded stdlibs sources
func EmbeddedMemPackage(pkgPath string) *gnovm.MemPackage {
	return embeddedMemPackages()[pkgPath]
}

var embeddedMemPackages = sync.OnceValue(func() map[string]*gnovm.MemPackage {
	initOrder := stdlibs.InitOrder()
	memPkgs := make(map[string]*gnovm.MemPackage, len(initOrder))

	for _, pkgPath := range initOrder {
		filesystems := []fs.FS{embeddedSources, stdlibs.EmbeddedSources()}
		filesystemsNames := []string{"test", "normal"}
		files := make([]string, 0, 32) // pre-alloc 32 as a likely high number of files
		for i, fsys := range filesystems {
			entries, err := fs.ReadDir(fsys, pkgPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				panic(fmt.Errorf("failed to read embedded stdlib %q in %q fsys: %w", pkgPath, filesystemsNames[i], err))
			}
			for _, f := range entries {
				// NOTE: RunMemPackage has other rules; those should be mostly useful
				// for on-chain packages (ie. include README and gno.mod).
				fp := filepath.Join(pkgPath, f.Name())
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".gno") && !slices.Contains(files, fp) {
					files = append(files, fp)
				}
			}
		}
		if len(files) == 0 {
			return nil
		}

		memPkgs[pkgPath] = gnolang.ReadMemPackageFromList(filesystems, files, pkgPath)
	}

	return memPkgs
})
