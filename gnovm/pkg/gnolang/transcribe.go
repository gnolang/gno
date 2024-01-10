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
	TRANS_BREAK
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
	TRANS_FUNCLIT_BODY
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
	TRANS_MAYBENATIVETYPE_TYPE
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
	TRANS_PANIC_EXCEPTION
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
	TRANS_VAR_TYPE
	TRANS_VAR_VALUE
	TRANS_TYPE_TYPE
	TRANS_FILE_BODY
)

// rules of closure capture
// closure captures namedExprs(nx)
// closure captures nx when:
// 1. defined outside the closure block, which implies nxs defined inside do not need to be captured,
//    it's evaluated naturally.
// 2. nx defined outside closure block are not all captured. it's captured only when it's volatile
//    e.g. a nx if mutated when the outside block is rangeStmt, for stmt, which dynamically
//    mutating the target nx. namely, any case you can not get a final value of the nx, you should
//    capture it. any time it's deemed final, no need to capture.

// FLOW:
// 1. start capture work when peeking funcLitExpr
// 2. do the capture work when traversing to nameExpr, check `hasClosure` to determine if capture needed
// 3. exclude any nxs that is not eligible:
//    a. defined locally
//    this is something tricky
//    one method is to give every nx an absolute level，so any nx can be compared by the `level`
//    with the level of the closure block to determine if it's defined inside or outside of the
//    closure block.
//    a second method is by checking the operators of `define` and `assign`, to filter out locally
//    defined nxs. TODO: is it doable?
//    ###whitelist is a collect of names to be whitelisted while capturing nxs, it contains:
//    1. params and results of funcLitExpr,
//    2. LHS of := and =
//    3. how about RHS? rhs can be either literals, or nxs(locally or not), it is defined locally,
//       it's bypassed by rule2. if is not local, capture it. so what we need is LHS.
//    FLOW of whitelist:
//    everytime encounter assign(assignStmt, init of if/range/for/switch stmt), push the whitelist
//    map with key of operator(define, assign), with value of assignStruct
//    while in traversing nx, first peek operator, the `check` the corresponding nx according to the
//    proper num set previously, the number should be counted, if counts to `num`, pop the operator,
//    implies an end for a assign/define stmt.
//    what the `check` does is to put the name of nx into the names, which is used to filter nxs.

//		b. final nx. TODO: this left a todo
//       it's can be done by traversing parent blocks, similar as `hasClosure`, to determine if there
//       is volatile blocks outside

var CX *ClosureContext

func init() {
	CX = &ClosureContext{whitelist: make(map[Name]bool)}
}

// AssignOperand is metadata to describe how much (left)nxs is related to an assign/define operator
type AssignOperand struct {
	num     int // num of lhs
	counter int // counter increased every time checked in traversing nx, if counter == num, mean it's resolved, and pop operator
}

func (ao *AssignOperand) String() string {
	var s string
	s += fmt.Sprintf("assign operand num: %d \n", ao.num)
	s += fmt.Sprintf("assign operand counter: %d \n", ao.counter)
	return s
}

type CNode struct {
	name Name
	n    Node
}

type ClosureContext struct {
	closures []*Closure
	nodes    []Node
	ops      []Word           // assign/define related logic
	operands []*AssignOperand // assign/define related logic, per operator, e.g. a, b := 0, 1
	//rc        []*RecursiveContext // detect cyclic, to find recursive closure, could converge with `nodes`
	whitelist map[Name]bool // use to filter out nxs
}

func (cx *ClosureContext) clearWhiteList() {
	debugPP.Println("clear whitelist")
	cx.whitelist = make(map[Name]bool)
	//cx.popOp()
	//cx.popOperand()
}

func (cx *ClosureContext) dumpWhitelist() string {
	var s string
	s += "===whitelist=== \n"
	for n, _ := range cx.whitelist {
		s += fmt.Sprintf("name is: %v \n", n)
	}
	return s
}

