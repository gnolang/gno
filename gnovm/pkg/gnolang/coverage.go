package gnolang

import "reflect"

// StatementCoverage tracks covered statements for a package.
// It is intentionally pointer-based and lightweight for runtime checks.
type StatementCoverage struct {
	tracked map[uintptr]struct{}
	hits    map[uintptr]struct{}
}

func NewStatementCoverage() *StatementCoverage {
	return &StatementCoverage{
		tracked: make(map[uintptr]struct{}, 256),
		hits:    make(map[uintptr]struct{}, 256),
	}
}

func statementID(s Stmt) uintptr {
	return reflect.ValueOf(s).Pointer()
}

func shouldTrackStmt(s Stmt) bool {
	switch s.(type) {
	case *BlockStmt, *EmptyStmt, *IfCaseStmt, *SwitchClauseStmt, *bodyStmt:
		return false
	default:
		return true
	}
}

func (c *StatementCoverage) TrackStmt(s Stmt) {
	if c == nil || !shouldTrackStmt(s) {
		return
	}
	c.tracked[statementID(s)] = struct{}{}
}

func (c *StatementCoverage) TrackNode(n Node) {
	if c == nil || n == nil {
		return
	}
	_ = Transcribe(n, func(_ []Node, _ TransField, _ int, cn Node, stage TransStage) (Node, TransCtrl) {
		if stage == TRANS_ENTER {
			if s, ok := cn.(Stmt); ok {
				c.TrackStmt(s)
			}
		}
		return cn, TRANS_CONTINUE
	})
}

func (c *StatementCoverage) MarkExecuted(s Stmt) {
	if c == nil || !shouldTrackStmt(s) {
		return
	}
	id := statementID(s)
	if _, ok := c.tracked[id]; ok {
		c.hits[id] = struct{}{}
	}
}

func (c *StatementCoverage) Percent() float64 {
	if c == nil || len(c.tracked) == 0 {
		return 0
	}
	return float64(len(c.hits)) * 100 / float64(len(c.tracked))
}
