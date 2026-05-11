package main

type arr1 [1]int

func main() {
	var i interface{} = arr1{1}
	_ = i.(arr1)[:]
}

// Error:
// main/addressable_6e_err.gno:7:6-17: invalid operation: i<VPBlock(1,0)>.(typeval{main.arr1}) (slice of unaddressable value)

// TypeCheckError:
// main/addressable_6e_err.gno:7:6: invalid operation: cannot slice i.(arr1) (comma, ok expression of array type arr1) (value not addressable)
