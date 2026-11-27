package main

var p *int
var A = *p

func main() {}

// Error:
// runtime error: invalid memory address or nil pointer dereference
