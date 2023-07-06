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
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
func Walk(in Visitor, out Visitor, node Node) {
	if in != nil {
		if in = in.Visit(node); in == nil && out == nil {
			return
		}
	}

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// todo we don't use comments in Gno, at the moment
	// Comments and fields
	// case *Comment:
	// nothing to do
	// case *CommentGroup:
	//	for _, c := range n.List {
	//		Walk(in, c)
	//	}

	case *FieldTypeExpr:
		//if n.Doc != nil {
		//	Walk(in, n.Doc)
		//}
		walkIdentList(in, out, []*NameExpr{Nx(n.Name)})
		if n.Type != nil {
			Walk(in, out, n.Type)
		}
		if n.Tag != nil {
			Walk(in, out, n.Tag)
		}
		//if n.Comment != nil {
		//	Walk(in, n.Comment)
		//}

	// case *FieldList:
	//	for _, f := range n.List {
	//		Walk(in, f)
	//	}

	// Expressions
	case *NameExpr, *BasicLitExpr:
	//	// nothing to do

	case *SliceTypeExpr:
		if n.Elt != nil {
			Walk(in, out, n.Elt)
		}

	case *FuncLitExpr:
		Walk(in, out, &n.Type)

		for _, stmt := range n.Body {
			Walk(in, out, stmt)
		}

	case *CompositeLitExpr:
		if n.Type != nil {
			Walk(in, out, n.Type)
		}

		for i := range n.Elts {
			Walk(in, out, &n.Elts[i])
		}

	//case *ParenExpr:
	//	Walk(in, n.X)
	//
	case *SelectorExpr:
		Walk(in, out, n.X)
		Walk(in, out, Nx(&n.Sel))

	case *IndexExpr:
		Walk(in, out, n.X)
		Walk(in, out, n.Index)

	// case *IndexListExpr:
	//	Walk(in, n.X)
	//	for _, index := range n.Indices {
	//		Walk(in, index)
	//	}

	case *SliceExpr:
		Walk(in, out, n.X)
		if n.Low != nil {
			Walk(in, out, n.Low)
		}
		if n.High != nil {
			Walk(in, out, n.High)
		}
		if n.Max != nil {
			Walk(in, out, n.Max)
		}

	case *TypeAssertExpr:
		Walk(in, out, n.X)
		if n.Type != nil {
			Walk(in, out, n.Type)
		}

	case *CallExpr:
		Walk(in, out, n.Func)
		walkExprList(in, out, n.Args)

	case *StarExpr:
		Walk(in, out, n.X)

	case *UnaryExpr:
		Walk(in, out, n.X)

	case *BinaryExpr:
		Walk(in, out, n.Left)
		Walk(in, out, n.Right)

	case *KeyValueExpr:
		Walk(in, out, n.Key)
		Walk(in, out, n.Value)

	// Types
	case *ArrayTypeExpr:
		if n.Len != nil {
			Walk(in, out, n.Len)
		}
		Walk(in, out, n.Elt)

	case *StructTypeExpr:
		for i := range n.Fields {
			Walk(in, out, &n.Fields[i])
		}

	case *FuncTypeExpr:
		if n.Params != nil {
			for i := range n.Params {
				Walk(in, out, &n.Params[i])
			}
		}
		if n.Results != nil {
			for i := range n.Results {
				Walk(in, out, &n.Results[i])
			}
		}

	case *InterfaceTypeExpr:
		for i := range n.Methods {
			Walk(in, out, &n.Methods[i])
		}

	case *MapTypeExpr:
		Walk(in, out, n.Key)
		Walk(in, out, n.Value)

	// we do not support this
	//case *ChanType:
	//	Walk(in, n.Value)
	//

	// Statements
	//case *BadStmt:
	//	// nothing to do
	//
	case *DeclStmt:
		for _, stmt := range n.Body {
			Walk(in, out, stmt)
		}

	//case *EmptyStmt:
	//	// nothing to do
	//
	//case *LabeledStmt:
	//	Walk(in, n.Label)
	//	Walk(in, n.Stmt)

	case *ExprStmt:
		Walk(in, out, n.X)

	case *SendStmt:
		Walk(in, out, n.Chan)
		Walk(in, out, n.Value)

	case *IncDecStmt:
		Walk(in, out, n.X)

	case *AssignStmt:
		walkExprList(in, out, n.Lhs)
		walkExprList(in, out, n.Rhs)

	case *GoStmt:
		Walk(in, out, &n.Call)

	case *DeferStmt:
		Walk(in, out, &n.Call)

	case *ReturnStmt:
		walkExprList(in, out, n.Results)

	case *BranchStmt:
		if n.Label != "" {
			Walk(in, out, Nx(n.Label))
		}

	case *BlockStmt:
		walkStmtList(in, out, n.Body)

	case *IfStmt:
		if n.Init != nil {
			Walk(in, out, n.Init)
		}
		Walk(in, out, n.Cond)
		Walk(in, out, n.Source)
		for _, stmt := range n.Else.Body {
			Walk(in, out, stmt)
		}

	case *SwitchClauseStmt:
		walkExprList(in, out, n.Cases)
		walkStmtList(in, out, n.Body)

	case *SwitchStmt:
		if n.Init != nil {
			Walk(in, out, n.Init)
		}
		if n.X != nil {
			Walk(in, out, n.X)
		}
		walkStmtList(in, out, n.bodyStmt.Body)

	// case *TypeSwitchStmt:
	//	if n.Init != nil {
	//		Walk(in, n.Init)
	//	}
	//	Walk(in, n.Assign)
	//	Walk(in, n.Body)

	// case *CommClause:
	//	if n.Comm != nil {
	//		Walk(in, n.Comm)
	//	}
	//	walkStmtList(in, n.Body)

	//case *SelectStmt:
	//	Walk(in, n.Body)
	//
	case *ForStmt:
		if n.Init != nil {
			Walk(in, out, n.Init)
		}
		if n.Cond != nil {
			Walk(in, out, n.Cond)
		}
		if n.Post != nil {
			Walk(in, out, n.Post)
		}
		walkStmtList(in, out, n.Body)

	case *RangeStmt:
		if n.Key != nil {
			Walk(in, out, n.Key)
		}
		if n.Value != nil {
			Walk(in, out, n.Value)
		}
		Walk(in, out, n.X)
		walkStmtList(in, out, n.Body)

	// Declarations
	case *ImportDecl:
		if n.Label != "" {
			Walk(in, out, Nx(n.Label))
		}
		if n.Name != "" {
			Walk(in, out, &n.NameExpr)
		}
		//Walk(in, n.Path)
		//if n.Comment != nil {
		//	Walk(in, n.Comment)
		//}

	case *ValueDecl:
		//if n.Doc != nil {
		//	Walk(in, n.Doc)
		//}

		for i := range n.NameExprs {
			Walk(in, out, &n.NameExprs[i])
		}

		if n.Type != nil {
			Walk(in, out, n.Type)
		}
		walkExprList(in, out, n.Values)
		//if n.Comment != nil {
		//	Walk(in, n.Comment)
		//}

	case *TypeDecl:
		//if n.Doc != nil {
		//	Walk(in, n.Doc)
		//}
		Walk(in, out, &n.NameExpr)
		//if n.TypeParams != nil {
		//	Walk(in, n.TypeParams)
		//}
		Walk(in, out, n.Type)
		//if n.Comment != nil {
		//	Walk(in, n.Comment)
		//}

	//case *BadDecl:
	//	// nothing to do
	//
	//case *GenDecl:
	//	if n.Doc != nil {
	//		Walk(in, n.Doc)
	//	}
	//	for _, s := range n.Specs {
	//		Walk(in, s)
	//	}
	//
	case *FuncDecl:
		//if n.Doc != nil {
		//	Walk(in, n.Doc)
		//}
		if n.Recv.Name != "" {
			Walk(in, out, &n.Recv)
		}
		Walk(in, out, &n.NameExpr)
		Walk(in, out, &n.Type)
		if n.Body != nil {
			walkStmtList(in, out, n.Body)
		}
	case *RefExpr:
		Walk(in, out, n.X)

	// Files and packages
	case *FileNode:
		//if n.Doc != nil {
		//	Walk(in, n.Doc)
		//}
		Walk(in, out, Nx(n.Name))
		walkDeclList(in, out, n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	// case *Package:
	//	for _, f := range n.Files {
	//		Walk(in, f)
	//	}

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	if out != nil {
		if out = out.Visit(node); out == nil {
			return
		}
		out.Visit(nil)
	}
	if in != nil {
		in.Visit(nil)
	}
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f != nil && f(node) {
		return f
	}
	return nil
}

// Inspect traverses an AST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).
func Inspect(node Node, in, out func(Node) bool) {
	Walk(inspector(in), inspector(out), node)
}
