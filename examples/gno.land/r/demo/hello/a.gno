package hello

import "std"

func SayHello() {
	println("say hello from A, going to call B")

	call := &std.CallMsg{
	    PkgPath: "gno.land/r/demo/greet",
	    Fn: "Greet",     
	    Args: []string{"hey"}, 
	}
	call := "gno.land/r/demo/greet/Greet/hey"

	callback := &std.CallMsg{
	    PkgPath: "gno.land/r/demo/hello",
	    Fn: "Reply",     
	}

	std.AsyncCall(call, callback)

	println("this is something behind async operation")
}

func Reply(r string){
	println("receive callback: ", r)
}