package gno

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

//----------------------------------------
// Machine

type Machine struct {

	// State
	Ops       []Op // main operations
	NumOps    int
	Values    []TypedValue  // buffer of values to be operated on
	NumValues int           // number of values
	Exprs     []Expr        // pending expressions
	Stmts     []Stmt        // pending statements
	Blocks    []*Block      // block (scope) stack
	Frames    []Frame       // func call stack
	Package   *PackageValue // active package
	Realm     *Realm        // active realm
	Exception *TypedValue   // if panic'd unless recovered

	// Volatile State
	NumResults int // number of results returned

	// Configuration
	CheckTypes bool
	Output     io.Writer
	Store      Store
}

// Machine with new package of given path.
// Creates a new MemRealmer for any new realms.
func NewMachine(pkgPath string, store Store) *Machine {
	pkgName := defaultPkgName(pkgPath)
	rlm := (*Realm)(nil)
	if IsRealmPath(pkgPath) {
		rlm = NewRealm(pkgPath)
	}
	pn := NewPackageNode(pkgName, pkgPath, &FileSet{})
	pv := pn.NewPackage(rlm)
	return NewMachineWithOptions(
		MachineOptions{
			Package: pv,
			Store:   store,
		})
}

type MachineOptions struct {
	Package    *PackageValue
	CheckTypes bool
	Output     io.Writer
	Store      Store
}

func NewMachineWithOptions(opts MachineOptions) *Machine {
	pkg := opts.Package
	if pkg == nil {
		pn := NewPackageNode("main", ".main", &FileSet{})
		pkg = pn.NewPackage(nil) // no realm by default.
	}
	rlm := pkg.GetRealm()
	checkTypes := opts.CheckTypes
	output := opts.Output
	if output == nil {
		output = os.Stdout
	}
	store := opts.Store
	if store == nil {
		// bare machine, no stdlibs.
	}
	blocks := []*Block{
		&pkg.Block,
	}
	return &Machine{
		Ops:        make([]Op, 1024),
		NumOps:     0,
		Values:     make([]TypedValue, 1024),
		NumValues:  0,
		Blocks:     blocks,
		Package:    pkg,
		Realm:      rlm,
		CheckTypes: checkTypes,
		Output:     output,
		Store:      store,
	}
}

//----------------------------------------
// top level Run* methods.

// Add files to the package's *FileSet and run them.
// This will also run each init function encountered.
func (m *Machine) RunFiles(fns ...*FileNode) {
	// Files' package names must match the machine's active one.
	// if there is one.
	for _, fn := range fns {
		if fn.PkgName != "" && fn.PkgName != m.Package.PkgName {
			panic(fmt.Sprintf("expected package name [%s] but got [%s]",
				m.Package.PkgName, fn.PkgName))
		}
	}
	// Add files to *PackageNode.FileSet.
	pv := m.Package
	pn := pv.Source.(*PackageNode)
	if pn.FileSet == nil {
		pn.FileSet = &FileSet{}
	}
	pn.FileSet.AddFiles(fns...)
	fileupdates := make([][]TypedValue, len(fns))

	// Preprocess each new file.
	for i, fn := range fns {
		// Preprocess file.
		// NOTE: Most of the declaration is handled by
		// Preprocess and any constant values set on
		// pn.StaticBlock, and those values are copied to the
		// runtime package value via PrepareNewValues.  Then,
		// non-constant var declarations and file-level imports
		// are re-set in runDeclaration(,true).
		fn = Preprocess(m.Store, pn, fn).(*FileNode)
		if debug {
			debug.Println("PREPROCESSED FILE: ", fn.String())
		}
		// Make block for fn.
		// Each file for each *PackageValue gets its own file *Block,
		// with values copied over from each file's
		// *FileNode.StaticBlock.
		fb := NewBlock(fn, &pv.Block)
		fb.Values = make([]TypedValue, len(fn.StaticBlock.Values))
		copy(fb.Values, fn.StaticBlock.Values)
		pv.AddFileBlock(fn.Name, fb)
		updates := pn.PrepareNewValues(pv) // with fb
		fileupdates[i] = updates
	}

	// exists if declaration run.
	var fdeclared = map[Name]struct{}{}
	// to detect loops in var declarations.
	var loopfindr = []Name{}
	// recursive function for var declarations.
	var runDeclarationFor func(fn *FileNode, decl Decl)
	runDeclarationFor = func(fn *FileNode, decl Decl) {
		// get dependencies of decl.
		deps := make(map[Name]struct{})
		findDependentNames(decl, deps)
		for dep, _ := range deps {
			// if dep already in fdeclared, skip.
			if _, ok := fdeclared[dep]; ok {
				continue
			}
			fn, depdecl := pn.FileSet.GetDeclFor(dep)
			// if dep already in loopfindr, abort.
			if hasName(dep, loopfindr) {
				if _, ok := (*depdecl).(*FuncDecl); ok {
					// recursive function dependencies
					// are OK with func decls.
					continue
				} else {
					panic(fmt.Sprintf(
						"loop in variable initialization: dependency trail %v circularly depends on %s", loopfindr, dep))
				}
			}
			// run dependecy declaration
			loopfindr = append(loopfindr, dep)
			runDeclarationFor(fn, *depdecl)
			loopfindr = loopfindr[:len(loopfindr)-1]
		}
		// run declaration
		fb := pv.GetFileBlock(m.Store, fn.Name)
		m.PushBlock(fb)
		m.runDeclaration(decl)
		m.PopBlock()
		for _, n := range decl.GetDeclNames() {
			fdeclared[n] = struct{}{}
		}
	}

	// Variable initialization.  This must happen after all
	// files are preprocessed, because value decl may be
	// out of order and depend on other files.
	for _, fn := range fns {
		// Run declarations.
		for _, decl := range fn.Decls {
			runDeclarationFor(fn, decl)
		}
	}

	// Run new init functions.
	// Go spec: "To ensure reproducible initialization
	// behavior, build systems are encouraged to present
	// multiple files belonging to the same package in
	// lexical file name order to a compiler."
	for i, fn := range fns {
		updates := fileupdates[i]
		fb := pv.GetFileBlock(m.Store, fn.Name)
		m.PushBlock(fb)
		for i := 0; i < len(updates); i++ {
			tv := &updates[i]
			if tv.IsDefined() && tv.T.Kind() == FuncKind && tv.V != nil {
				if fv, ok := tv.V.(*FuncValue); ok {
					fn := fv.Name
					if strings.HasPrefix(string(fn), "init.") {
						m.RunFunc(fn)
					}
				}
			}
		}
		m.PopBlock()
	}
}

