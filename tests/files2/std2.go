package main

import "std"

func main() {
	caller := std.GetCaller()
	result := std.ToBech32(caller)
	println(result)
}

// Output:
// g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj
