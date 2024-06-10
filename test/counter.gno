package main

// Define a simple struct with one field
type Counter struct {
	value int
}

// Method with a value receiver that attempts to modify the struct
func (c *Counter) Increment() {
	c.value++
}
