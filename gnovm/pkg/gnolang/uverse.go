package gnolang

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	// NativeCPUUversePrintInit is the base gas cost for the Print function.
	// The actual cost is 1800, but we subtract OpCPUCallNativeBody (424), resulting in 1376.
	NativeCPUUversePrintInit = 1376
	// NativeCPUUversePrintPerChar is now chars per gas unit.
	NativeCPUUversePrintCharsPerGas = 10
)

// ----------------------------------------
// non-primitive builtin types

var gErrorType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    "error",
	Base: &InterfaceType{
		PkgPath: uversePkgPath,
		Methods: []FieldType{
			{
				Name: "Error",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{
						{
							// Name: "",
							Type: StringType,
						},
					},
				},
			},
		},
	},
	sealed: true,
}

// IsErrorType returns true if the given type implements the error interface.
// This is useful for checking function return types without a TypedValue.
func IsErrorType(t Type) bool {
	if t == nil {
		return false
	}
	return IsImplementedBy(gErrorType, t)
}

var gStringerType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    "stringer",
	Base: &InterfaceType{
		PkgPath: uversePkgPath,
		Methods: []FieldType{
			{
				Name: "String",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{
						{
							// Name: "",
							Type: StringType,
						},
					},
				},
			},
		},
	},
	sealed: true,
}

var gAddressType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    "address",
	Base:    StringType,
	sealed:  true,
	// methods defined in makeUverseNode()
}

var gRealmType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    "realm",
	Base: &InterfaceType{
		PkgPath: uversePkgPath,
		Methods: []FieldType{
			{
				Name: "Address",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{{
						Type: gAddressType,
					}},
				},
			}, {
				Name: "PkgPath",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{{
						Type: StringType,
					}},
				},
			}, { // gets filled in init() below.
				Name: "Previous",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{{
						Type: nil,
					}},
				},
			}, {
				Name: "IsCode",
				Type: &FuncType{
					Params:  nil,
					Results: []FieldType{{Type: BoolType}},
				},
			}, {
				Name: "IsUser",
				Type: &FuncType{
					Params:  nil,
					Results: []FieldType{{Type: BoolType}},
				},
			}, {
				Name: "IsUserCall",
				Type: &FuncType{
					Params:  nil,
					Results: []FieldType{{Type: BoolType}},
				},
			}, {
				Name: "IsUserRun",
				Type: &FuncType{
					Params:  nil,
					Results: []FieldType{{Type: BoolType}},
				},
			}, {
				Name: "IsEphemeral",
				Type: &FuncType{
					Params:  nil,
					Results: []FieldType{{Type: BoolType}},
				},
			}, {
				Name: "IsCurrent",
				Type: &FuncType{
					Params:  nil,
					Results: []FieldType{{Type: BoolType}},
				},
			}, {
				Name: "String",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{{
						Type: StringType,
					}},
				},
			},
		},
	},
	sealed: true,
}

var gConcreteRealmType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    ".grealm",
	Base: &StructType{
		PkgPath: uversePkgPath,
		Fields: []FieldType{
			{Name: "addr", Type: gAddressType},
			{Name: "pkgPath", Type: StringType},
			// prev is *.grealm (PointerKind) so equality is identity
			// and the chain terminates cleanly at a nil PointerValue.
			// Type pointer-patched in init() below once
			// gConcreteRealmPtrType is visible.
			{Name: "prev", Type: nil},
		},
	},
	sealed: true,
	// methods defined in makeUverseNode()
}

// Singleton pointer type for *.grealm. Allocated once so TypeID memoization
// is stable across the realm machinery.
var gConcreteRealmPtrType = &PointerType{Elt: gConcreteRealmType}

// newRealmHIVPointer builds a *.grealm captured-realm TypedValue from
// (addr, pkgPath, prevField). All callers that produce a realm value go
// through this helper to keep the HIV+PointerValue construction in one
// place. Pass alloc=nil to skip allocator charging (used for the global
// placeholder origin built at package init).
func newRealmHIVPointer(alloc *Allocator, addr, pkgPath string, prevField TypedValue) TypedValue {
	// Realm-handle values are ephemeral and forbidden from
	// persistence (see refusePersistRealmHIV). They never reach
	// assignNewObjectID or saveObject — so deliberately don't
	// stamp PkgID here. Any attempt to persist them is caught by
	// the refuse-persist guard with a clearer error than what
	// cross-realm finalize would produce.
	sv := &StructValue{Fields: []TypedValue{
		{T: gAddressType, V: StringValue(addr)},
		{T: StringType, V: StringValue(pkgPath)},
		prevField,
	}}
	var hiv *HeapItemValue
	if alloc == nil {
		hiv = &HeapItemValue{Value: TypedValue{T: gConcreteRealmType, V: sv}}
	} else {
		// alloc.NewHeapItem would stamp PkgID = currentRealmID,
		// which routes through cross-realm finalize. Bypass it.
		alloc.AllocateHeapItem()
		hiv = &HeapItemValue{Value: TypedValue{T: gConcreteRealmType, V: sv}}
	}
	return TypedValue{
		T: gConcreteRealmPtrType,
		V: PointerValue{TV: &hiv.Value, Base: hiv, Index: 0},
	}
}

// gOriginRealmHIV / gOriginRealmTV is the preprocess-time placeholder
// origin realm. It stands in for the per-tx EOA realm value during
// preprocess (when ctx.OriginCaller is not yet known) and is replaced
// at runtime entry by buildOriginRealm, which builds a real per-tx
// origin realm with addr=OriginCaller, pkgPath="", prev=truly-nil.
//
// Non-nil pointer (with empty fields) rather than a typed-nil: matches
// runtime.PreviousRealm()'s shape for EOA callers (a zero Realm struct,
// not a nil). Structural shape (Value.T == gConcreteRealmType AND prev
// field has V == nil) is the persistent marker — see isOriginRealmHIV.
var (
	gOriginRealmHIV *HeapItemValue
	gOriginRealmTV  TypedValue
)

// isOriginRealmHIV reports whether hiv is an origin / EOA-shaped realm
// value (prev field truly nil). The shape persists across AST serialize/
// load — both the preprocess placeholder and the per-tx origin satisfy
// the predicate. All realm values are forbidden from persistence (see
// errPersistRealm / refusePersistRealmHIV); origin realms are exempted
// from the panic only because they're the shared chain-root marker that
// captured curs transitively reference, not a user-stashable value.
func isOriginRealmHIV(hiv *HeapItemValue) bool {
	if hiv == nil || hiv.Value.T != gConcreteRealmType {
		return false
	}
	sv, ok := hiv.Value.V.(*StructValue)
	if !ok || len(sv.Fields) < 3 {
		return false
	}
	return sv.Fields[2].V == nil
}

