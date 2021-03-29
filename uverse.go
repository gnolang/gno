package gno

import (
	"fmt"
	"reflect"
	"strings"
)

var uverseNode *PackageNode

const uversePkgPath = ".uverse"

// Always returns a new copy from the latest state of source.
func Uverse() *PackageValue {
	pn := UverseNode()
	pv := pn.NewPackage(nil) // no realms in uverse.
	return pv
}

// Always returns the same instance with possibly differing completeness.
func UverseNode() *PackageNode {
	// Global is singleton.
	if uverseNode != nil {
		return uverseNode
	}
	fmt.Println("baking uverse...")

	// NOTE: uverse node is hidden, thus the leading dot in pkgPath=".uverse".
	uverseNode = NewPackageNode("uverse", uversePkgPath, nil)

	// temporary convenience function.
	def := func(n Name, tv TypedValue) {
		uverseNode.Define(n, tv)
	}

	// temporary convenience function; type is filled later by caller.
	defNative := func(n Name, ps, rs FieldTypeExprs, native func(*Machine)) {
		fd := FuncD(n, ps, rs, nil)
		// Preprocess sets v.Source.Name on .Source.StaticBlock.
		fd = Preprocess(nil, uverseNode, fd).(*FuncDecl)
		ft := evalType(uverseNode, &fd.Type).(*FuncType)
		if debug {
			if ft == nil {
				panic("should not happen")
			}
		}
		/*
			tv := TypedValue{
				T: ft,
				V: &FuncValue{
					Type:       ft,
					Source:     fd,
					Name:       n,
					NativeBody: native,
				},
			}
		*/
		// Set the native override function,
		// which doesn't get interpeted as it
		// doesn't exist in the declaration node.
		fv := uverseNode.GetValueRef(n).V.(*FuncValue)
		fv.NativeBody = native
		// fv.Closure, fv.pkg set during .NewPackage().
	}

	// Primitive types
	undefined := TypedValue{}
	def("_", undefined)    // special, path is zero.
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
	def("error", asValue(
		&DeclaredType{
			PkgPath: uversePkgPath,
			Name:    "error",
			Base:    &InterfaceType{}, // XXX error() string
			Methods: nil,
			sealed:  true,
		}))

	// Values
	def("true", untypedBool(true))
	def("false", untypedBool(false))

	// Functions
	defNative("append",
		Flds( // params
			"x", SliceT(GenT("X", nil)), // args[0]
			"args", Vrd(GenT("X", nil)), // args[1]
		),
		Flds( // results
			"res", SliceT(GenT("X", nil)), // res
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2()
			xt := arg0.T
			switch xv := arg0.V.(type) {

			//----------------------------------------------------------------
			// append(*SliceValue, ???)
			case *SliceValue:
				xvl := xv.Length
				xvo := xv.Offset
				xvc := xv.Maxcap
				switch args := arg1.V.(type) {

				//------------------------------------------------------------
				// append(*SliceValue, *SliceValue)
				case *SliceValue:
					argsl := args.Length
					argso := args.Offset
					if xvl+argsl <= xvc {
						// append(*SliceValue, *SliceValue) w/i capacity -----
						if xv.Base.Data == nil {
							// append(*SliceValue.List, *SliceValue) ---------
							list := xv.Base.List
							if args.Base.Data == nil {
								copy(
									list[xvo+xvl:xvo+xvl+argsl],
									args.Base.List[argso:argso+argsl])
							} else {
								copyDataToList(
									list[xvo+xvl:xvo+xvl+argsl],
									args.Base.Data[argso:argso+argsl],
									xt.Elem())
							}
						} else {
							// append(*SliceValue.Data, *SliceValue) ---------
							data := xv.Base.Data
							if args.Base.Data == nil {
								copyListToData(
									data[xvo+xvl:xvo+xvl+argsl],
									args.Base.List[argso:argso+argsl])
							} else {
								copy(
									data[xvo+xvl:xvo+xvl+argsl],
									args.Base.Data[argso:argso+argsl])
							}
						}
						m.PushValue(TypedValue{
							T: xt,
							V: &SliceValue{
								Base:   xv.Base,
								Offset: xvo,
								Length: xvl + argsl,
								Maxcap: xvc - argsl,
							},
						})
						return
					} else if xt.Kind() == Uint8Kind {
						// append(*SliceValue, *SliceValue) new data bytes ---
						data := make([]byte, xvl+argsl)
						if xv.Base.Data == nil {
							copyListToData(
								data[:xvl],
								xv.Base.List[xvo:xvo+xvl])
						} else {
							copy(
								data[:xvl],
								xv.Base.Data[xvo:xvo+xvl])
						}
						if args.Base.Data == nil {
							copyListToData(
								data[xvl:xvl+argsl],
								args.Base.List[argso:argso+argsl])
						} else {
							copy(
								data[xvl:xvl+argsl],
								args.Base.Data[argso:argso+argsl])
						}
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromData(data),
						})
						return
					} else {
						// append(*SliceValue, *SliceValue) new list ---------
						list := make([]TypedValue, xvl+argsl)
						if 0 < xvl {
							copy(
								list[:xvl],
								xv.Base.List[xvo:xvo+xvl])
						}
						if 0 < argsl {
							copy(
								list[xvl:xvl+argsl],
								args.Base.List[argso:argso+argsl])
						}
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromList(list),
						})
						return
					}

				//------------------------------------------------------------
				// append(*SliceValue, *nativeValue)
				case *nativeValue:
					argsrv := args.Value
					argsl := argsrv.Len()
					if xvl+argsl <= xvc {
						// append(*SliceValue, *nativeValue) w/i capacity ----
						if xv.Base.Data == nil {
							// append(*SliceValue.List, *nativeValue) --------
							list := xv.Base.List
							copyNativeToList(
								list[xvo:xvo+argsl],
								argsrv, argsl)
						} else {
							// append(*SliceValue.Data, *nativeValue) --------
							data := xv.Base.Data
							copyNativeToData(
								data[xvo:xvo+argsl],
								argsrv, argsl)
						}
						m.PushValue(TypedValue{
							T: xt,
							V: &SliceValue{
								Base:   xv.Base,
								Offset: xvo,
								Length: xvl + argsl,
								Maxcap: xvc - argsl,
							},
						})
						return
					} else if xt.Kind() == Uint8Kind {
						// append(*SliceValue, *nativeValue) new data bytes --
						data := make([]byte, xvl+argsl)
						if xv.Base.Data == nil {
							copyListToData(
								data[:xvl],
								xv.Base.List[xvo:xvo+xvl])
						} else {
							copy(
								data[:xvl],
								xv.Base.Data[xvo:xvo+xvl])
						}
						copyNativeToData(
							data[xvl:xvl+argsl],
							argsrv, argsl)
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromData(data),
						})
						return
					} else {
						// append(*SliceValue, *nativeValue) new list --------
						list := make([]TypedValue, xvl+argsl)
						copy(
							list[:xvl],
							xv.Base.List[xvo:xvo+xvl])
						copyNativeToList(
							list[xvl:xvl+argsl],
							argsrv, argsl)
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromList(list),
						})
						return
					}

				//------------------------------------------------------------
				default:
					panic("should not happen")

				}

			//----------------------------------------------------------------
			// append(*nativeValue, ???)
			case *nativeValue:
				sv := xv.Value
				switch args := arg1.V.(type) {

				//------------------------------------------------------------
				// append(*nativeValue, *SliceValue)
				case *SliceValue:
					st := sv.Type()
					argso := args.Offset
					argsl := args.Length
					argsrv := reflect.MakeSlice(st, argsl, argsl)
					if args.Base.Data == nil {
						for i := 0; i < argsl; i++ {
							etv := &(args.Base.List[argso+i])
							if etv.IsUndefined() {
								continue
							}
							erv := gno2GoValue(etv, reflect.Value{})
							argsrv.Index(i).Set(erv)
						}
					} else {
						for i := 0; i < argsl; i++ {
							erv := argsrv.Index(i)
							erv.SetUint(uint64(args.Base.Data[argso+i]))
						}
					}
					resrv := reflect.AppendSlice(sv, argsrv)
					m.PushValue(TypedValue{
						T: xt,
						V: &nativeValue{Value: resrv},
					})

				//------------------------------------------------------------
				// append(*nativeValue, *nativeValue)
				case *nativeValue:
					argsrv := args.Value
					resrv := reflect.AppendSlice(sv, argsrv)
					m.PushValue(TypedValue{
						T: xt,
						V: &nativeValue{Value: resrv},
					})

				//------------------------------------------------------------
				// append(*nativeValue, StringValue)
				case StringValue:
					if xt.Elem().Kind() == Uint8Kind {
						// TODO this might be faster if reflect supports
						// appending this way without first converting to a slice.
						argrv := reflect.ValueOf([]byte(arg1.V.(StringValue)))
						resrv := reflect.AppendSlice(sv, argrv)
						m.PushValue(TypedValue{
							T: xt,
							V: &nativeValue{Value: resrv},
						})
					} else {
						panic(fmt.Sprintf(
							"cannot append %s to %s",
							arg1.T.String(), xt.String()))
					}

				//------------------------------------------------------------
				// append(*nativeValue, ???)
				default:
					panic(fmt.Sprintf(
						"cannot append %s to %s",
						arg1.T.String(), xt.String()))

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
			res0.SetInt(arg0.GetCapacity())
			m.PushValue(res0)
		},
	)
	def("close", undefined)
	def("complex", undefined)
	defNative("copy",
		Flds( // params
			"dst", GenT("X", nil),
			"src", GenT("X", nil),
		),
		Flds( // results
			"", "int",
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2()
			dst, src := arg0, arg1
			switch bdt := baseOf(dst.T).(type) {
			case *SliceType:
				switch bst := baseOf(src.T).(type) {
				case PrimitiveType:
					if debug {
						debug.Println("copy(<%s>,<%s>)", bdt.String(), bst.String())
					}
					if bst.Kind() == StringKind {
						panic("not yet implemented")
					} else {
						panic("should not happen")
					}
				case *SliceType:
					dstl := dst.GetLength()
					srcl := src.GetLength()
					minl := dstl
					if srcl < dstl {
						minl = srcl
					}
					if minl == 0 {
						return // do nothing.
					}
					dstv := dst.V.(*SliceValue)
					srcv := src.V.(*SliceValue)
					for i := 0; i < minl; i++ {
						dstev := dstv.GetPointerAtIndexInt2(i, bdt)
						srcev := srcv.GetPointerAtIndexInt2(i, bst)
						dstev.Assign(srcev.Deref())
					}
					res0 := TypedValue{
						T: IntType,
						V: nil,
					}
					res0.SetInt(minl)
					m.PushValue(res0)
				default:
					panic("should not happen")
				}
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
			switch baseOf(arg0.T).(type) {
			case *MapType:
				mv := arg0.V.(*MapValue)
				mv.DeleteForKey(&itv)
			case *nativeType:
				krv := gno2GoValue(&itv, reflect.Value{})
				mrv := arg0.V.(*nativeValue).Value
				mrv.SetMapIndex(krv, reflect.Value{})
			default:
				panic(fmt.Sprintf(
					"unexpected map type %s",
					arg0.T.String()))
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
			res0.SetInt(arg0.GetLength())
			m.PushValue(res0)
		},
	)
	defNative("make",
		Flds( // params
			"t", GenT("T.(type)", nil),
			"z", Vrd("int"),
		),
		Flds( // results
			"", GenT("T", nil),
		),
		func(m *Machine) {
			arg0, arg1 := m.LastBlock().GetParams2()
			vargs := arg1
			vargsl := vargs.GetLength()
			tt := arg0.GetType()
			switch bt := baseOf(tt).(type) {
			case *SliceType:
				if vargsl == 1 {
					lv := vargs.GetPointerAtIndexInt(0).Deref()
					li := lv.GetInt()
					list := make([]TypedValue, li)
					if et := bt.Elem(); et.Kind() == InterfaceKind {
						// leave as is
					} else {
						// init zero elements with concrete type.
						// XXX can this be removed?
						for i := 0; i < li; i++ {
							list[i].T = et
						}
					}
					m.PushValue(TypedValue{
						T: tt,
						V: newSliceFromList(list),
					})
					return
				} else if vargsl == 2 {
					lv := vargs.GetPointerAtIndexInt(0).Deref()
					li := lv.GetInt()
					cv := vargs.GetPointerAtIndexInt(1).Deref()
					ci := cv.GetInt()
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
							list2[i].T = et
						}
					}
					m.PushValue(TypedValue{
						T: tt,
						V: newSliceFromList(list),
					})
					return
				} else {
					panic("make() of slice type takes 2 or 3 arguments")
				}
			case *MapType:
				// NOTE: the type is not used.
				if vargsl == 0 {
					mv := &MapValue{}
					mv.MakeMap(0)
					m.PushValue(TypedValue{
						T: tt,
						V: mv,
					})
					return
				} else if vargsl == 1 {
					lv := vargs.GetPointerAtIndexInt(0).Deref()
					li := lv.GetInt()
					mv := &MapValue{}
					mv.MakeMap(li)
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
			default:
				panic(fmt.Sprintf(
					"cannot make type %s kind %v",
					tt.String(), tt.Kind()))
			}
		},
	)
	def("new", undefined)
	defNative("panic",
		Flds( // params
			"err", AnyT(), // args[0]
		),
		nil, // results
		func(m *Machine) {
			arg0 := m.LastBlock().GetParams1()
			xv := arg0.Deref()
			panic(sprintString(&xv))
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
			xvl := xv.GetLength()
			ss := make([]string, xvl)
			for i := 0; i < xvl; i++ {
				ev := xv.GetPointerAtIndexInt(i).Deref()
				ss[i] = sprintString(&ev)
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
			xvl := xv.GetLength()
			ss := make([]string, xvl)
			for i := 0; i < xvl; i++ {
				ev := xv.GetPointerAtIndexInt(i).Deref()
				ss[i] = sprintString(&ev)
			}
			rs := strings.Join(ss, " ") + "\n"
			m.Output.Write([]byte(rs))
		},
	)
	def("recover", undefined)
	return uverseNode
}

// sprintString returns the string to be printed for tv from
// print() and println().
func sprintString(tv *TypedValue) string {
	if tv.T == nil {
		return "undefined"
	}
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
	case PointerType:
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
		case BoundMethodValue:
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
	case *nativeType:
		return fmt.Sprintf("%v",
			tv.V.(*nativeValue).Value.Interface())
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

func copyNativeToList(dst []TypedValue, rv reflect.Value, rvl int) {
	// TODO: redundant go2GnoType() conversions.
	for i := 0; i < rvl; i++ {
		dst[i] = go2GnoValue(rv.Index(i))
	}
}

func copyNativeToData(dst []byte, rv reflect.Value, rvl int) {
	for i := 0; i < rvl; i++ {
		dst[i] = uint8(rv.Index(i).Uint())
	}
}