func (m *Machine) RunFunc(fn Name) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Machine.RunFunc(%q) panic: %v\n%s\n",
				fn, r, m.String())
			panic(r)
		}
	}()
	m.RunStatement(S(Call(Nx(fn))))
}

func (m *Machine) RunMain() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Machine.RunMain() panic: %v\n%s\n",
				r, m.String())
			panic(r)
		}
	}()
	m.RunStatement(S(Call(X("main"))))
}

// Evaluate throwaway expression in new block scope.
// If x is a function call, it must return 1 result.
// This function is mainly for debugging and testing,
// but it could also be useful for a repl.
// Input must not have been preprocessed, that is,
// it should not be the child of any parent.
func (m *Machine) Eval(x Expr) TypedValue {
	if debug {
		m.Printf("Machine.Eval(%v)\n", x)
	}
	// X must not have been preprocessed.
	if x.GetAttribute(ATTR_PREPROCESSED) != nil {
		panic(fmt.Sprintf(
			"Machine.Eval(x) expression already preprocessed: %s",
			x.String()))
	}
	// Preprocess input using package block.
	// There should only be one block, a *PackageNode.
	// Other usage styles not yet supported.
	pn := m.LastBlock().Source.(*PackageNode)
	// Transform expression to ensure isolation.
	// This is to ensure that the existing machine
	// context (ie **PackageNode) doesn't get modified.
	if _, ok := x.(*CallExpr); !ok {
		x = Call(Fn(nil, Flds("x", InterfaceT(nil)),
			Ss(
				Return(x),
			)))
	} else {
		// x already creates its own scope.
	}
	// Preprocess x.
	x = Preprocess(m.Store, pn, x).(Expr)
	// Evaluate x.
	start := m.NumValues
	m.PushOp(OpHalt)
	m.PushExpr(x)
	m.PushOp(OpEval)
	m.Run()
	res := m.ReapValues(start)
	if len(res) != 1 {
		panic("should not happen")
	}
	return res[0]
}

// Evaluate any preprocessed expression statically.
// This is primiarily used by the preprocessor to evaluate
// static types and values.
func (m *Machine) EvalStatic(last BlockNode, x Expr) TypedValue {
	if debug {
		m.Printf("Machine.EvalStatic(%v, %v)\n", last, x)
	}
	// X must have been preprocessed.
	if x.GetAttribute(ATTR_PREPROCESSED) == nil {
		panic(fmt.Sprintf(
			"Machine.EvalStatic(x) expression not yet preprocessed: %s",
			x.String()))
	}
	// Temporarily push last to m.Blocks.
	m.PushBlock(last.GetStaticBlock().GetBlock())
	// Evaluate x.
	start := m.NumValues
	m.PushOp(OpHalt)
	m.PushOp(OpPopBlock)
	m.PushExpr(x)
	m.PushOp(OpEval)
	m.Run()
	res := m.ReapValues(start)
	if len(res) != 1 {
		panic("should not happen")
	}
	return res[0]
}

// Evaluate the type of any preprocessed expression statically.
// This is primiarily used by the preprocessor to evaluate
// static types of nodes.
func (m *Machine) EvalStaticTypeOf(last BlockNode, x Expr) Type {
	if debug {
		m.Printf("Machine.EvalStaticTypeOf(%v, %v)\n", last, x)
	}
	// X must have been preprocessed.
	if x.GetAttribute(ATTR_PREPROCESSED) == nil {
		panic(fmt.Sprintf(
			"Machine.EvalStaticTypeOf(x) expression not yet preprocessed: %s",
			x.String()))
	}
	// Temporarily push last to m.Blocks.
	m.PushBlock(last.GetStaticBlock().GetBlock())
	// Evaluate x.
	start := m.NumValues
	m.PushOp(OpHalt)
	m.PushOp(OpPopBlock)
	m.PushExpr(x)
	m.PushOp(OpStaticTypeOf)
	m.Run()
	res := m.ReapValues(start)
	if len(res) != 1 {
		panic("should not happen")
	}
	tv := res[0].V.(TypeValue)
	return tv.Type
}

