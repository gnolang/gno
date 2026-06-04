package btree

import (
	"fmt"
	"sort"
	"testing"
)

// Content represents a key-value pair where the Key can be either an int or string
// and the Value can be any type.
type Content struct {
	Key   any
	Value any
}

// Less compares two Content records by their Keys.
// The Key must be either an int or a string.
func (c Content) Less(than Record) bool {
	other, ok := than.(Content)
	if !ok {
		panic("cannot compare: incompatible types")
	}

	switch key := c.Key.(type) {
	case int:
		switch otherKey := other.Key.(type) {
		case int:
			return key < otherKey
		case string:
			return true // ints are always less than strings
		default:
			panic("unsupported key type: must be int or string")
		}
	case string:
		switch otherKey := other.Key.(type) {
		case int:
			return false // strings are always greater than ints
		case string:
			return key < otherKey
		default:
			panic("unsupported key type: must be int or string")
		}
	default:
		panic("unsupported key type: must be int or string")
	}
}

type ContentSlice []Content

func (s ContentSlice) Len() int {
	return len(s)
}

func (s ContentSlice) Less(i, j int) bool {
	return s[i].Less(s[j])
}

func (s ContentSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ContentSlice) Copy() ContentSlice {
	newSlice := make(ContentSlice, len(s))
	copy(newSlice, s)
	return newSlice
}

// Ensure Content implements the Record interface.
var _ Record = Content{}

// ****************************************************************************
// Test helpers
// ****************************************************************************

func genericSeeding(tree *BTree, size int) *BTree {
	for i := 0; i < size; i++ {
		tree.Insert(Content{Key: i, Value: fmt.Sprintf("Value_%d", i)})
	}
	return tree
}

func intSlicesCompare(left, right []int) int {
	if len(left) != len(right) {
		if len(left) > len(right) {
			return 1
		} else {
			return -1
		}
	}

	for position, leftInt := range left {
		if leftInt != right[position] {
			if leftInt > right[position] {
				return 1
			} else {
				return -1
			}
		}
	}

	return 0
}

// ****************************************************************************
// Tests
// ****************************************************************************

func TestLen(t *testing.T) {
	length := genericSeeding(New(WithDegree(10)), 7).Len()
	if length != 7 {
		t.Errorf("Length is incorrect. Expected 7, but got %d.", length)
	}

	length = genericSeeding(New(WithDegree(5)), 111).Len()
	if length != 111 {
		t.Errorf("Length is incorrect. Expected 111, but got %d.", length)
	}

	length = genericSeeding(New(WithDegree(30)), 123).Len()
	if length != 123 {
		t.Errorf("Length is incorrect. Expected 123, but got %d.", length)
	}
}

func TestHas(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 40)

	if tree.Has(Content{Key: 7}) != true {
		t.Errorf("Has(7) reported false, but it should be true.")
	}
	if tree.Has(Content{Key: 39}) != true {
		t.Errorf("Has(40) reported false, but it should be true.")
	}
	if tree.Has(Content{Key: 1111}) == true {
		t.Errorf("Has(1111) reported true, but it should be false.")
	}
}

func TestMin(t *testing.T) {
	min := genericSeeding(New(WithDegree(10)), 53).Min().(Content)

	if min.Key != 0 {
		t.Errorf("Minimum should have been 0, but it was reported as %d.", min)
	}
}

func TestMax(t *testing.T) {
	max := genericSeeding(New(WithDegree(10)), 53).Max().(Content)

	if max.Key != 52 {
		t.Errorf("Maximum should have been 52, but it was reported as %d.", max)
	}
}

func TestGet(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 40)

	if val := tree.Get(Content{Key: 7}); val != nil && val.(Content).Value != "Value_7" {
		t.Errorf("Get(7) should have returned 'Value_7', but it returned %v.", val)
	}
	if val := tree.Get(Content{Key: 39}); val != nil && val.(Content).Value != "Value_39" {
		t.Errorf("Get(39) should have returned 'Value_39', but it returned %v.", val)
	}
	if val := tree.Get(Content{Key: 1111}); val != nil {
		t.Errorf("Get(1111) returned %v, but it should be nil.", val)
	}
}

