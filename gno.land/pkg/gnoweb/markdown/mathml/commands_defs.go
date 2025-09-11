package mathml

import "unicode"

func cmd_multirow(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	var attr string
	if name == "multirow" {
		attr = "rowspan"
	} else {
		attr = "columnspan"
	}
	n := converter.ParseTex(args[2], ctx)
	n.SetAttr(attr, StringifyTokens(args[0].Expr))
	return n
}

func cmd_prescript(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	super := args[0]
	sub := args[1]
	base := args[2]
	multi := NewMMLNode("mmultiscripts")
	multi.AppendChild(converter.ParseTex(base, ctx))
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
	left := args[0]
	right := args[1]
	base := args[2]
	multi := NewMMLNode("mmultiscripts")
	multi.Properties |= propLimitsunderover
	multi.AppendChild(converter.ParseTex(base, ctx))
	getScripts := func(side *TokenBuffer) []*MMLNode {
		subscripts := make([]*MMLNode, 0)
		superscripts := make([]*MMLNode, 0)
		var last string
		for !side.Empty() {
			t, err := side.GetNextToken()
			if err != nil {
				continue
			}
			switch t.Value {
			case "^":
				if last == t.Value {
					subscripts = append(subscripts, NewMMLNode("none"))
				}
				expr, err := side.GetNextExpr()
				if err != nil {
					expr, err = side.GetNextN(1, true)
				}
				superscripts = append(superscripts, converter.ParseTex(expr, ctx))
				last = t.Value
			case "_":
				if last == t.Value {
					superscripts = append(superscripts, NewMMLNode("none"))
				}
				expr, err := side.GetNextExpr()
				if err != nil {
					expr, err = side.GetNextN(1, true)
				}
				subscripts = append(subscripts, converter.ParseTex(expr, ctx))
				last = t.Value
			}
		}
		if len(superscripts) == 0 {
			superscripts = append(superscripts, NewMMLNode("none"))
		}
		if len(subscripts) == 0 {
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

func cmd_textcolor(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	n := converter.ParseTex(args[1], ctx)
	n.SetAttr("mathcolor", StringifyTokens(args[0].Expr))
	return n
}

func cmd_undersetOverset(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	var base, embellishment *MMLNode
	base = converter.ParseTex(args[1], ctx)
	embellishment = converter.ParseTex(args[0], ctx)
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
	n := converter.ParseTex(args[1], ctx)
	n.SetAttr("class", StringifyTokens(args[0].Expr))
	return n
}

func cmd_raisebox(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	n := NewMMLNode("mpadded").SetAttr("voffset", StringifyTokens(args[0].Expr))
	converter.ParseTex(args[1], ctx, n)
	return n
}

func cmd_cancel(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
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
	n := NewMMLNode("mo", StringifyTokens(args[0].Expr)).SetAttr("rspace", "0")
	n.Properties |= propLimitsunderover | propMovablelimits
	return n
}

func cmd_mod(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	n := NewMMLNode("mrow")
	if name == "pmod" {
		space := NewMMLNode("mspace").SetAttr("width", "0.7em")
		mod := NewMMLNode("mo", "mod").SetAttr("lspace", "0")
		n.AppendChild(space,
			NewMMLNode("mo", "("),
			mod,
			converter.ParseTex(args[0], ctx),
			NewMMLNode("mo", ")"),
		)
	} else {
		space := NewMMLNode("mspace").SetAttr("width", "0.5em")
		mod := NewMMLNode("mo", "mod")
		n.AppendChild(space,
			mod,
			converter.ParseTex(args[0], ctx),
		)
	}
	return n
}

func cmd_substack(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	n := converter.ParseTex(args[0], ctx|ctxTable)
	processTable(n)
	n.SetAttr("rowspacing", "0") // Incredibly, chrome does this by default
	n.SetFalse("displaystyle")
	return n
}

func cmd_underOverBrace(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	annotation := converter.ParseTex(args[0], ctx)
	n := NewMMLNode()
	brace := NewMMLNode("mo")
	brace.SetTrue("stretchy")
	n.Properties |= propLimitsunderover
	switch name {
	case "overbrace":
		n.Tag = "mover"
		brace.Text = "&OverBrace;"
	case "underbrace":
		n.Tag = "munder"
		brace.Text = "&UnderBrace;"
	}
	n.AppendChild(annotation, brace)
	return n

}

func cmd_not(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
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
	n := NewMMLNode("msqrt")
	n.AppendChild(converter.ParseTex(args[0], ctx))
	if opt != nil {
		n.Tag = "mroot"
		n.AppendChild(converter.ParseTex(opt, ctx))
	}
	return n
}

func cmd_text(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	return NewMMLNode("mtext", stringifyTokensHtml(args[0].Expr))
}

func cmd_frac(converter *MathMLConverter, name string, star bool, ctx parseContext, args []*TokenBuffer, opt *TokenBuffer) *MMLNode {
	// for a binomial coefficient, we need to wrap it in parentheses, so the "fraction" must
	// be a child of parent, and parent must be an mrow.
	wrapper := NewMMLNode("mrow")
	frac := NewMMLNode("mfrac")
	numerator := converter.ParseTex(args[0], ctx)
	denominator := converter.ParseTex(args[1], ctx)
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