func (m *Machine) RunStatement(s Stmt) {
	sn := m.LastBlock().Source
	s = Preprocess(m.Store, sn, s).(Stmt)
	m.PushOp(OpHalt)
	m.PushStmt(s)
	m.PushOp(OpExec)
	m.Run()
}

// Runs a declaration after preprocessing d.  If d was already
// preprocessed, call runDeclaration() instead.
func (m *Machine) RunDeclaration(d Decl) {
	// Preprocess input using package block.  There should only
	// be one block right now, and it's a *PackageNode.
	pn := m.LastBlock().Source.(*PackageNode)
	d = Preprocess(m.Store, pn, d).(Decl)
	pn.PrepareNewValues(m.Package)
	m.runDeclaration(d)
	if debug {
		if pn != m.Package.Source {
			panic("package mismatch")
		}
	}
}

// Declarations to be run within a body (not at the file or
// package level, for which evaluations happen during
// preprocessing).
func (m *Machine) runDeclaration(d Decl) {
	switch d := d.(type) {
	case *FuncDecl:
		// nothing to do.
		// closure and package already set
		// during PackageNode.NewPackage().
	case *ValueDecl:
		m.PushOp(OpHalt)
		m.PushStmt(d)
		m.PushOp(OpExec)
		m.Run()
	case *TypeDecl:
		m.PushOp(OpHalt)
		m.PushStmt(d)
		m.PushOp(OpExec)
		m.Run()
	default:
		// Do nothing for package constants.
	}
}

//----------------------------------------
// Op

type Op uint8

const (

	/* Control operators */
	OpInvalid             Op = 0x00 // invalid
	OpHalt                Op = 0x01 // halt (e.g. last statement)
	OpNoop                Op = 0x02 // no-op
	OpExec                Op = 0x03 // exec next statement
	OpPrecall             Op = 0x04 // sets X (func) to frame
	OpCall                Op = 0x05 // call(Frame.Func, [...])
	OpCallNativeBody      Op = 0x06 // call body is native
	OpReturn              Op = 0x07 // return ...
	OpReturnFromBlock     Op = 0x08 // return results (after defers)
	OpReturnToBlock       Op = 0x09 // copy results to block (before defer)
	OpDefer               Op = 0x0A // defer call(X, [...])
	OpCallDeferNativeBody Op = 0x0B // call body is native
	OpGo                  Op = 0x0C // go call(X, [...])
	OpSelect              Op = 0x0D // exec next select case
	OpSwitchClause        Op = 0x0E // exec next switch clause
	OpSwitchClauseCase    Op = 0x0F // exec next switch clause case
	OpTypeSwitch          Op = 0x10 // exec type switch clauses (all)
	OpIfCond              Op = 0x11 // eval cond
	OpPopValue            Op = 0x12 // pop X
	OpPopResults          Op = 0x13 // pop n call results
	OpPopBlock            Op = 0x14 // pop block NOTE breaks certain invariants.
	OpPanic1              Op = 0x15 // pop exception and pop call frames.
	OpPanic2              Op = 0x16 // pop call frames.

	/* Unary & binary operators */
	OpUpos  Op = 0x20 // + (unary)
	OpUneg  Op = 0x21 // - (unary)
	OpUnot  Op = 0x22 // ! (unary)
	OpUxor  Op = 0x23 // ^ (unary)
	OpUrecv Op = 0x25 // <- (unary) // TODO make expr
	OpLor   Op = 0x26 // ||
	OpLand  Op = 0x27 // &&
	OpEql   Op = 0x28 // ==
	OpNeq   Op = 0x29 // !=
	OpLss   Op = 0x2A // <
	OpLeq   Op = 0x2B // <=
	OpGtr   Op = 0x2C // >
	OpGeq   Op = 0x2D // >=
	OpAdd   Op = 0x2E // +
	OpSub   Op = 0x2F // -
	OpBor   Op = 0x30 // |
	OpXor   Op = 0x31 // ^
	OpMul   Op = 0x32 // *
	OpQuo   Op = 0x33 // /
	OpRem   Op = 0x34 // %
	OpShl   Op = 0x35 // <<
	OpShr   Op = 0x36 // >>
	OpBand  Op = 0x37 // &
	OpBandn Op = 0x38 // &^

	/* Other expression operators */
	OpEval         Op = 0x40 // eval next expression
	OpBinary1      Op = 0x41 // X op ?
	OpIndex1       Op = 0x42 // X[Y]
	OpIndex2       Op = 0x43 // (_, ok :=) X[Y]
	OpSelector     Op = 0x44 // X.Y
	OpSlice        Op = 0x45 // X[Low:High:Max]
	OpStar         Op = 0x46 // *X (deref or pointer-to)
	OpRef          Op = 0x47 // &X
	OpTypeAssert1  Op = 0x48 // X.(Type)
	OpTypeAssert2  Op = 0x49 // (_, ok :=) X.(Type)
	OpStaticTypeOf Op = 0x4A // static type of X
	OpCompositeLit Op = 0x4B // X{???}
	OpArrayLit     Op = 0x4C // [Len]{...}
	OpSliceLit     Op = 0x4D // []{...}
	OpMapLit       Op = 0x4E // X{...}
	OpStructLit    Op = 0x4F // X{...}
	OpFuncLit      Op = 0x50 // func(T){Body}
	OpConvert      Op = 0x51 // Y(X)

	/* Native operators */
	OpStructLitGoNative Op = 0x60
	OpCallGoNative      Op = 0x61

	/* Type operators */
	OpFieldType     Op = 0x70 // Name: X `tag`
	OpArrayType     Op = 0x71 // [X]Y{}
	OpSliceType     Op = 0x72 // []X{}
	OpPointerType   Op = 0x73 // *X
	OpInterfaceType Op = 0x74 // interface{...}
	OpChanType      Op = 0x75 // [<-]chan[<-]X
	OpFuncType      Op = 0x76 // func(params...)results...
	OpMapType       Op = 0x77 // map[X]Y
	OpStructType    Op = 0x78 // struct{...}

	/* Statement operators */
	OpAssign      Op = 0x80 // Lhs = Rhs
	OpAddAssign   Op = 0x81 // Lhs += Rhs
	OpSubAssign   Op = 0x82 // Lhs -= Rhs
	OpMulAssign   Op = 0x83 // Lhs *= Rhs
	OpQuoAssign   Op = 0x84 // Lhs /= Rhs
	OpRemAssign   Op = 0x85 // Lhs %= Rhs
	OpBandAssign  Op = 0x86 // Lhs &= Rhs
	OpBandnAssign Op = 0x87 // Lhs &^= Rhs
	OpBorAssign   Op = 0x88 // Lhs |= Rhs
	OpXorAssign   Op = 0x89 // Lhs ^= Rhs
	OpShlAssign   Op = 0x8A // Lhs <<= Rhs
	OpShrAssign   Op = 0x8B // Lhs >>= Rhs
	OpDefine      Op = 0x8C // X... := Y...
	OpInc         Op = 0x8D // X++
	OpDec         Op = 0x8E // X--

	/* Decl operators */
	OpValueDecl Op = 0x90 // var/const ...
	OpTypeDecl  Op = 0x91 // type ...

	/* Loop (sticky) operators (>= 0xD0) */
	OpSticky            Op = 0xD0 // not a real op.
	OpBody              Op = 0xD1 // if/block/switch/select.
	OpForLoop           Op = 0xD2
	OpRangeIter         Op = 0xD3
	OpRangeIterString   Op = 0xD4
	OpRangeIterMap      Op = 0xD5
	OpRangeIterArrayPtr Op = 0xD6
	OpReturnCallDefers  Op = 0xD7 // TODO rename?
)

