package collection

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
)

type Person struct {
	Name     string
	Age      int
	Email    string
	Username string
	Tags     []string
}

func (p Person) String() string {
	return ufmt.Sprintf("name=%s age=%d email=%s username=%s tags=%s",
		p.Name, p.Age, p.Email, p.Username, strings.Join(p.Tags, ","))
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
	c.AddIndex("name", func(v any) string {
		return v.(*Person).Name
	}, UniqueIndex)
	c.AddIndex("age", func(v any) string {
		return strconv.Itoa(v.(*Person).Age)
	}, DefaultIndex)

	// Test basic Set and Get
	p1 := &Person{Name: "Alice", Age: 30, Email: "alice@test.com"}
	id1 := c.Set(p1)
	if id1 == 0 {
		t.Error("Failed to set first object")
	}

	// Get by ID
	iter := c.Get(IDIndex, seqid.ID(id1).String())
	if !iter.Next() {
		t.Error("Failed to get object by ID")
	}
	entry := iter.Value()
	if entry.Obj.(*Person).Name != "Alice" {
		t.Error("Got wrong object")
	}
}

func TestUniqueConstraints(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*Collection) uint64
		wantID bool
	}{
		{
			name: "First person",
			setup: func(c *Collection) uint64 {
				return c.Set(&Person{Name: "Alice"})
			},
			wantID: true,
		},
		{
			name: "Duplicate name",
			setup: func(c *Collection) uint64 {
				c.Set(&Person{Name: "Alice"})
				return c.Set(&Person{Name: "Alice"})
			},
			wantID: false,
		},
		{
			name: "Same age (non-unique index)",
			setup: func(c *Collection) uint64 {
				c.AddIndex("age", func(v any) string {
					return strconv.Itoa(v.(*Person).Age)
				}, DefaultIndex)
				c.Set(&Person{Name: "Alice", Age: 30})
				return c.Set(&Person{Name: "Bob", Age: 30})
			},
			wantID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.AddIndex("name", func(v any) string {
				return v.(*Person).Name
			}, UniqueIndex)

			id := tt.setup(c)
			if (id != 0) != tt.wantID {
				t.Errorf("Set() got id = %v, want non-zero: %v", id, tt.wantID)
			}
		})
	}
}

func TestUpdates(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v any) string {
		return v.(*Person).Name
	}, UniqueIndex)
	c.AddIndex("username", func(v any) string {
		return strings.ToLower(v.(*Person).Username)
	}, UniqueIndex|CaseInsensitiveIndex)

	// Initial setup
	p1 := &Person{Name: "Alice", Username: "alice123"}
	p2 := &Person{Name: "Bob", Username: "bob456"}

	id1 := c.Set(p1)
	c.Set(p2)

	tests := []struct {
		name      string
		id        uint64
		newPerson *Person
		wantRet   bool
	}{
		{
			name:      "Update to non-conflicting values",
			id:        id1,
			newPerson: &Person{Name: "Alice2", Username: "alice1234"},
			wantRet:   true,
		},
		{
			name:      "Update to conflicting username",
			id:        id1,
			newPerson: &Person{Name: "Alice2", Username: "bob456"},
			wantRet:   false,
		},
		{
			name:      "Update non-existent ID",
			id:        99999,
			newPerson: &Person{Name: "Test", Username: "test"},
			wantRet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID := c.Update(tt.id, tt.newPerson)
			if gotID != tt.wantRet {
				t.Errorf("Update() got = %v, want %v", gotID, tt.wantRet)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v any) string {
		return v.(*Person).Name
	}, UniqueIndex)

	p1 := &Person{Name: "Alice"}
	id1 := c.Set(p1)

	tests := []struct {
		name    string
		id      uint64
		wantRet bool
	}{
		{
			name:    "Delete existing object",
			id:      id1,
			wantRet: true,
		},
		{
			name:    "Delete non-existent object",
			id:      99999,
			wantRet: false,
		},
		{
			name:    "Delete already deleted object",
			id:      id1,
			wantRet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID := c.Delete(tt.id)
			if gotID != tt.wantRet {
				t.Errorf("Delete() got = %v, want %v", gotID, tt.wantRet)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v any) string {
		return v.(*Person).Name
	}, UniqueIndex)

	tests := []struct {
		name      string
		operation func() bool
		wantPanic bool
	}{
		{
			name: "Set nil object",
			operation: func() bool {
				return c.Set(nil) != 0
			},
			wantPanic: false,
		},
		{
			name: "Set wrong type",
			operation: func() bool {
				return c.Set("not a person") != 0
			},
			wantPanic: true,
		},
		{
			name: "Update with nil",
			operation: func() bool {
				id := c.Set(&Person{Name: "Test"})
				return c.Update(id, nil)
			},
			wantPanic: false,
		},
		{
			name: "Get with invalid index name",
			operation: func() bool {
				iter := c.Get("invalid_index", "key")
				if iter.Empty() {
					return false
				}
				entry := iter.Value()
				if entry == nil {
					return false
				}
				_, err := seqid.FromString(entry.ID)
				if err != nil {
					return false
				}
				return true
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			panicked := false

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				got = tt.operation()
			}()

			if panicked != tt.wantPanic {
				t.Errorf("Operation panicked = %v, want panic = %v", panicked, tt.wantPanic)
			}
			if !panicked && got != false {
				t.Errorf("Operation returned %v, want 0", got)
			}
		})
	}
}

