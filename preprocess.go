package gno

import (
	"fmt"
	"math/big"
	"reflect"
)

// The ctx passed in may be mutated if there are any statements
// or declarations.  The file or package which contains ctx may
// be mutated if there are any file-level declarations.
//
// Store is used to load external package values, but otherwise
// the package and newly created blocks/values are expected
// to be non-RefValues -- in some cases, nil is passed for store
// to enforce this.
//
// List of what Preprocess() does:
//  * Assigns BlockValuePath to NameExprs.
//  * TODO document what it does.
func Preprocess(store Store, ctx BlockNode, n Node) Node {
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

		// if already preprocessed, skip it.
		if n.GetAttribute(ATTR_PREPROCESSED) == true {
			return n, TRANS_SKIP
		}

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

		switch stage {

		//----------------------------------------
		case TRANS_ENTER:
			switch n := n.(type) {

			// TRANS_ENTER -----------------------
			case *AssignStmt:
				if n.Op == DEFINE {
					var defined bool
					for _, lx := range n.Lhs {
						ln := lx.(*NameExpr).Name
						if ln == "_" {
							// ignore.
						} else {
							_, ok := last.GetLocalIndex(ln)
							if !ok {
								// initial declaration to be re-defined.
								last.Define(ln, anyValue(nil))
								defined = true
							} else {
								// do not redeclare.
							}
						}
					}
					if !defined {
						panic(fmt.Sprintf("nothing defined in asssignment %s", n.String()))
					}
				} else {
					// nothing defined.
				}

			// TRANS_ENTER -----------------------
			case *ImportDecl, *ValueDecl, *TypeDecl, *FuncDecl:
				// NOTE func decl usually must happen with a
				// file, and so last is usually a *FileNode,
				// but for testing convenience we allow
				// importing directly onto the package.
				// Uverse requires this.
				if n.GetAttribute(ATTR_PREDEFINED) == true {
					// skip declarations already predefined
					// (e.g. through recursion for a dependent)
				} else {
					// recursively predefine dependencies.
					d2, ppd := predefineNow(store, last, n.(Decl))
					if ppd {
						return d2, TRANS_SKIP
					} else {
						return d2, TRANS_CONTINUE
					}
				}

			// TRANS_ENTER -----------------------
			case *FuncTypeExpr:
				for i, _ := range n.Params {
					p := &n.Params[i]
					if p.Name == "" {
						p.Name = "_"
					}
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
				// create faux block to store .Init.
				// the contents are copied onto the case block
				// in the if case below for .Body and .Else.
				// NOTE: similar to *SwitchStmt.
				pushBlock(n, &last, &stack)

			// TRANS_BLOCK -----------------------
			case *IfCaseStmt:
				pushRealBlock(n, &last, &stack)
				// parent if statement.
				ifs := ns[len(ns)-1].(*IfStmt)
				// anything declared in ifs are copied.
				for _, n := range ifs.GetBlockNames() {
					tv := ifs.GetValueRef(nil, n)
					last.Define(n, *tv)
				}

			// TRANS_BLOCK -----------------------
			case *RangeStmt:
				pushBlock(n, &last, &stack)
				// NOTE: preprocess it here, so type can
				// be used to set n.IsMap/IsString and
				// define key/value.
				n.X = Preprocess(store, last, n.X).(Expr)
				xt := evalStaticTypeOf(store, last, n.X)
				switch xt.Kind() {
				case MapKind:
					n.IsMap = true
				case StringKind:
					n.IsString = true
				case PointerKind:
					if xt.Elem().Kind() != ArrayKind {
						panic("range iteration over pointer requires array elem type")
					}
					xt = xt.Elem()
					n.IsArrayPtr = true
				}
				// key value if define.
				if n.Op == DEFINE {
					if xt.Kind() == MapKind {
						if n.Key != nil {
							kt := baseOf(xt).(*MapType).Key
							kn := n.Key.(*NameExpr).Name
							last.Define(kn, anyValue(kt))
						}
						if n.Value != nil {
							vt := baseOf(xt).(*MapType).Value
							vn := n.Value.(*NameExpr).Name
							last.Define(vn, anyValue(vt))
						}
					} else if xt.Kind() == StringKind {
						if n.Key != nil {
							it := IntType
							kn := n.Key.(*NameExpr).Name
							last.Define(kn, anyValue(it))
						}
						if n.Value != nil {
							et := Int32Type
							vn := n.Value.(*NameExpr).Name
							last.Define(vn, anyValue(et))
						}
					} else {
						if n.Key != nil {
							it := IntType
							kn := n.Key.(*NameExpr).Name
							last.Define(kn, anyValue(it))
						}
						if n.Value != nil {
							et := xt.Elem()
							vn := n.Value.(*NameExpr).Name
							last.Define(vn, anyValue(et))
						}
					}
				}

			// TRANS_BLOCK -----------------------
			case *FuncLitExpr:
				// retrieve cached function type.
				ft := evalStaticType(store, last, &n.Type).(*FuncType)
				// push func body block.
				pushBlock(n, &last, &stack)
				// define parameters in new block.
				for _, p := range ft.Params {
					last.Define(p.Name, anyValue(p.Type))
				}
				// define results in new block.
				for i, rf := range ft.Results {
					if 0 < len(rf.Name) {
						last.Define(rf.Name, anyValue(rf.Type))
					} else {
						// create a hidden var with leading dot.
						// NOTE: document somewhere.
						rn := fmt.Sprintf(".res_%d", i)
						last.Define(Name(rn), anyValue(rf.Type))
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
				// NOTE: similar to *IfStmt, but with the major
				// difference that each clause block may have
				// different number of values.
				// To support the .Init statement and for
				// conceptual simplicity, we create a block in
				// OpExec.SwitchStmt, but since we don't initially
				// know which clause will match, we expand the
				// block once a clause has matched.
				pushBlock(n, &last, &stack)
				if n.VarName != "" {
					// NOTE: this defines for default clauses too,
					// see comment on block copying @
					// SwitchClauseStmt:TRANS_BLOCK.
					last.Define(n.VarName, anyValue(nil))
				}
				// Preprocess and convert tag if const.
				if n.X != nil {
					n.X = Preprocess(store, last, n.X).(Expr)
					convertIfConst(store, last, n.X)
				}

			// TRANS_BLOCK -----------------------
			case *SwitchClauseStmt:
				pushRealBlock(n, &last, &stack)
				// parent switch statement.
				ss := ns[len(ns)-1].(*SwitchStmt)
				// anything declared in ss are copied,
				// namely ss.VarName if defined.
				for _, n := range ss.GetBlockNames() {
					tv := ss.GetValueRef(nil, n)
					last.Define(n, *tv)
				}
				if ss.IsTypeSwitch {
					// evaluate case types.
					for i, cx := range n.Cases {
						cx = Preprocess(
							store, last, cx).(Expr)
						var ct Type
						if cxx, ok := cx.(*constExpr); ok {
							if !cxx.IsUndefined() {
								panic("should not happen")
							}
							// only in type switch cases, nil type allowed.
							ct = nil
						} else {
							ct = evalStaticType(store, last, cx)
						}
						n.Cases[i] = constType(cx, ct)
						// maybe type-switch def.
						if 0 < len(ss.VarName) {
							if len(n.Cases) == 1 {
								// If there is only 1 case, the
								// define applies with type.
								// (re-definition).
								last.Define(
									ss.VarName, anyValue(ct))
							} else {
								// If there are 2 or more
								// cases, the type is the tag type.
								tt := evalStaticTypeOf(store, last, ss.X)
								last.Define(
									ss.VarName, anyValue(tt))
							}
						}
					}
				} else {
					// evalualte tag type
					tt := evalStaticTypeOf(store, last, ss.X)
					// check or convert case types to tt.
					for i, cx := range n.Cases {
						cx = Preprocess(
							store, last, cx).(Expr)
						n.Cases[i] = cx
						checkOrConvertType(store, last, cx, tt)
					}
				}

			// TRANS_BLOCK -----------------------
			case *FuncDecl:
				// retrieve cached function type.
				ft := getType(&n.Type).(*FuncType)
				if n.IsMethod {
					// recv/type set @ predefineNow().
				} else {
					// type set @ predefineNow().
				}

				// push func body block.
				pushBlock(n, &last, &stack)
				// define receiver in new block, if method.
				if n.IsMethod {
					if 0 < len(n.Recv.Name) {
						rft := getType(&n.Recv).(FieldType)
						rt := rft.Type
						last.Define(n.Recv.Name, anyValue(rt))
					}
				}
				// define parameters in new block.
				for _, p := range ft.Params {
					last.Define(p.Name, anyValue(p.Type))
				}
				// define results in new block.
				for i, rf := range ft.Results {
					if 0 < len(rf.Name) {
						last.Define(rf.Name, anyValue(rf.Type))
					} else {
						// create a hidden var with leading dot.
						rn := fmt.Sprintf(".res_%d", i)
						last.Define(Name(rn), anyValue(rf.Type))
					}
				}

			// TRANS_BLOCK -----------------------
			case *FileNode:
				// only for imports.
				pushBlock(n, &last, &stack)
				{
					// This logic supports out-of-order
					// declarations.  this is required
					// separately from the direct
					// predefineNow() entry callbacks above,
					// for otherwise out-of-order
					// declarations would not get pre-defined
					// before (say) the body of a function
					// declaration or literl can refer to it.
					// (this must happen after pushBlock
					// above, otherwise it would happen @
					// *FileNode:ENTER)

					// Predefine all type decls and import decls.
					for i := 0; i < len(n.Decls); i++ {
						d := n.Decls[i]
						switch d.(type) {
						case *ImportDecl, *TypeDecl:
							if d.GetAttribute(ATTR_PREDEFINED) == true {
								// skip declarations already
								// predefined (e.g. through
								// recursion for a dependent)
							} else {
								// recursively predefine
								// dependencies.
								d2, _ := predefineNow(store, n, d)
								n.Decls[i] = d2
							}
						}
					}
					// Then, predefine all func/method decls.
					for i := 0; i < len(n.Decls); i++ {
						d := n.Decls[i]
						switch d.(type) {
						case *FuncDecl:
							if d.GetAttribute(ATTR_PREDEFINED) == true {
								// skip declarations already
								// predefined (e.g. through
								// recursion for a dependent)
							} else {
								// recursively predefine
								// dependencies.
								d2, _ := predefineNow(store, n, d)
								n.Decls[i] = d2
							}
						}
					}
					// Finally, predefine other decls and
					// preprocess ValueDecls..
					for i := 0; i < len(n.Decls); i++ {
						d := n.Decls[i]
						if d.GetAttribute(ATTR_PREDEFINED) == true {
							// skip declarations already
							// predefined (e.g. through
							// recursion for a dependent)
						} else {
							// recursively predefine
							// dependencies.
							d2, _ := predefineNow(store, n, d)
							n.Decls[i] = d2
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
			// in evalStaticType(store,).
			n.SetAttribute(ATTR_PREPROCESSED, true)

			//-There is still work to be done while leaving, but
			//once the logic of that is done, we will have to
			//perform additionally deferred logic that is best
			//handled with orthogonal switch conditions.
			//-For example, while leaving nodes w/
			//TRANS_COMPOSITE_TYPE, (regardless of whether name or
			//literal), any elided type names are inserted. (This
			//works because the transcriber leaves the composite
			//type before entering the kv elements.)
			defer func() {
				switch ftype {

				// TRANS_LEAVE (deferred)---------
				case TRANS_COMPOSITE_TYPE:
					// fill elided element composite lit type exprs
					clx := ns[len(ns)-1].(*CompositeLitExpr)
					// get or evaluate composite type.
					clt := evalStaticType(store, last, n.(Expr))
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
					clt := evalStaticType(store, last, clx.Type)
					switch bt := baseOf(clt).(type) {
					case *StructType:
						n.Path = bt.GetPathForName(n.Name)
						return n, TRANS_CONTINUE
					case *ArrayType, *SliceType:
						// Replace n with *constExpr.
						fillNameExprPath(last, n)
						cv := evalConst(store, last, n)
						return cv, TRANS_CONTINUE
					case *nativeType:
						switch bt.Type.Kind() {
						case reflect.Struct:
							// NOTE: For simplicity and some degree of
							// flexibility, do not use path indices for Go
							// native types, but use the name.
							n.Path = NewValuePathNative(n.Name)
							return n, TRANS_CONTINUE
						case reflect.Array, reflect.Slice:
							// Replace n with *constExpr.
							fillNameExprPath(last, n)
							cv := evalConst(store, last, n)
							return cv, TRANS_CONTINUE
						default:
							panic("should not happen")
						}
					}
				}
				// specific and general cases
				switch n.Name {
				case "_":
					n.Path = NewValuePathBlock(0, 0, "_")
					return n, TRANS_CONTINUE
				case "iota":
					pd := lastDecl(ns)
					io := pd.GetAttribute(ATTR_IOTA).(int)
					cx := constUntypedBigint(n, int64(io))
					return cx, TRANS_CONTINUE
				case "nil":
					// nil will be converted to typed-nils when
					// appropriate upon leaving the expression
					// nodes that contain nil nodes.
					fallthrough
				default:
					fillNameExprPath(last, n)
					// If uverse, return a *constExpr.
					if n.Path.Depth == 0 { // uverse
						cv := evalConst(store, last, n)
						// built-in functions must be called.
						if !cv.IsUndefined() &&
							cv.T.Kind() == FuncKind &&
							ftype != TRANS_CALL_FUNC {
							panic(fmt.Sprintf(
								"use of builtin %s not in function call",
								n.Name))
						}
						if !cv.IsUndefined() && cv.T.Kind() == TypeKind {
							return constType(n, cv.GetType()), TRANS_CONTINUE
						}
						return cv, TRANS_CONTINUE
					}
					// If untyped const, return it as *constExpr.
					nt := evalStaticTypeOf(store, last, n)
					if isUntyped(nt) {
						cx := evalConst(store, last, n)
						return cx, TRANS_CONTINUE
					}
				}

			// TRANS_LEAVE -----------------------
			case *BasicLitExpr:
				// Replace with *constExpr.
				cv := evalConst(store, last, n)
				return cv, TRANS_CONTINUE

			// TRANS_LEAVE -----------------------
			case *BinaryExpr:
				lt := evalStaticTypeOf(store, last, n.Left)
				rt := evalStaticTypeOf(store, last, n.Right)
				// Special (recursive) case if shift and right isn't uint.
				isShift := n.Op == SHL || n.Op == SHR
				if isShift && baseOf(rt) != UintType {
					// convert n.Right to (gno) uint type,
					rn := Expr(Call("uint", n.Right))
					// reset/create n2 to preprocess right child.
					n2 := &BinaryExpr{
						Left:  n.Left,
						Op:    n.Op,
						Right: rn,
					}
					resn := Preprocess(store, last, n2)
					return resn, TRANS_CONTINUE
				}
				// General case.
				lcx, lic := n.Left.(*constExpr)
				rcx, ric := n.Right.(*constExpr)
				if lic {
					if ric {
						// Left const, Right const ----------------------
						// Replace with *constExpr if const operands.
						cv := evalConst(store, last, n)
						return cv, TRANS_CONTINUE
					} else if isUntyped(lcx.T) {
						// Left untyped const, Right not ----------------
						if rnt, ok := rt.(*nativeType); ok {
							if isShift {
								panic("should not happen")
							}
							// get concrete native base type.
							pt := go2GnoBaseType(rnt.Type).(PrimitiveType)
							// convert n.Left to pt type,
							checkOrConvertType(store, last, n.Left, pt)
							// convert n.Right to (gno) pt type,
							rn := Expr(Call(pt.String(), n.Right))
							// and convert result back.
							tx := constType(n, rnt)
							// reset/create n2 to preprocess right child.
							n2 := &BinaryExpr{
								Left:  n.Left,
								Op:    n.Op,
								Right: rn,
							}
							resn := Node(Call(tx, n2))
							resn = Preprocess(store, last, resn)
							return resn, TRANS_CONTINUE
							// NOTE: binary operations are always computed in
							// gno, never with reflect.
						} else {
							if isShift {
								// nothing to do, right type is (already) uint type.
							} else {
								// convert n.Left to right type.
								checkOrConvertType(store, last, n.Left, rt)
							}
						}
					} else if lcx.T == nil {
						// convert n.Left to typed-nil type.
						checkOrConvertType(store, last, n.Left, rt)
					}
				} else if ric {
					if isUntyped(rcx.T) {
						// Left not, Right untyped const ----------------
						if isShift {
							if baseOf(rt) != UintType {
								// convert n.Right to (gno) uint type.
								checkOrConvertType(store, last, n.Right, UintType)
							} else {
								// leave n.Left as is and baseOf(n.Right) as UintType.
							}
						} else {
							if lnt, ok := lt.(*nativeType); ok {
								// get concrete native base type.
								pt := go2GnoBaseType(lnt.Type).(PrimitiveType)
								// convert n.Left to (gno) pt type,
								ln := Expr(Call(pt.String(), n.Left))
								// convert n.Right to pt type,
								checkOrConvertType(store, last, n.Right, pt)
								// and convert result back.
								tx := constType(n, lnt)
								// reset/create n2 to preprocess left child.
								n2 := &BinaryExpr{
									Left:  ln,
									Op:    n.Op,
									Right: n.Right,
								}
								resn := Node(Call(tx, n2))
								resn = Preprocess(store, last, resn)
								return resn, TRANS_CONTINUE
								// NOTE: binary operations are always computed in
								// gno, never with reflect.
							} else {
								// convert n.Right to left type.
								checkOrConvertType(store, last, n.Right, lt)
							}
						}
					} else if rcx.T == nil {
						// convert n.Right to typed-nil type.
						checkOrConvertType(store, last, n.Right, lt)
					}
				} else {
					// Left not const, Right not const ------------------
					if n.Op == EQL || n.Op == NEQ {
						// If == or !=, no conversions.
					} else if lnt, ok := lt.(*nativeType); ok {
						if debug {
							if !isShift {
								assertSameTypes(lt, rt)
							}
						}
						// If left and right are native type,
						// convert left and right to gno, then
						// convert result back to native.
						//
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
						tx := constType(n, lnt)
						// reset/create n2 to preprocess
						// children.
						n2 := &BinaryExpr{
							Left:  ln,
							Op:    n.Op,
							Right: rn,
						}
						resn := Node(Call(tx, n2))
						resn = Preprocess(store, last, resn)
						return resn, TRANS_CONTINUE
						// NOTE: binary operations are always
						// computed in gno, never with
						// reflect.
					} else {
						// nothing to do.
					}
				}

			// TRANS_LEAVE -----------------------
			case *CallExpr:
				// Func type evaluation.
				var ft *FuncType
				ift := evalStaticTypeOf(store, last, n.Func)
				switch cft := baseOf(ift).(type) {
				case *FuncType:
					ft = cft
				case *nativeType:
					ft = go2GnoFuncType(cft.Type)
				case *TypeType:
					if len(n.Args) != 1 {
						panic("type conversion requires single argument")
					}
					if _, ok := n.Args[0].(*constExpr); ok {
						convertIfConst(store, last, n.Args[0])
						cv := evalConst(store, last, n)
						return cv, TRANS_CONTINUE
					} else {
						ct := evalStaticType(store, last, n.Func)
						n.SetAttribute(ATTR_TYPEOF_VALUE, ct)
						return n, TRANS_CONTINUE
					}
				default:
					panic(fmt.Sprintf(
						"unexpected func type %v (%v)",
						ift, reflect.TypeOf(ift)))
				}
				hasVarg := ft.HasVarg()
				isVarg := n.Varg
				embedded := false
				argTVs := []TypedValue{}
				minArgs := len(ft.Params)
				if hasVarg {
					minArgs--
				}
				numArgs := countNumArgs(store, last, n) // isVarg?
				n.NumArgs = numArgs
				// Check input arg count.
				if len(n.Args) == 1 && numArgs > 1 {
					// special case of x(f()) form:
					// use the number of results instead.
					if isVarg {
						panic("should not happen")
					}
					embedded = true
					pcx := n.Args[0].(*CallExpr)
					argTVs = getResultTypedValues(pcx)
					if !hasVarg {
						if numArgs != len(ft.Params) {
							panic(fmt.Sprintf(
								"wrong argument count in call to %s; want %d got %d (with embedded call expr as arg)",
								n.Func.String(),
								len(ft.Params),
								numArgs,
							))
						}
					} else if hasVarg && !isVarg {
						if numArgs < len(ft.Params)-1 {
							panic(fmt.Sprintf(
								"not enough arguments in call to %s; want %d (besides variadic) got %d (with embedded call expr as arg)",
								n.Func.String(),
								len(ft.Params)-1,
								numArgs))
						}
					}
				} else if !hasVarg {
					argTVs = evalStaticTypedValues(store, last, n.Args...)
					if len(n.Args) != len(ft.Params) {
						panic(fmt.Sprintf(
							"wrong argument count in call to %s; want %d got %d",
							n.Func.String(),
							len(ft.Params),
							len(n.Args),
						))
					}
				} else if hasVarg && !isVarg {
					argTVs = evalStaticTypedValues(store, last, n.Args...)
					if len(n.Args) < len(ft.Params)-1 {
						panic(fmt.Sprintf(
							"not enough arguments in call to %s; want %d (besides variadic) got %d",
							n.Func.String(),
							len(ft.Params)-1,
							len(n.Args)))
					}
				} else if hasVarg && isVarg {
					argTVs = evalStaticTypedValues(store, last, n.Args...)
					if len(n.Args) != len(ft.Params) {
						panic(fmt.Sprintf(
							"not enough arguments in call to %s; want %d (including variadic) got %d",
							n.Func.String(),
							len(ft.Params),
							len(n.Args)))
					}
				} else {
					panic("should not happen")
				}
				// Specify function param/result generics.
				sft := ft.Specify(argTVs, isVarg)
				spts := sft.Params
				srts := FieldTypeList(sft.Results).Types()
				n.SetAttribute(ATTR_TYPEOF_VALUE,
					&tupleType{Elts: srts})
				// Replace const Args with *constExpr.
				if !embedded {
					for i, arg := range n.Args {
						if hasVarg {
							if (len(spts) - 1) <= i {
								if isVarg {
									if len(spts) <= i {
										panic("expected final vargs slice but got many")
									}
									checkOrConvertType(store, last, arg, spts[i].Type)
								} else {
									checkOrConvertType(store, last, arg,
										spts[len(spts)-1].Type.Elem())
								}
							} else {
								checkOrConvertType(store, last, arg, spts[i].Type)
							}
						} else {
							checkOrConvertType(store, last, arg, spts[i].Type)
						}
					}
				}
				// TODO in the future, pure results

			// TRANS_LEAVE -----------------------
			case *IndexExpr:
				dt := evalStaticTypeOf(store, last, n.X)
				if dt.Kind() == PointerKind {
					// if a is a pointer to an array,
					// a[low : high : max] is shorthand
					// for (*a)[low : high : max]
					dt = dt.Elem()
					n.X = &StarExpr{X: n.X}
					n.X.SetAttribute(ATTR_PREPROCESSED, true)
				}
				switch dt.Kind() {
				case StringKind, ArrayKind, SliceKind:
					// Replace const index with int *constExpr,
					// or if not const, assert integer type..
					checkOrConvertIntegerType(store, last, n.Index)
				case MapKind:
					mt := baseOf(gnoTypeOf(dt)).(*MapType)
					checkOrConvertType(store, last, n.Index, mt.Key)
				default:
					panic(fmt.Sprintf(
						"unexpected index base kind for type %s",
						dt.String()))

				}

			// TRANS_LEAVE -----------------------
			case *SliceExpr:
				// Replace const L/H/M with int *constExpr,
				// or if not const, assert integer type..
				checkOrConvertIntegerType(store, last, n.Low)
				checkOrConvertIntegerType(store, last, n.High)
				checkOrConvertIntegerType(store, last, n.Max)

			// TRANS_LEAVE -----------------------
			case *TypeAssertExpr:
				if n.Type == nil {
					panic("should not happen")
				}
				// ExprStmt of form `x.(<type>)`,
				// or special case form `c, ok := x.(<type>)`.
				evalStaticType(store, last, n.Type)

			// TRANS_LEAVE -----------------------
			case *UnaryExpr:
				xt := evalStaticTypeOf(store, last, n.X)
				if xnt, ok := xt.(*nativeType); ok {
					// get concrete native base type.
					pt := go2GnoBaseType(xnt.Type).(PrimitiveType)
					// convert n.X to gno type,
					xn := Expr(Call(pt.String(), n.X))
					// and convert result back.
					tx := constType(n, xnt)
					// reset/create n2 to preprocess children.
					n2 := &UnaryExpr{
						X:  xn,
						Op: n.Op,
					}
					resn := Node(Call(tx, n2))
					resn = Preprocess(store, last, resn)
					return resn, TRANS_CONTINUE
					// NOTE: like binary operations, unary operations are
					// always computed in gno, never with reflect.
				}
				// Replace with *constExpr if const X.
				if isConst(n.X) {
					cv := evalConst(store, last, n)
					return cv, TRANS_CONTINUE
				}

			// TRANS_LEAVE -----------------------
			case *CompositeLitExpr:
				// Get or evaluate composite type.
				clt := evalStaticType(store, last, n.Type)
				// Replace const Elts with default *constExpr.
			CLT_TYPE_SWITCH:
				switch cclt := baseOf(clt).(type) {
				case *StructType:
					if n.IsKeyed() {
						for i := 0; i < len(n.Elts); i++ {
							key := n.Elts[i].Key.(*NameExpr).Name
							path := cclt.GetPathForName(key)
							ft := cclt.GetStaticTypeOfAt(path)
							checkOrConvertType(store, last, n.Elts[i].Value, ft)
						}
					} else {
						for i := 0; i < len(n.Elts); i++ {
							ft := cclt.Fields[i].Type
							checkOrConvertType(store, last, n.Elts[i].Value, ft)
						}
					}
				case *ArrayType:
					for i := 0; i < len(n.Elts); i++ {
						checkOrConvertType(store, last, n.Elts[i].Key, IntType)
						checkOrConvertType(store, last, n.Elts[i].Value, cclt.Elt)
					}
				case *SliceType:
					for i := 0; i < len(n.Elts); i++ {
						checkOrConvertType(store, last, n.Elts[i].Key, IntType)
						checkOrConvertType(store, last, n.Elts[i].Value, cclt.Elt)
					}
				case *MapType:
					for i := 0; i < len(n.Elts); i++ {
						checkOrConvertType(store, last, n.Elts[i].Key, cclt.Key)
						checkOrConvertType(store, last, n.Elts[i].Value, cclt.Value)
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
								k := evalConst(store, last, elt.Key).ConvertGetInt()
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
				xt := evalStaticTypeOf(store, last, n.X)

				// Set selector path based on xt's type.
				switch cxt := xt.(type) {
				case *PointerType, *DeclaredType, *StructType, *InterfaceType:
					tr, _, rcvr, _ := findEmbeddedFieldType(cxt, n.Sel)
					if tr == nil {
						panic(fmt.Sprintf("missing field %s in %s", n.Sel,
							cxt.String()))
					}
					if len(tr) > 1 {
						// (the last vp, tr[len(tr)-1], is for n.Sel)
						if debug {
							if tr[len(tr)-1].Name != n.Sel {
								panic("should not happen")
							}
						}
						// replace n.X w/ tr[:len-1] selectors applied.
						nx2 := n.X
						for _, vp := range tr[:len(tr)-1] {
							nx2 = &SelectorExpr{
								X:    nx2,
								Path: vp,
								Sel:  vp.Name,
							}
						}
						// recursively preprocess new n.X.
						n.X = Preprocess(store, last, nx2).(Expr)
					}
					// nxt2 may not be xt anymore.
					// (even the dereferenced of xt and nxt2 may not
					// be the same, with embedded fields)
					nxt2 := evalStaticTypeOf(store, last, n.X)
					// Case 1: If receiver is pointer type but n.X is
					// not:
					if rcvr != nil &&
						rcvr.Kind() == PointerKind &&
						nxt2.Kind() != PointerKind {
						// Go spec: "If x is addressable and &x's
						// method set contains m, x.m() is shorthand
						// for (&x).m()"
						// Go spec: "As with method calls, a reference
						// to a non-interface method with a pointer
						// receiver using an addressable value will
						// automatically take the address of that
						// value: t.Mp is equivalent to (&t).Mp."
						//
						// convert to (&x).m, but leave xt as is.
						n.X = &RefExpr{X: n.X}
						n.X.SetAttribute(ATTR_PREPROCESSED, true)
						switch tr[len(tr)-1].Type {
						case VPDerefPtrMethod:
							// When ptr method was called like x.y.z(), where x
							// is a pointer, y is an embedded struct, and z
							// takes a pointer receiver.  That becomes
							// &(x.y).z().
							// The x.y receiver wasn't originally a pointer,
							// yet the trail was
							// [VPSubrefField,VPDerefPtrMethod].
						case VPPtrMethod:
							tr[len(tr)-1].Type = VPDerefPtrMethod
						default:
							panic(fmt.Sprintf(
								"expected ultimate VPPtrMethod but got %v in trail %v",
								tr[len(tr)-1].Type,
								tr,
							))
						}
					} else if len(tr) > 0 &&
						tr[len(tr)-1].IsDerefType() &&
						nxt2.Kind() != PointerKind {
						// Case 2: If tr[0] is deref type, but xt
						// is not pointer type, replace n.X with
						// &RefExpr{X: n.X}.
						n.X = &RefExpr{X: n.X}
						n.X.SetAttribute(ATTR_PREPROCESSED, true)
					}
					// bound method or underlying.
					// TODO check for unexported fields.
					n.Path = tr[len(tr)-1]
					// n.Path = cxt.GetPathForName(n.Sel)
				case *PackageType:
					// packages can only be referred to by
					// *NameExprs, and cannot be copied.
					nx := n.X.(*NameExpr)
					pv := last.GetValueRef(nil, nx.Name)
					pn := pv.V.(*PackageValue).Source
					n.Path = pn.GetPathForName(store, n.Sel)
				case *TypeType:
					// unbound method
					xt := evalStaticType(store, last, n.X)
					switch ct := xt.(type) {
					case *PointerType:
						dt := ct.Elt.(*DeclaredType)
						n.Path = dt.GetUnboundPathForName(n.Sel)
					case *DeclaredType:
						n.Path = ct.GetUnboundPathForName(n.Sel)
					default:
						panic(fmt.Sprintf(
							"unexpected selector expression type value %s",
							xt.String()))
					}
				case *nativeType:
					// NOTE: if type of n.X is native type, as in a native
					// interface method, n.Path may be VPNative but at
					// runtime, the value's type may be *gno.PointerType.
					//
					// native types don't use path indices.
					n.Path = NewValuePathNative(n.Sel)
				default:
					panic(fmt.Sprintf(
						"unexpected selector expression type %v",
						reflect.TypeOf(xt)))
				}

			// TRANS_LEAVE -----------------------
			case *FieldTypeExpr:
				// Replace const Tag with default *constExpr.
				convertIfConst(store, last, n.Tag)

			// TRANS_LEAVE -----------------------
			case *ArrayTypeExpr:
				if n.Len == nil {
					// Calculate length at *CompositeLitExpr:LEAVE
				} else {
					// Replace const Len with int *constExpr.
					cx := evalConst(store, last, n.Len)
					convertConst(store, last, cx, IntType)
					n.Len = cx
				}
				// NOTE: For all TypeExprs, the node is not replaced
				// with *constTypeExprs (as *constExprs are) because
				// we want to support type logic at runtime.
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *SliceTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *InterfaceTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *ChanTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *FuncTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *MapTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *StructTypeExpr:
				evalStaticType(store, last, n)

			// TRANS_LEAVE -----------------------
			case *AssignStmt:
				// NOTE: keep DEFINE and ASSIGN in sync.
				if n.Op == DEFINE {
					// Rhs consts become default *constExprs.
					for _, rx := range n.Rhs {
						// NOTE: does nothing if rx is "nil".
						convertIfConst(store, last, rx)
					}
					if len(n.Lhs) > len(n.Rhs) {
						// Unpack n.Rhs[0] to n.Lhs[:]
						if len(n.Rhs) != 1 {
							panic("should not happen")
						}
						switch cx := n.Rhs[0].(type) {
						case *CallExpr:
							// Call case: a, b := x(...)
							ift := evalStaticTypeOf(store, last, cx.Func)
							cft := getGnoFuncTypeOf(ift)
							if len(n.Lhs) != len(cft.Results) {
								panic(fmt.Sprintf(
									"assignment mismatch: "+
										"%d variables but %s returns %d values",
									len(n.Lhs), cx.Func.String(), len(cft.Results)))
							}
							for i, lx := range n.Lhs {
								ln := lx.(*NameExpr).Name
								rf := cft.Results[i]
								// re-definition
								last.Define(ln, anyValue(rf.Type))
							}
						case *TypeAssertExpr:
							// Type-assert case: a, ok := x.(type)
							if len(n.Lhs) != 2 {
								panic("should not happen")
							}
							cx.HasOK = true
							lhs0 := n.Lhs[0].(*NameExpr).Name
							lhs1 := n.Lhs[1].(*NameExpr).Name
							tt := evalStaticType(store, last, cx.Type)
							// re-definitions
							last.Define(lhs0, anyValue(tt))
							last.Define(lhs1, anyValue(BoolType))
						case *IndexExpr:
							// Index case: v, ok := x[k], x is map.
							if len(n.Lhs) != 2 {
								panic("should not happen")
							}
							cx.HasOK = true
							lhs0 := n.Lhs[0].(*NameExpr).Name
							lhs1 := n.Lhs[1].(*NameExpr).Name
							mt := evalStaticTypeOf(store, last, cx.X).(*MapType)
							// re-definitions
							last.Define(lhs0, anyValue(mt.Value))
							last.Define(lhs1, anyValue(BoolType))
						default:
							panic("should not happen")
						}
					} else {
						// General case: a, b := x, y
						for i, lx := range n.Lhs {
							ln := lx.(*NameExpr).Name
							rx := n.Rhs[i]
							rt := evalStaticTypeOf(store, last, rx)
							// re-definition
							if rt == nil {
								// e.g. (interface{})(nil), becomes constExpr(undefined).
								// last.Define(ln, undefined) complains, since redefinition.
							} else {
								last.Define(ln, anyValue(rt))
							}
						}
					}
				} else { // ASSIGN.
					// NOTE: Keep in sync with DEFINE above.
					if len(n.Lhs) > len(n.Rhs) {
						// TODO dry code w/ above.
						// Unpack n.Rhs[0] to n.Lhs[:]
						if len(n.Rhs) != 1 {
							panic("should not happen")
						}
						switch cx := n.Rhs[0].(type) {
						case *CallExpr:
							// Call case: a, b = x(...)
							ift := evalStaticTypeOf(store, last, cx.Func)
							cft := getGnoFuncTypeOf(ift)
							if len(n.Lhs) != len(cft.Results) {
								panic(fmt.Sprintf(
									"assignment mismatch: "+
										"%d variables but %s returns %d values",
									len(n.Lhs), cx.Func.String(), len(cft.Results)))
							}
						case *TypeAssertExpr:
							// Type-assert case: a, ok := x.(type)
							if len(n.Lhs) != 2 {
								panic("should not happen")
							}
							cx.HasOK = true
						case *IndexExpr:
							// Index case: v, ok := x[k], x is map.
							if len(n.Lhs) != 2 {
								panic("should not happen")
							}
							cx.HasOK = true
						default:
							panic("should not happen")
						}
					} else {
						// General case: a, b = x, y.
						for i, lx := range n.Lhs {
							lt := evalStaticTypeOf(store, last, lx)
							rx := n.Rhs[i]
							// converts if rx is "nil".
							checkOrConvertType(store, last, rx, lt)
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *BranchStmt:
				switch n.Op {
				case BREAK:
				case CONTINUE:
				case GOTO:
					_, depth, index := findLabel(last, n.Label)
					n.Depth = depth
					n.BodyIndex = index
				case FALLTHROUGH:
					// TODO CHALLENGE implement fallthrough
				default:
					panic("should not happen")
				}

			// TRANS_LEAVE -----------------------
			case *ForStmt:
				// Cond consts become bool *constExprs.
				checkOrConvertType(store, last, n.Cond, BoolType)

			// TRANS_LEAVE -----------------------
			case *IfStmt:
				// Cond consts become bool *constExprs.
				checkOrConvertType(store, last, n.Cond, BoolType)

			// TRANS_LEAVE -----------------------
			case *RangeStmt:
				// NOTE: k,v already defined @ TRANS_BLOCK.

			// TRANS_LEAVE -----------------------
			case *ReturnStmt:
				fnode, ft := funcNodeOf(last)
				// Check number of return arguments.
				if len(n.Results) != len(ft.Results) {
					if len(n.Results) == 0 {
						if ft.Results.IsNamed() {
							// ok, results already named.
						} else {
							panic(fmt.Sprintf("expected %d return values; got %d",
								len(ft.Results),
								len(n.Results),
							))
						}
					} else if len(n.Results) == 1 {
						if cx, ok := n.Results[0].(*CallExpr); ok {
							ift := evalStaticTypeOf(store, last, cx.Func)
							cft := getGnoFuncTypeOf(ift)
							if len(cft.Results) != len(ft.Results) {
								panic(fmt.Sprintf("expected %d return values; got %d",
									len(ft.Results),
									len(cft.Results),
								))
							} else {
								// nothing more to do.
							}
						} else {
							panic(fmt.Sprintf("expected %d return values; got %d",
								len(ft.Results),
								len(n.Results),
							))
						}
					} else {
						panic(fmt.Sprintf("expected %d return values; got %d",
							len(ft.Results),
							len(n.Results),
						))
					}
				} else {
					// Results consts become default *constExprs.
					for i, rx := range n.Results {
						rtx := ft.Results[i].Type
						rt := evalStaticType(store, fnode.GetParentNode(nil), rtx)
						if isGeneric(rt) {
							// cannot convert generic result,
							// the result type depends.
							// XXX how to deal?
							panic("not yet implemented")
						} else {
							checkOrConvertType(store, last, rx, rt)
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *SendStmt:
				// Value consts become default *constExprs.
				checkOrConvertType(store, last, n.Value, nil)

			// TRANS_LEAVE -----------------------
			case *SelectCaseStmt:
				// maybe receive defines.
				// if as, ok := n.Comm.(*AssignStmt); ok {
				//     handled by case *AssignStmt.
				// }

			// TRANS_LEAVE -----------------------
			case *SwitchStmt:
				// Ensure type switch cases are unique.
				if n.IsTypeSwitch {
					types := map[string]struct{}{}
					for _, clause := range n.Clauses {
						for _, casetype := range clause.Cases {
							var ctstr string
							ctype := casetype.(*constTypeExpr).Type
							if ctype == nil {
								ctstr = "nil"
							} else {
								ctstr = casetype.(*constTypeExpr).Type.String()
							}
							if _, exists := types[ctstr]; exists {
								panic(fmt.Sprintf(
									"duplicate type %s in type switch",
									ctstr))
							}
							types[ctstr] = struct{}{}
						}
					}
				}

			// TRANS_LEAVE -----------------------
			case *ValueDecl:
				// evaluate value if const expr.
				if n.Const {
					// NOTE: may or may not be a *constExpr,
					// but if not, make one now.
					for i, vx := range n.Values {
						n.Values[i] = evalConst(store, last, vx)
					}
				} else {
					// value(s) may already be *constExpr, but
					// otherwise as far as we know the
					// expression is not a const expr, so no
					// point evaluating it further.  this makes
					// the implementation differ from
					// runDeclaration(), as this uses OpStaticTypeOf.
				}
				var numNames = len(n.NameExprs)
				var sts = make([]Type, numNames) // static types
				var tvs = make([]TypedValue, numNames)
				if numNames > 1 && len(n.Values) == 1 {
					// special case if `var a, b, c T? = f()` form.
					cx := n.Values[0].(*CallExpr)
					tt := evalStaticTypeOfRaw(store, last, cx).(*tupleType)
					if len(tt.Elts) != numNames {
						panic("should not happen")
					}
					if n.Type != nil {
						// only a single type can be specified.
						nt := evalStaticType(store, last, n.Type)
						// TODO check tt and nt compat.
						for i := 0; i < numNames; i++ {
							sts[i] = nt
							tvs[i] = anyValue(nt)
						}
					} else {
						// set types as return types.
						for i := 0; i < numNames; i++ {
							et := tt.Elts[i]
							sts[i] = et
							tvs[i] = anyValue(et)
						}
					}
				} else if len(n.Values) != 0 && numNames != len(n.Values) {
					panic("should not happen")
				} else { // general case
					// evaluate types and convert consts.
					if n.Type != nil {
						// only a single type can be specified.
						nt := evalStaticType(store, last, n.Type)
						for i := 0; i < numNames; i++ {
							sts[i] = nt
						}
						// convert if const to nt.
						for _, vx := range n.Values {
							checkOrConvertType(store, last, vx, nt)
						}
					} else if n.Const {
						// derive static type from values.
						for i, vx := range n.Values {
							vt := evalStaticTypeOf(store, last, vx)
							sts[i] = vt
						}
					} else {
						// convert n.Value to default type.
						for i, vx := range n.Values {
							convertIfConst(store, last, vx)
							vt := evalStaticTypeOf(store, last, vx)
							sts[i] = vt
						}
					}
					// evaluate typed value for static definition.
					for i, vx := range n.Values {
						if cx, ok := vx.(*constExpr); ok &&
							!cx.TypedValue.IsUndefined() {
							// if value is non-nil const expr:
							tvs[i] = cx.TypedValue
						} else {
							// for var decls of non-const expr.
							st := sts[i]
							tvs[i] = anyValue(st)
						}
					}
				}
				// define.
				if fn, ok := last.(*FileNode); ok {
					pn := fn.GetParentNode(nil).(*PackageNode)
					for i := 0; i < numNames; i++ {
						nx := &n.NameExprs[i]
						if nx.Name == "_" {
							nx.Path = NewValuePathBlock(0, 0, "_")
						} else {
							pn.Define2(nx.Name, sts[i], tvs[i])
							nx.Path = last.GetPathForName(nil, nx.Name)
						}
					}
				} else {
					for i := 0; i < numNames; i++ {
						nx := &n.NameExprs[i]
						if nx.Name == "_" {
							nx.Path = NewValuePathBlock(0, 0, "_")
						} else {
							last.Define2(nx.Name, sts[i], tvs[i])
							nx.Path = last.GetPathForName(nil, nx.Name)
						}
					}
				}
				// TODO make note of constance in static block for
				// future use, or consider "const paths".  set as
				// preprocessed.

			// TRANS_LEAVE -----------------------
			case *TypeDecl:
				// Construct new Type, where any recursive
				// references refer to the old Type declared
				// during *TypeDecl:ENTER.  Then, copy over the
				// values, completing the recursion.
				tmp := evalStaticType(store, last, n.Type)
				dst := last.GetValueRef(store, n.Name).GetType()
				switch dst := dst.(type) {
				case *FuncType:
					*dst = *(tmp.(*FuncType))
				case *ArrayType:
					*dst = *(tmp.(*ArrayType))
				case *SliceType:
					*dst = *(tmp.(*SliceType))
				case *InterfaceType:
					*dst = *(tmp.(*InterfaceType))
				case *ChanType:
					*dst = *(tmp.(*ChanType))
				case *MapType:
					*dst = *(tmp.(*MapType))
				case *StructType:
					*dst = *(tmp.(*StructType))
				case *DeclaredType:
					pn := packageOf(last)
					// NOTE: this is where declared types are
					// actually instantiated, not in
					// interpret.go:runDeclaration().
					dt := declareWith(pn.PkgPath, n.Name, tmp)
					// if !n.IsAlias { // not sure why this was here.
					dt.Seal()
					//}
					*dst = *dt
				default:
					panic(fmt.Sprintf("unexpected type declaration type %v",
						reflect.TypeOf(dst)))
				}
				// We need to replace all references of the new
				// Type with old Type, including in attributes.
				n.Type.SetAttribute(ATTR_TYPE_VALUE, dst)
				// Replace the type with *constTypeExpr{},
				// otherwise methods would be un at runtime.
				n.Type = constType(n.Type, dst)
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

// like pushBlock(), but when the last block is a faux block,
// namely after SwitchStmt and IfStmt.
func pushRealBlock(bn BlockNode, last *BlockNode, stack *[]BlockNode) {
	orig := *last
	// skip the faux block for parent of bn.
	bn.InitStaticBlock(bn, (*last).GetParentNode(nil))
	*last = bn
	*stack = append(*stack, bn)
	// anything declared in orig are copied.
	for _, n := range orig.GetBlockNames() {
		tv := orig.GetValueRef(nil, n)
		bn.Define(n, *tv)
	}
}

// Evaluates the value of x which is expected to be a typeval.
// Caches the result as an attribute of x.
// To discourage mis-use, expects x to already be
// preprocessed.
func evalStaticType(store Store, last BlockNode, x Expr) Type {
	if t, ok := x.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
		return t
	} else if ctx, ok := x.(*constTypeExpr); ok {
		return ctx.Type // no need to set attribute.
	}
	pn := packageOf(last)
	tv := NewMachine(pn.PkgPath, store).EvalStatic(last, x)
	if _, ok := tv.V.(TypeValue); !ok {
		panic(fmt.Sprintf("%s is not a type", x.String()))
	}
	t := tv.GetType()
	x.SetAttribute(ATTR_TYPE_VALUE, t)
	return t
}

// If it is known that the type was already evaluated,
// use this function instead of evalStaticType(store,).
func getType(x Expr) Type {
	if ctx, ok := x.(*constTypeExpr); ok {
		return ctx.Type
	} else if t, ok := x.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
		return t
	} else {
		panic(fmt.Sprintf(
			"getType() called on expr not yet evaluated with evalStaticType(store,): %s",
			x.String(),
		))
	}
}

// If t is a native type, returns the gno type.
func gnoTypeOf(t Type) Type {
	if nt, ok := t.(*nativeType); ok {
		return nt.GnoType()
	} else {
		return t
	}
}

// Unlike evalStaticType, x is not expected to be a typeval,
// but rather computes the type OF x.
func evalStaticTypeOf(store Store, last BlockNode, x Expr) Type {
	t := evalStaticTypeOfRaw(store, last, x)
	if tt, ok := t.(*tupleType); ok {
		if len(tt.Elts) != 1 {
			panic(fmt.Sprintf(
				"evalStaticTypeOf() only supports *CallExpr with 1 result, got %s",
				tt.String(),
			))
		} else {
			return tt.Elts[0]
		}
	} else {
		return t
	}
}

// like evalStaticTypeOf() but returns the raw *tupleType for *CallExpr.
func evalStaticTypeOfRaw(store Store, last BlockNode, x Expr) (t Type) {
	if t, ok := x.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		return t
	} else if _, ok := x.(*constTypeExpr); ok {
		return gTypeType
	} else if ctx, ok := x.(*constExpr); ok {
		return ctx.T
	} else {
		pn := packageOf(last)
		t = NewMachine(pn.PkgPath, store).EvalStaticTypeOf(last, x)
		x.SetAttribute(ATTR_TYPEOF_VALUE, t)
		return t
	}
}

// If it is known that the type was already evaluated,
// use this function instead of evalStaticTypeOf().
func getTypeOf(x Expr) Type {
	if t, ok := x.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		if tt, ok := t.(*tupleType); ok {
			if len(tt.Elts) != 1 {
				panic(fmt.Sprintf(
					"getTypeOf() only supports *CallExpr with 1 result, got %s",
					tt.String(),
				))
			} else {
				return tt.Elts[0]
			}
		} else {
			return t
		}
	} else {
		panic(fmt.Sprintf(
			"getTypeOf() called on expr not yet evaluated with evalStaticTypeOf(): %s",
			x.String(),
		))
	}
}

// like evalStaticTypeOf() but for list of exprs, and the result
// includes the value if type is TypeKind.
func evalStaticTypedValues(store Store, last BlockNode, xs ...Expr) []TypedValue {
	res := make([]TypedValue, len(xs))
	for i, x := range xs {
		t := evalStaticTypeOf(store, last, x)
		if t != nil && t.Kind() == TypeKind {
			v := evalStaticType(store, last, x)
			res[i] = TypedValue{
				T: t,
				V: toTypeValue(v),
			}
		} else {
			res[i] = TypedValue{
				T: t,
				V: nil,
			}
		}
	}
	return res
}

func getGnoFuncTypeOf(it Type) *FuncType {
	bt := baseOf(it)
	ft := gnoTypeOf(bt).(*FuncType)
	return ft
}

func getResultTypedValues(cx *CallExpr) []TypedValue {
	if t, ok := cx.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		if tt, ok := t.(*tupleType); ok {
			res := make([]TypedValue, len(tt.Elts))
			for i, tte := range tt.Elts {
				res[i] = anyValue(tte)
			}
			return res
		} else {
			panic(fmt.Sprintf(
				"expected *tupleType of *CallExpr but got %v",
				reflect.TypeOf(t)))
		}
	} else {
		panic(fmt.Sprintf(
			"getResultTypedValues() called on call expr not yet evaluated: %s",
			cx.String(),
		))
	}
}

// Evaluate constant expressions.  Assumes all operands are
// already defined consts; the machine doesn't know whether a
// value is const or not, so this function always returns a
// *constExpr, even if the operands aren't actually consts in the
// code.
//
// No type conversion is done by the machine except as required by
// the expression (but otherwise the context is not considered).
// For example, untyped bigint types remain as untyped bigint
// types after evaluation.  Conversion happens in a separate step
// while leaving composite exprs/nodes that contain constant
// expression nodes (e.g. const exprs in the rhs of AssignStmts).
func evalConst(store Store, last BlockNode, x Expr) *constExpr {
	// TODO: some check or verification for ensuring x
	// is constant?  From the machine?
	pn := packageOf(last)
	cv := NewMachine(pn.PkgPath, store).EvalStatic(last, x)
	cx := &constExpr{
		Source:     x,
		TypedValue: cv,
	}
	cx.SetAttribute(ATTR_PREPROCESSED, true)
	setConstAttrs(cx)
	return cx
}

func constType(source Expr, t Type) *constTypeExpr {
	cx := &constTypeExpr{Source: source}
	cx.Type = t
	cx.SetAttribute(ATTR_PREPROCESSED, true)
	return cx
}

func setConstAttrs(cx *constExpr) {
	cv := &cx.TypedValue
	cx.SetAttribute(ATTR_TYPEOF_VALUE, cv.T)
	if cv.T != nil && cv.T.Kind() == TypeKind {
		if cv.GetType() == nil {
			panic("should not happen")
		}
		cx.SetAttribute(ATTR_TYPE_VALUE, cv.GetType())
	}
}

func packageOf(last BlockNode) *PackageNode {
	for {
		if pn, ok := last.(*PackageNode); ok {
			return pn
		}
		last = last.GetParentNode(nil)
	}
}

func funcNodeOf(last BlockNode) (BlockNode, *FuncTypeExpr) {
	for {
		if flx, ok := last.(*FuncLitExpr); ok {
			return flx, &flx.Type
		} else if fd, ok := last.(*FuncDecl); ok {
			return fd, &fd.Type
		}
		last = last.GetParentNode(nil)
	}
}

func findLabel(last BlockNode, label Name) (
	bn BlockNode, depth uint8, bodyIdx int) {

	for {
		switch cbn := last.(type) {
		case *IfStmt, *SwitchStmt:
			// These are faux blocks -- shouldn't happen.
			panic("unexpected faux blocknode")
		case *FuncLitExpr,
			*BlockStmt, *ForStmt, *IfCaseStmt, *RangeStmt,
			*SelectCaseStmt, *SwitchClauseStmt, *FuncDecl,
			*FileNode, *PackageNode:

			body := cbn.GetBody()
			_, bodyIdx = body.GetLabeledStmt(label)
			if bodyIdx != -1 {
				bn = cbn
				return
			} else {
				last = cbn.GetParentNode(nil)
				depth += 1
			}
		default:
			panic("unexpected block node")
		}
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
		V: toTypeValue(t),
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

func isConstType(x Expr) bool {
	_, ok := x.(*constTypeExpr)
	return ok
}

// 1. convert x to t if x is *constExpr.
// 2. otherwise, assert that x can be coerced to t.
// NOTE: also see checkOrConvertIntegerType()
func checkOrConvertType(store Store, last BlockNode, x Expr, t Type) {
	if cx, ok := x.(*constExpr); ok {
		convertConst(store, last, cx, t)
	} else if x != nil && t != nil {
		xt := evalStaticTypeOf(store, last, x)
		checkType(xt, t)
	}
}

// like checkOrConvertType(last, x, nil)
func convertIfConst(store Store, last BlockNode, x Expr) {
	if cx, ok := x.(*constExpr); ok {
		convertConst(store, last, cx, nil)
	}
}

func convertConst(store Store, last BlockNode, cx *constExpr, t Type) {
	if t != nil && t.Kind() == InterfaceKind {
		t = nil // signifies to convert to default type.
	}
	if isUntyped(cx.T) {
		ConvertUntypedTo(&cx.TypedValue, t)
		setConstAttrs(cx)
	} else if t != nil {
		// e.g. a named type or uint8 type to int for indexing.
		ConvertTo(store, &cx.TypedValue, t)
		setConstAttrs(cx)
	}
}

// assert that xt can be assigned as dt (dest type).
func checkType(xt Type, dt Type) {
	// Special case if dt is interface kind:
	if dt.Kind() == InterfaceKind {
		if idt, ok := baseOf(dt).(*InterfaceType); ok {
			if idt.IsEmptyInterface() {
				// if dt is an empty Gno interface, any x ok.
				return // ok
			} else if idt.IsImplementedBy(xt) {
				// if dt implements idt, ok.
				return // ok
			} else {
				panic(fmt.Sprintf(
					"%s does not implement %s",
					xt.String(),
					dt.String()))
			}
		} else if ndt, ok := baseOf(dt).(*nativeType); ok {
			nidt := ndt.Type
			if nidt.NumMethod() == 0 {
				// if dt is an empty Go native interface, ditto.
				return // ok
			} else if nxt, ok := baseOf(xt).(*nativeType); ok {
				// if xt has native base, do the naive native.
				if nxt.Type.AssignableTo(nidt) {
					return // ok
				} else {
					panic(fmt.Sprintf(
						"cannot use %s as %s",
						nxt.String(),
						nidt.String()))
				}
			} else if pxt, ok := baseOf(xt).(*PointerType); ok {
				nxt, ok := pxt.Elt.(*nativeType)
				if !ok {
					panic(fmt.Sprintf(
						"pointer to non-native type cannot satisfy non-empty native interface; %s doesn't implmeent %s",
						pxt.String(),
						nidt.String()))
				}
				// if xt has native base, do the naive native.
				if reflect.PtrTo(nxt.Type).AssignableTo(nidt) {
					return // ok
				} else {
					panic(fmt.Sprintf(
						"cannot use %s as %s",
						pxt.String(),
						nidt.String()))
				}
			} else {
				panic(fmt.Sprintf(
					"unexpected type pair: cannot use %s as %s",
					xt.String(),
					dt.String()))
			}
		} else {
			panic("should not happen")
		}
	}
	// Special case if xt or dt is *nativeType,
	// check with converted Gno type, and base of the other.
	if nxt, ok := xt.(*nativeType); ok {
		xt = go2GnoType2(nxt.Type)
		dt = baseOf(dt)
	} else if pxt, ok := xt.(*PointerType); ok {
		// *gonative{x} is gonative{*x}
		if enxt, ok := pxt.Elt.(*nativeType); ok {
			xt = &PointerType{Elt: go2GnoType2(enxt.Type)}
			dt = baseOf(dt)
		}
	}
	if nt, ok := dt.(*nativeType); ok {
		xt = baseOf(xt)
		dt = go2GnoType2(nt.Type)
	} else if pt, ok := dt.(*PointerType); ok {
		// *gonative{x} is gonative{*x}
		if ent, ok := pt.Elt.(*nativeType); ok {
			xt = baseOf(xt)
			dt = &PointerType{Elt: go2GnoType2(ent.Type)}
		}
	}
	// Special case of xt or dt is *DeclaredType,
	// allow implicit conversion unless both are declared.
	if dxt, ok := xt.(*DeclaredType); ok {
		if ddt, ok := dt.(*DeclaredType); ok {
			// types must match exactly.
			if !dxt.sealed && !ddt.sealed &&
				dxt.PkgPath == ddt.PkgPath &&
				dxt.Name == ddt.Name { // not yet sealed
				return // ok
			} else if dxt.TypeID() == ddt.TypeID() {
				return // ok
			} else {
				panic(fmt.Sprintf(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					ddt.String()))
			}
		} else {
			// carry on with baseOf(dxt)
			xt = dxt.Base
		}
	} else if ddt, ok := dt.(*DeclaredType); ok {
		// carry on with baseOf(ddt)
		dt = ddt.Base
	}
	bdt := baseOf(dt)
	// General cases.
	switch cxt := xt.(type) {
	case PrimitiveType:
		// if xt is untyped, ensure dt is compatible.
		switch xt {
		case UntypedBoolType:
			switch dt.Kind() {
			case BoolKind:
				return // ok
			default:
				panic(fmt.Sprintf(
					"cannot use untyped bool as %s",
					dt.Kind()))
			}
		case UntypedStringType:
			switch dt.Kind() {
			case StringKind:
				return // ok
			default:
				panic(fmt.Sprintf(
					"cannot use untyped string as %s",
					dt.Kind()))
			}
		case UntypedRuneType, UntypedBigintType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind:
				return // ok
			default:
				panic(fmt.Sprintf(
					"cannot use untyped rune as %s",
					dt.Kind()))
			}
		default:
			if isUntyped(xt) {
				panic("unexpected untyped type")
			}
			if cxt.TypeID() == bdt.TypeID() {
				return // ok
			}
		}
	case *PointerType:
		if pt, ok := bdt.(*PointerType); ok {
			checkType(cxt.Elt, pt.Elt)
			return // ok
		}
	case *ArrayType:
		if at, ok := bdt.(*ArrayType); ok {
			checkType(cxt.Elt, at.Elt)
			return // ok
		}
	case *SliceType:
		if st, ok := bdt.(*SliceType); ok {
			checkType(cxt.Elt, st.Elt)
			return // ok
		}
	case *MapType:
		if mt, ok := bdt.(*MapType); ok {
			checkType(cxt.Key, mt.Key)
			checkType(cxt.Value, mt.Value)
			return // ok
		}
	case *FuncType:
		if cxt.TypeID() == bdt.TypeID() {
			return // ok
		}
	case *InterfaceType:
		if cxt.Generic != "" {
			panic(fmt.Sprintf(
				"unexpected generic <%s>",
				cxt.Generic))
		} else {
			return // ok
		}
	case *DeclaredType:
		panic("should not happen")
	case *StructType, *PackageType, *ChanType:
		if cxt.TypeID() == bdt.TypeID() {
			return // ok
		}
	case *TypeType:
		if cxt.TypeID() == bdt.TypeID() {
			return // ok
		}
	default:
		panic(fmt.Sprintf(
			"unexpected type %s",
			xt.String()))
	}
	panic(fmt.Sprintf(
		"cannot use %s as %s",
		xt.String(),
		dt.String()))
}

// Returns any names not yet defined in expr.
// These happen upon enter from the top, so value paths cannot be
// used.  If no names are un and x is TypeExpr, evalStaticType(store,last, x)
// must not panic.
// NOTE: has no side effects except for the case of composite
// type expressions, which must get preprocessed for inner
// composite type eliding to work.
func findUndefined(store Store, last BlockNode, x Expr) (un Name) {
	return findUndefined2(store, last, x, nil)
}

func findUndefined2(store Store, last BlockNode, x Expr, t Type) (un Name) {
	if x == nil {
		return
	}
	switch cx := x.(type) {
	case *NameExpr:
		if _, ok := UverseNode().GetLocalIndex(cx.Name); ok {
			return
		}
		if tv := last.GetValueRef(store, cx.Name); tv != nil {
			return
		}
		return cx.Name
	case *BasicLitExpr:
		return
	case *BinaryExpr:
		un = findUndefined(store, last, cx.Left)
		if un != "" {
			return
		}
		un = findUndefined(store, last, cx.Right)
		if un != "" {
			return
		}
	case *SelectorExpr:
		return findUndefined(store, last, cx.X)
	case *SliceExpr:
		un = findUndefined(store, last, cx.X)
		if un != "" {
			return
		}
		if cx.Low != nil {
			un = findUndefined(store, last, cx.Low)
			if un != "" {
				return
			}
		}
		if cx.High != nil {
			un = findUndefined(store, last, cx.High)
			if un != "" {
				return
			}
		}
		if cx.Max != nil {
			un = findUndefined(store, last, cx.Max)
			if un != "" {
				return
			}
		}
	case *StarExpr:
		return findUndefined(store, last, cx.X)
	case *RefExpr:
		return findUndefined(store, last, cx.X)
	case *TypeAssertExpr:
		un = findUndefined(store, last, cx.X)
		if un != "" {
			return
		}
		return findUndefined(store, last, cx.Type)
	case *UnaryExpr:
		return findUndefined(store, last, cx.X)
	case *CompositeLitExpr:
		var ct Type
		if cx.Type == nil {
			if t == nil {
				panic("cannot elide unknown composite type")
			}
			ct = t
			cx.Type = constType(nil, t)
		} else {
			un = findUndefined(store, last, cx.Type)
			if un != "" {
				return
			}
			// preprocess now for eliding purposes.
			// TODO recursive preprocessing here is hacky, find a better
			// way.  This cannot be done asynchronously, cuz undefined
			// names ought to be returned immediately to let the caller
			// predefine it.
			cx.Type = Preprocess(store, last, cx.Type).(Expr) // recursive
			ct = evalStaticType(store, last, cx.Type)
			// elide composite lit element (nested) composite types.
			elideCompositeElements(cx, ct)
		}
		switch ct.Kind() {
		case ArrayKind, SliceKind, MapKind:
			for _, kvx := range cx.Elts {
				un = findUndefined(store, last, kvx.Key)
				if un != "" {
					return
				}
				un = findUndefined2(store, last, kvx.Value, ct.Elem())
				if un != "" {
					return
				}
			}
		case StructKind:
			for _, kvx := range cx.Elts {
				un = findUndefined(store, last, kvx.Value)
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
		return findUndefined(store, last, &cx.Type)
	case *FieldTypeExpr:
		return findUndefined(store, last, cx.Type)
	case *ArrayTypeExpr:
		if cx.Len != nil {
			un = findUndefined(store, last, cx.Len)
			if un != "" {
				return
			}
		}
		return findUndefined(store, last, cx.Elt)
	case *SliceTypeExpr:
		return findUndefined(store, last, cx.Elt)
	case *InterfaceTypeExpr:
		for i := range cx.Methods {
			un = findUndefined(store, last, &cx.Methods[i])
			if un != "" {
				return
			}
		}
	case *ChanTypeExpr:
		return findUndefined(store, last, cx.Value)
	case *FuncTypeExpr:
		for i := range cx.Params {
			un = findUndefined(store, last, &cx.Params[i])
			if un != "" {
				return
			}
		}
		for i := range cx.Results {
			un = findUndefined(store, last, &cx.Results[i])
			if un != "" {
				return
			}
		}
	case *MapTypeExpr:
		un = findUndefined(store, last, cx.Key)
		if un != "" {
			return
		}
		un = findUndefined(store, last, cx.Value)
		if un != "" {
			return
		}
	case *StructTypeExpr:
		for i := range cx.Fields {
			un = findUndefined(store, last, &cx.Fields[i])
			if un != "" {
				return
			}
		}
	case *CallExpr:
		un = findUndefined(store, last, cx.Func)
		if un != "" {
			return
		}
		for i := range cx.Args {
			un = findUndefined(store, last, cx.Args[i])
			if un != "" {
				return
			}
		}
	case *IndexExpr:
		un = findUndefined(store, last, cx.X)
		if un != "" {
			return
		}
		un = findUndefined(store, last, cx.Index)
		if un != "" {
			return
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

// like checkOrConvertType() but for any integer type.
func checkOrConvertIntegerType(store Store, last BlockNode, x Expr) {
	if cx, ok := x.(*constExpr); ok {
		convertConst(store, last, cx, IntType)
	} else if x != nil {
		xt := evalStaticTypeOf(store, last, x)
		checkIntegerType(xt)
	}
}

// assert that xt can be assigned as an integer type.
func checkIntegerType(xt Type) {
	switch xt.Kind() {
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind,
		BigintKind:
		return // ok
	default:
		panic(fmt.Sprintf(
			"expected integer type, but got %v",
			xt.Kind()))
	}
}

// predefineNow() pre-defines (with empty placeholders) all
// declaration names, and then preprocesses all type/value decls, and
// partially processes func decls.
//
// The recursive base procedure is split into two parts:
//
// First, tryPredefine(), which first predefines with placeholder
// values/types to support recursive types, then returns yet
// un-predefined dependencies.
//
// Second, which immediately preprocesses type/value declarations
// after dependencies have first been predefined, or partially
// preprocesses function declarations (which may not be completely
// preprocess-able before other file-level declarations are
// preprocessed).
func predefineNow(store Store, last BlockNode, d Decl) (Decl, bool) {
	m := make(map[Name]struct{})
	return predefineNow2(store, last, d, m)
}

func predefineNow2(store Store, last BlockNode, d Decl, m map[Name]struct{}) (Decl, bool) {
	pkg := packageOf(last)
	// pre-register d.GetName() to detect circular definition.
	for _, dn := range d.GetDeclNames() {
		m[dn] = struct{}{}
	}
	// recursively predefine dependencies.
	for {
		un := tryPredefine(store, last, d)
		if un != "" {
			// check circularity.
			if _, ok := m[un]; ok {
				panic("constant definition loop")
			}
			// look up dependency declaration from fileset.
			file, decl := pkg.FileSet.GetDeclFor(un)
			// preprocess if not already preprocessed.
			if file.GetParentNode(nil) == nil {
				file = Preprocess(store, pkg, file).(*FileNode)
			} else {
				// predefine dependency (recursive).
				*decl, _ = predefineNow2(store, file, *decl, m)
			}
		} else {
			break
		}
	}
	switch cd := d.(type) {
	case *FuncDecl:
		// *FuncValue/*FuncType is mostly empty still; here
		// we just fill the func type (and recv if method).
		// NOTE: unlike the *ValueDecl case, this case doesn't
		// preprocess d itself (only d.Type).
		if cd.IsMethod {
			if cd.Recv.Name == "" {
				cd.Recv.Name = "_"
			}
			cd.Recv =
				*Preprocess(store, last, &cd.Recv).(*FieldTypeExpr)
			cd.Type =
				*Preprocess(store, last, &cd.Type).(*FuncTypeExpr)
			rft := evalStaticType(store, last, &cd.Recv).(FieldType)
			rt := rft.Type
			ft := evalStaticType(store, last, &cd.Type).(*FuncType)
			ft = ft.UnboundType(rft)
			dt := (*DeclaredType)(nil)
			if pt, ok := rt.(*PointerType); ok {
				dt = pt.Elem().(*DeclaredType)
			} else {
				dt = rt.(*DeclaredType)
			}
			dt.DefineMethod(&FuncValue{
				Type:       ft,
				IsMethod:   true,
				SourceLoc:  cd.GetLocation(),
				Source:     cd,
				Name:       cd.Name,
				Body:       cd.Body,
				Closure:    nil, // set later, see PrepareNewValues().
				FileName:   filenameOf(last),
				PkgPath:    "", // set later, see PrepareNewValues().
				nativeBody: nil,
				pkg:        nil, // set later, see PrepareNewValues().
			})
		} else {
			ftv := pkg.GetValueRef(store, cd.Name)
			ft := ftv.T.(*FuncType)
			cd.Type = *Preprocess(store, last, &cd.Type).(*FuncTypeExpr)
			ft2 := evalStaticType(store, last, &cd.Type).(*FuncType)
			*ft = *ft2
			// XXX replace attr w/ ft?
			// return Preprocess(store, last, cd).(Decl), true
		}
		return d, false
	case *ValueDecl:
		return Preprocess(store, last, cd).(Decl), true
	case *TypeDecl:
		return Preprocess(store, last, cd).(Decl), true
	default:
		return d, false
	}
}

// If a dependent name is not yet defined, that name is
// returned; this return value is used by the caller to
// enforce declaration order.  If a dependent type is not yet
// defined (preprocessed), that type is fully preprocessed.
// Besides defining the type (and immediate dependent types
// of d) onto last (or packageOf(last)), there are no other
// side effects.  This function works for all block nodes and
// must be called for name declarations within (non-file,
// non-package) stmt bodies.
func tryPredefine(store Store, last BlockNode, d Decl) (un Name) {
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
		pv := store.GetPackage(d.PkgPath)
		if pv == nil {
			panic(fmt.Sprintf(
				"unknown import path %s",
				d.Path))
		}
		if d.Name == "" { // use default
			d.Name = pv.PkgName
		} else if d.Name == "_" { // no definition
			return
		}
		// NOTE imports usually must happen with a file,
		// and so last is usually a *FileNode, but for
		// testing convenience we allow importing
		// directly onto the package.
		last.Define(d.Name, TypedValue{
			T: gPackageType,
			V: pv,
		})
		d.Path = last.GetPathForName(store, d.Name)
	case *ValueDecl:
		un = findUndefined(store, last, d.Type)
		if un != "" {
			return
		}
		for _, vx := range d.Values {
			un = findUndefined(store, last, vx)
			if un != "" {
				return
			}
		}
		for i := 0; i < len(d.NameExprs); i++ {
			nx := &d.NameExprs[i]
			if nx.Name == "_" {
				nx.Path.Name = "_"
			} else {
				last2 := skipFile(last)
				last2.Define(nx.Name, anyValue(nil))
				nx.Path = last.GetPathForName(store, nx.Name)
			}
		}
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
					tv := Uverse().GetPointerTo(nil, NewValuePathUverse(idx, tx.Name))
					t = tv.TV.GetType()
				} else if tv := last.GetValueRef(store, tx.Name); tv != nil {
					// (file) block name
					t = tv.GetType()
					if dt, ok := t.(*DeclaredType); ok {
						if !dt.sealed {
							// predefineNow preprocessed dependent types.
							panic("should not happen")
						}
					} else {
						// all names are declared types.
						panic("should not happen")
					}
				} else {
					// yet undefined
					un = tx.Name
					return
				}
			case *SelectorExpr:
				// get package value.
				un = findUndefined(store, last, tx.X)
				if un != "" {
					return
				}
				pkgName := tx.X.(*NameExpr).Name
				tv := last.GetValueRef(store, pkgName)
				pv, ok := tv.V.(*PackageValue)
				if !ok {
					panic(fmt.Sprintf(
						"unknown package name %s in %s",
						pkgName,
						tx.String(),
					))
				}
				// check package node for name.
				pn := pv.Source.(*PackageNode)
				tx.Path = pn.GetPathForName(store, tx.Sel)
				ptr := pv.Block.GetPointerTo(store, tx.Path)
				t = ptr.TV.T
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
			d.Path = last.GetPathForName(store, d.Name)
		}
		// after predefinitions, return any undefined dependencies.
		un = findUndefined(store, last, d.Type)
		if un != "" {
			return
		}
	case *FuncDecl:
		un = findUndefined(store, last, &d.Type)
		if un != "" {
			return
		}
		if d.IsMethod {
			// define method.
			// methods are defined as struct fields, not
			// in the last block.  receiver isn't
			// processed until FuncDecl:BLOCK.
			un = findUndefined(store, last, &d.Recv)
			if un != "" {
				return
			}
		} else {
			// define package-level function.
			var ft = &FuncType{}
			pkg := skipFile(last).(*PackageNode)
			// special case: if d.Name == "init", assign unique suffix.
			if d.Name == "init" {
				idx := pkg.GetNumNames()
				// NOTE: use a dot for init func suffixing.
				// this also makes them unreferenceable.
				dname := Name(fmt.Sprintf("init.%d", idx))
				d.Name = dname
			}
			// define a FuncValue w/ above type as d.Name.
			// fill in later during *FuncDecl:BLOCK.
			pkg.Define(d.Name, TypedValue{
				T: ft,
				V: &FuncValue{
					Type:       ft,
					IsMethod:   false,
					SourceLoc:  d.GetLocation(),
					Source:     d,
					Name:       d.Name,
					Body:       d.Body,
					Closure:    nil, // set later, see PrepareNewValues().
					FileName:   filenameOf(last),
					PkgPath:    "", // set later, see PrepareNewValues().
					nativeBody: nil,
					pkg:        nil, // set later, see PrepareNewValues().
				},
			})
			if d.Name == "init" {
				// init functions can't be referenced.
			} else {
				d.Path = last.GetPathForName(store, d.Name)
			}
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
		nx.Path = last.GetPathForName(nil, nx.Name)
	}
}

func isFile(n BlockNode) bool {
	if _, ok := n.(*FileNode); ok {
		return true
	} else {
		return false
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
		case *PointerType:
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

// returns true of x is exactly `nil`.
func isNilExpr(x Expr) bool {
	if nx, ok := x.(*NameExpr); ok {
		return nx.Name == "nil"
	}
	return false
}

func isNilComparableKind(k Kind) bool {
	switch k {
	case SliceKind, MapKind, FuncKind:
		return true
	default:
		return false
	}
}

// returns number of args, or if arg is a call result,
// the number of results of the return tuple type.
func countNumArgs(store Store, last BlockNode, n *CallExpr) (numArgs int) {
	if len(n.Args) != 1 {
		return len(n.Args)
	} else if cx, ok := n.Args[0].(*CallExpr); ok {
		cxift := evalStaticTypeOf(store, last, cx.Func) // cx (iface) func type
		if cxift.Kind() == TypeKind {
			return 1 // type conversion
		} else {
			cxft := getGnoFuncTypeOf(cxift)
			numResults := len(cxft.Results)
			return numResults
		}
	} else {
		return 1
	}
}

func mergeNames(a, b []Name) []Name {
	c := make([]Name, len(a)+len(b))
	copy(c, a)
	copy(c[len(a):], b)
	return c
}

// This is to be run *after* preprocessing is done,
// to determine the order of var decl execution
// (which may include functions which may refer to package vars).
func findDependentNames(n Node, dst map[Name]struct{}) {
	switch cn := n.(type) {
	case *NameExpr:
		if _, ok := UverseNode().GetLocalIndex(cn.Name); ok {
			// skip global name
		} else {
			dst[cn.Name] = struct{}{}
		}
	case *BasicLitExpr:
	case *BinaryExpr:
		findDependentNames(cn.Left, dst)
		findDependentNames(cn.Right, dst)
	case *SelectorExpr:
		findDependentNames(cn.X, dst)
	case *SliceExpr:
		findDependentNames(cn.X, dst)
		if cn.Low != nil {
			findDependentNames(cn.Low, dst)
		}
		if cn.High != nil {
			findDependentNames(cn.High, dst)
		}
		if cn.Max != nil {
			findDependentNames(cn.Max, dst)
		}
	case *StarExpr:
		findDependentNames(cn.X, dst)
	case *RefExpr:
		findDependentNames(cn.X, dst)
	case *TypeAssertExpr:
		findDependentNames(cn.X, dst)
		findDependentNames(cn.Type, dst)
	case *UnaryExpr:
		findDependentNames(cn.X, dst)
	case *CompositeLitExpr:
		findDependentNames(cn.Type, dst)
		ct := getType(cn.Type)
		switch ct.Kind() {
		case ArrayKind, SliceKind, MapKind:
			for _, kvx := range cn.Elts {
				if kvx.Key != nil {
					findDependentNames(kvx.Key, dst)
				}
				findDependentNames(kvx.Value, dst)
			}
		case StructKind:
			for _, kvx := range cn.Elts {
				findDependentNames(kvx.Value, dst)
			}
		default:
			panic(fmt.Sprintf(
				"unexpected composite lit type %s",
				ct.String()))
		}
	case *FuncLitExpr:
		findDependentNames(&cn.Type, dst)
	case *FieldTypeExpr:
		findDependentNames(cn.Type, dst)
	case *ArrayTypeExpr:
		findDependentNames(cn.Elt, dst)
		if cn.Len != nil {
			findDependentNames(cn.Len, dst)
		}
	case *SliceTypeExpr:
		findDependentNames(cn.Elt, dst)
	case *InterfaceTypeExpr:
		for i := range cn.Methods {
			findDependentNames(&cn.Methods[i], dst)
		}
	case *ChanTypeExpr:
		findDependentNames(cn.Value, dst)
	case *FuncTypeExpr:
		for i := range cn.Params {
			findDependentNames(&cn.Params[i], dst)
		}
		for i := range cn.Results {
			findDependentNames(&cn.Results[i], dst)
		}
	case *MapTypeExpr:
		findDependentNames(cn.Key, dst)
		findDependentNames(cn.Value, dst)
	case *StructTypeExpr:
		for i := range cn.Fields {
			findDependentNames(&cn.Fields[i], dst)
		}
	case *CallExpr:
		findDependentNames(cn.Func, dst)
		for i := range cn.Args {
			findDependentNames(cn.Args[i], dst)
		}
	case *IndexExpr:
		findDependentNames(cn.X, dst)
		findDependentNames(cn.Index, dst)
	case *constTypeExpr:
	case *constExpr:
	case *ImportDecl:
	case *ValueDecl:
		if cn.Type != nil {
			findDependentNames(cn.Type, dst)
		}
		for _, vx := range cn.Values {
			findDependentNames(vx, dst)
		}
	case *TypeDecl:
		findDependentNames(cn.Type, dst)
	case *FuncDecl:
		findDependentNames(&cn.Type, dst)
		if cn.IsMethod {
			findDependentNames(&cn.Recv, dst)
			for _, n := range cn.GetExternNames() {
				dst[n] = struct{}{}
			}
		} else {
			for _, n := range cn.GetExternNames() {
				if n == cn.Name {
					// top-level function can
					// refer to itself without
					// depending on itself.
				} else {
					dst[n] = struct{}{}
				}
			}
		}
	default:
		panic(fmt.Sprintf(
			"unexpected node: %v (%v)",
			n, reflect.TypeOf(n)))
	}
}