//----------------------------------------
// main run loop.

func (m *Machine) Run() {
	for {
		op := m.PopOp()
		// TODO: this can be optimized manually, even into tiers.
		switch op {
		/* Control operators */
		case OpHalt:
			return
		case OpNoop:
			continue
		case OpExec:
			m.doOpExec(op)
		case OpPrecall:
			m.doOpPrecall()
		case OpCall:
			m.doOpCall()
		case OpCallNativeBody:
			m.doOpCallNativeBody()
		case OpReturn:
			m.doOpReturn()
		case OpReturnFromBlock:
			m.doOpReturnFromBlock()
		case OpReturnToBlock:
			m.doOpReturnToBlock()
		case OpDefer:
			m.doOpDefer()
		case OpPanic1:
			m.doOpPanic1()
		case OpPanic2:
			m.doOpPanic2()
		case OpCallDeferNativeBody:
			m.doOpCallDeferNativeBody()
		case OpGo:
			panic("not yet implemented")
		case OpSelect:
			panic("not yet implemented")
		case OpSwitchClause:
			m.doOpSwitchClause()
		case OpSwitchClauseCase:
			m.doOpSwitchClauseCase()
		case OpTypeSwitch:
			m.doOpTypeSwitch()
		case OpIfCond:
			m.doOpIfCond()
		case OpPopValue:
			m.PopValue()
		case OpPopResults:
			m.PopResults()
		case OpPopBlock:
			m.PopBlock()
		/* Unary operators */
		case OpUpos:
			m.doOpUpos()
		case OpUneg:
			m.doOpUneg()
		case OpUnot:
			m.doOpUnot()
		case OpUxor:
			m.doOpUxor()
		case OpUrecv:
			m.doOpUrecv()
		/* Binary operators */
		case OpLor:
			m.doOpLor()
		case OpLand:
			m.doOpLand()
		case OpEql:
			m.doOpEql()
		case OpNeq:
			m.doOpNeq()
		case OpLss:
			m.doOpLss()
		case OpLeq:
			m.doOpLeq()
		case OpGtr:
			m.doOpGtr()
		case OpGeq:
			m.doOpGeq()
		case OpAdd:
			m.doOpAdd()
		case OpSub:
			m.doOpSub()
		case OpBor:
			m.doOpBor()
		case OpXor:
			m.doOpXor()
		case OpMul:
			m.doOpMul()
		case OpQuo:
			m.doOpQuo()
		case OpRem:
			m.doOpRem()
		case OpShl:
			m.doOpShl()
		case OpShr:
			m.doOpShr()
		case OpBand:
			m.doOpBand()
		case OpBandn:
			m.doOpBandn()
		/* Expression operators */
		case OpEval:
			m.doOpEval()
		case OpBinary1:
			m.doOpBinary1()
		case OpIndex1:
			m.doOpIndex1()
		case OpIndex2:
			m.doOpIndex2()
		case OpSelector:
			m.doOpSelector()
		case OpSlice:
			m.doOpSlice()
		case OpStar:
			m.doOpStar()
		case OpRef:
			m.doOpRef()
		case OpTypeAssert1:
			m.doOpTypeAssert1()
		case OpTypeAssert2:
			m.doOpTypeAssert2()
		case OpStaticTypeOf:
			m.doOpStaticTypeOf()
		case OpCompositeLit:
			m.doOpCompositeLit()
		case OpArrayLit:
			m.doOpArrayLit()
		case OpSliceLit:
			m.doOpSliceLit()
		case OpFuncLit:
			m.doOpFuncLit()
		case OpMapLit:
			m.doOpMapLit()
		case OpStructLit:
			m.doOpStructLit()
		case OpConvert:
			m.doOpConvert()
		/* GoNative Operators */
		case OpStructLitGoNative:
			m.doOpStructLitGoNative()
		case OpCallGoNative:
			m.doOpCallGoNative()
		/* Type operators */
		case OpFieldType:
			m.doOpFieldType()
		case OpArrayType:
			m.doOpArrayType()
		case OpSliceType:
			m.doOpSliceType()
		case OpChanType:
			m.doOpChanType()
		case OpFuncType:
			m.doOpFuncType()
		case OpMapType:
			m.doOpMapType()
		case OpStructType:
			m.doOpStructType()
		case OpInterfaceType:
			m.doOpInterfaceType()
		/* Statement operators */
		case OpAssign:
			m.doOpAssign()
		case OpAddAssign:
			m.doOpAddAssign()
		case OpSubAssign:
			m.doOpSubAssign()
		case OpMulAssign:
			m.doOpMulAssign()
		case OpQuoAssign:
			m.doOpQuoAssign()
		case OpRemAssign:
			m.doOpRemAssign()
		case OpBandAssign:
			m.doOpBandAssign()
		case OpBandnAssign:
			m.doOpBandnAssign()
		case OpBorAssign:
			m.doOpBorAssign()
		case OpXorAssign:
			m.doOpXorAssign()
		case OpShlAssign:
			m.doOpShlAssign()
		case OpShrAssign:
			m.doOpShrAssign()
		case OpDefine:
			m.doOpDefine()
		case OpInc:
			m.doOpInc()
		case OpDec:
			m.doOpDec()
		/* Decl operators */
		case OpValueDecl:
			m.doOpValueDecl()
		case OpTypeDecl:
			m.doOpTypeDecl()
		/* Loop (sticky) operators */
		case OpBody:
			m.doOpExec(op)
		case OpForLoop:
			m.doOpExec(op)
		case OpRangeIter, OpRangeIterArrayPtr:
			m.doOpExec(op)
		case OpRangeIterString:
			m.doOpExec(op)
		case OpRangeIterMap:
			m.doOpExec(op)
		case OpReturnCallDefers:
			m.doOpReturnCallDefers()
		default:
			panic(fmt.Sprintf("unexpected opcode %s", op.String()))
		}
	}
}

