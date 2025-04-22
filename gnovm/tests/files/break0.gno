package main

func main() {
	break
}

// Error:
// main/files/break0.gno:4:2: break statement out of place

// TypeCheckError:
// main/files/break0.gno:4:2: break not in for, switch, or select statement
