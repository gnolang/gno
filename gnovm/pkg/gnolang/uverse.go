package gnolang

// XXX append and delete need checks too.

import (
	"bytes"
	"fmt"
	"io"

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

var gCoinType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    "gnocoin",
	Base: &StructType{
		PkgPath: uversePkgPath,
		Fields: []FieldType{
			{Name: "Denom", Type: StringType},
			{Name: "Amount", Type: Int64Type},
		},
	},
	sealed: true,
}

var gCoinsType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    "gnocoins",
	Base:    &SliceType{Elt: gCoinType},
	sealed:  true,
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
			}, {
				Name: "Coins",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{{
						Type: gCoinsType,
					}},
				},
			}, {
				Name: "Send",
				Type: &FuncType{
					Params: []FieldType{{
						Name: "coins", Type: gCoinsType,
					}, {
						Name: "to", Type: gAddressType,
					}},
					Results: []FieldType{{
						Type: gErrorType,
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
			}, { // gets filled in init() below.
				Name: "Origin",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{{
						Type: nil,
					}},
				},
			}, { // gets filled in init() below.
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

func init() {
	gRealmPrevious := gRealmType.Base.(*InterfaceType).GetMethodFieldType("Previous")
	gRealmOrigin := gRealmType.Base.(*InterfaceType).GetMethodFieldType("Origin")
	gRealmPrevious.Type.(*FuncType).Results[0].Type = gRealmType
	gRealmOrigin.Type.(*FuncType).Results[0].Type = gRealmType
}

var gConcreteRealmType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    ".grealm",
	Base: &StructType{
		PkgPath: uversePkgPath,
		Fields: []FieldType{
			{Name: "addr", Type: gAddressType},
			{Name: "pkgPath", Type: StringType},
			{Name: "prev", Type: gRealmType},
		},
	},
	sealed: true,
	// methods defined in makeUverseNode()
}

// NOTE: the value is set as a constExpr for the `.cur` in the preprocessor,
// and likewise for MsgCall cross-call of crossing functions, so the value
// should be deterministic, not dynamic, and only depend on the realm.
func NewConcreteRealm(pkgPath string) TypedValue {
	return TypedValue{
		T: gConcreteRealmType,
		V: &StructValue{
			Fields: []TypedValue{
				{T: gAddressType, V: nil}, // XXX
				{T: StringType, V: StringValue(pkgPath)},
				{T: gConcreteRealmType, V: nil}, // XXX
			},
		},
	}
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
	if !(bm.OpsEnabled || bm.StorageEnabled) {
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
		return &PackageNode{}
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
				newArrayValue := m.Alloc.NewDataArray(len(arg1String))
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
						arrayValue := m.Alloc.NewDataArray(arg1Length)
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
						arrayValue := m.Alloc.NewListArray(arg1Length)
						if arg1Length > 0 {
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
				// m.Realm.DidUpdate(arg0Base, nil, nil) FIRST TO CHECK WRITE PERMISSIONS.
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
							m.Realm.DidUpdate(arg0Base, nil, nil)

							if arg0Base.Data == nil {
								// append(*SliceValue.List, *SliceValue) ---------
								list := arg0Base.List
								if arg1Base.Data == nil {
									for i := range arg1Length {
										oldElem := list[arg0Offset+arg0Length+i]
										// unrefCopy will resolve references and copy their values
										// to copy by value rather than by reference.
										newElem := arg1Base.List[arg1Offset+i].unrefCopy(m.Alloc, m.Store)
										list[arg0Offset+arg0Length+i] = newElem

										m.Realm.DidUpdate(
											arg0Base,
											oldElem.GetFirstObject(m.Store),
											newElem.GetFirstObject(m.Store),
										)
									}
								} else {
									copyDataToList(
										list[arg0Offset+arg0Length:arg0Offset+arg0Length+arg1Length],
										arg1Base.Data[arg1Offset:arg1Offset+arg1Length],
										arg0Type.Elem())
								}
							} else {
								// append(*SliceValue.Data, *SliceValue) ---------
								data := arg0Base.Data
								if arg1Base.Data == nil {
									copyListToData(
										data[arg0Offset+arg0Length:arg0Offset+arg0Length+arg1Length],
										arg1Base.List[arg1Offset:arg1Offset+arg1Length])
								} else {
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
						arrayValue := m.Alloc.NewDataArray(newLength)
						if 0 < arg0Length {
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
						arrayValue := m.Alloc.NewListArray(arrayLen)
						if arg0Length > 0 {
							if arg0Base.Data == nil {
								for i := range arg0Length {
									arrayValue.List[i] = arg0Base.List[arg0Offset+i].unrefCopy(m.Alloc, m.Store)
								}
							} else {
								panic("should not happen")
							}
						}

						if arg1Length > 0 {
							if arg1Base.Data == nil {
								for i := range arg1Length {
									arrayValue.List[arg0Length+i] = arg1Base.List[arg1Offset+i].unrefCopy(m.Alloc, m.Store)
								}
							} else {
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
				// Guard for protecting dst against mutation by external realms.
				dstBase := dstv.GetBase(m.Store)
				m.Realm.DidUpdate(dstBase, nil, nil)
				// TODO: consider an optimization if dstv.Data != nil.
				for i := range minl {
					dstev := dstv.GetPointerAtIndexInt2(m.Store, i, bdt.Elt)
					srcev := src.TV.GetPointerAtIndexInt(m.Store, i)
					dstev.Assign2(m.Alloc, m.Store, m.Realm, srcev.Deref(), false)
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
				// Guard for protecting dst against mutation by external realms.
				dstBase := dstv.GetBase(m.Store)
				m.Realm.DidUpdate(dstBase, nil, nil)
				srcv := src.TV.V.(*SliceValue)
				for i := range minl {
					dstev := dstv.GetPointerAtIndexInt2(m.Store, i, bdt.Elt)
					srcev := srcv.GetPointerAtIndexInt2(m.Store, i, bst.Elt)
					dstev.Assign2(m.Alloc, m.Store, m.Realm, srcev.Deref(), false)
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

				// Guard for protecting map against mutation by external realms. This is necessary
				m.Realm.DidUpdate(mv, nil, nil)

				val, ok := mv.GetValueForKey(m.Store, &itv)
				if !ok {
					return
				}
				// delete
				mv.DeleteForKey(m.Store, &itv)

				if m.Realm != nil {
					// mark key as deleted
					keyObj := itv.GetFirstObject(m.Store)
					m.Realm.DidUpdate(mv, keyObj, nil)

					// mark value as deleted
					valObj := val.GetFirstObject(m.Store)
					m.Realm.DidUpdate(mv, valObj, nil)
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
			switch bt := baseOf(tt).(type) {
			case *SliceType:
				et := bt.Elem()
				switch vargsl {
				case 1:
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := int(lv.ConvertGetInt())
					if et.Kind() == Uint8Kind {
						arrayValue := m.Alloc.NewDataArray(li)
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSlice(arrayValue, 0, li, li),
						})
						return
					} else {
						arrayValue := m.Alloc.NewListArray(li)
						if et.Kind() == InterfaceKind {
							// leave as is
						} else {
							// init zero elements with concrete type.
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

					if ci < li {
						m.Panic(typedString(`makeslice: cap out of range`))
					}

					if et.Kind() == Uint8Kind {
						arrayValue := m.Alloc.NewDataArray(ci)
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSlice(arrayValue, 0, li, ci),
						})
						return
					} else {
						arrayValue := m.Alloc.NewListArray(ci)
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
				// NOTE: the type is not used.
				switch vargsl {
				case 0:
					m.PushValue(TypedValue{
						T: tt,
						V: m.Alloc.NewMap(0),
					})
					return
				case 1:
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := int(lv.ConvertGetInt())
					m.PushValue(TypedValue{
						T: tt,
						V: m.Alloc.NewMap(li),
					})
					return
				default:
					panic("make() of map type takes 1 or 2 arguments")
				}
			case *ChanType:
				switch vargsl {
				case 0, 1:
					panic("not yet implemented")
				default:
					panic("make() of chan type takes 1 or 2 arguments")
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
			tv := defaultTypedValue(m.Alloc, tt)
			m.Alloc.AllocatePointer()
			hi := m.Alloc.NewHeapItem(tv)
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
			// Todo: should stop op code benchmarking here.
			if bm.NativeEnabled {
				arg0 := m.LastBlock().GetParams1(m.Store)
				bm.StartNative(bm.GetNativePrintCode(len(formatUverseOutput(m, arg0, false))))
				prevOutput := m.Output
				m.Output = io.Discard
				defer func() {
					bm.StopNative()
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
			// Todo: should stop op code benchmarking here.
			if bm.NativeEnabled {
				arg0 := m.LastBlock().GetParams1(m.Store)
				bm.StartNative(bm.GetNativePrintCode(len(formatUverseOutput(m, arg0, false))))
				prevOutput := m.Output
				m.Output = io.Discard
				defer func() {
					bm.StopNative()
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
	def("gnocoin", asValue(gCoinType))
	def("gnocoins", asValue(gCoinsType))
	def("realm", asValue(gRealmType))
	def(".grealm", asValue(gConcreteRealmType))
	defNativeMethod(".grealm", "Address",
		nil, // params
		Flds( // results
			"", "address",
		),
		func(m *Machine) {
			panic("not yet implemented")
		},
	)
	defNativeMethod(".grealm", "PkgPath",
		nil, // params
		Flds( // results
			"", "string",
		),
		func(m *Machine) {
			panic("not yet implemented")
		},
	)
	defNativeMethod(".grealm", "Coins",
		nil, // params
		Flds( // results
			"", "gnocoins",
		),
		func(m *Machine) {
			panic("not yet implemented")
		},
	)
	defNativeMethod(".grealm", "Send",
		Flds( // params
			"coins", "gnocoins",
			"to", "address",
		),
		Flds( // results
			"", "error",
		),
		func(m *Machine) {
			panic("not yet implemented")
		},
	)
	defNativeMethod(".grealm", "Origin",
		nil, // params
		Flds( // results
			"", "realm",
		),
		func(m *Machine) {
			panic("not yet implemented")
		},
	)
	defNativeMethod(".grealm", "Previous",
		nil, // params
		Flds( // results
			"", "realm",
		),
		func(m *Machine) {
			panic("not yet implemented")
		},
	)
	defNativeMethod(".grealm", "String",
		nil, // params
		Flds( // results
			"", "string",
		),
		func(m *Machine) {
			panic("not yet implemented")
		},
	)
	defNative("crossing",
		nil, // params
		nil, // results
		func(m *Machine) {
			// should not happen since gno 0.9.
			panic("crossing() is reserved but deprecated")
		},
	)
	def("cross", undefined) // special keyword for cross-calling
	def(".cur", undefined)  // special keyword for non-cross-calling main(cur realm)
	// `cross` used to be a function, but it is now a special value.
	// XXX make this unavailable in prod 0.9.  Code that refers to this
	// intermediate name (gno fix > prepare()) will not pass type-checking
	// because it isn't available in .gnobuiltins.gno for gno 0.9, but this
	// name is unnecessarily reserved and brittle.
	defNative("_cross_gno0p0",
		Flds( // param
			"x", GenT("X", nil),
		),
		Flds( // results
			"x", GenT("X", nil),
		),
		func(m *Machine) {
			// This is handled by op_call instead.
			panic("cross is a virtual function")
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
