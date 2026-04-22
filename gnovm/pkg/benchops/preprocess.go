package benchops

import "time"

// PreprocessOp represents a preprocess operation for benchmarking.
// One code is emitted per TransStage (TRANS_ENTER/BLOCK/BLOCK2/LEAVE)
// visit of each relevant node type. Used to derive preprocess gas
// costs.
type PreprocessOp byte

// preprocess code
const (
	PreprocessOpInvalid PreprocessOp = 0x00 // invalid

	// generic stage codes (used when no node-specific code applies)
	PreprocessEnter  PreprocessOp = 0x01
	PreprocessBlock  PreprocessOp = 0x02
	PreprocessBlock2 PreprocessOp = 0x03
	PreprocessLeave  PreprocessOp = 0x04

	// TRANS_ENTER nodes with specific per-type enter-stage work in
	// preprocess1.
	PreprocessEnterAssignStmt   PreprocessOp = 0x05
	PreprocessEnterImportDecl   PreprocessOp = 0x06
	PreprocessEnterValueDecl    PreprocessOp = 0x07
	PreprocessEnterTypeDecl     PreprocessOp = 0x08
	PreprocessEnterFuncDecl     PreprocessOp = 0x09
	PreprocessEnterFuncTypeExpr PreprocessOp = 0x0A

	// TRANS_BLOCK nodes.
	PreprocessBlockBlockStmt        PreprocessOp = 0x0B
	PreprocessBlockForStmt          PreprocessOp = 0x0C
	PreprocessBlockIfStmt           PreprocessOp = 0x0D
	PreprocessBlockIfCaseStmt       PreprocessOp = 0x0E
	PreprocessBlockRangeStmt        PreprocessOp = 0x0F
	PreprocessBlockFuncLitExpr      PreprocessOp = 0x10
	PreprocessBlockSwitchStmt       PreprocessOp = 0x11
	PreprocessBlockSwitchClauseStmt PreprocessOp = 0x12
	PreprocessBlockFuncDecl         PreprocessOp = 0x13
	PreprocessBlockFileNode         PreprocessOp = 0x14

	// TRANS_BLOCK2 nodes.
	PreprocessBlock2SwitchStmt PreprocessOp = 0x15

	// TRANS_LEAVE nodes with specific per-type leave-stage work.
	PreprocessLeaveNameExpr          PreprocessOp = 0x16
	PreprocessLeaveBasicLitExpr      PreprocessOp = 0x17
	PreprocessLeaveBinaryExpr        PreprocessOp = 0x18
	PreprocessLeaveCallExpr          PreprocessOp = 0x19
	PreprocessLeaveIndexExpr         PreprocessOp = 0x1A
	PreprocessLeaveSliceExpr         PreprocessOp = 0x1B
	PreprocessLeaveTypeAssertExpr    PreprocessOp = 0x1C
	PreprocessLeaveUnaryExpr         PreprocessOp = 0x1D
	PreprocessLeaveCompositeLitExpr  PreprocessOp = 0x1E
	PreprocessLeaveStarExpr          PreprocessOp = 0x1F
	PreprocessLeaveSelectorExpr      PreprocessOp = 0x20
	PreprocessLeaveFieldTypeExpr     PreprocessOp = 0x21
	PreprocessLeaveArrayTypeExpr     PreprocessOp = 0x22
	PreprocessLeaveSliceTypeExpr     PreprocessOp = 0x23
	PreprocessLeaveInterfaceTypeExpr PreprocessOp = 0x24
	PreprocessLeaveFuncTypeExpr      PreprocessOp = 0x25
	PreprocessLeaveMapTypeExpr       PreprocessOp = 0x26
	PreprocessLeaveStructTypeExpr    PreprocessOp = 0x27
	PreprocessLeaveAssignStmt        PreprocessOp = 0x28
	PreprocessLeaveBranchStmt        PreprocessOp = 0x29
	PreprocessLeaveIncDecStmt        PreprocessOp = 0x2A
	PreprocessLeaveForStmt           PreprocessOp = 0x2B
	PreprocessLeaveIfStmt            PreprocessOp = 0x2C
	PreprocessLeaveRangeStmt         PreprocessOp = 0x2D
	PreprocessLeaveReturnStmt        PreprocessOp = 0x2E
	PreprocessLeaveSwitchStmt        PreprocessOp = 0x2F
	PreprocessLeaveValueDecl         PreprocessOp = 0x30
	PreprocessLeaveTypeDecl          PreprocessOp = 0x31

	// TRANS_LEAVE nodes without per-type work — each has a measured
	// cost distinct enough from the generic PreprocessLeave to warrant
	// its own code.
	PreprocessLeaveBlockStmt        PreprocessOp = 0x32
	PreprocessLeaveDeclStmt         PreprocessOp = 0x33
	PreprocessLeaveDeferStmt        PreprocessOp = 0x34
	PreprocessLeaveEmptyStmt        PreprocessOp = 0x35
	PreprocessLeaveExprStmt         PreprocessOp = 0x36
	PreprocessLeaveIfCaseStmt       PreprocessOp = 0x37
	PreprocessLeaveSwitchClauseStmt PreprocessOp = 0x38
	PreprocessLeaveFuncLitExpr      PreprocessOp = 0x39
	PreprocessLeaveFuncDecl         PreprocessOp = 0x3A
	PreprocessLeaveImportDecl       PreprocessOp = 0x3B
	PreprocessLeaveFileNode         PreprocessOp = 0x3C
	PreprocessLeaveRefExpr          PreprocessOp = 0x3D
	PreprocessLeaveConstExpr        PreprocessOp = 0x3E

	invalidPreprocessCode string = "PreprocessInvalid"
)

