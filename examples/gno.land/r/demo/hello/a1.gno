package hello

import "std"

func SayHello() {
	println("say hello from A, going to call B")

	call := &std.CallMsg{
	    PkgPath: "gno.land/r/demo/greet",
	    Fn: "Greet",     
	    Args: []string{"hey"}, 
	}

	callback := &std.CallMsg{
	    PkgPath: "gno.land/r/demo/hello",
	    Fn: "Reply",     
	}

	// std.AsyncCall(call, callback)

	r := std.SyncCall(call)
	println("res of sync call is: ", r)

	println("this is something behind async operation")
}

func Reply(r string){
	println("receive callback: ", r)
}