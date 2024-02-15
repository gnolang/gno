package hello

import (
	"gno.land/p/demo/ufmt"
)

func Hello() string {
	//println(int(1) + int8(1))
	println("hey")
	println(ufmt.Sprintf("%s \n", "hey"))
	return "hello world"
}

//func main() {
//	Hello()
//}
