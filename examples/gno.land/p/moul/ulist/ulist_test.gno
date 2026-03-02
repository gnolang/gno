package ulist

import (
	"testing"

	"gno.land/p/moul/typeutil"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
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
		expected any
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
func generateSequence(n int) []any {
	result := make([]any, n)
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
		values    []any
		start     int
		end       int
		expected  []Entry
		wantStop  bool
		stopAfter int // stop after N elements, -1 for no stop
	}{
		{
			name:      "empty list",
			values:    []any{},
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
			values: []any{42},
			start:  0,
			end:    0,
			expected: []Entry{
				{Index: 0, Value: 42},
			},
			stopAfter: -1,
		},
		{
			name:   "multiple elements forward",
			values: []any{1, 2, 3, 4, 5},
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
			values: []any{1, 2, 3, 4, 5},
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
			values: []any{1, 2, 3, 4, 5},
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
			values: []any{1, 2, 3, 4, 5},
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
			values:    []any{1, 2, 3, 4, 5},
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
			values: []any{1, 2, 3},
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
			values: []any{1, 2, 3},
			start:  0,
			end:    -2,
			expected: []Entry{
				{Index: 0, Value: 1},
			},
			stopAfter: -1,
		},
		{
			name:      "start beyond size",
			values:    []any{1, 2, 3},
			start:     5,
			end:       6,
			expected:  []Entry{},
			stopAfter: -1,
		},
		{
			name:   "end beyond size",
			values: []any{1, 2, 3},
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
			values: []any{1, 2, nil, 4, 5},
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
			values: []any{1, nil, 3, nil, 5},
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
			values:    []any{1, 2, 3},
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
			stopped := list.Iterator(tt.start, tt.end, func(index int, value any) bool {
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
				l.Iterator(0, 2, func(index int, value any) bool {
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

func TestIteratorByOffset(t *testing.T) {
	tests := []struct {
		name     string
		values   []any
		offset   int
		count    int
		expected []Entry
		wantStop bool
	}{
		{
			name:     "empty list",
			values:   []any{},
			offset:   0,
			count:    5,
			expected: []Entry{},
			wantStop: false,
		},
		{
			name:   "positive count forward iteration",
			values: []any{1, 2, 3, 4, 5},
			offset: 1,
			count:  2,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
			},
			wantStop: false,
		},
		{
			name:   "negative count backward iteration",
			values: []any{1, 2, 3, 4, 5},
			offset: 3,
			count:  -2,
			expected: []Entry{
				{Index: 3, Value: 4},
				{Index: 2, Value: 3},
			},
			wantStop: false,
		},
		{
			name:   "count exceeds available elements forward",
			values: []any{1, 2, 3},
			offset: 1,
			count:  5,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
			},
			wantStop: false,
		},
		{
			name:   "count exceeds available elements backward",
			values: []any{1, 2, 3},
			offset: 1,
			count:  -5,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 0, Value: 1},
			},
			wantStop: false,
		},
		{
			name:     "zero count",
			values:   []any{1, 2, 3},
			offset:   0,
			count:    0,
			expected: []Entry{},
			wantStop: false,
		},
		{
			name:   "negative offset",
			values: []any{1, 2, 3},
			offset: -1,
			count:  2,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
			},
			wantStop: false,
		},
		{
			name:   "offset beyond size",
			values: []any{1, 2, 3},
			offset: 5,
			count:  -2,
			expected: []Entry{
				{Index: 2, Value: 3},
				{Index: 1, Value: 2},
			},
			wantStop: false,
		},
		{
			name:   "with deleted elements",
			values: []any{1, nil, 3, nil, 5},
			offset: 0,
			count:  3,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 2, Value: 3},
				{Index: 4, Value: 5},
			},
			wantStop: false,
		},
		{
			name:   "early stop in forward iteration",
			values: []any{1, 2, 3, 4, 5},
			offset: 0,
			count:  5,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
			},
			wantStop: true, // The callback will return true after 2 elements
		},
		{
			name:   "early stop in backward iteration",
			values: []any{1, 2, 3, 4, 5},
			offset: 4,
			count:  -5,
			expected: []Entry{
				{Index: 4, Value: 5},
				{Index: 3, Value: 4},
			},
			wantStop: true, // The callback will return true after 2 elements
		},
		{
			name:     "nil list",
			values:   nil,
			offset:   0,
			count:    5,
			expected: []Entry{},
			wantStop: false,
		},
		{
			name:   "single element forward",
			values: []any{1},
			offset: 0,
			count:  5,
			expected: []Entry{
				{Index: 0, Value: 1},
			},
			wantStop: false,
		},
		{
			name:   "single element backward",
			values: []any{1},
			offset: 0,
			count:  -5,
			expected: []Entry{
				{Index: 0, Value: 1},
			},
			wantStop: false,
		},
		{
			name:     "all deleted elements",
			values:   []any{nil, nil, nil},
			offset:   0,
			count:    3,
			expected: []Entry{},
			wantStop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := New()
			list.Append(tt.values...)

			var result []Entry
			var cb IterCbFn
			if tt.wantStop {
				cb = func(index int, value any) bool {
					result = append(result, Entry{Index: index, Value: value})
					return len(result) >= 2 // Stop after 2 elements for early stop tests
				}
			} else {
				cb = func(index int, value any) bool {
					result = append(result, Entry{Index: index, Value: value})
					return false
				}
			}

			stopped := list.IteratorByOffset(tt.offset, tt.count, cb)

			uassert.Equal(t, len(tt.expected), len(result), "comparing length")
			for i := range result {
				uassert.Equal(t, tt.expected[i].Index, result[i].Index, "comparing index")
				uassert.Equal(t, typeutil.ToString(tt.expected[i].Value), typeutil.ToString(result[i].Value), "comparing value")
			}
			uassert.Equal(t, tt.wantStop, stopped, "comparing stopped")
		})
	}
}

