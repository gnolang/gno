package main

type IntPtr *int
type Int2 IntPtr

var ip IntPtr = new(int)

func (i2 Int2) Int() int {
	return *i2
}

func main() {
	println(Int2(ip).Int())
}

// Error:
// main/files/type37e.gno:8: invalid receiver type main.Int2 (base type is pointer type)
