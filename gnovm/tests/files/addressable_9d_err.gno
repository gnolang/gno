package main

func main() {
	m := map[string][1]int{"k": {1}}
	_ = m["k"][:]
}

// Error:
// main/addressable_9d_err.gno:5:6-15: invalid operation: m<VPBlock(1,0)>[(const ("k" string))] (slice of unaddressable value)

// TypeCheckError:
// main/addressable_9d_err.gno:5:6: invalid operation: cannot slice m["k"] (map index expression of type [1]int) (value not addressable)
