package std

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestPrevRealmIsOrigin(t *testing.T) {
	var (
		user = gno.DerivePkgAddr("user1.gno").Bech32()
		ctx  = ExecContext{
			OrigCaller: user,
		}
		msgCallFrame = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "main"}}
		msgRunFrame  = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/user-realm"}}
	)
	tests := []struct {
		name                 string
		machine              *gno.Machine
		expectedRealm        Realm
		expectedIsOriginCall bool
	}{
		{
			name: "no frames",
			machine: &gno.Machine{
				Context: ctx,
				Frames:  []*gno.Frame{},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "one frame w/o LastPackage",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					{LastPackage: nil},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "one package frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "one realm frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "one msgCall frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgCallFrame,
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: true,
		},
		{
			name: "one msgRun frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgRunFrame,
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "one package frame and one msgCall frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgCallFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: true,
		},
		{
			name: "one realm frame and one msgCall frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgCallFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: true,
		},
		{
			name: "one package frame and one msgRun frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgRunFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "one realm frame and one msgRun frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgRunFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "multiple frames with one realm",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    user,
				pkgPath: "",
			},
			expectedIsOriginCall: false,
		},
		{
			name: "multiple frames with multiple realms",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/zzz"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/zzz"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/yyy"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/yyy"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedRealm: Realm{
				addr:    gno.DerivePkgAddr("gno.land/r/yyy").Bech32(),
				pkgPath: "gno.land/r/yyy",
			},
			expectedIsOriginCall: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			realm := PrevRealm(tt.machine)
			isOrigin := IsOriginCall(tt.machine)

			assert.Equal(tt.expectedRealm, realm)
			assert.Equal(tt.expectedIsOriginCall, isOrigin)
		})
	}
}
