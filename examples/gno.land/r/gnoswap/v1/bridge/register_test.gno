package bridge

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
)

func TestVariousFunctionCallbacks(t *testing.T) {
	callbackTree = avl.NewTree()

	tests := []struct {
		name     string
		key      string
		callback any
		args     []any
		want     any
		wantErr  bool
	}{
		{
			name: "Simple addition function",
			key:  "test/add",
			callback: func(a, b int) int {
				return a + b
			},
			args: []any{5, 3},
			want: 8,
		},
		{
			name: "String concatenation",
			key:  "test/concat",
			callback: func(a, b string) string {
				return a + b
			},
			args: []any{"Hello, ", "World"},
			want: "Hello, World",
		},
		{
			name: "No arguments function",
			key:  "test/constant",
			callback: func() int {
				return 42
			},
			args: []any{},
			want: 42,
		},
		{
			name: "Multiple return values",
			key:  "test/divide",
			callback: func(a, b int) (int, bool) {
				if b == 0 {
					return 0, false
				}
				return a / b, true
			},
			args: []any{10, 2},
			want: [2]any{5, true},
		},
		{
			name: "Function with slice argument",
			key:  "test/sum",
			callback: func(numbers []int) int {
				sum := 0
				for _, n := range numbers {
					sum += n
				}
				return sum
			},
			args: []any{[]int{1, 2, 3, 4, 5}},
			want: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterCallback(tt.key, tt.callback)

			cb, exists := GetCallback(tt.key)
			uassert.True(t, exists)

			switch fn := cb.(type) {
			case func(int, int) int:
				if len(tt.args) == 2 {
					result := fn(tt.args[0].(int), tt.args[1].(int))
					uassert.Equal(t, result, tt.want)
				}
			case func(string, string) string:
				if len(tt.args) == 2 {
					result := fn(tt.args[0].(string), tt.args[1].(string))
					uassert.Equal(t, result, tt.want)
				}
			case func() int:
				result := fn()
				uassert.Equal(t, result, tt.want)
			case func(int, int) (int, bool):
				if len(tt.args) == 2 {
					quotient, ok := fn(tt.args[0].(int), tt.args[1].(int))
					expected := tt.want.([2]any)
					if quotient != expected[0].(int) || ok != expected[1].(bool) {
						t.Errorf("got (%v, %v), want (%v, %v)",
							quotient, ok, expected[0], expected[1])
					}
				}
			case func([]int) int:
				if len(tt.args) == 1 {
					result := fn(tt.args[0].([]int))
					uassert.Equal(t, result, tt.want)
				}
			default:
				t.Errorf("Unexpected function type for %s", tt.name)
			}
		})
	}
}

func TestDuplicateCallback(t *testing.T) {
	callbackTree = avl.NewTree()

	key := "test/duplicate"
	RegisterCallback(key, func() {})

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on duplicate registration")
		}
	}()

	RegisterCallback(key, func() {})
}

func TestNonExistentCallback(t *testing.T) {
	callbackTree = avl.NewTree()

	_, exists := GetCallback("non/existent/key")
	uassert.False(t, exists)
}
