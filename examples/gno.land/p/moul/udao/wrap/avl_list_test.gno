package wrap

import (
	"testing"

	"gno.land/p/demo/avl/list"
)

// MockProposal for testing wrap functionality
type MockProposal struct {
	id    uint64
	title string
}

func (m MockProposal) ID() uint64                         { return m.id }
func (m MockProposal) Definition() interface{}            { return nil }
func (m MockProposal) GetState() string                   { return "active" }
func (m MockProposal) GetYeaPercentage() float64          { return 0 }
func (m MockProposal) GetNayPercentage() float64          { return 0 }
func (m MockProposal) GetAbstainPercentage() float64      { return 0 }
func (m MockProposal) GetPendingVoterPercentage() float64 { return 0 }

func TestPropListWrapper(t *testing.T) {
	// Create an AVL list with some mock proposals
	avlList := &list.List{}
	prop1 := MockProposal{id: 1, title: "First Proposal"}
	prop2 := MockProposal{id: 2, title: "Second Proposal"}
	prop3 := MockProposal{id: 3, title: "Third Proposal"}
	
	avlList.Append(prop1)
	avlList.Append(prop2)
	avlList.Append(prop3)
	
	// Wrap it as a PropList
	wrapper := WrapAsPropList(avlList)
	
	// Test Len()
	if wrapper.Len() != 3 {
		t.Errorf("expected len=3, got %d", wrapper.Len())
	}
	
	// Test Get()
	firstProp := wrapper.Get(0)
	mockProp1 := firstProp.(MockProposal)
	if mockProp1.ID() != 1 {
		t.Errorf("expected first proposal ID=1, got %d", mockProp1.ID())
	}
	if mockProp1.title != "First Proposal" {
		t.Errorf("expected title 'First Proposal', got '%s'", mockProp1.title)
	}
	
	// Test Get() for different indices
	secondProp := wrapper.Get(1)
	mockProp2 := secondProp.(MockProposal)
	if mockProp2.ID() != 2 {
		t.Errorf("expected second proposal ID=2, got %d", mockProp2.ID())
	}
	
	thirdProp := wrapper.Get(2)
	mockProp3 := thirdProp.(MockProposal)
	if mockProp3.ID() != 3 {
		t.Errorf("expected third proposal ID=3, got %d", mockProp3.ID())
	}
	
	// Test Slice()
	slice := wrapper.Slice(0, 2)
	if len(slice) != 2 {
		t.Errorf("expected slice len=2, got %d", len(slice))
	}
	
	sliceProp1 := slice[0].(MockProposal)
	sliceProp2 := slice[1].(MockProposal)
	if sliceProp1.ID() != 1 || sliceProp2.ID() != 2 {
		t.Errorf("unexpected slice IDs: %d, %d", sliceProp1.ID(), sliceProp2.ID())
	}
	
	// Test Slice() with different range
	middleSlice := wrapper.Slice(1, 3)
	if len(middleSlice) != 2 {
		t.Errorf("expected middle slice len=2, got %d", len(middleSlice))
	}
	
	middleProp1 := middleSlice[0].(MockProposal)
	middleProp2 := middleSlice[1].(MockProposal)
	if middleProp1.ID() != 2 || middleProp2.ID() != 3 {
		t.Errorf("unexpected middle slice IDs: %d, %d", middleProp1.ID(), middleProp2.ID())
	}
	
	// Test single element slice
	singleSlice := wrapper.Slice(1, 2)
	if len(singleSlice) != 1 {
		t.Errorf("expected single slice len=1, got %d", len(singleSlice))
	}
	
	singleProp := singleSlice[0].(MockProposal)
	if singleProp.ID() != 2 {
		t.Errorf("expected single slice ID=2, got %d", singleProp.ID())
	}
}

func TestPropListWrapperEmptyList(t *testing.T) {
	// Test with empty list
	emptyList := &list.List{}
	wrapper := WrapAsPropList(emptyList)
	
	if wrapper.Len() != 0 {
		t.Errorf("expected empty list len=0, got %d", wrapper.Len())
	}
	
	// Test empty slice
	emptySlice := wrapper.Slice(0, 0)
	if len(emptySlice) != 0 {
		t.Errorf("expected empty slice len=0, got %d", len(emptySlice))
	}
}

func TestPropListWrapperSingleItem(t *testing.T) {
	// Test with single item
	singleList := &list.List{}
	prop := MockProposal{id: 42, title: "Single Proposal"}
	singleList.Append(prop)
	
	wrapper := WrapAsPropList(singleList)
	
	if wrapper.Len() != 1 {
		t.Errorf("expected single list len=1, got %d", wrapper.Len())
	}
	
	retrieved := wrapper.Get(0)
	mockProp := retrieved.(MockProposal)
	if mockProp.ID() != 42 {
		t.Errorf("expected ID=42, got %d", mockProp.ID())
	}
	if mockProp.title != "Single Proposal" {
		t.Errorf("expected title 'Single Proposal', got '%s'", mockProp.title)
	}
	
	// Test slice with single item
	slice := wrapper.Slice(0, 1)
	if len(slice) != 1 {
		t.Errorf("expected slice len=1, got %d", len(slice))
	}
	
	sliceProp := slice[0].(MockProposal)
	if sliceProp.ID() != 42 {
		t.Errorf("expected slice ID=42, got %d", sliceProp.ID())
	}
}

func TestPropListWrapperLargeList(t *testing.T) {
	// Test with larger list
	largeList := &list.List{}
	expectedCount := 10
	
	for i := 0; i < expectedCount; i++ {
		prop := MockProposal{id: uint64(i + 1), title: "Proposal " + string(rune(i + 1))}
		largeList.Append(prop)
	}
	
	wrapper := WrapAsPropList(largeList)
	
	if wrapper.Len() != expectedCount {
		t.Errorf("expected len=%d, got %d", expectedCount, wrapper.Len())
	}
	
	// Test access to all items
	for i := 0; i < expectedCount; i++ {
		prop := wrapper.Get(i)
		mockProp := prop.(MockProposal)
		if mockProp.ID() != uint64(i+1) {
			t.Errorf("expected ID=%d at index %d, got %d", i+1, i, mockProp.ID())
		}
	}
	
	// Test various slice operations
	testCases := []struct {
		start    int
		end      int
		expected int
	}{
		{0, 3, 3},
		{2, 7, 5},
		{5, 10, 5},
		{0, 10, 10},
		{8, 9, 1},
	}
	
	for _, tc := range testCases {
		slice := wrapper.Slice(tc.start, tc.end)
		if len(slice) != tc.expected {
			t.Errorf("expected slice[%d:%d] len=%d, got %d",
				tc.start, tc.end, tc.expected, len(slice))
		}
		
		// Verify slice content
		for i, prop := range slice {
			mockProp := prop.(MockProposal)
			expectedID := uint64(tc.start + i + 1)
			if mockProp.ID() != expectedID {
				t.Errorf("expected slice[%d:%d][%d] ID=%d, got %d",
					tc.start, tc.end, i, expectedID, mockProp.ID())
			}
		}
	}
}
