package mathml

import (
	"errors"
	"fmt"
	"math/bits"
)

type CommandSpec struct {
	F    func(*MathMLConverter, string, bool, parseContext, []*TokenBuffer, *TokenBuffer) *MMLNode
	argc int
	optc int
}

var (
	// maps commands to number of expected arguments
	command_args map[string]CommandSpec
	// Special properties of any identifiers accessed via a \command
	command_identifiers = map[string]NodeProperties{
		"arccos":   0,
		"arcsin":   0,
		"arctan":   0,
		"cos":      0,
		"cosh":     0,
		"cot":      0,
		"coth":     0,
		"csc":      0,
		"deg":      0,
		"dim":      0,
		"exp":      0,
		"hom":      0,
		"ker":      0,
		"ln":       0,
		"lg":       0,
		"log":      0,
		"sec":      0,
		"sin":      0,
		"sinh":     0,
		"tan":      0,
		"tanh":     0,
		"det":      propMovablelimits | propLimitsunderover,
		"gcd":      propMovablelimits | propLimitsunderover,
		"inf":      propMovablelimits | propLimitsunderover,
		"lim":      propMovablelimits | propLimitsunderover,
		"max":      propMovablelimits | propLimitsunderover,
		"min":      propMovablelimits | propLimitsunderover,
		"Pr":       propMovablelimits | propLimitsunderover,
		"sup":      propMovablelimits | propLimitsunderover,
		"limits":   propLimits | propNonprint,
		"nolimits": propNolimits | propNonprint,
	}

	precompiled_commands = map[string]*MMLNode{
		"varinjlim":  NewMMLNode("munder").SetProps(propMovablelimits|propLimitsunderover).AppendChild(NewMMLNode("mo", "lim"), NewMMLNode("mo", "→").SetTrue("stretchy")),
		"varprojlim": NewMMLNode("munder").SetProps(propMovablelimits|propLimitsunderover).AppendChild(NewMMLNode("mo", "lim"), NewMMLNode("mo", "←").SetTrue("stretchy")),
		// The rule under or over lim is drawn by the stylesheet.
		"varliminf": NewMMLNode("mpadded").SetProps(propMovablelimits | propLimitsunderover).AppendChild(NewMMLNode("mo", "lim").SetAttr("class", "math-liminf")),
		"varlimsup": NewMMLNode("mpadded").SetProps(propMovablelimits | propLimitsunderover).AppendChild(NewMMLNode("mo", "lim").SetAttr("class", "math-limsup")),
	}

	math_variants = map[string]parseContext{
		"mathbb":     ctxVarBb,
		"mathbf":     ctxVarBold,
		"boldsymbol": ctxVarBold,
		"mathbfit":   ctxVarBold | ctxVarItalic,
		"mathcal":    ctxVarScriptChancery,
		"mathfrak":   ctxVarFrak,
		"mathit":     ctxVarItalic,
		"mathrm":     ctxVarNormal,
		"mathscr":    ctxVarScriptRoundhand,
		"mathsf":     ctxVarSans,
		"mathsfbf":   ctxVarSans | ctxVarBold,
		"mathsfbfsl": ctxVarSans | ctxVarBold | ctxVarItalic,
		"mathsfsl":   ctxVarSans | ctxVarItalic,
		"mathtt":     ctxVarMono,
	}
	ctxSizeOffset int = bits.TrailingZeros64(uint64(ctxSize_1))
	// TODO: Not really using context for switch commands
	switches = map[string]parseContext{
		"color":             0,
		"bf":                ctxVarBold,
		"em":                ctxVarItalic,
		"rm":                ctxVarNormal,
		"displaystyle":      ctxDisplay,
		"textstyle":         ctxInline,
		"scriptstyle":       ctxScript,
		"scriptscriptstyle": ctxScriptscript,
		"tiny":              1 << ctxSizeOffset,
		"scriptsize":        2 << ctxSizeOffset,
		"footnotesize":      3 << ctxSizeOffset,
		"small":             4 << ctxSizeOffset,
		"normalsize":        5 << ctxSizeOffset,
		"large":             6 << ctxSizeOffset,
		"Large":             7 << ctxSizeOffset,
		"LARGE":             8 << ctxSizeOffset,
		"huge":              9 << ctxSizeOffset,
		"Huge":              10 << ctxSizeOffset,
	}
	// sizeSwitches gives the scale of each size switch, relative to the
	// surrounding math.
	sizeSwitches = map[string]float64{
		"tiny":         0.5,
		"scriptsize":   0.7,
		"footnotesize": 0.8,
		"small":        0.9,
		"normalsize":   1,
		"large":        1.2,
		"Large":        1.44,
		"LARGE":        1.728,
		"huge":         2.074,
		"Huge":         2.488,
	}
	accents = map[string]rune{
		"acute":          0x00b4,
		"bar":            0x00af,
		"breve":          0x02d8,
		"u":              0x02d8,
		"check":          0x02c7,
		"dot":            0x02d9,
		"ddot":           0x0308,
		"dddot":          0x20db,
		"ddddot":         0x20dc,
		"invbreve":       0x0311,
		"grave":          0x0060,
		"hat":            0x005e,
		"mathring":       0x02da,
		"overleftarrow":  0x2190,
		"overline":       0x203e,
		"overrightarrow": 0x2192,
		"tilde":          0x007e,
		"vec":            0x20d7,
		"widehat":        0x005e,
		"widetilde":      0x0360,
	}
	accents_below = map[string]rune{
		"underline": 0x0332,
	}
)