func (cx *ClosureContext) pushOp(op Word) {
	cx.ops = append(cx.ops, op)
}

func (cx *ClosureContext) popOp() {
	if len(cx.ops) != 0 {
		cx.ops = cx.ops[:len(cx.ops)-1]
	}
}

func (cx *ClosureContext) peekOp() Word {
	if len(cx.ops) != 0 {
		return cx.ops[len(cx.ops)-1]
	} else {
		return ILLEGAL
	}
}

func (cx *ClosureContext) popOperand() {
	if len(cx.operands) != 0 {
		cx.operands = cx.operands[:len(cx.operands)-1]
	}
}

func (cx *ClosureContext) peekOperand() *AssignOperand {
	if len(cx.operands) != 0 {
		return cx.operands[len(cx.operands)-1]
	} else {
		return nil
	}
}

func (cx *ClosureContext) dumpOps() string {
	var s string
	s += "\n"
	for _, o := range cx.ops {
		s += fmt.Sprintf("op: %v \n", o)
	}
	return s
}

// 1. has funcLitExpr
// 2. it's embedded in another volatile block, for/range stmt
// 2. no recursive closure(not support for now)
func (cx *ClosureContext) hasClosure() bool {
	for _, cn := range cx.nodes {
		if _, ok := cn.(*FuncLitExpr); ok {
			return true
		}
		//else if _, ok := c.(*FuncDecl); ok {
		//	return true
		//}
	}

	// detect cyclic
	// 1. encounter a nx, its type is funcLitExpr, name it t; how to get it?
	// 2. compare t with parent node type, could be direct ancestor, or indirect
	//    namely, if two(or more) same funcLitExpr appears in stack, have cyclic
	// which implies !hasClosure

	return false
}

func (cx *ClosureContext) push(n Node) bool {
	// push nodes stack
	//cn := &CNode{
	//	n: n,
	//}
	cx.nodes = append(cx.nodes, n)
	// push closure
	if cx.hasClosure() {
		debug.Println("+clo")
		debug.Println("before push closure")
		cx.dumpClosures()
		cx.closures = append(cx.closures, &Closure{}) // push empty closure to be filled in
		debug.Println("end push closure")
		cx.dumpClosures()
		return true
	} else {
		debug.Println("no fx in stack, no need push closure")
		return false
	}
}

func (cx *ClosureContext) pop(copy bool) *Closure {
	debug.Println("-clo")
	defer func() {
		if len(cx.nodes) != 0 {
			debug.Println("-node")
			cx.nodes = cx.nodes[:len(cx.nodes)-1]
		}
	}()

	if len(cx.closures) == 0 {
		return nil
	} else {
		if len(cx.closures) == 1 { // last one, clean context
			cx.whitelist = make(map[Name]bool)
		}
		c := cx.closures[len(cx.closures)-1] // get current
		for _, cnx := range c.cnxs {         // pop-> increase offset
			debug.Printf("+1 \n")
			cnx.offset += 1
		}
		cx.closures = cx.closures[:len(cx.closures)-1] // shrink

		if copy {
			currentClo := cx.currentClosure()
			if currentClo != nil { // if last closure, just pop, no copy
				// fill up current closure
				for _, cnx := range c.cnxs { // trace back captured nxs
					currentClo.Fill(*cnx)
				}
			}
		}
		debug.Println("after pop, dump")
		cx.dumpClosures()
		debug.Printf("c poped: %v, \n", c)
		return c
	}
}

func (cx *ClosureContext) peekNodes(offset int) Node {
	debug.Println("c:node")
	if len(cx.nodes) >= (1 + offset) {
		return cx.nodes[len(cx.nodes)-(1+offset)]
	}
	return nil
}