// errPersistRealm is the shared panic message for realm-value persistence
// refusals — kept as a const so the source of the string is single, and
// filetests can match it verbatim.
const errPersistRealm = "cannot persist realm value: realm values are ephemeral and tied to a call frame"

// refusePersistRealmHIV panics with errPersistRealm if hiv is a
// non-origin realm value. Used at every save/attach hook to keep the
// guard logic in one place.
func refusePersistRealmHIV(hiv *HeapItemValue) {
	if hiv == nil || isOriginRealmHIV(hiv) {
		return
	}
	if t := hiv.Value.T; t == gConcreteRealmType || t == gConcreteRealmPtrType {
		panic(errPersistRealm)
	}
}

// OriginCallerExtractor is set by the execctx package's init() so this
// package can read ctx.OriginCaller without importing execctx (which would
// be a cycle). Returns "" when no caller can be extracted.
var OriginCallerExtractor func(ctx any) string

// BuildOverridePrevField constructs a prev-field TypedValue for a captured
// realm whose addr/pkgPath have been overridden by testing.SetRealm with
// a CodeRealm. The prev carries the underlying-frame realm so that
// `cur.Previous()` after the override surfaces what X_getRealm surfaces
// as PreviousRealm of the override frame. Exposed for X_setContext.
func BuildOverridePrevField(addr, pkgPath string) TypedValue {
	return newRealmHIVPointer(nil, addr, pkgPath, TypedValue{})
}

// buildOriginRealm constructs a per-call origin realm matching what
// runtime.PreviousRealm() (via GetRealm) returns at the chain root for
// the same execution context: addr=OriginCaller; prev=truly-nil; pkgPath
// is "" for MsgCall / QueryEval / AddPkg, and is the /e/<addr>/run path
// for MsgRun. The MsgRun case is the one that keeps cur.Previous() in
// agreement with runtime.PreviousRealm() inside callees of `/e/` main —
// closing the IsUserCall() spoof gap. Fresh per call because OriginCaller
// can change between init() and main() within the same Machine (the test
// framework sets it after RunMemPackage but before RunMainMaybeCrossing),
// so a cached origin built at init time would be stale when main runs.
func buildOriginRealm(m *Machine) TypedValue {
	var addr string
	if OriginCallerExtractor != nil && m.Context != nil {
		addr = OriginCallerExtractor(m.Context)
	}
	var pkgPath string
	if len(m.Frames) > 0 {
		bp := m.Frames[0].LastPackage
		if bp != nil && IsEphemeralPath(bp.PkgPath) {
			pkgPath = bp.PkgPath
		}
	}
	return newRealmHIVPointer(nil, addr, pkgPath, TypedValue{})
}

// NOTE: this init() must run before makeUverseNode() (called from the init
// at the bottom of this file) so type-id memoization sees the patched
// prev-field type.
func init() {
	gRealmPrevious := gRealmType.Base.(*InterfaceType).GetMethodFieldType("Previous")
	gRealmPrevious.Type.(*FuncType).Results[0].Type = gRealmType

	// Patch the prev field's type (forward reference; see field 2 above).
	gConcreteRealmType.Base.(*StructType).Fields[2].Type = gConcreteRealmPtrType

	// Build the global placeholder origin realm now that types are wired.
	gOriginRealmTV = newRealmHIVPointer(nil, "", "", TypedValue{})
	gOriginRealmHIV = gOriginRealmTV.V.(PointerValue).Base.(*HeapItemValue)
}

// OriginRealmTV returns the typed-nil *.grealm used as the prev seed for
// realm captures at the top of the cross-call chain. Exposed for callers
// outside this package (e.g., pkg/test/test.go).
func OriginRealmTV() TypedValue { return gOriginRealmTV }

// NewOriginRealmTV builds a FRESH origin-shape realm value (addr="",
// pkgPath="", prev=truly-nil) backed by a brand-new *HeapItemValue. Use
// this when the caller intends to write to the cur (e.g. test frames
// that may receive testing.SetRealm overrides) — mutating the
// gOriginRealmTV-backed struct in place would corrupt the global
// placeholder for every subsequent caller. The shape still satisfies
// isOriginRealmHIV (prev truly-nil), so persistence guards continue
// to exempt it.
func NewOriginRealmTV(alloc *Allocator) TypedValue {
	return newRealmHIVPointer(alloc, "", "", TypedValue{})
}

// NewConcreteRealm builds a captured realm value as a pointer-typed
// TypedValue. Equality (PointerKind ==) is pointer-identity: two captured
// curs compare equal iff they reference the same *HeapItemValue — i.e.,
// the same cross-call snapshot. This is what makes a stashed cur
// unforgeable. Persistence of these values is forbidden.
func NewConcreteRealm(alloc *Allocator, pkgPath string, prev TypedValue) TypedValue {
	prevField := gOriginRealmTV
	if pv, ok := prev.V.(PointerValue); ok && pv.TV != nil {
		prevField = TypedValue{T: gConcreteRealmPtrType, V: pv}
	}
	return newRealmHIVPointer(alloc, string(DerivePkgBech32Addr(pkgPath)), pkgPath, prevField)
}

// MakeRealmValue builds a captured realm value with the given addr,
// pkgPath, and prev. Unlike NewConcreteRealm, addr is taken verbatim
// (not derived from pkgPath) so callers can construct UserRealm-shaped
// values (pkgPath="") with arbitrary addresses. Used by the testing
// stdlib to expose an explicit cur-value constructor — see X_makeRealm.
func MakeRealmValue(alloc *Allocator, addr, pkgPath string, prev TypedValue) TypedValue {
	prevField := gOriginRealmTV
	if pv, ok := prev.V.(PointerValue); ok && pv.TV != nil {
		prevField = TypedValue{T: gConcreteRealmPtrType, V: pv}
	}
	return newRealmHIVPointer(alloc, addr, pkgPath, prevField)
}

// derefRealmStruct unwraps a realm TypedValue (pointer-typed *.grealm, or
// value-receiver form) to its underlying *StructValue. Returns nil when
// the TypedValue's shape doesn't match (no PointerValue/StructValue), so
// callers that don't know the shape can bail safely.
func derefRealmStruct(tv *TypedValue) *StructValue {
	if pv, ok := tv.V.(PointerValue); ok {
		if pv.TV == nil {
			return nil
		}
		sv, _ := pv.TV.V.(*StructValue)
		return sv
	}
	sv, _ := tv.V.(*StructValue)
	return sv
}

