package mathml

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// MaxParseDepth is the maximum recursion depth of the TeX parser. Deeper
// expressions fail to convert.
const MaxParseDepth = 64

var errMaxDepth = errors.New("mathml: expression nested too deeply")

// MathMLConverter manages LaTeX to MathML conversion state
type MathMLConverter struct {
	currentExpr []rune  // the expression currently being evaluated
	depth       int     // current ParseTex recursion depth
	sizeScale   float64 // cumulative scale of the enclosing size switches; 0 means 1
	raisePt     float64 // cumulative shift of the enclosing \raisebox commands, in points
	switchDepth int     // nesting of the enclosing style switches
}

// NewMathMLConverter returns a converter. It keeps per-expression state, so
// it must not be shared across concurrent conversions.
func NewMathMLConverter() *MathMLConverter {
	return &MathMLConverter{}
}

func (converter *MathMLConverter) render(tex string, displaystyle bool) (result string, err error) {
	var ast *MMLNode
	var builder strings.Builder
	setStyle := func(math *MMLNode) {
		if displaystyle {
			math.SetAttr("display", "block")
			math.SetAttr("class", "math-displaystyle")
			math.SetAttr("displaystyle", "true")
		} else {
			math.SetAttr("display", "inline")
			math.SetAttr("class", "math-textstyle")
		}
	}
	defer func() {
		if r := recover(); r != nil {
			ast = makeMMLError()
			setStyle(ast)
			ast.Write(&builder)
			result = builder.String()
			if e, ok := r.(error); ok && errors.Is(e, errMaxDepth) {
				err = errMaxDepth
			} else {
				err = fmt.Errorf("MathML encountered an unexpected error")
			}
		}
	}()
	converter.currentExpr = []rune(strings.Clone(tex))
	tokens, err := tokenize(converter.currentExpr)
	if err != nil {
		return "", err
	}
	ctx := ctxRoot
	breaks := hasTopLevelLineBreak(tokens)
	if breaks {
		// \\ outside an environment breaks the line: the expression is
		// a one-column table, as in KaTeX.
		ctx |= ctxTable
	}
	root := converter.ParseTex(NewTokenBuffer(tokens), ctx)
	if breaks && root != nil {
		processTable(root, "")
		root.SetAttr("displaystyle", strconv.FormatBool(displaystyle))
	}
	ast = converter.wrapInMathTag(root, tex)
	ast.SetAttr("xmlns", "http://www.w3.org/1998/Math/MathML")
	setStyle(ast)
	// Write writes the MathML on one line: whitespace around inline math,
	// as in "($x$)", would show as spaces.
	ast.Write(&builder)
	if displaystyle {
		builder.WriteRune('\n')
	}
	return builder.String(), err
}

func (converter *MathMLConverter) wrapInMathTag(mrow *MMLNode, tex string) *MMLNode {
	node := NewMMLNode("math")
	semantics := node.AppendNew("semantics")
	if mrow != nil && mrow.Tag != "mrow" {
		root := semantics.AppendNew("mrow")
		root.AppendChild(mrow)
		root.doPostProcess()
	} else if mrow == nil {
		semantics.AppendNew("none")
	} else {
		semantics.AppendChild(mrow)
		semantics.doPostProcess()
	}
	annotation := NewMMLNode("annotation", tex)
	annotation.SetAttr("encoding", "application/x-tex")
	semantics.AppendChild(annotation)
	return node
}

// hasTopLevelLineBreak reports whether tokens hold a \\ outside any group.
func hasTopLevelLineBreak(tokens []Token) bool {
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if opensGroup(t) {
			i += t.MatchOffset
		} else if t.Kind&tokCommand > 0 && t.Value == `\` {
			return true
		}
	}
	return false
}

// ConvertToDisplay converts LaTeX to display MathML
func (converter *MathMLConverter) DisplayStyle(tex string) (string, error) {
	return converter.render(tex, true)
}

// ConvertToInline converts LaTeX to inline MathML
func (converter *MathMLConverter) TextStyle(tex string) (string, error) {
	return converter.render(tex, false)
}

// ConvertInline converts LaTeX to inline MathML
func (converter *MathMLConverter) ConvertInline(tex string) (string, error) {
	return converter.TextStyle(tex)
}

// ConvertDisplay converts LaTeX to display MathML
func (converter *MathMLConverter) ConvertDisplay(tex string) (string, error) {
	return converter.DisplayStyle(tex)
}
