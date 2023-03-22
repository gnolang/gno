package rand

// Crypto symbol.
func Crypto() {}

// A Rand is a test symbol for structs.
type Rand struct {
	Name     string   // comment1
	Value    int      // comment2
	Attempts bool     // comment3
	unexp    chan int // comment4
}

// NewRand generates a new Rand.
func NewRand() *Rand {
	return nil
}

// Flag is tested for constant doc.
type Flag int

// Common flag values.
const (
	FlagA = 1 << iota
	FlagB
	FlagC
)

var unexp = 1
