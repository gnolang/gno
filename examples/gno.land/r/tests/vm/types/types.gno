// package to test types behavior in various conditions (TXs, imports).
package types

import (
	"errors"
	"time"

	"gno.land/p/nt/avl"
)

var (
	gInt         int      = -42
	gUint        uint     = 42
	gString      string   = "a string"
	gStringSlice []string = []string{"a", "string", "slice"}
	gError       error    = errors.New("an error")
	gIntSlice    []int    = []int{-42, 0, 42}
	gUintSlice   []uint   = []uint{0, 42, 84}
	gTree        avl.Tree
	// gInterface  = any{-42, "a string", uint(42)}
)

func init() {
	gTree.Set("a", "content of A")
	gTree.Set("b", "content of B")
}

func Noop()                     {}
func RetTimeNow() time.Time     { return time.Now() }
func RetString() string         { return gString }
func RetStringPointer() *string { return &gString }
func RetUint() uint             { return gUint }
func RetInt() int               { return gInt }
func RetUintPointer() *uint     { return &gUint }
func RetIntPointer() *int       { return &gInt }
func RetTree() avl.Tree         { return gTree }
func RetIntSlice() []int        { return gIntSlice }
func RetUintSlice() []uint      { return gUintSlice }
func RetStringSlice() []string  { return gStringSlice }
func RetError() error           { return gError }
func Panic()                    { panic("PANIC!") }

// TODO: floats
// TODO: typed errors
// TODO: ret interface
// TODO: recover
// TODO: take types as input

func Render(path string) string {
	return "package to test data types."
}