func init() {
	command_args = map[string]CommandSpec{
		"multirow":    {F: cmd_multirow, argc: 3, optc: 0},
		"multicolumn": {F: cmd_multirow, argc: 3, optc: 0},
		"prescript":   {F: cmd_prescript, argc: 3, optc: 0},
		"sideset":     {F: cmd_sideset, argc: 3, optc: 0},
		"textcolor":   {F: cmd_textcolor, argc: 2, optc: 0},
		"frac":        {F: cmd_frac, argc: 2, optc: 0},
		"cfrac":       {F: cmd_frac, argc: 2, optc: 0},
		"binom":       {F: cmd_frac, argc: 2, optc: 0},
		"tbinom":      {F: cmd_frac, argc: 2, optc: 0},
		"dfrac":       {F: cmd_frac, argc: 2, optc: 0},
		"tfrac":       {F: cmd_frac, argc: 2, optc: 0},
		"overset":     {F: cmd_undersetOverset, argc: 2, optc: 0},
		"underset":    {F: cmd_undersetOverset, argc: 2, optc: 0},
		"class":       {F: cmd_class, argc: 2, optc: 0},
		"raisebox":    {F: cmd_raisebox, argc: 2, optc: 0},
		"cancel":      {F: cmd_cancel, argc: 1, optc: 0},
		"bcancel":     {F: cmd_cancel, argc: 1, optc: 0},
		"xcancel":     {F: cmd_cancel, argc: 1, optc: 0},
		"mathop":      {F: cmd_mathop, argc: 1, optc: 0},
		"bmod":        {F: cmd_mod, argc: 1, optc: 0},
		"pmod":        {F: cmd_mod, argc: 1, optc: 0},
		"substack":    {F: cmd_substack, argc: 1, optc: 0},
		"underbrace":  {F: cmd_underOverBrace, argc: 1, optc: 0},
		"overbrace":   {F: cmd_underOverBrace, argc: 1, optc: 0},
		"not":         {F: cmd_not, argc: 1, optc: 0},
		"sqrt":        {F: cmd_sqrt, argc: 1, optc: 1},
		"text":        {F: cmd_text, argc: 1, optc: 0},
	}
}

func isolateMathVariant(ctx parseContext) parseContext {
	return ctx & ^(ctxVarNormal - 1)
}