// realmHIV extracts the underlying *HeapItemValue from a realm TypedValue.
// Since all .grealm methods are pointer-receiver (see DefineNativePtrMethod
// usage in makeUverseNode), the outer PointerValue+HIV wrapper survives
// dispatch and HIV identity is always available.
func realmHIV(tv *TypedValue) *HeapItemValue {
	pv, ok := tv.V.(PointerValue)
	if !ok {
		return nil
	}
	hiv, _ := pv.Base.(*HeapItemValue)
	return hiv
}

// realmIsCurrentOnMachine reports whether tv is the topmost crossing
// frame's Cur, by HIV pointer identity. Used by both .grealm.IsCurrent
// (informational) and installCrossingCur's cross path (authority) —
// these share a single semantic: rlm.IsCurrent() ⇔ cross(rlm) accepts.
//
// Pointer-receiver method dispatch preserves the outer PointerValue, so
// recvHIV is always populated for any realm value the language allows
// to flow into a method or function call. A nil HIV here means tv isn't
// a real realm value (e.g., zero-value/uninitialized) and the check
// rejects.
func realmIsCurrentOnMachine(m *Machine, tv *TypedValue) bool {
	recvHIV := realmHIV(tv)
	if recvHIV == nil {
		return false
	}
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if !fr.IsCall() {
			continue
		}
		if !(fr.WithCross || fr.DidCrossing) {
			continue
		}
		if fr.Cur.T == nil {
			continue
		}
		return realmHIV(&fr.Cur) == recvHIV
	}
	return false
}

// realmIsEphemeral reports whether pkgPath matches the ephemeral pattern
// "domain/e/...". Mirrors chain/runtime.Realm.IsEphemeral.
func realmIsEphemeral(pkgPath string) bool {
	if pkgPath == "" {
		return false
	}
	idx := strings.Index(pkgPath, "/e/")
	if idx == -1 || len(pkgPath) <= idx+3 {
		return false
	}
	// Domain segment must not itself contain a '/'.
	return strings.Index(pkgPath[:idx], "/") == -1
}

// realmIsUserRun reports whether (addr, pkgPath) represents a user-run
// ephemeral realm: pkgPath == "<domain>/e/<addr>/run". Mirrors
// chain/runtime.Realm.IsUserRun.
func realmIsUserRun(addr, pkgPath string) bool {
	idx := strings.Index(pkgPath, "/")
	if idx == -1 {
		return false
	}
	return pkgPath == pkgPath[:idx]+"/e/"+addr+"/run"
}

// ----------------------------------------
// Uverse package

var (
	uverseNode  *PackageNode
	uverseValue *PackageValue
	uverseInit  = uverseUninitialized
)

const (
	uverseUninitialized = iota
	uverseInitializing
	uverseInitialized
)

func init() {
	// Skip Uverse init during benchmarking to load stdlibs in the benchmark main function.
	if !bm.Enabled {
		// Call Uverse() so we initialize the Uverse node ahead of any calls to the package.
		Uverse()
	}
}

const uversePkgPath = ".uverse"

// UverseNode returns the uverse PackageValue.
// If called while initializing the UverseNode itself, it will return an empty
// PackageValue.
func Uverse() *PackageValue {
	switch uverseInit {
	case uverseUninitialized:
		uverseInit = uverseInitializing
		makeUverseNode()
		uverseInit = uverseInitialized
	case uverseInitializing:
		return &PackageValue{}
	}

	return uverseValue
}

// UverseNode returns the uverse PackageNode.
// If called while initializing the UverseNode itself, it will return an empty
// PackageNode.
func UverseNode() *PackageNode {
	switch uverseInit {
	case uverseUninitialized:
		uverseInit = uverseInitializing
		makeUverseNode()
		uverseInit = uverseInitialized
	case uverseInitializing:
		// Return an empty stub; set location so debug code
		// can identify it as the uverse package.
		pn := &PackageNode{}
		pn.SetLocation(PackageNodeLocation(uversePkgPath))
		return pn
	}

	return uverseNode
}

