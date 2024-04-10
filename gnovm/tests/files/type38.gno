package main

type IntPtr *int

var ip IntPtr = new(int)

func (p IntPtr) Int() int {
	return *p
}

func main() {
	println(ip.Int())
}

// Error:
// main/debug/type38.gno:7: invalid receiver type IntPtr (pointer or interface type)
