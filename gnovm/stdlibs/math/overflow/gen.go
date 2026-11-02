//go:build ignore

//go:generate go run gen.go

// Code generator for math/overflow.
package main

import (
	"os"
	"strconv"
	"text/template"

	_ "embed"
)

func main() {
	var types, types256 []Type
	for _, bs := range [...]int{0, 8, 16, 32, 64, 256} {
		var s string
		if bs != 0 {
			s = strconv.Itoa(bs)
		} else {
			bs = 64
		}
		pair := []Type{
			{
				Name:   "int" + s,
				Short:  s,
				Signed: true,
				Min:    minOf(bs),
			},
			{
				Name:  "uint" + s,
				Short: "u" + s,
			},
		}
		if bs == 256 {
			types256 = append(types256, pair...)
		} else {
			types = append(types, pair...)
		}
	}
	// The 256-bit helpers are in a file of their own, so that tools that
	// turn Gno into Go, which has no 256-bit integers, can leave that file
	// out whole.
	generate("overflow_generated.gno", "", types)
	generate("overflow256_generated.gno", header256, types256)
}

const header256 = `
// minInt256 is the smallest int256. An int256 constant must fit an int64, so
// it is a variable, converted from uint256 bits.
var minInt256 = int256(minInt256Bits)

var minInt256Bits uint256 = 1 << 255
`

func generate(name, header string, types []Type) {
	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := parsedTmpl.Execute(file, struct {
		Header string
		Types  []Type
	}{header, types}); err != nil {
		panic(err)
	}
}

type Type struct {
	Name   string
	Short  string
	Signed bool
	Min    string // the smallest value, as a Gno constant expression
}

func minOf(bits int) string {
	if bits > 64 {
		// An int256 constant must fit an int64: see header256.
		return "minInt256"
	}
	return strconv.FormatInt(-1<<(bits-1), 10)
}

var parsedTmpl = template.Must(template.New("").Parse(tmpl))

//go:embed template.go.tpl
var tmpl string
