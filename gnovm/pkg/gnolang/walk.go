package gnolang

import "fmt"

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Helper functions for common node lists. They may be empty.

func walkIdentList(in Visitor, out Visitor, list []*NameExpr) {
	for _, x := range list {
		Walk(in, out, x)
	}
}

func walkExprList(in Visitor, out Visitor, list []Expr) {
	for _, x := range list {
		Walk(in, out, x)
	}
}

func walkStmtList(in Visitor, out Visitor, list []Stmt) {
	for _, x := range list {
		Walk(in, out, x)
	}
}

func walkDeclList(in Visitor, out Visitor, list []Decl) {
	for _, x := range list {
		Walk(in, out, x)
	}
}

// Walk traverses an AST in depth-first order: It starts by calling
// in.Visit(node); and out.Visit(node); node must not be nil. If the visitor w returned by
// in.Visit(node); and/or out.Visit(node); is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
func Walk(pre Visitor, post Visitor, node Node) {
	if pre != nil {
		if pre = pre.Visit(node); pre == nil && post == nil {
			return
		}
	}

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types pre ast.go)
	switch n := node.(type) {
	// todo we don't use comments pre Gno, at the moment
	// Comments and fields
	// case *Comment:
	// nothing to do
	// case *CommentGroup:
	// for _, c := range n.List {
	//    Walk(pre, c)
	// }

	case *FieldTypeExpr:
		//if n.Doc != nil {
		// Walk(pre, n.Doc)
		//}
		walkIdentList(pre, post, []*NameExpr{Nx(n.Name)})
		if n.Type != nil {
			Walk(pre, post, n.Type)
		}
		if n.Tag != nil {
			Walk(pre, post, n.Tag)
		}
		//if n.Comment != nil {
		// Walk(pre, n.Comment)
		//}

	// case *FieldList:
	// for _, f := range n.List {
	//    Walk(pre, f)
	// }

	// Expressions
	case *NameExpr, *BasicLitExpr:
	// // nothing to do

	case *SliceTypeExpr:
		if n.Elt != nil {
			Walk(pre, post, n.Elt)
		}

	case *FuncLitExpr:
		Walk(pre, post, &n.Type)

		for _, stmt := range n.Body {
			Walk(pre, post, stmt)
		}

	case *CompositeLitExpr:
		if n.Type != nil {
			Walk(pre, post, n.Type)
		}

		for i := range n.Elts {
			Walk(pre, post, &n.Elts[i])
		}

	//case *ParenExpr:
	// Walk(pre, n.X)
	//
	case *SelectorExpr:
		Walk(pre, post, n.X)
		Walk(pre, post, Nx(&n.Sel))

	case *IndexExpr:
		Walk(pre, post, n.X)
		Walk(pre, post, n.Index)

	// case *IndexListExpr:
	// Walk(pre, n.X)
	// for _, index := range n.Indices {
	//    Walk(pre, index)
	// }

	case *SliceExpr:
		Walk(pre, post, n.X)
		if n.Low != nil {
			Walk(pre, post, n.Low)
		}
		if n.High != nil {
			Walk(pre, post, n.High)
		}
		if n.Max != nil {
			Walk(pre, post, n.Max)
		}

	case *TypeAssertExpr:
		Walk(pre, post, n.X)
		if n.Type != nil {
			Walk(pre, post, n.Type)
		}

	case *CallExpr:
		Walk(pre, post, n.Func)
		walkExprList(pre, post, n.Args)

	case *StarExpr:
		Walk(pre, post, n.X)

	case *UnaryExpr:
		Walk(pre, post, n.X)

	case *BinaryExpr:
		Walk(pre, post, n.Left)
		Walk(pre, post, n.Right)

	case *KeyValueExpr:
		Walk(pre, post, n.Key)
		Walk(pre, post, n.Value)

	// Types
	case *ArrayTypeExpr:
		if n.Len != nil {
			Walk(pre, post, n.Len)
		}
		Walk(pre, post, n.Elt)

	case *StructTypeExpr:
		for i := range n.Fields {
			Walk(pre, post, &n.Fields[i])
		}

	case *FuncTypeExpr:
		if n.Params != nil {
			for i := range n.Params {
				Walk(pre, post, &n.Params[i])
			}
		}
		if n.Results != nil {
			for i := range n.Results {
				Walk(pre, post, &n.Results[i])
			}
		}

	case *InterfaceTypeExpr:
		for i := range n.Methods {
			Walk(pre, post, &n.Methods[i])
		}

	case *MapTypeExpr:
		Walk(pre, post, n.Key)
		Walk(pre, post, n.Value)

	// we do not support this
	//case *ChanType:
	// Walk(pre, n.Value)
	//

	// Statements
	//case *BadStmt:
	// // nothing to do
	//
	case *DeclStmt:
		for _, stmt := range n.Body {
			Walk(pre, post, stmt)
		}

	//case *EmptyStmt:
	// // nothing to do
	//
	//case *LabeledStmt:
	// Walk(pre, n.Label)
	// Walk(pre, n.Stmt)

	case *ExprStmt:
		Walk(pre, post, n.X)

	case *SendStmt:
		Walk(pre, post, n.Chan)
		Walk(pre, post, n.Value)

	case *IncDecStmt:
		Walk(pre, post, n.X)

	case *AssignStmt:
		walkExprList(pre, post, n.Lhs)
		walkExprList(pre, post, n.Rhs)

	case *GoStmt:
		Walk(pre, post, &n.Call)

	case *DeferStmt:
		Walk(pre, post, &n.Call)

	case *ReturnStmt:
		walkExprList(pre, post, n.Results)

	case *BranchStmt:
		if n.Label != "" {
			Walk(pre, post, Nx(n.Label))
		}

	case *BlockStmt:
		walkStmtList(pre, post, n.Body)

	case *IfStmt:
		if n.Init != nil {
			Walk(pre, post, n.Init)
		}
		Walk(pre, post, n.Cond)
		Walk(pre, post, n.Source)
		for _, stmt := range n.Else.Body {
			Walk(pre, post, stmt)
		}

	case *SwitchClauseStmt:
		walkExprList(pre, post, n.Cases)
		walkStmtList(pre, post, n.Body)

	case *SwitchStmt:
		if n.Init != nil {
			Walk(pre, post, n.Init)
		}
		if n.X != nil {
			Walk(pre, post, n.X)
		}
		walkStmtList(pre, post, n.bodyStmt.Body)

	// case *TypeSwitchStmt:
	// if n.Init != nil {
	//    Walk(pre, n.Init)
	// }
	// Walk(pre, n.Assign)
	// Walk(pre, n.Body)

	// case *CommClause:
	// if n.Comm != nil {
	//    Walk(pre, n.Comm)
	// }
	// walkStmtList(pre, n.Body)

	//case *SelectStmt:
	// Walk(pre, n.Body)
	//
	case *ForStmt:
		if n.Init != nil {
			Walk(pre, post, n.Init)
		}
		if n.Cond != nil {
			Walk(pre, post, n.Cond)
		}
		if n.Post != nil {
			Walk(pre, post, n.Post)
		}
		walkStmtList(pre, post, n.Body)

	case *RangeStmt:
		if n.Key != nil {
			Walk(pre, post, n.Key)
		}
		if n.Value != nil {
			Walk(pre, post, n.Value)
		}
		Walk(pre, post, n.X)
		walkStmtList(pre, post, n.Body)

	// Declarations
	case *ImportDecl:
		if n.Label != "" {
			Walk(pre, post, Nx(n.Label))
		}
		if n.Name != "" {
			Walk(pre, post, &n.NameExpr)
		}
		//Walk(pre, n.Path)
		//if n.Comment != nil {
		// Walk(pre, n.Comment)
		//}

	case *ValueDecl:
		//if n.Doc != nil {
		// Walk(pre, n.Doc)
		//}

		for i := range n.NameExprs {
			Walk(pre, post, &n.NameExprs[i])
		}

		if n.Type != nil {
			Walk(pre, post, n.Type)
		}
		walkExprList(pre, post, n.Values)
		//if n.Comment != nil {
		// Walk(pre, n.Comment)
		//}

	case *TypeDecl:
		//if n.Doc != nil {
		// Walk(pre, n.Doc)
		//}
		Walk(pre, post, &n.NameExpr)
		//if n.TypeParams != nil {
		// Walk(pre, n.TypeParams)
		//}
		Walk(pre, post, n.Type)
		//if n.Comment != nil {
		// Walk(pre, n.Comment)
		//}

	//case *BadDecl:
	// // nothing to do
	//
	//case *GenDecl:
	// if n.Doc != nil {
	//    Walk(pre, n.Doc)
	// }
	// for _, s := range n.Specs {
	//    Walk(pre, s)
	// }
	//
	case *FuncDecl:
		//if n.Doc != nil {
		// Walk(pre, n.Doc)
		//}
		if n.Recv.Name != "" {
			Walk(pre, post, &n.Recv)
		}
		Walk(pre, post, &n.NameExpr)
		Walk(pre, post, &n.Type)
		if n.Body != nil {
			walkStmtList(pre, post, n.Body)
		}
	case *RefExpr:
		Walk(pre, post, n.X)

	// Files and packages
	case *FileNode:
		//if n.Doc != nil {
		// Walk(pre, n.Doc)
		//}
		Walk(pre, post, Nx(n.Name))
		walkDeclList(pre, post, n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	// case *Package:
	// for _, f := range n.Files {
	//    Walk(pre, f)
	// }
	case *constTypeExpr:
		Walk(pre, post, n.Source)
	case *ConstExpr:
		Walk(pre, post, n.Source)

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T with value %+v", n, n))
	}

	if post != nil {
		if post = post.Visit(node); post == nil {
			return
		}
		post.Visit(nil)
	}
	if pre != nil {
		pre.Visit(nil)
	}
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f != nil && f(node) {
		return f
	}
	return nil
}

// Inspect traverses a syntax tree recursively, starting with root,
// and calling `pre` and `post` for each node as described below.
// It returns the syntax tree, possibly modified.
func Inspect(node Node, pre, post func(Node) bool) {
	Walk(inspector(pre), inspector(post), node)
}
