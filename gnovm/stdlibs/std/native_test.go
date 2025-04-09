package std

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func TestPreviousRealmIsOrigin(t *testing.T) {
	var (
		user = gno.DerivePkgAddr("user1.gno").Bech32()
		ctx  = ExecContext{
			OriginCaller: user,
		}
		msgCallFrame = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "main"}, Func: &gno.FuncValue{Name: "main"},
			WithSwitch: true, DidSwitch: true, LastRealm: &gno.Realm{Path: ""}}
		//msgRunFrame = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/g1337/run"}}
		withswitchFrame      = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}, Func: &gno.FuncValue{Name: "SetFoo"}, WithSwitch: true, DidSwitch: true, LastRealm: &gno.Realm{Path: ""}}
		switchrealmFrame     = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/yyy"}, Func: &gno.FuncValue{Name: "switchrealm"}, WithSwitch: false, DidSwitch: false}
		noswitchPackageFrame = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}, Func: &gno.FuncValue{Name: "foo"}, WithSwitch: false, DidSwitch: false}
		noswitchRealmFrame   = &gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}, Func: &gno.FuncValue{Name: "foo"}, WithSwitch: false, DidSwitch: false}
	)
	tests := []struct {
		name                 string
		machine              *gno.Machine
		expectedAddr         crypto.Bech32Address
		expectedPkgPath      string
		expectedIsOriginCall bool
	}{
		//{
		//	name: "no frames",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames:  []*gno.Frame{},
		//	},
		//	expectedAddr:         user,
		//	expectedPkgPath:      "",
		//	expectedIsOriginCall: false,
		//},
		//{
		//	name: "one frame w/o LastPackage",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames: []*gno.Frame{
		//			{LastPackage: nil},
		//		},
		//	},
		//	expectedAddr:         user,
		//	expectedPkgPath:      "",
		//	expectedIsOriginCall: false,
		//},
		//{
		//	name: "one package frame",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames: []*gno.Frame{
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
		//		},
		//	},
		//	expectedAddr:         user,
		//	expectedPkgPath:      "",
		//	expectedIsOriginCall: false,
		//},
		{
			name: "one realm frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					withswitchFrame,
					switchrealmFrame,
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
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
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: true,
		},
		//{
		//	name: "one msgRun frame",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames: []*gno.Frame{
		//			msgRunFrame,
		//		},
		//	},
		//	expectedAddr:         user,
		//	expectedPkgPath:      "",
		//	expectedIsOriginCall: false,
		//},
		{
			name: "one package frame and one msgCall frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgCallFrame,
					noswitchPackageFrame,
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: true,
		},
		{
			name: "one realm frame and one msgCall frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []*gno.Frame{
					msgCallFrame,
					noswitchRealmFrame,
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: true,
		},
		//{
		//	name: "one package frame and one msgRun frame",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames: []*gno.Frame{
		//			msgRunFrame,
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
		//		},
		//	},
		//	expectedAddr:         user,
		//	expectedPkgPath:      "",
		//	expectedIsOriginCall: false,
		//},
		//{
		//	name: "one realm frame and one msgRun frame",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames: []*gno.Frame{
		//			msgRunFrame,
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
		//		},
		//	},
		//	expectedAddr:         user,
		//	expectedPkgPath:      "",
		//	expectedIsOriginCall: false,
		//},
		//{
		//	name: "multiple frames with one realm",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames: []*gno.Frame{
		//			msgCallFrame,
		//			withswitchFrame,
		//			switchrealmFrame,
		//		},
		//	},
		//	expectedAddr:         user,
		//	expectedPkgPath:      "",
		//	expectedIsOriginCall: false,
		//},
		//{
		//	name: "multiple frames with multiple realms",
		//	machine: &gno.Machine{
		//		Context: ctx,
		//		Frames: []*gno.Frame{
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/zzz"}},
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/zzz"}},
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/yyy"}},
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/yyy"}},
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
		//			{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
		//		},
		//	},
		//	expectedAddr:         gno.DerivePkgAddr("gno.land/r/yyy").Bech32(),
		//	expectedPkgPath:      "gno.land/r/yyy",
		//	expectedIsOriginCall: false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			addr, pkgPath := X_getRealm(tt.machine, 1)
			isOrigin := isOriginCall(tt.machine)

			assert.Equal(string(tt.expectedAddr), addr)
			assert.Equal(tt.expectedPkgPath, pkgPath)
			assert.Equal(tt.expectedIsOriginCall, isOrigin)
		})
	}
}
