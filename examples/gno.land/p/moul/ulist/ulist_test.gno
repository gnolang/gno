package ulist

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/typeutil"
)

func TestNew(t *testing.T) {
	l := New()
	uassert.Equal(t, 0, l.Size())
	uassert.Equal(t, 0, l.TotalSize())
}

func TestListAppendAndGet(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *List
		index    int
		expected interface{}
	}{
		{
			name: "empty list",
			setup: func() *List {
				return New()
			},
			index:    0,
			expected: nil,
		},
		{
			name: "single append and get",
			setup: func() *List {
				l := New()
				l.Append(42)
				return l
			},
			index:    0,
			expected: 42,
		},
		{
			name: "multiple appends and get first",
			setup: func() *List {
				l := New()
				l.Append(1)
				l.Append(2)
				l.Append(3)
				return l
			},
			index:    0,
			expected: 1,
		},
		{
			name: "multiple appends and get last",
			setup: func() *List {
				l := New()
				l.Append(1)
				l.Append(2)
				l.Append(3)
				return l
			},
			index:    2,
			expected: 3,
		},
		{
			name: "get with invalid index",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:    1,
			expected: nil,
		},
		{
			name: "31 items get first",
			setup: func() *List {
				l := New()
				for i := 0; i < 31; i++ {
					l.Append(i)
				}
				return l
			},
			index:    0,
			expected: 0,
		},
		{
			name: "31 items get last",
			setup: func() *List {
				l := New()
				for i := 0; i < 31; i++ {
					l.Append(i)
				}
				return l
			},
			index:    30,
			expected: 30,
		},
		{
			name: "31 items get middle",
			setup: func() *List {
				l := New()
				for i := 0; i < 31; i++ {
					l.Append(i)
				}
				return l
			},
			index:    15,
			expected: 15,
		},
		{
			name: "values around power of 2 boundary",
			setup: func() *List {
				l := New()
				for i := 0; i < 18; i++ {
					l.Append(i)
				}
				return l
			},
			index:    15,
			expected: 15,
		},
		{
			name: "values at power of 2",
			setup: func() *List {
				l := New()
				for i := 0; i < 18; i++ {
					l.Append(i)
				}
				return l
			},
			index:    16,
			expected: 16,
		},
		{
			name: "values after power of 2",
			setup: func() *List {
				l := New()
				for i := 0; i < 18; i++ {
					l.Append(i)
				}
				return l
			},
			index:    17,
			expected: 17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			got := l.Get(tt.index)
			if got != tt.expected {
				t.Errorf("List.Get() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// generateSequence creates a slice of integers from 0 to n-1
func generateSequence(n int) []interface{} {
	result := make([]interface{}, n)
	for i := 0; i < n; i++ {
		result[i] = i
	}
	return result
}

func TestListDelete(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *List
		deleteIndices []int
		expectedErr   error
		expectedSize  int
	}{
		{
			name: "delete single element",
			setup: func() *List {
				l := New()
				l.Append(1, 2, 3)
				return l
			},
			deleteIndices: []int{1},
			expectedErr:   nil,
			expectedSize:  2,
		},
		{
			name: "delete multiple elements",
			setup: func() *List {
				l := New()
				l.Append(1, 2, 3, 4, 5)
				return l
			},
			deleteIndices: []int{0, 2, 4},
			expectedErr:   nil,
			expectedSize:  2,
		},
		{
			name: "delete with negative index",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			deleteIndices: []int{-1},
			expectedErr:   ErrOutOfBounds,
			expectedSize:  1,
		},
		{
			name: "delete beyond size",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			deleteIndices: []int{1},
			expectedErr:   ErrOutOfBounds,
			expectedSize:  1,
		},
		{
			name: "delete already deleted element",
			setup: func() *List {
				l := New()
				l.Append(1)
				l.Delete(0)
				return l
			},
			deleteIndices: []int{0},
			expectedErr:   ErrDeleted,
			expectedSize:  0,
		},
		{
			name: "delete multiple elements in reverse",
			setup: func() *List {
				l := New()
				l.Append(1, 2, 3, 4, 5)
				return l
			},
			deleteIndices: []int{4, 2, 0},
			expectedErr:   nil,
			expectedSize:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			initialSize := l.Size()
			err := l.Delete(tt.deleteIndices...)
			if err != nil && tt.expectedErr != nil {
				uassert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				uassert.Equal(t, tt.expectedErr, err)
			}
			uassert.Equal(t, tt.expectedSize, l.Size(),
				ufmt.Sprintf("Expected size %d after deleting %d elements from size %d, got %d",
					tt.expectedSize, len(tt.deleteIndices), initialSize, l.Size()))
		})
	}
}

func TestListSizeAndTotalSize(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		list := New()
		uassert.Equal(t, 0, list.Size())
		uassert.Equal(t, 0, list.TotalSize())
	})

	t.Run("list with elements", func(t *testing.T) {
		list := New()
		list.Append(1)
		list.Append(2)
		list.Append(3)
		uassert.Equal(t, 3, list.Size())
		uassert.Equal(t, 3, list.TotalSize())
	})

	t.Run("list with deleted elements", func(t *testing.T) {
		list := New()
		list.Append(1)
		list.Append(2)
		list.Append(3)
		list.Delete(1)
		uassert.Equal(t, 2, list.Size())
		uassert.Equal(t, 3, list.TotalSize())
	})
}

