package gno

// DEPRECATED
// This is a simple/naive way to implement a for-loop.
// Instead, we use Block.bodyStmts{}.
func (m *Machine) doOpForLoop1() {
	fs := m.PeekStmt1().(*ForStmt)
	cond := m.PopValue()
	if cond.GetBool() {
		// Run loop instance.
		// continuation .
		m.PushOp(OpForLoop1)
		// Evaluate condition for next loop.
		if fs.Cond != nil {
			m.PushExpr(fs.Cond)
			m.PushOp(OpEval)
		}
		// Exec post statement for next loop.
		if fs.Post != nil {
			m.PushStmt(fs.Post)
			m.PushOp(OpExec)
		}
		// Run the body.
		for i := len(fs.Body) - 1; 0 <= i; i-- {
			s := fs.Body[i]
			m.PushStmt(s)
			m.PushOp(OpExec)
		}
	} else {
		// Terminate loop.
		m.PopStmt()
		m.PopBlock()
	}
}
