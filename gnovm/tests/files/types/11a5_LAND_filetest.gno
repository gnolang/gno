package main

func main() {
	println(true && 0)
}

// Error:
// main/files/types/11a5_LAND_filetest.gno:4: operator && not defined on: <untyped> bigint
