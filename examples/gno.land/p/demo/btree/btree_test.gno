package btree

import (
	"errors"
	"testing"
)

func TestBTreeCreation(t *testing.T) {
	invalid := errors.New("invalid t")
	tests := []struct {
		name     string
		t        int
		expected error
	}{
		{
			"Negative t",
			-1,
			invalid,
		},
		{
			"Zero t",
			0,
			invalid,
		},
		{
			"t equal to 1",
			1,
			invalid,
		},
		{
			"Valid t",
			2,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewBTree(tt.t)
			if tt.expected == nil {
				if err != nil {
					t.Errorf("%s, expected <nil>", err)
				}
			} else {
				if err.Error() != tt.expected.Error() {
					t.Errorf("%s, expected %s", err, tt.expected)
				}
			}
		})
	}
}

func TestSearchBTree(t *testing.T) {
	dataGenerated := []int{
		1, 95, 45, 22, 31, 67, 78, 0, 11,
	}

	tree, err := NewBTree(2)
	if err != nil {
		t.Errorf("Unexpected error: %s; the value is valid", err)
	}
	found, err := tree.Search(-1)
	if err == nil {
		t.Errorf("%s, tree is empty", err)
	}

	for _, key := range dataGenerated {
		tree.Insert(Content{Key: key})
	}
	for _, key := range dataGenerated {
		found, err := tree.Search(key)
		if err != nil {
			t.Errorf("%s, Expected finding value", err)
		}
	}
}

func TestDeleteBTree(t *testing.T) {
	dataGenerated := []int{
		95, 45, 1, 22, 31, 67, 78, 0, 11, 2, 4, 10, 40,
	}

	tree, err := NewBTree(7)
	if err != nil {
		t.Errorf("Unexpected error: %s; the value is valid", err)
	}
	found, err := tree.Search(-1)
	if err == nil {
		t.Errorf("%s, tree is empty", err)
	}

	for _, key := range dataGenerated {
		tree.Insert(Content{Key: key})
	}
	for _, key := range dataGenerated {
		found, err := tree.Search(key)
		if err != nil {
			t.Errorf("%s, Expected finding value", err)
		}
	}
}

func TestGetSmallestBTree(t *testing.T) {
	dataGenerated := []int{
		95, 45, 1, 22, 31, 67, 78, 0, 11,
	}
	smallestOrdered := []int{
		95, 45, 1, 1, 1, 1, 1, 0, 0,
	}

	tree, err := NewBTree(2)
	if err != nil {
		t.Errorf("Unexpected error: %s; the value is valid", err)
	}
	_, err = tree.GetSmallest()
	if err == nil {
		t.Errorf("Unexpected error: %s; the tree is empty", err)
	}

	for idx, key := range dataGenerated {
		tree.Insert(Content{Key: key})
		mC, err := tree.GetSmallest()
		m := mC.Key
		if err != nil {
			t.Errorf("%s, Expected", err)
		}
		if m != smallestOrdered[idx] {
			t.Errorf("%s, Expected minimum value %d", err, m)
		}
	}
}

func TestGetGreatestBTree(t *testing.T) {
	dataGenerated := []int{
		11, 8, 78, 67, 31, 22, 1, 95, 45,
	}
	greatestOrdered := []int{
		11, 11, 78, 78, 78, 78, 78, 95, 95,
	}

	tree, err := NewBTree(3)
	if err != nil {
		t.Errorf("Unexpected error: %s; the value is valid", err)
	}
	_, err = tree.GetGreatest()
	if err == nil {
		t.Errorf("Unexpected error: %s; the tree is empty", err)
	}

	for idx, key := range dataGenerated {
		tree.Insert(Content{Key: key})
		mC, err := tree.GetGreatest()
		m := mC.Key
		if err != nil {
			t.Errorf("%s, Expected", err)
		}
		if m != greatestOrdered[idx] {
			t.Errorf("%s, Expected maximum value %d", err, m)
		}
	}
}

func TestGetSizeBTree(t *testing.T) {
	data := []int{
		11, 8, 78, 67, 31, 22, 1, 95, 45,
	}
	tree, err := NewBTree(3)

	// Verifying the size is continuously increased
	for idx, key := range data {
		tree.Insert(Content{Key: key})
		size := tree.GetSize()
		if size != idx+1 {
			t.Errorf("%d, Expected size %d", size, idx+1)
		}
	}
	lastSize := tree.GetSize()

	// Verify the size is affected by faking removing a value not in the tree
	tree.Delete(-1)
	if lastSize != tree.GetSize() {
		t.Errorf("No key deleted and size is %d, expected size %d", tree.GetSize(), lastSize)
	}

	// Verify the size is affected by removing a value in the tree
	tree.Delete(11)
	if lastSize-1 != tree.GetSize() {
		t.Errorf("Key deleted and size %d, expected size %d", tree.GetSize(), lastSize)
	}

	// Case for tree with diff grade
	tree, _ = NewBTree(7)

	// Verifying the size is continuously increased
	for i := 0; i <= 100; i += 3 {
		tree.Insert(Content{Key: i})
	}
	lastSize = tree.GetSize()

	// Verify the size is affected by faking removing a value not in the tree
	tree.Delete(-1)
	if lastSize != tree.GetSize() {
		t.Errorf("No key deleted and size is %d, expected size %d", tree.GetSize(), lastSize)
	}

	// Verify the size is affected by removing a value in the tree
	tree.Delete(69)
	if lastSize-1 != tree.GetSize() {
		t.Errorf("Key deleted and size %d, expected size %d", tree.GetSize(), lastSize)
	}
}

func TestValuesInOrderTree(t *testing.T) {
	tree, _ := NewBTree(3)

	// Test case tree is empty
	values := tree.ValuesInOrder()
	if len(values) != 0 {
		t.Errorf("Expected empty tree, got %v", values)
	}

	// Test case tree is not empty

	// Adding values to the tree
	dataGenerated := []int{95, 45, 1, 22, 31, 67, 78, 0, 11, 2, 4, 10, 40}
	for _, key := range dataGenerated {
		tree.Insert(Content{Key: key})
	}
	dataGenerated = []int{0, 1, 2, 4, 10, 11, 22, 31, 40, 45, 67, 78, 95}

	values = tree.ValuesInOrder()
	if len(values) != len(dataGenerated) {
		t.Errorf("Expected %d values, got %d", len(dataGenerated), len(values))
	}

	for idx, key := range values {
		if key != dataGenerated[idx] {
			t.Errorf("Expected value %d, got %d in index %d", dataGenerated[idx], key, idx)
		}
	}
}

// Simulate the usage of the tree as a whole
func TestAllAsAWhole(t *testing.T) {
	tree, _ := NewBTree(3)

	values := tree.ValuesInOrder()
	if len(values) != 0 {
		t.Errorf("Expected empty tree, got %v", values)
	}

	dataGenerated := []int{95, 45, 1, 22, 31, 67, 78, 0, 11, 2, 4, 10, 40}
	for _, key := range dataGenerated {
		tree.Insert(Content{Key: key})
	}

	dataGenerated = []int{0, 1, 2, 4, 11, 22, 31, 40, 67, 78, 95}

	tree.Delete(88)
	tree.Delete(10)
	tree.Delete(45)

	values = tree.ValuesInOrder()
	if len(values) != len(dataGenerated) {
		t.Errorf("Expected %d values, got %d", len(dataGenerated), len(values))
	}
	for idx, key := range values {
		if key != dataGenerated[idx] {
			t.Errorf("Expected value %d, got %d in index %d", dataGenerated[idx], key, idx)
		}
	}
}