func makeUverseNode() {
	// NOTE: uverse node is hidden, thus the leading dot in pkgPath=".uverse".
	uverseNode = NewPackageNode("uverse", uversePkgPath, nil)

	// temporary convenience functions.
	def := func(n Name, tv TypedValue) {
		uverseNode.Define2(true, n, tv.T, tv, NameSource{})
	}
	defNative := uverseNode.DefineNative
	defNativeMethod := uverseNode.DefineNativeMethod
	defNativePtrMethod := uverseNode.DefineNativePtrMethod

	// Primitive types
	undefined := TypedValue{}
	def("._", undefined)   // special, path is zero.
	def("iota", undefined) // special
	def("nil", undefined)
	def("bool", asValue(BoolType))
	def("byte", asValue(Uint8Type))
	def("float32", asValue(Float32Type))
	def("float64", asValue(Float64Type))
	def("int", asValue(IntType))
	def("int8", asValue(Int8Type))
	def("int16", asValue(Int16Type))
	def("int32", asValue(Int32Type))
	def("int64", asValue(Int64Type))
	def("rune", asValue(Int32Type))
	def("string", asValue(StringType))
	def("uint", asValue(UintType))
	def("uint8", asValue(Uint8Type))
	def("uint16", asValue(Uint16Type))
	def("uint32", asValue(Uint32Type))
	def("uint64", asValue(Uint64Type))
	def("error", asValue(gErrorType))
	def("any", asValue(&InterfaceType{}))

	// Values
	def("true", untypedBool(true))
	def("false", untypedBool(false))

	// Functions
	defNative("append",
		Flds( // params
			"x", GenT("X", nil), // args[0]
			"args", Vrd(GenT("X.Elem()", nil)), // args[1]
		),
		Flds( // results
			"res", GenT("X", nil), // res
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2(m.Store)
			// As a special case, if arg1 is a string type, first convert it into
			// a data slice type.
			if arg1.TV.T != nil && arg1.TV.T.Kind() == StringKind {
				arg1String := arg1.TV.GetString()
				// NOTE: this hack works because
				// arg1 PointerValue is not a pointer,
				// so the modification here is only local.
				newArrayValue := m.Alloc.NewDataArray(nil, len(arg1String))
				m.incrCPU(OpCPUSlopeCopyPrimitive * int64(len(arg1String)))
				copy(newArrayValue.Data, []byte(arg1String))
				arg1.TV = &TypedValue{
					T: m.Alloc.NewType(&SliceType{ // TODO: reuse
						Elt: Uint8Type,
						Vrd: true,
					}),
					V: m.Alloc.NewSlice(newArrayValue, 0, len(arg1String), len(arg1String)), // TODO: pool?
				}
			}
			arg0Type := arg0.TV.T
			arg1Type := arg1.TV.T
			switch arg0Value := arg0.TV.V.(type) {
			// ----------------------------------------------------------------
			// append(nil, ???)
			case nil:
				switch arg1Value := arg1.TV.V.(type) {
				// ------------------------------------------------------------
				// append(nil, nil)
				case nil: // no change
					m.PushValue(TypedValue{
						T: arg0Type,
						V: nil,
					})
					return

				// ------------------------------------------------------------
				// append(nil, *SliceValue)
				case *SliceValue:
					arg1Length := arg1Value.Length
					arg1Offset := arg1Value.Offset
					arg1Base := arg1Value.GetBase(m.Store)
					arg1EndIndex := arg1Offset + arg1Length

					if arg1Length == 0 { // no change
						m.PushValue(TypedValue{
							T: arg0Type,
							V: nil,
						})
						return
					} else if arg0Type.Elem().Kind() == Uint8Kind {
						// append(nil, *SliceValue) new data bytes ---
						arrayValue := m.Alloc.NewDataArray(nil, arg1Length)
						m.incrCPU(OpCPUSlopeCopyPrimitive * int64(arg1Length))
						if arg1Base.Data == nil {
							copyListToData(
								arrayValue.Data[:arg1Length],
								arg1Base.List[arg1Offset:arg1EndIndex])
						} else {
							copy(
								arrayValue.Data[:arg1Length],
								arg1Base.Data[arg1Offset:arg1EndIndex])
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSlice(arrayValue, 0, arg1Length, arg1Length),
						})
						return
					} else {
						// append(nil, *SliceValue) new list ---------
						arrayValue := m.Alloc.NewListArray(nil, arg1Length)
						if arg1Length > 0 {
							m.incrCPU(OpCPUSlopeCopyElement * int64(arg1Length))
							for i := range arg1Length {
								arrayValue.List[i] = arg1Base.List[arg1Offset+i].unrefCopy(m.Alloc, m.Store)
							}
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSlice(arrayValue, 0, arg1Length, arg1Length),
						})
						return
					}
				default:
					panic("should not happen")
				}

			// ----------------------------------------------------------------
			// append(*SliceValue, ???)
			case *SliceValue:
				arg0Length := arg0Value.Length
				arg0Offset := arg0Value.Offset
				arg0Capacity := arg0Value.Maxcap
				arg0Base := arg0Value.GetBase(m.Store)
				// NOTE, ANY MODIFICATION TO arg0 SHOULD ALWAYS CALL
				// m.Realm.DidUpdate(m, arg0Base, nil, nil) FIRST TO CHECK WRITE PERMISSIONS.
				switch arg1Value := arg1.TV.V.(type) {
				// ------------------------------------------------------------
				// append(*SliceValue, nil)
				case nil: // no change
					m.PushValue(TypedValue{
						T: arg0Type,
						V: arg0Value,
					})
					return

				// ------------------------------------------------------------
				// append(*SliceValue, *SliceValue)
				case *SliceValue:
					arg1Length := arg1Value.Length
					arg1Offset := arg1Value.Offset
					arg1Base := arg1Value.GetBase(m.Store)
					if arg0Length+arg1Length <= arg0Capacity {
						// append(*SliceValue, *SliceValue) w/i capacity -----
						if 0 < arg1Length { // implies 0 < xvc
							// DEFENSIVE: in this case, we're writing data directly
							// into the backing array of arg0. Ensure we can write
							// to it.
							if m.IsReadonly(arg0.TV) {
								m.Panic(typedString("cannot append to readonly tainted slice"))
							}

							if arg0Base.Data == nil {
								// append(*SliceValue.List, *SliceValue) ---------
								// Per-element DidUpdate calls below are sufficient
								// to mark arg0Base dirty; no top-level call needed.
								list := arg0Base.List
								if arg1Base.Data == nil {
									m.incrCPU(OpCPUSlopeCopyElement * int64(arg1Length))
									dstStart := arg0Offset + arg0Length
									srcStart := arg1Offset
									srcEnd := arg1Offset + arg1Length

									step := 1
									start := 0
									end := arg1Length
									// Overlap-safe copy: copy backward when dst starts after src to avoid clobbering.
									requiresBackwardCopy := arg0Base == arg1Base && dstStart > srcStart && dstStart < srcEnd
									if requiresBackwardCopy {
										step = -1
										start = arg1Length - 1
										end = -1
									}
									for i := start; i != end; i += step {
										oldElem := list[dstStart+i]
										// unrefCopy will resolve references and copy their values
										// to copy by value rather than by reference.
										newElem := arg1Base.List[arg1Offset+i].unrefCopy(m.Alloc, m.Store)
										list[dstStart+i] = newElem

										m.Realm.DidUpdate(m, 
											arg0Base,
											oldElem.GetFirstObject(m.Store),
											newElem.GetFirstObject(m.Store),
										)
									}
								} else {
									m.incrCPU(OpCPUSlopeCopyPrimitive * int64(arg1Length))
									copyDataToList(
										list[arg0Offset+arg0Length:arg0Offset+arg0Length+arg1Length],
										arg1Base.Data[arg1Offset:arg1Offset+arg1Length],
										arg0Type.Elem())
								}
							} else {
								// append(*SliceValue.Data, *SliceValue) ---------
								// DidUpdate is required here: raw byte copies do not
								// go through Assign2, so arg0Base would not be marked
								// dirty otherwise.
								m.Realm.DidUpdate(m, arg0Base, nil, nil)
								data := arg0Base.Data
								if arg1Base.Data == nil {
									m.incrCPU(OpCPUSlopeCopyPrimitive * int64(arg1Length))
									copyListToData(
										data[arg0Offset+arg0Length:arg0Offset+arg0Length+arg1Length],
										arg1Base.List[arg1Offset:arg1Offset+arg1Length])
								} else {
									m.incrCPU(OpCPUSlopeCopyPrimitive * int64(arg1Length))
									copy(
										data[arg0Offset+arg0Length:arg0Offset+arg0Length+arg1Length],
										arg1Base.Data[arg1Offset:arg1Offset+arg1Length])
								}
							}
							m.PushValue(TypedValue{
								T: arg0Type,
								V: m.Alloc.NewSlice(arg0Base, arg0Offset, arg0Length+arg1Length, arg0Capacity),
							})
							return
						} else { // no change
							m.PushValue(TypedValue{
								T: arg0Type,
								V: arg0Value,
							})
							return
						}
					} else if arg0Type.Elem().Kind() == Uint8Kind {
						// append(*SliceValue, *SliceValue) new data bytes ---
						newLength := arg0Length + arg1Length
						arrayValue := m.Alloc.NewDataArray(nil, newLength)
						if 0 < arg0Length {
							m.incrCPU(OpCPUSlopeCopyPrimitive * int64(arg0Length))
							if arg0Base.Data == nil {
								copyListToData(
									arrayValue.Data[:arg0Length],
									arg0Base.List[arg0Offset:arg0Offset+arg0Length])
							} else {
								copy(
									arrayValue.Data[:arg0Length],
									arg0Base.Data[arg0Offset:arg0Offset+arg0Length])
							}
						}
						if 0 < arg1Length {
							m.incrCPU(OpCPUSlopeCopyPrimitive * int64(arg1Length))
							if arg1Base.Data == nil {
								copyListToData(
									arrayValue.Data[arg0Length:newLength],
									arg1Base.List[arg1Offset:arg1Offset+arg1Length])
							} else {
								copy(
									arrayValue.Data[arg0Length:newLength],
									arg1Base.Data[arg1Offset:arg1Offset+arg1Length])
							}
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSlice(arrayValue, 0, newLength, newLength),
						})
						return
					} else {
						// append(*SliceValue, *SliceValue) new list ---------
						arrayLen := arg0Length + arg1Length
						arrayValue := m.Alloc.NewListArray(nil, arrayLen)
						if arg0Length > 0 {
							if arg0Base.Data == nil {
								m.incrCPU(OpCPUSlopeCopyElement * int64(arg0Length))
								for i := range arg0Length {
									arrayValue.List[i] = arg0Base.List[arg0Offset+i].unrefCopy(m.Alloc, m.Store)
								}
							} else {
								panic("should not happen")
							}
						}

						if arg1Length > 0 {
							if arg1Base.Data == nil {
								m.incrCPU(OpCPUSlopeCopyElement * int64(arg1Length))
								for i := range arg1Length {
									arrayValue.List[arg0Length+i] = arg1Base.List[arg1Offset+i].unrefCopy(m.Alloc, m.Store)
								}
							} else {
								m.incrCPU(OpCPUSlopeCopyPrimitive * int64(arg1Length))
								copyDataToList(
									arrayValue.List[arg0Length:arg0Length+arg1Length],
									arg1Base.Data[arg1Offset:arg1Offset+arg1Length],
									arg1Type.Elem(),
								)
							}
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSlice(arrayValue, 0, arrayLen, arrayLen),
						})
						return
					}
				// ------------------------------------------------------------
				default:
					panic("should not happen")
				}
			// ----------------------------------------------------------------
			// append(?!!, ???)
			default:
				panic("should not happen")
			}
		},
	)
	defNative("cap",
		Flds( // params
			"x", AnyT(),
		),
		Flds( // results
			"", "int",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(m.Store)
			res0 := TypedValue{
				T: IntType,
				V: nil,
			}
			res0.SetInt(int64(arg0.TV.GetCapacity()))
			m.PushValue(res0)
		},
	)
	defNative("copy",
		Flds( // params
			"dst", GenT("X", nil),
			"src", GenT("Y", nil),
		),
		Flds( // results
			"", "int",
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2(m.Store)
			dst, src := arg0, arg1
			bdt := baseOf(dst.TV.T).(*SliceType)
			switch bst := baseOf(src.TV.T).(type) {
			case PrimitiveType:
				if debug {
					debug.Println("copy(<%s>,<%s>)", bdt.String(), bst.String())
				}
				if bst.Kind() != StringKind {
					panic("should not happen")
				}
				if bdt.Elt != Uint8Type {
					panic("should not happen")
				}
				// NOTE: this implementation is almost identical to the next one.
				// note that in some cases optimization
				// is possible if dstv.Data != nil.
				dstl := dst.TV.GetLength()
				srcl := src.TV.GetLength()
				minl := min(srcl, dstl)
				if minl == 0 {
					// return 0.
					m.PushValue(defaultTypedValue(m.Alloc, IntType))
					return
				}
				dstv := dst.TV.V.(*SliceValue)
				if m.IsReadonly(dst.TV) {
					m.Panic(typedString("cannot copy to readonly tainted slice"))
				}
				dstBase := dstv.GetBase(m.Store)
				// DidUpdate is required here even though Assign2 is called per
				// element below: for byte slices (Data != nil), GetPointerAtIndexInt2
				// returns a DataByteType pointer and Assign2 returns early for that
				// case without calling DidUpdate. The top-level call ensures the
				// backing array is always marked dirty in the realm store.
				m.Realm.DidUpdate(m, dstBase, nil, nil)
				// Assign2 fast-paths DataByteType (values.go:217): just SetDataByte
				// + single DidUpdate. Per-byte cost lands in the Primitive tier.
				m.incrCPU(OpCPUSlopeCopyPrimitive * int64(minl))
				// TODO: consider an optimization if dstv.Data != nil.
				for i := range minl {
					dstev := dstv.GetPointerAtIndexInt2(m.Store, i, bdt.Elt)
					srcev := src.TV.GetPointerAtIndexInt(m.Store, i)
					dstev.Assign2(m, m.Alloc, m.Store, m.Realm, srcev.Deref(), false)
				}
				res0 := TypedValue{
					T: IntType,
					V: nil,
				}
				res0.SetInt(int64(minl))
				m.PushValue(res0)
				return
			case *SliceType:
				dstl := dst.TV.GetLength()
				srcl := src.TV.GetLength()
				minl := min(srcl, dstl)
				if minl == 0 {
					// return 0.
					m.PushValue(defaultTypedValue(m.Alloc, IntType))
					return
				}
				dstv := dst.TV.V.(*SliceValue)
				if m.IsReadonly(dst.TV) {
					m.Panic(typedString("cannot copy to readonly tainted slice"))
				}
				dstBase := dstv.GetBase(m.Store)
				// Same as above: DidUpdate is required for the DataByte case.
				m.Realm.DidUpdate(m, dstBase, nil, nil)
				srcv := src.TV.V.(*SliceValue)
				srcBase := srcv.GetBase(m.Store)
				dstStart := dstv.Offset
				srcStart := srcv.Offset
				srcEnd := srcStart + minl

				step := 1
				start := 0
				end := minl
				// Overlap-safe copy: copy backward when dst starts after src to avoid clobbering.
				requiresBackwardCopy := dstBase == srcBase && dstStart > srcStart && dstStart < srcEnd
				if requiresBackwardCopy {
					step = -1
					start = minl - 1
					end = -1
				}
				m.incrCPU(OpCPUSlopeCopyElement * int64(minl))
				for i := start; i != end; i += step {
					dstev := dstv.GetPointerAtIndexInt2(m.Store, i, bdt.Elt)
					srcev := srcv.GetPointerAtIndexInt2(m.Store, i, bst.Elt)
					dstev.Assign2(m, m.Alloc, m.Store, m.Realm, srcev.Deref(), false)
				}
				res0 := TypedValue{
					T: IntType,
					V: nil,
				}
				res0.SetInt(int64(minl))
				m.PushValue(res0)
				return
			default:
				panic("should not happen")
			}
		},
	)
	defNative("delete",
		Flds( // params
			"m", MapT(GenT("K", nil), GenT("V", nil)), // map type
			"k", GenT("K", nil), // map key
		),
		nil, // results
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2(m.Store)
			itv := arg1.Deref()
			switch baseOf(arg0.TV.T).(type) {
			case *MapType:
				mv := arg0.TV.V.(*MapValue)

				if m.IsReadonly(arg0.TV) {
					m.Panic(typedString("cannot delete from readonly tainted map"))
				}

				val, ok := mv.GetValueForKey(m.Store, &itv)
				if !ok {
					return
				}
				// delete
				mv.DeleteForKey(m.Store, &itv)

				if m.Realm != nil {
					// mark key as deleted
					keyObj := itv.GetFirstObject(m.Store)
					m.Realm.DidUpdate(m, mv, keyObj, nil)

					// mark value as deleted
					valObj := val.GetFirstObject(m.Store)
					m.Realm.DidUpdate(m, mv, valObj, nil)
				}

				return
			default:
				panic(fmt.Sprintf(
					"unexpected map type %s",
					arg0.TV.T.String()))
			}
		},
	)
	defNative("len",
		Flds( // params
			"x", AnyT(),
		),
		Flds( // results
			"", "int",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(m.Store)
			res0 := TypedValue{
				T: IntType,
				V: nil,
			}
			res0.SetInt(int64(arg0.TV.GetLength()))
			m.PushValue(res0)
		},
	)
	// NOTE: The variadic signature is intentionally permissive.
	// Actual argument count validation (e.g. slices require 2-3 args,
	// maps/channels require 1-2) is enforced at preprocess time in
	// the "make" special case of CallExpr, not here.
	defNative("make",
		Flds( // params
			"t", GenT("T.(type)", nil),
			"z", Vrd(AnyT()),
		),
		Flds( // results
			"", GenT("T", nil),
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2(m.Store)
			vargs := arg1
			vargsl := vargs.TV.GetLength()
			tt := arg0.TV.GetType()
			m.Alloc.checkConstructionTime(tt)
			switch bt := baseOf(tt).(type) {
			case *SliceType:
				et := bt.Elem()
				switch vargsl {
				case 1:
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := int(lv.ConvertGetInt())
					if li < 0 {
						m.Panic(typedString("runtime error: makeslice: len out of range"))
					}
					if et.Kind() == Uint8Kind {
						arrayValue := m.Alloc.NewDataArray(nil, li)
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSlice(arrayValue, 0, li, li),
						})
						return
					} else {
						arrayValue := m.Alloc.NewListArray(nil, li)
						if et.Kind() == InterfaceKind {
							// leave as is
						} else {
							// init zero elements with concrete type.
							// No CPU charge: for primitives defaultTypedValue is a
							// zero-cost struct literal; for composite types it
							// allocates via m.Alloc (covered by alloc gas).
							for i := range li {
								arrayValue.List[i] = defaultTypedValue(m.Alloc, et)
							}
						}
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSlice(arrayValue, 0, li, li),
						})
						return
					}
				case 2:
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := int(lv.ConvertGetInt())
					cv := vargs.TV.GetPointerAtIndexInt(m.Store, 1).Deref()
					ci := int(cv.ConvertGetInt())

					if li < 0 {
						m.Panic(typedString("runtime error: makeslice: len out of range"))
					}
					if ci < 0 {
						m.Panic(typedString("runtime error: makeslice: cap out of range"))
					}
					if ci < li {
						m.Panic(typedString("runtime error: makeslice: cap out of range"))
					}

					if et.Kind() == Uint8Kind {
						arrayValue := m.Alloc.NewDataArray(nil, ci)
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSlice(arrayValue, 0, li, ci),
						})
						return
					} else {
						arrayValue := m.Alloc.NewListArray(nil, ci)
						if et := bt.Elem(); et.Kind() == InterfaceKind {
							// leave as is
						} else {
							// Initialize all elements within capacity with default
							// type values. These need to be initialized because future
							// slice operations could get messy otherwise. Simple capacity
							// expansions like `a = a[:cap(a)]` would make it trivial to
							// initialize zero values at the time of the slice operation.
							// But sequences of operations like:
							// 		a := make([]int, 1, 10)
							// 		a = a[7:cap(a)]
							// 		a = a[3:5]
							//
							// require a bit more work to handle correctly, requiring that
							// all new TypedValue slice elements be checked to ensure they have
							// a value for every slice operation, which is not desirable.
							// No CPU charge: for primitives defaultTypedValue is a
							// zero-cost struct literal; for composite types it
							// allocates via m.Alloc (covered by alloc gas).
							for i := range ci {
								arrayValue.List[i] = defaultTypedValue(m.Alloc, et)
							}
						}
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSlice(arrayValue, 0, li, ci),
						})
						return
					}
				default:
					panic("make() of slice type takes 2 or 3 arguments")
				}
			case *MapType:
				switch vargsl {
				case 0:
					m.PushValue(TypedValue{
						T: tt,
						V: m.Alloc.NewMap(tt, 0),
					})
					return
				case 1:
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := int(lv.ConvertGetInt())
					m.PushValue(TypedValue{
						T: tt,
						V: m.Alloc.NewMap(tt, li),
					})
					return
				default:
					panic("make() of map type takes 1 or 2 arguments")
				}
			default:
				panic(fmt.Sprintf(
					"cannot make type %s kind %v",
					tt.String(), tt.Kind()))
			}
		},
	)
	defNative("new",
		Flds( // params
			"t", GenT("T.(type)", nil),
		),
		Flds( // results
			"", GenT("*T", nil),
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(m.Store)
			tt := arg0.TV.GetType()
			m.Alloc.checkConstructionTime(tt)
			tv := defaultTypedValue(m.Alloc, tt)
			m.Alloc.AllocatePointer()
			hi := m.Alloc.NewHeapItem(tt, tv)
			m.PushValue(TypedValue{
				T: m.Alloc.NewType(&PointerType{
					Elt: tt,
				}),
				V: PointerValue{
					TV:    &hi.Value,
					Base:  hi,
					Index: 0,
				},
			})
		},
	)

	// NOTE: panic is its own statement type, and is not defined as a function.
	defNative("print",
		Flds( // params
			"xs", Vrd(AnyT()), // args[0]
		),
		nil, // results
		func(m *Machine) {
			if bm.NativeEnabled {
				arg0 := m.LastBlock().GetParams1(m.Store)
				ncode := bm.GetNativePrintCode(len(formatUverseOutput(m, arg0, false)))
				old := bm.StartNative(ncode)
				prevOutput := m.Output
				m.Output = io.Discard
				defer func() {
					bm.StopNative(ncode, old)
					m.Output = prevOutput
				}()
			}

			arg0 := m.LastBlock().GetParams1(m.Store)
			uversePrint(m, arg0, false)
		},
	)
	defNative("println",
		Flds( // param
			"xs", Vrd(AnyT()), // args[0]
		),
		nil, // results
		func(m *Machine) {
			if bm.NativeEnabled {
				arg0 := m.LastBlock().GetParams1(m.Store)
				ncode := bm.GetNativePrintCode(len(formatUverseOutput(m, arg0, false)))
				old := bm.StartNative(ncode)
				prevOutput := m.Output
				m.Output = io.Discard
				defer func() {
					bm.StopNative(ncode, old)
					m.Output = prevOutput
				}()
			}
			arg0 := m.LastBlock().GetParams1(m.Store)
			uversePrint(m, arg0, true)
		},
	)
	defNative("panic",
		Flds( // params
			"exception", AnyT(),
		),
		nil, // results
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(m.Store)
			ex := arg0.TV.Copy(m.Alloc)
			// m.Panic(ex) also works, but after return will immediately OpPanic2.
			// This should be the only place .pushPanic() is called
			// outside of op_*.go doOp*() functions.
			m.pushPanic(ex)
		},
	)
	defNative("recover",
		nil, // params
		Flds( // results
			"exception", AnyT(),
		),
		func(m *Machine) {
			exception := m.Recover()
			if exception == nil {
				m.PushValue(TypedValue{})
			} else {
				m.PushValue(exception.Value)
			}
		},
	)

	//----------------------------------------
	// Gno2 types
	def("address", asValue(gAddressType))
	defNativeMethod("address", "String",
		nil, // params
		Flds( // results
			"", "string",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			res0 := typedString(arg0.TV.GetString())
			m.PushValue(res0)
		},
	)
	defNativeMethod("address", "IsValid",
		nil, // params
		Flds( // results
			"", "bool",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			b32addr := arg0.TV.GetString()
			addr, err := crypto.AddressFromBech32(b32addr)
			if err != nil {
				m.PushValue(typedBool(false))
				return
			}
			_ = addr
			m.PushValue(typedBool(len(addr) == 20))
		},
	)
	def("realm", asValue(gRealmType))
	def(".grealm", asValue(gConcreteRealmType))
	defNativePtrMethod(".grealm", "Address",
		nil, // params
		Flds( // results
			"", "address",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			addr := sv.Fields[0].GetString()
			m.PushValue(TypedValue{T: gAddressType, V: StringValue(addr)})
		},
	)
	defNativePtrMethod(".grealm", "PkgPath",
		nil, // params
		Flds( // results
			"", "string",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			path := sv.Fields[1].GetString()
			m.PushValue(typedString(path))
		},
	)
	defNativePtrMethod(".grealm", "Previous",
		nil, // params
		Flds( // results
			"", "realm",
		),
		func(m *Machine) {
			// Return the prev field verbatim when it carries a realm
			// value (non-nil prev). At the chain root or after a
			// testing.SetRealm UserRealm override, the prev field is a
			// truly-nil TypedValue — in that case panic to match
			// runtime.PreviousRealm()'s walk-end behavior, so user code
			// has a consistent boundary signal across both APIs and
			// can use defer-recover if it needs to detect the EOA root.
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			prev := sv.Fields[2]
			if prev.T == nil { // truly-nil — no previous beyond this.
				m.PanicString("frame not found: cannot seek beyond origin caller override")
				return
			}
			m.PushValue(prev)
		},
	)
	// IsCode / IsUser / IsUserCall / IsUserRun / IsEphemeral mirror the
	// classification methods on chain/runtime.Realm so a captured `cur`
	// can answer caller-shape questions without a runtime walk. The
	// derivations are pure on the (addr, pkgPath) stored in the .grealm
	// struct and match the chain/runtime implementations at
	// gnovm/stdlibs/chain/runtime/frame.gno.
	defNativePtrMethod(".grealm", "IsCode",
		nil,
		Flds("", "bool"),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			m.PushValue(typedBool(sv.Fields[1].GetString() != ""))
		},
	)
	defNativePtrMethod(".grealm", "IsUserCall",
		nil,
		Flds("", "bool"),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			m.PushValue(typedBool(sv.Fields[1].GetString() == ""))
		},
	)
	defNativePtrMethod(".grealm", "IsUserRun",
		nil,
		Flds("", "bool"),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			addr := sv.Fields[0].GetString()
			path := sv.Fields[1].GetString()
			m.PushValue(typedBool(realmIsUserRun(addr, path)))
		},
	)
	defNativePtrMethod(".grealm", "IsUser",
		nil,
		Flds("", "bool"),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			addr := sv.Fields[0].GetString()
			path := sv.Fields[1].GetString()
			m.PushValue(typedBool(path == "" || realmIsUserRun(addr, path)))
		},
	)
	defNativePtrMethod(".grealm", "IsEphemeral",
		nil,
		Flds("", "bool"),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			m.PushValue(typedBool(realmIsEphemeral(sv.Fields[1].GetString())))
		},
	)
	// IsCurrent returns true iff the receiver is the captured cur of the
	// topmost crossing frame on the live call stack — i.e., the receiver
	// was minted by installCrossingCur for the currently-executing
	// crossing-function invocation, not derived from a .Previous() walk
	// nor obtained from a sibling/ancestor crossing frame.
	//
	// Comparison is pointer-identity on the underlying *HeapItemValue,
	// not (addr, pkgPath) equality: two distinct cross-calls into the
	// same realm (A→B→A re-entry, or A→B return → A again) mint distinct
	// .grealm HIVs, so IsCurrent returns true for at most one frame's
	// cur at any moment.
	//
	// Receivers reached only via the bare-struct value-receiver path (no
	// HIV wrapper) always return false, since pointer-identity comparison
	// has no anchor there. Returns false when no crossing frame is in
	// scope (top of machine during package init's non-crossing entry,
	// MsgRun main, etc.).
	defNativePtrMethod(".grealm", "IsCurrent",
		nil,
		Flds("", "bool"),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			m.PushValue(typedBool(realmIsCurrentOnMachine(m, arg0.TV)))
		},
	)
	defNativePtrMethod(".grealm", "String",
		nil, // params
		Flds( // results
			"", "string",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			sv := derefRealmStruct(arg0.TV)
			addr := sv.Fields[0].GetString()
			path := sv.Fields[1].GetString()
			m.PushValue(typedString("realm{" + path + ":" + addr + "}"))
		},
	)
	def(".cur", undefined)    // special keyword for non-cross-calling main(cur realm)
	def(".origin", undefined) // sentinel for compiler-synthesized chain-root crossing calls (MsgCall keeper synthesis)
	// cross(rlm) is the explicit cross-call form. It validates
	// IsCurrent on rlm and returns it unchanged.
	//
	// When used at Args[0] of a crossing call (the intended usage),
	// the validated rlm flows through to the outer call's Args[0]
	// slot; installCrossingCur peeks the realm value and uses it as
	// the new cur's prev. No second IsCurrent check is needed
	// downstream because cross has already validated.
	//
	// realmIsCurrentOnMachine skips cross's own frame (cross is a
	// non-crossing native), finds the most recent crossing frame —
	// the caller of whatever evaluated cross(rlm) — and compares
	// rlm's HIV against that frame's Cur by pointer identity. Catches
	// stale rlm from sibling frames or captured-and-outlived frames.
	//
	// The Go-side typechecker shim narrows X to realm via the
	// .gnobuiltins.gno signature `func cross(rlm realm) realm`.
	defNative("cross",
		Flds( // param
			"rlm", GenT("X", nil),
		),
		Flds( // result
			"result", GenT("X", nil),
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(nil)
			if !realmIsCurrentOnMachine(m, arg0.TV) {
				panic("cross: rlm is not the current cur (stale capture or sibling frame)")
			}
			m.PushValue(*arg0.TV)
		},
	)
	defNative("attach",
		Flds( // params
			"xs", Vrd(AnyT()), // args[0]
		),
		nil, // results
		func(m *Machine) {
			panic("attach() is not yet supported")
		},
	)
	// Typed nils in Go1 are problematic.
	// https://dave.cheney.net/2017/08/09/typed-nils-in-go-2
	// Dave Cheney suggests typed-nil == nil when the typed-nil is not an
	// interface type, but arguably it should be the other way around, e.g.
	// > (*int)(nil) != nil.
	// Since Gno doesn't yet support reflect, and since even with reflect
	// implementing istypednil() is annoying, while istypednil() shouldn't
	// require reflect, Gno should therefore offer istypednil() as a uverse
	// function.
	// XXX REMOVE, move to std function.
	defNative("istypednil",
		Flds( // params
			"x", AnyT(),
		),
		Flds( // results
			"", "bool",
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(m.Store)
			m.PushValue(typedBool(arg0.TV.IsTypedNil()))
		},
	)
	// In the final form, it will do nothing if no abort; but otherwise
	// will make it as if nothing happened (with full cache wrapping). This
	// gives programs precognition, or at least hypotheticals.
	// e.g. "If it **would have** done this, do that instead".
	//
	// XXX This is only enabled in testing mode (for now), and test
	// developers should be aware that behavior will change to be like
	// above; currently it doesn't cache-wrap the fn function so residual
	// state mutations remain even after revive(), but they will be
	// "magically" rolled back upon panic in the future. The fn function
	// must *always* panic in the end in order to prevent state mutations
	// after a non-aborting transaction.
	defNative("revive",
		Flds( // params
			"fn", FuncT(nil, nil),
		),
		Flds( // results
			"ex", AnyT(),
		),
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1(m.Store)
			if m.ReviveEnabled {
				last := m.LastFrame()

				// Push the no-abort result.
				// last.SetRevive() marks the frame and this
				// value will get replaced w/ exception.
				m.PushValue(TypedValue{})
				last.SetIsRevive()

				// Push function and precall it.
				m.PushExpr(Call(&ConstExpr{Source: X("fn"), TypedValue: *arg0.TV}))
				m.PushOp(OpPrecall)
				m.PushValue(*arg0.TV)
			} else {
				// If revive isn't enabled just panic.
				m.pushPanic(typedString("revive() not enabled"))
				// m.PushValue(TypedValue{})
			}
		},
	)
	uverseValue = uverseNode.NewPackage(nilAllocator)
}

