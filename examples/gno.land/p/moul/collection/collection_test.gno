package collection

import (
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/seqid"
)

type Person struct {
	Name     string
	Age      int
	Email    string
	Username string
	Tags     []string
}

// TestOperation represents a single operation in a test sequence
type TestOperation struct {
	op      string // "set" or "update"
	person  *Person
	id      uint64 // for updates
	wantID  uint64
	wantErr bool
}

// TestCase represents a complete test case with setup and operations
type TestCase struct {
	name       string
	setupIndex func(*Collection)
	operations []TestOperation
}

func TestBasicOperations(t *testing.T) {
	c := New()

	// Add indexes
	c.AddIndex("name", func(v interface{}) string {
		return v.(*Person).Name
	}, UniqueIndex)
	c.AddIndex("age", func(v interface{}) string {
		return strconv.Itoa(v.(*Person).Age)
	}, DefaultIndex)

	// Test basic Set and Get
	p1 := &Person{Name: "Alice", Age: 30, Email: "alice@test.com"}
	id1 := c.Set(p1)
	if id1 == 0 {
		t.Error("Failed to set first object")
	}

	// Get by ID
	obj1, gotId := c.Get(IDIndex, seqid.ID(id1).String())
	if gotId == 0 {
		t.Error("Failed to get object by ID")
	}
	if obj1.(*Person).Name != "Alice" {
		t.Error("Got wrong object")
	}
}

func TestUniqueConstraints(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v interface{}) string {
		return v.(*Person).Name
	}, UniqueIndex)
	c.AddIndex("email", func(v interface{}) string {
		return strings.ToLower(v.(*Person).Email)
	}, UniqueIndex|CaseInsensitiveIndex)
	c.AddIndex("age", func(v interface{}) string {
		return strconv.Itoa(v.(*Person).Age)
	}, DefaultIndex)

	// Test cases for unique constraints
	tests := []struct {
		name     string
		person   *Person
		wantID   bool
		existing uint64
	}{
		{
			name:   "First person",
			person: &Person{Name: "Alice", Email: "alice@test.com", Age: 30},
			wantID: true,
		},
		{
			name:   "Duplicate name",
			person: &Person{Name: "Alice", Email: "alice2@test.com", Age: 25},
			wantID: false,
		},
		{
			name:   "Duplicate email (case insensitive)",
			person: &Person{Name: "Bob", Email: "ALICE@test.com", Age: 35},
			wantID: false,
		},
		{
			name:   "Same age (non-unique index)",
			person: &Person{Name: "Bob", Email: "bob@test.com", Age: 30},
			wantID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := c.Set(tt.person)
			if (id != 0) != tt.wantID {
				t.Errorf("Set() got id = %v, want non-zero: %v", id, tt.wantID)
			}
		})
	}
}

