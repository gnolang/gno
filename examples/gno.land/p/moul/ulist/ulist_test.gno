package ulist

import (
	"testing"
)

func TestNew(t *testing.T) {
	l := New()
	if l.Len() != 0 {
		t.Errorf("New list should be empty, got length %d", l.Len())
	}
}

func TestNewFromSlice(t *testing.T) {
	values := []interface{}{1, 2, 3, 4, 5}
	l := NewFromSlice(values)

	if l.Len() != len(values) {
		t.Errorf("Expected length %d, got %d", len(values), l.Len())
	}

	for i, want := range values {
		got, err := l.Get(i)
		if err != nil {
			t.Errorf("Unexpected error at index %d: %v", i, err)
		}
		if got != want {
			t.Errorf("At index %d: want %v, got %v", i, want, got)
		}
	}
}

func TestList_BasicOperations(t *testing.T) {
	l := New()

	// Test Append
	l.Append(1)
	l.Append(2)
	l.Append(3)

	if l.Len() != 3 {
		t.Errorf("Expected length 3, got %d", l.Len())
	}

	// Test Get
	tests := []struct {
		index   int
		want    interface{}
		wantErr error
	}{
		{0, 1, nil},
		{1, 2, nil},
		{2, 3, nil},
		{-1, nil, ErrIndexOutOfRange},
		{3, nil, ErrIndexOutOfRange},
	}

	for _, tt := range tests {
		got, err := l.Get(tt.index)
		if err != tt.wantErr {
			t.Errorf("Get(%d) error = %v, want %v", tt.index, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("Get(%d) = %v, want %v", tt.index, got, tt.want)
		}
	}

	// Test Set
	err := l.Set(1, 42)
	if err != nil {
		t.Errorf("Unexpected error on Set: %v", err)
	}

	got, err := l.Get(1)
	if err != nil || got != 42 {
		t.Errorf("After Set(1, 42), Get(1) = %v, want 42", got)
	}
}

func TestList_Insert(t *testing.T) {
	l := New()

	// Test inserting at various positions
	tests := []struct {
		index   int
		value   interface{}
		wantErr error
	}{
		{0, 1, nil},                 // Insert at beginning
		{1, 2, nil},                 // Insert at end
		{1, 3, nil},                 // Insert in middle
		{-1, 0, ErrIndexOutOfRange}, // Invalid index
		{4, 0, ErrIndexOutOfRange},  // Invalid index
	}

	for _, tt := range tests {
		err := l.Insert(tt.index, tt.value)
		if err != tt.wantErr {
			t.Errorf("Insert(%d, %v) error = %v, want %v", tt.index, tt.value, err, tt.wantErr)
		}
	}
}

func TestList_Delete(t *testing.T) {
	l := NewFromSlice([]interface{}{1, 2, 3, 4, 5})

	// Test deleting at various positions
	tests := []struct {
		index   int
		want    interface{}
		wantErr error
	}{
		{0, 1, nil},                   // Delete from beginning
		{3, 5, nil},                   // Delete from end
		{1, 3, nil},                   // Delete from middle
		{-1, nil, ErrIndexOutOfRange}, // Invalid index
		{2, nil, ErrIndexOutOfRange},  // Invalid index (after deletions)
	}

	for i, tt := range tests {
		got, err := l.Delete(tt.index)
		if err != tt.wantErr {
			t.Errorf("test %d: Delete(%d) error = %v, want %v", i, tt.index, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("test %d: Delete(%d) = %v, want %v", i, tt.index, got, tt.want)
		}
	}
}

func TestList_Range(t *testing.T) {
	l := NewFromSlice([]interface{}{1, 2, 3, 4, 5})

	tests := []struct {
		start   int
		end     int
		want    []interface{}
		wantErr error
	}{
		{0, 5, []interface{}{1, 2, 3, 4, 5}, nil},
		{1, 4, []interface{}{2, 3, 4}, nil},
		{0, 0, []interface{}{}, nil},
		{-1, 5, nil, ErrInvalidArgument},
		{0, 6, nil, ErrInvalidArgument},
		{3, 2, nil, ErrInvalidArgument},
	}

	for i, tt := range tests {
		got, err := l.Range(tt.start, tt.end)
		if err != tt.wantErr {
			t.Errorf("test %d: Range(%d, %d) error = %v, want %v", i, tt.start, tt.end, err, tt.wantErr)
			continue
		}
		if err == nil {
			if len(got) != len(tt.want) {
				t.Errorf("test %d: Range(%d, %d) len = %d, want %d", i, tt.start, tt.end, len(got), len(tt.want))
				continue
			}
			for j := range got {
				if got[j] != tt.want[j] {
					t.Errorf("test %d: Range(%d, %d)[%d] = %v, want %v", i, tt.start, tt.end, j, got[j], tt.want[j])
				}
			}
		}
	}
}

func TestList_Iterator(t *testing.T) {
	values := []interface{}{1, 2, 3, 4, 5}
	l := NewFromSlice(values)
	it := l.NewIterator()

	index := 0
	for {
		value, ok := it.Next()
		if !ok {
			break
		}
		if value != values[index] {
			t.Errorf("Iterator value at index %d = %v, want %v", index, value, values[index])
		}
		index++
	}

	if index != len(values) {
		t.Errorf("Iterator traversed %d elements, want %d", index, len(values))
	}
}

func TestList_Clear(t *testing.T) {
	l := NewFromSlice([]interface{}{1, 2, 3})
	l.Clear()

	if l.Len() != 0 {
		t.Errorf("After Clear(), Len() = %d, want 0", l.Len())
	}

	_, err := l.Get(0)
	if err != ErrIndexOutOfRange {
		t.Errorf("After Clear(), Get(0) error = %v, want %v", err, ErrIndexOutOfRange)
	}
}

// Benchmark tests
func BenchmarkList_Append(b *testing.B) {
	l := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Append(i)
	}
}

func BenchmarkList_Get(b *testing.B) {
	l := New()
	for i := 0; i < 1000; i++ {
		l.Append(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = l.Get(i % 1000)
	}
}

func BenchmarkList_Insert(b *testing.B) {
	l := New()
	for i := 0; i < 1000; i++ {
		l.Append(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Insert(500, i)
		_, _ = l.Delete(500)
	}
}
