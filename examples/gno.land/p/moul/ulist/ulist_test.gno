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