func copyDataToList(dst []TypedValue, data []byte, et Type) {
	for i := range data {
		dst[i] = TypedValue{T: et}
		dst[i].SetUint8(data[i])
	}
}

func copyListToData(dst []byte, tvs []TypedValue) {
	for i := range tvs {
		dst[i] = tvs[i].GetUint8()
	}
}

func copyListToRunes(dst []rune, tvs []TypedValue) {
	for i := range tvs {
		dst[i] = tvs[i].GetInt32()
	}
}

func consumeGas(m *Machine, amount types.Gas) {
	if m.GasMeter != nil {
		m.GasMeter.ConsumeGas(amount, "CPUCycles")
	}
}

// uversePrint is used for the print and println functions.
// println passes newline = true.
// xv contains the variadic argument passed to the function.
func uversePrint(m *Machine, xv PointerValue, newline bool) {
	consumeGas(m, NativeCPUUversePrintInit)
	output := formatUverseOutput(m, xv, newline)
	consumeGas(m, overflow.Divp(types.Gas(len(output)), NativeCPUUversePrintCharsPerGas))
	// For debugging:
	// fmt.Println(colors.Cyan(string(output)))
	m.Output.Write(output)
}

func formatUverseOutput(m *Machine, xv PointerValue, newline bool) []byte {
	xvl := xv.TV.GetLength()
	switch xvl {
	case 0:
		if newline {
			return bNewline
		}
	case 1:
		ev := xv.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
		res := ev.Sprint(m)
		if newline {
			res += "\n"
		}
		return []byte(res)
	default:
		var buf bytes.Buffer

		for i := range xvl {
			if i != 0 { // Not the last item.
				buf.WriteByte(' ')
			}
			ev := xv.TV.GetPointerAtIndexInt(m.Store, i).Deref()
			res := ev.Sprint(m)
			buf.WriteString(res)
		}
		if newline {
			buf.WriteByte('\n')
		}
		return buf.Bytes()
	}

	return nil
}

var bNewline = []byte("\n")
