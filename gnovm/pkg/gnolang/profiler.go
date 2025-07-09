package gnolang

import (
	"fmt"
	"strings"

	// "github.com/google/pprof/profile"
)

type OpStats struct {
	Count         int64 // number of times the opcode has been executed
	CumulativeCPU int64 // total "CPU" cycles consumed by this opcode
	Allocations   int64 // total memory allocated in bytes by this opcode
}

type Position struct {
	File       string // source file name
	Line       int    // source line number
	FuncSource string // name of the function where the opcode occurs
}

type OpTrace struct {
	Op        Op        // the opcode executed
	IndexOp   int       // the index of the opcode in the program
	FuncName  string    // the function name in which the opcode was executed
	SourcePos Position  // position in the source code (file + line)
}

var opStats = make(map[Op]*OpStats) // maps Op → stats
var traces []OpTrace                // list of all opcode executions traced


// getOpCPUCost returns the number of "CPU" cycles associated with a given opcode.
func (m *Machine) getOpCPUCost(op Op) int64 {
	switch op {
	case OpInvalid:
		return OpCPUInvalid
	case OpHalt:
		return OpCPUHalt
	case OpNoop:
		return OpCPUNoop
	case OpExec:
		return OpCPUExec
	case OpPrecall:
		return OpCPUPrecall
	case OpCall:
		return OpCPUCall
	case OpCallNativeBody:
		return OpCPUCallNativeBody
	case OpDefer:
		return OpCPUDefer
	case OpCallDeferNativeBody:
		return OpCPUCallDeferNativeBody
	case OpGo:
		return OpCPUGo
	case OpSelect:
		return OpCPUSelect
	case OpSwitchClause:
		return OpCPUSwitchClause
	case OpSwitchClauseCase:
		return OpCPUSwitchClauseCase
	case OpTypeSwitch:
		return OpCPUTypeSwitch
	case OpIfCond:
		return OpCPUIfCond
	case OpPopValue:
		return OpCPUPopValue
	case OpPopResults:
		return OpCPUPopResults
	case OpPopBlock:
		return OpCPUPopBlock
	case OpPopFrameAndReset:
		return OpCPUPopFrameAndReset
	case OpPanic1:
		return OpCPUPanic1
	case OpPanic2:
		return OpCPUPanic2
	case OpReturn:
		return OpCPUReturn
	case OpReturnAfterCopy:
		return OpCPUReturnAfterCopy
	case OpReturnFromBlock:
		return OpCPUReturnFromBlock
	case OpReturnToBlock:
		return OpCPUReturnToBlock
	case OpUpos:
		return OpCPUUpos
	case OpUneg:
		return OpCPUUneg
	case OpUnot:
		return OpCPUUnot
	case OpUxor:
		return OpCPUUxor
	case OpUrecv:
		return OpCPUUrecv
	case OpLor:
		return OpCPULor
	case OpLand:
		return OpCPULand
	case OpEql:
		return OpCPUEql
	case OpNeq:
		return OpCPUNeq
	case OpLss:
		return OpCPULss
	case OpLeq:
		return OpCPULeq
	case OpGtr:
		return OpCPUGtr
	case OpGeq:
		return OpCPUGeq
	case OpAdd:
		return OpCPUAdd
	case OpSub:
		return OpCPUSub
	case OpBor:
		return OpCPUBor
	case OpXor:
		return OpCPUXor
	case OpMul:
		return OpCPUMul
	case OpQuo:
		return OpCPUQuo
	case OpRem:
		return OpCPURem
	case OpShl:
		return OpCPUShl
	case OpShr:
		return OpCPUShr
	case OpBand:
		return OpCPUBand
	case OpBandn:
		return OpCPUBandn
	case OpEval:
		return OpCPUEval
	case OpBinary1:
		return OpCPUBinary1
	case OpIndex1:
		return OpCPUIndex1
	case OpIndex2:
		return OpCPUIndex2
	case OpSelector:
		return OpCPUSelector
	case OpSlice:
		return OpCPUSlice
	case OpStar:
		return OpCPUStar
	case OpRef:
		return OpCPURef
	case OpTypeAssert1:
		return OpCPUTypeAssert1
	case OpTypeAssert2:
		return OpCPUTypeAssert2
	case OpStaticTypeOf:
		return OpCPUStaticTypeOf
	case OpCompositeLit:
		return OpCPUCompositeLit
	case OpArrayLit:
		return OpCPUArrayLit
	case OpSliceLit:
		return OpCPUSliceLit
	case OpSliceLit2:
		return OpCPUSliceLit2
	case OpMapLit:
		return OpCPUMapLit
	case OpStructLit:
		return OpCPUStructLit
	case OpFuncLit:
		return OpCPUFuncLit
	case OpConvert:
		return OpCPUConvert
	case OpFieldType:
		return OpCPUFieldType
	case OpArrayType:
		return OpCPUArrayType
	case OpSliceType:
		return OpCPUSliceType
	case OpPointerType:
		return OpCPUPointerType
	case OpInterfaceType:
		return OpCPUInterfaceType
	case OpChanType:
		return OpCPUChanType
	case OpFuncType:
		return OpCPUFuncType
	case OpMapType:
		return OpCPUMapType
	case OpStructType:
		return OpCPUStructType
	case OpAssign:
		return OpCPUAssign
	case OpAddAssign:
		return OpCPUAddAssign
	case OpSubAssign:
		return OpCPUSubAssign
	case OpMulAssign:
		return OpCPUMulAssign
	case OpQuoAssign:
		return OpCPUQuoAssign
	case OpRemAssign:
		return OpCPURemAssign
	case OpBandAssign:
		return OpCPUBandAssign
	case OpBandnAssign:
		return OpCPUBandnAssign
	case OpBorAssign:
		return OpCPUBorAssign
	case OpXorAssign:
		return OpCPUXorAssign
	case OpShlAssign:
		return OpCPUShlAssign
	case OpShrAssign:
		return OpCPUShrAssign
	case OpDefine:
		return OpCPUDefine
	case OpInc:
		return OpCPUInc
	case OpDec:
		return OpCPUDec
	case OpValueDecl:
		return OpCPUValueDecl
	case OpTypeDecl:
		return OpCPUTypeDecl
	case OpSticky:
		return OpCPUSticky
	case OpBody:
		return OpCPUBody
	case OpForLoop:
		return OpCPUForLoop
	case OpRangeIter:
		return OpCPURangeIter
	case OpRangeIterString:
		return OpCPURangeIterString
	case OpRangeIterMap:
		return OpCPURangeIterMap
	case OpRangeIterArrayPtr:
		return OpCPURangeIterArrayPtr
	case OpReturnCallDefers:
		return OpCPUReturnCallDefers
	case OpVoid:
		return 0
	}
	return 1
}

