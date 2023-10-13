package greet

func Greet(name string, country string, age int, isSomething bool) string {

	//chainID := std.GetChainID()
	//call := std.EncodeMsgCall(chainID, "gno.land/r/demo/x/calls/await/hola", "Hola", []string{"hola"})
	//
	//r, err := std.Await(call)
	//println("done ")
	//println("r is :", r)
	//println("errMsg is: ", err)
	//return "greet" + r

	println("name is: ", name)
	println("country is: ", country)
	println("age is: ", age)
	println("isSomething is: ", isSomething)

	return "greet"
}
