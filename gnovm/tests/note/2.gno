package main

var e error

func main() {
	e = 1 // this check is currently happening in checkOrConvert, should be in assertCompatibility
}

// Error:
// main/debug/2.gnoa:6: <untyped> bigint does not implement .uverse.error
