package collection

import "gno.land/p/nt/ufmt/v0"

// Entry represents a single object in the collection with its ID
type Entry struct {
	ID  string
	Obj any
}

// String returns a string representation of the Entry
func (e *Entry) String() string {
	if e == nil {
		return "<nil>"
	}
	return ufmt.Sprintf("Entry{ID: %s, Obj: %v}", e.ID, e.Obj)
}

// EntryIterator provides iteration over collection entries
type EntryIterator struct {
	collection *Collection
	indexName  string
	key        string
	currentID  string
	currentObj any
	err        error
	closed     bool

	// For multi-value cases
	ids        []string
	currentIdx int
}

func (ei *EntryIterator) Close() error {
	ei.closed = true
	ei.currentID = ""
	ei.currentObj = nil
	ei.ids = nil
	return nil
}

func (ei *EntryIterator) Next() bool {
	if ei == nil || ei.closed || ei.err != nil {
		return false
	}

	// Handle ID index specially
	if ei.indexName == IDIndex {
		if ei.currentID != "" { // We've already returned the single value
			return false
		}
		obj, exists := ei.collection.indexes[IDIndex].tree.Get(ei.key)
		if !exists {
			return false
		}
		ei.currentID = ei.key
		ei.currentObj = obj
		return true
	}

	// Get the index
	idx, exists := ei.collection.indexes[ei.indexName]
	if !exists {
		return false
	}

	// Initialize ids slice if needed
	if ei.ids == nil {
		idData, exists := idx.tree.Get(ei.key)
		if !exists {
			return false
		}

		switch stored := idData.(type) {
		case []string:
			ei.ids = stored
			ei.currentIdx = -1
		case string:
			ei.ids = []string{stored}
			ei.currentIdx = -1
		default:
			return false
		}
	}

	// Move to next ID
	ei.currentIdx++
	if ei.currentIdx >= len(ei.ids) {
		return false
	}

	// Fetch the actual object
	ei.currentID = ei.ids[ei.currentIdx]
	obj, exists := ei.collection.indexes[IDIndex].tree.Get(ei.currentID)
	if !exists {
		// Skip invalid entries
		return ei.Next()
	}
	ei.currentObj = obj
	return true
}

func (ei *EntryIterator) Error() error {
	return ei.err
}

func (ei *EntryIterator) Value() *Entry {
	if ei == nil || ei.closed || ei.currentID == "" {
		return nil
	}
	return &Entry{
		ID:  ei.currentID,
		Obj: ei.currentObj,
	}
}

func (ei *EntryIterator) Empty() bool {
	if ei == nil || ei.closed || ei.err != nil {
		return true
	}

	// Handle ID index specially
	if ei.indexName == IDIndex {
		_, exists := ei.collection.indexes[IDIndex].tree.Get(ei.key)
		return !exists
	}

	// Get the index
	idx, exists := ei.collection.indexes[ei.indexName]
	if !exists {
		return true
	}

	// Check if key exists in index
	idData, exists := idx.tree.Get(ei.key)
	if !exists {
		return true
	}

	// Check if there are any valid IDs
	switch stored := idData.(type) {
	case []string:
		return len(stored) == 0
	case string:
		return stored == ""
	default:
		return true
	}
}
