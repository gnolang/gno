package rand

import (
	"io"
)

// Crypto symbol.
func Crypto() {}

func NewWriter() io.Writer {}

// A Rand is a test symbol for structs.
type Rand struct {
	Name     string   // comment1
	Value    int      // comment2
	Attempts bool     // comment3
	unexp    chan int // comment4
}

type Rander interface {
	Generate() // To generate
	// comment1
	unexp(x int) string
}

type RandEmbedder struct {
	A string
	Rand
}

// NewRand generates a new Rand.
func NewRand() *Rand {
	return nil
}

func (*Rand) Generate() {
}

// Flag is tested for constant doc.
type Flag int

// Common flag values.
const (
	FlagA Flag = 1 << iota
	FlagB
	FlagC
)

var FlagVar Flag = 9999

var ExportedVar = true

var unexp = 1

func SetUnexp(cur realm, val int) {
	unexp = val
}
