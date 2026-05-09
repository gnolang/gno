package gnolang

import (
	"testing"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
)

// TestPreprocessDefaultGasCostInvariant guards the safety contract
// documented at preprocessDefaultGasCost: the fallback charged for an
// uncalibrated code in production must be >= the heaviest calibrated
// cost, otherwise a calibration gap becomes an under-charge DoS
// vector.
//
// init() in preprocess.go derives the fallback from the table; this
// test fails loudly if anyone ever turns it back into a hand-tuned
// constant lower than max(table).
func TestPreprocessDefaultGasCostInvariant(t *testing.T) {
	t.Parallel()

	var maxCost int64
	var maxName string
	for code, cost := range preprocessGasCosts {
		if int64(cost) > maxCost {
			maxCost = int64(cost)
			maxName = bm.PreprocessCodeString(bm.PreprocessOp(code))
		}
	}
	if maxCost == 0 {
		t.Fatal("preprocessGasCosts is empty — calibration table not loaded")
	}
	if int64(preprocessDefaultGasCost) < maxCost {
		t.Errorf(
			"preprocessDefaultGasCost (%d) < max(preprocessGasCosts) "+
				"(%d, from %s); fallback is cheaper than the heaviest "+
				"calibrated code, which would let an unmapped (stage,node) "+
				"pair under-charge gas relative to a real visit. "+
				"Adjust preprocessDefaultGasHeadroomBps or recalibrate.",
			preprocessDefaultGasCost, maxCost, maxName,
		)
	}
}

// TestPreprocessNodeCodeReachable asserts that every (TransStage, Node)
// pair preprocessNodeCode currently dispatches on returns a non-zero
// (i.e. calibrated, non-fallback) code. If a new node type is added to
// Transcribe without updating preprocessNodeCode + the cost table, this
// test will fail in non-debug builds (where the missing case silently
// falls through to a stage-generic code that may or may not be
// calibrated).
//
// We exercise just the cases that preprocessNodeCode actually maps,
// using one representative Node per type — the goal is not to find new
// missing cases (that's debugFind / preprocessNodeCode's debug panic),
// but to lock the current mapping against accidental zero-cost regressions.
func TestPreprocessNodeCodeAllMappedHaveCost(t *testing.T) {
	t.Parallel()

	type sample struct {
		stage TransStage
		node  Node
	}
	samples := []sample{
		// TRANS_ENTER specific.
		{TRANS_ENTER, &AssignStmt{}},
		{TRANS_ENTER, &ImportDecl{}},
		{TRANS_ENTER, &ValueDecl{}},
		{TRANS_ENTER, &TypeDecl{}},
		{TRANS_ENTER, &FuncDecl{}},
		{TRANS_ENTER, &FuncTypeExpr{}},
		// TRANS_ENTER grouped (PreprocessEnter).
		{TRANS_ENTER, &NameExpr{}},
		{TRANS_ENTER, &BasicLitExpr{}},
		// TRANS_BLOCK.
		{TRANS_BLOCK, &BlockStmt{}},
		{TRANS_BLOCK, &ForStmt{}},
		{TRANS_BLOCK, &IfStmt{}},
		{TRANS_BLOCK, &IfCaseStmt{}},
		{TRANS_BLOCK, &RangeStmt{}},
		{TRANS_BLOCK, &FuncLitExpr{}},
		{TRANS_BLOCK, &SwitchStmt{}},
		{TRANS_BLOCK, &SwitchClauseStmt{}},
		{TRANS_BLOCK, &FuncDecl{}},
		{TRANS_BLOCK, &FileNode{}},
		// TRANS_BLOCK2.
		{TRANS_BLOCK2, &SwitchStmt{}},
		// TRANS_LEAVE — sample of specific cases.
		{TRANS_LEAVE, &NameExpr{}},
		{TRANS_LEAVE, &BasicLitExpr{}},
		{TRANS_LEAVE, &StructTypeExpr{}},
		{TRANS_LEAVE, &AssignStmt{}},
		{TRANS_LEAVE, &FuncDecl{}},
		{TRANS_LEAVE, &FileNode{}},
		{TRANS_LEAVE, &ConstExpr{}},
	}

	for _, s := range samples {
		code := preprocessNodeCode(s.stage, s.node)
		if code == bm.PreprocessOpInvalid {
			t.Errorf("preprocessNodeCode(%v, %T) = PreprocessOpInvalid", s.stage, s.node)
			continue
		}
		if preprocessGasCosts[code] == 0 {
			t.Errorf(
				"preprocessNodeCode(%v, %T) = %s (code 0x%x) but cost is 0; "+
					"calibrate preprocessGasCosts or this visit will fall back "+
					"to preprocessDefaultGasCost in production",
				s.stage, s.node,
				bm.PreprocessCodeString(code), byte(code),
			)
		}
	}
}

// TestConsumePreprocessGasNilMeter asserts the no-op contract for
// non-transaction contexts (REPL, stdlib predefine, tests).
func TestConsumePreprocessGasNilMeter(t *testing.T) {
	t.Parallel()

	// Should not panic, should not allocate, should not deadlock.
	consumePreprocessGas(nil, bm.PreprocessLeaveStructTypeExpr)
}
