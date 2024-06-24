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
		switch {
		case e.Stmt == nil:
			builder.WriteString(fmt.Sprintf("%s()\n", e.Frame.Func))
			builder.WriteString(fmt.Sprintf("    %s/%s:%d\n", e.Frame.Func.PkgPath, e.Frame.Func.FileName, e.Frame.Func.Source.GetLine()))
		case e.Frame.Func != nil && e.Frame.Func.IsNative():
			builder.WriteString(fmt.Sprintf("%s()\n", e.Frame.Func))
			builder.WriteString(fmt.Sprintf("    %s.%s\n", e.Frame.Func.NativePkg, e.Frame.Func.NativeName))
		case e.Frame.Func != nil:
			builder.WriteString(fmt.Sprintf("%s\n", e.Stmt.String()))
			builder.WriteString(fmt.Sprintf("    %s/%s:%d\n", e.Frame.Func.PkgPath, e.Frame.Func.FileName, e.Stmt.GetLine()))
		default:
			builder.WriteString(fmt.Sprintf("%s\n", e.Frame.GoFunc.Value.Type()))
			builder.WriteString("    gonative\n")
		}
	}
	return builder.String()
}

func (m *Machine) Stacktrace() Stacktrace {
	stacktrace := Stacktrace{}

	if len(m.Frames) == 0 {
		return stacktrace
	}

	index := len(m.Stmts)
	for i := len(m.Frames) - 1; i >= 0; i-- {
		if m.Frames[i].IsCall() {
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

	stacktrace.Executions = append(stacktrace.Executions, Execution{
		Frame: m.Frames[0],
	})
	return stacktrace
}
