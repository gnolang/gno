package gno

/*
State transition map.
NOTE: does not show frames. We use frames for each of
these except IfStmt to support break/continue branch
statements. Omitting frames requires more complex logic
during break/continue and results in brittle code, so we
choose to use frames for all but IfStmt block nodes.

CallExpr ->
  OpPrecall->
    OpCall-> +block
	  OpReturn?,OpExec.*
	  OpReturn,OpCallNativeBody
    OpCallGoNative
	OpConvert

ForStmt ->
  OpForLoop2 +block

RangeStmt ->
  OpRangeIterList +block
  OpRangeIterMap +block
  OpRangeIterString +block

IfStmt ->
  OpIfCond -> +block
    OpPopBlock

SwitchStmt -> +block
  OpSwitchCase

SelectStmt ->
  OpSelectCase +block

*/

func (m *Machine) doOpForLoop1() {
	fs := m.PeekStmt(1).(*ForStmt)
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

func (m *Machine) doOpIfCond() {
	is := m.PopStmt().(*IfStmt)
	// final continuation
	m.PushOp(OpPopBlock)
	// Test cond and Run the body or else.
	cond := m.PopValue()
	if cond.GetBool() {
		// Run the body.
		for i := len(is.Body) - 1; 0 <= i; i-- {
			s := is.Body[i]
			m.PushStmt(s)
			m.PushOp(OpExec)
		}
	} else {
		// Run the else body.
		for i := len(is.Else) - 1; 0 <= i; i-- {
			s := is.Else[i]
			m.PushStmt(s)
			m.PushOp(OpExec)
		}
	}
}
