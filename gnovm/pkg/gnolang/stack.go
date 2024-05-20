package gnolang

type execution struct {
	Stmt Stmt
	Fun  *FuncValue
}
type Stack struct {
	Execs []execution
}

func NewStack() *Stack {
	return &Stack{
		Execs: make([]execution, 0),
	}
}

func (s *Stack) onStmtPopped(stmt Stmt) {
	if len(s.Execs) > 0 {
		s.Execs[len(s.Execs)-1].Stmt = stmt
	}
}

func (s *Stack) OnFramePushed(frame *Frame) {
	if frame.Func != nil {
		s.Execs = append(s.Execs, execution{Fun: frame.Func})
	}
}

func (s *Stack) OnFramePopped(frame *Frame) {
	if frame.Func != nil {
		s.Execs = s.Execs[:len(s.Execs)-1]
	}
}

func (s *Stack) Copy() *Stack {
	cpy := NewStack()
	cpy.Execs = make([]execution, len(s.Execs))
	copy(cpy.Execs, s.Execs)
	return cpy
}
