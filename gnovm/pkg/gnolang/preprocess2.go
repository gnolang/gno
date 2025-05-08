package gnolang

import (
	"fmt"
	"reflect"
)

type XRealmItem struct {
	Type string
	Location
}

// t: type, p: pkgpath, f: filename, l: line, c: column
func addXRealmItem(n Node, t string, p string, f string, l int, c int) {
	x, _ := n.GetAttribute("XREALMITEM").(map[string]string) // p/f:l:c -> t
	if x == nil {
		x = make(map[string]string)
		n.SetAttribute("XREALMITEM", x)
	}
	key := fmt.Sprintf("%s/%s:%d:%d", p, f, l, c)
	x[key] = t
}

// Finds XRealmItems for interream spec 2 transpiling tool.
// Sets to bn attribute XREALMITEM.
func FindXRealmItems(store Store, pn *PackageNode, bn BlockNode) {
	// create stack of BlockNodes.
	var stack []BlockNode = make([]BlockNode, 0, 32)
	var last BlockNode = pn
	stack = append(stack, last)

	// Iterate over all nodes recursively.
	_ = Transcribe(bn, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		defer doRecover(stack, n)

		if debug {
			debug.Printf("FindXRealmItems %s (%v) stage:%v\n", n.String(), reflect.TypeOf(n), stage)
		}

		switch stage {
		// ----------------------------------------
		case TRANS_BLOCK:
			pushInitBlock(n.(BlockNode), &last, &stack)

		// ----------------------------------------
		case TRANS_LEAVE:
			// Pop block from stack.
			// NOTE: DO NOT USE TRANS_SKIP WITHIN BLOCK
			// NODES, AS TRANS_LEAVE WILL BE SKIPPED; OR
			// POP BLOCK YOURSELF.
			defer func() {
				switch n.(type) {
				case BlockNode:
					stack = stack[:len(stack)-1]
					last = stack[len(stack)-1]
				}
			}()

			switch n := n.(type) {
			case *CallExpr:
				if _, ok := n.Func.(*constTypeExpr); ok {
					fmt.Println("CONOST1", n)
					return n, TRANS_CONTINUE
				} else if cx, ok := n.Func.(*ConstExpr); ok {
					fmt.Println("CONOST2", n)
					if cx.TypedValue.T.Kind() != FuncKind {
						return n, TRANS_CONTINUE
					}
					fv := cx.GetFunc()
					if fv.PkgPath == uversePkgPath && fv.Name == "cross" {
						// XXX Add a 'nilrealm' as first argument.
						// This is not part of the proposed spec,
						// but Go doesn't support generic cross[T any]()T
						// that curries the first `cur realm` argument.
						// Gno2 can simply omit the `nilrealm` argument.
						pc, ok := ns[len(ns)-1].(*CallExpr)
						if !ok {
							panic("cross(fn) must be followed by a call")
						}
						loc := last.GetLocation()
						fmt.Printf("add nilrealm: %s/%s:%d:%d\n", loc.PkgPath, loc.File, pc.GetLine(), pc.GetColumn())
						addXRealmItem(pn, "add nilrealm", loc.PkgPath, loc.File, pc.GetLine(), pc.GetColumn())
					} else if fv.PkgPath == uversePkgPath && fv.Name == "crossing" {
						if !IsRealmPath(pn.PkgPath) {
							panic("crossing() is only allowed in realm packages")
						}
						// XXX Add `cur realm` as first argument to func decl.
						loc := last.GetLocation()
						fmt.Printf("add curfunc: %s/%s:%d:%d\n", loc.PkgPath, loc.File, loc.Line, loc.Column)
						addXRealmItem(pn, "add curfunc", loc.PkgPath, loc.File, loc.Line, loc.Column)
					} else if fv.PkgPath == uversePkgPath && fv.Name == "attach" {
						// reserve attach() so we can support it later.
						panic("attach() not yet supported")
					}
				} else {
					fmt.Println("CONOST3", n)
					// Already handled, added "add nilrealm"
					// from the "cross" case above.
					if n.WithCross {
						fmt.Println("CONOST4", n)
						// Is a cross(fn)(...) call.
						// Leave it alone.
						return n, TRANS_CONTINUE
					}
					pv := pn.NewPackage() // temporary
					store := store.BeginTransaction(nil, nil, nil)
					store.SetCachePackage(pv)
					m := NewMachine("x", store)
					defer m.Release()
					tv := TypedValue{}
					func() {
						// cannot be resolved statically
						defer func() {
							recover()
							//fmt.Println("FAILED TO EVALSTATIC", n.Func, r)
						}()
						// try to evaluate n.Func.
						tv = m.EvalStatic(last, n.Func)
					}()
					switch cv := tv.V.(type) {
					case nil:
						return n, TRANS_CONTINUE
					case TypeValue:
						panic("wtf")
					case *FuncValue:
						if cv.IsCrossing() {
							// XXX Not cross-called, so add `cur` as first argument.
							loc := last.GetLocation()
							fmt.Printf("add curcall: %s/%s:%d:%d\n", loc.PkgPath, loc.File, n.GetLine(), n.GetColumn())
							addXRealmItem(pn, "add curcall", loc.PkgPath, loc.File, n.GetLine(), n.GetColumn())
						}
					case *BoundMethodValue:
						if cv.IsCrossing() {
							// XXX Not cross-called, so add `cur` as first argument.
							loc := last.GetLocation()
							fmt.Printf("add curcall: %s/%s:%d:%d\n", loc.PkgPath, loc.File, n.GetLine(), n.GetColumn())
							addXRealmItem(pn, "add curcall", loc.PkgPath, loc.File, n.GetLine(), n.GetColumn())
						}
					}
				}
			}
			// end type switch statement
			// END TRANS_LEAVE -----------------------
			return n, TRANS_CONTINUE
		}
		return n, TRANS_CONTINUE
	})
}
