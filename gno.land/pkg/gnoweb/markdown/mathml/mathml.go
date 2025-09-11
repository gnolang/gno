package mathml

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func init() {
	logger = log.New(os.Stderr, "MathML: ", log.LstdFlags)
}

// TexToMML converts LaTeX to MathML
func TexToMML(tex string, macros map[string]string, block, displaystyle bool) (result string, err error) {
	var ast *MMLNode
	var builder strings.Builder
	defer func() {
		if r := recover(); r != nil {
			ast = makeMMLError()
			if block {
				ast.SetAttr("display", "block")
			} else {
				ast.SetAttr("display", "inline")
			}
			if displaystyle {
				ast.SetTrue("displaystyle")
			}
			ast.Write(&builder, 0)
			result = builder.String()
			err = fmt.Errorf("MathML encountered an unexpected error while processing\n%s\n", tex)
		}
	}()
	converter := NewMathMLConverter()
	converter.currentExpr = []rune(strings.Clone(tex))
	tokens, err := tokenize(converter.currentExpr)
	if err != nil {
		return "", err
	}
	ast = wrapInMathTag(converter.ParseTex(NewTokenBuffer(tokens), ctxRoot), tex)
	if block {
		ast.SetAttr("display", "block")
	} else {
		ast.SetAttr("display", "inline")
	}
	if displaystyle {
		ast.SetTrue("displaystyle")
	}
	ast.Write(&builder, 1)
	return builder.String(), err
}
func wrapInMathTag(mrow *MMLNode, tex string) *MMLNode {
	node := NewMMLNode("math")
	node.SetAttr("style", "font-feature-settings: 'dtls' off;").SetAttr("xmlns", "http://www.w3.org/1998/Math/MathML")
	semantics := node.AppendNew("semantics")
	if mrow != nil && mrow.Tag != "mrow" {
		root := semantics.AppendNew("mrow")
		root.AppendChild(mrow)
		root.doPostProcess()
	} else {
		semantics.AppendChild(mrow)
		semantics.doPostProcess()
	}
	annotation := NewMMLNode("annotation", strings.ReplaceAll(tex, "<", "&lt;"))
	annotation.SetAttr("encoding", "application/x-tex")
	semantics.AppendChild(annotation)
	return node
}

// DisplayStyle renders LaTeX as display MathML
func DisplayStyle(tex string, macros map[string]string) (string, error) {
	return TexToMML(tex, macros, true, false)
}

// InlineStyle renders LaTeX as inline MathML
func InlineStyle(tex string, macros map[string]string) (string, error) {
	return TexToMML(tex, macros, false, false)
}

// MathMLConverter manages LaTeX to MathML conversion state
type MathMLConverter struct {
	EQCount              int             // used for numbering display equations
	DoNumbering          bool            // Whether or not to number equations in a document
	PrintOneLine         bool            // If true, print the MathML on a single line
	currentExpr          []rune          // the expression currently being evaluated
	currentIsDisplay     bool            // true if the current expression is being rendered in displaystyle
	needMacroExpansion   map[string]bool // used if any \newcommand definitions are encountered.
	unknownCommandsAsOps bool            // treat unknown \commands as operators
}

// NewDocument creates a MathMLConverter for a document
func NewDocument(macros map[string]string, doNumbering bool) *MathMLConverter {
	converter := NewMathMLConverter(macros)
	converter.DoNumbering = doNumbering
	return converter
}

func NewMathMLConverter(macros ...map[string]string) *MathMLConverter {
	var out MathMLConverter
	out.needMacroExpansion = make(map[string]bool)
	return &out
}

func (converter *MathMLConverter) render(tex string, displaystyle bool) (result string, err error) {
	var ast *MMLNode
	var builder strings.Builder
	var indent int
	if converter.PrintOneLine {
		indent = -1
	}
	defer func() {
		if r := recover(); r != nil {
			ast = makeMMLError()
			if displaystyle {
				ast.SetAttr("display", "block")
				ast.SetAttr("class", "math-displaystyle")
				ast.SetAttr("displaystyle", "true")
			} else {
				ast.SetAttr("display", "inline")
				ast.SetAttr("class", "math-textstyle")
			}
			ast.Write(&builder, indent)
			result = builder.String()
			err = fmt.Errorf("MathML encountered an unexpected error")
		}
		converter.currentIsDisplay = false
	}()
	converter.currentExpr = []rune(strings.Clone(tex))
	tokens, err := tokenize(converter.currentExpr)
	if err != nil {
		return "", err
	}
	ast = converter.wrapInMathTag(converter.ParseTex(NewTokenBuffer(tokens), ctxRoot), tex)
	ast.SetAttr("xmlns", "http://www.w3.org/1998/Math/MathML")
	if displaystyle {
		ast.SetAttr("display", "block")
		ast.SetAttr("class", "math-displaystyle")
		ast.SetAttr("displaystyle", "true")
	} else {
		ast.SetAttr("display", "inline")
		ast.SetAttr("class", "math-textstyle")
	}
	builder.WriteRune('\n')
	ast.Write(&builder, indent)
	builder.WriteRune('\n')
	return builder.String(), err
}

func (converter *MathMLConverter) wrapInMathTag(mrow *MMLNode, tex string) *MMLNode {
	node := NewMMLNode("math")
	node.SetAttr("style", "font-feature-settings: 'dtls' off;")
	semantics := node.AppendNew("semantics")
	if converter.DoNumbering && converter.currentIsDisplay {
		converter.EQCount++
		numberedEQ := NewMMLNode("mtable")
		row := numberedEQ.AppendNew("mlabeledtr")
		num := row.AppendNew("mtd")
		eq := row.AppendNew("mtd")
		num.AppendNew("mtext", fmt.Sprintf("(%d)", converter.EQCount))
		if mrow != nil && mrow.Tag != "mrow" {
			root := NewMMLNode("mrow")
			root.AppendChild(mrow)
			root.doPostProcess()
			eq.AppendChild(root)
		} else {
			eq.AppendChild(mrow)
			eq.doPostProcess()
		}
		semantics.AppendChild(numberedEQ)
	} else {
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
	}
	annotation := NewMMLNode("annotation", strings.ReplaceAll(tex, "<", "&lt;"))
	annotation.SetAttr("encoding", "application/x-tex")
	semantics.AppendChild(annotation)
	return node
}

// ConvertToDisplay converts LaTeX to display MathML
func (converter *MathMLConverter) DisplayStyle(tex string) (string, error) {
	converter.currentIsDisplay = true
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

// ConvertToMathML converts LaTeX to MathML without semantics wrapper
func (converter *MathMLConverter) SemanticsOnly(tex string) (string, error) {
	converter.currentExpr = []rune(strings.Clone(tex))
	tokens, err := tokenize(converter.currentExpr)
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	if err != nil {
		return "", err
	}

	ast := converter.ParseTex(NewTokenBuffer(tokens), ctxRoot)
	var builder strings.Builder
	var indent int
	if converter.PrintOneLine {
		indent = -1
	}
	ast.Write(&builder, indent)
	return builder.String(), err
}
