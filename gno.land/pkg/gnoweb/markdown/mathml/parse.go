package mathml

import (
	"errors"
	"fmt"
	"slices"
)

type NodeClass uint64
type NodeProperties uint64
type parseContext uint64

const (
	propNull NodeProperties = 1 << iota
	propNonprint
	propLargeop
	propSuperscript
	propSubscript
	propMovablelimits
	propLimitsunderover
	propCellSep
	propRowSep
	propLimits
	propNolimits
	propSymUpright
	propStretchy
	propHorzArrow
	propVertArrow
	propInfixOver
	propInfixChoose
	propInfixAtop
	propOperatorName

	propInfix = propInfixOver | propInfixChoose | propInfixAtop
)

const (
	ctxRoot parseContext = 1 << iota
	ctxDisplay
	ctxInline
	ctxScript
	ctxScriptscript
	ctxText
	// SIZES (interpreted as a 4-bit unsigned int)
	ctxSize_1
	ctxSize_2
	ctxSize_3
	ctxSize_4
	// ENVIRONMENTS
	ctxTable
	ctxEnvHasArg // the environment takes a {spec}, after an optional [position]
	ctxEnvHasOpt // the environment takes an optional [spec]
	// ONLY FONT VARIANTS AFTER THIS POINT
	ctxVarNormal
	ctxVarBb
	ctxVarMono
	ctxVarScriptChancery
	ctxVarScriptRoundhand
	ctxVarFrak
	ctxVarBold
	ctxVarItalic
	ctxVarSans
)

var (
	self_closing_tags = map[string]bool{
		"malignmark":  true,
		"maligngroup": true,
		"mspace":      true,
		"mprescripts": true,
		"none":        true,
	}
)

func (converter *MathMLConverter) OriginalString(b *TokenBuffer) string {
	if b.Empty() {
		return ""
	}
	start := b.Expr[0].start
	end := b.Expr[len(b.Expr)-1].end
	return string(converter.currentExpr[start:end])
}

