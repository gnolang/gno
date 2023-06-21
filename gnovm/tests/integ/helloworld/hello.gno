package helloworld

import "gno.land/r/demo/helloworld/world"

func SayHelloWorld() string {
	return "Hello " + world.World()
}
