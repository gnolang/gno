package press

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	t.Parallel()

	p := NewPress()
	assert.Equal(t, "", p.Print())
}

func TestBasic(t *testing.T) {
	t.Parallel()

	p := NewPress()
	p.P("this ")
	p.P("is ")
	p.P("a test")
	assert.Equal(t, "this is a test", p.Print())
}

func TestBasicLn(t *testing.T) {
	t.Parallel()

	p := NewPress()
	p.P("this ")
	p.P("is ")
	p.Pl("a test")
	assert.Equal(t, "this is a test\n", p.Print())
}

func TestNewlineStr(t *testing.T) {
	t.Parallel()

	p := NewPress().SetNewlineStr("\r\n")
	p.P("this ")
	p.P("is ")
	p.Pl("a test")
	p.Pl("a test")
	p.Pl("a test")
	assert.Equal(t, "this is a test\r\na test\r\na test\r\n", p.Print())
}

func TestIndent(t *testing.T) {
	t.Parallel()

	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.Pl("second line")
		p.Pl("third line")
	}).P("}")
	assert.Equal(t, `first line {
	second line
	third line
}`, p.Print())
}

func TestIndent2(t *testing.T) {
	t.Parallel()

	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.P("second ")
		p.P("line")
		// Regardless of whether Pl or Ln is called on cp2,
		// the indented lines terminate with newlineDelim before
		// the next unindented line.
	}).P("}")
	assert.Equal(t, `first line {
	second line
}`, p.Print())
}

func TestIndent3(t *testing.T) {
	t.Parallel()

	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.P("second ")
		p.Pl("line")
	}).P("}")
	assert.Equal(t, `first line {
	second line
}`, p.Print())
}

func TestIndentLn(t *testing.T) {
	t.Parallel()

	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.Pl("second line")
		p.Pl("third line")
	}).Pl("}")
	assert.Equal(t, `first line {
	second line
	third line
}
`, p.Print())
}

func TestNestedIndent(t *testing.T) {
	t.Parallel()

	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.Pl("second line")
		p.Pl("third line")
		p.I(func(p *Press) {
			p.Pl("fourth line")
			p.Pl("fifth line")
		})
	}).Pl("}")
	assert.Equal(t, `first line {
	second line
	third line
		fourth line
		fifth line
}
`, p.Print())
}
