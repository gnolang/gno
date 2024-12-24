package ulist

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	l := New()
	if l.Len() != 0 {
		t.Errorf("New list should be empty, got length %d", l.Len())
	}
}

func TestAppendMany(t *testing.T) {
	values := []interface{}{1, 2, 3, 4, 5}
	l := New()
	l.AppendMany(values...)

	if l.Len() != len(values) {
		t.Errorf("Expected length %d, got %d", len(values), l.Len())
	}

	for i, want := range values {
		got, err := l.Get(i)
		if err != nil || got != want {
			t.Errorf("At index %d: want %v, got %v, err %v", i, want, got, err)
		}
	}
}

func TestList_Set(t *testing.T) {
	l := New()
	l.AppendMany(1, 2, 3)

	tests := []struct {
		index   int
		value   interface{}
		wantErr error
	}{
		{1, 42, nil},
		{-1, 0, ErrIndexOutOfRange},
		{3, 0, ErrIndexOutOfRange},
	}

	for _, tt := range tests {
		err := l.Set(tt.index, tt.value)
		if err != tt.wantErr {
			t.Errorf("Set(%d, %v) error = %v, want %v", tt.index, tt.value, err, tt.wantErr)
			continue
		}
		if err == nil {
			got, _ := l.Get(tt.index)
			if got != tt.value {
				t.Errorf("After Set(%d, %v), Get(%d) = %v", tt.index, tt.value, tt.index, got)
			}
		}
	}
}

func TestList_Delete(t *testing.T) {
	l := New()
	l.AppendMany(1, 2, 3, 4, 5)

	tests := []struct {
		index   int
		want    interface{}
		wantErr error
	}{
		{0, 1, nil},                   // beginning
		{3, 5, nil},                   // end
		{1, 3, nil},                   // middle
		{-1, nil, ErrIndexOutOfRange}, // invalid
	}

	for _, tt := range tests {
		got, err := l.Delete(tt.index)
		if err != tt.wantErr || (err == nil && got != tt.want) {
			t.Errorf("Delete(%d) = %v, %v; want %v, %v", tt.index, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestList_Range(t *testing.T) {
	l := New()
	l.AppendMany(1, 2, 3, 4, 5)

	tests := []struct {
		start, end int
		want       []interface{}
		wantErr    error
	}{
		{0, 5, []interface{}{1, 2, 3, 4, 5}, nil},
		{1, 4, []interface{}{2, 3, 4}, nil},
		{0, 0, []interface{}{}, nil},
		{-1, 5, nil, ErrInvalidArgument},
		{3, 2, nil, ErrInvalidArgument},
	}

	for _, tt := range tests {
		got, err := l.Range(tt.start, tt.end)
		if err != tt.wantErr {
			t.Errorf("Range(%d, %d) error = %v, want %v", tt.start, tt.end, err, tt.wantErr)
			continue
		}
		if err == nil && !sliceEqual(got, tt.want) {
			t.Errorf("Range(%d, %d) = %v, want %v", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestList_Iterator(t *testing.T) {
	values := []interface{}{1, 2, 3, 4, 5}
	l := New()
	l.AppendMany(values...)
	it := l.NewIterator()

	var got []interface{}
	for value, ok := it.Next(); ok; value, ok = it.Next() {
		got = append(got, value)
	}

	if !sliceEqual(got, values) {
		t.Errorf("Iterator values = %v, want %v", got, values)
	}
}

func TestList_Clear(t *testing.T) {
	l := New()
	l.AppendMany(1, 2, 3)
	l.Clear()

	if l.Len() != 0 {
		t.Errorf("After Clear(), Len() = %d, want 0", l.Len())
	}
}

func TestNewWithSize(t *testing.T) {
	tests := []struct {
		size     int
		wantSize int
	}{
		{16, 16}, // default
		{2, 4},   // too small
		{15, 16}, // odd
		{32, 32}, // large even
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("size_%d", tt.size), func(t *testing.T) {
			l := NewWithSize(tt.size)
			if l.maxNodeSize != tt.wantSize {
				t.Errorf("NewWithSize(%d) got size = %d, want %d", tt.size, l.maxNodeSize, tt.wantSize)
			}
		})
	}
}

func BenchmarkListWithDifferentSizes(b *testing.B) {
	sizes := []int{4, 8, 16, 32, 64, 128}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			l := NewWithSize(size)

			b.Run("append", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					l.Append(i)
				}
			})

			l = NewWithSize(size)
			for i := 0; i < 1000; i++ {
				l.Append(i)
			}

			b.Run("get", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, _ = l.Get(i % 1000)
				}
			})

			b.Run("insert_delete", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = l.Insert(500, i)
					_, _ = l.Delete(500)
				}
			})
		})
	}
}