func TestDescend(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 5)

	expected := []int{4, 3, 2, 1, 0}
	found := []int{}

	tree.Descend(func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("Descend returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestDescendGreaterThan(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 10)

	expected := []int{9, 8, 7, 6, 5}
	found := []int{}

	tree.DescendGreaterThan(Content{Key: 4}, func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("DescendGreaterThan returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestDescendLessOrEqual(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 10)

	expected := []int{4, 3, 2, 1, 0}
	found := []int{}

	tree.DescendLessOrEqual(Content{Key: 4}, func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("DescendLessOrEqual returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestDescendRange(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 10)

	expected := []int{6, 5, 4, 3, 2}
	found := []int{}

	tree.DescendRange(Content{Key: 6}, Content{Key: 1}, func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("DescendRange returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestAscend(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 5)

	expected := []int{0, 1, 2, 3, 4}
	found := []int{}

	tree.Ascend(func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("Ascend returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestAscendGreaterOrEqual(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 10)

	expected := []int{5, 6, 7, 8, 9}
	found := []int{}

	tree.AscendGreaterOrEqual(Content{Key: 5}, func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("AscendGreaterOrEqual returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestAscendLessThan(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 10)

	expected := []int{0, 1, 2, 3, 4}
	found := []int{}

	tree.AscendLessThan(Content{Key: 5}, func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("AscendLessThan returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestAscendRange(t *testing.T) {
	tree := genericSeeding(New(WithDegree(10)), 10)

	expected := []int{2, 3, 4, 5, 6}
	found := []int{}

	tree.AscendRange(Content{Key: 2}, Content{Key: 7}, func(_record Record) bool {
		record := _record.(Content)
		found = append(found, record.Key.(int))
		return true
	})

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("AscendRange returned the wrong sequence. Expected %v, but got %v.", expected, found)
	}
}

func TestDeleteMin(t *testing.T) {
	tree := genericSeeding(New(WithDegree(3)), 100)

	expected := []int{0, 1, 2, 3, 4}
	found := []int{}

	found = append(found, tree.DeleteMin().(Content).Key.(int))
	found = append(found, tree.DeleteMin().(Content).Key.(int))
	found = append(found, tree.DeleteMin().(Content).Key.(int))
	found = append(found, tree.DeleteMin().(Content).Key.(int))
	found = append(found, tree.DeleteMin().(Content).Key.(int))

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("5 rounds of DeleteMin returned the wrong elements. Expected  %v, but got %v.", expected, found)
	}
}

func TestShift(t *testing.T) {
	tree := genericSeeding(New(WithDegree(3)), 100)

	expected := []int{0, 1, 2, 3, 4}
	found := []int{}

	found = append(found, tree.Shift().(Content).Key.(int))
	found = append(found, tree.Shift().(Content).Key.(int))
	found = append(found, tree.Shift().(Content).Key.(int))
	found = append(found, tree.Shift().(Content).Key.(int))
	found = append(found, tree.Shift().(Content).Key.(int))

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("5 rounds of Shift returned the wrong elements. Expected  %v, but got %v.", expected, found)
	}
}

func TestDeleteMax(t *testing.T) {
	tree := genericSeeding(New(WithDegree(3)), 100)

	expected := []int{99, 98, 97, 96, 95}
	found := []int{}

	found = append(found, tree.DeleteMax().(Content).Key.(int))
	found = append(found, tree.DeleteMax().(Content).Key.(int))
	found = append(found, tree.DeleteMax().(Content).Key.(int))
	found = append(found, tree.DeleteMax().(Content).Key.(int))
	found = append(found, tree.DeleteMax().(Content).Key.(int))

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("5 rounds of DeleteMax returned the wrong elements. Expected  %v, but got %v.", expected, found)
	}
}

func TestPop(t *testing.T) {
	tree := genericSeeding(New(WithDegree(3)), 100)

	expected := []int{99, 98, 97, 96, 95}
	found := []int{}

	found = append(found, tree.Pop().(Content).Key.(int))
	found = append(found, tree.Pop().(Content).Key.(int))
	found = append(found, tree.Pop().(Content).Key.(int))
	found = append(found, tree.Pop().(Content).Key.(int))
	found = append(found, tree.Pop().(Content).Key.(int))

	if intSlicesCompare(expected, found) != 0 {
		t.Errorf("5 rounds of Pop returned the wrong elements. Expected  %v, but got %v.", expected, found)
	}
}

func TestInsertGet(t *testing.T) {
	tree := New(WithDegree(4))

	expected := []Content{}

	for count := 0; count < 20; count++ {
		value := fmt.Sprintf("Value_%d", count)
		tree.Insert(Content{Key: count, Value: value})
		expected = append(expected, Content{Key: count, Value: value})
	}

	for count := 0; count < 20; count++ {
		val := tree.Get(Content{Key: count})
		if val == nil || val.(Content) != expected[count] {
			t.Errorf("Insert/Get doesn't appear to be working. Expected to retrieve %v with key %d, but got %v.", expected[count], count, val)
		}
	}
}

func TestClone(t *testing.T) {
	// Implement the clone test
}

// ***** The following tests are functional or stress testing type tests.

func TestBTree(t *testing.T) {
	// Create a B-Tree of degree 3
	tree := New(WithDegree(3))

	// insertData := []Content{}
	var insertData ContentSlice

	// Insert integer keys
	intKeys := []int{10, 20, 5, 6, 12, 30, 7, 17}
	for _, key := range intKeys {
		content := Content{Key: key, Value: fmt.Sprintf("Value_%d", key)}
		insertData = append(insertData, content)
		result := tree.Insert(content)
		if result != nil {
			t.Errorf("**** Already in the tree?  %v", result)
		}
	}

	// Insert string keys
	stringKeys := []string{"apple", "banana", "cherry", "date", "fig", "grape"}
	for _, key := range stringKeys {
		content := Content{Key: key, Value: fmt.Sprintf("Fruit_%s", key)}
		insertData = append(insertData, content)
		tree.Insert(content)
	}

	if tree.Len() != 14 {
		t.Errorf("Tree length wrong. Expected 14 but got %d", tree.Len())
	}

	// Search for existing and non-existing keys
	searchTests := []struct {
		test     Content
		expected bool
	}{
		{Content{Key: 10, Value: "Value_10"}, true},
		{Content{Key: 15, Value: ""}, false},
		{Content{Key: "banana", Value: "Fruit_banana"}, true},
		{Content{Key: "kiwi", Value: ""}, false},
	}

	t.Logf("Search Tests:\n")
	for _, test := range searchTests {
		val := tree.Get(test.test)

		if test.expected {
			if val != nil && val.(Content).Value == test.test.Value {
				t.Logf("Found expected key:value %v:%v", test.test.Key, test.test.Value)
			} else {
				if val == nil {
					t.Logf("Didn't find %v, but expected", test.test.Key)
				} else {
					t.Errorf("Expected key %v:%v, but found %v:%v.", test.test.Key, test.test.Value, val.(Content).Key, val.(Content).Value)
				}
			}
		} else {
			if val != nil {
				t.Errorf("Did not expect key %v, but found key:value %v:%v", test.test.Key, val.(Content).Key, val.(Content).Value)
			} else {
				t.Logf("Didn't find %v, but wasn't expected", test.test.Key)
			}
		}
	}

	// Iterate in order
	t.Logf("\nIn-order Iteration:\n")
	pos := 0

	if tree.Len() != 14 {
		t.Errorf("Tree length wrong. Expected 14 but got %d", tree.Len())
	}

	sortedInsertData := insertData.Copy()
	sort.Sort(sortedInsertData)

	t.Logf("Insert Data Length: %d", len(insertData))
	t.Logf("Sorted Data Length: %d", len(sortedInsertData))
	t.Logf("Tree Length: %d", tree.Len())

	tree.Ascend(func(_record Record) bool {
		record := _record.(Content)
		t.Logf("Key:Value == %v:%v", record.Key, record.Value)
		if record.Key != sortedInsertData[pos].Key {
			t.Errorf("Out of order! Expected %v, but got %v", sortedInsertData[pos].Key, record.Key)
		}
		pos++
		return true
	})
	// // Reverse Iterate
	t.Logf("\nReverse-order Iteration:\n")
	pos = len(sortedInsertData) - 1

	tree.Descend(func(_record Record) bool {
		record := _record.(Content)
		t.Logf("Key:Value == %v:%v", record.Key, record.Value)
		if record.Key != sortedInsertData[pos].Key {
			t.Errorf("Out of order! Expected %v, but got %v", sortedInsertData[pos].Key, record.Key)
		}
		pos--
		return true
	})

	deleteTests := []Content{
		{Key: 10, Value: "Value_10"},
		{Key: 15, Value: ""},
		{Key: "banana", Value: "Fruit_banana"},
		{Key: "kiwi", Value: ""},
	}
	for _, test := range deleteTests {
		fmt.Printf("\nDeleting %+v\n", test)
		tree.Delete(test)
	}

	if tree.Len() != 12 {
		t.Errorf("Tree length wrong. Expected 12 but got %d", tree.Len())
	}

	for _, test := range deleteTests {
		val := tree.Get(test)
		if val != nil {
			t.Errorf("Did not expect key %v, but found key:value %v:%v", test.Key, val.(Content).Key, val.(Content).Value)
		} else {
			t.Logf("Didn't find %v, but wasn't expected", test.Key)
		}
	}
}

// Write a test that populates a large B-Tree with 1000 records.
// It should then `Clone` the tree, make some changes to both the original and the clone,
// And then clone the clone, and make some changes to all three trees, and then check that the changes are isolated
// to the tree they were made in.
func TestBTreeCloneIsolation(t *testing.T) {
	t.Logf("Creating B-Tree of degree 10 with 1000 records\n")
	size := 1000
	tree := genericSeeding(New(WithDegree(10)), size)

	// Clone the tree
	t.Logf("Cloning the tree\n")
	clone := tree.Clone()

	// Make some changes to the original and the clone
	t.Logf("Making changes to the original and the clone\n")
	for i := 0; i < size; i += 2 {
		content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
		tree.Delete(content)
		content = Content{Key: i + 1, Value: fmt.Sprintf("Value_%d", i+1)}
		clone.Delete(content)
	}

	// Clone the clone
	t.Logf("Cloning the clone\n")
	clone2 := clone.Clone()

	// Make some changes to all three trees
	t.Logf("Making changes to all three trees\n")
	for i := 0; i < size; i += 3 {
		content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
		tree.Delete(content)
		content = Content{Key: i, Value: fmt.Sprintf("Value_%d", i+1)}
		clone.Delete(content)
		content = Content{Key: i + 2, Value: fmt.Sprintf("Value_%d", i+2)}
		clone2.Delete(content)
	}

	// Check that the changes are isolated to the tree they were made in
	t.Logf("Checking that the changes are isolated to the tree they were made in\n")
	for i := 0; i < size; i++ {
		content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
		val := tree.Get(content)

		if i%3 == 0 || i%2 == 0 {
			if val != nil {
				t.Errorf("Didn't expect key %v, but found key:value %v:%v", content.Key, val.(Content).Key, val.(Content).Value)
			}
		} else {
			if val == nil {
				t.Errorf("Expected key %v, but didn't find it", content.Key)
			}
		}

		val = clone.Get(content)
		if i%2 != 0 || i%3 == 0 {
			if val != nil {
				t.Errorf("Didn't expect key %v, but found key:value %v:%v", content.Key, val.(Content).Key, val.(Content).Value)
			}
		} else {
			if val == nil {
				t.Errorf("Expected key %v, but didn't find it", content.Key)
			}
		}

		val = clone2.Get(content)
		if i%2 != 0 || (i-2)%3 == 0 {
			if val != nil {
				t.Errorf("Didn't expect key %v, but found key:value %v:%v", content.Key, val.(Content).Key, val.(Content).Value)
			}
		} else {
			if val == nil {
				t.Errorf("Expected key %v, but didn't find it", content.Key)
			}
		}
	}
}

// --------------------
// Stress tests. Disabled for testing performance

//func TestStress(t *testing.T) {
//	// Loop through creating B-Trees with a range of degrees from 3 to 12, stepping by 3.
//	// Insert 1000 records into each tree, then search for each record.
//	// Delete half of the records, skipping every other one, then search for each record.
//
//	for degree := 3; degree <= 12; degree += 3 {
//		t.Logf("Testing B-Tree of degree %d\n", degree)
//		tree := New(WithDegree(degree))
//
//		// Insert 1000 records
//		t.Logf("Inserting 1000 records\n")
//		for i := 0; i < 1000; i++ {
//			content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
//			tree.Insert(content)
//		}
//
//		// Search for all records
//		for i := 0; i < 1000; i++ {
//			content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
//			val := tree.Get(content)
//			if val == nil {
//				t.Errorf("Expected key %v, but didn't find it", content.Key)
//			}
//		}
//
//		// Delete half of the records
//		for i := 0; i < 1000; i += 2 {
//			content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
//			tree.Delete(content)
//		}
//
//		// Search for all records
//		for i := 0; i < 1000; i++ {
//			content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
//			val := tree.Get(content)
//			if i%2 == 0 {
//				if val != nil {
//					t.Errorf("Didn't expect key %v, but found key:value %v:%v", content.Key, val.(Content).Key, val.(Content).Value)
//				}
//			} else {
//				if val == nil {
//					t.Errorf("Expected key %v, but didn't find it", content.Key)
//				}
//			}
//		}
//	}
//
//	// Now create a very large tree, with 100000 records
//	// Then delete roughly one third of them, using a very basic random number generation scheme
//	// (implement it right here) to determine which records to delete.
//	// Print a few lines using Logf to let the user know what's happening.
//
//	t.Logf("Testing B-Tree of degree 10 with 100000 records\n")
//	tree := New(WithDegree(10))
//
//	// Insert 100000 records
//	t.Logf("Inserting 100000 records\n")
//	for i := 0; i < 100000; i++ {
//		content := Content{Key: i, Value: fmt.Sprintf("Value_%d", i)}
//		tree.Insert(content)
//	}
//
//	// Implement a very basic random number generator
//	seed := 0
//	random := func() int {
//		seed = (seed*1103515245 + 12345) & 0x7fffffff
//		return seed
//	}
//
//	// Delete one third of the records
//	t.Logf("Deleting one third of the records\n")
//	for i := 0; i < 35000; i++ {
//		content := Content{Key: random() % 100000, Value: fmt.Sprintf("Value_%d", i)}
//		tree.Delete(content)
//	}
//}
