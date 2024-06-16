package doctest

import (
	"bytes"
	"fmt"

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

	m := gno.NewMachine("main", store)
	m.RunMemPackageWithOverrides(&std.MemPackage{
		Name: c.Package,
		Path: c.Package,
		Files: []*std.MemFile{
			{Name: fmt.Sprintf("%d.%s", c.Index, c.T), Body: c.Content},
		},
	}, true)

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