func TestMustDelete(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *List
		indices     []int
		shouldPanic bool
		panicMsg    string
	}{
		{
			name: "successful delete",
			setup: func() *List {
				l := New()
				l.Append(1, 2, 3)
				return l
			},
			indices:     []int{1},
			shouldPanic: false,
		},
		{
			name: "out of bounds",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			indices:     []int{1},
			shouldPanic: true,
			panicMsg:    ErrOutOfBounds.Error(),
		},
		{
			name: "already deleted",
			setup: func() *List {
				l := New()
				l.Append(1)
				l.Delete(0)
				return l
			},
			indices:     []int{0},
			shouldPanic: true,
			panicMsg:    ErrDeleted.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			if tt.shouldPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Error("Expected panic but got none")
					}
					err, ok := r.(error)
					if !ok {
						t.Errorf("Expected error but got %v", r)
					}
					uassert.Equal(t, tt.panicMsg, err.Error())
				}()
			}
			l.MustDelete(tt.indices...)
			if tt.shouldPanic {
				t.Error("Expected panic")
			}
		})
	}
}

func TestMustGet(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *List
		index       int
		expected    any
		shouldPanic bool
		panicMsg    string
	}{
		{
			name: "successful get",
			setup: func() *List {
				l := New()
				l.Append(42)
				return l
			},
			index:       0,
			expected:    42,
			shouldPanic: false,
		},
		{
			name: "out of bounds negative",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:       -1,
			shouldPanic: true,
			panicMsg:    ErrOutOfBounds.Error(),
		},
		{
			name: "out of bounds positive",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:       1,
			shouldPanic: true,
			panicMsg:    ErrOutOfBounds.Error(),
		},
		{
			name: "deleted element",
			setup: func() *List {
				l := New()
				l.Append(1)
				l.Delete(0)
				return l
			},
			index:       0,
			shouldPanic: true,
			panicMsg:    ErrDeleted.Error(),
		},
		{
			name: "nil list",
			setup: func() *List {
				return nil
			},
			index:       0,
			shouldPanic: true,
			panicMsg:    ErrOutOfBounds.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			if tt.shouldPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Error("Expected panic but got none")
					}
					err, ok := r.(error)
					if !ok {
						t.Errorf("Expected error but got %v", r)
					}
					uassert.Equal(t, tt.panicMsg, err.Error())
				}()
			}
			result := l.MustGet(tt.index)
			if tt.shouldPanic {
				t.Error("Expected panic")
			}
			uassert.Equal(t, typeutil.ToString(tt.expected), typeutil.ToString(result))
		})
	}
}

