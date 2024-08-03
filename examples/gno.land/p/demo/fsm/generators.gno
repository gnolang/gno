package fsm

import (
	"gno.land/p/demo/avl"
)

func NewFSM(initial_ State) *FSM { 
	_t := NewTransitions()
	_from := avl.Tree{}
	_from.Set(string(initial_),_t)
	
	return &FSM{
		previousState: State(""),
		currentState: initial_,
		transitionsFrom: _from,
	}
}

func NewContext(ctx_ interface{}) *Context { 
	return &Context{
		ctx: ctx_,
	}
}

func NewTransitions() *Transitions { 
	return &Transitions{
		transitionsTo: avl.Tree{},
	}
}

func NewTransition(cur_ State, next_ State, rule_ Rule, ctx_ *Context) *Transition { 
	return &Transition{
		fromState: cur_, 
		toState: next_,
		rule: rule_,
		context: ctx_,
	}
}