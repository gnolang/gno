package main

import "gno.land/r/morgan/caller"

func main() { caller.StoreClosure(cross, func() { println("12345") }) }