// Parse a list of TeX tokens into a MathML node tree
func (converter *MathMLConverter) ParseTex(b *TokenBuffer, context parseContext, parent ...*MMLNode) *MMLNode {
	// Bound recursion so adversarial nesting cannot blow up time, memory or
	// output size. The panic is recovered by the converter entry points.
	converter.depth++
	defer func() { converter.depth-- }()
	if converter.depth > MaxParseDepth {
		panic(errMaxDepth)
	}
	var node *MMLNode
	siblings := make([]*MMLNode, 0)
	var optionString string
	if context&(ctxEnvHasArg|ctxEnvHasOpt) > 0 {
		opt, err := b.GetOptions()
		if err == nil && context&ctxEnvHasOpt > 0 {
			optionString = StringifyTokens(opt.Expr)
		} else if spec, err := b.GetNextExpr(); err == nil {
			optionString = StringifyTokens(spec.Expr)
		}
		context &^= ctxEnvHasArg | ctxEnvHasOpt
	}
	doFence := func(tok Token) *MMLNode {
		var n *MMLNode
		if tok.Kind&tokCommand > 0 {
			n = converter.ProcessCommand(context&^ctxRoot, tok, b)
		} else {
			n = NewMMLNode("mo")
			n.Text = tok.Value
		}
		if tok.Kind&tokOpen == tokOpen {
			n.SetAttr("form", "prefix")
		}
		if tok.Kind&tokMiddle == tokMiddle {
			n.SetAttr("form", "infix")
		}
		if tok.Kind&tokClose == tokClose {
			n.SetAttr("form", "postfix")
		}
		n.SetTrue("fence")
		n.SetTrue("stretchy")
		return n
	}
	// properties granted by a previous node
	var promotedProperties NodeProperties
	for !b.Empty() {
		var child *MMLNode
		tok, err := b.GetNextToken()
		if errors.Is(err, ErrTokenBufferEnd) {
			siblings = append(siblings, nil)
			promotedProperties = 0
			continue
		}
		if errors.Is(err, ErrTokenBufferExpr) {
			expr, _ := b.GetNextExpr()
			temp := converter.ParseTex(expr, context&^ctxRoot)
			if temp == nil {
				// {} is an empty atom, which can carry a script.
				temp = NewMMLNode("mrow")
			}
			temp.Properties |= promotedProperties
			siblings = append(siblings, temp)
			promotedProperties = 0
			continue
		}
		if context&ctxTable > 0 {
			switch tok.Value {
			case "&":
				// Do not count an escaped \& command
				if tok.Kind&tokReserved > 0 {
					child = NewMMLNode()
					child.Properties = propCellSep
					siblings = append(siblings, child)
					continue
				}
			case "\\", "cr":
				child = NewMMLNode()
				child.Properties = propRowSep
				option, err := b.GetOptions()
				if err == nil {
					dummy := NewMMLNode("rowspacing")
					dummy.Properties = propNonprint
					dummy.SetAttr("rowspacing", StringifyTokens(option.Expr))
					siblings = append(siblings, dummy)
				}
				siblings = append(siblings, child)
				continue
			}
		}
		switch {
		case tok.Kind&(tokClose|tokCurly) == tokClose|tokCurly:
			continue
		case tok.Kind&(tokClose|tokEnv) == tokClose|tokEnv:
			continue
		case tok.Kind&tokComment > 0:
			continue
		case tok.Kind&(tokSubsup|tokInfix) > 0:
			switch tok.Value {
			case "^":
				promotedProperties |= propSuperscript
				// handle the case where no base for the superscript is given
				if len(siblings) == 0 {
					siblings = append(siblings, nil)
				}
			case "_":
				promotedProperties |= propSubscript
				if len(siblings) == 0 {
					siblings = append(siblings, nil)
				}
			case "over":
				promotedProperties |= propInfixOver
			case "choose":
				promotedProperties |= propInfixChoose
			case "atop":
				promotedProperties |= propInfixAtop
			}
			// tell the next sibling to be a super- or subscript
			continue
		case tok.Kind&tokMacroarg > 0:
			child = NewMMLNode("merror", "?"+tok.Value)
			child.SetAttr("title", "Unexpanded macro argument")
		case tok.Kind&tokEscaped > 0:
			child = NewMMLNode("mo", tok.Value)
			if tok.Kind&(tokOpen|tokClose|tokFence) > 0 {
				child.SetTrue("stretchy")
			}
		case tok.Kind&(tokOpen|tokEnv) == tokOpen|tokEnv:
			ctx := setEnvironmentContext(tok, context) &^ ctxRoot
			env, _ := b.GetNextN(tok.MatchOffset)
			// The body is parsed into an mrow of its own even when it is a
			// single node: processTable splits that mrow into cells.
			child = processEnv(converter.ParseTex(env, ctx, NewMMLNode("mrow")), tok.Value, ctx)
		case tok.Kind&tokOpen > 0:
			child = NewMMLNode("mo")
			if tok.Kind&tokCommand > 0 {
				child = converter.ProcessCommand(context&^ctxRoot, tok, b)
			} else {
				child.Text = tok.Value
			}
			child.SetAttr("form", "prefix")
			if tok.Kind&tokFence > 0 {
				child.SetTrue("fence")
				child.SetTrue("stretchy")
			} else {
				child.SetFalse("stretchy")
			}
			if tok.Kind&tokFence == tokFence {
				container := NewMMLNode("mrow")
				if tok.Kind&tokNullDelim == 0 {
					container.AppendChild(child)
				}
				temp, _ := b.GetNextN(tok.MatchOffset)
				converter.ParseTex(temp, context&^ctxRoot, container)
				siblings = append(siblings, container)
				//don't need to worry about promotedProperties here.
				continue
			}
		case tok.Kind&tokClose > 0:
			child = NewMMLNode("mo")
			if tok.Kind&tokCommand > 0 {
				child = converter.ProcessCommand(context&^ctxRoot, tok, b)
			} else {
				child.Text = tok.Value
			}
			child.SetAttr("form", "postfix")
			if tok.Kind&tokNullDelim > 0 {
				child = nil
				break
			}
			if tok.Kind&tokFence > 0 {
				child.SetTrue("fence")
				child.SetTrue("stretchy")
			} else {
				child.SetFalse("stretchy")
			}
		case tok.Kind&tokFence > 0:
			if tok.Kind&tokNullDelim > 0 {
				continue
			}
			child = doFence(tok)
		case tok.Kind&tokLetter > 0:
			child = NewMMLNode("mi", tok.Value)
			child.set_variants_from_context(context &^ ctxRoot)
		case tok.Kind&tokNumber > 0:
			child = NewMMLNode("mn", tok.Value)
			child.set_variants_from_context(context &^ ctxRoot)
		case tok.Kind&tokCommand > 0:
			child = converter.ProcessCommand(context&^ctxRoot, tok, b)
		case tok.Kind&tokWhitespace > 0:
			if context&ctxText > 0 {
				child = NewMMLNode("mspace", " ")
				child.Tok.Value = " "
				child.SetAttr("width", "1em")
				siblings = append(siblings, child)
				continue
			} else {
				continue
			}
		default:
			child = NewMMLNode("mo", tok.Value)
		}
		if child == nil {
			continue
		}
		child.Tok = tok
		switch k := tok.Kind & (tokBigness1 | tokBigness2 | tokBigness3 | tokBigness4); k {
		case tokBigness1:
			child.SetAttr("scriptlevel", "-1")
			child.SetFalse("stretchy")
		case tokBigness2:
			child.SetAttr("scriptlevel", "-2")
			child.SetFalse("stretchy")
		case tokBigness3:
			child.SetAttr("scriptlevel", "-3")
			child.SetFalse("stretchy")
		case tokBigness4:
			child.SetAttr("scriptlevel", "-4")
			child.SetFalse("stretchy")
		}
		if child.Tag == "mo" && child.Text == "|" && tok.Kind&tokFence > 0 {
			child.SetTrue("symmetric")
		}
		// apply properties granted by previous sibling, if any
		child.Properties |= promotedProperties
		promotedProperties = 0
		siblings = append(siblings, child)
	}
	if len(parent) > 0 && parent[0] != nil {
		node = parent[0]
		node.Children = append(node.Children, siblings...)

		if node.Tag == "" {
			node.Tag = "mrow"
		}
	} else if len(siblings) > 1 {
		node = NewMMLNode("mrow")
		node.Children = append(node.Children, siblings...)
	} else if len(siblings) == 1 {
		if siblings[0] == nil {
			return nil
		}
		// A lone \over still needs its fraction built here, or it would
		// take its operands from the group around this one.
		if context&ctxRoot == ctxRoot && !(siblings[0].Tag == "mrow" || siblings[0].Tag == "mtd") || siblings[0].Properties&propInfix > 0 {
			node = NewMMLNode("mrow")
			node.Children = append(node.Children, siblings...)
		} else {
			return siblings[0]
		}
	} else {
		return nil
	}
	if len(node.Children) == 0 && len(node.Text) == 0 {
		return nil
	}
	node.Option = optionString
	node.doPostProcess()
	return node
}