func TestUpdates(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v interface{}) string {
		return v.(*Person).Name
	}, UniqueIndex)
	c.AddIndex("username", func(v interface{}) string {
		return strings.ToLower(v.(*Person).Username)
	}, UniqueIndex|CaseInsensitiveIndex)

	// Initial setup
	p1 := &Person{Name: "Alice", Username: "alice123"}
	p2 := &Person{Name: "Bob", Username: "bob456"}

	id1 := c.Set(p1)
	id2 := c.Set(p2)

	tests := []struct {
		name      string
		id        uint64
		newPerson *Person
		wantID    uint64
	}{
		{
			name:      "Update to non-conflicting values",
			id:        id1,
			newPerson: &Person{Name: "Alice2", Username: "alice1234"},
			wantID:    id1,
		},
		{
			name:      "Update to conflicting username",
			id:        id1,
			newPerson: &Person{Name: "Alice2", Username: "bob456"},
			wantID:    0,
		},
		{
			name:      "Update non-existent ID",
			id:        99999,
			newPerson: &Person{Name: "Test", Username: "test"},
			wantID:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID := c.Update(tt.id, tt.newPerson)
			if gotID != tt.wantID {
				t.Errorf("Update() got = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v interface{}) string {
		return v.(*Person).Name
	}, UniqueIndex)

	// Setup
	p1 := &Person{Name: "Alice"}
	id1 := c.Set(p1)

	tests := []struct {
		name    string
		id      uint64
		wantObj bool
	}{
		{
			name:    "Delete existing object",
			id:      id1,
			wantObj: false,
		},
		{
			name:    "Delete non-existent object",
			id:      99999,
			wantObj: false,
		},
		{
			name:    "Delete already deleted object",
			id:      id1,
			wantObj: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Delete(tt.id)
			_, gotId := c.Get(IDIndex, seqid.ID(tt.id).String())
			if (gotId != 0) != tt.wantObj {
				t.Errorf("After Delete(), Get() returned object: %v, want object: %v", gotId != 0, tt.wantObj)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v interface{}) string {
		return v.(*Person).Name
	}, UniqueIndex)

	tests := []struct {
		name      string
		operation func() uint64
		wantID    uint64
		wantPanic bool
	}{
		{
			name: "Set nil object",
			operation: func() uint64 {
				return c.Set(nil)
			},
			wantID:    0,
			wantPanic: false,
		},
		{
			name: "Set wrong type",
			operation: func() uint64 {
				return c.Set("not a person")
			},
			wantID:    0,
			wantPanic: false,
		},
		{
			name: "Update with nil",
			operation: func() uint64 {
				id := c.Set(&Person{Name: "Test"})
				return c.Update(id, nil)
			},
			wantID:    0,
			wantPanic: false,
		},
		{
			name: "Get with invalid index name",
			operation: func() uint64 {
				_, id := c.Get("invalid_index", "key")
				return id
			},
			wantID:    0,
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID uint64
			panicked := false

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				gotID = tt.operation()
			}()

			if panicked != tt.wantPanic {
				t.Errorf("Operation panicked = %v, want panic = %v", panicked, tt.wantPanic)
			}
			if !panicked && gotID != tt.wantID {
				t.Errorf("Operation returned %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestIndexTypes(t *testing.T) {
	c := New()

	// Test different types of indexes
	c.AddIndex("composite", func(v interface{}) string {
		p := v.(*Person)
		return p.Name + ":" + strconv.Itoa(p.Age)
	}, UniqueIndex)

	c.AddIndex("case_insensitive", func(v interface{}) string {
		return strings.ToLower(v.(*Person).Username)
	}, UniqueIndex|CaseInsensitiveIndex)

	// Test composite index
	p1 := &Person{Name: "Alice", Age: 30, Username: "Alice123"}
	id1 := c.Set(p1)
	if id1 == 0 {
		t.Error("Failed to set object with composite index")
	}

	// Test case-insensitive index
	p2 := &Person{Name: "Bob", Age: 25, Username: "alice123"}
	id2 := c.Set(p2)
	if id2 != 0 {
		t.Error("Case-insensitive index failed to prevent duplicate")
	}
}

func TestIndexOptions(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Collection) uint64
		wantID  bool
		wantErr bool
	}{
		{
			name: "Unique case-sensitive index",
			setup: func(c *Collection) uint64 {
				c.AddIndex("username", func(v interface{}) string {
					return v.(*Person).Username
				}, UniqueIndex)

				id1 := c.Set(&Person{Username: "Alice"})
				return c.Set(&Person{Username: "Alice"}) // Should fail
			},
			wantID: false,
		},
		{
			name: "Unique case-insensitive index",
			setup: func(c *Collection) uint64 {
				c.AddIndex("email", func(v interface{}) string {
					return v.(*Person).Email
				}, UniqueIndex|CaseInsensitiveIndex)

				id1 := c.Set(&Person{Email: "test@example.com"})
				return c.Set(&Person{Email: "TEST@EXAMPLE.COM"}) // Should fail
			},
			wantID: false,
		},
		{
			name: "Default index",
			setup: func(c *Collection) uint64 {
				c.AddIndex("age", func(v interface{}) string {
					return strconv.Itoa(v.(*Person).Age)
				}, DefaultIndex)

				// First person with age 30
				id1 := c.Set(&Person{Age: 30})
				if id1 == 0 {
					t.Error("Failed to set first person")
				}

				// Second person with same age should succeed
				return c.Set(&Person{Age: 30})
			},
			wantID: true,
		},
		{
			name: "Multiple options",
			setup: func(c *Collection) uint64 {
				c.AddIndex("name", func(v interface{}) string {
					return v.(*Person).Name
				}, UniqueIndex|CaseInsensitiveIndex|SparseIndex)

				id1 := c.Set(&Person{Name: "Alice"})
				return c.Set(&Person{Name: "ALICE"}) // Should fail
			},
			wantID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New() // Create new collection for each test
			id := tt.setup(c)
			if (id != 0) != tt.wantID {
				t.Errorf("got id = %v, want non-zero: %v", id, tt.wantID)
			}
		})
	}
}

func TestConcurrentOperations(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v interface{}) string {
		return v.(*Person).Name
	}, UniqueIndex)

	// Test concurrent reads and writes
	p1 := &Person{Name: "Alice"}
	id1 := c.Set(p1)

	// Get and Update simultaneously
	obj, gotId := c.Get(IDIndex, seqid.ID(id1).String())
	newID := c.Update(id1, &Person{Name: "Alice2"})

	if gotId == 0 || newID == 0 {
		t.Error("Concurrent operations failed")
	}
}

