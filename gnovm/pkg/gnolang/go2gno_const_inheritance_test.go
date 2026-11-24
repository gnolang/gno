package gnolang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// An oversized const block is refused by the checkCopies bound at type check,
// not by ParseFile: ParseFile also re-parses stored packages at boot, so a bound
// there would stop a node on code it accepted before.
func TestOversizedConstInheritanceRejectedAtTypeCheck(t *testing.T) {
	block := inheritedConstBlock(200, 2000)
	tests := []struct {
		name string
		src  string
	}{
		{"top level", "package grid\n" + block},
		{"function body", "package grid\nfunc f() {\n" + block + "}\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorContains(t, tcCopies(tt.src), "declarations copy")
			var m *Machine
			_, err := m.ParseFile("test.gno", tt.src)
			require.NoError(t, err, "ParseFile must accept what boot re-parses")
		})
	}
}

func TestParseFileAllowsLegitimateConstInheritance(t *testing.T) {
	tests := []struct {
		name string
		decl string
	}{
		{
			"iota enum",
			`const (
	Sunday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)`,
		},
		{
			"unit multipliers",
			`const (
	KB = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)`,
		},
		{
			"simple expression",
			`const (
	a = 1
	b
	c
	d
)`,
		},
		{
			"intervening explicit spec",
			`const (
	a = 1
	b
	c = 2
	d
	e
)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m *Machine
			_, err := m.ParseFile("test.gno", "package test\n"+tt.decl)
			require.NoError(t, err)
		})
	}
}

func inheritedConstBlock(terms, specs int) string {
	var b strings.Builder
	b.WriteString("const (\nc0 = ")
	b.WriteString(strings.Repeat("1 + ", terms-1))
	b.WriteString("1\n")
	for i := 1; i <= specs; i++ {
		fmt.Fprintf(&b, "c%d\n", i)
	}
	b.WriteString(")\n")
	return b.String()
}

func TestParseFileRejectsInheritedFuncLit(t *testing.T) {
	src := `package test
const (
	a = len([1]func(){func() { {}; {} }})
	b
)
`
	var m *Machine
	_, err := m.ParseFile("test.gno", src)
	require.ErrorContains(t, err, "const declaration inheritance of a function literal is not supported")

	// Written out explicitly, the same values are accepted.
	src = `package test
const (
	a = len([1]func(){func() { {}; {} }})
	b = len([1]func(){func() { {}; {} }})
)
`
	_, err = m.ParseFile("test.gno", src)
	require.NoError(t, err)
}

func TestParseExprConstInheritance(t *testing.T) {
	// ParseExpr discards the parser's FileSet, so the budget must not
	// depend on one.
	var m *Machine
	_, err := m.ParseExpr("func() int { const (a = iota; b); return b }()")
	require.NoError(t, err)

	_, err = m.ParseExpr("func() int {\n" + inheritedConstBlock(200, 2000) + "return 0\n}()")
	require.ErrorContains(t, err, "declarations copy")
}
