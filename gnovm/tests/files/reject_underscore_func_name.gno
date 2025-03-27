// https://github.com/gnolang/gno/issues/4019

package main

func _() {}

func main() {}

// Error:
// main/files/reject_underscore_func_name.gno:5:1: "_" as a function name is not permitted, cannot be referenced
