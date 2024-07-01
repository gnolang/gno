package gnoland

import (
	"fmt"
	"testing"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestLoadPackagesFromDir(t *testing.T) {
	var (
		memPkg = &std.MemPackage{
			Name: "foo",
			Path: "gno.land/r/demo/foo",
			Files: []*std.MemFile{
				{Name: "empty.gno", Body: "package foo"},
				{Name: "generated.gno.tpl", Body: `package foo

/* {{.}} */
`},
			},
		}
		creator = bft.Address{}
		fee     = std.Fee{}
		deposit = std.Coins{}
	)
	tx, err := LoadPackage(memPkg, creator, fee, deposit)
	fmt.Printf("$$$$$$$$$ tx=%v, err=%v\n", tx, err)
}
