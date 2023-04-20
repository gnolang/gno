package hello

import "std"

func SayHello() {
	println("say hello from A, going to call B")

	// call := "gno.land/r/demo/greet#Greet#hey&hi&ho"
	call := "gno.land/r/demo/greet#Greet#hey"
	callback := "gno.land/r/demo/hello#Reply"

	std.Send(call, callback)

	println("this is something behind async operation")
}

func Reply(r string){
	println("receive callback: ", r)
}