//----------------------------------------
// push pop methods.

func (m *Machine) PushOp(op Op) {
	if debug {
		m.Printf("+o %v\n", op)
	}
	m.Ops[m.NumOps] = op
	m.NumOps++
}

func (m *Machine) PopOp() Op {
	numOps := m.NumOps
	op := m.Ops[numOps-1]
	if debug {
		m.Printf("-o %v\n", op)
	}
	if OpSticky <= op {
		// do not pop persistent op types.
	} else {
		m.NumOps--
	}
	return op
}

func (m *Machine) ForcePopOp() {
	if debug {
		m.Printf("-o! %v\n", m.Ops[m.NumOps-1])
	}
	m.NumOps--
}

// Offset starts at 1.
// DEPRECATED use PeekStmt1() instead.
func (m *Machine) PeekStmt(offset int) Stmt {
	if debug {
		if offset != 1 {
			panic("should not happen")
		}
	}
	return m.Stmts[len(m.Stmts)-offset]
}

func (m *Machine) PeekStmt1() Stmt {
	numStmts := len(m.Stmts)
	s := m.Stmts[numStmts-1]
	if bs, ok := s.(*bodyStmt); ok {
		return bs.Active
	} else {
		return m.Stmts[numStmts-1]
	}
}

func (m *Machine) PushStmt(s Stmt) {
	if debug {
		m.Printf("+s %v\n", s)
	}
	m.Stmts = append(m.Stmts, s)
}

func (m *Machine) PushStmts(ss ...Stmt) {
	if debug {
		for _, s := range ss {
			m.Printf("+s %v\n", s)
		}
	}
	m.Stmts = append(m.Stmts, ss...)
}

func (m *Machine) PopStmt() Stmt {
	numStmts := len(m.Stmts)
	s := m.Stmts[numStmts-1]
	if debug {
		m.Printf("-s %v\n", s)
	}
	if bs, ok := s.(*bodyStmt); ok {
		return bs.PopActiveStmt()
	} else {
		// general case.
		m.Stmts = m.Stmts[:numStmts-1]
		return s
	}
}

func (m *Machine) ForcePopStmt() (s Stmt) {
	numStmts := len(m.Stmts)
	s = m.Stmts[numStmts-1]
	if debug {
		m.Printf("-s %v\n", s)
	}
	// TODO debug lines and assertions.
	m.Stmts = m.Stmts[:len(m.Stmts)-1]
	return
}