// isLaTeXLogo argument is true for \LaTeX and false for \TeX. The kerning
// that browsers do not apply from the MathML attributes alone comes from the
// stylesheet, through the math-latex-* and math-tex-* classes: the page CSP
// blocks inline styles.
func makeTexLogo(isLaTeXLogo bool) *MMLNode {
	mrow := NewMMLNode("mrow")
	if isLaTeXLogo {
		mrow.AppendNew("mtext", "L")
		mrow.AppendNew("mspace").SetAttr("class", "math-latex-kern-la")

		mpadded := mrow.AppendNew("mpadded").SetAttr("voffset", "0.2em").SetAttr("class", "math-latex-a")
		mstyle1 := mpadded.AppendNew("mstyle").SetAttr("scriptlevel", "0").SetAttr("displaystyle", "false")
		mstyle1.AppendNew("mtext", "A")

		mrow.AppendNew("mspace").SetAttr("width", "-0.15em").SetAttr("class", "math-latex-kern-at")
	}
	mrow.AppendNew("mtext", "T")
	mrow.AppendNew("mspace").SetAttr("width", "-0.1667em").SetAttr("class", "math-tex-kern-te")

	mpadded := mrow.AppendNew("mpadded").SetAttr("voffset", "-0.2155em").SetAttr("class", "math-tex-e")
	mstyle := mpadded.AppendNew("mstyle").SetAttr("scriptlevel", "0").SetAttr("displaystyle", "false")
	mstyle.AppendNew("mtext", "E")

	mrow.AppendNew("mspace").SetAttr("width", "-0.125em").SetAttr("class", "math-tex-kern-ex")
	mrow.AppendNew("mtext", "X")

	return mrow
}

