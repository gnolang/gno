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
// main/type37d.gno:7:1-9:2: invalid receiver type main.IntPtr (base type is pointer type)

// TypeCheckError:
// main/type37d.gno:7:9: invalid receiver type IntPtr (pointer or interface type); main/type37d.gno:12:13: ip.Int undefined (type IntPtr has no field or method Int)
