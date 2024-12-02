// PKGPATH: gno.land/r/ptr_map
package ptr_map

import "fmt"

type MyStruct struct {
	Name string
}

var (
	i0 = &MyStruct{Name: "initial"}

	m = map[[1]*MyStruct]string{[1]*MyStruct{i0}: "zero"}

	i1 = &MyStruct{Name: "alice"}
	i2 = &MyStruct{Name: "bob"}
)

func mutate() {
	i1.Name = "Clark"
}

func main() {

	// Create an array of pointers to MyStruct
	key1 := [1]*MyStruct{i1}
	key2 := [1]*MyStruct{i2}

	// Add entries to the map using arrays of pointers as keys
	m[key1] = "first key"
	m[key2] = "second key"

	// Access the value by key
	fmt.Println("Value for key1:", m[key1])
	fmt.Println("Value for key2:", m[key2])
	//
	// Example of pointer equality
	fmt.Println("Are key1 and key2 equal?", key1 == key2)

	mutate()
	fmt.Println("Value for key1:", m[key1])
}

// Output:
// Value for key1: first key
// Value for key2: second key
// Are key1 and key2 equal? false
// Value for key1: first key
