package main

var T struct{}

func (t T) do() {}

func main() {}

// Error:
// T (variable of type struct{}) is not a type