func TestGetRange(t *testing.T) {
	tests := []struct {
		name     string
		values   []any
		start    int
		end      int
		expected []Entry
	}{
		{
			name:     "empty list",
			values:   []any{},
			start:    0,
			end:      10,
			expected: []Entry{},
		},
		{
			name:   "single element",
			values: []any{42},
			start:  0,
			end:    0,
			expected: []Entry{
				{Index: 0, Value: 42},
			},
		},
		{
			name:   "multiple elements forward",
			values: []any{1, 2, 3, 4, 5},
			start:  1,
			end:    3,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
				{Index: 3, Value: 4},
			},
		},
		{
			name:   "multiple elements reverse",
			values: []any{1, 2, 3, 4, 5},
			start:  3,
			end:    1,
			expected: []Entry{
				{Index: 3, Value: 4},
				{Index: 2, Value: 3},
				{Index: 1, Value: 2},
			},
		},
		{
			name:   "with deleted elements",
			values: []any{1, nil, 3, nil, 5},
			start:  0,
			end:    4,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 2, Value: 3},
				{Index: 4, Value: 5},
			},
		},
		{
			name:     "nil list",
			values:   nil,
			start:    0,
			end:      5,
			expected: []Entry{},
		},
		{
			name:     "negative indices",
			values:   []any{1, 2, 3},
			start:    -1,
			end:      -2,
			expected: []Entry{},
		},
		{
			name:   "indices beyond size",
			values: []any{1, 2, 3},
			start:  1,
			end:    5,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := New()
			list.Append(tt.values...)

			result := list.GetRange(tt.start, tt.end)

			uassert.Equal(t, len(tt.expected), len(result), "comparing length")
			for i := range result {
				uassert.Equal(t, tt.expected[i].Index, result[i].Index, "comparing index")
				uassert.Equal(t, typeutil.ToString(tt.expected[i].Value), typeutil.ToString(result[i].Value), "comparing value")
			}
		})
	}
}

func TestGetByOffset(t *testing.T) {
	tests := []struct {
		name     string
		values   []any
		offset   int
		count    int
		expected []Entry
	}{
		{
			name:     "empty list",
			values:   []any{},
			offset:   0,
			count:    5,
			expected: []Entry{},
		},
		{
			name:   "positive count forward",
			values: []any{1, 2, 3, 4, 5},
			offset: 1,
			count:  2,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
			},
		},
		{
			name:   "negative count backward",
			values: []any{1, 2, 3, 4, 5},
			offset: 3,
			count:  -2,
			expected: []Entry{
				{Index: 3, Value: 4},
				{Index: 2, Value: 3},
			},
		},
		{
			name:   "count exceeds available elements",
			values: []any{1, 2, 3},
			offset: 1,
			count:  5,
			expected: []Entry{
				{Index: 1, Value: 2},
				{Index: 2, Value: 3},
			},
		},
		{
			name:     "zero count",
			values:   []any{1, 2, 3},
			offset:   0,
			count:    0,
			expected: []Entry{},
		},
		{
			name:   "with deleted elements",
			values: []any{1, nil, 3, nil, 5},
			offset: 0,
			count:  3,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 2, Value: 3},
				{Index: 4, Value: 5},
			},
		},
		{
			name:   "negative offset",
			values: []any{1, 2, 3},
			offset: -1,
			count:  2,
			expected: []Entry{
				{Index: 0, Value: 1},
				{Index: 1, Value: 2},
			},
		},
		{
			name:   "offset beyond size",
			values: []any{1, 2, 3},
			offset: 5,
			count:  -2,
			expected: []Entry{
				{Index: 2, Value: 3},
				{Index: 1, Value: 2},
			},
		},
		{
			name:     "nil list",
			values:   nil,
			offset:   0,
			count:    5,
			expected: []Entry{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := New()
			list.Append(tt.values...)

			result := list.GetByOffset(tt.offset, tt.count)

			uassert.Equal(t, len(tt.expected), len(result), "comparing length")
			for i := range result {
				uassert.Equal(t, tt.expected[i].Index, result[i].Index, "comparing index")
				uassert.Equal(t, typeutil.ToString(tt.expected[i].Value), typeutil.ToString(result[i].Value), "comparing value")
			}
		})
	}
}

