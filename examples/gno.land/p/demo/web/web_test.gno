package web

 import (
 	"std"
 	"testing"	 
	 "gno.land/p/demo/ufmt"
 )

const (
	helloString string = "Hello World"
	thingsString string = "Hello Things"
	thingString string = "Hello Thing: %s"
	thingName string = "bob"
)

 func HelloWorld(req_ *Request) *Response {
	return req_.Respond(helloString, Stringer)
 }
 
 func Things(req_ *Request) *Response {
	return req_.Respond(thingsString, Stringer)
 }
 
 func Thing(req_ *Request) *Response {	
	return req_.Respond(ufmt.Sprintf(thingString, req_.Value("thingID")), Stringer)
 }

func ThingsRouter() *Router {
	_r := NewRouter()
	_r.AddRoute("",HelloWorld)
	_r.AddRoute("things",Things)
	_r.AddRoute("thing/{thingID}",Thing)
	return _r
}

 func TestWeb(t *testing.T) {
 	type BooleanTest struct {
 		name    string
 		response bool
 		fn      func() bool
 	}

 	// walk through creating organic and org ids, asserting only one of each per person
 	{
		t.Logf("Starting WebApp Test")
		_router := ThingsRouter()
		t.Logf("Created Things Router")

		tests := []BooleanTest{
			{"Hello World", true, func() bool { 
				_req, _action := _router.Resolve("")				
				return _action(_req).Render() == helloString
			}},
			{"Hello Things", true, func() bool { 
				_req, _action := _router.Resolve("things")
				return _action(_req).Render() == thingsString
			}},
			{"Hello Thing", true, func() bool { 
				_req, _action := _router.Resolve(ufmt.Sprintf("thing/%s", thingName))
				return _action(_req).Render() == ufmt.Sprintf(thingString, thingName)
			}},
		}
		for _, tc := range tests {
			if tc.fn() != tc.response {
				t.Errorf("%s: have: %t want: %t", tc.name, tc.fn(), tc.response)
			}
		}
		
		// _helloReq, _helloAction := _router.Resolve("/")
		// t.Logf("Hello World Response: %s", _helloAction(_helloReq).Render())
		
		// _thingsReq, _thingsAction := _router.Resolve("/things")
		// t.Logf("Things Response: %s", _thingsAction(_thingsReq).Render())
		
		// _thingReq, _thingAction := _router.Resolve("/thing/bob")
		// t.Logf("Thing Response: %s", _thingAction(_thingReq).Render())

		// t.Errorf("SUCCESS! FAILING NOW!")					
 	}	
 }


 func panicValue(fn func()) (recovered interface{}) {
     defer func() {
         recovered = recover()
     }()
     fn()    
 	return
 }