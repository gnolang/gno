package stdlibs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func TestPrevRealm(t *testing.T) {
	var (
		user = gno.DerivePkgAddr("user1.gno").Bech32()
		ctx  = ExecContext{
			OrigCaller: user,
		}
	)
	tests := []struct {
		name                 string
		machine              *gno.Machine
		expectedLastCaller   crypto.Bech32Address
		expecteedLastPkgPath string
	}{
		{
			name: "no frames",
			machine: &gno.Machine{
				Context: ctx,
				Frames:  []gno.Frame{},
			},
			expectedLastCaller:   user,
			expecteedLastPkgPath: "",
		},
		{
			name: "one frame w/o LastPackage",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: nil},
				},
			},
			expectedLastCaller:   user,
			expecteedLastPkgPath: "",
		},
		{
			name: "one non-realm frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
				},
			},
			expectedLastCaller:   user,
			expecteedLastPkgPath: "",
		},
		{
			name: "one realm frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedLastCaller:   user,
			expecteedLastPkgPath: "",
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
			expectedLastCaller:   user,
			expecteedLastPkgPath: "",
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
			expectedLastCaller:   gno.DerivePkgAddr("gno.land/r/yyy").Bech32(),
			expecteedLastPkgPath: "gno.land/r/yyy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			lastCaller, lastPkgPath := prevRealm(tt.machine)

			assert.Equal(tt.expectedLastCaller, lastCaller)
			assert.Equal(tt.expecteedLastPkgPath, lastPkgPath)
		})
	}
}
