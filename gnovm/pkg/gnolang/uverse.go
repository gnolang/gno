package gnolang

import (
	"fmt"
	"reflect"
	"strings"
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

// ----------------------------------------
// Uverse package

var (
	uverseNode  *PackageNode
	uverseValue *PackageValue
)

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
	def("bigint", asValue(BigintType))
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
						data := make([]byte, arg1Length)
						if arg1Base.Data == nil {
							copyListToData(
								data[:arg1Length],
								arg1Base.List[arg1Offset:arg1EndIndex])
						} else {
							copy(
								data[:arg1Length],
								arg1Base.Data[arg1Offset:arg1EndIndex])
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(nil, *SliceValue) new list ---------
						list := make([]TypedValue, arg1Length)
						if 0 < arg1Length {
							for i := 0; i < arg1Length; i++ {
								list[i] = arg1Base.List[arg1Offset+i].unrefCopy(m.Alloc, m.Store)
							}
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				// ------------------------------------------------------------
				// append(nil, *NativeValue)
				case *NativeValue:
					arg1NativeValue := arg1Value.Value
					arg1NativeValueLength := arg1NativeValue.Len()
					if arg1NativeValueLength == 0 { // no change
						m.PushValue(TypedValue{
							T: arg0Type,
							V: nil,
						})
						return
					} else if arg0Type.Elem().Kind() == Uint8Kind {
						// append(nil, *NativeValue) new data bytes --
						data := make([]byte, arg1NativeValueLength)
						copyNativeToData(
							data[:arg1NativeValueLength],
							arg1NativeValue, arg1NativeValueLength)
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(nil, *NativeValue) new list --------
						list := make([]TypedValue, arg1NativeValueLength)
						if 0 < arg1NativeValueLength {
							copyNativeToList(
								m.Alloc,
								list[:arg1NativeValueLength],
								arg1NativeValue, arg1NativeValueLength)
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				// ------------------------------------------------------------
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
							if arg0Base.Data == nil {
								// append(*SliceValue.List, *SliceValue) ---------
								list := arg0Base.List
								if arg1Base.Data == nil {
									for i := 0; i < arg1Length; i++ {
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
									m.Realm.DidUpdate(arg1Base, nil, nil)
								}
							} else {
								// append(*SliceValue.Data, *SliceValue) ---------
								data := arg0Base.Data
								if arg1Base.Data == nil {
									copyListToData(
										data[arg0Offset+arg0Length:arg0Offset+arg0Length+arg1Length],
										arg1Base.List[arg1Offset:arg1Offset+arg1Length])
									m.Realm.DidUpdate(arg0Base, nil, nil)
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
						data := make([]byte, arg0Length+arg1Length)
						if 0 < arg0Length {
							if arg0Base.Data == nil {
								copyListToData(
									data[:arg0Length],
									arg0Base.List[arg0Offset:arg0Offset+arg0Length])
							} else {
								copy(
									data[:arg0Length],
									arg0Base.Data[arg0Offset:arg0Offset+arg0Length])
							}
						}
						if 0 < arg1Length {
							if arg1Base.Data == nil {
								copyListToData(
									data[arg0Length:arg0Length+arg1Length],
									arg1Base.List[arg1Offset:arg1Offset+arg1Length])
							} else {
								copy(
									data[arg0Length:arg0Length+arg1Length],
									arg1Base.Data[arg1Offset:arg1Offset+arg1Length])
							}
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(*SliceValue, *SliceValue) new list ---------
						list := make([]TypedValue, arg0Length+arg1Length)
						if 0 < arg0Length {
							if arg0Base.Data == nil {
								for i := 0; i < arg0Length; i++ {
									list[i] = arg0Base.List[arg0Offset+i].unrefCopy(m.Alloc, m.Store)
								}
							} else {
								panic("should not happen")
							}
						}

						if 0 < arg1Length {
							if arg1Base.Data == nil {
								for i := 0; i < arg1Length; i++ {
									list[arg0Length+i] = arg1Base.List[arg1Offset+i].unrefCopy(m.Alloc, m.Store)
								}
							} else {
								copyDataToList(
									list[arg0Length:arg0Length+arg1Length],
									arg1Base.Data[arg1Offset:arg1Offset+arg1Length],
									arg1Type.Elem(),
								)
							}
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				// ------------------------------------------------------------
				// append(*SliceValue, *NativeValue)
				case *NativeValue:
					arg1NativeValue := arg1Value.Value
					arg1NativeValueLength := arg1NativeValue.Len()
					if arg0Length+arg1NativeValueLength <= arg0Capacity {
						// append(*SliceValue, *NativeValue) w/i capacity ----
						if 0 < arg1NativeValueLength { // implies 0 < xvc
							if arg0Base.Data == nil {
								// append(*SliceValue.List, *NativeValue) --------
								list := arg0Base.List
								copyNativeToList(
									m.Alloc,
									list[arg0Offset:arg0Offset+arg1NativeValueLength],
									arg1NativeValue, arg1NativeValueLength)
							} else {
								// append(*SliceValue.Data, *NativeValue) --------
								data := arg0Base.Data
								copyNativeToData(
									data[arg0Offset:arg0Offset+arg1NativeValueLength],
									arg1NativeValue, arg1NativeValueLength)
							}
							m.PushValue(TypedValue{
								T: arg0Type,
								V: m.Alloc.NewSlice(arg0Base, arg0Offset, arg0Length+arg1NativeValueLength, arg0Capacity),
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
						// append(*SliceValue, *NativeValue) new data bytes --
						data := make([]byte, arg0Length+arg1NativeValueLength)
						if 0 < arg0Length {
							if arg0Base.Data == nil {
								copyListToData(
									data[:arg0Length],
									arg0Base.List[arg0Offset:arg0Offset+arg0Length])
							} else {
								copy(
									data[:arg0Length],
									arg0Base.Data[arg0Offset:arg0Offset+arg0Length])
							}
						}
						if 0 < arg1NativeValueLength {
							copyNativeToData(
								data[arg0Length:arg0Length+arg1NativeValueLength],
								arg1NativeValue, arg1NativeValueLength)
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromData(data),
						})
						return
					} else {
						// append(*SliceValue, *NativeValue) new list --------
						listLen := arg0Length + arg1NativeValueLength
						list := make([]TypedValue, listLen)
						if 0 < arg0Length {
							for i := 0; i < listLen; i++ {
								list[i] = arg0Base.List[arg0Offset+i].unrefCopy(m.Alloc, m.Store)
							}
						}
						if 0 < arg1NativeValueLength {
							copyNativeToList(
								m.Alloc,
								list[arg0Length:listLen],
								arg1NativeValue, arg1NativeValueLength)
						}
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewSliceFromList(list),
						})
						return
					}

				// ------------------------------------------------------------
				default:
					panic("should not happen")
				}

			// ----------------------------------------------------------------
			// append(*NativeValue, ???)
			case *NativeValue:
				arg0NativeValue := arg0Value.Value
				switch arg1Value := arg1.TV.V.(type) {
				// ------------------------------------------------------------
				// append(*NativeValue, nil)
				case nil: // no change
					m.PushValue(TypedValue{
						T: arg0Type,
						V: arg0Value,
					})
					return

				// ------------------------------------------------------------
				// append(*NativeValue, *SliceValue)
				case *SliceValue:
					arg0NativeValueType := arg0NativeValue.Type()
					arg1Offset := arg1Value.Offset
					arg1Length := arg1Value.Length
					arg1Base := arg1Value.GetBase(m.Store)
					if 0 < arg1Length {
						newNativeArg1Slice := reflect.MakeSlice(arg0NativeValueType, arg1Length, arg1Length)
						if arg1Base.Data == nil {
							for i := 0; i < arg1Length; i++ {
								etv := &(arg1Base.List[arg1Offset+i])
								if etv.IsUndefined() {
									continue
								}
								erv := gno2GoValue(etv, reflect.Value{})
								newNativeArg1Slice.Index(i).Set(erv)
							}
						} else {
							for i := 0; i < arg1Length; i++ {
								erv := newNativeArg1Slice.Index(i)
								erv.SetUint(uint64(arg1Base.Data[arg1Offset+i]))
							}
						}
						modifiedNativeSlice := reflect.AppendSlice(arg0NativeValue, newNativeArg1Slice)
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewNative(modifiedNativeSlice),
						})
						return
					} else { // no change
						m.PushValue(TypedValue{
							T: arg0Type,
							V: arg0Value,
						})
						return
					}

				// ------------------------------------------------------------
				// append(*NativeValue, *NativeValue)
				case *NativeValue:
					arg1ReflectValue := arg1Value.Value
					modifiedNativeSlice := reflect.AppendSlice(arg0NativeValue, arg1ReflectValue)
					m.PushValue(TypedValue{
						T: arg0Type,
						V: m.Alloc.NewNative(modifiedNativeSlice),
					})
					return

				// ------------------------------------------------------------
				// append(*NativeValue, StringValue)
				case StringValue:
					if arg0Type.Elem().Kind() == Uint8Kind {
						// TODO this might be faster if reflect supports
						// appending this way without first converting to a slice.
						arg1ReflectValue := reflect.ValueOf([]byte(arg1.TV.GetString()))
						modifiedNativeSlice := reflect.AppendSlice(arg0NativeValue, arg1ReflectValue)
						m.PushValue(TypedValue{
							T: arg0Type,
							V: m.Alloc.NewNative(modifiedNativeSlice),
						})
						return
					} else {
						panic(fmt.Sprintf(
							"cannot append %s to %s",
							arg1.TV.T.String(), arg0Type.String()))
					}

				// ------------------------------------------------------------
				// append(*NativeValue, ???)
				default:
					panic(fmt.Sprintf(
						"cannot append %s to %s",
						arg1.TV.T.String(), arg0Type.String()))
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
					m.PushValue(TypedValue{
						T: tt,
						V: m.Alloc.NewMap(0),
					})
					return
				} else if vargsl == 1 {
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := lv.ConvertGetInt()
					m.PushValue(TypedValue{
						T: tt,
						V: m.Alloc.NewMap(li),
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
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewNative(
								reflect.MakeMap(bt.Type),
							),
						})
						return
					} else if vargsl == 1 {
						sv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
						si := sv.ConvertGetInt()
						m.PushValue(TypedValue{
							T: tt,
							V: m.Alloc.NewNative(
								reflect.MakeMapWithSize(
									bt.Type, si),
							),
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
			m.Alloc.AllocatePointer()
			m.PushValue(TypedValue{
				T: m.Alloc.NewType(&PointerType{
					Elt: tt,
				}),
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
			panic(xv.Sprint(m))
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
				ss[i] = ev.Sprint(m)
			}
			rs := strings.Join(ss, " ")
			if debug {
				print(rs)
			}
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
				ss[i] = ev.Sprint(m)
			}
			rs := strings.Join(ss, " ") + "\n"
			if debug {
				println("DEBUG/stdout: " + rs)
			}
			m.Output.Write([]byte(rs))
		},
	)
	defNative("recover",
		nil, // params
		Flds( // results
			"exception", AnyT(),
		),
		func(m *Machine) {
			if len(m.Exceptions) == 0 {
				m.PushValue(TypedValue{})
				return
			}

			// If the exception is out of scope, this recover can't help; return nil.
			if m.PanicScope <= m.DeferPanicScope {
				m.PushValue(TypedValue{})
				return
			}

			exception := &m.Exceptions[len(m.Exceptions)-1]

			// If the frame the exception occurred in is not popped, it's possible that
			// the exception is still in scope and can be recovered.
			if !exception.Frame.Popped {
				// If the frame is not the current frame, the exception is not in scope; return nil.
				// This retrieves the second most recent call frame because the first most recent
				// is the call to recover itself.
				if frame := m.LastCallFrame(2); frame == nil || (frame != nil && frame != exception.Frame) {
					m.PushValue(TypedValue{})
					return
				}
			}

			m.PushValue(exception.Value)
			// Recover complete; remove exceptions.
			m.Exceptions = nil
		},
	)
	return uverseNode
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