func TestMustSet(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *List
		index       int
		value       any
		shouldPanic bool
		panicMsg    string
	}{
		{
			name: "successful set",
			setup: func() *List {
				l := New()
				l.Append(42)
				return l
			},
			index:       0,
			value:       99,
			shouldPanic: false,
		},
		{
			name: "restore deleted element",
			setup: func() *List {
				l := New()
				l.Append(42)
				l.Delete(0)
				return l
			},
			index:       0,
			value:       99,
			shouldPanic: false,
		},
		{
			name: "out of bounds negative",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:       -1,
			value:       99,
			shouldPanic: true,
			panicMsg:    ErrOutOfBounds.Error(),
		},
		{
			name: "out of bounds positive",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:       1,
			value:       99,
			shouldPanic: true,
			panicMsg:    ErrOutOfBounds.Error(),
		},
		{
			name: "nil list",
			setup: func() *List {
				return nil
			},
			index:       0,
			value:       99,
			shouldPanic: true,
			panicMsg:    ErrOutOfBounds.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			if tt.shouldPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Error("Expected panic but got none")
					}
					err, ok := r.(error)
					if !ok {
						t.Errorf("Expected error but got %v", r)
					}
					uassert.Equal(t, tt.panicMsg, err.Error())
				}()
			}
			l.MustSet(tt.index, tt.value)
			if tt.shouldPanic {
				t.Error("Expected panic")
			}
			// Verify the value was set correctly for non-panic cases
			if !tt.shouldPanic {
				result := l.Get(tt.index)
				uassert.Equal(t, typeutil.ToString(tt.value), typeutil.ToString(result))
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *List
		index       int
		value       any
		expectedErr error
		verify      func(t *testing.T, l *List)
	}{
		{
			name: "set value in empty list",
			setup: func() *List {
				return New()
			},
			index:       0,
			value:       42,
			expectedErr: ErrOutOfBounds,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 0, l.Size())
			},
		},
		{
			name: "set value at valid index",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index: 0,
			value: 42,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 42, l.Get(0))
				uassert.Equal(t, 1, l.Size())
				uassert.Equal(t, 1, l.TotalSize())
			},
		},
		{
			name: "set value at negative index",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:       -1,
			value:       42,
			expectedErr: ErrOutOfBounds,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 1, l.Get(0))
			},
		},
		{
			name: "set value beyond size",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index:       1,
			value:       42,
			expectedErr: ErrOutOfBounds,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 1, l.Get(0))
				uassert.Equal(t, 1, l.Size())
			},
		},
		{
			name: "set nil value",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index: 0,
			value: nil,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, nil, l.Get(0))
				uassert.Equal(t, 0, l.Size())
			},
		},
		{
			name: "set value at deleted index",
			setup: func() *List {
				l := New()
				l.Append(1, 2, 3)
				l.Delete(1)
				return l
			},
			index: 1,
			value: 42,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 42, l.Get(1))
				uassert.Equal(t, 3, l.Size())
				uassert.Equal(t, 3, l.TotalSize())
			},
		},
		{
			name: "set value in nil list",
			setup: func() *List {
				return nil
			},
			index:       0,
			value:       42,
			expectedErr: ErrOutOfBounds,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 0, l.Size())
			},
		},
		{
			name: "set multiple values at same index",
			setup: func() *List {
				l := New()
				l.Append(1)
				return l
			},
			index: 0,
			value: 42,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 42, l.Get(0))
				err := l.Set(0, 99)
				uassert.Equal(t, nil, err)
				uassert.Equal(t, 99, l.Get(0))
				uassert.Equal(t, 1, l.Size())
			},
		},
		{
			name: "set value at last index",
			setup: func() *List {
				l := New()
				l.Append(1, 2, 3)
				return l
			},
			index: 2,
			value: 42,
			verify: func(t *testing.T, l *List) {
				uassert.Equal(t, 42, l.Get(2))
				uassert.Equal(t, 3, l.Size())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			err := l.Set(tt.index, tt.value)

			if tt.expectedErr != nil {
				uassert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				uassert.Equal(t, nil, err)
			}

			tt.verify(t, l)
		})
	}
}
