package fsm

type State string

type Rule func(*FSM, *Context) bool

type IFSM interface{
	Current() State 
	Previous() State 
	Next() *Transitions
	Transition(State) bool
	AddTransition(State, State, Rule, *Context) bool
	RemoveTransition(State, State) bool
}

type Contextualized interface{
	Context() interface{}
}