// https://github.com/gnolang/gno/issues/3727

package main

func () A()

func main() {}

// Error:
// files/parse_err0.gno:5:1: method has no receiver

// TypeCheckError:
// main/files/parse_err0.gno:5:6: method has no receiver
