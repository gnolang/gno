package fsm

import (
	"fmt"
	"testing"
	"gno.land/p/demo/avl"
)

const (
	ON = State("ON")
	OFF = State("OFF")

	ADMIN = "ADMIN"
	USER = "USER"
	OUT = State("UNAUTHED")
	ADMINDASH = State("ADMIN")
	USERDASH = State("USER")
)

	

func NewAdminContext() *Context{
	return NewContext(ADMIN)
}

func NewUserContext() *Context{
	return NewContext(USER)
}

func NewEmptyContext() *Context{
	return NewContext(interface{})
}

func NewLoginManager(ctx_ *Context) *FSM {
	_login := NewFSM(OUT)
	_login.AddTransition(OUT, ADMINDASH, LoginAdmin, ctx_)
	_login.AddTransition(OUT, USERDASH, LoginUser, ctx_)
	_login.AddTransition(ADMINDASH, OUT, LoggedIn, ctx_)
	_login.AddTransition(USERDASH, OUT, LoggedIn, ctx_)
		
	return _login
}

func NewLightswitch() *FSM {
	_ctx := NewEmptyContext()	
	_switch := NewFSM(OFF)
	_switch.AddTransition(OFF, ON, SwitchOn, _ctx)
	_switch.AddTransition(ON, OFF, SwitchOff, _ctx)
		
	return _switch
}

func SwitchOn(fsm_ *FSM, ctx_ *Context) bool {
	return fsm_.Current() == OFF
}

func SwitchOff(fsm_ *FSM, ctx_ *Context) bool {
	return fsm_.Current() == ON
}

func LoggedIn(fsm_ *FSM, ctx_ *Context) bool {
	return ADMINDASH == fsm_.Current() || USERDASH == fsm_.Current()
}

func LoginAdmin(fsm_ *FSM, ctx_ *Context) bool {
	return string(ctx_.Context()) == ADMIN
}

func LoginUser(fsm_ *FSM, ctx_ *Context) bool {
	return string(ctx_.Context()) == USER
}

func TestFSM(t *testing.T) {

	type BoolTest struct {
		name    string
		response bool
		fn      func() bool
	}

	// check lightswitch.
	{
		_switch := NewLightswitch()
		switchTests := []BoolTest{
			{"Initial State", true, func() bool { return _switch.Current() == OFF }},
			{"Off Fail", false, func() bool { return _switch.Transition(OFF) }},
			{"On Success", true, func() bool { return _switch.Transition(ON) }},
			{"On Fail", false, func() bool { return _switch.Transition(ON) }},
			{"Off Success", true, func() bool { return _switch.Transition(OFF) }},
			{"Removed On Success", true, func() bool { return _switch.RemoveTransition(OFF,ON) }},
			{"Doesn't Contain On", false, func() bool { return _switch.Next().Contains(ON) }},
			{"On Fail", false, func() bool { return _switch.Transition(ON) }},
		}
		for _, tc := range switchTests {
			if tc.fn() != tc.response {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.response)
			}
		}
	}
	
	// check login.
	{
		_ctx := NewAdminContext()	
		_login := NewLoginManager(_ctx)
		loginTests := []BoolTest{
			{"Initial State", true, func() bool { return _login.Current() == OUT }},
			{"Logout Fail", false, func() bool { return _login.Transition(OUT) }},
			{"Contains Admin Transision", true, func() bool { return _login.Next().Contains(ADMINDASH) }},
			{"Contains User Transition", true, func() bool { return _login.Next().Contains(USERDASH) }},
			{"Doesn't Contain Logout Transition", false, func() bool { return _login.Next().Contains(OUT) }},
			{"Login to wrong account fail", false, func() bool { return _login.Transition(USERDASH) }},
			{"Login to admin success", true, func() bool { return _login.Transition(ADMINDASH) }},
			{"Admin state check", true, func() bool { return _login.Current() == ADMINDASH }},
			{"Logout success", true, func() bool { return _login.Transition(OUT) }},
		}
		for _, tc := range loginTests {
			if tc.fn() != tc.response {
				t.Errorf("%s: have: %d want: %d", tc.name, tc.fn(), tc.response)
			}
		}
	}
}