func TestIterator(t *testing.T) {
	tests := []struct {
		name      string
		values    []interface{}
		start     int
		end       int
		expected  []Entry
		wantStop  bool
		stopAfter int // stop after N elements, -1 for no stop
	}{
		{
			name:      "empty list",
			values:    []interface{}{},
			start:     0,
			end:       10,
			expected:  []Entry{},
			stopAfter: -1,
		},
		{
			name:      "nil list",
			values:    nil,
			start:     0,
			end:       0,
			expected:  []Entry{},
			stopAfter: -1,
		},
		{
			name:   "single element forward",
			values: []interface{}{42},
			start:  0,
			end:    0,
			expected: []Entry{
				{Index: 0, Value: 42},
			},
			stopAfter: -1,
		},
		{
			name:   "multiple elements forward",
			values: []interface{}{1, 2, 3, 4, 5},
			start:  0,
			end:    4,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
				{Index: 3, Value: 4},
				{Index: 4, Value: 5},
			},
			stopAfter: -1,
		},
		{
			name:   "multiple elements reverse",
			values: []interface{}{1, 2, 3, 4, 5},
			start:  4,
			end:    0,
			expected: []Entry{
				{Index: 4, Value: 5},
				{Index: 3, Value: 4},
				{Index: 2, Value: 3},
				{Index: 1, Value: 2},
				{Index: 0, Value: 1},
			},
			stopAfter: -1,
		},
		{
			name:   "partial range forward",
			values: []interface{}{1, 2, 3, 4, 5},
			start:  1,
			end:    3,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
				{Index: 3, Value: 4},
			},
			stopAfter: -1,
		},
		{
			name:   "partial range reverse",
			values: []interface{}{1, 2, 3, 4, 5},
			start:  3,
			end:    1,
			expected: []Entry{
				{Index: 3, Value: 4},
				{Index: 2, Value: 3},
				{Index: 1, Value: 2},
			},
			stopAfter: -1,
		},
		{
			name:      "stop iteration early",
			values:    []interface{}{1, 2, 3, 4, 5},
			start:     0,
			end:       4,
			wantStop:  true,
			stopAfter: 2,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
			},
		},
		{
			name:   "negative start",
			values: []interface{}{1, 2, 3},
			start:  -1,
			end:    2,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
			},
			stopAfter: -1,
		},
		{
			name:   "negative end",
			values: []interface{}{1, 2, 3},
			start:  0,
			end:    -2,
			expected: []Entry{
				{Index: 0, Value: 1},
			},
			stopAfter: -1,
		},
		{
			name:      "start beyond size",
			values:    []interface{}{1, 2, 3},
			start:     5,
			end:       6,
			expected:  []Entry{},
			stopAfter: -1,
		},
		{
			name:   "end beyond size",
			values: []interface{}{1, 2, 3},
			start:  0,
			end:    5,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
			},
			stopAfter: -1,
		},
		{
			name:   "with deleted elements",
			values: []interface{}{1, 2, nil, 4, 5},
			start:  0,
			end:    4,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
				{Index: 3, Value: 4},
				{Index: 4, Value: 5},
			},
			stopAfter: -1,
		},
		{
			name:   "with deleted elements reverse",
			values: []interface{}{1, nil, 3, nil, 5},
			start:  4,
			end:    0,
			expected: []Entry{
				{Index: 4, Value: 5},
				{Index: 2, Value: 3},
				{Index: 0, Value: 1},
			},
			stopAfter: -1,
		},
		{
			name:      "start equals end",
			values:    []interface{}{1, 2, 3},
			start:     1,
			end:       1,
			expected:  []Entry{{Index: 1, Value: 2}},
			stopAfter: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := New()
			list.Append(tt.values...)

			var result []Entry
			stopped := list.Iterator(tt.start, tt.end, func(index int, value interface{}) bool {
				result = append(result, Entry{Index: index, Value: value})
				return tt.stopAfter >= 0 && len(result) >= tt.stopAfter
			})

			uassert.Equal(t, len(result), len(tt.expected), "comparing length")

			for i := range result {
				uassert.Equal(t, result[i].Index, tt.expected[i].Index, "comparing index")
				uassert.Equal(t, typeutil.ToString(result[i].Value), typeutil.ToString(tt.expected[i].Value), "comparing value")
			}

			uassert.Equal(t, stopped, tt.wantStop, "comparing stopped")
		})
	}
}

