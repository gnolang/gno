package main

// Regression test: `defer f(g())` where the single syntactic argument is a
// multi-value call must bind the deferred func and its expanded arguments
// correctly. doOpDefer previously used len(Call.Args) (== 1 here) instead of
// Call.NumArgs (the expanded value count == 2), so it peeked an argument value
// as the func. Case 1 then panicked with "defer called a nil function"; case 2
// ran the wrong function (marker) instead of f.

func twovals() (int, int) { return 7, 8 }
func h(a, b int)          { println("h:", a+b) }

func marker()            { println("WRONG: marker ran") }
func g() (func(), int)   { return marker, 99 }
func f(fn func(), n int) { println("f: n =", n) }

func main() {
	defer f(g())       // runs 2nd: must print "f: n = 99", must not run marker
	defer h(twovals()) // runs 1st: must print "h: 15"
	println("body")
}

// Output:
// body
// h: 15
// f: n = 99
