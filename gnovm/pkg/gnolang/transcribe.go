package gnolang

import (
	"fmt"
	"reflect"
)

type (
	TransCtrl  uint8
	TransStage uint8
	TransField uint8
)

const (
	TRANS_CONTINUE TransCtrl = iota
	TRANS_SKIP
	TRANS_EXIT
)

const (
	TRANS_ENTER TransStage = iota
	TRANS_BLOCK
	TRANS_BLOCK2
	TRANS_LEAVE
)

const (
	TRANS_ROOT TransField = iota
	TRANS_BINARY_LEFT
	TRANS_BINARY_RIGHT
	TRANS_CALL_FUNC
	TRANS_CALL_ARG
	TRANS_INDEX_X
	TRANS_INDEX_INDEX
	TRANS_SELECTOR_X
	TRANS_SLICE_X
	TRANS_SLICE_LOW
	TRANS_SLICE_HIGH
	TRANS_SLICE_MAX
	TRANS_STAR_X
	TRANS_REF_X
	TRANS_TYPEASSERT_X
	TRANS_TYPEASSERT_TYPE
	TRANS_UNARY_X
	TRANS_COMPOSITE_TYPE
	TRANS_COMPOSITE_KEY
	TRANS_COMPOSITE_VALUE
	TRANS_FUNCLIT_TYPE
	TRANS_FUNCLIT_HEAP_CAPTURE
	TRANS_FUNCLIT_BODY
	TRANS_FIELDTYPE_NAME
	TRANS_FIELDTYPE_TYPE
	TRANS_FIELDTYPE_TAG
	TRANS_ARRAYTYPE_LEN
	TRANS_ARRAYTYPE_ELT
	TRANS_SLICETYPE_ELT
	TRANS_INTERFACETYPE_METHOD
	TRANS_CHANTYPE_VALUE
	TRANS_FUNCTYPE_PARAM
	TRANS_FUNCTYPE_RESULT
	TRANS_MAPTYPE_KEY
	TRANS_MAPTYPE_VALUE
	TRANS_STRUCTTYPE_FIELD
	TRANS_ASSIGN_LHS
	TRANS_ASSIGN_RHS
	TRANS_BLOCK_BODY
	TRANS_DECL_BODY
	TRANS_DEFER_CALL
	TRANS_EXPR_X
	TRANS_FOR_INIT
	TRANS_FOR_COND
	TRANS_FOR_POST
	TRANS_FOR_BODY
	TRANS_GO_CALL
	TRANS_IF_INIT
	TRANS_IF_COND
	TRANS_IF_BODY
	TRANS_IF_ELSE
	TRANS_IF_CASE_BODY
	TRANS_INCDEC_X
	TRANS_RANGE_X
	TRANS_RANGE_KEY
	TRANS_RANGE_VALUE
	TRANS_RANGE_BODY
	TRANS_RETURN_RESULT
	TRANS_SELECT_CASE
	TRANS_SELECTCASE_COMM
	TRANS_SELECTCASE_BODY
	TRANS_SEND_CHAN
	TRANS_SEND_VALUE
	TRANS_SWITCH_INIT
	TRANS_SWITCH_X
	TRANS_SWITCH_CASE
	TRANS_SWITCHCASE_CASE
	TRANS_SWITCHCASE_BODY
	TRANS_FUNC_RECV
	TRANS_FUNC_TYPE
	TRANS_FUNC_BODY
	TRANS_IMPORT_PATH
	TRANS_CONST_TYPE
	TRANS_CONST_VALUE
	TRANS_VAR_NAME
	TRANS_VAR_TYPE
	TRANS_VAR_VALUE
	TRANS_TYPE_TYPE
	TRANS_FILE_BODY
)