func TestList_InsertEdgeCases(t *testing.T) {
	defaultNodeSize := 16
	tests := []struct {
		name     string
		setup    func() *List
		index    int
		value    interface{}
		wantErr  error
		validate func(t *testing.T, l *List)
	}{
		{
			name: "insert_at_start_empty_list",
			setup: func() *List {
				return New()
			},
			index: 0,
			value: 1,
			validate: func(t *testing.T, l *List) {
				if l.Len() != 1 {
					t.Errorf("length = %d, want 1", l.Len())
				}
				if v, _ := l.Get(0); v != 1 {
					t.Errorf("value = %v, want 1", v)
				}
			},
		},
		{
			name: "insert_at_start_nonempty_list",
			setup: func() *List {
				l := New()
				l.AppendMany(1, 2, 3)
				return l
			},
			index: 0,
			value: 0,
			validate: func(t *testing.T, l *List) {
				want := []interface{}{0, 1, 2, 3}
				for i, w := range want {
					if v, _ := l.Get(i); v != w {
						t.Errorf("at index %d: got %v, want %v", i, v, w)
					}
				}
			},
		},
		{
			name: "insert_at_end_with_space",
			setup: func() *List {
				l := New()
				l.AppendMany(1, 2, 3)
				return l
			},
			index: 3,
			value: 4,
			validate: func(t *testing.T, l *List) {
				want := []interface{}{1, 2, 3, 4}
				for i, w := range want {
					if v, _ := l.Get(i); v != w {
						t.Errorf("at index %d: got %v, want %v", i, v, w)
					}
				}
			},
		},
		{
			name: "insert_at_end_full_chunk",
			setup: func() *List {
				l := NewWithSize(4)
				l.AppendMany(1, 2, 3, 4) // Fill first chunk
				return l
			},
			index: 4,
			value: 5,
			validate: func(t *testing.T, l *List) {
				want := []interface{}{1, 2, 3, 4, 5}
				for i, w := range want {
					if v, _ := l.Get(i); v != w {
						t.Errorf("at index %d: got %v, want %v", i, v, w)
					}
				}
			},
		},
		{
			name: "insert_middle_with_space",
			setup: func() *List {
				l := New()
				l.AppendMany(1, 2, 4, 5)
				return l
			},
			index: 2,
			value: 3,
			validate: func(t *testing.T, l *List) {
				want := []interface{}{1, 2, 3, 4, 5}
				for i, w := range want {
					if v, _ := l.Get(i); v != w {
						t.Errorf("at index %d: got %v, want %v", i, v, w)
					}
				}
			},
		},
		{
			name: "insert_middle_full_chunk",
			setup: func() *List {
				l := NewWithSize(defaultNodeSize)
				for i := 1; i <= 4; i++ {
					l.Append(i)
				}
				return l
			},
			index: 2,
			value: 2.5,
			validate: func(t *testing.T, l *List) {
				expected := []interface{}{1, 2, 2.5, 3, 4}
				for i, want := range expected {
					got, err := l.Get(i)
					if err != nil {
						t.Errorf("Get(%d) error = %v", i, err)
						continue
					}
					if got != want {
						t.Errorf("Get(%d) = %v, want %v", i, got, want)
					}
				}
			},
		},
		{
			name: "sparse_insert_creating_gap",
			setup: func() *List {
				l := NewWithSize(16)
				l.Append(1)
				l.Append(2)
				l.Append(3)
				return l
			},
			index:   18,
			value:   int(4),
			wantErr: nil,
			validate: func(t *testing.T, l *List) {
				// Validation logic here
			},
		},
		{
			name: "insert_invalid_negative",
			setup: func() *List {
				return New()
			},
			index:   -1,
			value:   1,
			wantErr: ErrIndexOutOfRange,
		},
		{
			name: "insert_invalid_too_large",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:   2, // Beyond length
			value:   2,
			wantErr: ErrIndexOutOfRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			err := l.Insert(tt.index, tt.value)

			if err != tt.wantErr {
				t.Errorf("Insert() error = %v, want %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, l)
			}
		})
	}
}

// Additional test for chunk boundaries
func TestList_ChunkBoundaries(t *testing.T) {
	l := NewWithSize(4) // Small size to test boundaries

	// Fill first chunk
	l.AppendMany(1, 2, 3, 4)

	// Insert at various positions to test chunk splitting
	tests := []struct {
		index int
		value interface{}
		want  []interface{}
	}{
		{0, 0, []interface{}{0, 1, 2, 3, 4}},           // Start of first chunk
		{5, 5, []interface{}{0, 1, 2, 3, 4, 5}},        // New chunk
		{3, 2.5, []interface{}{0, 1, 2, 2.5, 3, 4, 5}}, // Middle of first chunk
	}

	for i, tt := range tests {
		err := l.Insert(tt.index, tt.value)
		if err != nil {
			t.Errorf("test %d: Insert(%d, %v) unexpected error: %v", i, tt.index, tt.value, err)
			continue
		}

		// Verify entire list content
		for j, want := range tt.want {
			if got, _ := l.Get(j); got != want {
				t.Errorf("test %d: at index %d: got %v, want %v", i, j, got, want)
			}
		}
	}
}

// Helper function for comparing slices
func sliceEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
