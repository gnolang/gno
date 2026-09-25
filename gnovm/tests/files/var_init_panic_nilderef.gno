package main

var p *int
var A = *p

func main() {}

// Error:
// runtime error: nil pointer dereference
