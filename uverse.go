package gno

import (
	"fmt"
	"reflect"
	"strings"
)

//----------------------------------------
// non-primitive builtin types

var gErrorType = &DeclaredType{
	PkgPath: uversePkgPath,
	Name:    "error",
	Base: &InterfaceType{
		PkgPath: uversePkgPath,
		Methods: []FieldType{
			FieldType{
				Name: "Error",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{
						FieldType{
							//Name: "",
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
			FieldType{
				Name: "String",
				Type: &FuncType{
					Params: nil,
					Results: []FieldType{
						FieldType{
							//Name: "",
							Type: StringType,
						},
					},
				},
			},
		},
	},
	sealed: true,
}

//----------------------------------------
// Uverse package

var uverseNode *PackageNode
var uverseValue *PackageValue

const uversePkgPath = ".uverse"

// Always returns a new copy from the latest state of source.
func Uverse() *PackageValue {
	if uverseValue == nil {
		pn := UverseNode()
		uverseValue = pn.NewPackage()
	}
	return uverseValue
}

// Always returns the same instance with possibly differing completeness.
func UverseNode() *PackageNode {
	// Global is singleton.
	if uverseNode != nil {
		return uverseNode
	}
	fmt.Print("baking uverse...")
	defer fmt.Println(" done")

	// NOTE: uverse node is hidden, thus the leading dot in pkgPath=".uverse".
	uverseNode = NewPackageNode("uverse", uversePkgPath, nil)

	// temporary convenience functions.
	def := func(n Name, tv TypedValue) {
		uverseNode.Define(n, tv)
	}
	defNative := uverseNode.DefineNative

	// Primitive types
	undefined := TypedValue{}
	def("._", undefined)   // special, path is zero.
	def("iota", undefined) // special
	def("nil", undefined)
	def("bool", asValue(BoolType))
	def("string", asValue(StringType))
	def("int", asValue(IntType))
	def("int8", asValue(Int8Type))
	def("int16", asValue(Int16Type))
	def("rune", asValue(Int32Type))
	def("int32", asValue(Int32Type))
	def("int64", asValue(Int64Type))
	def("uint", asValue(UintType))
	def("byte", asValue(Uint8Type))
	def("uint8", asValue(Uint8Type))
	def("uint16", asValue(Uint16Type))
	def("uint32", asValue(Uint32Type))
	def("uint64", asValue(Uint64Type))
	def("bigint", asValue(BigintType))
	def("float32", asValue(Float32Type))
	def("float64", asValue(Float64Type))
	// NOTE on 'typeval': We can't call the type of a TypeValue a
	// "type", even though we want to, because it conflicts with
	// the pre-existing syntax for type-switching, `switch
	// x.(type) {case SomeType:...}`, for if x.(type) were not a
	// type-switch but a type-assertion, and the resulting value
	// could be any type, such as an IntType; whereas as the .X of
	// a SwitchStmt, the type of an IntType value is not IntType
	// but always a TypeType (all types are of type TypeType).
	//
	// The ideal solution is to keep the syntax consistent for
	// type-assertions, but for backwards compatibility, the
	// keyword that represents the TypeType type is not "type" but
	// "typeval".  The value of a "typeval" value is represented
	// by a TypeValue.
	def("typeval", asValue(gTypeType))
	def("error", asValue(gErrorType))

	// Values
	def("true", untypedBool(true))
	def("false", untypedBool(false))

	// Functions
	defNative("append",
		Flds( // params
			"x", GenT("X", nil), // args[0]
			"args", MaybeNativeT(Vrd(GenT("X.Elem()", nil))), // args[1]
		),
		Flds( // results
			"res", GenT("X", nil), // res
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2()
			// As a special case, if arg1 is a string type, first convert it into
			// a data slice type.
			if arg1.TV.T != nil && arg1.TV.T.Kind() == StringKind {
				arg1s := arg1.TV.GetString()
				// NOTE: this hack works because
				// arg1 PointerValue is not a pointer,
				// so the modification here is only local.
				m.Alloc.AllocateSlice()
				m.Alloc.AllocateDataArray(int64(len(arg1s)))
				arg1.TV = &TypedValue{
					T: &SliceType{ // TODO: reuse
						Elt: Uint8Type,
						Vrd: true,
					},
					V: &SliceValue{ // TODO: pool?
						Base: &ArrayValue{
							Data: []byte(arg1s),
						},
						Offset: 0,
						Length: len(arg1s),
						Maxcap: len(arg1s),
					},
				}
			}
			xt := arg0.TV.T
			argt := arg1.TV.T
			switch xv := arg0.TV.V.(type) {

			//----------------------------------------------------------------
			// append(nil, ???)
			case nil:
				switch args := arg1.TV.V.(type) {

				//------------------------------------------------------------
				// append(nil, nil)
				case nil: // no change
					m.PushValue(TypedValue{
						T: xt,
						V: nil,
					})
					return

				//------------------------------------------------------------
				// append(nil, *SliceValue)
				case *SliceValue:
					argsl := args.Length
					argso := args.Offset
					argsb := args.GetBase(m.Store)
					if argsl == 0 { // no change
						m.PushValue(TypedValue{
							T: xt,
							V: nil,
						})
						return
					} else if xt.Elem().Kind() == Uint8Kind {
						// append(nil, *SliceValue) new data bytes ---
						data := make([]byte, argsl)
						if argsb.Data == nil {
							copyListToData(
								data[:argsl],
								argsb.List[argso:argso+argsl])
						} else {
							copy(
								data[:argsl],
								argsb.Data[argso:argso+argsl])
						}
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(nil, *SliceValue) new list ---------
						list := make([]TypedValue, argsl)
						if 0 < argsl {
							copy(
								list[:argsl],
								argsb.List[argso:argso+argsl])
						}
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				//------------------------------------------------------------
				// append(nil, *NativeValue)
				case *NativeValue:
					argsrv := args.Value
					argsl := argsrv.Len()
					if argsl == 0 { // no change
						m.PushValue(TypedValue{
							T: xt,
							V: nil,
						})
						return
					} else if xt.Elem().Kind() == Uint8Kind {
						// append(nil, *NativeValue) new data bytes --
						data := make([]byte, argsl)
						copyNativeToData(
							data[:argsl],
							argsrv, argsl)
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(nil, *NativeValue) new list --------
						list := make([]TypedValue, argsl)
						if 0 < argsl {
							copyNativeToList(
								m.Alloc,
								list[:argsl],
								argsrv, argsl)
						}
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				//------------------------------------------------------------
				default:
					panic("should not happen")

				}

			//----------------------------------------------------------------
			// append(*SliceValue, ???)
			case *SliceValue:
				xvl := xv.Length
				xvo := xv.Offset
				xvc := xv.Maxcap
				xvb := xv.GetBase(m.Store)
				switch args := arg1.TV.V.(type) {

				//------------------------------------------------------------
				// append(*SliceValue, nil)
				case nil: // no change
					m.PushValue(TypedValue{
						T: xt,
						V: xv,
					})
					return

				//------------------------------------------------------------
				// append(*SliceValue, *SliceValue)
				case *SliceValue:
					argsl := args.Length
					argso := args.Offset
					argsb := args.GetBase(m.Store)
					if xvl+argsl <= xvc {
						// append(*SliceValue, *SliceValue) w/i capacity -----
						if 0 < argsl { // implies 0 < xvc
							if xvb.Data == nil {
								// append(*SliceValue.List, *SliceValue) ---------
								list := xvb.List
								if argsb.Data == nil {
									copy(
										list[xvo+xvl:xvo+xvl+argsl],
										argsb.List[argso:argso+argsl])
								} else {
									copyDataToList(
										list[xvo+xvl:xvo+xvl+argsl],
										argsb.Data[argso:argso+argsl],
										xt.Elem())
								}
							} else {
								// append(*SliceValue.Data, *SliceValue) ---------
								data := xvb.Data
								if argsb.Data == nil {
									copyListToData(
										data[xvo+xvl:xvo+xvl+argsl],
										argsb.List[argso:argso+argsl])
								} else {
									copy(
										data[xvo+xvl:xvo+xvl+argsl],
										argsb.Data[argso:argso+argsl])
								}
							}
							m.Alloc.AllocateSlice()
							m.PushValue(TypedValue{
								T: xt,
								V: &SliceValue{
									Base:   xvb,
									Offset: xvo,
									Length: xvl + argsl,
									Maxcap: xvc,
								},
							})
							return
						} else { // no change
							m.PushValue(TypedValue{
								T: xt,
								V: xv,
							})
							return
						}
					} else if xt.Elem().Kind() == Uint8Kind {
						// append(*SliceValue, *SliceValue) new data bytes ---
						data := make([]byte, xvl+argsl)
						if 0 < xvl {
							if xvb.Data == nil {
								copyListToData(
									data[:xvl],
									xvb.List[xvo:xvo+xvl])
							} else {
								copy(
									data[:xvl],
									xvb.Data[xvo:xvo+xvl])
							}
						}
						if 0 < argsl {
							if argsb.Data == nil {
								copyListToData(
									data[xvl:xvl+argsl],
									argsb.List[argso:argso+argsl])
							} else {
								copy(
									data[xvl:xvl+argsl],
									argsb.Data[argso:argso+argsl])
							}
						}
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(*SliceValue, *SliceValue) new list ---------
						list := make([]TypedValue, xvl+argsl)
						if 0 < xvl {
							if xvb.Data == nil {
								copy(
									list[:xvl],
									xvb.List[xvo:xvo+xvl])
							} else {
								panic("should not happen")
								/*
									copyDataToList(
										list[:xvl],
										xvb.Data[xvo:xvo+xvl],
										xt.Elem(),
									)
								*/
							}
						}
						if 0 < argsl {
							if argsb.Data == nil {
								copy(
									list[xvl:xvl+argsl],
									argsb.List[argso:argso+argsl])
							} else {
								copyDataToList(
									list[xvl:xvl+argsl],
									argsb.Data[argso:argso+argsl],
									argt.Elem(),
								)
							}
						}
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				//------------------------------------------------------------
				// append(*SliceValue, *NativeValue)
				case *NativeValue:
					argsrv := args.Value
					argsl := argsrv.Len()
					if xvl+argsl <= xvc {
						// append(*SliceValue, *NativeValue) w/i capacity ----
						if 0 < argsl { // implies 0 < xvc
							if xvb.Data == nil {
								// append(*SliceValue.List, *NativeValue) --------
								list := xvb.List
								copyNativeToList(
									m.Alloc,
									list[xvo:xvo+argsl],
									argsrv, argsl)
							} else {
								// append(*SliceValue.Data, *NativeValue) --------
								data := xvb.Data
								copyNativeToData(
									data[xvo:xvo+argsl],
									argsrv, argsl)
							}
							m.Alloc.AllocateSlice()
							m.PushValue(TypedValue{
								T: xt,
								V: &SliceValue{
									Base:   xvb,
									Offset: xvo,
									Length: xvl + argsl,
									Maxcap: xvc,
								},
							})
							return
						} else { // no change
							m.PushValue(TypedValue{
								T: xt,
								V: xv,
							})
							return
						}
					} else if xt.Elem().Kind() == Uint8Kind {
						// append(*SliceValue, *NativeValue) new data bytes --
						data := make([]byte, xvl+argsl)
						if 0 < xvl {
							if xvb.Data == nil {
								copyListToData(
									data[:xvl],
									xvb.List[xvo:xvo+xvl])
							} else {
								copy(
									data[:xvl],
									xvb.Data[xvo:xvo+xvl])
							}
						}
						if 0 < argsl {
							copyNativeToData(
								data[xvl:xvl+argsl],
								argsrv, argsl)
						}
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(*SliceValue, *NativeValue) new list --------
						list := make([]TypedValue, xvl+argsl)
						if 0 < xvl {
							copy(
								list[:xvl],
								xvb.List[xvo:xvo+xvl])
						}
						if 0 < argsl {
							copyNativeToList(
								m.Alloc,
								list[xvl:xvl+argsl],
								argsrv, argsl)
						}
						m.PushValue(TypedValue{
							T: xt,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				//------------------------------------------------------------
				default:
					panic("should not happen")

				}

			//----------------------------------------------------------------
			// append(*NativeValue, ???)
			case *NativeValue:
				sv := xv.Value
				switch args := arg1.TV.V.(type) {

				//------------------------------------------------------------
				// append(*NativeValue, nil)
				case nil: // no change
					m.PushValue(TypedValue{
						T: xt,
						V: xv,
					})
					return

				//------------------------------------------------------------
				// append(*NativeValue, *SliceValue)
				case *SliceValue:
					st := sv.Type()
					argso := args.Offset
					argsl := args.Length
					argsb := args.GetBase(m.Store)
					if 0 < argsl {
						argsrv := reflect.MakeSlice(st, argsl, argsl)
						if argsb.Data == nil {
							for i := 0; i < argsl; i++ {
								etv := &(argsb.List[argso+i])
								if etv.IsUndefined() {
									continue
								}
								erv := gno2GoValue(etv, reflect.Value{})
								argsrv.Index(i).Set(erv)
							}
						} else {
							for i := 0; i < argsl; i++ {
								erv := argsrv.Index(i)
								erv.SetUint(uint64(argsb.Data[argso+i]))
							}
						}
						resrv := reflect.AppendSlice(sv, argsrv)
						m.Alloc.AllocateNative()
						m.PushValue(TypedValue{
							T: xt,
							V: &NativeValue{Value: resrv},
						})
						return
					} else { // no change
						m.PushValue(TypedValue{
							T: xt,
							V: xv,
						})
						return
					}

				//------------------------------------------------------------
				// append(*NativeValue, *NativeValue)
				case *NativeValue:
					argsrv := args.Value
					resrv := reflect.AppendSlice(sv, argsrv)
					m.Alloc.AllocateNative()
					m.PushValue(TypedValue{
						T: xt,
						V: &NativeValue{Value: resrv},
					})
					return

				//------------------------------------------------------------
				// append(*NativeValue, StringValue)
				case StringValue:
					if xt.Elem().Kind() == Uint8Kind {
						// TODO this might be faster if reflect supports
						// appending this way without first converting to a slice.
						argrv := reflect.ValueOf([]byte(arg1.TV.V.(StringValue)))
						resrv := reflect.AppendSlice(sv, argrv)
						m.Alloc.AllocateNative()
						m.PushValue(TypedValue{
							T: xt,
							V: &NativeValue{Value: resrv},
						})
						return
					} else {
						panic(fmt.Sprintf(
							"cannot append %s to %s",
							arg1.TV.T.String(), xt.String()))
					}

				//------------------------------------------------------------
				// append(*NativeValue, ???)
				default:
					panic(fmt.Sprintf(
						"cannot append %s to %s",
						arg1.TV.T.String(), xt.String()))

				}

			//----------------------------------------------------------------
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
			arg0 := m.LastBlock().GetParams1()
			res0 := TypedValue{
				T: IntType,
				V: nil,
			}
			res0.SetInt(arg0.TV.GetCapacity())
			m.PushValue(res0)
			return
		},
	)
	def("close", undefined)
	def("complex", undefined)
	defNative("copy",
		Flds( // params
			"dst", GenT("X", nil),
			"src", GenT("Y", nil),
		),
		Flds( // results
			"", "int",
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2()
			dst, src := arg0, arg1
			switch bdt := baseOf(dst.TV.T).(type) {
			case *SliceType:
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
					minl := dstl
					if srcl < dstl {
						minl = srcl
					}
					if minl == 0 {
						// return 0.
						m.PushValue(defaultTypedValue(m.Alloc, IntType))
						return
					}
					dstv := dst.TV.V.(*SliceValue)
					// TODO: consider an optimization if dstv.Data != nil.
					for i := 0; i < minl; i++ {
						dstev := dstv.GetPointerAtIndexInt2(m.Store, i, bdt.Elt)
						srcev := src.TV.GetPointerAtIndexInt(m.Store, i)
						dstev.Assign2(m.Alloc, m.Store, m.Realm, srcev.Deref(), false)
					}
					res0 := TypedValue{
						T: IntType,
						V: nil,
					}
					res0.SetInt(minl)
					m.PushValue(res0)
					return
				case *SliceType:
					dstl := dst.TV.GetLength()
					srcl := src.TV.GetLength()
					minl := dstl
					if srcl < dstl {
						minl = srcl
					}
					if minl == 0 {
						// return 0.
						m.PushValue(defaultTypedValue(m.Alloc, IntType))
						return
					}
					dstv := dst.TV.V.(*SliceValue)
					srcv := src.TV.V.(*SliceValue)
					for i := 0; i < minl; i++ {
						dstev := dstv.GetPointerAtIndexInt2(m.Store, i, bdt.Elt)
						srcev := srcv.GetPointerAtIndexInt2(m.Store, i, bst.Elt)
						dstev.Assign2(m.Alloc, m.Store, m.Realm, srcev.Deref(), false)
					}
					res0 := TypedValue{
						T: IntType,
						V: nil,
					}
					res0.SetInt(minl)
					m.PushValue(res0)
					return
				case *NativeType:
					panic("copy from native slice not yet implemented") // XXX
				default:
					panic("should not happen")
				}
			case *NativeType:
				panic("copy to native slice not yet implemented") // XXX
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
			arg0, arg1 := m.LastBlock().GetParams2()
			itv := arg1.Deref()
			switch cbt := baseOf(arg0.TV.T).(type) {
			case *MapType:
				mv := arg0.TV.V.(*MapValue)
				mv.DeleteForKey(m.Store, &itv)
				return
			case *NativeType:
				krv := reflect.New(cbt.Type.Key()).Elem()
				krv = gno2GoValue(&itv, krv)
				mrv := arg0.TV.V.(*NativeValue).Value
				mrv.SetMapIndex(krv, reflect.Value{})
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
			arg0 := m.LastBlock().GetParams1()
			res0 := TypedValue{
				T: IntType,
				V: nil,
			}
			res0.SetInt(arg0.TV.GetLength())
			m.PushValue(res0)
			return
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
			arg0, arg1 := m.LastBlock().GetParams2()
			vargs := arg1
			vargsl := vargs.TV.GetLength()
			tt := arg0.TV.GetType()
			switch bt := baseOf(tt).(type) {
			case *SliceType:
				et := bt.Elem()
				if vargsl == 1 {
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := lv.ConvertGetInt()
					if et.Kind() == Uint8Kind {
						data := make([]byte, li)
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						list := make([]TypedValue, li)
						if et.Kind() == InterfaceKind {
							// leave as is
						} else {
							// init zero elements with concrete type.
							for i := 0; i < li; i++ {
								list[i] = defaultTypedValue(m.Alloc, et)
							}
						}
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}
				} else if vargsl == 2 {
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := lv.ConvertGetInt()
					cv := vargs.TV.GetPointerAtIndexInt(m.Store, 1).Deref()
					ci := cv.ConvertGetInt()
					if et.Kind() == Uint8Kind {
						data := make([]byte, li, ci)
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						list := make([]TypedValue, li, ci)
						if et := bt.Elem(); et.Kind() == InterfaceKind {
							// leave as is
						} else {
							// init zero elements with concrete type.
							// the elements beyond len l within cap c
							// must also be initialized, for a future
							// slice operation may refer to them.
							// XXX can this be removed?
							list2 := list[:ci]
							for i := 0; i < ci; i++ {
								list2[i] = defaultTypedValue(m.Alloc, et)
							}
						}
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}
				} else {
					panic("make() of slice type takes 2 or 3 arguments")
				}
			case *MapType:
				// NOTE: the type is not used.
				if vargsl == 0 {
					mv := &MapValue{}
					mv.MakeMap(0)
					m.Alloc.AllocateMap(0)
					m.PushValue(TypedValue{
						T: tt,
						V: mv,
					})
					return
				} else if vargsl == 1 {
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := lv.ConvertGetInt()
					mv := &MapValue{}
					mv.MakeMap(li)
					m.Alloc.AllocateMap(int64(li))
					m.PushValue(TypedValue{
						T: tt,
						V: mv,
					})
					return
				} else {
					panic("make() of map type takes 1 or 2 arguments")
				}
			case *ChanType:
				if vargsl == 0 {
					panic("not yet implemented")
				} else if vargsl == 1 {
					panic("not yet implemented")
				} else {
					panic("make() of chan type takes 1 or 2 arguments")
				}
			case *NativeType:
				switch bt.Type.Kind() {
				case reflect.Map:
					if vargsl == 0 {
						m.Alloc.AllocateNative()
						m.PushValue(TypedValue{
							T: tt,
							V: &NativeValue{
								Value: reflect.MakeMap(bt.Type),
							},
						})
						return
					} else if vargsl == 1 {
						sv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
						si := sv.ConvertGetInt()
						m.Alloc.AllocateNative()
						m.PushValue(TypedValue{
							T: tt,
							V: &NativeValue{
								Value: reflect.MakeMapWithSize(
									bt.Type, si),
							},
						})
						return
					} else {
						panic("make() of map type takes 1 or 2 arguments")
					}
				default:
					panic("not yet implemented")
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
			arg0 := m.LastBlock().GetParams1()
			tt := arg0.TV.GetType()
			vv := defaultValue(m.Alloc, tt)
			m.PushValue(TypedValue{
				T: &PointerType{
					Elt: tt,
				},
				V: PointerValue{
					TV: &TypedValue{
						T: tt,
						V: vv,
					},
					Base: nil,
				},
			})
			return
		},
	)
	defNative("panic",
		Flds( // params
			"err", AnyT(), // args[0]
		),
		nil, // results
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1()
			xv := arg0.Deref()
			panic(sprintString(m, &xv))
		},
	)
	defNative("print",
		Flds( // params
			"xs", Vrd(AnyT()), // args[0]
		),
		nil, // results
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1()
			xv := arg0
			xvl := xv.TV.GetLength()
			ss := make([]string, xvl)
			for i := 0; i < xvl; i++ {
				ev := xv.TV.GetPointerAtIndexInt(m.Store, i).Deref()
				ss[i] = sprintString(m, &ev)
			}
			rs := strings.Join(ss, " ")
			m.Output.Write([]byte(rs))
		},
	)
	defNative("println",
		Flds( // param
			"xs", Vrd(AnyT()), // args[0]
		),
		nil, // results
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1()
			xv := arg0
			xvl := xv.TV.GetLength()
			ss := make([]string, xvl)
			for i := 0; i < xvl; i++ {
				ev := xv.TV.GetPointerAtIndexInt(m.Store, i).Deref()
				ss[i] = sprintString(m, &ev)
			}
			rs := strings.Join(ss, " ") + "\n"
			m.Output.Write([]byte(rs))
		},
	)
	defNative("recover",
		nil, // params
		Flds( // results
			"exception", AnyT(),
		),
		func(m *Machine) {
			if m.Exception == nil {
				m.PushValue(TypedValue{})
				return
			} else {
				m.PushValue(*m.Exception)
				m.Exception = nil
				return
			}
		},
	)
	return uverseNode
}

// sprintString returns the string to be printed for tv from
// print() and println().
// XXX rename to sprintTypedValue.
func sprintString(m *Machine, tv *TypedValue) string {
	// if undefined, just "undefined".
	if tv.T == nil {
		return "undefined"
	}
	// if implements .String(), return it.
	if IsImplementedBy(gStringerType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "String")))
		return res[0].GetString()
	}
	// if implements .Error(), return it.
	if IsImplementedBy(gErrorType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "Error")))
		return res[0].GetString()
	}
	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			return fmt.Sprintf("%t", tv.GetBool())
		case UntypedStringType, StringType:
			return string(tv.GetString())
		case IntType:
			return fmt.Sprintf("%d", tv.GetInt())
		case Int8Type:
			return fmt.Sprintf("%d", tv.GetInt8())
		case Int16Type:
			return fmt.Sprintf("%d", tv.GetInt16())
		case UntypedRuneType, Int32Type:
			return fmt.Sprintf("%d", tv.GetInt32())
		case Int64Type:
			return fmt.Sprintf("%d", tv.GetInt64())
		case UintType:
			return fmt.Sprintf("%d", tv.GetUint())
		case Uint8Type:
			return fmt.Sprintf("%d", tv.GetUint8())
		case Uint16Type:
			return fmt.Sprintf("%d", tv.GetUint16())
		case Uint32Type:
			return fmt.Sprintf("%d", tv.GetUint32())
		case Uint64Type:
			return fmt.Sprintf("%d", tv.GetUint64())
		case UntypedBigintType, BigintType:
			return tv.V.(BigintValue).V.String()
		default:
			panic("should not happen")
		}
	case *PointerType:
		return tv.V.(PointerValue).String()
	case *ArrayType:
		return tv.V.(*ArrayValue).String()
	case *SliceType:
		return tv.V.(*SliceValue).String()
	case *StructType:
		return tv.V.(*StructValue).String()
	case *MapType:
		return tv.V.(*MapValue).String()
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			ft := tv.T.String()
			return "nil " + ft
		case *FuncValue:
			return fv.String()
		case *BoundMethodValue:
			return fv.String()
		default:
			panic(fmt.Sprintf(
				"unexpected func type %v",
				reflect.TypeOf(tv.V)))
		}
	case *InterfaceType:
		if debug {
			if tv.DebugHasValue() {
				panic("should not happen")
			}
		}
		return "nil"
	case *TypeType:
		return tv.V.(TypeValue).String()
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		return tv.V.(*PackageValue).String()
	case *ChanType:
		panic("not yet implemented")
		//return tv.V.(*ChanValue).String()
	case *NativeType:
		return fmt.Sprintf("%v",
			tv.V.(*NativeValue).Value.Interface())
	default:
		if debug {
			panic(fmt.Sprintf(
				"unexpected type %s",
				tv.T.String()))
		} else {
			panic("should not happen")
		}
	}
}

func copyDataToList(dst []TypedValue, data []byte, et Type) {
	for i := 0; i < len(data); i++ {
		dst[i] = TypedValue{T: et}
		dst[i].SetUint8(data[i])
	}
}

func copyListToData(dst []byte, tvs []TypedValue) {
	for i := 0; i < len(tvs); i++ {
		dst[i] = tvs[i].GetUint8()
	}
}

func copyListToRunes(dst []rune, tvs []TypedValue) {
	for i := 0; i < len(tvs); i++ {
		dst[i] = tvs[i].GetInt32()
	}
}

func copyNativeToList(alloc *Allocator, dst []TypedValue, rv reflect.Value, rvl int) {
	// TODO: redundant go2GnoType() conversions.
	for i := 0; i < rvl; i++ {
		dst[i] = go2GnoValue(alloc, rv.Index(i))
	}
}

func copyNativeToData(dst []byte, rv reflect.Value, rvl int) {
	for i := 0; i < rvl; i++ {
		dst[i] = uint8(rv.Index(i).Uint())
	}
}
