package doctest

import (
	"bytes"
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	IGNORE       = "ignore"
	SHOULD_PANIC = "should_panic"
	NO_RUN       = "no_run"
)

func ExecuteCodeBlock(c CodeBlock) (string, error) {
	if c.ContainsOptions(IGNORE) {
		return "", nil
	}

	if c.T == "go" {
		c.T = "gno"
	} else if c.T != "gno" {
		return "", fmt.Errorf("unsupported language: %s", c.T)
	}

	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := gno.NewStore(nil, baseStore, iavlStore)
	store.SetStrictGo2GnoMapping(true)

	m := gno.NewMachine("main", store)

	importPaths := extractImportPaths(c.Content)
	for _, path := range importPaths {
		if !gno.IsRealmPath(path) {
			pkgName := defaultPkgName(path)
			pn := gno.NewPackageNode(pkgName, path, &gno.FileSet{})
			pv := pn.NewPackage()
			store.SetBlockNode(pn)
			store.SetCachePackage(pv)
			m.SetActivePackage(pv)
		} else {
			dir := "gnovm/stdlibs/" + path
			memPkg := gno.ReadMemPackage(dir, path)
			m.RunMemPackage(memPkg, true)
		}
	}

	memPkg := &std.MemPackage{
		Name: c.Package,
		Path: c.Package,
		Files: []*std.MemFile{
			{
				Name: fmt.Sprintf("%d.%s", c.Index, c.T),
				Body: c.Content,
			},
		},
	}

	if !gno.IsRealmPath(c.Package) {
		pkgName := defaultPkgName(c.Package)
		pn := gno.NewPackageNode(pkgName, c.Package, &gno.FileSet{})
		pv := pn.NewPackage()
		store.SetBlockNode(pn)
		m.SetActivePackage(pv)
		m.RunMemPackage(memPkg, true)
	} else {
		store.ClearCache()
		m.PreprocessAllFilesAndSaveBlockNodes()
		pv := store.GetPackage(c.Package, false)
		m.SetActivePackage(pv)
	}

	// Capture output
	var output bytes.Buffer
	m.Output = &output

	if c.ContainsOptions(NO_RUN) {
		return "", nil
	}

	m.RunMain()

	result := output.String()
	if c.ContainsOptions(SHOULD_PANIC) {
		return "", fmt.Errorf("expected panic, got %q", result)
	}

	return result, nil
}

func defaultPkgName(gopkgPath string) gno.Name {
	parts := strings.Split(gopkgPath, "/")
	last := parts[len(parts)-1]
	parts = strings.Split(last, "-")
	name := parts[len(parts)-1]
	name = strings.ToLower(name)
	return gno.Name(name)
}
