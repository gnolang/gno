package mathml

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// maxCellSpan caps \multirow and \multicolumn spans. A span sets the
// minimum size of a stretched arrow, so an unbounded one draws a glyph
// thousands of em tall over the rest of the page.
const maxCellSpan = 64

// texLength matches the length of a \raisebox, \kern, \hspace or \\[len]:
// a signed decimal and an optional unit, which TeX lets a space precede.
var texLength = regexp.MustCompile(`^([+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)) *([a-zA-Z]{2})?$`)

// emPt is the size of an em in TeX points, at TeX's default 10pt font.
const emPt = 10

// texUnits gives the size in TeX points of each unit a length accepts:
// TeX's own units, and px, which CSS knows. The font-relative ones assume
// TeX's default 10pt font. css marks the units a browser understands; the
// others are converted to em.
var texUnits = map[string]struct {
	pt  float64
	css bool
}{
	"pt": {1, true},
	"pc": {12, true},
	"in": {72.27, true},
	"cm": {72.27 / 2.54, true},
	"mm": {72.27 / 25.4, true},
	"px": {72.27 / 96, true},
	"em": {emPt, true},
	"ex": {5, true}, // CSS's fallback x-height of 0.5em
	"bp": {72.27 / 72, false},
	"dd": {1238.0 / 1157, false},
	"cc": {12 * 1238.0 / 1157, false},
	"sp": {1.0 / 65536, false},
	"mu": {10.0 / 18, false},
}

// maxRaisePt is the largest \raisebox shift accepted, 2em of the math
// around the formula, counting the shifts of the enclosing \raisebox
// commands and the size switches in between. Larger shifts would let math
// move over the page around it, so they are ignored.
const maxRaisePt = 2 * emPt

// maxRowSpacingPt bounds the space a \\[len] puts after a table row: a
// negative space would draw the rows over each other and the page above,
// and a large one would stretch a short formula down the page.
const maxRowSpacingPt = 2 * emPt

// minKernPt and maxKernPt bound a \kern, \mkern or \hspace: a negative
// space could pull the math over the text before it, as a large shift
// would.
const (
	minKernPt = -2 * emPt
	maxKernPt = 20 * emPt
)

// fontScale returns the cumulative scale of the enclosing size switches.
func (converter *MathMLConverter) fontScale() float64 {
	if converter.sizeScale == 0 {
		return 1
	}
	return converter.sizeScale
}

// parseLength returns s, a TeX length, as a CSS length, and its size in
// points of the math around the formula: a length in em, or in a unit
// written as em, grows with the enclosing size switches. A bare number is
// taken in em.
func (converter *MathMLConverter) parseLength(s string) (string, float64, bool) {
	m := texLength.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return "", 0, false
	}
	unit := strings.ToLower(m[2])
	if unit == "" {
		unit = "em"
	}
	u, ok := texUnits[unit]
	if !ok {
		return "", 0, false
	}
	// The regexp only admits decimals, so ParseFloat cannot fail.
	v, _ := strconv.ParseFloat(m[1], 64)
	pt := v * u.pt
	num := strings.TrimPrefix(m[1], "+")
	if !u.css {
		em := strconv.FormatFloat(v*u.pt/emPt, 'f', 4, 64)
		num, unit = strings.TrimSuffix(strings.TrimRight(em, "0"), "."), "em"
	}
	if unit == "em" || unit == "ex" {
		pt *= converter.fontScale()
	}
	return num + unit, pt, true
}

func cmd_multirow(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 3 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	var attr string
	if name == "multirow" {
		attr = "rowspan"
	} else {
		attr = "columnspan"
	}
	n := converter.parseArg(args[2], ctx) // #nosec G602 - bounds checked above
	span := strings.TrimSpace(StringifyTokens(args[0].Expr))
	if v, err := strconv.Atoi(span); err == nil && v >= 1 && v <= maxCellSpan {
		n.SetAttr(attr, strconv.Itoa(v))
	}
	return n
}

