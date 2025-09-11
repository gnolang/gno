package mathml

import (
	"errors"
	"fmt"
)

type NodeClass uint64
type NodeProperties uint64
type parseContext uint64

const (
	propNull NodeProperties = 1 << iota
	propNonprint
	propLargeop
	propScriptBase
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
)

const (
	ctxRoot parseContext = 1 << iota
	ctxDisplay
	ctxInline
	ctxScript
	ctxScriptscript
	ctxText
	ctxBracketed
	// SIZES (interpreted as a 4-bit unsigned int)
	ctxSize_1
	ctxSize_2
	ctxSize_3
	ctxSize_4
	// ENVIRONMENTS
	ctxTable
	ctxEnvHasArg
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
	var node *MMLNode
	siblings := make([]*MMLNode, 0)
	var optionString string
	if context&ctxEnvHasArg > 0 {
		_, err := b.GetNextToken()
		if errors.Is(err, ErrTokenBufferExpr) {
			temp, _ := b.GetNextExpr()
			optionString = StringifyTokens(temp.Expr)
		} else {
			b.Unget()
		}
		context ^= ctxEnvHasArg
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
			if temp != nil {
				temp.Properties |= promotedProperties
			}
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
		case tok.Kind&tokBadmacro > 0:
			child = NewMMLNode("merror", tok.Value)
			child.SetAttr("title", "cyclic dependency in macro definition")
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
			child = processEnv(converter.ParseTex(env, ctx), tok.Value, ctx)
		case tok.Kind&(tokOpen|tokCurly) == tokOpen|tokCurly:
			child = converter.ParseTex(b, context&^ctxRoot)
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
				if tok.Kind&tokNull == 0 {
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
			if tok.Kind&tokNull > 0 {
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

		if context&ctxRoot == ctxRoot && !(siblings[0].Tag == "mrow" || siblings[0].Tag == "mtd") {
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
	if n != nil {
		n.postProcessInfix()
		n.postProcessLimitSwitch()
		n.postProcessScripts()
		n.postProcessSpace()
		n.postProcessChars()
	}
	begin := 0
	for n.Children[begin] == nil && begin < len(n.Children)-1 {
		begin++
	}
	n.Children = n.Children[begin:]
}

func (n *MMLNode) postProcessLimitSwitch() {
	var i int
	for i = 1; i < len(n.Children); i++ {
		child := n.Children[i]
		if child == nil {
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

func (n *MMLNode) postProcessSpace() {
	i := 0
	limit := len(n.Children)
	for ; i < limit; i++ {
		if n.Children[i] == nil || space_widths[n.Children[i].Tok.Value] == 0 {
			continue
		}
		if n.Children[i].Tok.Kind&tokCommand == 0 {
			continue
		}
		j := i + 1
		width := space_widths[n.Children[i].Tok.Value]
		for j < limit && space_widths[n.Children[j].Tok.Value] > 0 && n.Children[j].Tok.Kind&tokCommand > 0 {
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
		case "<":
			n.Children[i].Text = "&lt;"
		case ">":
			n.Children[i].Text = "&gt;"
		case "&":
			n.Children[i].Text = "&amp;"
		case "'", "’", "ʹ":
			combinePrimes(i)
		}
		i++
	}
}

// Look for any ^ or _ among siblings and convert to a msub, msup, or msubsup
func (n *MMLNode) postProcessScripts() {
	var base, super, sub *MMLNode
	var i int
	for i = 0; i < len(n.Children); i++ {
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
		if i > 0 {
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

func (n *MMLNode) postProcessInfix() {
	doFraction := func(name string, numerator *MMLNode, denominator *MMLNode) *MMLNode {
		// for a binomial coefficient, we need to wrap it in parentheses, so the "fraction" must
		// be a child of parent, and parent must be an mrow.
		wrapper := NewMMLNode("mrow")
		frac := NewMMLNode("mfrac")
		frac.AppendChild(numerator, denominator)
		switch name {
		case "", "frac":
			return frac
		case "cfrac", "dfrac":
			frac.SetTrue("displaystyle")
			return frac
		case "tfrac":
			frac.SetFalse("displaystyle")
			return frac
		case "binom":
			frac.SetAttr("linethickness", "0")
			wrapper.AppendChild(strechyOP("("), frac, strechyOP(")"))
		case "tbinom":
			wrapper.SetFalse("displaystyle")
			frac.SetAttr("linethickness", "0")
			wrapper.AppendChild(strechyOP("("), frac, strechyOP(")"))
		}
		return wrapper
	}
	for i := 1; i < len(n.Children); i++ {
		a := n.Children[i-1]
		b := n.Children[i]
		if b == nil {
			continue
		}
		if b.Properties&propInfixOver > 0 {
			n.Children[i-1] = doFraction("frac", a, b)
		} else if b.Properties&propInfixChoose > 0 {
			n.Children[i-1] = doFraction("binom", a, b)
		} else if b.Properties&propInfixAtop > 0 {
			n.Children[i-1] = doFraction("frac", a, b).SetAttr("linethickness", "0")
		}
		if b.Properties&(propInfixOver|propInfixChoose|propInfixAtop) > 0 {
			n.Children[i] = nil
		}
	}
}