func (n *MMLNode) doPostProcess() {
	if n == nil {
		return
	}
	n.postProcessInfix()
	n.postProcessLimitSwitch()
	n.postProcessScripts()
	n.postProcessOperatorNames()
	n.postProcessSpace()
	n.postProcessChars()
	begin := 0
	for begin < len(n.Children)-1 && n.Children[begin] == nil {
		begin++
	}
	n.Children = n.Children[begin:]
}

// isSeparator reports whether n separates the cells or rows of a table:
// a & or \\, or the spacing marker of a \\[len].
func isSeparator(n *MMLNode) bool {
	return n != nil && (n.Properties&(propCellSep|propRowSep) > 0 || isRowSpacing(n))
}

// isRowSpacing reports whether n is the marker a \\[len] leaves before its
// row separator.
func isRowSpacing(n *MMLNode) bool {
	return n.Tag == "rowspacing" && n.Properties&propNonprint > 0
}

func (n *MMLNode) postProcessLimitSwitch() {
	var i int
	for i = 1; i < len(n.Children); i++ {
		child := n.Children[i]
		if child == nil || n.Children[i-1] == nil || isSeparator(n.Children[i-1]) {
			continue
		}
		if child.Properties&propLimits > 0 {
			n.Children[i-1].Properties |= propLimitsunderover
			n.Children[i-1].Properties &= ^propMovablelimits
			n.Children[i-1].SetFalse("movablelimits")
			placeholder := NewMMLNode()
			placeholder.Properties = propNonprint
			n.Children[i-1], n.Children[i] = placeholder, n.Children[i-1]
		} else if child.Properties&propNolimits > 0 {
			n.Children[i-1].Properties &= ^propLimitsunderover
			n.Children[i-1].Properties &= ^propMovablelimits
			placeholder := NewMMLNode()
			placeholder.Properties = propNonprint
			n.Children[i-1], n.Children[i] = placeholder, n.Children[i-1]
		}
	}
}

// embellishedCore returns the node whose spacing applies to n: n itself, or
// the base of n if n is a script or a one-child mrow around one.
func embellishedCore(n *MMLNode) *MMLNode {
	for n != nil {
		switch n.Tag {
		case "msub", "msup", "msubsup", "munder", "mover", "munderover":
			if len(n.Children) == 0 {
				return n
			}
			n = n.Children[0]
		case "mrow":
			var only *MMLNode
			for _, c := range n.Children {
				if c != nil && c.Properties&propNonprint == 0 {
					if only != nil {
						return n
					}
					only = c
				}
			}
			if only == nil {
				return n
			}
			n = only
		default:
			return n
		}
	}
	return nil
}