func cmd_prescript(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 3 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	super := args[0]
	sub := args[1]
	base := args[2]
	multi := NewMMLNode("mmultiscripts")
	multi.AppendChild(converter.parseArg(base, ctx))
	multi.AppendChild(NewMMLNode("none"), NewMMLNode("none"), NewMMLNode("mprescripts"))
	temp := converter.ParseTex(sub, ctx)
	if temp != nil {
		multi.AppendChild(temp)
	}
	temp = converter.ParseTex(super, ctx)
	if temp != nil {
		multi.AppendChild(temp)
	}
	return multi
}

func cmd_sideset(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 3 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	left := args[0]
	right := args[1]
	base := args[2]
	multi := NewMMLNode("mmultiscripts")
	multi.Properties |= propLimitsunderover
	multi.AppendChild(converter.parseArg(base, ctx))
	getScripts := func(side *TokenBuffer) []*MMLNode {
		subscripts := make([]*MMLNode, 0)
		superscripts := make([]*MMLNode, 0)
		var last string
		for !side.Empty() {
			t, err := side.GetNextToken()
			if errors.Is(err, ErrTokenBufferExpr) {
				// A brace group where a ^ or _ is expected is not a script:
				// skip it. GetNextToken does not advance past a group, so
				// it must be consumed here or the loop never ends.
				if _, err := side.GetNextExpr(); err != nil {
					break
				}
				continue
			}
			if err != nil {
				break
			}
			switch t.Value {
			case "^":
				if last == t.Value {
					subscripts = append(subscripts, NewMMLNode("none"))
				}
				expr, err := side.GetNextExpr()
				if err != nil {
					expr, _ = side.GetNextN(1, true)
				}
				superscripts = append(superscripts, converter.parseArg(expr, ctx))
				last = t.Value
			case "_":
				if last == t.Value {
					superscripts = append(superscripts, NewMMLNode("none"))
				}
				expr, err := side.GetNextExpr()
				if err != nil {
					expr, _ = side.GetNextN(1, true)
				}
				subscripts = append(subscripts, converter.parseArg(expr, ctx))
				last = t.Value
			}
		}
		// Pad the shorter list, as after a repeated script (_a_b) or on
		// a side with no script.
		for len(superscripts) < max(len(subscripts), 1) {
			superscripts = append(superscripts, NewMMLNode("none"))
		}
		for len(subscripts) < len(superscripts) {
			subscripts = append(subscripts, NewMMLNode("none"))
		}
		result := make([]*MMLNode, len(subscripts)+len(superscripts))
		for i := range len(subscripts) {
			result[2*i] = subscripts[i]
			result[2*i+1] = superscripts[i]
		}
		return result
	}
	multi.AppendChild(getScripts(right)...)
	multi.AppendChild(NewMMLNode("mprescripts"))
	multi.AppendChild(getScripts(left)...)
	return multi
}

// themeColors maps the colour names \color and \textcolor accept to the
// suffix of the class the stylesheet colours: only theme colours, which
// keep their contrast in light and dark mode. Other colours, as names, hex
// or rgb values, are ignored and the content keeps the text colour.
var themeColors = map[string]string{
	"red":    "red",
	"blue":   "blue",
	"green":  "green",
	"orange": "orange",
	"purple": "purple",
	"gray":   "gray",
	"grey":   "gray",
}

// setColor gives n the class of the theme colour named by tex, if any.
func setColor(n *MMLNode, tex []Token) {
	if c, ok := themeColors[strings.ToLower(strings.TrimSpace(StringifyTokens(tex)))]; ok {
		n.SetAttr("class", "math-color-"+c)
	}
}

func cmd_textcolor(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 2 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	n := NewMMLNode("mstyle")
	setColor(n, args[0].Expr)
	converter.ParseTex(args[1], ctx, n)
	return n
}

func cmd_undersetOverset(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 2 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	base := converter.parseArg(args[1], ctx)
	embellishment := converter.parseArg(args[0], ctx)
	if base.Tag == "mo" {
		base.SetTrue("stretchy")
	}
	tag := "munder"
	if name == "overset" {
		tag = "mover"
	}
	underover := NewMMLNode(tag)
	underover.AppendChild(base, embellishment)
	n := NewMMLNode("mrow")
	n.AppendChild(underover)
	return n
}

