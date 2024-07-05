package gnoland

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
)

func TestLoadPackagesFromDir(t *testing.T) {
	var (
		memPkg = &std.MemPackage{
			Name: "foo",
			Path: "gno.land/r/demo/foo",
			Files: []*std.MemFile{
				{Name: "empty.gno", Body: "package foo"},
				{Name: "generated.gno.tpl", Body: `package foo

var ChainID = {{.Genesis.ChainID | printf "%q"}}
`},
			},
		}
		creator = bft.Address{}
		fee     = std.Fee{}
		deposit = std.Coins{}
	)
	tplData := GenesisTplData{}
	tplData.Genesis.ChainID = "test"
	tx, err := LoadPackage(memPkg, creator, fee, deposit, tplData)
	assert.NoError(t, err)
	var msg vm.MsgAddPackage
	bz := tx.Msgs[0].GetSignBytes()
	amino.MustUnmarshalJSON(bz, &msg)
	for _, file := range msg.Package.Files {
		assert.NotEqual(t, "generated.gno.tpl", file.Name, "*.tpl files should be removed")
		if file.Name == "generated.gno.tpl" {
			expected := `package foo

var ChainID = "test"
`
			assert.Equal(t, expected, file.Body)
		}
	}
}