func (cx *ClosureContext) dumpNodes() {
	if debug {
		println("============Dump fxs===========")
		println("len: ", len(cx.nodes))
		for i, n := range cx.nodes {
			fmt.Printf("node[%d]: %v\n", i, n)
		}
		println("============end===============")
		println("\n")
	}
}
func (cx *ClosureContext) dumpClosures() {
	if debug {
		println("============Dump closures start=======")
		debug.Printf("depth of closures: %d \n", len(cx.closures))
		for _, c := range cx.closures {
			fmt.Printf("===>: %v \n", c)
		}
		println("============Dump closures end===============")
		println("\n")
	}
}

func (cx *ClosureContext) currentClosure() *Closure {
	debug.Printf("currentClosure, len: %d \n", len(cx.closures))
	if len(cx.closures) == 0 {
		return nil
	}
	return cx.closures[len(cx.closures)-1]
}

type CapturedNx struct {
	nx     *NameExpr
	offset uint8 // every captured nx has an offset, represents its distance to the funcLitExpr
}

func (cnx *CapturedNx) String() string {
	return fmt.Sprintf("nx is: %v, offset is: %d \n", cnx.nx, cnx.offset)
}

// captured NameExpr
type Closure struct {
	names     []Name
	cnxs      []*CapturedNx
	recursive bool
}

func (c *Closure) String() string {
	var s string
	s += "\n===========closure start============\n"
	for i, n := range c.names {
		s += fmt.Sprintf("names[%d] is: %s \n", i, string(n))
	}
	for i, c := range c.cnxs {
		s += fmt.Sprintf("cnxs[%d] is : [nx:%v, offset:%d] \n", i, c.nx, c.offset)
	}
	s += "===========closure end=============\n"
	return s
}

//func NewClosure() *Closure {
//	return &Closure{}
//}

func (clo *Closure) Fill(cnx CapturedNx) {
	debug.Printf("+nx: %v \n", cnx.nx)
	for _, n := range clo.names { // filter out existed nx
		if cnx.nx.Name == n {
			debug.Println("exist, return")
			return
		}
	}
	clo.names = append(clo.names, cnx.nx.Name)
	clo.cnxs = append(clo.cnxs, &cnx)
}

// return:
//   - TRANS_CONTINUE to visit children recursively;
//   - TRANS_SKIP to break out of the
//     ENTER,CHILDS1,[BLOCK,CHILDS2]?,LEAVE sequence for that node,
//     i.e. skipping (the rest of) it;
//   - TRANS_BREAK to break out of looping in CHILDS1 or CHILDS2,
//   - TRANS_EXIT to stop traversing altogether.
//
// Do not mutate ns.
// Must return a new node to replace the old one,
// or the node will be deleted (or set to nil).
// Read: transform the ftype/index of context ns which
// is n during stage.
type Transform func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl)

// n MUST be a pointer to a node struct.
// returns the transcribe code returned for n.
// returns new node nn to replace n.
func Transcribe(n Node, t Transform) (nn Node) {
	if reflect.TypeOf(n).Kind() != reflect.Ptr {
		panic("Transcribe() expects a non-pointer concrete Node struct")
	}
	var ns []Node = make([]Node, 0, 32)
	var nc TransCtrl
	nn = transcribe(t, ns, TRANS_ROOT, 0, n, &nc)
	return
}