// getOpAllocation returns the number of bytes allocated for a given opcode. (estimations)
// TODO: Replace with real allocation measurements from benchmarks for example
func (m *Machine) getOpAllocation(op Op) int64 {
	switch op {
	case OpArrayLit:
		return 64
	case OpSliceLit, OpSliceLit2:
		return 64
	case OpMapLit:
		return 128
	case OpStructLit:
		return 48
	case OpFuncLit:
		return 72
	case OpCompositeLit:
		return 56
	case OpInterfaceType:
		return 40
	case OpFuncType:
		return 48
	case OpFieldType:
		return 24
	case OpSliceType:
		return 32
	case OpMapType:
		return 48
	case OpStructType:
		return 56
	case OpArrayType:
		return 32
	case OpPointerType:
		return 24
	case OpChanType:
		return 32
	case OpEval:
		return 32
	case OpCall:
		return 80
	case OpCallNativeBody:
		return 96
	case OpPrecall:
		return 40
	case OpExec:
		return 48
	case OpDefine:
		return 32
	case OpAssign:
		return 16
	case OpConvert:
		return 24
	case OpTypeAssert1, OpTypeAssert2:
		return 32
	case OpStaticTypeOf:
		return 24
	case OpIndex1, OpIndex2:
		return 8
	case OpSelector:
		return 16
	case OpSlice:
		return 32
	case OpStar:
		return 8
	case OpRef:
		return 16
	case OpForLoop:
		return 24
	case OpBody:
		return 32
	case OpReturn, OpReturnFromBlock:
		return 16
	case OpReturnAfterCopy:
		return 24
	case OpPopBlock:
		return 8
	case OpPopResults:
		return 8
	case OpPopValue:
		return 8
	case OpAdd, OpSub, OpMul, OpQuo, OpRem:
		return 0
	case OpShl, OpShr, OpBand, OpBor, OpXor, OpBandn:
		return 0
	case OpUpos, OpUneg, OpUnot, OpUxor:
		return 0
	case OpEql, OpNeq, OpLss, OpLeq, OpGtr, OpGeq:
		return 0
	case OpLor, OpLand:
		return 0
	case OpHalt:
		return 0
	case OpNoop:
		return 0
	case OpVoid:
		return 0
	case OpAddAssign, OpSubAssign, OpMulAssign, OpQuoAssign, OpRemAssign:
		return 8
	case OpBandAssign, OpBandnAssign, OpBorAssign, OpXorAssign:
		return 8
	case OpShlAssign, OpShrAssign:
		return 8
	case OpInc, OpDec:
		return 0
	case OpValueDecl, OpTypeDecl:
		return 24
	case OpSticky:
		return 16
	default:
		return 8
	}
}

// UpdateOpStats is called each time an opcode is executed.
// It updates cumulative stats and stores trace data.
func (m *Machine) UpdateOpStats(op Op, index int, funcName string, pos Position) {
	// if op == OpInvalid {
	// 	return
	// }
	cost := m.getOpCPUCost(op)
	alloc := m.getOpAllocation(op)

	stat, ok := opStats[op]
	if !ok {
		stat = &OpStats{}
		opStats[op] = stat
	}
	stat.Count++
	stat.CumulativeCPU += cost
	stat.Allocations += alloc

	trace := OpTrace{
		Op:        op,
		IndexOp:   index,
		FuncName:  funcName,
		SourcePos: pos,
	}
	traces = append(traces, trace)

	if stat.Count <= 5 || stat.Count%100 == 0 {
		fmt.Printf("[TRACE] Opcode=%-18s | Index=%3d | Func=%-20s | %s:%d | CPU=%4d | Alloc=%d | Count=%d\n",
			op.String(), index, funcName, pos.File, pos.Line, cost, alloc, stat.Count)
	}
}

func (m *Machine) PrintOpStats() {
	fmt.Printf("\n%-20s | %-10s | %-15s | %-10s\n", "Opcode", "Count", "Cumulative CPU", "Alloc (KB)")
	fmt.Println(strings.Repeat("-", 70))

	var totalCount, totalCPU, totalAlloc int64
	for _, stat := range opStats {
		totalCount += stat.Count
		totalCPU += stat.CumulativeCPU
		totalAlloc += stat.Allocations
	}
	for op, stat := range opStats {
		allocKB := float64(stat.Allocations) / 1024
		fmt.Printf("%-20s | %-10d | %-15d | %-10.2f\n",
			op.String(), stat.Count, stat.CumulativeCPU, allocKB)
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-20s | %-10d | %-15d | %-10.2f\n",
		"TOTAL", totalCount, totalCPU, float64(totalAlloc)/1024)
}