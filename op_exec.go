package gno

import (
	"fmt"
	"unicode/utf8"
)

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

	// NOTE this could go in the switch statement, and we could use the
	// EXEC_SWITCH to jump back, rather than putting this in front like so, but
	// loops are so common that this is likely faster overall, as the type
	// switch is slower than this type assertion conditional.
	switch op {
	case OpForLoop2:
		ls := s.(*loopStmt)
		fs := ls.ForStmt
		// loopStmt is for ForStmt.
		// evaluate .Cond.
		if ls.BodyIndex == -1 {
			if fs.Cond != nil {
				cond := m.PopValue()
				if !cond.GetBool() {
					// done with loop.
					m.PopFrameAndReset()
					return
				}
			}
			ls.BodyIndex++ // TODO remove
		}
		// execute body statement.
		if ls.BodyIndex < ls.BodyLen {
			next := fs.Body[ls.BodyIndex]
			ls.BodyIndex++
			// continue onto exec stmt.
			ls.Active = next
			s = next
			goto EXEC_SWITCH
		} else if ls.BodyIndex == ls.BodyLen {
			// (queue to) go back.
			if fs.Cond != nil {
				m.PushExpr(fs.Cond)
				m.PushOp(OpEval)
			}
			ls.BodyIndex = -1
			if next := fs.Post; next == nil {
				ls.Active = nil
				return // go back now.
			} else {
				// continue onto post stmt.
				ls.Active = next
				s = next
				goto EXEC_SWITCH
			}
		} else {
			panic("should not happen")
		}
	case OpRangeIter:
		ls := s.(*loopStmt)
		rs := ls.RangeStmt
		// loopStmt is for RangeStmt.
		xv := m.PeekValue(1)
		// TODO check length.
		switch ls.BodyIndex {
		case -2: // init.
			ls.ListLen = xv.GetLength()
			b := NewBlock(ls.RangeStmt, m.LastBlock())
			m.PushBlock(b)
			ls.BodyIndex++
			fallthrough
		case -1: // assign list element.
			if rs.Key != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(ls.ListIndex)
				if ls.ListIndex == 0 {
					switch rs.Op {
					case ASSIGN:
						m.PopAsPointer(rs.Key).Assign(iv)
					case DEFINE:
						knxp := rs.Key.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(knxp)
						ptr.Assign(iv)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(rs.Key).Assign(iv)
				}
			}
			if rs.Value != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(ls.ListIndex)
				ev := xv.GetPointerAtIndex(&iv).Deref()
				if ls.ListIndex == 0 {
					switch rs.Op {
					case ASSIGN:
						m.PopAsPointer(rs.Value).Assign(ev)
					case DEFINE:
						vnxp := rs.Value.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(vnxp)
						ptr.Assign(ev)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(rs.Value).Assign(ev)
				}
			}
			ls.BodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIterMap,
			// but without tracking Next.
			if ls.BodyIndex < ls.BodyLen {
				next := rs.Body[ls.BodyIndex]
				ls.BodyIndex++
				// continue onto exec stmt.
				ls.Active = next
				s = next // switch on ls.Active
				goto EXEC_SWITCH
			} else if ls.BodyIndex == ls.BodyLen {
				if ls.ListIndex < ls.ListLen-1 {
					// set up next assign if needed.
					switch rs.Op {
					case ASSIGN:
						if rs.Key != nil {
							m.PushForPointer(rs.Key)
						}
						if rs.Value != nil {
							m.PushForPointer(rs.Value)
						}
					case DEFINE:
						// do nothing
					case ILLEGAL:
						// do nothing, no assignment
					default:
						panic("should not happen")
					}
					ls.ListIndex++
					ls.BodyIndex = -1
					ls.Active = nil
					return // redo doOpExec:*loopStmt
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
		ls := s.(*loopStmt)
		rs := ls.RangeStmt
		// loopStmt is for RangeStmt.
		xv := m.PeekValue(1)
		sv := xv.GetString()
		switch ls.BodyIndex {
		case -2: // init.
			// We decode utf8 runes in order --
			// we don't yet know the number of runes.
			ls.StrLen = xv.GetLength()
			r, size := utf8.DecodeRuneInString(sv)
			ls.NextRune = r
			ls.StrIndex += size
			b := NewBlock(ls.RangeStmt, m.LastBlock())
			m.PushBlock(b)
			ls.BodyIndex++
			fallthrough
		case -1: // assign list element.
			if rs.Key != nil {
				iv := TypedValue{T: IntType}
				iv.SetInt(ls.ListIndex)
				if ls.ListIndex == 0 {
					switch rs.Op {
					case ASSIGN:
						m.PopAsPointer(rs.Key).Assign(iv)
					case DEFINE:
						knxp := rs.Key.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(knxp)
						ptr.Assign(iv)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(rs.Key).Assign(iv)
				}
			}
			if rs.Value != nil {
				ev := typedRune(ls.NextRune)
				if ls.ListIndex == 0 {
					switch rs.Op {
					case ASSIGN:
						m.PopAsPointer(rs.Value).Assign(ev)
					case DEFINE:
						vnxp := rs.Value.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(vnxp)
						ptr.Assign(ev)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(rs.Value).Assign(ev)
				}
			}
			ls.BodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIterMap,
			// but without tracking Next.
			if ls.BodyIndex < ls.BodyLen {
				next := rs.Body[ls.BodyIndex]
				ls.BodyIndex++
				// continue onto exec stmt.
				ls.Active = next
				s = next // switch on ls.Active
				goto EXEC_SWITCH
			} else if ls.BodyIndex == ls.BodyLen {
				if ls.StrIndex < ls.StrLen {
					// set up next assign if needed.
					switch rs.Op {
					case ASSIGN:
						if rs.Key != nil {
							m.PushForPointer(rs.Key)
						}
						if rs.Value != nil {
							m.PushForPointer(rs.Value)
						}
					case DEFINE:
						// do nothing
					case ILLEGAL:
						// do nothing, no assignment
					default:
						panic("should not happen")
					}
					rsv := sv[ls.StrIndex:]
					r, size := utf8.DecodeRuneInString(rsv)
					ls.NextRune = r
					ls.StrIndex += size
					ls.ListIndex++
					ls.BodyIndex = -1
					ls.Active = nil
					return // redo doOpExec:*loopStmt
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
		ls := s.(*loopStmt)
		rs := ls.RangeStmt
		// loopStmt is for RangeStmt.
		xv := m.PeekValue(1)
		mv := xv.V.(*MapValue)
		switch ls.BodyIndex {
		case -2: // init.
			// ls.ListLen = xv.GetLength()
			ls.NextItem = mv.List.Head
			b := NewBlock(ls.RangeStmt, m.LastBlock())
			m.PushBlock(b)
			ls.BodyIndex++
			fallthrough
		case -1: // assign list element.
			next := ls.NextItem
			if rs.Key != nil {
				kv := next.Key
				if ls.ListIndex == 0 {
					switch rs.Op {
					case ASSIGN:
						m.PopAsPointer(rs.Key).Assign(kv)
					case DEFINE:
						knxp := rs.Key.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(knxp)
						ptr.Assign(kv)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(rs.Key).Assign(kv)
				}
			}
			if rs.Value != nil {
				vv := next.Value
				if ls.ListIndex == 0 {
					switch rs.Op {
					case ASSIGN:
						m.PopAsPointer(rs.Value).Assign(vv)
					case DEFINE:
						vnxp := rs.Value.(*NameExpr).Path
						ptr := m.LastBlock().GetPointerTo(vnxp)
						ptr.Assign(vv)
					default:
						panic("should not happen")
					}
				} else {
					// Already defined, use assign.
					m.PopAsPointer(rs.Value).Assign(vv)
				}
			}
			ls.BodyIndex++
			fallthrough
		default:
			// NOTE: duplicated for OpRangeIter,
			// with slight modification to track Next.
			if ls.BodyIndex < ls.BodyLen {
				next := rs.Body[ls.BodyIndex]
				ls.BodyIndex++
				// continue onto exec stmt.
				ls.Active = next
				s = next // switch on ls.Active
				goto EXEC_SWITCH
			} else if ls.BodyIndex == ls.BodyLen {
				nnext := ls.NextItem.Next
				if nnext == nil {
					// done with range.
					m.PopFrameAndReset()
					return
				} else {
					// set up next assign if needed.
					switch rs.Op {
					case ASSIGN:
						if rs.Key != nil {
							m.PushForPointer(rs.Key)
						}
						if rs.Value != nil {
							m.PushForPointer(rs.Value)
						}
					case DEFINE:
						// do nothing
					case ILLEGAL:
						// do nothing, no assignment
					default:
						panic("should not happen")
					}
					ls.NextItem = nnext
					ls.ListIndex++
					ls.BodyIndex = -1
					ls.Active = nil
					return // redo doOpExec:*loopStmt
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
		m.PushBlock(b)
		// continuation (persistent)
		m.PushOp(OpForLoop2)
		m.PushStmt(&loopStmt{
			ForStmt:   cs,
			BodyLen:   len(cs.Body),
			BodyIndex: -1,
		})
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
		// continuation (persistent)
		if cs.IsMap {
			m.PushOp(OpRangeIterMap)
		} else if cs.IsString {
			m.PushOp(OpRangeIterString)
		} else {
			m.PushOp(OpRangeIter)
		}
		m.PushStmt(&loopStmt{
			RangeStmt: cs,
			ListLen:   0, // set later
			ListIndex: 0, // set later
			BodyLen:   len(cs.Body),
			BodyIndex: -2,
		})
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
			panic("not yet implemented")
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
	default:
		panic(fmt.Sprintf("unexpected statement %#v", s))
	}
}