// ProcessCommand sets the value of n and returns the next index of tokens to be processed.
func (converter *MathMLConverter) ProcessCommand(context parseContext, tok Token, b *TokenBuffer) *MMLNode {
	star := tok.Kind&tokStarSuffix > 0
	name := tok.Value
	switch name {
	case "newcommand", "def", "renewcommand":
		return converter.newCommand(b)
	case "LaTeX":
		return makeTexLogo(true)
	case "TeX":
		return makeTexLogo(false)
	}

	if prop, ok := command_identifiers[name]; ok {
		n := NewMMLNode("mi")
		n.Properties = prop
		n.Text = name
		n.SetAttr("lspace", "0.11111em")
		n.Tok = tok
		n.set_variants_from_context(context)
		n.setAttribsFromProperties()
		return n
	} else if sym, ok := symbolTable[name]; ok {
		return makeSymbol(sym, tok, context)
	}
	if node, ok := precompiled_commands[tok.Value]; ok {
		// We must wrap this node in a new mrow since all instances point to the same memory location. This way, we can
		// perform modifications on the newly created mrow without affecting all other instances of the precompiled
		// command.
		return NewMMLNode("mrow").AppendChild(node).SetProps(node.Properties)
	}
	if variant, ok := math_variants[name]; ok {
		nextExpr, err := b.GetNextExpr()
		if errors.Is(err, ErrTokenBufferSingle) {
			nextExpr, err = b.GetNextN(1, true)
		}
		var wrapper *MMLNode
		if name == "mathrm" {
			wrapper = NewMMLNode("mpadded").SetAttr("lspace", "0")
		}
		if err != nil {
			// treat the remainder of the buffer as argument
			return converter.ParseTex(b, context|variant, wrapper)
		}
		return converter.ParseTex(nextExpr, context|variant, wrapper)
	}
	if width, ok := space_widths[name]; ok {
		n := NewMMLNode("mspace")
		n.Tok = tok
		if name == `\` {
			n.SetAttr("linebreak", "newline")
		} else {
			n.SetAttr("width", fmt.Sprintf("%.7fem", float32(width)/18.0))
		}
		return n
	}
	if sw, ok := switches[name]; ok {
		cellEnd := func(t Token) bool {
			if t.Kind&tokReserved > 0 && t.Value == "&" {
				return true
			}
			if t.Value == "\\" || t.Value == "cr" {
				return true
			}
			return false
		}
		// The switch applies up to the end of the cell or row it is in. The
		// cells of a group after it ({...}, an environment, a \left...\right
		// pair) belong to that group.
		var i int
		for i = b.idx; i < len(b.Expr); i++ {
			t := b.Expr[i]
			if opensGroup(t) {
				i += t.MatchOffset
				continue
			}
			if cellEnd(t) {
				break
			}
		}
		switchExpressions, _ := b.GetNextN(min(i, len(b.Expr)) - b.idx)

		n := NewMMLNode("mstyle")
		if name == "color" {
			expr, err := switchExpressions.GetNextExpr()
			if err == nil {
				setColor(n, expr.Expr)
				converter.ParseTex(switchExpressions, context|sw, n)
				return n
			}
			b.Unget()
			return NewMMLNode("merror", name).SetAttr("title", fmt.Sprintf("%s expects an argument", name))
		}
		if size, ok := sizeSwitches[name]; ok {
			converter.parseSized(switchExpressions, context|sw, n, size)
			return n
		}
		converter.ParseTex(switchExpressions, context|sw, n)
		switch name {
		case "displaystyle":
			n.SetTrue("displaystyle")
			n.SetAttr("scriptlevel", "0")
		case "textstyle":
			n.SetFalse("displaystyle")
			n.SetAttr("scriptlevel", "0")
		case "scriptstyle":
			n.SetFalse("displaystyle")
			n.SetAttr("scriptlevel", "1")
		case "scriptscriptstyle":
			n.SetFalse("displaystyle")
			n.SetAttr("scriptlevel", "2")
		case "rm":
			n.SetAttr("mathvariant", "normal")
		}
		return n
	}
	var n *MMLNode
	if spec, ok := command_args[name]; ok {
		n = converter.processCommandArgs(context, name, star, b, spec)
	} else if ch, ok := accents[name]; ok {
		n = NewMMLNode("mover").SetTrue("accent")
		acc := NewMMLNode("mo", string(ch))
		acc.SetTrue("stretchy")
		tempbuf, err := b.GetNextExpr()
		if errors.Is(err, ErrTokenBufferSingle) {
			tempbuf, _ = b.GetNextN(1, true)
		}
		base := converter.ParseTex(tempbuf, context)
		if base.Tag == "mi" {
			// Dotless i and j under the accent, through the stylesheet.
			base.AddClass("math-dtls-on")
		}
		n.AppendChild(base, acc)
	} else if ch, ok := accents_below[name]; ok {
		n = NewMMLNode("munder").SetTrue("accent")
		acc := NewMMLNode("mo", string(ch))
		acc.SetTrue("stretchy")
		tempbuf, err := b.GetNextExpr()
		if errors.Is(err, ErrTokenBufferSingle) {
			tempbuf, _ = b.GetNextN(1, true)
		}
		base := converter.ParseTex(tempbuf, context)
		if base.Tag == "mi" {
			base.AddClass("math-dtls-on")
		}
		n.AppendChild(base, acc)
	} else {
		n = NewMMLNode("merror", tok.Value)
	}
	n.Tok = tok
	n.set_variants_from_context(context)
	n.setAttribsFromProperties()
	return n
}

// minSize and maxSize bound the size of math under nested size switches,
// relative to the math around it: a mathsize percentage is relative to the
// enclosing font, so each nested \Huge would otherwise multiply the size
// again, letting a short input draw glyphs over the whole page.
const (
	minSize = 0.5   // \tiny
	maxSize = 2.488 // \Huge
)

// parseSized parses b into n under a size switch of the given scale. The
// cumulative scale of the enclosing switches is clamped to [minSize,
// maxSize], and n gets the mathsize that reaches it from the enclosing one.
func (converter *MathMLConverter) parseSized(b *TokenBuffer, context parseContext, n *MMLNode, size float64) {
	outer := converter.sizeScale
	if outer == 0 {
		outer = 1
	}
	scale := min(max(outer*size, minSize), maxSize)
	n.SetAttr("mathsize", fmt.Sprintf("%05.1f%%", 100*scale/outer))
	converter.sizeScale = scale
	defer func() { converter.sizeScale = outer }()
	converter.ParseTex(b, context, n)
}

func makeSymbol(t symbol, tok Token, context parseContext) *MMLNode {
	n := NewMMLNode()
	n.Properties = t.properties
	n.Text = t.char
	if context&ctxTable > 0 && t.properties&(propHorzArrow|propVertArrow) > 0 {
		n.SetTrue("stretchy")
	}
	if n.Properties&propSymUpright > 0 {
		context |= ctxVarNormal
	}
	switch t.kind {
	case sym_binaryop, sym_opening, sym_closing, sym_relation, sym_operator:
		n.Tag = "mo"
	case sym_large:
		n.Tag = "mo"
		// we do an XOR rather than an OR here to remove this property
		// from any of the integral symbols from symbolTable.
		n.Properties ^= propLimitsunderover
		n.Properties |= propLargeop | propMovablelimits
	case sym_alphabetic:
		n.Tag = "mi"
	default:
		if tok.Kind&tokFence > 0 {
			n.Tag = "mo"
		} else {
			n.Tag = "mi"
		}
	}
	n.Tok = tok
	n.set_variants_from_context(context)
	n.setAttribsFromProperties()
	return n
}

// Process commands that take arguments
func (converter *MathMLConverter) processCommandArgs(context parseContext, name string, star bool, b *TokenBuffer, spec CommandSpec) *MMLNode {
	args := make([]*TokenBuffer, 0)
	if b.Empty() {
		return NewMMLNode("merror", name).SetAttr("title", name+" requires one or more arguments")
	}
	opt, _ := b.GetOptions()
	for !b.Empty() && len(args) < spec.argc {
		arg, err := b.GetNextExpr()
		if err == nil {
			args = append(args, arg)
		} else if errors.Is(err, ErrTokenBufferSingle) {
			arg, err := b.GetNextN(1, true)
			if err == nil {
				args = append(args, arg)
			}
		}
	}
	if len(args) != spec.argc {
		return NewMMLNode("merror", name).SetAttr("title", "wrong number of arguments")
	}
	return spec.F(converter, name, star, context, args, opt)
}

func (converter *MathMLConverter) newCommand(b *TokenBuffer) (errNode *MMLNode) {
	var definition *TokenBuffer
	var name string
	makeMerror := func(msg string) *MMLNode {
		n := NewMMLNode("merror", `\newcommand`)
		n.SetAttr("title", msg)
		return n
	}
	t, err := b.GetNextToken()
	if err == nil && t.Kind&tokCommand == 0 {
		errNode = makeMerror("newcommand expects an argument of exactly one \\command")
		return
	} else if errors.Is(err, ErrTokenBufferExpr) {
		temp, err := b.GetNextExpr()
		if len(temp.Expr) != 1 || err != nil {
			errNode = makeMerror("newcommand expects an argument of exactly one \\command")
			return
		} else if temp.Expr[0].Kind&tokCommand == 0 {
			errNode = makeMerror("newcommand expects an argument of exactly one \\command")
			return
		}
		t = temp.Expr[0]
	}
	name = t.Value

	definition, err = b.GetNextExpr()
	if errors.Is(err, ErrTokenBufferSingle) {
		definition, err = b.GetNextN(1, true)
	}
	if err != nil {
		errNode = makeMerror("malformed macro definition")
		return
	}
	for _, t := range definition.Expr {
		if t.Value == name && t.Kind&tokCommand > 0 {
			errNode = makeMerror("Recursive macro definition detected")
			return
		}
	}

	return
}