// Offset starts at 1.
func (m *Machine) PeekExpr(offset int) Expr {
	return m.Exprs[len(m.Exprs)-offset]
}

func (m *Machine) PushExpr(x Expr) {
	if debug {
		m.Printf("+x %v\n", x)
	}
	m.Exprs = append(m.Exprs, x)
}

func (m *Machine) PopExpr() Expr {
	numExprs := len(m.Exprs)
	x := m.Exprs[numExprs-1]
	if debug {
		m.Printf("-x %v\n", x)
	}
	m.Exprs = m.Exprs[:numExprs-1]
	return x
}

// Returns reference to value in Values stack.  Offset starts at 1.
func (m *Machine) PeekValue(offset int) *TypedValue {
	return &m.Values[m.NumValues-offset]
}

// XXX delete?
func (m *Machine) PeekType(offset int) Type {
	return m.Values[m.NumValues-offset].T
}

func (m *Machine) PushValue(tv TypedValue) {
	if debug {
		m.Printf("+v %s\n", tv.String())
	}
	if len(m.Values) == m.NumValues {
		// TODO tune.
		newValues := make([]TypedValue, len(m.Values)*2)
		copy(newValues, m.Values)
		m.Values = newValues
	}
	m.Values[m.NumValues] = tv
	m.NumValues++
	return
}

// Resulting reference is volatile.
func (m *Machine) PopValue() (tv *TypedValue) {
	tv = &m.Values[m.NumValues-1]
	if debug {
		m.Printf("-v %s\n", tv.String())
	}
	m.NumValues--
	return tv
}

// Returns a slice of n values in the stack and decrements NumValues.
// NOTE: The results are on the values stack, so they must be copied or used
// immediately.  If you need to use the machine before or during usage,
// consider using PopCopyValues().
// NOTE: the values are in stack order, oldest first, the opposite order of
// multiple pop calls.  This is used for params assignment, for example.
func (m *Machine) PopValues(n int) []TypedValue {
	if debug {
		for i := 0; i < n; i++ {
			tv := m.Values[m.NumValues-n+i]
			m.Printf("-vs[%d/%d] %s\n", i, n, tv.String())
		}
	}
	m.NumValues -= n
	return m.Values[m.NumValues : m.NumValues+n]
}

// Like PopValues(), but copies the values onto a new slice.
func (m *Machine) PopCopyValues(n int) []TypedValue {
	res := make([]TypedValue, n)
	ptvs := m.PopValues(n)
	copy(res, ptvs)
	return res
}

// Decrements NumValues by number of last results.
func (m *Machine) PopResults() {
	if debug {
		for i := 0; i < m.NumResults; i++ {
			m.PopValue()
		}
	} else {
		m.NumValues -= m.NumResults
	}
	m.NumResults = 0
}

// Pops values with index start or greater.
func (m *Machine) ReapValues(start int) []TypedValue {
	end := m.NumValues
	rs := make([]TypedValue, end-start)
	copy(rs, m.Values[start:end])
	m.NumValues = start
	return rs
}

func (m *Machine) PushBlock(b *Block) {
	if debug {
		m.Println("+B")
	}
	m.Blocks = append(m.Blocks, b)
}

func (m *Machine) PopBlock() (b *Block) {
	if debug {
		m.Println("-B")
	}
	numBlocks := len(m.Blocks)
	b = m.Blocks[numBlocks-1]
	m.Blocks = m.Blocks[:numBlocks-1]
	return b
}

// The result is a volatile reference in the machine's type stack.
// Mutate and forget.
func (m *Machine) LastBlock() *Block {
	return m.Blocks[len(m.Blocks)-1]
}

// Pushes a frame with one less statement.
func (m *Machine) PushFrameBasic(s Stmt) {
	label := s.GetAttribute(ATTR_LABEL)
	lname := Name("")
	if label != nil {
		lname = label.(Name)
	}
	fr := Frame{
		Label:     lname,
		Source:    s,
		NumOps:    m.NumOps,
		NumValues: m.NumValues,
		NumExprs:  len(m.Exprs),
		NumStmts:  len(m.Stmts),
		NumBlocks: len(m.Blocks),
	}
	if debug {
		m.Printf("+F %#v\n", fr)
	}
	m.Frames = append(m.Frames, fr)
}

// TODO: track breaks/panics/returns on frame and
// ensure the counts are consistent, otherwise we mask
// bugs with frame pops.
func (m *Machine) PushFrameCall(cx *CallExpr, fv *FuncValue, recv TypedValue) {
	fr := Frame{
		Source:      cx,
		NumOps:      m.NumOps,
		NumValues:   m.NumValues,
		NumExprs:    len(m.Exprs),
		NumStmts:    len(m.Stmts),
		NumBlocks:   len(m.Blocks),
		Func:        fv,
		GoFunc:      nil,
		Receiver:    recv,
		NumArgs:     cx.NumArgs,
		IsVarg:      cx.Varg,
		Defers:      nil,
		LastPackage: m.Package,
		LastRealm:   m.Realm,
	}
	if debug {
		if m.Package == nil {
			panic("should not happen")
		}
	}
	if debug {
		m.Printf("+F %#v\n", fr)
	}
	m.Frames = append(m.Frames, fr)
	pkg := fv.GetPackage()
	if debug {
		if pkg == nil {
			panic("should not happen")
		}
	}
	m.Package = pkg
	rlm := pkg.GetRealm()
	if rlm != nil && m.Realm != rlm {
		m.Realm = rlm // enter new realm
	}
}

