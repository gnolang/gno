package runtime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/chain/runtime/unsafe"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/crypto"
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

// TestObjectID covers the shapes the native has to tell apart. The fixture
// builds what the VM actually hands over — a pointer to a standalone heap
// item — rather than a bare value the machine never produces.
func TestObjectID(t *testing.T) {
	const (
		realmPath = "gno.land/r/x/objectidtest"
		otherPath = "gno.land/r/x/objectidother"
		purePath  = "gno.land/p/x/objectidpure"
	)

	// ptrTo wraps a heap item the way a Gno pointer to a standalone object
	// does: PointerValue with the heap item itself as Base.
	ptrTo := func(hiv *gno.HeapItemValue) gno.TypedValue {
		return gno.TypedValue{V: gno.PointerValue{TV: &hiv.Value, Base: hiv}}
	}
	heapItem := func(pkgPath string, newTime uint64) *gno.HeapItemValue {
		hiv := gno.NewAllocator(1<<20).NewHeapItem(nil, gno.TypedValue{})
		hiv.SetPkgID(gno.PkgIDFromPkgPath(pkgPath))
		if newTime != 0 {
			hiv.SetNewTime(newTime)
		}
		return hiv
	}

	stamped := heapItem(realmPath, 7)
	wantStamped := stamped.GetObjectID().String()

	tests := []struct {
		name string
		tv   gno.TypedValue
		want string
	}{
		{
			name: "persisted realm object",
			tv:   ptrTo(stamped),
			want: wantStamped,
		},
		{
			// The realm half is stamped at allocation, the clock half only
			// when the owning realm persists the object. Half an ID names
			// nothing, so the caller is told to ask again later.
			name: "created by the running call",
			tv:   ptrTo(heapItem(realmPath, 0)),
			want: "",
		},
		{
			// Immutable packages are outside what #6026 is about and
			// reporting "" keeps the question open; widening is compatible.
			name: "owned by a p/ package",
			tv:   ptrTo(heapItem(purePath, 3)),
			want: "",
		},
		{
			// A pointer into a struct field resolves to the container, and
			// every sibling field would report the container's identity.
			name: "pointer into a struct",
			tv: gno.TypedValue{V: gno.PointerValue{
				TV:   &gno.TypedValue{},
				Base: &gno.StructValue{},
			}},
			want: "",
		},
		{
			name: "nil pointer",
			tv:   gno.TypedValue{V: gno.PointerValue{}},
			want: "",
		},
		{
			name: "nil interface",
			tv:   gno.TypedValue{},
			want: "",
		},
		{
			name: "string",
			tv:   gno.TypedValue{T: gno.StringType},
			want: "",
		},
	}

	m := &gno.Machine{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ObjectID(m, tt.tv))
		})
	}

	// Reading is pure: the same object answers the same thing every time.
	assert.Equal(t, wantStamped, ObjectID(m, ptrTo(stamped)))

	// What #6026 needs: two objects one realm minted under one name still
	// differ, and no realm can reach across to another realm's answers.
	second := heapItem(realmPath, 8)
	foreign := heapItem(otherPath, 7)
	assert.NotEqual(t, ObjectID(m, ptrTo(stamped)), ObjectID(m, ptrTo(second)))
	assert.NotEqual(t, ObjectID(m, ptrTo(stamped)), ObjectID(m, ptrTo(foreign)))
}