func cmd_class(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 2 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	// The class name is ignored: letting page authors apply arbitrary site
	// CSS classes would allow restyling (e.g. overlaying) the page.
	return converter.ParseTex(args[1], ctx)
}

func cmd_raisebox(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 2 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	n := NewMMLNode("mpadded")
	outer := converter.raisePt
	if v, pt, ok := converter.parseLength(StringifyTokens(args[0].Expr)); ok && math.Abs(outer+pt) <= maxRaisePt {
		n.SetAttr("voffset", v)
		converter.raisePt += pt
		defer func() { converter.raisePt = outer }()
	}
	converter.ParseTex(args[1], ctx, n)
	return n
}

func cmd_cancel(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	var notation string
	switch name {
	case "cancel":
		notation = "updiagonalstrike"
	case "bcancel":
		notation = "downdiagonalstrike"
	case "xcancel":
		notation = "updiagonalstrike downdiagonalstrike"
	}

	n := NewMMLNode("menclose")
	n.SetAttr("notation", notation)
	converter.ParseTex(args[0], ctx, n)
	return n
}

func cmd_mathop(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	n := converter.operator(args[0], ctx)
	n.Properties |= propLimitsunderover | propMovablelimits
	if n.Tag == "mo" {
		n.SetAttr("rspace", "0")
	}
	return n
}

// operator parses b, the argument of \mathop or \operatorname, as one
// operator. MathML Core honours movablelimits and the operator spacing on
// <mo> only, so an argument that is only text (a word, a symbol, under a
// font command or not) is written as one <mo>; anything else is kept as
// parsed.
func (converter *MathMLConverter) operator(b *TokenBuffer, ctx parseContext) *MMLNode {
	n := converter.parseArg(b, ctx)
	if n.Tag == "mo" {
		return n
	}
	var sb strings.Builder
	var leafText func(*MMLNode) bool
	leafText = func(n *MMLNode) bool {
		switch n.Tag {
		case "mi", "mn", "mo", "mtext":
			sb.WriteString(n.Text)
		case "mspace":
			sb.WriteRune('\u2009') // thin space
		case "mrow", "mpadded", "mstyle":
			for _, c := range n.Children {
				if c != nil && !leafText(c) {
					return false
				}
			}
		default:
			return false
		}
		return true
	}
	if !leafText(n) || sb.Len() == 0 {
		return n
	}
	return NewMMLNode("mo", sb.String())
}

func cmd_mod(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	return NewMMLNode("mrow").AppendChild(
		NewMMLNode("mspace").SetAttr("width", "0.7em"),
		NewMMLNode("mo", "("),
		NewMMLNode("mo", "mod").SetAttr("lspace", "0"),
		converter.parseArg(args[0], ctx),
		NewMMLNode("mo", ")"),
	)
}

func cmd_substack(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	n := converter.ParseTex(args[0], ctx|ctxTable, NewMMLNode("mrow"))
	if n == nil {
		n = NewMMLNode("mrow")
	}
	processTable(n, name)
	n.SetAttr("rowspacing", "0")
	n.SetFalse("displaystyle")
	return n
}

func cmd_underOverBrace(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	annotation := converter.parseArg(args[0], ctx)
	n := NewMMLNode()
	brace := NewMMLNode("mo")
	brace.SetTrue("stretchy")
	n.Properties |= propLimitsunderover
	switch name {
	case "overbrace":
		n.Tag = "mover"
		brace.Text = "⏞"
	case "underbrace":
		n.Tag = "munder"
		brace.Text = "⏟"
	}
	n.AppendChild(annotation, brace)
	return n
}

func cmd_not(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	if len(args[0].Expr) < 1 {
		return NewMMLNode("merror", name).SetAttr("title", " requires an argument")
	} else if len(args[0].Expr) == 1 {
		t := args[0].Expr[0]
		sym, ok := symbolTable[t.Value]
		n := NewMMLNode()
		if ok {
			n.Text = sym.char
		} else {
			n.Text = t.Value
		}
		if sym.kind == sym_alphabetic || (len(t.Value) == 1 && unicode.IsLetter([]rune(t.Value)[0])) {
			n.Tag = "mi"
		} else {
			n.Tag = "mo"
		}
		if neg, ok := negation_map[t.Value]; ok {
			n.Text = neg
		} else {
			n.Text += "̸" //Once again we have chrome to thank for not implementing menclose
		}
		return n
	} else {
		n := NewMMLNode("menclose")
		n.SetAttr("notation", "updiagonalstrike")
		converter.ParseTex(args[0], ctx, n)
		return n
	}
}