func (m *Machine) PushFrameGoNative(cx *CallExpr, fv *nativeValue) {
	fr := Frame{
		Source:      cx,
		NumOps:      m.NumOps,
		NumValues:   m.NumValues,
		NumExprs:    len(m.Exprs),
		NumStmts:    len(m.Stmts),
		NumBlocks:   len(m.Blocks),
		Func:        nil,
		GoFunc:      fv,
		Receiver:    TypedValue{},
		NumArgs:     cx.NumArgs,
		IsVarg:      cx.Varg,
		Defers:      nil,
		LastPackage: m.Package,
		LastRealm:   m.Realm,
	}
	if debug {
		m.Printf("+F %#v\n", fr)
	}
	m.Frames = append(m.Frames, fr)
	// keep m.Package the same.
}

func (m *Machine) PopFrame() Frame {
	numFrames := len(m.Frames)
	f := m.Frames[numFrames-1]
	if debug {
		m.Printf("-F %#v\n", f)
	}
	m.Frames = m.Frames[:numFrames-1]
	return f
}

func (m *Machine) PopFrameAndReset() {
	fr := m.PopFrame()
	m.NumOps = fr.NumOps
	m.NumValues = fr.NumValues
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts]
	m.Blocks = m.Blocks[:fr.NumBlocks]
	m.PopStmt() // may be sticky
}

// TODO: optimize by passing in last frame.
func (m *Machine) PopFrameAndReturn() {
	fr := m.PopFrame()
	if debug {
		// TODO: optimize with fr.IsCall
		if fr.Func == nil && fr.GoFunc == nil {
			panic("unexpected non-call (loop) frame")
		}
	}
	rtypes := fr.Func.GetType(m.Store).Results
	numRes := len(rtypes)
	m.NumOps = fr.NumOps
	m.NumResults = numRes
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts]
	m.Blocks = m.Blocks[:fr.NumBlocks]
	// shift and convert results to typed-nil if undefined and not iface kind.
	// and not func result type isn't interface kind.
	resStart := m.NumValues - numRes
	for i := 0; i < numRes; i++ {
		res := m.Values[resStart+i]
		if res.IsUndefined() && rtypes[i].Type.Kind() != InterfaceKind {
			res.T = rtypes[i].Type
		}
		m.Values[fr.NumValues+i] = res
	}
	m.NumValues = fr.NumValues + numRes
	m.Package = fr.LastPackage
}

func (m *Machine) PeekFrameAndContinueFor() {
	fr := m.LastFrame()
	m.NumOps = fr.NumOps + 1
	m.NumValues = fr.NumValues
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts+1]
	m.Blocks = m.Blocks[:fr.NumBlocks+1]
	ls := m.PeekStmt(1).(*bodyStmt)
	ls.NextBodyIndex = ls.BodyLen
}

func (m *Machine) PeekFrameAndContinueRange() {
	fr := m.LastFrame()
	m.NumOps = fr.NumOps + 1
	m.NumValues = fr.NumValues + 1
	m.Exprs = m.Exprs[:fr.NumExprs]
	m.Stmts = m.Stmts[:fr.NumStmts+1]
	m.Blocks = m.Blocks[:fr.NumBlocks+1]
	ls := m.PeekStmt(1).(*bodyStmt)
	ls.NextBodyIndex = ls.BodyLen
}

func (m *Machine) NumFrames() int {
	return len(m.Frames)
}

func (m *Machine) LastFrame() *Frame {
	return &m.Frames[len(m.Frames)-1]
}

// TODO: this function and PopUntilLastCallFrame() is used in conjunction
// spanning two disjoint operations upon return. Optimize.
func (m *Machine) LastCallFrame() *Frame {
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if fr.Func != nil || fr.GoFunc != nil {
			// TODO: optimize with fr.IsCall
			return fr
		}
	}
	panic("missing call frame")
}

// pops the last non-call (loop) frames
// and returns the last call frame (which is left on stack).
func (m *Machine) PopUntilLastCallFrame() *Frame {
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if fr.Func != nil || fr.GoFunc != nil {
			// TODO: optimize with fr.IsCall
			m.Frames = m.Frames[:i+1]
			return fr
		}
	}
	panic("missing call frame")
}

func (m *Machine) PushForPointer(lx Expr) {
	switch lx := lx.(type) {
	case *NameExpr:
		// no Lhs eval needed.
	case *IndexExpr:
		// evaluate Index
		m.PushExpr(lx.Index)
		m.PushOp(OpEval)
		// evaluate X
		m.PushExpr(lx.X)
		m.PushOp(OpEval)
	case *SelectorExpr:
		// evaluate X
		m.PushExpr(lx.X)
		m.PushOp(OpEval)
	case *StarExpr:
		// evaluate X (a reference)
		m.PushExpr(lx.X)
		m.PushOp(OpEval)
	case *CompositeLitExpr: // for *RefExpr e.g. &mystruct{}
		// evaluate lx.
		m.PushExpr(lx)
		m.PushOp(OpEval)
	default:
		panic(fmt.Sprintf(
			"illegal assignment X expression type %v",
			reflect.TypeOf(lx)))
	}
}

