// Package stateshowcase exercises every kind of declaration the state
// explorer can render. Used as a visual fixture during gnoweb UI work
// — drop the realm once the state explorer is finalised.
package stateshowcase

import (
	"strconv"

	"gno.land/p/nt/ufmt/v0"
)

// Counter tracks how many times the realm has been pinged. Plain int
// to exercise the simplest primitive rendering.
var Counter int = 42

// Big is a 64-bit integer, used to demonstrate the wider-int kind in
// the state explorer (some chains return int as int64).
var Big int64 = 1_000_000_000_000

// Small is a signed 8-bit integer. Negative value picked deliberately
// to test that the value-badge handles the leading minus sign.
var Small int8 = -7

// UCount is an unsigned int — non-negative by construction.
var UCount uint = 100

// UByte is a single byte. Hex literal exercises radix-aware rendering.
var UByte byte = 0xFF

// Pi is a 64-bit float. The state explorer should show the trailing
// digits without precision loss.
var Pi float64 = 3.1415927

// Tau is a 32-bit float — same as Pi but lower precision.
var Tau float32 = 6.283185

// Active is the canonical "is the realm enabled" flag. The walker
// should render this as a semantic green badge.
var Active bool = true

// Disabled is the false counterpart — should render with a muted/red
// treatment so the two states read at a glance.
var Disabled bool = false

// Slogan is the realm's short tagline. Strings render quoted in the
// state explorer so they're visually distinct from numbers.
var Slogan string = "hello, gno"

// EmptyStr is the zero-length string — tests the empty-leaf path.
var EmptyStr string = ""

// Score, Multiplier and Bonus form a multi-value declaration group:
// same `var (..)` block in source, each becomes its own card.
var (
	// Score is the running points total.
	Score int = 1200
	// Multiplier scales reward computations.
	Multiplier int = 3
	// Bonus is a free-form label for the active reward.
	Bonus string = "double-points"
)

// User describes a participant in the realm. Flat struct: primitive
// fields plus a slice of role tags. Used as building block for nested
// containers below.
type User struct {
	Name     string
	Age      int
	Verified bool
	Roles    []string
}

// Owner is the realm's administrator. Top-level struct value renders
// as a card whose body is a fields grid.
var Owner = User{
	Name:     "alice",
	Age:      30,
	Verified: true,
	Roles:    []string{"admin", "editor"},
}

// Org represents a community group within the realm. Carries two
// nested User values to exercise struct-in-struct rendering.
type Org struct {
	Title    string
	Treasury int64
	Founder  User
	Member   User
}

// Acme is a sample organisation pre-seeded so the nested-struct path
// has visible content. Click to expand Founder/Member into their own
// fields grids.
var Acme = Org{
	Title:    "Acme Realm",
	Treasury: 1_000_000,
	Founder:  User{Name: "alice", Age: 30, Verified: true},
	Member:   User{Name: "bob", Age: 27, Verified: false},
}

// CurrentUser is a pointer to a User — exercises the *T heap chain
// path: heap-item → ref → struct. EnrichInlinePreviews resolves the
// indirection so the inner fields surface inline.
var CurrentUser *User = &User{
	Name:     "carol",
	Age:      41,
	Verified: true,
	Roles:    []string{"founder"},
}

// Tags is a string slice with several entries — preview shows a
// sample, the dedicated page lists them all.
var Tags = []string{"alpha", "beta", "gamma", "delta", "epsilon"}

// EmptyTags is a zero-length slice; tests the empty-state rendering.
var EmptyTags = []string{}

// SinglePost has exactly one element — tests the "single value" path.
var SinglePost = []string{"first post"}

// Numbers is a slice of primitives. Cards should right-align the
// numeric values when rendered as field rows.
var Numbers = []int{1, 1, 2, 3, 5, 8, 13, 21}

// Squad is a slice of structs. Each entry expands inline; navigation
// goes to the entry's own dedicated page.
var Squad = []User{
	{Name: "alice", Age: 30, Verified: true},
	{Name: "bob", Age: 27, Verified: false},
	{Name: "carol", Age: 41, Verified: true},
}

// Matrix is a 9-element fixed-size array — tests the [N]T kind.
var Matrix = [9]int{1, 2, 3, 4, 5, 6, 7, 8, 9}

// Settings holds runtime configuration knobs. String→int map exercises
// the basic map rendering with a "Keys of …" scope label.
var Settings = map[string]int{
	"max_users":    1000,
	"max_attempts": 3,
	"timeout_ms":   5000,
}

// Members is a registry of users keyed by login. Heterogeneous map
// values: each entry expands into a nested struct grid.
var Members = map[string]User{
	"alice": {Name: "alice", Age: 30, Verified: true},
	"bob":   {Name: "bob", Age: 27, Verified: false},
}

// EmptySettings — zero-length map for the empty-state path.
var EmptySettings = map[string]int{}

// IntKeys exercises non-string map keys (int → string here).
var IntKeys = map[int]string{
	1: "one",
	2: "two",
	3: "three",
}

// Add returns the sum of two integers. Plain function — the explorer
// renders it as a header-only card on the realm root, with full source
// shown on the dedicated page.
func Add(a, b int) int {
	return a + b
}

// NextID is a closure: it captures `idSeq` from init() and returns
// monotonically-increasing IDs. The state explorer renders closures
// with their captured variables visible inline alongside the source.
var NextID func() int

// AddToBuffer is a closure capturing `Buffer` and `bufferMax`. Tests
// the multi-capture rendering path (slice + int).
var AddToBuffer func(string)

// Buffer is the shared target of AddToBuffer — exposed so the
// captured-state pattern is visible from both sides.
var Buffer []string

func init() {
	idSeq := 0
	NextID = func() int {
		idSeq++
		return idSeq
	}

	bufferMax := 5
	Buffer = make([]string, 0, bufferMax)
	AddToBuffer = func(s string) {
		if len(Buffer) < bufferMax {
			Buffer = append(Buffer, s)
		}
	}
}

// Greet builds a salutation from the realm's slogan. No params.
func Greet() string {
	return "hello, " + Slogan
}

// Stats returns a tuple — count, score and the owner's name. Tests
// multi-return rendering in the action page.
func Stats() (count, score int, name string) {
	return Counter, Score, Owner.Name
}

// VariadicSum sums an arbitrary number of integers. Variadic params
// exercise the `nums ...int` formatter in signatures.
func VariadicSum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Render produces the markdown shown on the realm root page. The
// state explorer is reachable via the State tab.
func Render(path string) string {
	out := "# State Showcase\n\n"
	out += "This realm exercises every kind of top-level declaration\n"
	out += "the state explorer can render. Open `/r/demo/stateshowcase$state`.\n\n"
	out += ufmt.Sprintf("- Counter: %d\n", Counter)
	out += ufmt.Sprintf("- Score:   %d\n", Score)
	out += ufmt.Sprintf("- Tags:    %s\n", strconv.Itoa(len(Tags)))
	return out
}