func cmd_sqrt(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	n := NewMMLNode("msqrt")
	n.AppendChild(converter.parseArg(args[0], ctx))
	if opt != nil && !opt.Empty() {
		n.Tag = "mroot"
		n.AppendChild(converter.parseArg(opt, ctx))
	}
	return n
}

func cmd_text(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	return NewMMLNode("mtext", stringifyTokensHtml(args[0].Expr))
}

func cmd_frac(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 2 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	return makeFraction(name, converter.parseArg(args[0], ctx), converter.parseArg(args[1], ctx))
}

// makeFraction returns the fraction of command name (frac, binom...).
func makeFraction(name string, numerator, denominator *MMLNode) *MMLNode {
	frac := NewMMLNode("mfrac").AppendChild(numerator, denominator)
	switch name {
	case "cfrac", "dfrac":
		frac.SetTrue("displaystyle")
	case "tfrac":
		frac.SetFalse("displaystyle")
	case "binom", "tbinom":
		// A binomial coefficient is a fraction with no bar, in
		// parentheses: the fraction goes in an mrow with them.
		frac.SetAttr("linethickness", "0")
		wrapper := NewMMLNode("mrow").AppendChild(strechyOP("("), frac, strechyOP(")"))
		if name == "tbinom" {
			wrapper.SetFalse("displaystyle")
		}
		return wrapper
	}
	return frac
}

// parseArg parses the argument b of a command. An empty argument is an
// empty mrow, as an empty group is an empty atom in TeX: elements with a
// fixed number of children (mfrac, mover...) keep all of them.
func (converter *MathMLConverter) parseArg(b *TokenBuffer, ctx parseContext) *MMLNode {
	if n := converter.ParseTex(b, ctx); n != nil {
		return n
	}
	return NewMMLNode("mrow")
}

func cmd_operatorname(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	// Spaced as \sin is; the starred form puts its scripts under and over,
	// as \lim does.
	n := converter.operator(args[0], withVariant(ctx, ctxVarNormal))
	if n.Tag == "mo" {
		n.Properties |= propOperatorName
		n.SetAttr("lspace", "0").SetAttr("rspace", "0")
	}
	if star {
		n.Properties |= propLimitsunderover | propMovablelimits
	}
	return n
}

func cmd_phantom(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	n := NewMMLNode("mphantom").AppendChild(converter.parseArg(args[0], ctx))
	switch name {
	case "hphantom":
		return NewMMLNode("mpadded").SetAttr("height", "0").SetAttr("depth", "0").AppendChild(n)
	case "vphantom":
		return NewMMLNode("mpadded").SetAttr("width", "0").AppendChild(n)
	}
	return n
}

// makeSpace returns the space of a \kern, \mkern or \hspace of length s,
// or nil if s is not a length within [minKernPt, maxKernPt].
func (converter *MathMLConverter) makeSpace(s string) *MMLNode {
	v, pt, ok := converter.parseLength(s)
	if !ok || pt < minKernPt || pt > maxKernPt {
		return nil
	}
	return NewMMLNode("mspace").SetAttr("width", v)
}

func cmd_hspace(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if len(args) < 1 {
		return NewMMLNode("mtext", "Error: insufficient arguments")
	}
	if n := converter.makeSpace(StringifyTokens(args[0].Expr)); n != nil {
		return n
	}
	return NewMMLNode("merror", `\`+name).SetAttr("title", "invalid length")
}

// cmd_tag keeps the tag of the formula, which render writes after it: there
// is no equation numbering, so the tag is shown as written. \label is read
// and dropped.
func cmd_tag(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	if name == "tag" && len(args) == 1 {
		text := stringifyTokensHtml(args[0].Expr)
		if !star {
			text = "(" + text + ")"
		}
		converter.tag = NewMMLNode("mtext", text)
	}
	return NewMMLNode().SetProps(propNonprint)
}
