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
		// The two inert-flow messages. They arrived with about 400 lines of
		// hand-written pb3_gen.go and were not covered here, so a wrong field
		// number or a dropped field in either would not have been caught -- while
		// the same diff extended this test for Params.
		//
		// A dropped field matters more than usual for these two: Approver is the
		// whole authorization, and PkgPath is what the chain compiles. Losing
		// either on the wire is not a decoding curiosity.
		{"MsgEnablePackage", &vm.MsgEnablePackage{
			Approver: caller,
			PkgPath:  "gno.land/r/demo/foo",
		}},
		{"MsgDisablePackage", &vm.MsgDisablePackage{
			Approver: caller,
			PkgPath:  "gno.land/r/demo/foo",
		}},
		{"Params", &vm.Params{}},
		// Populated address lists. The empty Params case above never enters the
		// repeated-field marshal/size/unmarshal loops in pb3_gen.go, so it passes
		// even if a repeated field has no codec at all — and DefaultParams()
		// leaves all three lists nil, so GenesisState below does not cover them
		// either. That gap is not cosmetic: every one of these lists is
		// fail-closed, so a field silently dropped on the amino-binary path
		// disables the capability it guards rather than erroring.
		//
		// Two entries per list, not one: a single-element list never exercises
		// the unmarshal loop's multi-element continuation, which is the site most
		// easily got wrong. Distinct addresses per list so a field number copied
		// from the block above cannot pass by writing into the wrong slice.
		{"ParamsAddressLists", &vm.Params{
			CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissioned,
			CodeSubmitters: []crypto.Address{
				crypto.AddressFromPreimage([]byte("submitter1")),
				crypto.AddressFromPreimage([]byte("submitter2")),
			},
			PkgApprovers: []crypto.Address{
				crypto.AddressFromPreimage([]byte("approver1")),
				crypto.AddressFromPreimage([]byte("approver2")),
			},
			RunSubmitters: []crypto.Address{
				crypto.AddressFromPreimage([]byte("runner1")),
				crypto.AddressFromPreimage([]byte("runner2")),
			},
		}},
		// One entry each, to catch a codec that only works for even counts.
		{"ParamsAddressListsSingle", &vm.Params{
			CodeSubmitters: []crypto.Address{crypto.AddressFromPreimage([]byte("s"))},
			PkgApprovers:   []crypto.Address{crypto.AddressFromPreimage([]byte("a"))},
			RunSubmitters:  []crypto.Address{crypto.AddressFromPreimage([]byte("r"))},
		}},
		{"GenesisState", &vm.GenesisState{Params: vm.DefaultParams()}},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
