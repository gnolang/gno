package main

// Regression test for the nil-func branch of doOpDefer. Deferring a call of a
// nil func value with arguments must (a) raise call-of-nil when the deferred
// call runs and (b) leave the value stack balanced. The nil branch previously
// popped only the func value and left the evaluated arguments on the stack;
// the leak was masked by frame truncation on return, but the branch must still
// consume its arguments like the *FuncValue / *BoundMethodValue branches do.

func main() {
	defer func() {
		if r := recover(); r != nil {
			println("recovered")
		}
	}()
	var nf func(int, int)
	defer nf(1, 2)
	println("body")
}

// Output:
// body
// recovered
