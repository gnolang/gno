package gnolang

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
  OpForLoop +block

RangeStmt ->
  OpRangeIterList +block
  OpRangeIterMap +block
  OpRangeIterString +block

IfStmt ->
  OpIfCond -> +block

SwitchStmt -> +block
  OpSwitchClause
    OpSwitchClauseCase
  OpTypeSwitch

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
	s := m.PeekStmt(1) // TODO: PeekStmt1()?
	m.Lastline = s.GetLine()
	if debug {
		debug.Printf("PEEK STMT: %v\n", s)
		debug.Printf("%v\n", m)
	}

	// NOTE this could go in the switch statement, and we could
	// use the EXEC_SWITCH to jump back, rather than putting this
	// in front like so, but loops are so common that this is
	// likely faster overall, as the type switch is slower than
	// this type assertion conditional.
	switch op {
	case OpBody:
		bs := m.LastBlock().GetBodyStmt()
		if bs.NextBodyIndex == -2 { // init
			bs.NumOps = len(m.Ops)
			bs.NumValues = len(m.Values)
			bs.NumExprs = len(m.Exprs)
			bs.NumStmts = len(m.Stmts)
			bs.NextBodyIndex = 0
		}
		if bs.NextBodyIndex < bs.BodyLen {
			next := bs.Body[bs.NextBodyIndex]
			bs.NextBodyIndex++
			// continue onto exec stmt.
			bs.Active = next
			s = next
			goto EXEC_SWITCH
		} else {
			m.ForcePopOp()
			m.ForcePopStmt()
			return
		}
	case OpForLoop:
		bs := m.LastBlock().GetBodyStmt()
		// evaluate .Cond.
		if bs.NextBodyIndex == -2 { // init
			bs.NumOps = len(m.Ops)
			bs.NumValues = len(m.Values)
			bs.NumExprs = len(m.Exprs)
			bs.NumStmts = len(m.Stmts)
			bs.NextBodyIndex = -1
		}
		if bs.NextBodyIndex == -1 {
			if bs.Cond != nil {
				cond := m.PopValue()
				if !cond.GetBool() {
					// done with loop.
					m.PopFrameAndReset()
					return
				}
			}
			bs.NextBodyIndex++
		}
		// execute body statement.
		if bs.NextBodyIndex < bs.BodyLen {
			next := bs.Body[bs.NextBodyIndex]
			bs.NextBodyIndex++
			// continue onto exec stmt.
			bs.Active = next
			s = next
			goto EXEC_SWITCH
		} else if bs.NextBodyIndex == bs.BodyLen {
			// (queue to) go back.
			if bs.Cond != nil {
				m.PushExpr(bs.Cond)
				m.PushOp(OpEval)
			}
			bs.NextBodyIndex = -1
			if next := bs.Post; next == nil {
				bs.Active = nil
				return // go back now.
			} else {
				// continue onto post stmt.
				// XXX this is a kind of exception....
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
	case OpRangeIter, OpRangeIterArrayPtr:
		bs := s.(*bodyStmt)
		xv := m.PeekValue(1)
		// TODO check length.
		switch bs.NextBodyIndex {
		case -2: // init.
			var ll int
			var dv *TypedValue
			if op == OpRangeIterArrayPtr {
				dv = xv.V.(PointerValue).TV
				*xv = *dv
			} else {
				dv = xv
				*xv = xv.Copy(m.Alloc)
			}
			ll = dv.GetLength()
			if ll == 0 { // early termination
				m.PopFrameAndReset()
				return
			}
			bs.ListLen = ll
			bs.NumOps = len(m.Ops)
			bs.NumValues = len(m.Values)
			bs.NumExprs = len(m.Exprs)
			bs.NumStmts = len(m.Stmts)
			bs.NextBodyIndex++
			fallthrough
		case -1: // assign list element.
			if bs.Key != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(int64(bs.ListIndex))
				switch bs.Op {
				case ASSIGN:
					m.PopAsPointer(bs.Key).Assign2(m.Alloc, m.Store, m.Realm, iv, false)
				case DEFINE:
					knx := bs.Key.(*NameExpr)
					ptr := m.LastBlock().GetPointerToMaybeHeapDefine(m.Store, knx)
					ptr.TV.Assign(m.Alloc, iv, false)
				default:
					panic("should not happen")
				}
			}
			if bs.Value != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(int64(bs.ListIndex))
				ev := xv.GetPointerAtIndex(m.Realm, m.Alloc, m.Store, &iv).Deref()
				switch bs.Op {
				case ASSIGN:
					m.PopAsPointer(bs.Value).Assign2(m.Alloc, m.Store, m.Realm, ev, false)
				case DEFINE:
					vnx := bs.Value.(*NameExpr)
					ptr := m.LastBlock().GetPointerToMaybeHeapDefine(m.Store, vnx)
					ptr.TV.Assign(m.Alloc, ev, false)
				default:
					panic("should not happen")
				}
			}
			bs.NextBodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIterMap,
			// but without tracking Next.
			if bs.NextBodyIndex < bs.BodyLen {
				next := bs.Body[bs.NextBodyIndex]
				bs.NextBodyIndex++
				// continue onto exec stmt.
				bs.Active = next
				s = next // switch on bs.Active
				goto EXEC_SWITCH
			} else if bs.NextBodyIndex == bs.BodyLen {
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
					bs.NextBodyIndex = -1
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
		switch bs.NextBodyIndex {
		case -2: // init.
			// We decode utf8 runes in order --
			// we don't yet know the number of runes.
			strLen := xv.GetLength()
			if strLen == 0 { // early termination
				m.PopFrameAndReset()
				return
			}
			bs.StrLen = strLen
			r, size := utf8.DecodeRuneInString(sv)
			bs.NextRune = r
			bs.StrIndex += size
			bs.NumOps = len(m.Ops)
			bs.NumValues = len(m.Values)
			bs.NumExprs = len(m.Exprs)
			bs.NumStmts = len(m.Stmts)
			bs.NextBodyIndex++
			fallthrough
		case -1: // assign list element.
			if bs.Key != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(int64(bs.ListIndex))
				switch bs.Op {
				case ASSIGN:
					m.PopAsPointer(bs.Key).Assign2(m.Alloc, m.Store, m.Realm, iv, false)
				case DEFINE:
					knx := bs.Key.(*NameExpr)
					ptr := m.LastBlock().GetPointerToMaybeHeapDefine(m.Store, knx)
					ptr.TV.Assign(m.Alloc, iv, false)
				default:
					panic("should not happen")
				}
			}
			if bs.Value != nil {
				ev := typedRune(bs.NextRune)
				switch bs.Op {
				case ASSIGN:
					m.PopAsPointer(bs.Value).Assign2(m.Alloc, m.Store, m.Realm, ev, false)
				case DEFINE:
					vnx := bs.Value.(*NameExpr)
					ptr := m.LastBlock().GetPointerToMaybeHeapDefine(m.Store, vnx)
					ptr.TV.Assign(m.Alloc, ev, false)
				default:
					panic("should not happen")
				}
			}
			bs.NextBodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIterMap,
			// but without tracking Next.
			if bs.NextBodyIndex < bs.BodyLen {
				next := bs.Body[bs.NextBodyIndex]
				bs.NextBodyIndex++
				// continue onto exec stmt.
				bs.Active = next
				s = next // switch on bs.Active
				goto EXEC_SWITCH
			} else if bs.NextBodyIndex == bs.BodyLen {
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
					bs.ListIndex = bs.StrIndex // trails StrIndex.
					bs.StrIndex += size
					bs.NextBodyIndex = -1
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
		var mv *MapValue
		if xv.V != nil {
			mv = xv.V.(*MapValue)
		}
		switch bs.NextBodyIndex {
		case -2: // init.
			if mv == nil || mv.GetLength() == 0 { // early termination
				m.PopFrameAndReset()
				return
			}
			// initialize bs.
			bs.NextItem = mv.List.Head
			bs.NumOps = len(m.Ops)
			bs.NumValues = len(m.Values)
			bs.NumExprs = len(m.Exprs)
			bs.NumStmts = len(m.Stmts)
			bs.NextBodyIndex++
			fallthrough
		case -1: // assign list element.
			next := bs.NextItem
			if bs.Key != nil {
				kv := *fillValueTV(m.Store, &next.Key)
				switch bs.Op {
				case ASSIGN:
					m.PopAsPointer(bs.Key).Assign2(m.Alloc, m.Store, m.Realm, kv, false)
				case DEFINE:
					knx := bs.Key.(*NameExpr)
					ptr := m.LastBlock().GetPointerToMaybeHeapDefine(m.Store, knx)
					ptr.TV.Assign(m.Alloc, kv, false)
				default:
					panic("should not happen")
				}
			}
			if bs.Value != nil {
				vv := *fillValueTV(m.Store, &next.Value)
				switch bs.Op {
				case ASSIGN:
					m.PopAsPointer(bs.Value).Assign2(m.Alloc, m.Store, m.Realm, vv, false)
				case DEFINE:
					vnx := bs.Value.(*NameExpr)
					ptr := m.LastBlock().GetPointerToMaybeHeapDefine(m.Store, vnx)
					ptr.TV.Assign(m.Alloc, vv, false)
				default:
					panic("should not happen")
				}
			}
			bs.NextBodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIter,
			// with slight modification to track Next.
			if bs.NextBodyIndex < bs.BodyLen {
				next := bs.Body[bs.NextBodyIndex]
				bs.NextBodyIndex++
				// continue onto exec stmt.
				bs.Active = next
				s = next // switch on bs.Active
				goto EXEC_SWITCH
			} else if bs.NextBodyIndex == bs.BodyLen {
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
					bs.NextBodyIndex = -1
					bs.Active = nil
					return // redo doOpExec:*bodyStmt
				}
			} else {
				panic("should not happen")
			}
		}
	}

EXEC_SWITCH:
	if debug {
		debug.Printf("EXEC: %v\n", s)
	}
	switch cs := s.(type) {
	case *AssignStmt:
		switch cs.Op {
		case ASSIGN:
			m.PushOp(OpAssign)
		case ADD_ASSIGN:
			m.PushOp(OpAddAssign)
		case SUB_ASSIGN:
			m.PushOp(OpSubAssign)
		case MUL_ASSIGN:
			m.PushOp(OpMulAssign)
		case QUO_ASSIGN:
			m.PushOp(OpQuoAssign)
		case REM_ASSIGN:
			m.PushOp(OpRemAssign)
		case BAND_ASSIGN:
			m.PushOp(OpBandAssign)
		case BOR_ASSIGN:
			m.PushOp(OpBorAssign)
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
			panic(fmt.Sprintf(
				"unexpected assign type %s",
				cs.Op,
			))
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
		b := m.Alloc.NewBlock(cs, m.LastBlock())
		b.bodyStmt = bodyStmt{
			Body:          cs.Body,
			BodyLen:       len(cs.Body),
			NextBodyIndex: -2,
			Cond:          cs.Cond,
			Post:          cs.Post,
		}
		m.PushBlock(b)
		m.PushOp(OpForLoop)
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
		b := m.Alloc.NewBlock(cs, m.LastBlock())
		m.PushBlock(b)
		m.PushOp(OpPopBlock)
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
			m.PushOp(OpInc)
		case DEC:
			m.PushOp(OpDec)
		default:
			panic("unexpected inc/dec operation")
		}
		// Push eval operations if needed.
		m.PushForPointer(cs.X)
	case *ReturnStmt:
		m.PopStmt()
		fr := m.MustPeekCallFrame(1)
		ft := fr.Func.GetType(m.Store)
		hasDefers := 0 < len(fr.Defers)
		hasResults := 0 < len(ft.Results)
		// If has defers, return from the block stack.
		if hasDefers {
			// NOTE: unnamed results are given hidden names
			// ".res%d" from the preprocessor, so they are
			// present in the func block.
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
			} else if cs.CopyResults {
				m.PushOp(OpReturnAfterCopy)
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
		b := m.Alloc.NewBlock(cs, m.LastBlock())
		b.bodyStmt = bodyStmt{
			Body:          cs.Body,
			BodyLen:       len(cs.Body),
			NextBodyIndex: -2,
			Key:           cs.Key,
			Value:         cs.Value,
			Op:            cs.Op,
		}
		m.PushBlock(b)
		// TODO: replace with "cs.Op".
		if cs.IsMap {
			m.PushOp(OpRangeIterMap)
		} else if cs.IsString {
			m.PushOp(OpRangeIterString)
		} else if cs.IsArrayPtr {
			m.PushOp(OpRangeIterArrayPtr)
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
				case *ForStmt, *RangeStmt, *SwitchStmt:
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
			m.GotoJump(int(cs.FrameDepth), int(cs.BlockDepth))
			last := m.LastBlock()
			bs := last.GetBodyStmt()
			m.Ops = m.Ops[:bs.NumOps]
			m.Values = m.Values[:bs.NumValues]
			m.Exprs = m.Exprs[:bs.NumExprs]
			m.Stmts = m.Stmts[:bs.NumStmts]
			bs.NextBodyIndex = cs.BodyIndex
			bs.Active = bs.Body[cs.BodyIndex] // prefill
		case FALLTHROUGH:
			ss, ok := m.LastFrame().Source.(*SwitchStmt)
			// this is handled in the preprocessor
			// should never happen
			if !ok {
				panic("fallthrough statement out of place")
			}

			b := m.LastBlock()
			// compute next switch clause from BodyIndex (assigned in preprocess)
			nextClause := cs.BodyIndex + 1
			// expand block size
			cl := &ss.Clauses[nextClause]
			b.ExpandWith(m.Alloc, cl)
			// exec clause body
			b.bodyStmt = bodyStmt{
				Body:          cl.Body,
				BodyLen:       len(cl.Body),
				NextBodyIndex: -2,
			}
			m.PushOp(OpBody)
			m.PushStmt(b.GetBodyStmt())
		default:
			panic("unknown branch op")
		}
	case *DeclStmt:
		m.PopStmt()
		for i := len(cs.Body) - 1; 0 <= i; i-- {
			m.PushStmt(cs.Body[i])
			m.PushOp(OpExec)
		}
	case *ValueDecl: // SimpleDeclStmt
		m.PushOp(OpValueDecl)
		if cs.Type != nil {
			m.PushExpr(cs.Type)
			m.PushOp(OpEval)
		}
		for i := len(cs.Values) - 1; 0 <= i; i-- {
			m.PushExpr(cs.Values[i])
			m.PushOp(OpEval)
		}
	case *TypeDecl: // SimpleDeclStmt
		m.PushOp(OpTypeDecl)
		m.PushExpr(cs.Type)
		m.PushOp(OpEval)
	case *DeferStmt:
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
	case *SwitchStmt:
		m.PushFrameBasic(cs)
		m.PushOp(OpPopFrameAndReset)
		b := m.Alloc.NewBlock(cs, m.LastBlock())
		m.PushBlock(b)
		m.PushOp(OpPopBlock)
		if cs.IsTypeSwitch {
			m.PushOp(OpTypeSwitch)
			// evaluate x
			m.PushExpr(cs.X)
			m.PushOp(OpEval)
		} else {
			m.PushOp(OpSwitchClause)
			// push clause index 0
			m.PushValue(typedInt(0))
			// push clause case index 0
			m.PushValue(typedInt(0))
			// evaluate x
			m.PushExpr(cs.X)
			m.PushOp(OpEval)
		}
		// exec init
		if cs.Init != nil {
			m.PushOp(OpExec)
			m.PushStmt(cs.Init)
		}
	case *BlockStmt:
		b := m.Alloc.NewBlock(cs, m.LastBlock())
		m.PushBlock(b)
		m.PushOp(OpPopBlock)
		b.bodyStmt = bodyStmt{
			Body:          cs.Body,
			BodyLen:       len(cs.Body),
			NextBodyIndex: -2,
		}
		m.PushOp(OpBody)
		m.PushStmt(b.GetBodyStmt())
	case *EmptyStmt:
	default:
		panic(fmt.Sprintf("unexpected statement %#v", s))
	}
}

func (m *Machine) doOpIfCond() {
	is := m.PopStmt().(*IfStmt)
	b := m.LastBlock()
	// Test cond and run Body or Else.
	cond := m.PopValue()
	if cond.GetBool() {
		if len(is.Then.Body) != 0 {
			// expand block size
			b.ExpandWith(m.Alloc, &is.Then)
			// exec then body
			b.bodyStmt = bodyStmt{
				Body:          is.Then.Body,
				BodyLen:       len(is.Then.Body),
				NextBodyIndex: -2,
			}
			m.PushOp(OpBody)
			m.PushStmt(b.GetBodyStmt())
		}
	} else {
		if len(is.Else.Body) != 0 {
			// expand block size
			b.ExpandWith(m.Alloc, &is.Else)
			// exec then body
			b.bodyStmt = bodyStmt{
				Body:          is.Else.Body,
				BodyLen:       len(is.Else.Body),
				NextBodyIndex: -2,
			}
			m.PushOp(OpBody)
			m.PushStmt(b.GetBodyStmt())
		}
	}
}

func (m *Machine) doOpTypeSwitch() {
	ss := m.PopStmt().(*SwitchStmt)
	xv := m.PopValue()
	xtid := TypeID("")
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
				if ct == nil {
					if xv.IsUndefined() {
						// match nil type with undefined
						match = true
					}
				} else if ct.Kind() == InterfaceKind {
					gnot := ct
					if baseOf(gnot).(*InterfaceType).IsImplementedBy(xv.T) {
						// match
						match = true
					}
				} else {
					ctid := TypeID("")
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
			if len(cs.Body) != 0 {
				b := m.LastBlock()
				// remember size (from init)
				size := len(b.Values)
				// expand block size
				b.ExpandWith(m.Alloc, cs)
				// define if varname
				if ss.VarName != "" {
					// NOTE: assumes the var is first after size.
					vp := NewValuePath(
						VPBlock, 1, uint16(size), ss.VarName)
					// NOTE: GetPointerToMaybeHeapDefine not needed,
					// because this type is in new type switch clause block.
					ptr := b.GetPointerTo(m.Store, vp)
					ptr.TV.Assign(m.Alloc, *xv, false)
				}
				// exec clause body
				b.bodyStmt = bodyStmt{
					Body:          cs.Body,
					BodyLen:       len(cs.Body),
					NextBodyIndex: -2,
				}
				m.PushOp(OpBody)
				m.PushStmt(b.GetBodyStmt())
			}
			return // done!
		}
	}
}

func (m *Machine) doOpSwitchClause() {
	ss := m.PeekStmt1().(*SwitchStmt)
	// tv := m.PeekValue(1) // switch tag value
	// caiv := m.PeekValue(2) // switch clause case index (reuse)
	cliv := m.PeekValue(3) // switch clause index (reuse)
	idx := cliv.GetInt()
	if int(idx) >= len(ss.Clauses) {
		// no clauses matched: do nothing.
		m.PopStmt()  // pop switch stmt
		m.PopValue() // pop switch tag value
		m.PopValue() // pop clause case index
		m.PopValue() // pop clause index
		// done!
	} else {
		cl := &ss.Clauses[idx]
		if len(cl.Cases) == 0 {
			// default clause
			m.PopStmt()  // pop switch stmt
			m.PopValue() // pop switch tag value
			m.PopValue() // clause case index no longer needed
			m.PopValue() // clause index no longer needed
			// expand block size
			b := m.LastBlock()
			b.ExpandWith(m.Alloc, cl)
			// exec clause body
			b.bodyStmt = bodyStmt{
				Body:          cl.Body,
				BodyLen:       len(cl.Body),
				NextBodyIndex: -2,
			}
			m.PushOp(OpBody)
			m.PushStmt(b.GetBodyStmt())
		} else {
			// try to match switch clause case(s).
			m.PushOp(OpSwitchClauseCase)
			// push first case expr
			m.PushOp(OpEval)
			m.PushExpr(cl.Cases[0])
		}
	}
}

func (m *Machine) doOpSwitchClauseCase() {
	cv := m.PopValue()     // switch case value
	tv := m.PeekValue(1)   // switch tag value
	caiv := m.PeekValue(2) // clause case index (reuse)
	cliv := m.PeekValue(3) // clause index (reuse)

	// eval whether cv == tv.
	if debug {
		debugAssertEqualityTypes(cv.T, tv.T)
	}
	match := isEql(m.Store, cv, tv)
	if match {
		// matched clause
		ss := m.PopStmt().(*SwitchStmt) // pop switch stmt
		m.PopValue()                    // pop switch tag value
		m.PopValue()                    // pop clause case index
		m.PopValue()                    // pop clause index
		// expand block size
		clidx := cliv.GetInt()
		cl := &ss.Clauses[clidx]
		b := m.LastBlock()
		b.ExpandWith(m.Alloc, cl)
		// exec clause body
		b.bodyStmt = bodyStmt{
			Body:          cl.Body,
			BodyLen:       len(cl.Body),
			NextBodyIndex: -2,
		}
		m.PushOp(OpBody)
		m.PushStmt(b.GetBodyStmt())
	} else {
		// try next case or clause.
		ss := m.PeekStmt1().(*SwitchStmt) // peek switch stmt
		clidx := cliv.GetInt()
		cl := ss.Clauses[clidx]
		caidx := caiv.GetInt()
		if int(caidx+1) < len(cl.Cases) {
			// try next clause case.
			m.PushOp(OpSwitchClauseCase) // TODO consider sticky
			caiv.SetInt(caidx + 1)
			m.PushOp(OpEval)
			m.PushExpr(cl.Cases[caidx+1])
		} else {
			// no more cases: next clause.
			m.PushOp(OpSwitchClause) // TODO make sticky
			cliv.SetInt(clidx + 1)
			caiv.SetInt(0)
		}
	}
}
