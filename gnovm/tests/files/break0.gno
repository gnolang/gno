package main

func main() {
	break
}

// Error:
// main/break0.gno:4:2-7: break statement out of place

// TypeCheckError:
// main/break0.gno:4:2: break not in for, switch, or select statement
