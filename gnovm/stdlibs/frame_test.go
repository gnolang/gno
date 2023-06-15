package stdlibs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestPrevRealm(t *testing.T) {
	var (
		user = gno.DerivePkgAddr("user1.gno").Bech32()
		ctx  = ExecContext{
			OrigCaller: user,
		}
	)
	tests := []struct {
		name          string
		machine       *gno.Machine
		expectedRealm Realm
	}{
		{
			name: "no frames",
			machine: &gno.Machine{
				Context: ctx,
				Frames:  []gno.Frame{},
			},
			expectedRealm: Realm{
				addr: user,
				path: "",
			},
		},
		{
			name: "one frame w/o LastPackage",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: nil},
				},
			},
			expectedRealm: Realm{
				addr: user,
				path: "",
			},
		},
		{
			name: "one non-realm frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr: user,
				path: "",
			},
		},
		{
			name: "one realm frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr: user,
				path: "",
			},
		},
		{
			name: "multiple frames with one realm",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr: user,
				path: "",
			},
		},
		{
			name: "multiple frames with multiple realms",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/zzz"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/zzz"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/yyy"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/yyy"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr: gno.DerivePkgAddr("gno.land/r/yyy").Bech32(),
				path: "gno.land/r/yyy",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			realm := prevRealm(tt.machine)

			assert.Equal(t, tt.expectedRealm, realm)
		})
	}
}