// Transform node `n` of `ftype`/`index` in context `ns` during `stage`.
// Return a new node to replace the old one, or the node will be deleted (or
// set to nil).  NOTE: Do not mutate stack or ns.  NOTE: Consider using
// TranscribeB() which is safer.
//
// Returns:
//   - TRANS_CONTINUE to visit children recursively;
//   - TRANS_SKIP to skip the following stages for the node
//     (BLOCK/BLOCK2/LEAVE), but a skip from LEAVE will
//     skip the following stages of the parent.
//   - TRANS_EXIT to stop traversing altogether.
//
// XXX Replace usage of Transcribe() with TranscribeB().
type Transform func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl)

// n MUST be a pointer to a node struct.
// returns the transcribe code returned for n.
// returns new node nn to replace n.
func Transcribe(n Node, t Transform) (nn Node) {
	if reflect.TypeOf(n).Kind() != reflect.Ptr {
		panic("Transcribe() expects a non-pointer concrete Node struct")
	}
	ns := make([]Node, 0, 32)
	var nc TransCtrl
	nn = transcribe(t, ns, TRANS_ROOT, 0, n, &nc)
	return
}

func transcribe(t Transform, ns []Node, ftype TransField, index int, n Node, nc *TransCtrl) (nn Node) {
	// transcribe n on the way in.
	var c TransCtrl
	nn, c = t(ns, ftype, index, n, TRANS_ENTER)
	if stopOrSkip(nc, c) {
		return
	}

	// push nn to node stack.
	nns := append(ns, nn)

	// visit any children of n.
	switch cnn := nn.(type) {
	case *NameExpr:
	case *BasicLitExpr:
	case *BinaryExpr:
		cnn.Left = transcribe(t, nns, TRANS_BINARY_LEFT, 0, cnn.Left, &c).(Expr) // XXX wished this worked with nil.
		if stopOrSkip(nc, c) {
			return
		}
		cnn.Right = transcribe(t, nns, TRANS_BINARY_RIGHT, 0, cnn.Right, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *CallExpr:
		cnn.Func = transcribe(t, nns, TRANS_CALL_FUNC, 0, cnn.Func, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		for idx := range cnn.Args {
			cnn.Args[idx] = transcribe(t, nns, TRANS_CALL_ARG, idx, cnn.Args[idx], &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *IndexExpr:
		cnn.X = transcribe(t, nns, TRANS_INDEX_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		cnn.Index = transcribe(t, nns, TRANS_INDEX_INDEX, 0, cnn.Index, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *SelectorExpr:
		cnn.X = transcribe(t, nns, TRANS_SELECTOR_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *SliceExpr:
		cnn.X = transcribe(t, nns, TRANS_SLICE_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		if cnn.Low != nil {
			cnn.Low = transcribe(t, nns, TRANS_SLICE_LOW, 0, cnn.Low, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		if cnn.High != nil {
			cnn.High = transcribe(t, nns, TRANS_SLICE_HIGH, 0, cnn.High, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		if cnn.Max != nil {
			cnn.Max = transcribe(t, nns, TRANS_SLICE_MAX, 0, cnn.Max, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *StarExpr:
		cnn.X = transcribe(t, nns, TRANS_STAR_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *RefExpr:
		cnn.X = transcribe(t, nns, TRANS_REF_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *TypeAssertExpr:
		cnn.X = transcribe(t, nns, TRANS_TYPEASSERT_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		if cnn.Type != nil {
			cnn.Type = transcribe(t, nns, TRANS_TYPEASSERT_TYPE, 0, cnn.Type, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *UnaryExpr:
		cnn.X = transcribe(t, nns, TRANS_UNARY_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *CompositeLitExpr:
		if cnn.Type != nil {
			cnn.Type = transcribe(t, nns, TRANS_COMPOSITE_TYPE, 0, cnn.Type, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		for idx, kvx := range cnn.Elts {
			k, v := kvx.Key, kvx.Value
			if k != nil {
				k = transcribe(t, nns, TRANS_COMPOSITE_KEY, idx, k, &c).(Expr)
				if stopOrSkip(nc, c) {
					return
				}
			}
			v = transcribe(t, nns, TRANS_COMPOSITE_VALUE, idx, v, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
			cnn.Elts[idx] = KeyValueExpr{Key: k, Value: v}
		}
	case *FuncLitExpr:
		cnn.Type = *transcribe(t, nns, TRANS_FUNCLIT_TYPE, 0, &cnn.Type, &c).(*FuncTypeExpr)
		if stopOrSkip(nc, c) {
			return
		}
		for idx := range cnn.HeapCaptures {
			cnn.HeapCaptures[idx] = *(transcribe(t, nns, TRANS_FUNCLIT_HEAP_CAPTURE, idx, &cnn.HeapCaptures[idx], &c).(*NameExpr))
			if stopOrSkip(nc, c) {
				return
			}
		}
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*FuncLitExpr)
		}
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_FUNCLIT_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *FieldTypeExpr:
		/* XXX make this an option. these are not normal names.
		cnn.NameExpr = *(transcribe(t, nns, TRANS_FIELDTYPE_NAME, 0, &cnn.NameExpr, &c).(*NameExpr))
		*/
		cnn.Type = transcribe(t, nns, TRANS_FIELDTYPE_TYPE, 0, cnn.Type, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		if cnn.Tag != nil {
			cnn.Tag = transcribe(t, nns, TRANS_FIELDTYPE_TAG, 0, cnn.Tag, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *ArrayTypeExpr:
		if cnn.Len != nil {
			cnn.Len = transcribe(t, nns, TRANS_ARRAYTYPE_LEN, 0, cnn.Len, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		cnn.Elt = transcribe(t, nns, TRANS_ARRAYTYPE_ELT, 0, cnn.Elt, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *SliceTypeExpr:
		cnn.Elt = transcribe(t, nns, TRANS_SLICETYPE_ELT, 0, cnn.Elt, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *InterfaceTypeExpr:
		for idx := range cnn.Methods {
			cnn.Methods[idx] = *transcribe(t, nns, TRANS_INTERFACETYPE_METHOD, idx, &cnn.Methods[idx], &c).(*FieldTypeExpr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *ChanTypeExpr:
		cnn.Value = transcribe(t, nns, TRANS_CHANTYPE_VALUE, 0, cnn.Value, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *FuncTypeExpr:
		for idx := range cnn.Params {
			cnn.Params[idx] = *transcribe(t, nns, TRANS_FUNCTYPE_PARAM, idx, &cnn.Params[idx], &c).(*FieldTypeExpr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		for idx := range cnn.Results {
			cnn.Results[idx] = *transcribe(t, nns, TRANS_FUNCTYPE_RESULT, idx, &cnn.Results[idx], &c).(*FieldTypeExpr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *MapTypeExpr:
		cnn.Key = transcribe(t, nns, TRANS_MAPTYPE_KEY, 0, cnn.Key, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		cnn.Value = transcribe(t, nns, TRANS_MAPTYPE_VALUE, 0, cnn.Value, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *StructTypeExpr:
		for idx := range cnn.Fields {
			cnn.Fields[idx] = *transcribe(t, nns, TRANS_STRUCTTYPE_FIELD, idx, &cnn.Fields[idx], &c).(*FieldTypeExpr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *AssignStmt:
		for idx := range cnn.Lhs {
			cnn.Lhs[idx] = transcribe(t, nns, TRANS_ASSIGN_LHS, idx, cnn.Lhs[idx], &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		for idx := range cnn.Rhs {
			cnn.Rhs[idx] = transcribe(t, nns, TRANS_ASSIGN_RHS, idx, cnn.Rhs[idx], &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *BlockStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*BlockStmt)
		}
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_BLOCK_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *BranchStmt:
	case *DeclStmt:
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_DECL_BODY, idx, cnn.Body[idx], &c).(SimpleDeclStmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *DeferStmt:
		cnn.Call = *transcribe(t, nns, TRANS_DEFER_CALL, 0, &cnn.Call, &c).(*CallExpr)
		if stopOrSkip(nc, c) {
			return
		}
	case *EmptyStmt:
	case *ExprStmt:
		cnn.X = transcribe(t, nns, TRANS_EXPR_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *ForStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*ForStmt)
		}
		if cnn.Init != nil {
			cnn.Init = transcribe(t, nns, TRANS_FOR_INIT, 0, cnn.Init, &c).(SimpleStmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
		if cnn.Cond != nil {
			cnn.Cond = transcribe(t, nns, TRANS_FOR_COND, 0, cnn.Cond, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		if cnn.Post != nil {
			cnn.Post = transcribe(t, nns, TRANS_FOR_POST, 0, cnn.Post, &c).(SimpleStmt)
			if stopOrSkip(nc, c) {
				return
			}
		}

		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_FOR_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *GoStmt:
		cnn.Call = *transcribe(t, nns, TRANS_GO_CALL, 0, &cnn.Call, &c).(*CallExpr)
		if stopOrSkip(nc, c) {
			return
		}
	case *IfStmt:
		// NOTE: like switch stmts, both if statements AND
		// contained cases visit with the TRANS_BLOCK stage, even
		// though during runtime only one block is created.
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*IfStmt)
		}
		if cnn.Init != nil {
			cnn.Init = transcribe(t, nns, TRANS_IF_INIT, 0, cnn.Init, &c).(SimpleStmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
		cnn.Cond = transcribe(t, nns, TRANS_IF_COND, 0, cnn.Cond, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		cnn.Then = *transcribe(t, nns, TRANS_IF_BODY, 0, &cnn.Then, &c).(*IfCaseStmt)
		if stopOrSkip(nc, c) {
			return
		}
		cnn.Else = *transcribe(t, nns, TRANS_IF_ELSE, 0, &cnn.Else, &c).(*IfCaseStmt)
		if stopOrSkip(nc, c) {
			return
		}
	case *IfCaseStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*IfCaseStmt)
		}
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_IF_CASE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *IncDecStmt:
		cnn.X = transcribe(t, nns, TRANS_INCDEC_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *RangeStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*RangeStmt)
		}
		cnn.X = transcribe(t, nns, TRANS_RANGE_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		if cnn.Key != nil {
			cnn.Key = transcribe(t, nns, TRANS_RANGE_KEY, 0, cnn.Key, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		if cnn.Value != nil {
			cnn.Value = transcribe(t, nns, TRANS_RANGE_VALUE, 0, cnn.Value, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_RANGE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *ReturnStmt:
		for idx := range cnn.Results {
			cnn.Results[idx] = transcribe(t, nns, TRANS_RETURN_RESULT, idx, cnn.Results[idx], &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *SelectStmt:
		for idx := range cnn.Cases {
			cnn.Cases[idx] = *transcribe(t, nns, TRANS_SELECT_CASE, idx, &cnn.Cases[idx], &c).(*SelectCaseStmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *SelectCaseStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*SelectCaseStmt)
		}
		cnn.Comm = transcribe(t, nns, TRANS_SELECTCASE_COMM, 0, cnn.Comm, &c).(Stmt)
		if stopOrSkip(nc, c) {
			return
		}
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_SELECTCASE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *SendStmt:
		cnn.Chan = transcribe(t, nns, TRANS_SEND_CHAN, 0, cnn.Chan, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		cnn.Value = transcribe(t, nns, TRANS_SEND_VALUE, 0, cnn.Value, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *SwitchStmt:
		// NOTE: unlike the select case, and like if stmts, both
		// switch statements AND contained cases visit with the
		// TRANS_BLOCK stage, even though during runtime only one
		// block is created.
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*SwitchStmt)
		}
		if cnn.Init != nil {
			cnn.Init = transcribe(t, nns, TRANS_SWITCH_INIT, 0, cnn.Init, &c).(SimpleStmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
		cnn.X = transcribe(t, nns, TRANS_SWITCH_X, 0, cnn.X, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
		// NOTE: special block case for after .Init and .X.
		cnn2, c2 = t(ns, ftype, index, cnn, TRANS_BLOCK2)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*SwitchStmt)
		}
		for idx := range cnn.Clauses {
			cnn.Clauses[idx] = *transcribe(t, nns, TRANS_SWITCH_CASE, idx, &cnn.Clauses[idx], &c).(*SwitchClauseStmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *SwitchClauseStmt:
		// NOTE: unlike the select case, both switch
		// statements AND switch cases visit with the
		// TRANS_BLOCK stage, even though during runtime
		// only one block is created.
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*SwitchClauseStmt)
		}
		for idx := range cnn.Cases {
			cnn.Cases[idx] = transcribe(t, nns, TRANS_SWITCHCASE_CASE, idx, cnn.Cases[idx], &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_SWITCHCASE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *FuncDecl:
		if cnn.Recv.Type != nil {
			cnn.Recv = *transcribe(t, nns, TRANS_FUNC_RECV, 0, &cnn.Recv, &c).(*FieldTypeExpr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		cnn.Type = *transcribe(t, nns, TRANS_FUNC_TYPE, 0, &cnn.Type, &c).(*FuncTypeExpr)
		if stopOrSkip(nc, c) {
			return
		}
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*FuncDecl)
		}
		// iterate over Body; its length can change if a statement is decomposed.
		for idx := 0; idx < len(cnn.Body); idx++ {
			cnn.Body[idx] = transcribe(t, nns, TRANS_FUNC_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *ImportDecl:
		// nothing to do
	case *ValueDecl:
		if cnn.Type != nil {
			cnn.Type = transcribe(t, nns, TRANS_VAR_TYPE, 0, cnn.Type, &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
		// XXX consider RHS, LHS, RHS, LHS, ... order.
		for idx := range cnn.NameExprs {
			cnn.NameExprs[idx] = *(transcribe(t, nns, TRANS_VAR_NAME, idx, &cnn.NameExprs[idx], &c).(*NameExpr))
		}
		for idx := range cnn.Values {
			cnn.Values[idx] = transcribe(t, nns, TRANS_VAR_VALUE, idx, cnn.Values[idx], &c).(Expr)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *TypeDecl:
		cnn.Type = transcribe(t, nns, TRANS_TYPE_TYPE, 0, cnn.Type, &c).(Expr)
		if stopOrSkip(nc, c) {
			return
		}
	case *FileNode:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if stopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*FileNode)
		}
		for idx := range cnn.Decls {
			cnn.Decls[idx] = transcribe(t, nns, TRANS_FILE_BODY, idx, cnn.Decls[idx], &c).(Decl)
			if stopOrSkip(nc, c) {
				return
			}
		}
	case *ConstExpr, *constTypeExpr: // leaf nodes
		// These nodes get created by the preprocessor while
		// leaving the type expression of a composite lit, before
		// visiting the key value elements of the composite lit.
	default:
		if n == nil {
			panic(fmt.Sprintf("node missing for %v", ftype))
		} else {
			panic(fmt.Sprintf("unexpected node type %#v", n))
		}
	}

	// transcribe n on the way out.
	nn, *nc = t(ns, ftype, index, nn, TRANS_LEAVE)

	return
}

// returns true if transcribe() should stop or skip or exit (& if so then sets
// *c to TRANS_EXIT if exit, or TRANS_CONTINUE if skip).
func stopOrSkip(oldnc *TransCtrl, nc TransCtrl) (stop bool) {
	switch nc {
	case TRANS_EXIT:
		*oldnc = TRANS_EXIT
		return true
	case TRANS_SKIP:
		*oldnc = TRANS_CONTINUE
		return true
	case TRANS_CONTINUE:
		return false
	default:
		panic("should not happen")
	}
}
