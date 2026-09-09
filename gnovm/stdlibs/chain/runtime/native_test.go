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

// newObjectIDMachine returns a machine whose realm is a fresh persistent realm
// at time 1 (an owner was finalized there), plus that owner and the store
// transaction, so the caller can create objects and finalize them.
func newObjectIDMachine(t *testing.T, pkgPath string) (*gno.Machine, gno.TransactionStore, *gno.Allocator, gno.Object) {
	t.Helper()

	baseStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	store := gno.NewStore(gno.NewAllocator(math.MaxInt64), baseStore, iavlStore)
	rlm := gno.NewRealm(pkgPath)
	rlm.Time = 1
	store.SetPackageRealm(rlm)

	tx := store.BeginTransaction(baseStore.CacheWrap(), iavlStore.CacheWrap(), nil, nil)
	m := gno.NewMachineWithOptions(gno.MachineOptions{Store: tx})
	m.Realm = tx.GetPackageRealm(pkgPath)

	alloc := gno.NewAllocator(math.MaxInt64)
	owner := alloc.NewStruct(nil, nil)
	owner.SetPkgID(m.Realm.ID)
	owner.SetNewTime(1)
	return m, tx, alloc, owner
}

// newOwnedObject allocates an object under m's realm, owned by owner, ready to
// be marked new-real.
func newOwnedObject(m *gno.Machine, alloc *gno.Allocator, owner gno.Object) gno.Object {
	oo := alloc.NewStruct(nil, nil)
	oo.SetPkgID(m.Realm.ID)
	oo.SetOwner(owner)
	oo.IncRefCount()
	return oo
}

func TestObjectAddress(t *testing.T) {
	m, tx, alloc, owner := newObjectIDMachine(t, "gno.land/r/demo/objectid")
	object := newOwnedObject(m, alloc, owner)
	tv := gno.TypedValue{V: object}

	tests := []struct {
		name string
		// finalize runs before the assertion, if set.
		finalize bool
		want     func() string
	}{
		{
			// NewTime is stamped at finalization, so an object the running
			// call created has no ID and no address to derive.
			name: "unstamped object has no address",
			want: func() string { return "" },
		},
		{
			// Reading is not issuing: it must not advance the realm clock, so
			// a second read is identical.
			name: "reading again does not stamp it",
			want: func() string { return "" },
		},
		{
			name:     "finalized object derives its own address",
			finalize: true,
			want:     func() string { return gno.DeriveObjectIDCryptoAddr(object.GetObjectID()).String() },
		},
		{
			name: "and keeps deriving the same one",
			want: func() string { return gno.DeriveObjectIDCryptoAddr(object.GetObjectID()).String() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeBefore := m.Realm.Time
			if tt.finalize {
				m.Realm.MarkNewReal(object)
				m.Realm.FinalizeRealmTransaction(tx)
				timeBefore = m.Realm.Time
			}
			require.Equal(t, tt.want(), X_objectAddress(m, tv))
			require.Equal(t, timeBefore, m.Realm.Time, "objectAddress must not advance the realm clock")
		})
	}
}

// Two objects finalized by the same pass take distinct ticks of the realm
// clock, so their addresses differ — which is what makes one usable as an
// identifier for the object that carries it.
func TestObjectAddressIsUniquePerObject(t *testing.T) {
	m, tx, alloc, owner := newObjectIDMachine(t, "gno.land/r/demo/objectid_unique")

	objects := make([]gno.Object, 0, 2)
	for range 2 {
		oo := newOwnedObject(m, alloc, owner)
		m.Realm.MarkNewReal(oo)
		objects = append(objects, oo)
	}
	m.Realm.FinalizeRealmTransaction(tx)

	seen := make(map[string]int, len(objects))
	for i, oo := range objects {
		addr := X_objectAddress(m, gno.TypedValue{V: oo})
		require.NotEmpty(t, addr)
		require.NotContains(t, seen, addr, "objects %d and %d share an address", seen[addr], i)
		seen[addr] = i
	}
}

func TestObjectAddressRejectsValuesWithoutIdentity(t *testing.T) {
	m := gno.NewMachineWithOptions(gno.MachineOptions{})

	tests := []struct {
		name string
		tv   gno.TypedValue
	}{
		{name: "nil interface"},
		{name: "primitive", tv: typedString("not an object")},
		{name: "nil pointer", tv: gno.TypedValue{V: gno.PointerValue{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Panics(t, func() { X_objectAddress(m, tt.tv) })
		})
	}
}