func TestLargeListAppendGetAndDelete(t *testing.T) {
	l := New()
	size := 100

	// Append values from 0 to 99
	for i := 0; i < size; i++ {
		l.Append(i)
		val := l.Get(i)
		uassert.Equal(t, i, val)
	}

	// Verify size
	uassert.Equal(t, size, l.Size())
	uassert.Equal(t, size, l.TotalSize())

	// Get and verify each value
	for i := 0; i < size; i++ {
		val := l.Get(i)
		uassert.Equal(t, i, val)
	}

	// Get and verify each value
	for i := 0; i < size; i++ {
		err := l.Delete(i)
		uassert.Equal(t, nil, err)
	}

	// Verify size
	uassert.Equal(t, 0, l.Size())
	uassert.Equal(t, size, l.TotalSize())

	// Get and verify each value
	for i := 0; i < size; i++ {
		val := l.Get(i)
		uassert.Equal(t, nil, val)
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "nil list operations",
			test: func(t *testing.T) {
				var l *List
				uassert.Equal(t, 0, l.Size())
				uassert.Equal(t, 0, l.TotalSize())
				uassert.Equal(t, nil, l.Get(0))
				err := l.Delete(0)
				uassert.Equal(t, ErrOutOfBounds.Error(), err.Error())
			},
		},
		{
			name: "delete empty indices slice",
			test: func(t *testing.T) {
				l := New()
				l.Append(1)
				err := l.Delete()
				uassert.Equal(t, nil, err)
				uassert.Equal(t, 1, l.Size())
			},
		},
		{
			name: "append nil values",
			test: func(t *testing.T) {
				l := New()
				l.Append(nil, nil)
				uassert.Equal(t, 2, l.Size())
				uassert.Equal(t, nil, l.Get(0))
				uassert.Equal(t, nil, l.Get(1))
			},
		},
		{
			name: "delete same index multiple times",
			test: func(t *testing.T) {
				l := New()
				l.Append(1, 2, 3)
				err := l.Delete(1)
				uassert.Equal(t, nil, err)
				err = l.Delete(1)
				uassert.Equal(t, ErrDeleted.Error(), err.Error())
			},
		},
		{
			name: "iterator with all deleted elements",
			test: func(t *testing.T) {
				l := New()
				l.Append(1, 2, 3)
				l.Delete(0, 1, 2)
				var count int
				l.Iterator(0, 2, func(index int, value interface{}) bool {
					count++
					return false
				})
				uassert.Equal(t, 0, count)
			},
		},
		{
			name: "append after delete",
			test: func(t *testing.T) {
				l := New()
				l.Append(1, 2)
				l.Delete(1)
				l.Append(3)
				uassert.Equal(t, 2, l.Size())
				uassert.Equal(t, 3, l.TotalSize())
				uassert.Equal(t, 1, l.Get(0))
				uassert.Equal(t, nil, l.Get(1))
				uassert.Equal(t, 3, l.Get(2))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}
