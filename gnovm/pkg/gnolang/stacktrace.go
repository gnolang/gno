package gnolang

import (
	"fmt"
	"strings"
)

const maxStacktraceSize = 128

type Execution struct {
	Stmt  Stmt
	Frame *Frame
}
type Stacktrace struct {
	Executions []Execution
}

func (s Stacktrace) String() string {
	var builder strings.Builder

	excecutions := s.Executions
	numFramesElided := 0
	if len(s.Executions) > maxStacktraceSize {
		excecutions = s.Executions[:maxStacktraceSize/2]
		excecutions = append(excecutions, s.Executions[len(s.Executions)-maxStacktraceSize/2:]...)
		numFramesElided = len(s.Executions) - maxStacktraceSize
	}

	for i, e := range excecutions {
		if numFramesElided > 0 && i == maxStacktraceSize/2 {
			builder.WriteString(fmt.Sprintf("...%d frame(s) elided...\n", numFramesElided))
		}

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

// Stacktrace returns the stack trace of the machine.
// It collects the executions and frames from the machine's frames and statements.
func (m *Machine) Stacktrace() (stacktrace Stacktrace) {
	if len(m.Frames) == 0 {
		return
	}

	nextStmtIndex := len(m.Stmts) - 1
	for i := len(m.Frames) - 1; i >= 0; i-- {
		if m.Frames[i].IsCall() {
			stm := m.Stmts[nextStmtIndex]
			bs := stm.(*bodyStmt)
			stm = bs.Body[bs.NextBodyIndex-1]
			stacktrace.Executions = append(stacktrace.Executions, Execution{
				Stmt:  stm,
				Frame: m.Frames[i],
			})
		}
		// if the frame is a call, the next statement is the last statement of the frame.
		nextStmtIndex = m.Frames[i].NumStmts - 1
	}

	stacktrace.Executions = append(stacktrace.Executions, Execution{
		Frame: m.Frames[0],
	})

	return
}