// index matches PreprocessOp value
var preprocessCodeNames = []string{
	invalidPreprocessCode,
	"PreprocessEnter",
	"PreprocessBlock",
	"PreprocessBlock2",
	"PreprocessLeave",
	"PreprocessEnterAssignStmt",
	"PreprocessEnterImportDecl",
	"PreprocessEnterValueDecl",
	"PreprocessEnterTypeDecl",
	"PreprocessEnterFuncDecl",
	"PreprocessEnterFuncTypeExpr",
	"PreprocessBlockBlockStmt",
	"PreprocessBlockForStmt",
	"PreprocessBlockIfStmt",
	"PreprocessBlockIfCaseStmt",
	"PreprocessBlockRangeStmt",
	"PreprocessBlockFuncLitExpr",
	"PreprocessBlockSwitchStmt",
	"PreprocessBlockSwitchClauseStmt",
	"PreprocessBlockFuncDecl",
	"PreprocessBlockFileNode",
	"PreprocessBlock2SwitchStmt",
	"PreprocessLeaveNameExpr",
	"PreprocessLeaveBasicLitExpr",
	"PreprocessLeaveBinaryExpr",
	"PreprocessLeaveCallExpr",
	"PreprocessLeaveIndexExpr",
	"PreprocessLeaveSliceExpr",
	"PreprocessLeaveTypeAssertExpr",
	"PreprocessLeaveUnaryExpr",
	"PreprocessLeaveCompositeLitExpr",
	"PreprocessLeaveStarExpr",
	"PreprocessLeaveSelectorExpr",
	"PreprocessLeaveFieldTypeExpr",
	"PreprocessLeaveArrayTypeExpr",
	"PreprocessLeaveSliceTypeExpr",
	"PreprocessLeaveInterfaceTypeExpr",
	"PreprocessLeaveFuncTypeExpr",
	"PreprocessLeaveMapTypeExpr",
	"PreprocessLeaveStructTypeExpr",
	"PreprocessLeaveAssignStmt",
	"PreprocessLeaveBranchStmt",
	"PreprocessLeaveIncDecStmt",
	"PreprocessLeaveForStmt",
	"PreprocessLeaveIfStmt",
	"PreprocessLeaveRangeStmt",
	"PreprocessLeaveReturnStmt",
	"PreprocessLeaveSwitchStmt",
	"PreprocessLeaveValueDecl",
	"PreprocessLeaveTypeDecl",
	"PreprocessLeaveBlockStmt",
	"PreprocessLeaveDeclStmt",
	"PreprocessLeaveDeferStmt",
	"PreprocessLeaveEmptyStmt",
	"PreprocessLeaveExprStmt",
	"PreprocessLeaveIfCaseStmt",
	"PreprocessLeaveSwitchClauseStmt",
	"PreprocessLeaveFuncLitExpr",
	"PreprocessLeaveFuncDecl",
	"PreprocessLeaveImportDecl",
	"PreprocessLeaveFileNode",
	"PreprocessLeaveRefExpr",
	"PreprocessLeaveConstExpr",
}

// PreprocessCodeString returns the name for a preprocess code.
// Used in gas descriptions and benchmark output.
func PreprocessCodeString(code PreprocessOp) string {
	if int(code) >= len(preprocessCodeNames) {
		return invalidPreprocessCode
	}
	return preprocessCodeNames[code]
}

// ---- Preprocess timing ----
//
// Independent from the VM op timeline. The stack supports recursive
// Preprocess calls (sub-tree re-preprocess during constant folding,
// etc.).

var (
	preprocessCounts   [256]int64
	preprocessAccumDur [256]time.Duration
	preprocessStack    []preprocessFrame
)

// preprocessFrame is one stack entry: the current code and its
// resume time.
type preprocessFrame struct {
	code  PreprocessOp
	start time.Time
}

// StartPreprocess begins timing a preprocess op, pausing any outer
// op currently being timed.
func StartPreprocess(code PreprocessOp) {
	if code == PreprocessOpInvalid {
		panic("the PreprocessOp is invalid")
	}
	now := time.Now()
	// Pause outer preprocess measurement, if any.
	if n := len(preprocessStack); n > 0 {
		outer := &preprocessStack[n-1]
		if outer.start != measure.timeZero {
			preprocessAccumDur[outer.code] += now.Sub(outer.start)
			outer.start = measure.timeZero
		}
	}
	preprocessStack = append(preprocessStack, preprocessFrame{
		code:  code,
		start: now,
	})
	preprocessCounts[code]++
}

// StopPreprocess finalizes the current preprocess op. Any outer
// paused preprocess op is resumed.
func StopPreprocess() {
	now := time.Now()
	n := len(preprocessStack)
	if n == 0 {
		return
	}
	f := preprocessStack[n-1]
	preprocessStack = preprocessStack[:n-1]
	if f.start != measure.timeZero {
		preprocessAccumDur[f.code] += now.Sub(f.start)
	}
	// Resume outer frame.
	if n2 := len(preprocessStack); n2 > 0 {
		preprocessStack[n2-1].start = now
	}
}

// PreprocessAccumDur returns the accumulated duration for a
// preprocess code.
func PreprocessAccumDur(code PreprocessOp) time.Duration {
	return preprocessAccumDur[code]
}

// PreprocessCount returns the invocation count for a preprocess code.
func PreprocessCount(code PreprocessOp) int64 {
	return preprocessCounts[code]
}

// ResetPreprocess clears all preprocess timing state. Typically
// called from a benchmark's setup before b.ResetTimer().
func ResetPreprocess() {
	preprocessCounts = [256]int64{}
	preprocessAccumDur = [256]time.Duration{}
	preprocessStack = nil
}