func TestSparseIndexBehavior(t *testing.T) {
	c := New()
	c.AddIndex("optional_field", func(v interface{}) string {
		return v.(*Person).Username
	}, SparseIndex)

	tests := []struct {
		name   string
		person *Person
		wantID bool
	}{
		{
			name:   "Empty optional field",
			person: &Person{Name: "Alice", Email: "alice@test.com"},
			wantID: true,
		},
		{
			name:   "Populated optional field",
			person: &Person{Name: "Bob", Email: "bob@test.com", Username: "bobby"},
			wantID: true,
		},
		{
			name:   "Multiple empty fields",
			person: &Person{Name: "Charlie"},
			wantID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := c.Set(tt.person)
			if (id != 0) != tt.wantID {
				t.Errorf("Set() got id = %v, want non-zero: %v", id, tt.wantID)
			}
		})
	}
}

func TestIndexKeyGeneration(t *testing.T) {
	c := New()
	c.AddIndex("composite", func(v interface{}) string {
		p := v.(*Person)
		return p.Name + ":" + strconv.Itoa(p.Age)
	}, UniqueIndex)

	tests := []struct {
		name   string
		person *Person
		wantID bool
	}{
		{
			name:   "Valid composite key",
			person: &Person{Name: "Alice", Age: 30},
			wantID: true,
		},
		{
			name:   "Duplicate composite key",
			person: &Person{Name: "Alice", Age: 30},
			wantID: false,
		},
		{
			name:   "Different composite key",
			person: &Person{Name: "Alice", Age: 31},
			wantID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := c.Set(tt.person)
			if (id != 0) != tt.wantID {
				t.Errorf("Set() got id = %v, want non-zero: %v", id, tt.wantID)
			}
		})
	}
}

func TestGetIndex(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v interface{}) string {
		return v.(*Person).Name
	}, UniqueIndex)

	tests := []struct {
		name      string
		indexName string
		wantNil   bool
	}{
		{
			name:      "Get existing index",
			indexName: "name",
			wantNil:   false,
		},
		{
			name:      "Get _id index",
			indexName: IDIndex,
			wantNil:   false,
		},
		{
			name:      "Get non-existent index",
			indexName: "invalid",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := c.GetIndex(tt.indexName)
			if (tree == nil) != tt.wantNil {
				t.Errorf("GetIndex() got nil = %v, want nil = %v", tree == nil, tt.wantNil)
			}
		})
	}
}

func TestAddIndexPanic(t *testing.T) {
	c := New()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when adding _id index")
		}
	}()

	c.AddIndex(IDIndex, func(v interface{}) string {
		return ""
	}, DefaultIndex)
}

func TestCaseInsensitiveGet(t *testing.T) {
	c := New()
	c.AddIndex("email", func(v interface{}) string {
		return v.(*Person).Email
	}, UniqueIndex|CaseInsensitiveIndex)

	p := &Person{Email: "Test@Example.com"}
	id := c.Set(p)

	tests := []struct {
		name    string
		key     string
		wantObj bool
	}{
		{
			name:    "Exact match",
			key:     "Test@Example.com",
			wantObj: true,
		},
		{
			name:    "Different case",
			key:     "test@example.COM",
			wantObj: true,
		},
		{
			name:    "Non-existent",
			key:     "other@example.com",
			wantObj: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, gotId := c.Get("email", tt.key)
			if (obj != nil) != tt.wantObj {
				t.Errorf("Get() got object = %v, want object = %v", obj != nil, tt.wantObj)
			}
			if tt.wantObj && gotId != id {
				t.Errorf("Get() got id = %v, want id = %v", gotId, id)
			}
		})
	}
}

