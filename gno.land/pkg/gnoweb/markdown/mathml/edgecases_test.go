package mathml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMMLNodeExtraArgs(t *testing.T) {
	n := NewMMLNode("mi", "x", "ignored")
	assert.Equal(t, "mi", n.Tag)
	assert.Equal(t, "x", n.Text)
}

func TestDepthErrorIsReported(t *testing.T) {
	_, err := NewMathMLConverter().ConvertInline(strings.Repeat("{", 100) + "x" + strings.Repeat("}", 100))
	assert.ErrorIs(t, err, errMaxDepth)
}

func TestBraceErrorHasNoMarkup(t *testing.T) {
	_, err := NewMathMLConverter().ConvertInline(`{<b>x`)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "<pre>")
}
