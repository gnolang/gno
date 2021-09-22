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
		ft := evalStaticType(nil, uverseNode, &fd.Type).(*FuncType)
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
					SourceLoc:  fd.GetLocation(),
					Source:     fd,
					Name:       n,
					NativeBody: native,
				},
			}
		*/
		// Set the native override function,
		// which doesn't get interpeted as it
		// doesn't exist in the declaration node.
		fv := uverseNode.GetValueRef(nil, n).V.(*FuncValue)
		fv.nativeBody = native
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
	def("error", asValue(
		&DeclaredType{
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
			xt := arg0.TV.T
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
					} else if xt.Kind() == Uint8Kind {
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
							V: newSliceFromData(data),
						})
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
							V: newSliceFromList(list),
						})
					}

				//------------------------------------------------------------
				// append(nil, *nativeValue)
				case *nativeValue:
					argsrv := args.Value
					argsl := argsrv.Len()
					if argsl == 0 { // no change
						m.PushValue(TypedValue{
							T: xt,
							V: nil,
						})
					} else if xt.Kind() == Uint8Kind {
						// append(nil, *nativeValue) new data bytes --
						data := make([]byte, argsl)
						copyNativeToData(
							data[:argsl],
							argsrv, argsl)
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromData(data),
						})
					} else {
						// append(nil, *nativeValue) new list --------
						list := make([]TypedValue, argsl)
						if 0 < argsl {
							copyNativeToList(
								list[:argsl],
								argsrv, argsl)
						}
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromList(list),
						})
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
							m.PushValue(TypedValue{
								T: xt,
								V: &SliceValue{
									Base:   xvb,
									Offset: xvo,
									Length: xvl + argsl,
									Maxcap: xvc - argsl,
								},
							})
						} else { // no change
							m.PushValue(TypedValue{
								T: xt,
								V: xv,
							})
						}
					} else if xt.Kind() == Uint8Kind {
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
							V: newSliceFromData(data),
						})
					} else {
						// append(*SliceValue, *SliceValue) new list ---------
						list := make([]TypedValue, xvl+argsl)
						if 0 < xvl {
							copy(
								list[:xvl],
								xvb.List[xvo:xvo+xvl])
						}
						if 0 < argsl {
							copy(
								list[xvl:xvl+argsl],
								argsb.List[argso:argso+argsl])
						}
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromList(list),
						})
					}

				//------------------------------------------------------------
				// append(*SliceValue, *nativeValue)
				case *nativeValue:
					argsrv := args.Value
					argsl := argsrv.Len()
					if xvl+argsl <= xvc {
						// append(*SliceValue, *nativeValue) w/i capacity ----
						if 0 < argsl { // implies 0 < xvc
							if xvb.Data == nil {
								// append(*SliceValue.List, *nativeValue) --------
								list := xvb.List
								copyNativeToList(
									list[xvo:xvo+argsl],
									argsrv, argsl)
							} else {
								// append(*SliceValue.Data, *nativeValue) --------
								data := xvb.Data
								copyNativeToData(
									data[xvo:xvo+argsl],
									argsrv, argsl)
							}
							m.PushValue(TypedValue{
								T: xt,
								V: &SliceValue{
									Base:   xvb,
									Offset: xvo,
									Length: xvl + argsl,
									Maxcap: xvc - argsl,
								},
							})
						} else { // no change
							m.PushValue(TypedValue{
								T: xt,
								V: xv,
							})
						}
					} else if xt.Kind() == Uint8Kind {
						// append(*SliceValue, *nativeValue) new data bytes --
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
							V: newSliceFromData(data),
						})
					} else {
						// append(*SliceValue, *nativeValue) new list --------
						list := make([]TypedValue, xvl+argsl)
						if 0 < xvl {
							copy(
								list[:xvl],
								xvb.List[xvo:xvo+xvl])
						}
						if 0 < argsl {
							copyNativeToList(
								list[xvl:xvl+argsl],
								argsrv, argsl)
						}
						m.PushValue(TypedValue{
							T: xt,
							V: newSliceFromList(list),
						})
					}

				//------------------------------------------------------------
				default:
					panic("should not happen")

				}

			//----------------------------------------------------------------
			// append(*nativeValue, ???)
			case *nativeValue:
				sv := xv.Value
				switch args := arg1.TV.V.(type) {

				//------------------------------------------------------------
				// append(*nativeValue, nil)
				case nil: // no change
					m.PushValue(TypedValue{
						T: xt,
						V: xv,
					})

				//------------------------------------------------------------
				// append(*nativeValue, *SliceValue)
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
						m.PushValue(TypedValue{
							T: xt,
							V: &nativeValue{Value: resrv},
						})
					} else { // no change
						m.PushValue(TypedValue{
							T: xt,
							V: xv,
						})
					}

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
						argrv := reflect.ValueOf([]byte(arg1.TV.V.(StringValue)))
						resrv := reflect.AppendSlice(sv, argrv)
						m.PushValue(TypedValue{
							T: xt,
							V: &nativeValue{Value: resrv},
						})
					} else {
						panic(fmt.Sprintf(
							"cannot append %s to %s",
							arg1.TV.T.String(), xt.String()))
					}

				//------------------------------------------------------------
				// append(*nativeValue, ???)
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
			switch bdt := baseOf(dst.TV.T).(type) {
			case *SliceType:
				switch bst := baseOf(src.TV.T).(type) {
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
					dstl := dst.TV.GetLength()
					srcl := src.TV.GetLength()
					minl := dstl
					if srcl < dstl {
						minl = srcl
					}
					if minl == 0 {
						return // do nothing.
					}
					dstv := dst.TV.V.(*SliceValue)
					srcv := src.TV.V.(*SliceValue)
					for i := 0; i < minl; i++ {
						dstev := dstv.GetPointerAtIndexInt2(m.Store, i, bdt.Elt)
						srcev := srcv.GetPointerAtIndexInt2(m.Store, i, bst.Elt)
						dstev.TV.Assign(srcev.Deref(), false)
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
			switch baseOf(arg0.TV.T).(type) {
			case *MapType:
				mv := arg0.TV.V.(*MapValue)
				mv.DeleteForKey(m.Store, &itv)
			case *nativeType:
				krv := gno2GoValue(&itv, reflect.Value{})
				mrv := arg0.TV.V.(*nativeValue).Value
				mrv.SetMapIndex(krv, reflect.Value{})
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
				if vargsl == 1 {
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := lv.ConvertGetInt()
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
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := lv.ConvertGetInt()
					cv := vargs.TV.GetPointerAtIndexInt(m.Store, 1).Deref()
					ci := cv.ConvertGetInt()
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
					lv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
					li := lv.ConvertGetInt()
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
			case *nativeType:
				switch bt.Type.Kind() {
				case reflect.Map:
					if vargsl == 0 {
						m.PushValue(TypedValue{
							T: tt,
							V: &nativeValue{
								Value: reflect.MakeMap(bt.Type),
							},
						})
						return
					} else if vargsl == 1 {
						sv := vargs.TV.GetPointerAtIndexInt(m.Store, 0).Deref()
						si := sv.ConvertGetInt()
						m.PushValue(TypedValue{
							T: tt,
							V: &nativeValue{
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
			vv := defaultValue(tt)
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
			xvl := xv.TV.GetLength()
			ss := make([]string, xvl)
			for i := 0; i < xvl; i++ {
				ev := xv.TV.GetPointerAtIndexInt(m.Store, i).Deref()
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
			xvl := xv.TV.GetLength()
			ss := make([]string, xvl)
			for i := 0; i < xvl; i++ {
				ev := xv.TV.GetPointerAtIndexInt(m.Store, i).Deref()
				ss[i] = sprintString(&ev)
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
			} else {
				m.PushValue(*m.Exception)
				m.Exception = nil
			}
		},
	)
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

func copyListToRunes(dst []rune, tvs []TypedValue) {
	for i := 0; i < len(tvs); i++ {
		dst[i] = tvs[i].GetInt32()
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
