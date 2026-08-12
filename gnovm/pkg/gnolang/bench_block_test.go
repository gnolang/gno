package gnolang

import (
	"testing"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
)

// Calibration benchmarks for the block-creation gas split (see acquireBlock
// and the ADR). Block creation is charged separately from the enclosing op:
// OpCPUAcquireBlock on every acquire, plus AllocateBlock on a pool miss. The
// enclosing-op constants (OpCPUCall, OpCPUReturnCallDefers) were therefore
// re-derived to EXCLUDE block creation. Reference hardware is Xeon 8168
// (1 gas = 1 ns); on other hardware, anchor to BenchmarkOpAdd_Int (Xeon=81):
//
//	ratio                  = 81 / Add_Int(ns)
//	OpCPUAcquireBlock      = recover(hit) ns                       * ratio
//	OpCPUCall (sans block) = (OpCallWarm - recover) ns             * ratio
//	OpCPUReturnCallDefers  = (OpReturnCallDefersWarm - recover) ns * ratio
//
// The "warm" variants pre-fill the pool so acquireBlock hits cheaply, leaving
// only the recover cost to subtract (isolating the non-block op cost).

func benchAcquireSrc(numNames int) *BlockStmt {
	src := &BlockStmt{}
	src.NumNames = uint16(numNames)
	src.HeapItems = make([]bool, numNames)
	return src
}

func warmPool(m *Machine) {
	for range 8 {
		m.blockPool = append(m.blockPool, &Block{Values: make([]TypedValue, 0, blockPoolValueCap)})
	}
}

// returnBlock recycles the top block back into the warm pool (not measured).
func returnBlock(m *Machine, blk *Block) {
	m.Blocks = m.Blocks[:0]
	m.Ops = m.Ops[:0]
	m.Stmts = m.Stmts[:0]
	m.Values = m.Values[:0]
	m.Frames = m.Frames[:0]
	vals := blk.Values[:cap(blk.Values)]
	clear(vals)
	*blk = Block{Values: vals[:0]}
	m.blockPool = append(m.blockPool, blk)
}

func BenchmarkOpAcquireBlockRecycle(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	src := benchAcquireSrc(2)
	var parent Block
	warmPool(m)
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		bm.SwitchOpCode(bmTarget)
		blk := m.acquireBlock(src, &parent) // hit
		bm.SwitchOpCode(bmSetup)
		vals := blk.Values[:cap(blk.Values)]
		clear(vals)
		*blk = Block{Values: vals[:0]}
		m.blockPool = append(m.blockPool, blk)
	}
	reportBenchops(b)
}

func BenchmarkOpCallWarm(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	ft := &FuncType{Params: []FieldType{}, Results: []FieldType{}}
	fd := benchFuncDeclNode(0, nil)
	fv := &FuncValue{Type: ft, IsClosure: true, Source: fd, PkgPath: "bench", body: []Stmt{}}
	cx := &CallExpr{NumArgs: 0}
	warmPool(m)
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		bm.SwitchOpCode(bmTarget)
		m.doOpCall() // acquireBlock hits
		bm.SwitchOpCode(bmSetup)
		returnBlock(m, m.Blocks[len(m.Blocks)-1])
	}
	reportBenchops(b)
}

func BenchmarkOpReturnCallDefersWarm(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	ft := &FuncType{Params: []FieldType{}, Results: []FieldType{}}
	fd := benchFuncDeclNode(0, nil)
	fv := &FuncValue{Type: ft, IsClosure: true, Source: fd, PkgPath: "bench", body: []Stmt{}}
	cx := &CallExpr{NumArgs: 0}
	warmPool(m)
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		m.Blocks = append(m.Blocks, &Block{})
		cfr := m.LastFrame()
		cfr.PushDefer(Defer{
			Callable: fv,
			Args:     []TypedValue{},
			Source:   &DeferStmt{Call: CallExpr{NumArgs: 0, Args: []Expr{}}},
			Parent:   &Block{},
		})
		m.PushOp(OpReturnCallDefers)
		bm.SwitchOpCode(bmTarget)
		m.doOpReturnCallDefers() // acquireBlock hits
		bm.SwitchOpCode(bmSetup)
		returnBlock(m, m.Blocks[len(m.Blocks)-1])
	}
	reportBenchops(b)
}