func (m *Machine) PopAsPointer(lx Expr) PointerValue {
	switch lx := lx.(type) {
	case *NameExpr:
		lb := m.LastBlock()
		return lb.GetPointerTo(m.Store, lx.Path)
	case *IndexExpr:
		iv := m.PopValue()
		xv := m.PopValue()
		return xv.GetPointerAtIndex(m.Store, iv)
	case *SelectorExpr:
		xv := m.PopValue()
		return xv.GetPointerTo(m.Store, lx.Path)
	case *StarExpr:
		ptr := m.PopValue().V.(PointerValue)
		return ptr
	case *CompositeLitExpr: // for *RefExpr
		tv := *m.PopValue()
		return PointerValue{
			TV:   &tv, // heap alloc
			Base: nil,
		}
	default:
		panic("should not happen")
	}
}

// for testing.
func (m *Machine) CheckEmpty() error {
	found := ""
	if m.NumOps > 0 {
		found = "op"
	} else if m.NumValues > 0 {
		found = "value"
	} else if len(m.Exprs) > 0 {
		found = "expr"
	} else if len(m.Stmts) > 0 {
		found = "stmt"
	} else if len(m.Blocks) > 0 {
		for _, b := range m.Blocks {
			_, isPkg := b.Source.(*PackageNode)
			if isPkg {
				// ok
			} else {
				found = "(non-package) block"
			}
		}
	} else if len(m.Frames) > 0 {
		found = "frame"
	} else if m.NumResults > 0 {
		found = ".NumResults != 0"
	}
	if found != "" {
		return fmt.Errorf("found leftover %s", found)
	} else {
		return nil
	}
}

func (m *Machine) Panic(ex TypedValue) {
	// TODO: chain exceptions if preexisting unrecovered exception.
	m.Exception = &ex
	m.PopUntilLastCallFrame()
	m.PushOp(OpPanic2)
	m.PushOp(OpReturnCallDefers)
}

//----------------------------------------
// inspection methods

func (m *Machine) Println(args ...interface{}) {
	if debug {
		s := strings.Repeat("|", m.NumOps)
		fmt.Println(append([]interface{}{"DEBUG:", s}, args...)...)
	}
}

func (m *Machine) Printf(format string, args ...interface{}) {
	if debug {
		s := strings.Repeat("|", m.NumOps)
		fmt.Printf("DEBUG: "+s+" "+format, args...)
	}
}

func (m *Machine) String() string {
	vs := []string{}
	for i := m.NumValues - 1; i >= 0; i-- {
		v := m.Values[i]
		vs = append(vs, fmt.Sprintf("          #%d %v", i, v))
	}
	ss := []string{}
	for i := len(m.Stmts) - 1; i >= 0; i-- {
		s := m.Stmts[i]
		ss = append(ss, fmt.Sprintf("          #%d %v", i, s))
	}
	xs := []string{}
	for i := len(m.Exprs) - 1; i >= 0; i-- {
		x := m.Exprs[i]
		xs = append(xs, fmt.Sprintf("          #%d %v", i, x))
	}
	bs := []string{}
	for b := m.LastBlock(); b != nil; {
		gen := len(bs)/2 + 1
		gens := strings.Repeat("@", gen)
		bsi := b.StringIndented("            ")
		bs = append(bs, fmt.Sprintf("          %s(%d) %s", gens, gen, bsi))
		if b.Source != nil {
			sb := b.Source.GetStaticBlock().GetBlock()
			bs = append(bs, fmt.Sprintf(" (static values) %s(%d) %s", gens, gen,
				sb.StringIndented("            ")))
			sts := b.Source.GetStaticBlock().Types
			bs = append(bs, fmt.Sprintf(" (static types) %s(%d) %s", gens, gen, sts))
		}
		// b = b.Parent.(*Block|RefValue)
		switch bp := b.Parent.(type) {
		case nil:
			b = nil
			break
		case *Block:
			b = bp
		case RefValue:
			bs = append(bs, fmt.Sprintf("            (block ref %v)", bp.ObjectID))
			b = nil
			break
		default:
			panic("should not happen")
		}
	}
	obs := []string{}
	for i := len(m.Blocks) - 2; i >= 0; i-- {
		b := m.Blocks[i]
		obs = append(obs, fmt.Sprintf("          #%d %s", i,
			b.StringIndented("            ")))
		if b.Source != nil {
			sb := b.Source.GetStaticBlock().GetBlock()
			obs = append(obs, fmt.Sprintf(" (static) #%d %s", i,
				sb.StringIndented("            ")))
		}
	}
	fs := []string{}
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := m.Frames[i]
		fs = append(fs, fmt.Sprintf("          #%d %s", i, fr.String()))
	}
	return fmt.Sprintf(`Machine:
    CheckTypes: %v
	Op: %v
	Values: (len: %d)
%s
	Exprs:
%s
	Stmts:
%s
	Blocks:
%s
	Blocks (other):
%s
	Frames:
%s
	Exception:
	  %s`,
		m.CheckTypes,
		m.Ops[:m.NumOps],
		m.NumValues,
		strings.Join(vs, "\n"),
		strings.Join(xs, "\n"),
		strings.Join(ss, "\n"),
		strings.Join(bs, "\n"),
		strings.Join(obs, "\n"),
		strings.Join(fs, "\n"),
		m.Exception,
	)
}

//----------------------------------------

func hasName(n Name, ns []Name) bool {
	for _, n2 := range ns {
		if n == n2 {
			return true
		}
	}
	return false
}