func TestIndexTypes(t *testing.T) {
	c := New()

	// Test different types of indexes
	c.AddIndex("composite", func(v any) string {
		p := v.(*Person)
		return p.Name + ":" + strconv.Itoa(p.Age)
	}, UniqueIndex)

	c.AddIndex("case_insensitive", func(v any) string {
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
				c.AddIndex("username", func(v any) string {
					return v.(*Person).Username
				}, UniqueIndex)

				c.Set(&Person{Username: "Alice"})
				return c.Set(&Person{Username: "Alice"}) // Should fail
			},
			wantID: false,
		},
		{
			name: "Unique case-insensitive index",
			setup: func(c *Collection) uint64 {
				c.AddIndex("email", func(v any) string {
					return v.(*Person).Email
				}, UniqueIndex|CaseInsensitiveIndex)

				c.Set(&Person{Email: "test@example.com"})
				return c.Set(&Person{Email: "TEST@EXAMPLE.COM"}) // Should fail
			},
			wantID: false,
		},
		{
			name: "Default index",
			setup: func(c *Collection) uint64 {
				c.AddIndex("age", func(v any) string {
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
				c.AddIndex("name", func(v any) string {
					return v.(*Person).Name
				}, UniqueIndex|CaseInsensitiveIndex|SparseIndex)

				c.Set(&Person{Name: "Alice"})
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
	c.AddIndex("name", func(v any) string {
		return v.(*Person).Name
	}, UniqueIndex)

	p1 := &Person{Name: "Alice"}
	id1 := c.Set(p1)
	iter := c.Get("_id", seqid.ID(id1).String())
	success := c.Update(id1, &Person{Name: "Alice2"})

	if iter.Empty() || !success {
		t.Error("Concurrent operations failed")
	}
}

func TestSparseIndexBehavior(t *testing.T) {
	c := New()
	c.AddIndex("optional_field", func(v any) string {
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
	c.AddIndex("composite", func(v any) string {
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
	c.AddIndex("name", func(v any) string {
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

	c.AddIndex(IDIndex, func(v any) string {
		return ""
	}, DefaultIndex)
}

func TestCaseInsensitiveOperations(t *testing.T) {
	c := New()
	c.AddIndex("email", func(v any) string {
		return v.(*Person).Email
	}, UniqueIndex|CaseInsensitiveIndex)

	p := &Person{Email: "Test@Example.com"}
	id := c.Set(p)

	tests := []struct {
		name      string
		key       string
		wantObj   bool
		operation string // "get" or "getAll"
		wantCount int
	}{
		{"Get exact match", "Test@Example.com", true, "get", 1},
		{"Get different case", "test@example.COM", true, "get", 1},
		{"Get non-existent", "other@example.com", false, "get", 0},
		{"GetAll exact match", "Test@Example.com", true, "getAll", 1},
		{"GetAll different case", "test@example.COM", true, "getAll", 1},
		{"GetAll non-existent", "other@example.com", false, "getAll", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.operation == "get" {
				iter := c.Get("email", tt.key)
				if iter.Empty() {
					if tt.wantObj {
						t.Error("Expected iterator to not be empty")
					}
					return
				}
				hasValue := iter.Next()
				if hasValue != tt.wantObj {
					t.Errorf("Get() got object = %v, want object = %v", hasValue, tt.wantObj)
				}
				if hasValue {
					entry := iter.Value()
					if entry.ID != seqid.ID(id).String() {
						t.Errorf("Get() got id = %v, want id = %v", entry.ID, seqid.ID(id).String())
					}
				}
			} else {
				entries := c.GetAll("email", tt.key)
				if len(entries) != tt.wantCount {
					t.Errorf("GetAll() returned %d entries, want %d", len(entries), tt.wantCount)
				}
				if tt.wantCount > 0 {
					entry := entries[0]
					if entry.ID != seqid.ID(id).String() {
						t.Errorf("GetAll() returned ID %s, want %s", entry.ID, seqid.ID(id).String())
					}
				}
			}
		})
	}
}

func TestGetInvalidID(t *testing.T) {
	c := New()
	iter := c.Get(IDIndex, "not-a-valid-id")
	if !iter.Empty() {
		t.Errorf("Get() with invalid ID format got an entry, want nil")
	}
}

func TestGetAll(t *testing.T) {
	c := New()
	c.AddIndex("tags", func(v any) []string {
		return v.(*Person).Tags
	}, DefaultIndex)
	c.AddIndex("age", func(v any) string {
		return strconv.Itoa(v.(*Person).Age)
	}, DefaultIndex)
	c.AddIndex("name", func(v any) string {
		return v.(*Person).Name
	}, UniqueIndex)

	// Create test data
	people := []*Person{
		{Name: "Alice", Age: 30, Tags: []string{"dev", "go"}},
		{Name: "Bob", Age: 30, Tags: []string{"dev", "python"}},
		{Name: "Charlie", Age: 25, Tags: []string{"dev", "rust"}},
	}

	ids := make([]uint64, len(people))
	for i, p := range people {
		ids[i] = c.Set(p)
		if ids[i] == 0 {
			t.Fatalf("Failed to set person %s", p.Name)
		}
	}

	tests := []struct {
		name      string
		indexName string
		key       string
		wantCount int
	}{
		{
			name:      "Multi-value index with multiple matches",
			indexName: "tags",
			key:       "dev",
			wantCount: 3,
		},
		{
			name:      "Multi-value index with single match",
			indexName: "tags",
			key:       "go",
			wantCount: 1,
		},
		{
			name:      "Multi-value index with no matches",
			indexName: "tags",
			key:       "java",
			wantCount: 0,
		},
		{
			name:      "Single-value non-unique index with multiple matches",
			indexName: "age",
			key:       "30",
			wantCount: 2,
		},
		{
			name:      "Single-value unique index",
			indexName: "name",
			key:       "Alice",
			wantCount: 1,
		},
		{
			name:      "Non-existent index",
			indexName: "invalid",
			key:       "value",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := c.Get(tt.indexName, tt.key)
			count := 0
			for iter.Next() {
				entry := iter.Value()
				if entry.ID == "" {
					t.Error("Got entry with empty ID")
				}
				if entry.Obj == nil {
					t.Error("Got entry with nil Obj")
				}
				count++
			}
			if count != tt.wantCount {
				t.Errorf("Got %d entries, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestIndexOperations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Collection) (uint64, error)
		verify  func(*Collection, uint64) error
		wantErr bool
	}{
		{
			name: "Basic set and get",
			setup: func(c *Collection) (uint64, error) {
				c.AddIndex("name", func(v any) string {
					return v.(*Person).Name
				}, UniqueIndex)
				return c.Set(&Person{Name: "Alice", Age: 30}), nil
			},
			verify: func(c *Collection, id uint64) error {
				iter := c.Get(IDIndex, seqid.ID(id).String())
				if !iter.Next() {
					return errors.New("failed to get object by ID")
				}
				entry := iter.Value()
				if entry.Obj.(*Person).Name != "Alice" {
					return errors.New("got wrong object")
				}
				return nil
			},
		},
		{
			name: "Composite index",
			setup: func(c *Collection) (uint64, error) {
				c.AddIndex("composite", func(v any) string {
					p := v.(*Person)
					return p.Name + ":" + strconv.Itoa(p.Age)
				}, UniqueIndex)
				return c.Set(&Person{Name: "Alice", Age: 30}), nil
			},
			verify: func(c *Collection, id uint64) error {
				iter := c.Get("composite", "Alice:30")
				if !iter.Next() {
					return errors.New("failed to get object by composite index")
				}
				return nil
			},
		},
		// Add more test cases combining unique scenarios from original tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			id, err := tt.setup(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("setup error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if err := tt.verify(c, id); err != nil {
					t.Errorf("verification failed: %v", err)
				}
			}
		})
	}
}

func TestMultiValueIndexes(t *testing.T) {
	c := New()
	c.AddIndex("tags", func(v any) []string {
		return v.(*Person).Tags
	}, DefaultIndex)

	tests := []struct {
		name      string
		setup     []*Person
		searchTag string
		wantCount int
	}{
		{
			name: "Multiple tags, multiple matches",
			setup: []*Person{
				{Name: "Alice", Tags: []string{"dev", "go"}},
				{Name: "Bob", Tags: []string{"dev", "python"}},
				{Name: "Charlie", Tags: []string{"dev", "rust"}},
			},
			searchTag: "dev",
			wantCount: 3,
		},
		{
			name: "Single tag match",
			setup: []*Person{
				{Name: "Alice", Tags: []string{"dev", "go"}},
				{Name: "Bob", Tags: []string{"dev", "python"}},
			},
			searchTag: "go",
			wantCount: 1,
		},
		{
			name: "No matches",
			setup: []*Person{
				{Name: "Alice", Tags: []string{"dev", "go"}},
				{Name: "Bob", Tags: []string{"dev", "python"}},
			},
			searchTag: "java",
			wantCount: 0,
		},
		{
			name: "Empty tags",
			setup: []*Person{
				{Name: "Alice", Tags: []string{}},
				{Name: "Bob", Tags: nil},
			},
			searchTag: "dev",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.AddIndex("tags", func(v any) []string {
				return v.(*Person).Tags
			}, DefaultIndex)

			// Setup test data
			for _, p := range tt.setup {
				if id := c.Set(p); id == 0 {
					t.Fatalf("Failed to set person %s", p.Name)
				}
			}

			// Test Get operation
			iter := c.Get("tags", tt.searchTag)
			count := 0
			for iter.Next() {
				count++
			}
			if count != tt.wantCount {
				t.Errorf("Get() got %d matches, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestGetOperations(t *testing.T) {
	c := New()
	c.AddIndex("name", func(v any) string {
		return v.(*Person).Name
	}, UniqueIndex)
	c.AddIndex("age", func(v any) string {
		return strconv.Itoa(v.(*Person).Age)
	}, DefaultIndex)

	// Setup test data
	testPeople := []*Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 30},
		{Name: "Charlie", Age: 25},
	}

	ids := make([]uint64, len(testPeople))
	for i, p := range testPeople {
		ids[i] = c.Set(p)
		if ids[i] == 0 {
			t.Fatalf("Failed to set person %s", p.Name)
		}
	}

	tests := []struct {
		name      string
		indexName string
		key       string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Get by ID",
			indexName: IDIndex,
			key:       seqid.ID(ids[0]).String(),
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "Get by unique index",
			indexName: "name",
			key:       "Alice",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "Get by non-unique index",
			indexName: "age",
			key:       "30",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "Get with invalid index",
			indexName: "invalid_index",
			key:       "value",
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "Get with invalid ID format",
			indexName: IDIndex,
			key:       "not-a-valid-id",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := c.Get(tt.indexName, tt.key)
			if iter.Empty() {
				if !tt.wantErr {
					t.Errorf("Get() returned empty iterator, wanted %d results", tt.wantCount)
				}
				return
			}

			count := 0
			for iter.Next() {
				entry := iter.Value()
				if entry.ID == "" {
					t.Error("Got entry with empty ID")
				}
				if entry.Obj == nil {
					t.Error("Got entry with nil Obj")
				}
				count++
			}

			if count != tt.wantCount {
				t.Errorf("Get() returned %d results, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestEntryString(t *testing.T) {
	tests := []struct {
		name     string
		entry    *Entry
		expected string
	}{
		{
			name:     "Nil entry",
			entry:    nil,
			expected: "<nil>",
		},
		{
			name: "Person entry",
			entry: &Entry{
				ID:  "123",
				Obj: &Person{Name: "Alice", Age: 30},
			},
			expected: `Entry{ID: 123, Obj: name=Alice age=30 email= username= tags=}`,
		},
		{
			name: "Entry with nil object",
			entry: &Entry{
				ID:  "456",
				Obj: nil,
			},
			expected: `Entry{ID: 456, Obj: <nil>}`,
		},
		{
			name: "Entry with complete person",
			entry: &Entry{
				ID: "789",
				Obj: &Person{
					Name:     "Bob",
					Age:      25,
					Email:    "bob@example.com",
					Username: "bobby",
					Tags:     []string{"dev", "go"},
				},
			},
			expected: `Entry{ID: 789, Obj: name=Bob age=25 email=bob@example.com username=bobby tags=dev,go}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.String()
			if got != tt.expected {
				t.Errorf("Entry.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
