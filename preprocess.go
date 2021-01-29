package gno

import (
	"fmt"
	"math/big"
	"reflect"
)

// The ctx passed in may be mutated if there are any statements or
// declarations.  The file or package which contains ctx may be mutated if
// there are any file-level declarations.
//
// List of what Preprocess() does:
//  * Assigns BlockValuePath to NameExprs.
//  * TODO document what it does.
func Preprocess(imp Importer, ctx BlockNode, n Node) Node {
	if ctx == nil {
		panic("Preprocess requires context")
	}

	// create stack of BlockNodes.
	var stack []BlockNode = make([]BlockNode, 0, 32)
	var last BlockNode = ctx
	stack = append(stack, last)

	// iterate over all nodes recursively and calculate
	// BlockValuePath for each NameExpr.
	nn := Transcribe(n, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {

		defer func() {
			if r := recover(); r != nil {
				for i, sbn := range stack {
					fmt.Printf("stack #%d: %s\n", i, sbn.String())
				}
				panic(r)
			}
		}()
		if debug {
			debug.Printf("Transcribe %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		// if already preprocessed, break and do nothing.
		if n.GetAttribute(ATTR_PREPROCESSED) == true {
			return n, TRANS_BREAK
		}

		switch stage {

		//----------------------------------------
		case TRANS_ENTER:
			switch n := n.(type) {

			// TRANS_ENTER -----------------------
			case *AssignStmt:
				if n.Op == DEFINE {
					for _, lx := range n.Lhs {
						ln := lx.(*NameExpr).Name
						if ln == "_" {
							panic("cannot define special name \"_\"")
						}
						// initial declaration to be re-defined.
						last.Define(ln, anyValue(nil))
					}
				} else {
					// nothing defined.
				}

			// TRANS_ENTER -----------------------
			case *ImportDecl, *ValueDecl, *TypeDecl, *FuncDecl:
				// NOTE func decl usually must happen with a file, and so
				// last is usually a *FileNode, but for testing
				// convenience we allow importing directly onto the
				// package. Uverse requires this.
				if n.GetAttribute(ATTR_PREDEFINED) == true {
					// skip declarations already predefined
					// (e.g. through recursion for a dependent)
				} else {
					// recursively predefine dependencies.
					predefineNow(imp, last, n.(Decl))
				}

			}

			// TRANS_ENTER -----------------------
			return n, TRANS_CONTINUE

		//----------------------------------------
		case TRANS_BLOCK:

			switch n := n.(type) {

			// TRANS_BLOCK -----------------------
			case *BlockStmt:
				pushBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *ForStmt:
				pushBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *IfStmt:
				pushBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *RangeStmt:
				pushBlock(n, &last, &stack)
				// key value if define.
				if n.Op == DEFINE {
					if n.Key != nil {
						kn := n.Key.(*NameExpr).Name
						last.Define(kn, anyValue(IntType))
					}
					if n.Value != nil {
						// initial declaration to be re-defined.
						vn := n.Value.(*NameExpr).Name
						last.Define(vn, anyValue(nil))
					}
				}

			// TRANS_BLOCK -----------------------
			case *FuncLitExpr:
				// retrieve cached function type.
				ft := evalType(last, &n.Type).(*FuncType)
				// push func body block.
				pushBlock(n, &last, &stack)
				// define parameters in new block.
				for _, p := range ft.Params {
					last.Define(p.Name, anyValue(p.Type))
				}
				// define results in new block.
				for i, r := range ft.Results {
					if 0 < len(r.Name) {
						last.Define(r.Name, anyValue(r.Type))
					} else {
						// create a hidden var with leading dot.
						// NOTE: document somewhere.
						rn := fmt.Sprintf(".res_%d", i)
						last.Define(Name(rn), anyValue(r.Type))
					}
				}

			// TRANS_BLOCK -----------------------
			case *SelectCaseStmt:
				pushBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *SwitchStmt:
				// create faux block to store .Init/.Varname.
				// the contents are copied onto the case block
				// in the switch case below for switch cases.
				pushBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *SwitchCaseStmt:
				pushBlock(n, &last, &stack)
				// parent switch statement.
				ss := ns[len(ns)-1].(*SwitchStmt)
				// anything declared in ss are copied.
				for _, n := range ss.GetNames() {
					tv := ss.GetValueRef(n)
					last.Define(n, *tv)
				}
				// maybe type-switch def.
				if 0 < len(ss.VarName) {
					// if there is only 1 case, the define applies.  if
					// there are multiple, the definition is void(?),
					// TODO TestSwitchDefine
					if len(n.Cases) == 1 {
						ct := evalType(last, n.Cases[0])
						last.Define(ss.VarName, anyValue(ct))
					}
				}

			// TRANS_BLOCK -----------------------
			case *FuncDecl:
				// retrieve cached function type.
				ft := evalType(last, &n.Type).(*FuncType)
				if n.IsMethod {
					// set method onto declared type.
					rft := evalType(last, &n.Recv).(FieldType)
					ft = ft.UnboundType(rft)
					rt := rft.Type
					dt := (*DeclaredType)(nil)
					if pt, ok := rt.(PointerType); ok {
						dt = pt.Elem().(*DeclaredType)
					} else {
						dt = rt.(*DeclaredType)
					}
					dt.DefineMethod(&FuncValue{
						Type:       ft,
						IsMethod:   true,
						Source:     n,
						Name:       n.Name,
						Body:       n.Body,
						Closure:    nil, // set later.
						NativeBody: nil,
						FileName:   filenameOf(last),
						pkg:        nil, // set later.
					})
				} else {
					// type fills in @ predefineNow().
					/*
						// fill in ft constructed at
						// *FuncDecl:ENTER.
						tv := last.GetValueRef(n.Name)
						fv := tv.V.(*FuncValue)
						*(fv.Type) = *ft
					*/
				}

				// push func body block.
				parent := last
				pushBlock(n, &last, &stack)
				// define receiver in new block, if method.
				if n.IsMethod {
					if 0 < len(n.Recv.Name) {
						rt := evalType(parent, n.Recv.Type)
						last.Define(n.Recv.Name, anyValue(rt))
					}
				}
				// define parameters in new block.
				for _, p := range ft.Params {
					last.Define(p.Name, anyValue(p.Type))
				}
				// define results in new block.
				for i, r := range ft.Results {
					if 0 < len(r.Name) {
						last.Define(r.Name, anyValue(r.Type))
					} else {
						// create a hidden var with leading dot.
						rn := fmt.Sprintf(".res_%d", i)
						last.Define(Name(rn), anyValue(r.Type))
					}
				}

			// TRANS_BLOCK -----------------------
			case *FileNode:
				// only for imports.
				pushBlock(n, &last, &stack)
				{
					// support out-of-order declarations.  this is required
					// separately from the direct predefineNow() entry
					// callbacks above, for otherwise out-of-order declarations
					// would not get pre-defined before (say) the body of a
					// function declaration or literl can refer to it.
					// (this must happen after pushBlock above, otherwise it
					// would happen @ *FileNode:ENTER)
					for _, d := range n.Body {
						if d.GetAttribute(ATTR_PREDEFINED) == true {
							// skip declarations already predefined
							// (e.g. through recursion for a dependent)
						} else {
							// recursively predefine dependencies.
							predefineNow(imp, n, d)
						}
					}
				}

			// TRANS_BLOCK -----------------------
			default:
				panic("should not happen")
			}
			return n, TRANS_CONTINUE

		//----------------------------------------
		case TRANS_LEAVE:
			// mark as preprocessed so that it can be used
			// in evalType().
			n.SetAttribute(ATTR_PREPROCESSED, true)

			//-There is still work to be done while leaving, but once the
			//logic of that is done, we will have to perform additionally
			//deferred logic that is best handled with orthogonal switch
			//conditions.
			//-For example, while leaving nodes w/ TRANS_COMPOSITE_TYPE,
			//(regardless of whether name or literal), any elided type
			//names are inserted. (This works because the transcriber
			//leaves the composite type before entering the kv elements.)
			defer func() {
				switch ftype {

				// TRANS_LEAVE (deferred)---------
				case TRANS_COMPOSITE_TYPE:
					// fill elided element composite lit type exprs.
					clx := ns[len(ns)-1].(*CompositeLitExpr)
					// get or evaluate composite type.
					clt := evalType(last, n.(Expr))
					// elide composite lit element (nested) composite types.
					elideCompositeElements(clx, clt)
				}
			}()

			// The main TRANS_LEAVE switch.
			switch n := n.(type) {

			// TRANS_LEAVE -----------------------
			case *NameExpr:
				// special case if struct composite key.
				if ftype == TRANS_COMPOSITE_KEY {
					clx := ns[len(ns)-1].(*CompositeLitExpr)
					clt := evalType(last, clx.Type)
					switch bt := baseOf(clt).(type) {
					case *StructType:
						n.Path = bt.GetPathForName(n.Name)
						return n, TRANS_CONTINUE
					case *ArrayType, *SliceType:
						// Replace n with *constExpr.
						fillNameExprPath(last, n)
						cv := evalConst(last, n)
						return cv, TRANS_CONTINUE
					case *nativeType:
						switch bt.Type.Kind() {
						case reflect.Struct:
							// NOTE Gno embedded fields are flattened,
							// whereas in Go fields are nested, and a
							// complete index is a slice of indices.  For
							// simplicity and some degree of flexibility,
							// do not use path indices for Go native
							// types, but use the name.
							n.Path = NewValuePath(n.Name, 0, 0)
							return n, TRANS_CONTINUE
						case reflect.Array, reflect.Slice:
							// Replace n with *constExpr.
							fillNameExprPath(last, n)
							cv := evalConst(last, n)
							return cv, TRANS_CONTINUE
						default:
							panic("should not happen")
						}
					}
				}
				// specific and general cases
				switch n.Name {
				case "_":
					return n, TRANS_CONTINUE
				case "iota":
					pd := lastDecl(ns)
					io := pd.GetAttribute(ATTR_IOTA).(int)
					cx := constUntypedBigint(n, int64(io))
					return cx, TRANS_CONTINUE
				case "nil":
					// nil will be converted to typed-nils when appropriate
					// upon leaving the expression nodes that contain nil
					// nodes.
					fallthrough
				default:
					fillNameExprPath(last, n)
					if n.Path.Depth == 0 { // uverse
						cv := evalConst(last, n)
						// built-in functions must be called.
						if !cv.IsUndefined() &&
							cv.T.Kind() == FuncKind &&
							ftype != TRANS_CALL_FUNC {
							panic(fmt.Sprintf(
								"use of builtin %s not in function call",
								n.Name))
						}
						return cv, TRANS_CONTINUE
					}
				}

			// TRANS_LEAVE -----------------------
			case *BasicLitExpr:
				// Replace with *constExpr.
				cv := evalConst(last, n)
				return cv, TRANS_CONTINUE

			// TRANS_LEAVE -----------------------
			case *BinaryExpr:
				// Replace with *constExpr if const operands.
				isShift := n.Op == SHL || n.Op == SHR
				rt := evalTypeOf(last, n.Right)
				// Special (recursive) case if shift and right isn't uint.
				if isShift && baseOf(rt) != UintType {
					// convert n.Right to (gno) uint type,
					rn := Expr(Call("uint", n.Right))
					// reset/create n2 to preprocess right child.
					n2 := &BinaryExpr{
						Left:  n.Left,
						Op:    n.Op,
						Right: rn,
					}
					resn := Preprocess(imp, last, n2)
					return resn, TRANS_CONTINUE
				}
				// General case.
				lcx, lic := n.Left.(*constExpr)
				rcx, ric := n.Right.(*constExpr)
				if lic {
					if ric {
						cv := evalConst(last, n)
						return cv, TRANS_CONTINUE
					} else if isUntyped(lcx.T) {
						if rnt, ok := rt.(*nativeType); ok {
							if isShift {
								panic("should not happen")
							}
							// get concrete native base type.
							pt := go2GnoBaseType(rnt.Type).(PrimitiveType)
							// convert n.Left to pt type,
							convertIfConst(last, n.Left, pt)
							// convert n.Right to (gno) pt type,
							rn := Expr(Call(pt.String(), n.Right))
							// and convert result back.
							tx := &constTypeExpr{
								Source: n,
								Type:   rnt,
							}
							// reset/create n2 to preprocess right child.
							n2 := &BinaryExpr{
								Left:  n.Left,
								Op:    n.Op,
								Right: rn,
							}
							resn := Node(Call(tx, n2))
							resn = Preprocess(imp, last, resn)
							return resn, TRANS_CONTINUE
							// NOTE: binary operations are always computed in
							// gno, never with reflect.
						} else {
							if isShift {
								// nothing to do, right type is (already) uint type.
							} else {
								// convert n.Left to right type.
								convertIfConst(last, n.Left, rt)
							}
						}
					}
				} else {
					if ric && isUntyped(rcx.T) {
						if isShift {
							if baseOf(rt) != UintType {
								// convert n.Right to (gno) uint type.
								convertIfConst(last, n.Right, UintType)
							} else {
								// leave n.Left as is and baseOf(n.Right) as UintType.
							}
						} else {
							lt := evalTypeOf(last, n.Left)
							if lnt, ok := lt.(*nativeType); ok {
								// get concrete native base type.
								pt := go2GnoBaseType(lnt.Type).(PrimitiveType)
								// convert n.Left to (gno) pt type,
								ln := Expr(Call(pt.String(), n.Left))
								// convert n.Right to pt type,
								convertIfConst(last, n.Right, pt)
								// and convert result back.
								tx := &constTypeExpr{
									Source: n,
									Type:   lnt,
								}
								// reset/create n2 to preprocess left child.
								n2 := &BinaryExpr{
									Left:  ln,
									Op:    n.Op,
									Right: n.Right,
								}
								resn := Node(Call(tx, n2))
								resn = Preprocess(imp, last, resn)
								return resn, TRANS_CONTINUE
								// NOTE: binary operations are always computed in
								// gno, never with reflect.
							} else {
								// convert n.Right to left type.
								convertIfConst(last, n.Right, lt)
							}
						}
					} else {
						lt := evalTypeOf(last, n.Left)
						if debug {
							if !isShift {
								assertTypes(lt, rt)
							}
						}
						if lnt, ok := lt.(*nativeType); ok {
							// get concrete native base type.
							pt := go2GnoBaseType(lnt.Type).(PrimitiveType)
							// convert n.Left to (gno) pt type,
							ln := Expr(Call(pt.String(), n.Left))
							// convert n.Right to pt or uint type,
							rn := n.Right
							if isShift {
								if baseOf(rt) != UintType {
									rn = Expr(Call("uint", n.Right))
								}
							} else {
								rn = Expr(Call(pt.String(), n.Right))
							}
							// and convert result back.
							tx := &constTypeExpr{
								Source: n,
								Type:   lnt,
							}
							// reset/create n2 to preprocess children.
							n2 := &BinaryExpr{
								Left:  ln,
								Op:    n.Op,
								Right: rn,
							}
							resn := Node(Call(tx, n2))
							resn = Preprocess(imp, last, resn)
							return resn, TRANS_CONTINUE
							// NOTE: binary operations are always computed in
							// gno, never with reflect.
						} else {
							// nothing to do.
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *CallExpr:
				// Special case conversions.
				if cx, ok := n.Func.(*constExpr); ok {
					if nx, ok := cx.Source.(*NameExpr); ok &&
						nx.Name == "append" {
						st := evalTypeOf(last, n.Args[0])
						if debug {
							if st.Kind() != SliceKind {
								panic(fmt.Sprintf(
									"cannot append to non-slice kind %s",
									st.Kind()))
							}
						}
						// Replace const Args with *constExpr.
						set := st.Elem()
						for i := 1; i < len(n.Args); i++ {
							arg := n.Args[i]
							if n.Varg && i == len(n.Args)-1 {
								convertIfConst(last, arg, nil)
							} else {
								convertIfConst(last, arg, set)
							}
						}
						break // done with "append" special case.
					}
				}
				// Func type evaluation.
				var ft *FuncType
				ift := evalTypeOf(last, n.Func)
				switch cft := baseOf(ift).(type) {
				case *FuncType:
					ft = cft
				case *nativeType:
					ft = go2GnoFuncType(cft.Type)
				case *TypeType:
					if len(n.Args) != 1 {
						panic("type conversion requires single argument")
					}
					convertIfConst(last, n.Args[0], nil)
					return n, TRANS_CONTINUE
				default:
					panic(fmt.Sprintf(
						"unexpected func type %v (%v)",
						ift, reflect.TypeOf(ift)))
				}
				// Replace const Args with *constExpr.
				hasVarg := ft.HasVarg()
				isVargX := n.Varg
				for i, arg := range n.Args {
					if hasVarg && (len(ft.Params)-1) <= i {
						if isVargX {
							if len(ft.Params) <= i {
								panic("expected final vargs slice but got many")
							}
							convertIfConst(last, arg,
								ft.Params[i].Type)
						} else {
							convertIfConst(last, arg,
								ft.Params[len(ft.Params)-1].Type.Elem())
						}
					} else {
						convertIfConst(last, arg, ft.Params[i].Type)
					}
				}
				// TODO in the future, pure results

			// TRANS_LEAVE -----------------------
			case *IndexExpr:
				xt := evalTypeOf(last, n.X)
				switch xt.Kind() {
				case ArrayKind:
					convertIfConst(last, n.Index, IntType)
				case SliceKind:
					convertIfConst(last, n.Index, IntType)
				case MapKind:
					mt := baseOf(gnoTypeOf(xt)).(*MapType)
					convertIfConst(last, n.Index, mt.Key)
				default:
					panic(fmt.Sprintf(
						"unexpected index base kind for type %s",
						xt.String()))

				}

			// TRANS_LEAVE -----------------------
			case *SliceExpr:
				// Replace const L/H/M with int *constExpr.
				convertIfConst(last, n.Low, IntType)
				convertIfConst(last, n.High, IntType)
				convertIfConst(last, n.Max, IntType)

			// TRANS_LEAVE -----------------------
			case *TypeAssertExpr:
				n.Type = evalConst(last, n.Type)
				if ftype == TRANS_ASSIGN_RHS {
					as := ns[len(ns)-1].(*AssignStmt)
					if len(as.Lhs) == 1 {
						n.HasOK = false
					} else if len(as.Lhs) == 2 {
						n.HasOK = true
					} else {
						panic(fmt.Sprintf(
							"type assert assignment takes 1 or 2 lhs operands, got %v",
							len(as.Lhs),
						))
					}
				}

			// TRANS_LEAVE -----------------------
			case *UnaryExpr:
				xt := evalTypeOf(last, n.X)
				if xnt, ok := xt.(*nativeType); ok {
					// get concrete native base type.
					pt := go2GnoBaseType(xnt.Type).(PrimitiveType)
					// convert n.X to gno type,
					xn := Expr(Call(pt.String(), n.X))
					// and convert result back.
					tx := &constTypeExpr{
						Source: n,
						Type:   xnt,
					}
					// reset/create n2 to preprocess children.
					n2 := &UnaryExpr{
						X:  xn,
						Op: n.Op,
					}
					resn := Node(Call(tx, n2))
					resn = Preprocess(imp, last, resn)
					return resn, TRANS_CONTINUE
					// NOTE: like binary operations, unary operations are
					// always computed in gno, never with reflect.
				}
				// Replace with *constExpr if const X.
				if isConst(n.X) {
					cv := evalConst(last, n)
					return cv, TRANS_CONTINUE
				}

			// TRANS_LEAVE -----------------------
			case *CompositeLitExpr:
				// Get or evaluate composite type.
				clt := evalType(last, n.Type)
				// Replace const Elts with default *constExpr.
			CLT_TYPE_SWITCH:
				switch cclt := baseOf(clt).(type) {
				case *StructType:
					for i := 0; i < len(n.Elts); i++ {
						flat := cclt.Mapping[i]
						ft := cclt.Fields[flat].Type
						convertIfConst(last, n.Elts[i].Value, ft)
					}
				case *ArrayType:
					for i := 0; i < len(n.Elts); i++ {
						convertIfConst(last, n.Elts[i].Key, IntType)
						convertIfConst(last, n.Elts[i].Value, cclt.Elt)
					}
				case *SliceType:
					for i := 0; i < len(n.Elts); i++ {
						convertIfConst(last, n.Elts[i].Key, IntType)
						convertIfConst(last, n.Elts[i].Value, cclt.Elt)
					}
				case *MapType:
					for i := 0; i < len(n.Elts); i++ {
						convertIfConst(last, n.Elts[i].Key, cclt.Key)
						convertIfConst(last, n.Elts[i].Value, cclt.Value)
					}
				case *nativeType:
					clt = cclt.GnoType()
					goto CLT_TYPE_SWITCH
				default:
					panic(fmt.Sprintf(
						"unexpected composite type %s",
						clt.String()))
				}
				// If variadic array lit, measure.
				if at, ok := clt.(*ArrayType); ok {
					if at.Vrd {
						idx := 0
						for _, elt := range n.Elts {
							if elt.Key == nil {
								idx++
							} else {
								// XXX why convert?
								k := evalConst(last, elt.Key).ConvertGetInt()
								if idx <= k {
									idx = k + 1
								} else {
									panic("array lit key out of order")
								}
							}
						}
						// update type
						// (dontcare)
						// at.Vrd = false
						at.Len = idx
						// update node
						cx := constInt(n, idx)
						n.Type.(*ArrayTypeExpr).Len = cx
					}
				}

			// TRANS_LEAVE -----------------------
			case *KeyValueExpr:
				// NOTE: For simplicity we just
				// use the *CompositeLitExpr.

			// TRANS_LEAVE -----------------------
			case *SelectorExpr:
				xt := evalTypeOf(last, n.X)
				if pt, ok := xt.(PointerType); ok {
					if dt, ok := pt.Elt.(*DeclaredType); ok {
						mthd := dt.GetMethod(n.Sel)
						if mthd == nil {
							// Go spec: "if the type of x is a
							// defined pointer type and (*x).f is a
							// valid selector expression denoting a
							// field (but not a method), x.f is
							// shorthand for (*x).f."
						} else {
							if _, ok := mthd.Type.Params[0].Type.(PointerType); ok {
								xt = xt.Elem()
								goto SEL_TYPE_SWITCH
							} else {
								// Go spec: "As with selectors,
								// a reference to a
								// non-interface method with a
								// value receiver using a
								// pointer will automatically
								// dereference that pointer:
								// pt.Mv is equivalent to
								// (*pt).Mv."
							}
						}
					}
					// convert to (*x).f.
					n.X = &StarExpr{X: n.X}
					n.X.SetAttribute(ATTR_PREPROCESSED, true)
					xt = xt.Elem()
				} else if dt, ok := xt.(*DeclaredType); ok {
					mthd := dt.GetMethod(n.Sel)
					if mthd != nil {
						if _, ok := mthd.Type.Params[0].Type.(PointerType); ok {
							// Go spec: "If x is addressable
							// and &x's method set contains
							// m, x.m() is shorthand for
							// (&x).m()"
							// Go spec: "As with method
							// calls, a reference to a
							// non-interface method with a
							// pointer receiver using an
							// addressable value will
							// automatically take the
							// address of that value: t.Mp
							// is equivalent to (&t).Mp."
						} else {
							goto SEL_TYPE_SWITCH
						}
						// convert to (&x).m.
						n.X = &RefExpr{X: n.X}
						n.X.SetAttribute(ATTR_PREPROCESSED, true)
					}
				}
			SEL_TYPE_SWITCH:
				// Set selector path.
				switch xt := xt.(type) {
				case *DeclaredType:
					// bound method or underlying.
					// TODO check for unexported fields.
					n.Path = xt.GetPathForName(n.Sel)
				case *StructType:
					// struct field
					// TODO check for unexported fields.
					n.Path = xt.GetPathForName(n.Sel)
				case *PackageType:
					// packages can only be referred to by
					// *NameExprs, and cannot be copied.
					nx := n.X.(*NameExpr)
					pv := last.GetValueRef(nx.Name)
					pn := pv.V.(*PackageValue).Source
					n.Path = pn.GetPathForName(n.Sel)
				case *InterfaceType:
					// first implement interfaaces
					// TODO check for unexported fields.
					panic("not yet implemented")
				case *TypeType:
					// unbound method
					xv := evalType(last, n.X)
					switch xv := xv.(type) {
					case *DeclaredType:
						n.Path = xv.GetPathForName(n.Sel)
						if n.Path.Depth > 1 {
							panic(fmt.Sprintf(
								"DeclaredType has no method %s",
								n.Sel))
						}
					default:
						panic(fmt.Sprintf(
							"unexpected selector expression type value %s",
							xv.String()))
					}
				case *nativeType:
					// native types don't use path indices.
					n.Path = NewValuePath(n.Sel, 0, 0)
				default:
					panic(fmt.Sprintf(
						"unexpected selector expression type %s",
						xt.String()))
				}

			// TRANS_LEAVE -----------------------
			case *FieldTypeExpr:
				// Replace const Tag with default *constExpr.
				convertIfConst(last, n.Tag, nil)

			// TRANS_LEAVE -----------------------
			case *ArrayTypeExpr:
				if n.Len == nil {
					// Calculate length at *CompositeLitExpr:LEAVE
				} else {
					// Replace const Len with int *constExpr.
					evalConst(last, n.Len)
					convertIfConst(last, n.Len, IntType)
				}
				// TODO *constTypeExpr?
				evalType(last, n)

			// TRANS_LEAVE -----------------------
			case *SliceTypeExpr:
				// TODO *constTypeExpr?
				evalType(last, n)

			// TRANS_LEAVE -----------------------
			case *InterfaceTypeExpr:
				// TODO *constTypeExpr?
				evalType(last, n)

			// TRANS_LEAVE -----------------------
			case *ChanTypeExpr:
				// TODO *constTypeExpr?
				evalType(last, n)

			// TRANS_LEAVE -----------------------
			case *FuncTypeExpr:
				// TODO *constTypeExpr?
				evalType(last, n)

			// TRANS_LEAVE -----------------------
			case *MapTypeExpr:
				// TODO *constTypeExpr?
				evalType(last, n)

			// TRANS_LEAVE -----------------------
			case *StructTypeExpr:
				// TODO *constTypeExpr?
				evalType(last, n)

			// TRANS_LEAVE -----------------------
			case *AssignStmt:
				// Rhs consts become default *constExprs.
				for _, rx := range n.Rhs {
					// NOTE: does nothing if rx is "nil".
					convertIfConst(last, rx, nil)
				}
				// Handle any definitions/assignments.
				if n.Op == DEFINE {
					if len(n.Lhs) > len(n.Rhs) {
						// Unpack n.Rhs[0] to n.Lhs[:]
						if len(n.Rhs) != 1 {
							panic("should not happen")
						}
						cx, ok := n.Rhs[0].(*CallExpr)
						if !ok {
							panic("should not happen")
						}
						ft := gnoTypeOf(evalTypeOf(last, cx.Func)).(*FuncType)
						if len(n.Lhs) != len(ft.Results) {
							panic(fmt.Sprintf(
								"assignment mismatch: "+
									"%d variables but %s returns %d values",
								len(n.Lhs), cx.Func.String(), len(ft.Results)))
						}
						for i, lx := range n.Lhs {
							ln := lx.(*NameExpr).Name
							rt := ft.Results[i]
							// re-definition
							last.Define(ln, anyValue(rt))
						}
					} else {
						for i, lx := range n.Lhs {
							ln := lx.(*NameExpr).Name
							rx := n.Rhs[i]
							rt := evalTypeOf(last, rx)
							// re-definition
							last.Define(ln, anyValue(rt))
						}
					}
				} else {
					if len(n.Lhs) > len(n.Rhs) {
						// TODO dry code w/ above.
						// Unpack n.Rhs[0] to n.Lhs[:]
						if len(n.Rhs) != 1 {
							panic("should not happen")
						}
						cx, ok := n.Rhs[0].(*CallExpr)
						if !ok {
							panic("should not happen")
						}
						ft := gnoTypeOf(evalTypeOf(last, cx.Func)).(*FuncType)
						if len(n.Lhs) != len(ft.Results) {
							panic(fmt.Sprintf(
								"assignment mismatch: "+
									"%d variables but %s returns %d values",
								len(n.Lhs), cx.Func.String(), len(ft.Results)))
						}
						// No conversion to do.
					} else {
						for i, lx := range n.Lhs {
							lt := evalTypeOf(last, lx)
							rx := n.Rhs[i]
							// converts if rx is "nil".
							convertIfConst(last, rx, lt)
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *ForStmt:
				// Cond consts become bool *constExprs.
				convertIfConst(last, n.Cond, BoolType)

			// TRANS_LEAVE -----------------------
			case *IfStmt:
				// Cond consts become bool *constExprs.
				convertIfConst(last, n.Cond, BoolType)

			// TRANS_LEAVE -----------------------
			case *RangeStmt:
				// key value if define.
				if n.Op == DEFINE {
					if n.Key != nil {
						// already defined @TRANS_BLOCK.
					}
					if n.Value != nil {
						vn := n.Value.(*NameExpr).Name
						pb := last.GetParent()
						xt := evalTypeOf(pb, n.X)
						et := xt.Elem()
						// re-definition.
						last.Define(vn, anyValue(et))
					}
				}

			// TRANS_LEAVE -----------------------
			case *ReturnStmt:
				fnode, ft := funcNodeOf(last)
				// Results consts become default *constExprs.
				for i, rx := range n.Results {
					rtx := ft.Results[i].Type
					rt := evalType(fnode.GetParent(), rtx)
					convertIfConst(last, rx, rt)
				}

			// TRANS_LEAVE -----------------------
			case *SendStmt:
				// Value consts become default *constExprs.
				convertIfConst(last, n.Value, nil)

			// TRANS_LEAVE -----------------------
			case *SelectCaseStmt:
				// maybe receive defines.
				// if as, ok := n.Comm.(*AssignStmt); ok {
				//     handled by case *AssignStmt.
				// }

			// TRANS_LEAVE -----------------------
			case *ValueDecl:
				// evaluate value if const expr.
				if n.Const {
					// not necessarily a *constExpr (yet).
					n.Value = evalConst(last, n.Value)
				} else {
					// value may already be *constExpr, but
					// otherwise as far as we know the
					// expression is not a const expr, so no
					// point evaluating it further.  this makes
					// the implementation differ from
					// runDeclaration(), as this uses OpTypeOf.
				}
				// convert and evaluate type.
				var t Type
				if n.Type != nil {
					t = evalType(last, n.Type)
					convertIfConst(last, n.Value, t)
				} else {
					convertIfConst(last, n.Value, nil)
					t = evalTypeOf(last, n.Value)
				}
				// evaluate typed value for static definition.
				var tv TypedValue
				if cx, ok := n.Value.(*constExpr); ok {
					// if value is const expr; const and var decls.
					tv = cx.TypedValue
				} else {
					// for var decls of non-const expr.
					tv = anyValue(t)
				}
				// define.
				if fn, ok := last.(*FileNode); ok {
					pn := fn.GetParent().(*PackageNode)
					pn.Define(n.Name, tv)
				} else {
					last.Define(n.Name, tv)
				}
				n.Path = last.GetPathForName(n.Name)
				// TODO make note of constance in static block for future
				// use, or consider "const paths".
				// set as preprocessed.

			// TRANS_LEAVE -----------------------
			case *TypeDecl:
				// Construct new Type, where any recursive references
				// refer to the old Type declared during
				// *TypeDecl:ENTER.  Then, copy over the values,
				// completing the recursion.
				temp := evalType(last, n.Type)
				tipe := last.GetValueRef(n.Name).GetType()
				switch oldt := tipe.(type) {
				case *FuncType:
					*oldt = *(temp.(*FuncType))
				case *ArrayType:
					*oldt = *(temp.(*ArrayType))
				case *SliceType:
					*oldt = *(temp.(*SliceType))
				case *InterfaceType:
					*oldt = *(temp.(*InterfaceType))
				case *ChanType:
					*oldt = *(temp.(*ChanType))
				case *MapType:
					*oldt = *(temp.(*MapType))
				case *StructType:
					*oldt = *(temp.(*StructType))
				case *DeclaredType:
					pn := packageOf(last)
					// XXX this is wrong,
					// this makes static type and runtime
					// type be different somehow.
					// this makes a differnt one
					dt := declareWith(pn.PkgPath, n.Name, temp)
					*oldt = *dt
				default:
					panic(fmt.Sprintf("unexpected type declaration type %v",
						reflect.TypeOf(tipe)))
				}
				// We need to replace all references of the new
				// Type with old Type, including in attributes.
				n.Type.SetAttribute(ATTR_TYPE_VALUE, tipe)
				// Replace the type with *constTypeExpr{},
				// otherwise methods would be un at runtime.
				n.Type = constType(n.Type, tipe)
			}
			// end type switch statement

			// TRANS_LEAVE -----------------------
			// finalization.
			if _, ok := n.(BlockNode); ok {
				// Pop block.
				stack = stack[:len(stack)-1]
				last = stack[len(stack)-1]
				return n, TRANS_CONTINUE
			} else {
				return n, TRANS_CONTINUE
			}
		}

		panic(fmt.Sprintf(
			"unknown stage %v", stage))
	})

	return nn
}

func pushBlock(bn BlockNode, last *BlockNode, stack *[]BlockNode) {
	bn.InitStaticBlock(bn, *last)
	*last = bn
	*stack = append(*stack, bn)
}

// Evaluates the value of x which is expected to be a typeval.
// Caches the result as an attribute of x.
// To discourage mis-use, expects x to already be
// preprocessed.
func evalType(last BlockNode, x Expr) Type {
	if t, ok := x.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
		return t
	} else if ctx, ok := x.(*constTypeExpr); ok {
		return ctx.Type // no need to set attribute.
	}
	pn := packageOf(last)
	tv := NewMachine(pn.PkgPath).StaticEval(last, x)
	t := tv.GetType()
	x.SetAttribute(ATTR_TYPE_VALUE, t)
	return t
}

// If t is a native type, returns the gno type.
func gnoTypeOf(t Type) Type {
	if nt, ok := t.(*nativeType); ok {
		return nt.GnoType()
	} else {
		return t
	}
}

// If it is known that the type was already evaluated,
// use this function instead of evalType().
// TODO not used.
func getType(x Expr) Type {
	if t, ok := x.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
		return t
	} else {
		panic(fmt.Sprintf(
			"getType() called on expr not yet evaluated with evalType(): %s",
			x.String(),
		))
	}
}

// Unlike evalType, x is not expected to be a typeval,
// but rather computes the type OF x.
func evalTypeOf(last BlockNode, x Expr) Type {
	if t, ok := x.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		return t
	} else if _, ok := x.(*constTypeExpr); ok {
		return gTypeType
	} else if ctx, ok := x.(*constExpr); ok {
		return ctx.T
	} else {
		pn := packageOf(last)
		t = NewMachine(pn.PkgPath).StaticEvalTypeOf(last, x)
		x.SetAttribute(ATTR_TYPEOF_VALUE, t)
		return t
	}
}

// Evaluate constant expressions.  Assumes all operands
// are already defined.  No type conversion is done by
// the machine except as required by the expression (but
// otherwise the context is not considered).  For
// example, untyped bigint types remain as untyped bigint
// types after evaluation.  Conversion happens in a
// separate step while leaving composite exprs/nodes that
// contain constant expression nodes (e.g. const exprs in
// the rhs of AssignStmts).
func evalConst(last BlockNode, x Expr) *constExpr {
	// TODO: some check or verification for ensuring x
	// is constant?  From the machine?
	pn := packageOf(last)
	cv := NewMachine(pn.PkgPath).StaticEval(last, x)
	cx := &constExpr{
		Source:     x,
		TypedValue: cv,
	}
	cx.SetAttribute(ATTR_PREPROCESSED, true)
	return cx
}

func packageOf(last BlockNode) *PackageNode {
	for {
		if pn, ok := last.(*PackageNode); ok {
			return pn
		}
		last = last.GetParent()
	}
}

func funcNodeOf(last BlockNode) (BlockNode, *FuncTypeExpr) {
	for {
		if flx, ok := last.(*FuncLitExpr); ok {
			return flx, &flx.Type
		} else if fd, ok := last.(*FuncDecl); ok {
			return fd, &fd.Type
		}
		last = last.GetParent()
	}
}

func lastDecl(ns []Node) Decl {
	for i := len(ns) - 1; 0 <= i; i-- {
		if d, ok := ns[i].(Decl); ok {
			return d
		}
	}
	return nil
}

func asValue(t Type) TypedValue {
	return TypedValue{
		T: gTypeType,
		V: TypeValue{t},
	}
}

func anyValue(t Type) TypedValue {
	return TypedValue{
		T: t,
		V: nil,
	}
}

func isConst(x Expr) bool {
	_, ok := x.(*constExpr)
	return ok
}

func convertIfConst(last BlockNode, x Expr, t Type) {
	if x == nil {
		return
	}
	if t != nil && t.Kind() == InterfaceKind {
		// TODO type check?
		return
	}
	if cx, ok := x.(*constExpr); ok {
		if isUntyped(cx.T) {
			ConvertUntypedTo(&cx.TypedValue, t)
		} else if t != nil {
			ConvertTo(&cx.TypedValue, t)
		}
	}
}

// Returns any names not yet defined in expr.
// These happen upon enter from the top, so value paths cannot be used.
// If no names are un and x is TypeExpr, evalType(last, x) must not
// panic.
// NOTE: has no side effects except for the case of composite type expressions,
// which must get preprocessed for inner composite type eliding to work.
func findUndefined(imp Importer, last BlockNode, x Expr) (un Name) {
	return findUndefined2(imp, last, x, nil)
}

func findUndefined2(imp Importer, last BlockNode, x Expr, t Type) (un Name) {
	if x == nil {
		return
	}
	switch cx := x.(type) {
	case *NameExpr:
		if _, ok := UverseNode().GetLocalIndex(cx.Name); ok {
			return
		}
		if tv := last.GetValueRef(cx.Name); tv != nil {
			return
		}
		return cx.Name
	case *BasicLitExpr:
		return
	case *BinaryExpr:
		un = findUndefined(imp, last, cx.Left)
		if un != "" {
			return
		}
		un = findUndefined(imp, last, cx.Right)
		if un != "" {
			return
		}
	case *SelectorExpr:
		return findUndefined(imp, last, cx.X)
	case *StarExpr:
		return findUndefined(imp, last, cx.X)
	case *RefExpr:
		return findUndefined(imp, last, cx.X)
	case *TypeAssertExpr:
		un = findUndefined(imp, last, cx.X)
		if un != "" {
			return
		}
		return findUndefined(imp, last, cx.Type)
	case *UnaryExpr:
		return findUndefined(imp, last, cx.X)
	case *CompositeLitExpr:
		var ct Type
		if cx.Type == nil {
			if t == nil {
				panic("cannot elide unknown composite type")
			}
			ct = t
		} else {
			un = findUndefined(imp, last, cx.Type)
			if un != "" {
				return
			}
			// preprocess now for eliding purposes.
			// TODO recursive preprocessing here is hacky, find a better way.
			// This cannot be done asynchronously, cuz undefined names ought to
			// be returned immediately to let the caller predefine it.
			cx.Type = Preprocess(imp, last, cx.Type).(Expr) // recursive
			ct = evalType(last, cx.Type)
			// elide composite lit element (nested) composite types.
			elideCompositeElements(cx, ct)
		}
		switch ct.Kind() {
		case ArrayKind, SliceKind, MapKind:
			for _, kvx := range cx.Elts {
				un = findUndefined(imp, last, kvx.Key)
				if un != "" {
					return
				}
				un = findUndefined2(imp, last, kvx.Value, ct.Elem())
				if un != "" {
					return
				}
			}
		case StructKind:
			for _, kvx := range cx.Elts {
				un = findUndefined(imp, last, kvx.Value)
				if un != "" {
					return
				}
			}
		default:
			panic(fmt.Sprintf(
				"unexpected composite lit type %s",
				ct.String()))
		}
	case *FuncLitExpr:
		return findUndefined(imp, last, &cx.Type)
	case *FieldTypeExpr:
		return findUndefined(imp, last, cx.Type)
	case *ArrayTypeExpr:
		if cx.Len != nil {
			un = findUndefined(imp, last, cx.Len)
			if un != "" {
				return
			}
		}
		return findUndefined(imp, last, cx.Elt)
	case *SliceTypeExpr:
		return findUndefined(imp, last, cx.Elt)
	case *InterfaceTypeExpr:
		for i := range cx.Methods {
			un = findUndefined(imp, last, &cx.Methods[i])
			if un != "" {
				return
			}
		}
	case *ChanTypeExpr:
		return findUndefined(imp, last, cx.Value)
	case *FuncTypeExpr:
		for i := range cx.Params {
			un = findUndefined(imp, last, &cx.Params[i])
			if un != "" {
				return
			}
		}
		for i := range cx.Results {
			un = findUndefined(imp, last, &cx.Results[i])
			if un != "" {
				return
			}
		}
	case *MapTypeExpr:
		un = findUndefined(imp, last, cx.Key)
		if un != "" {
			return
		}
		un = findUndefined(imp, last, cx.Value)
		if un != "" {
			return
		}
	case *StructTypeExpr:
		for i := range cx.Fields {
			un = findUndefined(imp, last, &cx.Fields[i])
			if un != "" {
				return
			}
		}
	case *constTypeExpr:
		return
	case *constExpr:
		return
	default:
		panic(fmt.Sprintf(
			"unexpected expr: %v (%v)",
			x, reflect.TypeOf(x)))
	}
	return
}

// The purpose of this function is to split declarations into two parts; the
// first part creates empty placeholder type instances, and the second part
// to fill in the details while supporting recursive and cyclic definitions.
// preedefineNow handles the first part while the Preprocessor handles the
// rest.
// The exception to this separation is for *ValueDecls, which get
// preprocessed immediately, recursively, and also the function signature of
// any function or method declarations.
func predefineNow(imp Importer, last BlockNode, d Decl) Decl {
	pkg := packageOf(last)
	// recursively predefine dependencies.
	for {
		un := tryPredefine(imp, last, d)
		if un != "" {
			// look up dependency declaration from fileset.
			file, decl := pkg.FileSet.GetDeclFor(un)
			// predefine dependency (recursive).
			predefineNow(imp, file, *decl)
			// if value decl, preprocess now (greatly recursive).
			if vd, ok := (*decl).(*ValueDecl); ok {
				*decl = Preprocess(imp, file, vd).(Decl)
			}
		} else {
			break
		}
	}
	if fd, ok := d.(*FuncDecl); ok {
		// *FuncValue/*FuncType is mostly empty still; here we just fill the
		// func type.
		// NOTE: unlike the *ValueDecl case, this case doesn't preprocess d
		// itself (only d.Type).
		if !fd.IsMethod {
			ftv := pkg.GetValueRef(fd.Name)
			ft := ftv.T.(*FuncType)
			fd.Type = *Preprocess(imp, last, &fd.Type).(*FuncTypeExpr)
			ft2 := evalType(last, &fd.Type).(*FuncType)
			*ft = *ft2
			// XXX replace attr w/ ft?
		}
	}
	return d
}

// If a dependent name is not yet defined, that name is returned; this return
// value is used by the caller to enforce declaration order.  If a dependent
// type is not yet defined (preprocessed), that type is fully preprocessed.
// Besides defining the type (and immediate dependent types of d) onto last
// (or packageOf(last)), there are no other side effects.  This function works
// for all block nodes and must be called for name declarations within
// (non-file, non-package) stmt bodies.
func tryPredefine(imp Importer, last BlockNode, d Decl) (un Name) {
	if d.GetAttribute(ATTR_PREDEFINED) == true {
		panic("decl node already predefined!")
	}

	// If un is blank, it means the predefine succeeded.
	defer func() {
		if un == "" {
			d.SetAttribute(ATTR_PREDEFINED, true)
		}
	}()

	// NOTE: These happen upon enter from the top,
	// so value paths cannot be used here.
	switch d := d.(type) {
	case *ImportDecl:
		pv := imp(d.PkgPath)
		if pv == nil {
			panic(fmt.Sprintf(
				"unknown import path %s",
				d.Path))
		}
		if d.Name == "" { // use default
			d.Name = pv.PkgName
		}
		// NOTE imports usually must happen with a file,
		// and so last is usually a *FileNode, but for
		// testing convenience we allow importing
		// directly onto the package.
		last.Define(d.Name, TypedValue{
			T: gPackageType,
			V: pv,
		})
		d.Path = last.GetPathForName(d.Name)
	case *ValueDecl:
		un = findUndefined(imp, last, d.Type)
		if un != "" {
			return
		}
		un = findUndefined(imp, last, d.Value)
		if un != "" {
			return
		}
		last2 := skipFile(last)
		last2.Define(d.Name, anyValue(nil))
		d.Path = last.GetPathForName(d.Name)
	case *TypeDecl:
		// before looking for dependencies, predefine empty type.
		last2 := skipFile(last)
		_, ok := last2.GetLocalIndex(d.Name)
		if !ok {
			// construct empty t type
			var t Type
			switch tx := d.Type.(type) {
			case *FuncTypeExpr:
				t = &FuncType{}
			case *ArrayTypeExpr:
				t = &ArrayType{}
			case *SliceTypeExpr:
				t = &SliceType{}
			case *InterfaceTypeExpr:
				t = &InterfaceType{}
			case *ChanTypeExpr:
				t = &ChanType{}
			case *MapTypeExpr:
				t = &MapType{}
			case *StructTypeExpr:
				t = &StructType{}
			case *NameExpr:
				if idx, ok := UverseNode().GetLocalIndex(tx.Name); ok {
					// uverse name
					tv := Uverse().GetValueRefAt(NewValuePath(tx.Name, 0, idx))
					t = tv.GetType()
				} else if tv := last.GetValueRef(tx.Name); tv != nil {
					// block name
					t = tv.GetType()
				} else {
					// yet undefined
					un = tx.Name
					return
				}
			default:
				panic(fmt.Sprintf(
					"unexpected type declaration type %v",
					reflect.TypeOf(d.Type)))
			}
			if d.IsAlias {
				// use t directly.
			} else {
				// create new declared type.
				pn := packageOf(last)
				t = declareWith(pn.PkgPath, d.Name, t)
			}
			// fill in later.
			last2.Define(d.Name, asValue(t))
			d.Path = last.GetPathForName(d.Name)
		}
		// after predefinitions, return any undefined dependencies.
		un = findUndefined(imp, last, d.Type)
		if un != "" {
			return
		}
	case *FuncDecl:
		un = findUndefined(imp, last, &d.Type)
		if un != "" {
			return
		}
		// define function name.
		if d.IsMethod {
			// methods are defined as struct fields, not
			// in the last block.  receiver isn't
			// processed until FuncDecl:BLOCK.
			un = findUndefined(imp, last, &d.Recv)
			if un != "" {
				return
			}
		} else {
			// define a FuncValue w/ above type as d.Name.
			// fill in later during *FuncDecl:BLOCK.
			var ft = &FuncType{}
			pkg := skipFile(last).(*PackageNode)
			pkg.Define(d.Name, TypedValue{
				T: ft,
				V: &FuncValue{
					Type:       ft,
					IsMethod:   false,
					Source:     d,
					Name:       d.Name,
					Body:       d.Body,
					Closure:    nil, // set later.
					NativeBody: nil,
					FileName:   filenameOf(last),
					pkg:        nil, // set later.
				},
			})
			d.Path = last.GetPathForName(d.Name)
		}
	default:
		panic(fmt.Sprintf(
			"unexpected declaration type %v",
			d.String()))
	}
	return ""
}

func constInt(source Expr, i int) *constExpr {
	cx := &constExpr{Source: source}
	cx.T = IntType
	cx.SetInt(i)
	cx.SetAttribute(ATTR_PREPROCESSED, true)
	return cx
}

func constUntypedBigint(source Expr, i64 int64) *constExpr {
	cx := &constExpr{Source: source}
	cx.T = UntypedBigintType
	cx.V = BigintValue{big.NewInt(i64)}
	cx.SetAttribute(ATTR_PREPROCESSED, true)
	return cx
}

func constType(source Expr, t Type) *constTypeExpr {
	cx := &constTypeExpr{Source: source}
	cx.Type = t
	cx.SetAttribute(ATTR_PREPROCESSED, true)
	return cx
}

func fillNameExprPath(last BlockNode, nx *NameExpr) {
	if nx.Name == "_" {
		// Blank name has no path; caller error.
		panic("should not happen")
	}
	// check to see if name is global.
	if idx, ok := UverseNode().GetLocalIndex(nx.Name); ok {
		// Keep generation as 0 (instead of 1),
		// as 1 is for last, and 0 is for uverse.
		nx.Path.Depth = 0
		nx.Path.Index = idx
	} else {
		// Set path for name.
		nx.Path = last.GetPathForName(nx.Name)
	}
}

func skipFile(n BlockNode) BlockNode {
	if fn, ok := n.(*FileNode); ok {
		return packageOf(fn)
	} else {
		return n
	}
}

// If n is a *FileNode, return name, otherwise empty.
func filenameOf(n BlockNode) Name {
	if fnode, ok := n.(*FileNode); ok {
		return fnode.Name
	} else {
		return ""
	}
}

func elideCompositeElements(clx *CompositeLitExpr, clt Type) {
	switch clt := baseOf(clt).(type) {
	/*
		case PointerType:
			det := clt.Elt.Elt
			for _, ex := range clx.Elts {
				vx := evx.Value
				if vclx, ok := vx.(*CompositeLitExpr); ok {
					if vclx.Type == nil {
						vclx.Type = &constTypeExpr{
							Source: vx,
							Type:   et,
						}
					}
				}
			}
	*/
	case *ArrayType:
		et := clt.Elt
		el := len(clx.Elts)
		for i := 0; i < el; i++ {
			kvx := &clx.Elts[i]
			elideCompositeExpr(&kvx.Value, et)
		}
	case *SliceType:
		et := clt.Elt
		el := len(clx.Elts)
		for i := 0; i < el; i++ {
			kvx := &clx.Elts[i]
			elideCompositeExpr(&kvx.Value, et)
		}
	case *MapType:
		kt := clt.Key
		vt := clt.Value
		el := len(clx.Elts)
		for i := 0; i < el; i++ {
			kvx := &clx.Elts[i]
			elideCompositeExpr(&kvx.Key, kt)
			elideCompositeExpr(&kvx.Value, vt)
		}
	case *StructType:
		// Struct fields cannot be elided in Go for
		// legibility, but Gno could support them (e.g. for
		// certain tagged struct fields).
		// TODO: support eliding.
		for _, kvx := range clx.Elts {
			vx := kvx.Value
			if vclx, ok := vx.(*CompositeLitExpr); ok {
				if vclx.Type == nil {
					panic("types cannot be elided in composite literals for struct types")
				}
			}
		}
	case *nativeType:
		// TODO: support eliding.
		for _, kvx := range clx.Elts {
			vx := kvx.Value
			if vclx, ok := vx.(*CompositeLitExpr); ok {
				if vclx.Type == nil {
					panic("types cannot be elided in composite literals for native types")
				}
			}
		}
	default:
		panic(fmt.Sprintf(
			"unexpected composite lit type %s",
			clt.String()))
	}
}

// if *vx is composite lit type, fill in elided type.
// if composite type is pointer type, replace composite
// expression with ref expr.
func elideCompositeExpr(vx *Expr, vt Type) {
	if vclx, ok := (*vx).(*CompositeLitExpr); ok {
		if vclx.Type == nil {
			if vt.Kind() == PointerKind {
				vclx.Type = &constTypeExpr{
					Source: *vx,
					Type:   vt.Elem(),
				}
				*vx = &RefExpr{
					X: vclx,
				}
			} else {
				vclx.Type = &constTypeExpr{
					Source: *vx,
					Type:   vt,
				}
			}
		}
	}
}
