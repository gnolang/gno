package fsm

import (
	"gno.land/p/demo/avl"
)

type FSM struct{
	previousState State 
	currentState State 
	transitionsFrom avl.Tree
}

type Transitions struct{
	transitionsTo avl.Tree
}

type Transition struct{
	fromState State
	toState State
	rule Rule
	context *Context
}

type Context struct {
	ctx interface{}
}

func (ts *Transitions) Contains(next_ State) bool { return ts.transitionsTo.Has(string(next_))}

func (ts *Transitions) Add(next_ State, transition_ *Transition) bool { 
	ts.transitionsTo.Set(string(next_),transition_)
	return true
}

func (ts *Transitions) Remove(next_ State) bool { 
	ts.transitionsTo.Remove(string(next_))
	return true
}

func (t *Transition) From() State { return t.fromState }

func (t *Transition) To() State { return t.toState }

func (t *Transition) Rule() Rule { return t.rule }

func (t *Transition) Context() *Context { return t.context }

func (c *Context) Context() interface{} { return c.ctx }

func (f *FSM) Previous() State { return f.previousState }

func (f *FSM) Current() State { return f.currentState }

func (f *FSM) Next() *Transitions { 
	_t, _exists := f.transitionsFrom.Get(string(f.currentState))
	if !_exists {
		panic("No next transition for current state.")
	}
	return _t.(*Transitions)	
}

func (f *FSM) Transition(next_ State) bool { 
	_transitions, _stateExists := f.transitionsFrom.Get(string(f.currentState))
	if !_stateExists {
		panic("Invalid current state. Exiting.")
	}
	_t, _transExists := _transitions.(*Transitions).transitionsTo.Get(string(next_))
	if !_transExists {
		return false
	}
	_transition := _t.(*Transition)
	_rule := _transition.Rule()
	_ctx := _transition.Context()
	if _rule(f, _ctx) {
		f.previousState = f.currentState
		f.currentState = next_
		return true
	}
	return false
}

func (f *FSM) AddTransition(cur_ State, next_ State, rule_ Rule, ctx_ *Context) bool { 
	_curStr := string(cur_)

	if !f.transitionsFrom.Has(_curStr) {
		f.transitionsFrom.Set(_curStr, NewTransitions())
	}
	_ts, _ := f.transitionsFrom.Get(_curStr)
	_transitions := _ts.(*Transitions)
	
	if _transitions.Contains(next_) {
		return false
	}

	_transition := NewTransition(cur_,next_,rule_,ctx_)
	return _transitions.Add(next_,_transition)
}

func (f *FSM) RemoveTransition(cur_ State, next_ State) bool { 
	_curStr := string(cur_)

	if !f.transitionsFrom.Has(_curStr) {
		return false
	}
	_ts, _ := f.transitionsFrom.Get(_curStr)
	_transitions := _ts.(*Transitions)
	
	if !_transitions.Contains(next_) {
		return false
	}

	return _transitions.Remove(next_)
}