// postProcessOperatorNames sets the space TeX puts around an operator name
// (\sin, \lim...): a thin space on each side that faces an ordinary
// neighbour (a letter, a number, a fraction, a \left...\right group), and
// after the name if another name follows. Next to an operator, the
// operator's own spacing applies; before an opening bracket, as in \sin(x),
// TeX puts none. The names are <mo> elements, so the space is set as their
// lspace and rspace, which MathML Core honours on <mo> only.
func (n *MMLNode) postProcessOperatorNames() {
	// kinds holds the kind of each printed child, and cores the node whose
	// spacing applies to it.
	const (
		separator = iota
		operator
		name
		ordinary
	)
	var kinds []int
	var cores []*MMLNode
	hasName := false
	for _, c := range n.Children {
		if c == nil || c.Properties&propNonprint > 0 {
			continue
		}
		core := embellishedCore(c)
		k := ordinary
		switch {
		case c.Properties&(propCellSep|propRowSep) > 0:
			k = separator
		case core.Properties&propOperatorName > 0:
			k, hasName = name, true
		case core.Tag == "mo":
			k = operator
		}
		kinds = append(kinds, k)
		cores = append(cores, core)
	}
	if !hasName {
		return
	}
	for i, k := range kinds {
		if k != name {
			continue
		}
		lspace, rspace := "0", "0"
		if i > 0 && kinds[i-1] == ordinary {
			lspace = thinSpace
		}
		if i+1 < len(kinds) && kinds[i+1] >= name {
			rspace = thinSpace
		}
		cores[i].SetAttr("lspace", lspace)
		cores[i].SetAttr("rspace", rspace)
	}
}

// thinSpace is TeX's \, (3/18 em), the space around an operator name.
const thinSpace = "0.1667em"

func (n *MMLNode) postProcessSpace() {
	i := 0
	limit := len(n.Children)
	isSpace := func(c *MMLNode) bool {
		return c != nil && space_widths[c.Tok.Value] != 0 && c.Tok.Kind&tokCommand > 0
	}
	for ; i < limit; i++ {
		if !isSpace(n.Children[i]) {
			continue
		}
		j := i + 1
		width := space_widths[n.Children[i].Tok.Value]
		for j < limit && isSpace(n.Children[j]) && space_widths[n.Children[j].Tok.Value] > 0 {
			width += space_widths[n.Children[j].Tok.Value]
			n.Children[j] = nil
			j++
		}
		n.Children[i].SetAttr("width", fmt.Sprintf("%.2fem", float64(width)/18.0))
		i = j
	}
}

func (n *MMLNode) postProcessChars() {
	combinePrimes := func(idx int) int {
		children := n.Children
		var i, nillifyUpTo int
		count := 1
		nillifyUpTo = idx
		keepgoing := true
		for i = idx + 1; i < len(children) && keepgoing; i++ {
			if children[i] == nil {
				continue
			} else if children[i].Text == "'" && children[i].Tok.Kind != tokCommand {
				count++
				nillifyUpTo = i
			} else {
				keepgoing = false
			}
		}
		var temp rune
		text := make([]rune, 0, 1+(count/4))
		for count > 0 {
			switch count {
			case 1:
				temp = '′'
			case 2:
				temp = '″'
			case 3:
				temp = '‴'
			default:
				temp = '⁗'
			}
			count -= 4
			text = append(text, temp)
		}
		for _, primes := range text {
			n.Children[idx] = NewMMLNode("mo", string(primes))
			idx++
		}
		for i = idx; i <= nillifyUpTo; i++ {
			n.Children[i] = nil
		}
		return i
	}
	i := 0
	var child *MMLNode
	for i < len(n.Children) {
		child = n.Children[i]
		if child == nil {
			i++
			continue
		}
		switch child.Text {
		case "-":
			n.Children[i].Text = "−"
		case "'", "’", "ʹ":
			combinePrimes(i)
		}
		i++
	}
}

