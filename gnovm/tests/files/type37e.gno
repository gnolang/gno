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
// main/type37e.gno:8:1-10:2: invalid receiver type main.Int2 (base type is pointer type)

// TypeCheckError:
// main/type37e.gno:8:10: invalid receiver type Int2 (pointer or interface type); main/type37e.gno:13:19: Int2(ip).Int undefined (type Int2 has no field or method Int)
