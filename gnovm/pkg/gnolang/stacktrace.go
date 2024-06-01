package gnolang

import (
	"fmt"
	"strings"
)

func (m *Machine) buildStackTrace() string {
	var builder strings.Builder

	index := len(m.Stmts)
	for i := len(m.Frames) - 1; i >= 0; i-- {
		if m.Frames[i].Func != nil {
			stm := m.Stmts[index-1]
			bs := stm.(*bodyStmt)
			stm = bs.Body[bs.NextBodyIndex-1]
			builder.WriteString(fmt.Sprintf("%s\n", stmtString(stm)))
			builder.WriteString(fmt.Sprintf("    %s\n", frameFuncLocation(m.Frames[i], stm)))
		}
		index = m.Frames[i].NumStmts
	}

	if len(m.Frames) > 0 {
		builder.WriteString(fmt.Sprintf("%s()\n", m.Frames[0].Func))
		builder.WriteString(fmt.Sprintf("    %s/%s:%d\n", m.Frames[0].Func.PkgPath, m.Frames[0].Func.FileName, m.Frames[0].Func.Source.GetLine()))
	}

	return builder.String()
}

func (m *Machine) Stacktrace() string {
	return m.buildStackTrace()
}

func frameFuncLocation(f *Frame, stmt Stmt) string {
	return fmt.Sprintf("%s/%s:%d", f.Func.PkgPath, f.Func.FileName, stmt.GetLine())
}

func stmtString(stm Stmt) string {
	switch s := stm.(type) {
	case *bodyStmt:
		return s.String()
	case *ExprStmt:
		return s.String()
	case *IfStmt:
		return s.String()
	case *ForStmt:
		return s.String()
	case *RangeStmt:
		return s.String()
	case *SwitchStmt:
		return s.String()
	case *ReturnStmt:
		return s.String()
	case *DeferStmt:
		return s.String()
	case *GoStmt:
		return s.String()
	case *BlockStmt:
		return s.String()
	case *EmptyStmt:
		return s.String()
	case *DeclStmt:
		return s.String()
	case *IncDecStmt:
		return s.String()
	case *PanicStmt:
		return s.String()
	case *SelectStmt:
		return s.String()
	case *SendStmt:
		return s.String()
	case *BranchStmt:
		return s.String()
	default:
		return "unknown statement"
	}
}