// Look for any ^ or _ among siblings and convert to a msub, msup, or msubsup
func (n *MMLNode) postProcessScripts() {
	for i := 0; i < len(n.Children); i++ {
		var base, super, sub *MMLNode
		child := n.Children[i]
		if child == nil {
			continue
		}
		if child.Properties&(propSubscript|propSuperscript) == 0 {
			continue
		}
		var hasSuper, hasSub, hasBoth bool
		var script, next *MMLNode
		skip := 0
		if i < len(n.Children)-1 {
			next = n.Children[i+1]
		}
		// A script at the start of a cell or row has no base: it must
		// not take the separator before it.
		if i > 0 && !isSeparator(n.Children[i-1]) {
			base = n.Children[i-1]
		}
		if child.Properties&propSubscript > 0 {
			hasSub = true
			sub = child
			skip++
			if next != nil && next.Properties&propSuperscript > 0 {
				hasBoth = true
				super = next
				skip++
			}
		} else if child.Properties&propSuperscript > 0 {
			hasSuper = true
			super = child
			skip++
			if next != nil && next.Properties&propSubscript > 0 {
				hasBoth = true
				sub = next
				skip++
			}
		}
		pos := i - 1 //we want to replace the base with our script node
		if base == nil {
			pos++ //there is no base so we have to replace the zeroth node
			base = NewMMLNode("none")
			skip-- // there is one less node to nillify
		}
		// munder and mover tags must be encapsulated in an mrow for firefox to correctly render strechy fences
		// surrounding them.
		needs_mrow := false
		if hasBoth {
			if base.Properties&propLimitsunderover > 0 {
				script = NewMMLNode("munderover")
				needs_mrow = true
			} else {
				script = NewMMLNode("msubsup")
			}
			script.Children = append(script.Children, base, sub, super)
		} else if hasSub {
			if base.Properties&propLimitsunderover > 0 {
				script = NewMMLNode("munder")
				needs_mrow = true
			} else {
				script = NewMMLNode("msub")
			}
			script.Children = append(script.Children, base, sub)
		} else if hasSuper {
			if base.Properties&propLimitsunderover > 0 {
				script = NewMMLNode("mover")
				needs_mrow = true
			} else {
				script = NewMMLNode("msup")
			}
			script.Children = append(script.Children, base, super)
		} else {
			continue
		}
		if needs_mrow {
			n.Children[pos] = NewMMLNode("mrow").AppendChild(script)
		} else {
			n.Children[pos] = script
		}
		for j := pos + 1; j <= skip+pos && j < len(n.Children); j++ {
			n.Children[j] = nil
		}
	}
}

// postProcessInfix builds the fraction of \over, \atop and \choose,
// which take everything before them in their group, or in their cell or row
// of a table, as numerator, and everything after as denominator.
func (n *MMLNode) postProcessInfix() {
	isInfix := func(c *MMLNode) bool { return c != nil && c.Properties&propInfix > 0 }
	if !slices.ContainsFunc(n.Children, isInfix) {
		return
	}
	out := make([]*MMLNode, 0, len(n.Children))
	start := 0
	for i := 0; i <= len(n.Children); i++ {
		if i < len(n.Children) && !isSeparator(n.Children[i]) {
			continue
		}
		seg := n.Children[start:i]
		k := slices.IndexFunc(seg, isInfix)
		if k < 0 {
			out = append(out, seg...)
		} else {
			infix := seg[k].Properties & propInfix
			// TeX rejects a second \over in the same group as ambiguous:
			// it is shown as an error around the fraction of the first.
			ambiguous := false
			for _, c := range seg[k:] {
				if isInfix(c) {
					ambiguous = ambiguous || c != seg[k]
					c.Properties &^= propInfix
				}
			}
			name := "frac"
			if infix&propInfixChoose > 0 {
				name = "binom"
			}
			frac := makeFraction(name, infixOperand(seg[:k]), infixOperand(seg[k:]))
			if infix&propInfixAtop > 0 {
				frac.SetAttr("linethickness", "0")
			}
			if ambiguous {
				frac = NewMMLNode("merror").SetAttr("title", "ambiguous fraction: add braces").AppendChild(frac)
			}
			out = append(out, frac)
		}
		if i < len(n.Children) {
			out = append(out, n.Children[i])
		}
		start = i + 1
	}
	n.Children = out
}

// infixOperand returns the nodes on one side of an \over as one node, post
// processed as a group of their own.
func infixOperand(nodes []*MMLNode) *MMLNode {
	row := NewMMLNode("mrow").AppendChild(nodes...)
	row.doPostProcess()
	row.Children = slices.DeleteFunc(row.Children, func(c *MMLNode) bool { return c == nil })
	if len(row.Children) == 1 {
		return row.Children[0]
	}
	return row
}
