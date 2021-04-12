package gno

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

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
  OpSwitchClause
  OpTypeSwitchClause

SelectStmt ->
  OpSelectCase +block

*/

//----------------------------------------
// doOpExec
//
// NOTE: Push operations appear in opposite order (filo).  The end result of
// running this operation and any queued in the op stack following this
// operation is that the value of the expression is pushed onto the stack.

func (m *Machine) doOpExec(op Op) {
	s := m.PeekStmt(1)
	if debug {
		debug.Printf("EXEC: %v\n", s)
	}

	// NOTE this could go in the switch statement, and we could
	// use the EXEC_SWITCH to jump back, rather than putting this
	// in front like so, but loops are so common that this is
	// likely faster overall, as the type switch is slower than
	// this type assertion conditional.
	switch op {
	case OpBody:
		bs := m.LastBlock().GetBodyStmt()
		if bs.BodyIndex == -2 { // init
			bs.NumOps = m.NumOps
			bs.NumStmts = len(m.Stmts)
			bs.BodyIndex = 0
		}
		if bs.BodyIndex < bs.BodyLen {
			next := bs.Body[bs.BodyIndex]
			bs.BodyIndex++
			// continue onto exec stmt.
			bs.Active = next
			s = next
			goto EXEC_SWITCH
		} else {
			m.ForcePopOp()
			m.ForcePopStmt()
			return
		}
	case OpForLoop2:
		bs := m.LastBlock().GetBodyStmt()
		// evaluate .Cond.
		if bs.BodyIndex == -2 { // init
			bs.NumOps = m.NumOps
			bs.NumStmts = len(m.Stmts)
			bs.BodyIndex = -1
		}
		if bs.BodyIndex == -1 {
			if bs.Cond != nil {
				cond := m.PopValue()
				if !cond.GetBool() {
					// done with loop.
					m.PopFrameAndReset()
					return
				}
			}
			bs.BodyIndex++ // TODO remove
		}
		// execute body statement.
		if bs.BodyIndex < bs.BodyLen {
			next := bs.Body[bs.BodyIndex]
			bs.BodyIndex++
			// continue onto exec stmt.
			bs.Active = next
			s = next
			goto EXEC_SWITCH
		} else if bs.BodyIndex == bs.BodyLen {
			// (queue to) go back.
			if bs.Cond != nil {
				m.PushExpr(bs.Cond)
				m.PushOp(OpEval)
			}
			bs.BodyIndex = -1
			if next := bs.Post; next == nil {
				bs.Active = nil
				return // go back now.
			} else {
				// continue onto post stmt.
				// XXX this is a kind of excewption....
				// that is, this needs to run after
				// the bodyStmt is force popped?
				// or uh...
				bs.Active = next
				s = next
				goto EXEC_SWITCH
			}
		} else {
			panic("should not happen")
		}
	case OpRangeIter:
		bs := s.(*bodyStmt)
		xv := m.PeekValue(1)
		// TODO check length.
		switch bs.BodyIndex {
		case -2: // init.
			bs.ListLen = xv.GetLength()
			bs.NumOps = m.NumOps
			bs.NumStmts = len(m.Stmts)
			bs.BodyIndex++
			fallthrough
		case -1: // assign list element.
			if bs.Key != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(bs.ListIndex)
				if bs.ListIndex == 0 {
					switch bs.Op {
					case ASSIGN:
						m.PopAsPointer(bs.Key).Assign(iv, false)
					case DEFINE:
						knxp := bs.Key.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(knxp)
						ptr.Assign(iv, false)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(bs.Key).Assign(iv, false)
				}
			}
			if bs.Value != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(bs.ListIndex)
				ev := xv.GetPointerAtIndex(&iv).Deref()
				if bs.ListIndex == 0 {
					switch bs.Op {
					case ASSIGN:
						m.PopAsPointer(bs.Value).Assign(ev, false)
					case DEFINE:
						vnxp := bs.Value.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(vnxp)
						ptr.Assign(ev, false)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(bs.Value).Assign(ev, false)
				}
			}
			bs.BodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIterMap,
			// but without tracking Next.
			if bs.BodyIndex < bs.BodyLen {
				next := bs.Body[bs.BodyIndex]
				bs.BodyIndex++
				// continue onto exec stmt.
				bs.Active = next
				s = next // switch on bs.Active
				goto EXEC_SWITCH
			} else if bs.BodyIndex == bs.BodyLen {
				if bs.ListIndex < bs.ListLen-1 {
					// set up next assign if needed.
					switch bs.Op {
					case ASSIGN:
						if bs.Key != nil {
							m.PushForPointer(bs.Key)
						}
						if bs.Value != nil {
							m.PushForPointer(bs.Value)
						}
					case DEFINE:
						// do nothing
					case ILLEGAL:
						// do nothing, no assignment
					default:
						panic("should not happen")
					}
					bs.ListIndex++
					bs.BodyIndex = -1
					bs.Active = nil
					return // redo doOpExec:*bodyStmt
				} else {
					// done with range.
					m.PopFrameAndReset()
					return
				}
			} else {
				panic("should not happen")
			}
		}
	case OpRangeIterString:
		bs := s.(*bodyStmt)
		xv := m.PeekValue(1)
		sv := xv.GetString()
		switch bs.BodyIndex {
		case -2: // init.
			// We decode utf8 runes in order --
			// we don't yet know the number of runes.
			bs.StrLen = xv.GetLength()
			r, size := utf8.DecodeRuneInString(sv)
			bs.NextRune = r
			bs.StrIndex += size
			bs.NumOps = m.NumOps
			bs.NumStmts = len(m.Stmts)
			bs.BodyIndex++
			fallthrough
		case -1: // assign list element.
			if bs.Key != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(bs.ListIndex)
				if bs.ListIndex == 0 {
					switch bs.Op {
					case ASSIGN:
						m.PopAsPointer(bs.Key).Assign(iv, false)
					case DEFINE:
						knxp := bs.Key.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(knxp)
						ptr.Assign(iv, false)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(bs.Key).Assign(iv, false)
				}
			}
			if bs.Value != nil {
				ev := typedRune(bs.NextRune)
				if bs.ListIndex == 0 {
					switch bs.Op {
					case ASSIGN:
						m.PopAsPointer(bs.Value).Assign(ev, false)
					case DEFINE:
						vnxp := bs.Value.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(vnxp)
						ptr.Assign(ev, false)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(bs.Value).Assign(ev, false)
				}
			}
			bs.BodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIterMap,
			// but without tracking Next.
			if bs.BodyIndex < bs.BodyLen {
				next := bs.Body[bs.BodyIndex]
				bs.BodyIndex++
				// continue onto exec stmt.
				bs.Active = next
				s = next // switch on bs.Active
				goto EXEC_SWITCH
			} else if bs.BodyIndex == bs.BodyLen {
				if bs.StrIndex < bs.StrLen {
					// set up next assign if needed.
					switch bs.Op {
					case ASSIGN:
						if bs.Key != nil {
							m.PushForPointer(bs.Key)
						}
						if bs.Value != nil {
							m.PushForPointer(bs.Value)
						}
					case DEFINE:
						// do nothing
					case ILLEGAL:
						// do nothing, no assignment
					default:
						panic("should not happen")
					}
					rsv := sv[bs.StrIndex:]
					r, size := utf8.DecodeRuneInString(rsv)
					bs.NextRune = r
					bs.StrIndex += size
					bs.ListIndex++
					bs.BodyIndex = -1
					bs.Active = nil
					return // redo doOpExec:*bodyStmt
				} else {
					// done with range.
					m.PopFrameAndReset()
					return
				}
			} else {
				panic("should not happen")
			}
		}
	case OpRangeIterMap:
		bs := s.(*bodyStmt)
		xv := m.PeekValue(1)
		mv := xv.V.(*MapValue)
		switch bs.BodyIndex {
		case -2: // init.
			// bs.ListLen = xv.GetLength()
			bs.NextItem = mv.List.Head
			bs.NumOps = m.NumOps
			bs.NumStmts = len(m.Stmts)
			bs.BodyIndex++
			fallthrough
		case -1: // assign list element.
			next := bs.NextItem
			if bs.Key != nil {
				kv := next.Key
				if bs.ListIndex == 0 {
					switch bs.Op {
					case ASSIGN:
						m.PopAsPointer(bs.Key).Assign(kv, false)
					case DEFINE:
						knxp := bs.Key.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(knxp)
						ptr.Assign(kv, false)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(bs.Key).Assign(kv, false)
				}
			}
			if bs.Value != nil {
				vv := next.Value
				if bs.ListIndex == 0 {
					switch bs.Op {
					case ASSIGN:
						m.PopAsPointer(bs.Value).Assign(vv, false)
					case DEFINE:
						vnxp := bs.Value.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(vnxp)
						ptr.Assign(vv, false)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(bs.Value).Assign(vv, false)
				}
			}
			bs.BodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIter,
			// with slight modification to track Next.
			if bs.BodyIndex < bs.BodyLen {
				next := bs.Body[bs.BodyIndex]
				bs.BodyIndex++
				// continue onto exec stmt.
				bs.Active = next
				s = next // switch on bs.Active
				goto EXEC_SWITCH
			} else if bs.BodyIndex == bs.BodyLen {
				nnext := bs.NextItem.Next
				if nnext == nil {
					// done with range.
					m.PopFrameAndReset()
					return
				} else {
					// set up next assign if needed.
					switch bs.Op {
					case ASSIGN:
						if bs.Key != nil {
							m.PushForPointer(bs.Key)
						}
						if bs.Value != nil {
							m.PushForPointer(bs.Value)
						}
					case DEFINE:
						// do nothing
					case ILLEGAL:
						// do nothing, no assignment
					default:
						panic("should not happen")
					}
					bs.NextItem = nnext
					bs.ListIndex++
					bs.BodyIndex = -1
					bs.Active = nil
					return // redo doOpExec:*bodyStmt
				}
			} else {
				panic("should not happen")
			}
		}
	}

EXEC_SWITCH:
	switch cs := s.(type) {
	case *AssignStmt:
		// continuation
		switch cs.Op {
		case ASSIGN:
			m.PushOp(OpAssign)
		case ADD_ASSIGN:
			m.PushOp(OpAddAssign)
		case SUB_ASSIGN:
			m.PushOp(OpSubAssign)
		case XOR_ASSIGN:
			m.PushOp(OpXorAssign)
		case SHL_ASSIGN:
			m.PushOp(OpShlAssign)
		case SHR_ASSIGN:
			m.PushOp(OpShrAssign)
		case BAND_NOT_ASSIGN:
			m.PushOp(OpBandnAssign)
		case DEFINE:
			m.PushOp(OpDefine)
		default:
			panic("unexpected assign type")
		}
		// For each Rhs, push eval operation.
		for i := len(cs.Rhs) - 1; 0 <= i; i-- {
			rx := cs.Rhs[i]
			// evaluate Rhs
			m.PushExpr(rx)
			m.PushOp(OpEval)
		}
		if cs.Op != DEFINE {
			// For each Lhs, push eval operation if needed.
			for i := len(cs.Lhs) - 1; 0 <= i; i-- {
				lx := cs.Lhs[i]
				m.PushForPointer(lx)
			}
		}
	case *ExprStmt:
		m.PopStmt()
		// All expressions push 1 value except calls,
		// which push as many as there are results.
		if _, ok := cs.X.(*CallExpr); ok {
			m.PushOp(OpPopResults)
		} else {
			m.PushOp(OpPopValue)
		}
		// eval X
		m.PushExpr(cs.X)
		m.PushOp(OpEval)
	case *ForStmt:
		m.PushFrameBasic(cs)
		b := NewBlock(cs, m.LastBlock())
		b.bodyStmt = bodyStmt{
			Body:      cs.Body,
			BodyLen:   len(cs.Body),
			BodyIndex: -1,
			Cond:      cs.Cond,
			Post:      cs.Post,
		}
		m.PushBlock(b)
		// continuation (persistent)
		m.PushOp(OpForLoop2)
		m.PushStmt(b.GetBodyStmt())
		// evaluate condition
		if cs.Cond != nil {
			m.PushExpr(cs.Cond)
			m.PushOp(OpEval)
		}
		// exec init statement
		if cs.Init != nil {
			m.PushStmt(cs.Init)
			m.PushOp(OpExec)
		}
	case *IfStmt:
		b := NewBlock(cs, m.LastBlock())
		m.PushBlock(b)
		// continuation
		m.PushOp(OpIfCond)
		// evaluate condition
		m.PushExpr(cs.Cond)
		m.PushOp(OpEval)
		// initializer
		if cs.Init != nil {
			m.PushStmt(cs.Init)
			m.PushOp(OpExec)
		}
	case *IncDecStmt:
		switch cs.Op {
		case INC:
			// continuation
			m.PushOp(OpInc)
		case DEC:
			// continuation
			m.PushOp(OpDec)
		default:
			panic("unexpected inc/dec operation")
		}
		// Push eval operations if needed.
		m.PushForPointer(cs.X)
	case *ReturnStmt:
		m.PopStmt()
		fr := m.LastFrame()
		hasDefers := 0 < len(fr.Defers)
		hasResults := 0 < len(fr.Func.Type.Results)
		// If has defers, return from the block stack.
		if hasDefers {
			// NOTE: unnamed results are given hidden names
			// ".res%d" from the preprocessor, so they are
			// present in the func block.
			m.PushOp(OpReturnFromBlock)
			m.PushOp(OpReturnCallDefers) // sticky
			if cs.Results == nil {
				// results already in block, if any.
			} else if hasResults {
				// copy return results to block.
				m.PushOp(OpReturnToBlock)
			}
		} else {
			if cs.Results == nil {
				m.PushOp(OpReturnFromBlock)
			} else {
				m.PushOp(OpReturn)
			}
		}
		// Evaluate results in order, if any.
		for i := len(cs.Results) - 1; 0 <= i; i-- {
			res := cs.Results[i]
			m.PushExpr(res)
			m.PushOp(OpEval)
		}
	case *RangeStmt:
		m.PushFrameBasic(cs)
		b := NewBlock(cs, m.LastBlock())
		b.bodyStmt = bodyStmt{
			Body:      cs.Body,
			BodyLen:   len(cs.Body),
			BodyIndex: -2,
			Key:       cs.Key,
			Value:     cs.Value,
			Op:        cs.Op,
		}
		m.PushBlock(b)
		// continuation (persistent)
		if cs.IsMap {
			m.PushOp(OpRangeIterMap)
		} else if cs.IsString {
			m.PushOp(OpRangeIterString)
		} else {
			m.PushOp(OpRangeIter)
		}
		m.PushStmt(b.GetBodyStmt())
		// evaluate eval for assign if needed.
		switch cs.Op {
		case ASSIGN:
			if cs.Key != nil {
				m.PushForPointer(cs.Key)
			}
			if cs.Value != nil {
				m.PushForPointer(cs.Value)
			}
		case DEFINE:
			// do nothing
		case ILLEGAL:
			// do nothing, no assignment
		default:
			panic("should not happen")
		}
		// evaluate X
		m.PushExpr(cs.X)
		m.PushOp(OpEval)
	case *BranchStmt:
		switch cs.Op {
		case BREAK:
			// Pop frames until for/range
			// statement (which matches
			// label, if labeled), and reset.
			for {
				fr := m.LastFrame()
				switch fr.Source.(type) {
				case *ForStmt, *RangeStmt:
					if cs.Label != "" && cs.Label != fr.Label {
						m.PopFrame()
					} else {
						m.PopFrameAndReset()
						return
					}
				default:
					m.PopFrame()
				}
			}
		case CONTINUE:
			// TODO document
			for {
				fr := m.LastFrame()
				switch fr.Source.(type) {
				case *ForStmt:
					if cs.Label != "" && cs.Label != fr.Label {
						m.PopFrame()
					} else {
						m.PeekFrameAndContinueFor()
						return
					}
				case *RangeStmt:
					if cs.Label != "" && cs.Label != fr.Label {
						m.PopFrame()
					} else {
						m.PeekFrameAndContinueRange()
						return
					}
				default:
					m.PopFrame()
				}
			}
		case GOTO:
			for i := uint8(0); i < cs.Depth; i++ {
				m.PopBlock()
			}
			last := m.LastBlock()
			bs := last.GetBodyStmt()
			m.NumOps = bs.NumOps
			m.NumValues = 0
			m.Exprs = nil
			m.Stmts = m.Stmts[:bs.NumStmts]
			bs.BodyIndex = cs.BodyIndex
			bs.Active = bs.Body[cs.BodyIndex]
		case FALLTHROUGH:
			panic("not yet implemented")
		default:
			panic("unknown branch op")
		}
	case *DeclStmt:
		m.PopStmt()
		for _, d := range cs.Decls {
			m.runDeclaration(d)
		}
	case *DeferStmt:
		// continuation
		m.PushOp(OpDefer)
		// evaluate args
		args := cs.Call.Args
		for i := len(args) - 1; 0 <= i; i-- {
			m.PushExpr(args[i])
			m.PushOp(OpEval)
		}
		// evaluate func
		m.PushExpr(cs.Call.Func)
		m.PushOp(OpEval)
	case *LabeledStmt:
		s = cs.Stmt
		goto EXEC_SWITCH
	case *SwitchStmt:
		b := NewBlock(cs, m.LastBlock())
		m.PushBlock(b)
		if cs.IsTypeSwitch {
			// continuation
			m.PushOp(OpTypeSwitchClause)
			// evaluate x
			m.PushExpr(cs.X)
			m.PushOp(OpEval)
		} else {
			// continuation
			m.PushOp(OpSwitchClause)
			// evaluate x
			m.PushExpr(cs.X)
			m.PushOp(OpEval)
		}
	default:
		panic(fmt.Sprintf("unexpected statement %#v", s))
	}
}

func (m *Machine) doOpIfCond() {
	is := m.PopStmt().(*IfStmt)
	b := m.LastBlock()
	// final continuation
	m.PushOp(OpPopBlock)
	// Test cond and run Body or Else.
	cond := m.PopValue()
	if cond.GetBool() {
		if len(is.Then.Body) != 0 {
			b.bodyStmt = bodyStmt{
				Body:      is.Then.Body,
				BodyLen:   len(is.Then.Body),
				BodyIndex: -2,
			}
			m.PushOp(OpBody)
			m.PushStmt(b.GetBodyStmt())
		}
	} else {
		if len(is.Else.Body) != 0 {
			b.bodyStmt = bodyStmt{
				Body:      is.Else.Body,
				BodyLen:   len(is.Else.Body),
				BodyIndex: -2,
			}
			m.PushOp(OpBody)
			m.PushStmt(b.GetBodyStmt())
		}
	}
}

func (m *Machine) doOpSwitchClause() {
	panic("not yet implemented")
}

func (m *Machine) doOpTypeSwitchClause() {
	ss := m.PopStmt().(*SwitchStmt)
	xv := m.PopValue()
	xtid := TypeID{}
	if xv.T != nil {
		xtid = xv.T.TypeID()
	}
	// NOTE: all cases should be *constTypeExprs, which
	// lets us optimize the implementation by
	// iterating over all clauses and cases here.
	for i := range ss.Clauses {
		match := false
		cs := &ss.Clauses[i]
		if len(cs.Cases) > 0 {
			// see if any clause cases match.
			for _, cx := range cs.Cases {
				if debug {
					if !isConstType(cx) {
						panic(fmt.Sprintf(
							"should not happen, expected const type expr for case(s) but got %s",
							reflect.TypeOf(cx)))
					}
				}
				ct := cx.(*constTypeExpr).Type
				if ct.Kind() == InterfaceKind {
					if baseOf(ct).(*InterfaceType).IsImplementedBy(xv.T) {
						// match
						match = true
					}
				} else {
					ctid := TypeID{}
					if ct != nil {
						ctid = ct.TypeID()
					}
					if xtid == ctid {
						// match
						match = true
					}
				}
			}
		} else { // default
			match = true
		}
		if match { // did match
			// final continuation
			m.PushOp(OpPopBlock)
			if len(cs.Body) != 0 {
				b := m.LastBlock()
				// define if varname
				if ss.VarName != "" && len(cs.Cases) <= 1 {
					// NOTE: assumes the var is first in block.
					vp := NewValuePath(
						VPTypeDefault, 1, 0, ss.VarName)
					ptr := b.GetPointerTo(vp)
					ptr.Assign(*xv, false)
				}
				// expand block size
				if nn := cs.GetNumNames(); nn > 1 {
					b.ExpandToSize(nn)
				}
				// exec clause body
				b.bodyStmt = bodyStmt{
					Body:      cs.Body,
					BodyLen:   len(cs.Body),
					BodyIndex: -2,
				}
				m.PushOp(OpBody)
				m.PushStmt(b.GetBodyStmt())
			}
			return // done!
		}
	}
}
