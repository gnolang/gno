package memdb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterator(t *testing.T) {
	tests := []struct {
		name       string
		setupDB    func(*MemDB)
		start      []byte
		end        []byte
		wantKeys   []string
		wantValues []string
	}{
		{
			name:    "empty database",
			setupDB: func(db *MemDB) {},
			start:   nil,
			end:     nil,
		},
		{
			name: "single key",
			setupDB: func(db *MemDB) {
				db.Set([]byte("key1"), []byte("value1"))
			},
			start:      nil,
			end:        nil,
			wantKeys:   []string{"key1"},
			wantValues: []string{"value1"},
		},
		{
			name: "multiple keys in range",
			setupDB: func(db *MemDB) {
				db.Set([]byte("key1"), []byte("value1"))
				db.Set([]byte("key2"), []byte("value2"))
				db.Set([]byte("key3"), []byte("value3"))
			},
			start:      []byte("key1"),
			end:        []byte("key3"),
			wantKeys:   []string{"key1", "key2"},
			wantValues: []string{"value1", "value2"},
		},
		{
			name: "prefix iteration",
			setupDB: func(db *MemDB) {
				db.Set([]byte("prefix1_a"), []byte("value1"))
				db.Set([]byte("prefix1_b"), []byte("value2"))
				db.Set([]byte("prefix2_a"), []byte("value3"))
			},
			start:      []byte("prefix1_"),
			end:        []byte("prefix1_\xff"),
			wantKeys:   []string{"prefix1_a", "prefix1_b"},
			wantValues: []string{"value1", "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMemDB()
			tt.setupDB(db)

			iter := db.Iterator(tt.start, tt.end)
			defer iter.Close()

			var gotKeys []string
			var gotValues []string

			for ; iter.Valid(); iter.Next() {
				gotKeys = append(gotKeys, string(iter.Key()))
				gotValues = append(gotValues, string(iter.Value()))
			}

			assert.Equal(t, tt.wantKeys, gotKeys)
			assert.Equal(t, tt.wantValues, gotValues)
		})
	}
}

func TestIterator_Domain(t *testing.T) {
	db := NewMemDB()
	start := []byte("0")
	end := []byte("9")

	iter := db.Iterator(start, end)
	gotStart, gotEnd := iter.Domain()

	if !bytes.Equal(gotStart, start) {
		t.Errorf("Domain start: got %v, want %v", gotStart, start)
	}
	if !bytes.Equal(gotEnd, end) {
		t.Errorf("Domain end: got %v, want %v", gotEnd, end)
	}
}

func TestIterator_InvalidOperation(t *testing.T) {
	db := NewMemDB()
	db.Set([]byte("key"), []byte("value"))

	iter := db.Iterator(nil, nil)
	iter.Next() // Move past the only key

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on Next() after iterator becomes invalid")
		}
	}()

	iter.Next() // Should panic
}

func TestReverseIterator(t *testing.T) {
	tests := []struct {
		name       string
		setupDB    func(*MemDB)
		start      []byte
		end        []byte
		wantKeys   []string
		wantValues []string
	}{
		{
			name: "reverse multiple keys",
			setupDB: func(db *MemDB) {
				db.Set([]byte("key1"), []byte("value1"))
				db.Set([]byte("key2"), []byte("value2"))
				db.Set([]byte("key3"), []byte("value3"))
			},
			start:      []byte("key1"),
			end:        []byte("key3"),
			wantKeys:   []string{"key2", "key1"},
			wantValues: []string{"value2", "value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewMemDB()
			tt.setupDB(db)

			iter := db.ReverseIterator(tt.start, tt.end)
			defer iter.Close()

			var gotKeys []string
			var gotValues []string

			for ; iter.Valid(); iter.Next() {
				gotKeys = append(gotKeys, string(iter.Key()))
				gotValues = append(gotValues, string(iter.Value()))
			}

			assert.Equal(t, tt.wantKeys, gotKeys)
			assert.Equal(t, tt.wantValues, gotValues)
		})
	}
}