func TestGetInvalidID(t *testing.T) {
	c := New()
	obj, id := c.Get(IDIndex, "not-a-valid-id")
	if obj != nil || id != 0 {
		t.Errorf("Get() with invalid ID format got (obj=%v, id=%v), want (nil, 0)", obj, id)
	}
}

func TestMultiValueIndex(t *testing.T) {
	c := New()

	// Add a multi-value index for "tags"
	c.AddIndex("tags", func(v interface{}) []string {
		return v.(*Person).Tags
	}, MultiValueIndex)

	// Create a person with multiple tags
	p := &Person{Name: "AliceWithTags", Tags: []string{"golang", "engineer", "travel"}}
	id := c.Set(p)

	if id == 0 {
		t.Fatal("Failed to set object with multi-value index")
	}

	// Check each key in the "tags" index to confirm insertion
	for _, tag := range p.Tags {
		val, ok := c.GetIndex("tags").Get(tag)
		if !ok {
			t.Errorf("Expected to find key '%s' in multi-value index, but did not", tag)
			continue
		}
		storedIDs, ok := val.([]string)
		if !ok {
			t.Errorf("Expected a []string for multi-value index under key '%s', but got %T", tag, val)
			continue
		}

		found := false
		for _, storedID := range storedIDs {
			if storedID == seqid.ID(id).String() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find ID %s in multi-value index for key '%s', but it was missing", seqid.ID(id).String(), tag)
		}
	}
}

func TestMultiValueDelete(t *testing.T) {
	c := New()

	// Add a multi-value index for "tags"
	c.AddIndex("tags", func(v interface{}) []string {
		return v.(*Person).Tags
	}, MultiValueIndex)

	// Create two people with overlapping tags.
	pAlice := &Person{Name: "Alice", Tags: []string{"a", "dev", "gopher"}}
	pBob := &Person{Name: "Bob", Tags: []string{"a", "cooking"}}

	idAlice := c.Set(pAlice)
	if idAlice == 0 {
		t.Fatal("Failed to set Alice with multi-value index")
	}
	idBob := c.Set(pBob)
	if idBob == 0 {
		t.Fatal("Failed to set Bob with multi-value index")
	}

	// Confirm that both Alice and Bob's IDs appear under the tag "a".
	val, ok := c.GetIndex("tags").Get("a")
	if !ok {
		t.Error("Expected to find key 'a' for multi-value index but did not.")
	} else {
		storedIDs, ok := val.([]string)
		if !ok {
			t.Errorf("Expected a []string for multi-value key 'a', got %T", val)
		} else {
			// Check that both IDs are present
			var aliceFound, bobFound bool
			for _, sid := range storedIDs {
				switch sid {
				case seqid.ID(idAlice).String():
					aliceFound = true
				case seqid.ID(idBob).String():
					bobFound = true
				}
			}
			if !aliceFound || !bobFound {
				t.Errorf("Expected tag 'a' to contain both Alice (%v) and Bob (%v). Found: %v", idAlice, idBob, storedIDs)
			}
		}
	}

	// Now delete Alice
	c.Delete(idAlice)

	// After deleting Alice, "a" should remain for Bob only
	valAfterDelete, ok := c.GetIndex("tags").Get("a")
	if !ok {
		t.Error("Expected to still find key 'a' pointing to Bob after deleting Alice, but key was not found.")
	} else {
		storedIDs, ok := valAfterDelete.([]string)
		if !ok {
			t.Errorf("Expected a []string for multi-value key 'a', got %T", valAfterDelete)
		} else {
			// Check that only Bob remains
			if len(storedIDs) != 1 || storedIDs[0] != seqid.ID(idBob).String() {
				t.Errorf("Expected 'a' to only have Bob's ID %v after deleting Alice, got %v", seqid.ID(idBob).String(), storedIDs)
			}
		}
	}

	// Double-check that Alice is not retrievable by her ID
	_, gotId := c.Get(IDIndex, seqid.ID(idAlice).String())
	if gotId != 0 {
		t.Errorf("Expected Alice with ID %v to be deleted but she is still present", idAlice)
	}
}
