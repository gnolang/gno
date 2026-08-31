package runtime

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/chain/runtime/unsafe"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestNewRealmID(t *testing.T) {
	baseStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	store := gno.NewStore(gno.NewAllocator(math.MaxInt64), baseStore, iavlStore)
	pkgPath := "gno.land/r/demo/realm_id"
	store.SetPackageRealm(gno.NewRealm(pkgPath))

	baseTx := baseStore.CacheWrap()
	iavlTx := iavlStore.CacheWrap()
	tx := store.BeginTransaction(baseTx, iavlTx, nil, nil)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Store:   tx,
		Context: execctx.ExecContext{RealmIDEnabled: true},
	})
	m.Realm = tx.GetPackageRealm(pkgPath)

	first := NewRealmID(m)
	second := NewRealmID(m)
	require.Equal(t, "gno:realm-id:0510954b76375c303d02cacf694f6135b1adc474:1", first)
	require.Equal(t, "gno:realm-id:0510954b76375c303d02cacf694f6135b1adc474:2", second)
	require.NotEqual(t, first, second)
	var oid gno.ObjectID
	require.Error(t, oid.UnmarshalAmino(first))

	tx.Write()
	baseTx.Write()
	iavlTx.Write()

	fresh := gno.NewStore(gno.NewAllocator(math.MaxInt64), baseStore, iavlStore)
	require.Equal(t, uint64(2), fresh.GetPackageRealm(pkgPath).Time)

	baseTx = baseStore.CacheWrap()
	iavlTx = iavlStore.CacheWrap()
	tx = fresh.BeginTransaction(baseTx, iavlTx, nil, nil)
	m = gno.NewMachineWithOptions(gno.MachineOptions{
		Store:   tx,
		Context: execctx.ExecContext{RealmIDEnabled: true},
	})
	m.Realm = tx.GetPackageRealm(pkgPath)
	require.Equal(t, "gno:realm-id:0510954b76375c303d02cacf694f6135b1adc474:3", NewRealmID(m))
}

func TestNewRealmIDPanicsInQueryContext(t *testing.T) {
	pkgPath := "gno.land/r/demo/realm_id"
	m := &gno.Machine{
		Realm:   gno.NewRealm(pkgPath),
		Context: execctx.ExecContext{},
	}
	require.Panics(t, func() { NewRealmID(m) })
}

func TestNewRealmIDRejectsInvalidRealm(t *testing.T) {
	tests := []struct {
		name     string
		realm    *gno.Realm
		wantTime uint64
	}{
		{name: "nil realm"},
		{name: "package", realm: gno.NewRealm("gno.land/p/demo/realm_id")},
		{name: "ephemeral", realm: gno.NewRealm("gno.land/e/g1user/run")},
		{name: "overflow", realm: &gno.Realm{Path: "gno.land/r/demo/realm_id", Time: math.MaxUint64}, wantTime: math.MaxUint64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &gno.Machine{
				Realm:   tt.realm,
				Context: execctx.ExecContext{RealmIDEnabled: true},
			}
			require.Panics(t, func() { NewRealmID(m) })
			if tt.realm != nil {
				require.Equal(t, tt.wantTime, tt.realm.Time)
			}
		})
	}
}

func TestPreviousRealmIsOrigin(t *testing.T) {
	var (
		user = gno.DerivePkgBech32Addr("user1.gno")
		ctx  = execctx.ExecContext{
			OriginCaller: user,
		}
		msgCallFrame = gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "main"}}
		msgRunFrame  = gno.Frame{LastPackage: &gno.PackageValue{PkgPath: "gno.land/e/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/run"}}
	)
	type expectations struct {
		addr         crypto.Bech32Address
		pkgPath      string
		isOriginCall bool
		doesPanic    bool
	}
	tests := []struct {
		name                 string
		machine              *gno.Machine
		expectedAddr         crypto.Bech32Address
		expectedPkgPath      string
		expectedIsOriginCall bool
	}{
		{
			name: "no frames",
			machine: &gno.Machine{
				Context: ctx,
				Frames:  []gno.Frame{},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: false,
		},
		{
			name: "one frame w/o LastPackage",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: nil},
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: false,
		},
		{
			name: "one package frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: false,
		},
		{
			name: "one realm frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
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
				Frames: []gno.Frame{
					msgCallFrame,
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: true,
		},
		{
			name: "one msgRun frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					msgRunFrame,
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: false,
		},
		{
			name: "one package frame and one msgCall frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					msgCallFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
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
				Frames: []gno.Frame{
					msgCallFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: true,
		},
		{
			name: "one package frame and one msgRun frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					msgRunFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/p/xxx"}},
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: false,
		},
		{
			name: "one realm frame and one msgRun frame",
			machine: &gno.Machine{
				Context: ctx,
				Frames: []gno.Frame{
					msgRunFrame,
					{LastPackage: &gno.PackageValue{PkgPath: "gno.land/r/xxx"}},
				},
			},
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: false,
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
			expectedAddr:         user,
			expectedPkgPath:      "",
			expectedIsOriginCall: false,
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
			expectedAddr:         gno.DerivePkgBech32Addr("gno.land/r/yyy"),
			expectedPkgPath:      "gno.land/r/yyy",
			expectedIsOriginCall: false,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("fail", i)
				}
			}()
			assert := assert.New(t)

			addr, pkgPath := unsafe.X_getRealm(tt.machine, 1)
			isOrigin := isOriginCall(tt.machine)

			assert.Equal(string(tt.expectedAddr), addr)
			assert.Equal(tt.expectedPkgPath, pkgPath)
			assert.Equal(tt.expectedIsOriginCall, isOrigin)
		})
	}
}
