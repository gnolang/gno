package gnolang

import (
	"fmt"
	"strings"
)

type Execution struct {
	Stmt  Stmt
	Frame *Frame
}
type Stacktrace struct {
	Executions []Execution
}

func (s Stacktrace) String() string {
	var builder strings.Builder
	for _, e := range s.Executions {
		if e.Stmt == nil {
			builder.WriteString(fmt.Sprintf("%s()\n", e.Frame.Func))
			builder.WriteString(fmt.Sprintf("    %s/%s:%d\n", e.Frame.Func.PkgPath, e.Frame.Func.FileName, e.Frame.Func.Source.GetLine()))
		} else {
			builder.WriteString(fmt.Sprintf("%s\n", e.Stmt.String()))
			builder.WriteString(fmt.Sprintf("    %s\n", frameFuncLocation(e.Frame, e.Stmt)))
		}
	}
	return builder.String()
}

func (m *Machine) Stacktrace() Stacktrace {
	stacktrace := Stacktrace{}

	index := len(m.Stmts)
	for i := len(m.Frames) - 1; i >= 0; i-- {
		if m.Frames[i].Func != nil {
			stm := m.Stmts[index-1]
			bs := stm.(*bodyStmt)
			stm = bs.Body[bs.NextBodyIndex-1]
			stacktrace.Executions = append(stacktrace.Executions, Execution{
				Stmt:  stm,
				Frame: m.Frames[i],
			})
		}
		index = m.Frames[i].NumStmts
	}

	if len(m.Frames) > 0 {
		stacktrace.Executions = append(stacktrace.Executions, Execution{
			Frame: m.Frames[0],
		})
	}

	return stacktrace
}

func frameFuncLocation(f *Frame, stmt Stmt) string {
	return fmt.Sprintf("%s/%s:%d", f.Func.PkgPath, f.Func.FileName, stmt.GetLine())
}
