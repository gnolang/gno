package vm_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCodecParity_VM(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(std.Package)
	cdc.RegisterPackage(vm.Package)
	cdc.Seal()

	caller := crypto.AddressFromPreimage([]byte("caller"))
	pkg := &std.MemPackage{
		Name: "foo",
		Path: "gno.land/r/demo/foo",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package foo\nfunc Hello() {}\n"},
		},
	}

	cases := []struct {
		name string
		v    any
	}{
		{"MsgCall", &vm.MsgCall{
			Caller:  caller,
			Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
			PkgPath: "gno.land/r/demo/foo",
			Func:    "Hello",
			Args:    []string{"world"},
		}},
		{"MsgAddPackage", &vm.MsgAddPackage{
			Creator: caller,
			Package: pkg,
			Send:    std.Coins{{Denom: "ugnot", Amount: 50}},
		}},
		{"MsgRun", &vm.MsgRun{
			Caller:  caller,
			Package: pkg,
		}},
		{"Params", &vm.Params{}},
		{"GenesisState", &vm.GenesisState{Params: vm.DefaultParams()}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