func transcribe(t Transform, ns []Node, ftype TransField, index int, n Node, nc *TransCtrl) (nn Node) {
	// transcribe n on the way in.
	var c TransCtrl
	nn, c = t(ns, ftype, index, n, TRANS_ENTER)
	if isStopOrSkip(nc, c) {
		return
	}

	// push nn to node stack.
	nns := append(ns, nn)

	// visit any children of n.
	switch cnn := nn.(type) {
	case *NameExpr:
		debugPP.Printf("-----trans, nameExpr: %v \n", cnn)
		if CX.hasClosure() {
			debugPP.Printf("---has Closure, check: %v \n", cnn)
			debugPP.Println("---currentOp: ", CX.peekOp())
			debugPP.Println("---dump ops: ", CX.dumpOps())
			// recording names defined in closure as a whitelist
			if CX.peekOp() == DEFINE || CX.peekOp() == ASSIGN {
				ao := CX.peekOperand()
				if ao != nil { // staff to do
					debugPP.Printf("ao is: %v \n", ao)
					// in scope of define/assign op, record to whitelist
					ao.counter += 1
					// add nx to whitelist until resolved
					CX.whitelist[cnn.Name] = true
					debugPP.Println(CX.dumpWhitelist())
					if ao.counter == ao.num { // all resolved
						//CX.clearWhiteList()
						CX.popOp()
						CX.popOperand()
					}
				}
			}
			// capture logic
			// not exist in whitelist, capture
			if _, ok := CX.whitelist[cnn.Name]; !ok {
				debugPP.Printf("nx need capture: %s \n", string(cnn.Name))
				currentClo := CX.currentClosure()
				debugPP.Printf("currentClo: %v \n", currentClo)
				if currentClo != nil { // a closure to fill
					//if cnn.Path.Depth < 1 { // if local defined, no capture
					debugPP.Printf("---capture: %v \n", cnn)
					cnx := CapturedNx{
						nx:     cnn,
						offset: 0,
					}
					currentClo.Fill(cnx)
					CX.dumpClosures()
					//CX.dumpNodes()
				}
			}
		}
	case *BasicLitExpr:
	case *BinaryExpr:
		cnn.Left = transcribe(t, nns, TRANS_BINARY_LEFT, 0, cnn.Left, &c).(Expr) // XXX wished this worked with nil.
		if isStopOrSkip(nc, c) {
			return
		}
		cnn.Right = transcribe(t, nns, TRANS_BINARY_RIGHT, 0, cnn.Right, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *CallExpr:
		cnn.Func = transcribe(t, nns, TRANS_CALL_FUNC, 0, cnn.Func, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
		for idx := range cnn.Args {
			cnn.Args[idx] = transcribe(t, nns, TRANS_CALL_ARG, idx, cnn.Args[idx], &c).(Expr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *IndexExpr:
		cnn.X = transcribe(t, nns, TRANS_INDEX_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
		cnn.Index = transcribe(t, nns, TRANS_INDEX_INDEX, 0, cnn.Index, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *SelectorExpr:
		cnn.X = transcribe(t, nns, TRANS_SELECTOR_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *SliceExpr:
		cnn.X = transcribe(t, nns, TRANS_SLICE_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
		if cnn.Low != nil {
			cnn.Low = transcribe(t, nns, TRANS_SLICE_LOW, 0, cnn.Low, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		if cnn.High != nil {
			cnn.High = transcribe(t, nns, TRANS_SLICE_HIGH, 0, cnn.High, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		if cnn.Max != nil {
			cnn.Max = transcribe(t, nns, TRANS_SLICE_MAX, 0, cnn.Max, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
	case *StarExpr:
		cnn.X = transcribe(t, nns, TRANS_STAR_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *RefExpr:
		cnn.X = transcribe(t, nns, TRANS_REF_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *TypeAssertExpr:
		cnn.X = transcribe(t, nns, TRANS_TYPEASSERT_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
		if cnn.Type != nil {
			cnn.Type = transcribe(t, nns, TRANS_TYPEASSERT_TYPE, 0, cnn.Type, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
	case *UnaryExpr:
		cnn.X = transcribe(t, nns, TRANS_UNARY_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *CompositeLitExpr:
		if cnn.Type != nil {
			cnn.Type = transcribe(t, nns, TRANS_COMPOSITE_TYPE, 0, cnn.Type, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		for idx, kvx := range cnn.Elts {
			k, v := kvx.Key, kvx.Value
			if k != nil {
				k = transcribe(t, nns, TRANS_COMPOSITE_KEY, idx, k, &c).(Expr)
				if isBreak(c) {
					break
				} else if isStopOrSkip(nc, c) {
					return
				}
			}
			v = transcribe(t, nns, TRANS_COMPOSITE_VALUE, idx, v, &c).(Expr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
			cnn.Elts[idx] = KeyValueExpr{Key: k, Value: v}
		}
	case *FuncLitExpr:
		debug.Printf("-----trans, funcLitExpr: %v \n", cnn)
		//CX.dumpNodes()
		CX.dumpClosures()

		cnn.Type = *transcribe(t, nns, TRANS_FUNCLIT_TYPE, 0, &cnn.Type, &c).(*FuncTypeExpr)
		if isStopOrSkip(nc, c) {
			return
		}
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*FuncLitExpr)
		}
		// TODO: get all param and result names to filter out captured nxs
		// whitelist
		for i, n := range cnn.Names {
			debugPP.Printf("name[%d] in staticBlock is: %s \n", i, string(n))
			// put in whitelist, which is per traverse
			CX.whitelist[n] = true
		}
		debugPP.Println("---start trans funcLit body stmt, push initial closure and fx")
		pushed := CX.push(cnn)
		//var pushed bool

		debugPP.Printf("---stop or skip, pop and return \n")
		node := CX.peekNodes(1)
		isCopy := true
		if _, ok := node.(*FuncLitExpr); ok {
			isCopy = false
		}

		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_FUNCLIT_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				// pop before return
				if pushed {
					CX.pop(isCopy)
				}
				return
			}
		}
		// defer pop
		debugPP.Printf("---done trans body \n")
		// TODO: set fx.Closure, and level as well
		debugPP.Println("funcLit pop c-----")

		if pushed {
			pc := CX.pop(isCopy)
			if pc != nil {
				cnn.SetClosure(pc)
				debugPP.Printf("---done FuncLit trans, fx: %v, closure: %+v \n", cnn, cnn.Closure.String())
			}
		}
	case *FieldTypeExpr:
		cnn.Type = transcribe(t, nns, TRANS_FIELDTYPE_TYPE, 0, cnn.Type, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
		if cnn.Tag != nil {
			cnn.Tag = transcribe(t, nns, TRANS_FIELDTYPE_TAG, 0, cnn.Tag, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
	case *ArrayTypeExpr:
		if cnn.Len != nil {
			cnn.Len = transcribe(t, nns, TRANS_ARRAYTYPE_LEN, 0, cnn.Len, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		cnn.Elt = transcribe(t, nns, TRANS_ARRAYTYPE_ELT, 0, cnn.Elt, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *SliceTypeExpr:
		cnn.Elt = transcribe(t, nns, TRANS_SLICETYPE_ELT, 0, cnn.Elt, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *InterfaceTypeExpr:
		for idx := range cnn.Methods {
			cnn.Methods[idx] = *transcribe(t, nns, TRANS_INTERFACETYPE_METHOD, idx, &cnn.Methods[idx], &c).(*FieldTypeExpr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *ChanTypeExpr:
		cnn.Value = transcribe(t, nns, TRANS_CHANTYPE_VALUE, 0, cnn.Value, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *FuncTypeExpr:
		for idx := range cnn.Params {
			cnn.Params[idx] = *transcribe(t, nns, TRANS_FUNCTYPE_PARAM, idx, &cnn.Params[idx], &c).(*FieldTypeExpr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		for idx := range cnn.Results {
			cnn.Results[idx] = *transcribe(t, nns, TRANS_FUNCTYPE_RESULT, idx, &cnn.Results[idx], &c).(*FieldTypeExpr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *MapTypeExpr:
		cnn.Key = transcribe(t, nns, TRANS_MAPTYPE_KEY, 0, cnn.Key, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
		cnn.Value = transcribe(t, nns, TRANS_MAPTYPE_VALUE, 0, cnn.Value, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *StructTypeExpr:
		for idx := range cnn.Fields {
			cnn.Fields[idx] = *transcribe(t, nns, TRANS_STRUCTTYPE_FIELD, idx, &cnn.Fields[idx], &c).(*FieldTypeExpr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *MaybeNativeTypeExpr:
		cnn.Type = transcribe(t, nns, TRANS_MAYBENATIVETYPE_TYPE, 0, cnn.Type, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *AssignStmt:
		debugPP.Printf("---assignStmt: %v \n", cnn)
		if CX.hasClosure() {
			debugPP.Println("---push op and operands")
			// push op(assign/define) and operands
			CX.ops = append(CX.ops, cnn.Op)
			ao := &AssignOperand{}
			ao.num = len(cnn.Lhs)
			CX.operands = append(CX.operands, ao)
		}
		for idx := range cnn.Lhs {
			cnn.Lhs[idx] = transcribe(t, nns, TRANS_ASSIGN_LHS, idx, cnn.Lhs[idx], &c).(Expr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
		for idx := range cnn.Rhs {
			cnn.Rhs[idx] = transcribe(t, nns, TRANS_ASSIGN_RHS, idx, cnn.Rhs[idx], &c).(Expr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *BlockStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*BlockStmt)
		}
		pushed := CX.push(cnn)
		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_BLOCK_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if pushed {
			CX.pop(true)
		}
	case *BranchStmt:
	case *DeclStmt:
		//CX.pushOp(ASSIGN)
		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_DECL_BODY, idx, cnn.Body[idx], &c).(SimpleDeclStmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
		//CX.popOp()
	case *DeferStmt:
		cnn.Call = *transcribe(t, nns, TRANS_DEFER_CALL, 0, &cnn.Call, &c).(*CallExpr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *EmptyStmt:
	case *ExprStmt:
		cnn.X = transcribe(t, nns, TRANS_EXPR_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *ForStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*ForStmt)
		}

		pushed := CX.push(cnn)

		if cnn.Init != nil {
			cnn.Init = transcribe(t, nns, TRANS_FOR_INIT, 0, cnn.Init, &c).(SimpleStmt)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		if cnn.Cond != nil {
			cnn.Cond = transcribe(t, nns, TRANS_FOR_COND, 0, cnn.Cond, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		if cnn.Post != nil {
			cnn.Post = transcribe(t, nns, TRANS_FOR_POST, 0, cnn.Post, &c).(SimpleStmt)
			if isStopOrSkip(nc, c) {
				return
			}
		}

		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_FOR_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if pushed {
			CX.pop(true)
		}
	case *GoStmt:
		cnn.Call = *transcribe(t, nns, TRANS_GO_CALL, 0, &cnn.Call, &c).(*CallExpr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *IfStmt:
		debug.Println("-----trans, if stmt")
		// NOTE: like switch stmts, both if statements AND
		// contained cases visit with the TRANS_BLOCK stage, even
		// though during runtime only one block is created.
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*IfStmt)
		}

		pushed := CX.push(cnn)

		// nx in init is always treat defined locally
		if cnn.Init != nil {
			cnn.Init = transcribe(t, nns, TRANS_IF_INIT, 0, cnn.Init, &c).(SimpleStmt)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		cnn.Cond = transcribe(t, nns, TRANS_IF_COND, 0, cnn.Cond, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}

		cnn.Then = *transcribe(t, nns, TRANS_IF_BODY, 0, &cnn.Then, &c).(*IfCaseStmt)
		if isStopOrSkip(nc, c) {
			return
		}
		cnn.Else = *transcribe(t, nns, TRANS_IF_ELSE, 0, &cnn.Else, &c).(*IfCaseStmt)
		if isStopOrSkip(nc, c) {
			return
		}
		if pushed {
			CX.pop(true)
		}
	case *IfCaseStmt:
		debug.Printf("-----trans, (if---case) stmt: %v \n", cnn)
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*IfCaseStmt)
		}
		//pushClosure(&Closure{})
		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_IF_CASE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
		//debug.Println("if-case pop c-----")

		//popClosure()
	case *IncDecStmt:
		cnn.X = transcribe(t, nns, TRANS_INCDEC_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *RangeStmt:
		debugPP.Printf("---range stmt: %v \n", cnn)
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*RangeStmt)
		}

		pushed := CX.push(cnn)

		cnn.X = transcribe(t, nns, TRANS_RANGE_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			if pushed {
				CX.pop(true)
			}
			return
		}

		if CX.hasClosure() { // TODO: do we need this?
			debugPP.Println("---push op and operands")
			// push op(assign/define) and operands
			CX.ops = append(CX.ops, cnn.Op)
			ao := &AssignOperand{}
			ao.num = 2 // key, value
			CX.operands = append(CX.operands, ao)
		}
		if cnn.Key != nil {
			cnn.Key = transcribe(t, nns, TRANS_RANGE_KEY, 0, cnn.Key, &c).(Expr)
			if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if cnn.Value != nil {
			cnn.Value = transcribe(t, nns, TRANS_RANGE_VALUE, 0, cnn.Value, &c).(Expr)
			if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}

		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_RANGE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if pushed {
			CX.pop(true)
		}
	case *ReturnStmt:
		debug.Printf("-----trans, return stmt: %v \n", cnn)
		for idx := range cnn.Results {
			cnn.Results[idx] = transcribe(t, nns, TRANS_RETURN_RESULT, idx, cnn.Results[idx], &c).(Expr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *PanicStmt:
		cnn.Exception = transcribe(t, nns, TRANS_PANIC_EXCEPTION, 0, cnn.Exception, &c).(Expr)
	case *SelectStmt:
		for idx := range cnn.Cases {
			cnn.Cases[idx] = *transcribe(t, nns, TRANS_SELECT_CASE, idx, &cnn.Cases[idx], &c).(*SelectCaseStmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *SelectCaseStmt:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*SelectCaseStmt)
		}
		pushed := CX.push(cnn)
		cnn.Comm = transcribe(t, nns, TRANS_SELECTCASE_COMM, 0, cnn.Comm, &c).(Stmt)
		if isStopOrSkip(nc, c) {
			return
		}
		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_SELECTCASE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if pushed {
			CX.pop(true)
		}
	case *SendStmt:
		cnn.Chan = transcribe(t, nns, TRANS_SEND_CHAN, 0, cnn.Chan, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
		cnn.Value = transcribe(t, nns, TRANS_SEND_VALUE, 0, cnn.Value, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *SwitchStmt:
		debugPP.Printf("---switchStmt: %v \n", cnn)
		// NOTE: unlike the select case, and like if stmts, both
		// switch statements AND contained cases visit with the
		// TRANS_BLOCK stage, even though during runtime only one
		// block is created.
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*SwitchStmt)
		}

		if cnn.IsTypeSwitch {
			debugPP.Printf("is type switch, init is :%v \n", cnn.Init)
			debugPP.Printf("is type switch, X is :%v \n", cnn.X)
			debugPP.Printf("is type switch, varName is :%v \n", cnn.VarName)
			CX.whitelist[cnn.VarName] = true
		}
		pushed := CX.push(cnn)

		if cnn.Init != nil {
			cnn.Init = transcribe(t, nns, TRANS_SWITCH_INIT, 0, cnn.Init, &c).(SimpleStmt)
			if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		cnn.X = transcribe(t, nns, TRANS_SWITCH_X, 0, cnn.X, &c).(Expr)
		if isStopOrSkip(nc, c) {
			if pushed {
				CX.pop(true)
			}
			return
		}
		// NOTE: special block case for after .Init and .X.
		cnn2, c2 = t(ns, ftype, index, cnn, TRANS_BLOCK2)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			if pushed {
				CX.pop(true)
			}
			return
		} else {
			cnn = cnn2.(*SwitchStmt)
		}
		for idx := range cnn.Clauses {
			cnn.Clauses[idx] = *transcribe(t, nns, TRANS_SWITCH_CASE, idx, &cnn.Clauses[idx], &c).(*SwitchClauseStmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if pushed {
			CX.pop(true)
		}
	case *SwitchClauseStmt:
		// NOTE: unlike the select case, both switch
		// statements AND switch cases visit with the
		// TRANS_BLOCK stage, even though during runtime
		// only one block is created.
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*SwitchClauseStmt)
		}
		for idx := range cnn.Cases {
			cnn.Cases[idx] = transcribe(t, nns, TRANS_SWITCHCASE_CASE, idx, cnn.Cases[idx], &c).(Expr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_SWITCHCASE_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
	case *FuncDecl:
		if cnn.Recv.Type != nil {
			cnn.Recv = *transcribe(t, nns, TRANS_FUNC_RECV, 0, &cnn.Recv, &c).(*FieldTypeExpr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		cnn.Type = *transcribe(t, nns, TRANS_FUNC_TYPE, 0, &cnn.Type, &c).(*FuncTypeExpr)
		if isStopOrSkip(nc, c) {
			return
		}
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*FuncDecl)
		}

		pushed := CX.push(cnn)
		for idx := range cnn.Body {
			cnn.Body[idx] = transcribe(t, nns, TRANS_FUNC_BODY, idx, cnn.Body[idx], &c).(Stmt)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if pushed {
			CX.pop(true)
		}
	case *ImportDecl:
		// nothing to do
	case *ValueDecl:
		debugPP.Println("---value decl")
		if CX.hasClosure() {
			debugPP.Println("---push op and operands")
			// push op(assign/define) and operands
			CX.ops = append(CX.ops, ASSIGN)
			ao := &AssignOperand{}
			ao.num = len(cnn.NameExprs)
			CX.operands = append(CX.operands, ao)
		}

		if cnn.Type != nil {
			cnn.Type = transcribe(t, nns, TRANS_VAR_TYPE, 0, cnn.Type, &c).(Expr)
			if isStopOrSkip(nc, c) {
				return
			}
		}
		for idx := range cnn.Values {
			cnn.Values[idx] = transcribe(t, nns, TRANS_VAR_VALUE, idx, cnn.Values[idx], &c).(Expr)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				return
			}
		}
		CX.popOp()
	case *TypeDecl:
		cnn.Type = transcribe(t, nns, TRANS_TYPE_TYPE, 0, cnn.Type, &c).(Expr)
		if isStopOrSkip(nc, c) {
			return
		}
	case *FileNode:
		cnn2, c2 := t(ns, ftype, index, cnn, TRANS_BLOCK)
		if isStopOrSkip(nc, c2) {
			nn = cnn2
			return
		} else {
			cnn = cnn2.(*FileNode)
		}

		pushed := CX.push(cnn)

		for idx := range cnn.Decls {
			cnn.Decls[idx] = transcribe(t, nns, TRANS_FILE_BODY, idx, cnn.Decls[idx], &c).(Decl)
			if isBreak(c) {
				break
			} else if isStopOrSkip(nc, c) {
				if pushed {
					CX.pop(true)
				}
				return
			}
		}
		if pushed {
			CX.pop(true)
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

// returns true if transcribe() should stop or skip or exit (& if so then sets *c to TRANS_EXIT if exit, or TRANS_CONTINUE if break).
func isStopOrSkip(oldnc *TransCtrl, nc TransCtrl) (stop bool) {
	if nc == TRANS_EXIT {
		*oldnc = TRANS_EXIT
		return true
	} else if nc == TRANS_SKIP {
		*oldnc = TRANS_CONTINUE
		return true
	} else if nc == TRANS_CONTINUE {
		return false
	} else {
		panic("should not happen")
	}
}

// returns true if transcribe() should break (a loop).
func isBreak(nc TransCtrl) (brek bool) {
	if nc == TRANS_BREAK {
		return true
	} else {
		return false
	}
}
