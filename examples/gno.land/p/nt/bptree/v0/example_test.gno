package bptree

// ExampleNew shows NewBPTree32() and adding a value
func ExampleNew() {
	var tree *BPTree
	tree = NewBPTree32()
	tree.Set("key0", "value0")

	var updated bool
	updated = tree.Set("key1", "value1")
	println(updated, tree.Size())

	// Output:
	// false 2
}

// ExampleUpdate shows NewBPTree32() and updating a value
func ExampleUpdate() {
	var tree *BPTree
	tree = NewBPTree32()
	tree.Set("key0", "value0")

	var updated bool
	updated = tree.Set("key0", "new_value0")
	println(updated, tree.Size())

	// Output:
	// true 1
}

// ExampleZeroValue shows updating the zero value of a BPTree without NewBPTree32()
func ExampleZeroValue() {
	var tree BPTree
	tree.Set("key0", "value0")
	tree.Set("key1", "value1")

	var updated bool
	updated = tree.Set("key2", "value2")
	println(updated, tree.Size())

	// Output:
	// false 3
}
