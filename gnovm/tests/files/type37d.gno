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
// main/files/type37d.gno:7: invalid receiver type main.IntPtr (base type is pointer type)
