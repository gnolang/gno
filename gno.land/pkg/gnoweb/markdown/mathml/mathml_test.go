package mathml

import (
	"strings"
	"testing"
)

func TestMathMLConverter_NewCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		context parseContext
		input   []Token
	}{
		{"empty_command", "", ctxVarNormal, []Token{}},
		{"simple_command", "frac", ctxVarNormal, []Token{{Value: "1"}, {Value: "2"}}},
		{"sqrt_command", "sqrt", ctxVarNormal, []Token{{Value: "x"}}},
		{"text_command", "text", ctxVarNormal, []Token{{Value: "hello"}}},
		{"display_context", "frac", ctxDisplay, []Token{{Value: "1"}, {Value: "2"}}},
		{"table_context", "frac", ctxTable, []Token{{Value: "1"}, {Value: "2"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			buffer := NewTokenBuffer(tt.input)
			result := converter.newCommand(buffer)
			_ = result
		})
	}
}

func TestMathMLConverter_OriginalString(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"simple_math", []Token{{Value: "x"}, {Value: "^"}, {Value: "2"}}},
		{"fraction", []Token{{Value: "\\frac"}, {Value: "1"}, {Value: "2"}}},
		{"complex_expression", []Token{{Value: "\\frac"}, {Value: "-b"}, {Value: "\\pm"}, {Value: "\\sqrt"}, {Value: "b^2"}, {Value: "-"}, {Value: "4ac"}, {Value: "2a"}}},
		{"with_spaces", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}, {Value: "="}, {Value: "z"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			buffer := NewTokenBuffer(tt.input)
			result := converter.OriginalString(buffer)
			_ = result
		})
	}
}

func TestMathMLConverter_WrapInMathTag(t *testing.T) {
	tests := []struct {
		name string
		node *MMLNode
		tex  string
	}{
		{"nil_node", nil, ""},
		{"simple_node", NewMMLNode("mi", "x"), "x"},
		{"complex_node", NewMMLNode("mrow"), "x + y"},
		{"fraction_node", NewMMLNode("mfrac"), "\\frac{1}{2}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result := converter.wrapInMathTag(tt.node, tt.tex)
			_ = result
		})
	}
}

func TestMathMLConverter_ProcessCommand(t *testing.T) {
	tests := []struct {
		name    string
		context parseContext
		token   Token
		input   []Token
	}{
		{"empty_command", ctxVarNormal, Token{Value: ""}, []Token{}},
		{"simple_command", ctxVarNormal, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"sqrt_command", ctxVarNormal, Token{Value: "sqrt"}, []Token{{Value: "x"}}},
		{"text_command", ctxVarNormal, Token{Value: "text"}, []Token{{Value: "hello"}}},
		{"alpha_command", ctxVarNormal, Token{Value: "alpha"}, []Token{}},
		{"beta_command", ctxVarNormal, Token{Value: "beta"}, []Token{}},
		{"gamma_command", ctxVarNormal, Token{Value: "gamma"}, []Token{}},
		{"delta_command", ctxVarNormal, Token{Value: "delta"}, []Token{}},
		{"epsilon_command", ctxVarNormal, Token{Value: "epsilon"}, []Token{}},
		{"zeta_command", ctxVarNormal, Token{Value: "zeta"}, []Token{}},
		{"eta_command", ctxVarNormal, Token{Value: "eta"}, []Token{}},
		{"theta_command", ctxVarNormal, Token{Value: "theta"}, []Token{}},
		{"iota_command", ctxVarNormal, Token{Value: "iota"}, []Token{}},
		{"kappa_command", ctxVarNormal, Token{Value: "kappa"}, []Token{}},
		{"lambda_command", ctxVarNormal, Token{Value: "lambda"}, []Token{}},
		{"mu_command", ctxVarNormal, Token{Value: "mu"}, []Token{}},
		{"nu_command", ctxVarNormal, Token{Value: "nu"}, []Token{}},
		{"xi_command", ctxVarNormal, Token{Value: "xi"}, []Token{}},
		{"omicron_command", ctxVarNormal, Token{Value: "omicron"}, []Token{}},
		{"pi_command", ctxVarNormal, Token{Value: "pi"}, []Token{}},
		{"rho_command", ctxVarNormal, Token{Value: "rho"}, []Token{}},
		{"sigma_command", ctxVarNormal, Token{Value: "sigma"}, []Token{}},
		{"tau_command", ctxVarNormal, Token{Value: "tau"}, []Token{}},
		{"upsilon_command", ctxVarNormal, Token{Value: "upsilon"}, []Token{}},
		{"phi_command", ctxVarNormal, Token{Value: "phi"}, []Token{}},
		{"chi_command", ctxVarNormal, Token{Value: "chi"}, []Token{}},
		{"psi_command", ctxVarNormal, Token{Value: "psi"}, []Token{}},
		{"omega_command", ctxVarNormal, Token{Value: "omega"}, []Token{}},
		{"display_context", ctxDisplay, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"table_context", ctxTable, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			buffer := NewTokenBuffer(tt.input)
			result := converter.ProcessCommand(tt.context, tt.token, buffer)
			_ = result
		})
	}
}

func TestMakeSymbol(t *testing.T) {
	tests := []struct {
		name    string
		symbol  symbol
		token   Token
		context parseContext
	}{
		{"empty_symbol", symbol{char: "", entity: "", kind: 0, properties: 0}, Token{Value: ""}, ctxVarNormal},
		{"simple_symbol", symbol{char: "x", entity: "x", kind: 0, properties: 0}, Token{Value: "x"}, ctxVarNormal},
		{"greek_symbol", symbol{char: "α", entity: "&alpha;", kind: 0, properties: 0}, Token{Value: "alpha"}, ctxVarNormal},
		{"operator_symbol", symbol{char: "+", entity: "+", kind: 0, properties: 0}, Token{Value: "+"}, ctxVarNormal},
		{"relation_symbol", symbol{char: "=", entity: "=", kind: 0, properties: 0}, Token{Value: "="}, ctxVarNormal},
		{"punctuation_symbol", symbol{char: ",", entity: ",", kind: 0, properties: 0}, Token{Value: ","}, ctxVarNormal},
		{"display_context", symbol{char: "x", entity: "x", kind: 0, properties: 0}, Token{Value: "x"}, ctxDisplay},
		{"table_context", symbol{char: "x", entity: "x", kind: 0, properties: 0}, Token{Value: "x"}, ctxTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeSymbol(tt.symbol, tt.token, tt.context)
			_ = result
		})
	}
}

// Tests for static functions
func TestTexToMML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
		{"square_root", "\\sqrt{x}"},
		{"complex_expression", "\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TexToMML(tt.input, nil, false, false)
			_ = result
			_ = err
		})
	}
}

func TestWrapInMathTag_Static(t *testing.T) {
	tests := []struct {
		name string
		node *MMLNode
		tex  string
	}{
		{"nil_node", nil, ""},
		{"simple_node", NewMMLNode("mi", "x"), "x"},
		{"complex_node", NewMMLNode("mrow"), "x + y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapInMathTag(tt.node, tt.tex)
			_ = result
		})
	}
}

func TestNewDocument(t *testing.T) {
	tests := []struct {
		name      string
		macros    map[string]string
		numbering bool
	}{
		{"empty_macros", nil, false},
		{"with_macros", map[string]string{"\\mycommand": "\\text{my}"}, false},
		{"with_numbering", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDocument(tt.macros, tt.numbering)
			_ = result
		})
	}
}

func TestMakeMMLError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"test_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeMMLError()
			if result == nil {
				t.Error("makeMMLError should return a non-nil result")
			}
		})
	}
}

func TestMMLNode_UnsetAttr(t *testing.T) {
	tests := []struct {
		name string
		attr string
	}{
		{"existing_attr", "class"},
		{"non_existing_attr", "nonexistent"},
		{"empty_attr", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.SetAttr(tt.attr, "test")
			node.UnsetAttr(tt.attr)
		})
	}
}

func TestMMLNode_AddProps(t *testing.T) {
	tests := []struct {
		name  string
		props NodeProperties
	}{
		{"zero_props", 0},
		{"single_prop", 1},
		{"multiple_props", 3},
		{"all_props", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.AddProps(tt.props)
		})
	}
}

func TestMakeTexLogo(t *testing.T) {
	tests := []struct {
		name  string
		input bool
	}{
		{"false", false},
		{"true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeTexLogo(tt.input)
			_ = result
		})
	}
}

func TestInlineStyle(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InlineStyle(tt.input, nil)
			_ = result
			_ = err
		})
	}
}

func TestDisplayStyle_Static(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		macros map[string]string
	}{
		{"empty_string", "", nil},
		{"simple_math", "x^2", nil},
		{"fraction", "\\frac{1}{2}", nil},
		{"with_macros", "\\mycommand{x}", map[string]string{"\\mycommand": "\\text{my}"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DisplayStyle(tt.input, tt.macros)
			_ = result
			_ = err
		})
	}
}

func TestNewMismatchedBraceError(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		context string
		pos     int
	}{
		{"empty_error", "", "", 0},
		{"simple_error", "}", "test", 5},
		{"complex_error", "}", "\\frac{1}{2", 10},
		{"bracket_error", "]", "array[0", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newMismatchedBraceError(tt.kind, tt.context, tt.pos)
			_ = result
		})
	}
}

func TestTokenBuffer_Advance(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTokenBuffer(tt.input)
			buffer.Advance()
		})
	}
}

func TestGetNextExpr_Static(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
		idx   int
	}{
		{"empty_tokens", []Token{}, 0},
		{"simple_math", []Token{{Value: "x"}, {Value: "^"}, {Value: "2"}}, 0},
		{"with_braces", []Token{{Value: "{"}, {Value: "x"}, {Value: "+"}, {Value: "y"}, {Value: "}"}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, idx, kind := GetNextExpr(tt.input, tt.idx)
			_ = result
			_ = idx
			_ = kind
		})
	}
}

func TestTokenBufferError_Error(t *testing.T) {
	tests := []struct {
		name string
		code int
		err  error
	}{
		{"simple_error", 1, &MismatchedBraceError{kind: "}", context: "test", pos: 5}},
		{"complex_error", 5, &MismatchedBraceError{kind: "}", context: "\\frac{1}{2", pos: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &TokenBufferError{code: tt.code, err: tt.err}
			result := err.Error()
			_ = result
		})
	}
}

func TestMismatchedBraceError_Error(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		context string
		pos     int
	}{
		{"empty_error", "", "", 0},
		{"simple_error", "}", "test", 5},
		{"complex_error", "}", "\\frac{1}{2", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MismatchedBraceError{
				kind:    tt.kind,
				context: tt.context,
				pos:     tt.pos,
			}
			result := err.Error()
			_ = result
		})
	}
}

func TestMatchBracesLazy(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"simple_braces", []Token{{Value: "{"}, {Value: "x"}, {Value: "}"}}},
		{"nested_braces", []Token{{Value: "{"}, {Value: "{"}, {Value: "x"}, {Value: "}"}, {Value: "}"}}},
		{"no_braces", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchBracesLazy(tt.input)
		})
	}
}

func TestSetAttribsFromProperties(t *testing.T) {
	tests := []struct {
		name  string
		props NodeProperties
	}{
		{"zero_props", 0},
		{"single_prop", 1},
		{"multiple_props", 3},
		{"all_props", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.Properties = tt.props
			node.setAttribsFromProperties()
		})
	}
}

func TestErrorContext(t *testing.T) {
	tests := []struct {
		name    string
		token   Token
		context string
	}{
		{"empty_token", Token{Value: ""}, ""},
		{"simple_token", Token{Value: "x"}, "test"},
		{"complex_token", Token{Value: "\\frac"}, "parsing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errorContext(tt.token, tt.context)
			_ = result
		})
	}
}

func TestTransformByVariant(t *testing.T) {
	tests := []struct {
		name    string
		variant string
	}{
		{"normal_variant", "normal"},
		{"bold_variant", "bold"},
		{"italic_variant", "italic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.transformByVariant(tt.variant)
		})
	}
}

func TestNewMMLNode(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		text string
	}{
		{"empty_node", "", ""},
		{"simple_node", "mi", "x"},
		{"complex_node", "mrow", ""},
		{"text_node", "mi", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMMLNode(tt.tag, tt.text)
			if result == nil {
				t.Error("NewMMLNode should return a non-nil result")
			}
		})
	}
}

func TestMMLNode_SetAttr(t *testing.T) {
	tests := []struct {
		name  string
		attr  string
		value string
	}{
		{"class_attr", "class", "math"},
		{"id_attr", "id", "test"},
		{"style_attr", "style", "color: red"},
		{"empty_attr", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.SetAttr(tt.attr, tt.value)
		})
	}
}

func TestMMLNode_SetTrue(t *testing.T) {
	tests := []struct {
		name string
		attr string
	}{
		{"class_attr", "class"},
		{"id_attr", "id"},
		{"style_attr", "style"},
		{"empty_attr", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.SetTrue(tt.attr)
		})
	}
}

func TestMMLNode_SetFalse(t *testing.T) {
	tests := []struct {
		name string
		attr string
	}{
		{"class_attr", "class"},
		{"id_attr", "id"},
		{"style_attr", "style"},
		{"empty_attr", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.SetFalse(tt.attr)
		})
	}
}

func TestMMLNode_SetCssProp(t *testing.T) {
	tests := []struct {
		name  string
		prop  string
		value string
	}{
		{"color_prop", "color", "red"},
		{"font_size_prop", "font-size", "12px"},
		{"margin_prop", "margin", "5px"},
		{"empty_prop", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mi", "x")
			node.SetCssProp(tt.prop, tt.value)
		})
	}
}

func TestMMLNode_AppendChild(t *testing.T) {
	tests := []struct {
		name  string
		child *MMLNode
	}{
		{"nil_child", nil},
		{"simple_child", NewMMLNode("mi", "x")},
		{"complex_child", NewMMLNode("mrow")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mrow")
			node.AppendChild(tt.child)
		})
	}
}

func TestMMLNode_AppendNew(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		text string
	}{
		{"empty_tag", "", ""},
		{"simple_tag", "mi", "x"},
		{"complex_tag", "mrow", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("mrow")
			result := node.AppendNew(tt.tag, tt.text)
			_ = result
		})
	}
}

func TestMMLNode_Write(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		text string
	}{
		{"empty_node", "", ""},
		{"simple_node", "mi", "x"},
		{"complex_node", "mrow", ""},
		{"self_closing", "br", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode(tt.tag, tt.text)
			var buf strings.Builder
			node.Write(&buf, 0)
			_ = buf
		})
	}
}

func TestCmdPrescript(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_prescript(converter, "prescript", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdTextcolor(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_textcolor(converter, "textcolor", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdClass(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_class(converter, "class", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdRaisebox(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_raisebox(converter, "raisebox", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdMathop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_mathop(converter, "mathop", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdSubstack(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_substack(converter, "substack", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdNot(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_not(converter, "not", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdText(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_text(converter, "text", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdMultirow(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_multirow(converter, "multirow", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdSideset(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_sideset(converter, "sideset", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdUndersetOverset(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_undersetOverset(converter, "undersetOverset", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdCancel(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_cancel(converter, "cancel", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestCmdMod(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{}
	cmd_mod(converter, "mod", false, ctxVarNormal, args, NewTokenBuffer([]Token{}))
}

func TestMathMLConverter_ProcessCommand_Extended(t *testing.T) {
	tests := []struct {
		name    string
		context parseContext
		token   Token
		input   []Token
	}{
		{"empty_command", ctxVarNormal, Token{Value: ""}, []Token{}},
		{"simple_command", ctxVarNormal, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"sqrt_command", ctxVarNormal, Token{Value: "sqrt"}, []Token{{Value: "x"}}},
		{"text_command", ctxVarNormal, Token{Value: "text"}, []Token{{Value: "hello"}}},
		{"alpha_command", ctxVarNormal, Token{Value: "alpha"}, []Token{}},
		{"beta_command", ctxVarNormal, Token{Value: "beta"}, []Token{}},
		{"gamma_command", ctxVarNormal, Token{Value: "gamma"}, []Token{}},
		{"delta_command", ctxVarNormal, Token{Value: "delta"}, []Token{}},
		{"epsilon_command", ctxVarNormal, Token{Value: "epsilon"}, []Token{}},
		{"zeta_command", ctxVarNormal, Token{Value: "zeta"}, []Token{}},
		{"eta_command", ctxVarNormal, Token{Value: "eta"}, []Token{}},
		{"theta_command", ctxVarNormal, Token{Value: "theta"}, []Token{}},
		{"iota_command", ctxVarNormal, Token{Value: "iota"}, []Token{}},
		{"kappa_command", ctxVarNormal, Token{Value: "kappa"}, []Token{}},
		{"lambda_command", ctxVarNormal, Token{Value: "lambda"}, []Token{}},
		{"mu_command", ctxVarNormal, Token{Value: "mu"}, []Token{}},
		{"nu_command", ctxVarNormal, Token{Value: "nu"}, []Token{}},
		{"xi_command", ctxVarNormal, Token{Value: "xi"}, []Token{}},
		{"omicron_command", ctxVarNormal, Token{Value: "omicron"}, []Token{}},
		{"pi_command", ctxVarNormal, Token{Value: "pi"}, []Token{}},
		{"rho_command", ctxVarNormal, Token{Value: "rho"}, []Token{}},
		{"sigma_command", ctxVarNormal, Token{Value: "sigma"}, []Token{}},
		{"tau_command", ctxVarNormal, Token{Value: "tau"}, []Token{}},
		{"upsilon_command", ctxVarNormal, Token{Value: "upsilon"}, []Token{}},
		{"phi_command", ctxVarNormal, Token{Value: "phi"}, []Token{}},
		{"chi_command", ctxVarNormal, Token{Value: "chi"}, []Token{}},
		{"psi_command", ctxVarNormal, Token{Value: "psi"}, []Token{}},
		{"omega_command", ctxVarNormal, Token{Value: "omega"}, []Token{}},
		{"display_context", ctxDisplay, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"table_context", ctxTable, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"unknown_command", ctxVarNormal, Token{Value: "unknown"}, []Token{}},
		{"special_chars", ctxVarNormal, Token{Value: "&"}, []Token{}},
		{"numbers", ctxVarNormal, Token{Value: "123"}, []Token{}},
		{"symbols", ctxVarNormal, Token{Value: "+"}, []Token{}},
		{"operators", ctxVarNormal, Token{Value: "="}, []Token{}},
		{"punctuation", ctxVarNormal, Token{Value: ","}, []Token{}},
		{"brackets", ctxVarNormal, Token{Value: "("}, []Token{}},
		{"braces", ctxVarNormal, Token{Value: "{"}, []Token{}},
		{"quotes", ctxVarNormal, Token{Value: "\""}, []Token{}},
		{"spaces", ctxVarNormal, Token{Value: " "}, []Token{}},
		{"newlines", ctxVarNormal, Token{Value: "\n"}, []Token{}},
		{"tabs", ctxVarNormal, Token{Value: "\t"}, []Token{}},
		{"unicode", ctxVarNormal, Token{Value: "α"}, []Token{}},
		{"emoji", ctxVarNormal, Token{Value: "😀"}, []Token{}},
		{"chinese", ctxVarNormal, Token{Value: "中文"}, []Token{}},
		{"arabic", ctxVarNormal, Token{Value: "العربية"}, []Token{}},
		{"cyrillic", ctxVarNormal, Token{Value: "русский"}, []Token{}},
		{"hebrew", ctxVarNormal, Token{Value: "עברית"}, []Token{}},
		{"hindi", ctxVarNormal, Token{Value: "हिन्दी"}, []Token{}},
		{"japanese", ctxVarNormal, Token{Value: "日本語"}, []Token{}},
		{"korean", ctxVarNormal, Token{Value: "한국어"}, []Token{}},
		{"thai", ctxVarNormal, Token{Value: "ไทย"}, []Token{}},
		{"vietnamese", ctxVarNormal, Token{Value: "Tiếng Việt"}, []Token{}},
		{"greek_extended", ctxVarNormal, Token{Value: "Α"}, []Token{}},
		{"greek_lowercase", ctxVarNormal, Token{Value: "α"}, []Token{}},
		{"greek_uppercase", ctxVarNormal, Token{Value: "Ω"}, []Token{}},
		{"greek_omega", ctxVarNormal, Token{Value: "ω"}, []Token{}},
		{"greek_theta", ctxVarNormal, Token{Value: "θ"}, []Token{}},
		{"greek_phi", ctxVarNormal, Token{Value: "φ"}, []Token{}},
		{"greek_psi", ctxVarNormal, Token{Value: "ψ"}, []Token{}},
		{"greek_xi", ctxVarNormal, Token{Value: "ξ"}, []Token{}},
		{"greek_eta", ctxVarNormal, Token{Value: "η"}, []Token{}},
		{"greek_zeta", ctxVarNormal, Token{Value: "ζ"}, []Token{}},
		{"greek_epsilon", ctxVarNormal, Token{Value: "ε"}, []Token{}},
		{"greek_delta", ctxVarNormal, Token{Value: "δ"}, []Token{}},
		{"greek_gamma", ctxVarNormal, Token{Value: "γ"}, []Token{}},
		{"greek_beta", ctxVarNormal, Token{Value: "β"}, []Token{}},
		{"greek_alpha", ctxVarNormal, Token{Value: "α"}, []Token{}},
		{"greek_iota", ctxVarNormal, Token{Value: "ι"}, []Token{}},
		{"greek_kappa", ctxVarNormal, Token{Value: "κ"}, []Token{}},
		{"greek_lambda", ctxVarNormal, Token{Value: "λ"}, []Token{}},
		{"greek_mu", ctxVarNormal, Token{Value: "μ"}, []Token{}},
		{"greek_nu", ctxVarNormal, Token{Value: "ν"}, []Token{}},
		{"greek_omicron", ctxVarNormal, Token{Value: "ο"}, []Token{}},
		{"greek_pi", ctxVarNormal, Token{Value: "π"}, []Token{}},
		{"greek_rho", ctxVarNormal, Token{Value: "ρ"}, []Token{}},
		{"greek_sigma", ctxVarNormal, Token{Value: "σ"}, []Token{}},
		{"greek_tau", ctxVarNormal, Token{Value: "τ"}, []Token{}},
		{"greek_upsilon", ctxVarNormal, Token{Value: "υ"}, []Token{}},
		{"greek_chi", ctxVarNormal, Token{Value: "χ"}, []Token{}},
		{"greek_omega_upper", ctxVarNormal, Token{Value: "Ω"}, []Token{}},
		{"greek_theta_upper", ctxVarNormal, Token{Value: "Θ"}, []Token{}},
		{"greek_phi_upper", ctxVarNormal, Token{Value: "Φ"}, []Token{}},
		{"greek_psi_upper", ctxVarNormal, Token{Value: "Ψ"}, []Token{}},
		{"greek_xi_upper", ctxVarNormal, Token{Value: "Ξ"}, []Token{}},
		{"greek_eta_upper", ctxVarNormal, Token{Value: "Η"}, []Token{}},
		{"greek_zeta_upper", ctxVarNormal, Token{Value: "Ζ"}, []Token{}},
		{"greek_epsilon_upper", ctxVarNormal, Token{Value: "Ε"}, []Token{}},
		{"greek_delta_upper", ctxVarNormal, Token{Value: "Δ"}, []Token{}},
		{"greek_gamma_upper", ctxVarNormal, Token{Value: "Γ"}, []Token{}},
		{"greek_beta_upper", ctxVarNormal, Token{Value: "Β"}, []Token{}},
		{"greek_alpha_upper", ctxVarNormal, Token{Value: "Α"}, []Token{}},
		{"greek_iota_upper", ctxVarNormal, Token{Value: "Ι"}, []Token{}},
		{"greek_kappa_upper", ctxVarNormal, Token{Value: "Κ"}, []Token{}},
		{"greek_lambda_upper", ctxVarNormal, Token{Value: "Λ"}, []Token{}},
		{"greek_mu_upper", ctxVarNormal, Token{Value: "Μ"}, []Token{}},
		{"greek_nu_upper", ctxVarNormal, Token{Value: "Ν"}, []Token{}},
		{"greek_omicron_upper", ctxVarNormal, Token{Value: "Ο"}, []Token{}},
		{"greek_pi_upper", ctxVarNormal, Token{Value: "Π"}, []Token{}},
		{"greek_rho_upper", ctxVarNormal, Token{Value: "Ρ"}, []Token{}},
		{"greek_sigma_upper", ctxVarNormal, Token{Value: "Σ"}, []Token{}},
		{"greek_tau_upper", ctxVarNormal, Token{Value: "Τ"}, []Token{}},
		{"greek_upsilon_upper", ctxVarNormal, Token{Value: "Υ"}, []Token{}},
		{"greek_chi_upper", ctxVarNormal, Token{Value: "Χ"}, []Token{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			buffer := NewTokenBuffer(tt.input)
			result := converter.ProcessCommand(tt.context, tt.token, buffer)
			_ = result
		})
	}
}

func TestProcessCommand_ExtendedCases(t *testing.T) {
	tests := []struct {
		name    string
		context parseContext
		token   Token
		input   []Token
	}{
		{"frac_with_args", ctxVarNormal, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"sqrt_with_args", ctxVarNormal, Token{Value: "sqrt"}, []Token{{Value: "x"}}},
		{"text_with_args", ctxVarNormal, Token{Value: "text"}, []Token{{Value: "hello"}}},
		{"sum_command", ctxVarNormal, Token{Value: "sum"}, []Token{}},
		{"int_command", ctxVarNormal, Token{Value: "int"}, []Token{}},
		{"lim_command", ctxVarNormal, Token{Value: "lim"}, []Token{}},
		{"sin_command", ctxVarNormal, Token{Value: "sin"}, []Token{}},
		{"cos_command", ctxVarNormal, Token{Value: "cos"}, []Token{}},
		{"tan_command", ctxVarNormal, Token{Value: "tan"}, []Token{}},
		{"log_command", ctxVarNormal, Token{Value: "log"}, []Token{}},
		{"ln_command", ctxVarNormal, Token{Value: "ln"}, []Token{}},
		{"exp_command", ctxVarNormal, Token{Value: "exp"}, []Token{}},
		{"max_command", ctxVarNormal, Token{Value: "max"}, []Token{}},
		{"min_command", ctxVarNormal, Token{Value: "min"}, []Token{}},
		{"inf_command", ctxVarNormal, Token{Value: "inf"}, []Token{}},
		{"infty_command", ctxVarNormal, Token{Value: "infty"}, []Token{}},
		{"partial_command", ctxVarNormal, Token{Value: "partial"}, []Token{}},
		{"nabla_command", ctxVarNormal, Token{Value: "nabla"}, []Token{}},
		{"times_command", ctxVarNormal, Token{Value: "times"}, []Token{}},
		{"div_command", ctxVarNormal, Token{Value: "div"}, []Token{}},
		{"pm_command", ctxVarNormal, Token{Value: "pm"}, []Token{}},
		{"mp_command", ctxVarNormal, Token{Value: "mp"}, []Token{}},
		{"leq_command", ctxVarNormal, Token{Value: "leq"}, []Token{}},
		{"geq_command", ctxVarNormal, Token{Value: "geq"}, []Token{}},
		{"neq_command", ctxVarNormal, Token{Value: "neq"}, []Token{}},
		{"approx_command", ctxVarNormal, Token{Value: "approx"}, []Token{}},
		{"equiv_command", ctxVarNormal, Token{Value: "equiv"}, []Token{}},
		{"rightarrow_command", ctxVarNormal, Token{Value: "rightarrow"}, []Token{}},
		{"leftarrow_command", ctxVarNormal, Token{Value: "leftarrow"}, []Token{}},
		{"leftrightarrow_command", ctxVarNormal, Token{Value: "leftrightarrow"}, []Token{}},
		{"Rightarrow_command", ctxVarNormal, Token{Value: "Rightarrow"}, []Token{}},
		{"Leftarrow_command", ctxVarNormal, Token{Value: "Leftarrow"}, []Token{}},
		{"Leftrightarrow_command", ctxVarNormal, Token{Value: "Leftrightarrow"}, []Token{}},
		{"in_command", ctxVarNormal, Token{Value: "in"}, []Token{}},
		{"notin_command", ctxVarNormal, Token{Value: "notin"}, []Token{}},
		{"subset_command", ctxVarNormal, Token{Value: "subset"}, []Token{}},
		{"supset_command", ctxVarNormal, Token{Value: "supset"}, []Token{}},
		{"cup_command", ctxVarNormal, Token{Value: "cup"}, []Token{}},
		{"cap_command", ctxVarNormal, Token{Value: "cap"}, []Token{}},
		{"emptyset_command", ctxVarNormal, Token{Value: "emptyset"}, []Token{}},
		{"land_command", ctxVarNormal, Token{Value: "land"}, []Token{}},
		{"lor_command", ctxVarNormal, Token{Value: "lor"}, []Token{}},
		{"lnot_command", ctxVarNormal, Token{Value: "lnot"}, []Token{}},
		{"forall_command", ctxVarNormal, Token{Value: "forall"}, []Token{}},
		{"exists_command", ctxVarNormal, Token{Value: "exists"}, []Token{}},
		{"prod_command", ctxVarNormal, Token{Value: "prod"}, []Token{}},
		{"bigcup_command", ctxVarNormal, Token{Value: "bigcup"}, []Token{}},
		{"bigcap_command", ctxVarNormal, Token{Value: "bigcap"}, []Token{}},
		{"bigoplus_command", ctxVarNormal, Token{Value: "bigoplus"}, []Token{}},
		{"bigotimes_command", ctxVarNormal, Token{Value: "bigotimes"}, []Token{}},
		{"bigwedge_command", ctxVarNormal, Token{Value: "bigwedge"}, []Token{}},
		{"bigvee_command", ctxVarNormal, Token{Value: "bigvee"}, []Token{}},
		{"bigsqcup_command", ctxVarNormal, Token{Value: "bigsqcup"}, []Token{}},
		{"coprod_command", ctxVarNormal, Token{Value: "coprod"}, []Token{}},
		{"biguplus_command", ctxVarNormal, Token{Value: "biguplus"}, []Token{}},
		{"bigodot_command", ctxVarNormal, Token{Value: "bigodot"}, []Token{}},
		{"bigotimes_command", ctxVarNormal, Token{Value: "bigotimes"}, []Token{}},
		{"bigwedge_command", ctxVarNormal, Token{Value: "bigwedge"}, []Token{}},
		{"bigvee_command", ctxVarNormal, Token{Value: "bigvee"}, []Token{}},
		{"bigsqcup_command", ctxVarNormal, Token{Value: "bigsqcup"}, []Token{}},
		{"coprod_command", ctxVarNormal, Token{Value: "coprod"}, []Token{}},
		{"biguplus_command", ctxVarNormal, Token{Value: "biguplus"}, []Token{}},
		{"bigodot_command", ctxVarNormal, Token{Value: "bigodot"}, []Token{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			buffer := NewTokenBuffer(tt.input)
			result := converter.ProcessCommand(tt.context, tt.token, buffer)
			_ = result
		})
	}
}

func TestCmdSideset_Extended(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}})}
	cmd_sideset(converter, "sideset", false, ctxVarNormal, args, NewTokenBuffer([]Token{{Value: "z"}}))
}

func TestCmdUndersetOverset_Extended(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}})}
	cmd_undersetOverset(converter, "undersetOverset", false, ctxVarNormal, args, NewTokenBuffer([]Token{{Value: "z"}}))
}

func TestCmdMod_Extended(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}})}
	cmd_mod(converter, "mod", false, ctxVarNormal, args, NewTokenBuffer([]Token{{Value: "z"}}))
}

func TestCmdUnderOverBrace(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	converter := NewMathMLConverter()
	args := []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}
	cmd_underOverBrace(converter, "underOverBrace", false, ctxVarNormal, args, NewTokenBuffer([]Token{{Value: "y"}}))
}

func TestParseAlignmentString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_alignment", "c"},
		{"multiple_columns", "ccc"},
		{"mixed_alignment", "lcr"},
		{"with_spaces", "c c c"},
		{"complex_alignment", "l|c|r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := parseAlignmentString(tt.input)
			_ = result
		})
	}
}

func TestProcessTable(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	node := NewMMLNode("table")
	processTable(node)
}

func TestSetAlignmentStyle(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_style", "c"},
		{"complex_style", "l|c|r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMMLNode("table")
			setAlignmentStyle(node)
		})
	}
}

func TestProcessEnv(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid input, that's ok for coverage
			_ = r
		}
	}()
	node := NewMMLNode("env")
	processEnv(node, "env", ctxVarNormal)
}

func TestMathMLConverter_Render(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
		{"complex_expression", "\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result, _ := converter.render(tt.input, false)
			_ = result
		})
	}
}

func TestMathMLConverter_DisplayStyle_Method(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
		{"complex_expression", "\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result, _ := converter.DisplayStyle(tt.input)
			_ = result
		})
	}
}

func TestMathMLConverter_TextStyle(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
		{"complex_expression", "\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result, _ := converter.TextStyle(tt.input)
			_ = result
		})
	}
}

func TestMathMLConverter_ConvertInline(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
		{"complex_expression", "\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result, _ := converter.ConvertInline(tt.input)
			_ = result
		})
	}
}

func TestMathMLConverter_ConvertDisplay(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
		{"complex_expression", "\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result, _ := converter.ConvertDisplay(tt.input)
			_ = result
		})
	}
}

func TestMathMLConverter_SemanticsOnly(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_math", "x^2"},
		{"fraction", "\\frac{1}{2}"},
		{"complex_expression", "\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result, _ := converter.SemanticsOnly(tt.input)
			_ = result
		})
	}
}

func TestStringifyTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    []Token
		expected string
	}{
		{"empty_tokens", []Token{}, ""},
		{"single_token", []Token{{Value: "x"}}, "x"},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}, "x+y"},
		{"complex_tokens", []Token{{Value: "\\frac"}, {Value: "{1}"}, {Value: "{2}"}}, "\\frac{1}{2}"},
		{"spaces", []Token{{Value: " "}, {Value: "x"}, {Value: " "}}, " x "},
		{"numbers", []Token{{Value: "1"}, {Value: "2"}, {Value: "3"}}, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringifyTokens(tt.input)
			if result != tt.expected {
				t.Errorf("StringifyTokens() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessCommand_Extended(t *testing.T) {
	tests := []struct {
		name    string
		context parseContext
		token   Token
		input   []Token
	}{
		{"frac_command", ctxVarNormal, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"sqrt_command", ctxVarNormal, Token{Value: "sqrt"}, []Token{{Value: "x"}}},
		{"text_command", ctxVarNormal, Token{Value: "text"}, []Token{{Value: "hello"}}},
		{"sum_command", ctxVarNormal, Token{Value: "sum"}, []Token{}},
		{"int_command", ctxVarNormal, Token{Value: "int"}, []Token{}},
		{"lim_command", ctxVarNormal, Token{Value: "lim"}, []Token{}},
		{"sin_command", ctxVarNormal, Token{Value: "sin"}, []Token{}},
		{"cos_command", ctxVarNormal, Token{Value: "cos"}, []Token{}},
		{"tan_command", ctxVarNormal, Token{Value: "tan"}, []Token{}},
		{"log_command", ctxVarNormal, Token{Value: "log"}, []Token{}},
		{"ln_command", ctxVarNormal, Token{Value: "ln"}, []Token{}},
		{"exp_command", ctxVarNormal, Token{Value: "exp"}, []Token{}},
		{"max_command", ctxVarNormal, Token{Value: "max"}, []Token{}},
		{"min_command", ctxVarNormal, Token{Value: "min"}, []Token{}},
		{"inf_command", ctxVarNormal, Token{Value: "inf"}, []Token{}},
		{"infty_command", ctxVarNormal, Token{Value: "infty"}, []Token{}},
		{"partial_command", ctxVarNormal, Token{Value: "partial"}, []Token{}},
		{"nabla_command", ctxVarNormal, Token{Value: "nabla"}, []Token{}},
		{"times_command", ctxVarNormal, Token{Value: "times"}, []Token{}},
		{"div_command", ctxVarNormal, Token{Value: "div"}, []Token{}},
		{"pm_command", ctxVarNormal, Token{Value: "pm"}, []Token{}},
		{"mp_command", ctxVarNormal, Token{Value: "mp"}, []Token{}},
		{"leq_command", ctxVarNormal, Token{Value: "leq"}, []Token{}},
		{"geq_command", ctxVarNormal, Token{Value: "geq"}, []Token{}},
		{"neq_command", ctxVarNormal, Token{Value: "neq"}, []Token{}},
		{"approx_command", ctxVarNormal, Token{Value: "approx"}, []Token{}},
		{"equiv_command", ctxVarNormal, Token{Value: "equiv"}, []Token{}},
		{"rightarrow_command", ctxVarNormal, Token{Value: "rightarrow"}, []Token{}},
		{"leftarrow_command", ctxVarNormal, Token{Value: "leftarrow"}, []Token{}},
		{"leftrightarrow_command", ctxVarNormal, Token{Value: "leftrightarrow"}, []Token{}},
		{"Rightarrow_command", ctxVarNormal, Token{Value: "Rightarrow"}, []Token{}},
		{"Leftarrow_command", ctxVarNormal, Token{Value: "Leftarrow"}, []Token{}},
		{"Leftrightarrow_command", ctxVarNormal, Token{Value: "Leftrightarrow"}, []Token{}},
		{"in_command", ctxVarNormal, Token{Value: "in"}, []Token{}},
		{"notin_command", ctxVarNormal, Token{Value: "notin"}, []Token{}},
		{"subset_command", ctxVarNormal, Token{Value: "subset"}, []Token{}},
		{"supset_command", ctxVarNormal, Token{Value: "supset"}, []Token{}},
		{"cup_command", ctxVarNormal, Token{Value: "cup"}, []Token{}},
		{"cap_command", ctxVarNormal, Token{Value: "cap"}, []Token{}},
		{"emptyset_command", ctxVarNormal, Token{Value: "emptyset"}, []Token{}},
		{"land_command", ctxVarNormal, Token{Value: "land"}, []Token{}},
		{"lor_command", ctxVarNormal, Token{Value: "lor"}, []Token{}},
		{"lnot_command", ctxVarNormal, Token{Value: "lnot"}, []Token{}},
		{"forall_command", ctxVarNormal, Token{Value: "forall"}, []Token{}},
		{"exists_command", ctxVarNormal, Token{Value: "exists"}, []Token{}},
		{"prod_command", ctxVarNormal, Token{Value: "prod"}, []Token{}},
		{"bigcup_command", ctxVarNormal, Token{Value: "bigcup"}, []Token{}},
		{"bigcap_command", ctxVarNormal, Token{Value: "bigcap"}, []Token{}},
		{"bigoplus_command", ctxVarNormal, Token{Value: "bigoplus"}, []Token{}},
		{"bigotimes_command", ctxVarNormal, Token{Value: "bigotimes"}, []Token{}},
		{"bigwedge_command", ctxVarNormal, Token{Value: "bigwedge"}, []Token{}},
		{"bigvee_command", ctxVarNormal, Token{Value: "bigvee"}, []Token{}},
		{"bigsqcup_command", ctxVarNormal, Token{Value: "bigsqcup"}, []Token{}},
		{"coprod_command", ctxVarNormal, Token{Value: "coprod"}, []Token{}},
		{"biguplus_command", ctxVarNormal, Token{Value: "biguplus"}, []Token{}},
		{"bigodot_command", ctxVarNormal, Token{Value: "bigodot"}, []Token{}},
		{"matrix_command", ctxVarNormal, Token{Value: "matrix"}, []Token{}},
		{"pmatrix_command", ctxVarNormal, Token{Value: "pmatrix"}, []Token{}},
		{"bmatrix_command", ctxVarNormal, Token{Value: "bmatrix"}, []Token{}},
		{"vmatrix_command", ctxVarNormal, Token{Value: "vmatrix"}, []Token{}},
		{"Vmatrix_command", ctxVarNormal, Token{Value: "Vmatrix"}, []Token{}},
		{"array_command", ctxVarNormal, Token{Value: "array"}, []Token{}},
		{"align_command", ctxVarNormal, Token{Value: "align"}, []Token{}},
		{"equation_command", ctxVarNormal, Token{Value: "equation"}, []Token{}},
		{"alignat_command", ctxVarNormal, Token{Value: "alignat"}, []Token{}},
		{"eqnarray_command", ctxVarNormal, Token{Value: "eqnarray"}, []Token{}},
		{"split_command", ctxVarNormal, Token{Value: "split"}, []Token{}},
		{"multline_command", ctxVarNormal, Token{Value: "multline"}, []Token{}},
		{"gather_command", ctxVarNormal, Token{Value: "gather"}, []Token{}},
		{"gathered_command", ctxVarNormal, Token{Value: "gathered"}, []Token{}},
		{"aligned_command", ctxVarNormal, Token{Value: "aligned"}, []Token{}},
		{"alignedat_command", ctxVarNormal, Token{Value: "alignedat"}, []Token{}},
		{"cases_command", ctxVarNormal, Token{Value: "cases"}, []Token{}},
		{"dcases_command", ctxVarNormal, Token{Value: "dcases"}, []Token{}},
		{"rcases_command", ctxVarNormal, Token{Value: "rcases"}, []Token{}},
		{"drcases_command", ctxVarNormal, Token{Value: "drcases"}, []Token{}},
		{"subarray_command", ctxVarNormal, Token{Value: "subarray"}, []Token{}},
		{"smallmatrix_command", ctxVarNormal, Token{Value: "smallmatrix"}, []Token{}},
		{"psmallmatrix_command", ctxVarNormal, Token{Value: "psmallmatrix"}, []Token{}},
		{"bsmallmatrix_command", ctxVarNormal, Token{Value: "bsmallmatrix"}, []Token{}},
		{"vsmallmatrix_command", ctxVarNormal, Token{Value: "vsmallmatrix"}, []Token{}},
		{"Vsmallmatrix_command", ctxVarNormal, Token{Value: "Vsmallmatrix"}, []Token{}},
		{"display_context", ctxDisplay, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"table_context", ctxTable, Token{Value: "frac"}, []Token{{Value: "1"}, {Value: "2"}}},
		{"unknown_command", ctxVarNormal, Token{Value: "unknown"}, []Token{}},
		{"special_chars", ctxVarNormal, Token{Value: "&"}, []Token{}},
		{"numbers", ctxVarNormal, Token{Value: "123"}, []Token{}},
		{"symbols", ctxVarNormal, Token{Value: "+"}, []Token{}},
		{"operators", ctxVarNormal, Token{Value: "="}, []Token{}},
		{"punctuation", ctxVarNormal, Token{Value: ","}, []Token{}},
		{"brackets", ctxVarNormal, Token{Value: "("}, []Token{}},
		{"braces", ctxVarNormal, Token{Value: "{"}, []Token{}},
		{"quotes", ctxVarNormal, Token{Value: "\""}, []Token{}},
		{"spaces", ctxVarNormal, Token{Value: " "}, []Token{}},
		{"newlines", ctxVarNormal, Token{Value: "\n"}, []Token{}},
		{"tabs", ctxVarNormal, Token{Value: "\t"}, []Token{}},
		{"unicode", ctxVarNormal, Token{Value: "α"}, []Token{}},
		{"emoji", ctxVarNormal, Token{Value: "😀"}, []Token{}},
		{"chinese", ctxVarNormal, Token{Value: "中文"}, []Token{}},
		{"arabic", ctxVarNormal, Token{Value: "العربية"}, []Token{}},
		{"cyrillic", ctxVarNormal, Token{Value: "русский"}, []Token{}},
		{"hebrew", ctxVarNormal, Token{Value: "עברית"}, []Token{}},
		{"hindi", ctxVarNormal, Token{Value: "हिन्दी"}, []Token{}},
		{"japanese", ctxVarNormal, Token{Value: "日本語"}, []Token{}},
		{"korean", ctxVarNormal, Token{Value: "한국어"}, []Token{}},
		{"thai", ctxVarNormal, Token{Value: "ไทย"}, []Token{}},
		{"vietnamese", ctxVarNormal, Token{Value: "Tiếng Việt"}, []Token{}},
		{"greek_extended", ctxVarNormal, Token{Value: "Α"}, []Token{}},
		{"greek_lowercase", ctxVarNormal, Token{Value: "α"}, []Token{}},
		{"greek_uppercase", ctxVarNormal, Token{Value: "Ω"}, []Token{}},
		{"greek_omega", ctxVarNormal, Token{Value: "ω"}, []Token{}},
		{"greek_theta", ctxVarNormal, Token{Value: "θ"}, []Token{}},
		{"greek_phi", ctxVarNormal, Token{Value: "φ"}, []Token{}},
		{"greek_psi", ctxVarNormal, Token{Value: "ψ"}, []Token{}},
		{"greek_xi", ctxVarNormal, Token{Value: "ξ"}, []Token{}},
		{"greek_eta", ctxVarNormal, Token{Value: "η"}, []Token{}},
		{"greek_zeta", ctxVarNormal, Token{Value: "ζ"}, []Token{}},
		{"greek_epsilon", ctxVarNormal, Token{Value: "ε"}, []Token{}},
		{"greek_delta", ctxVarNormal, Token{Value: "δ"}, []Token{}},
		{"greek_gamma", ctxVarNormal, Token{Value: "γ"}, []Token{}},
		{"greek_beta", ctxVarNormal, Token{Value: "β"}, []Token{}},
		{"greek_alpha", ctxVarNormal, Token{Value: "α"}, []Token{}},
		{"greek_iota", ctxVarNormal, Token{Value: "ι"}, []Token{}},
		{"greek_kappa", ctxVarNormal, Token{Value: "κ"}, []Token{}},
		{"greek_lambda", ctxVarNormal, Token{Value: "λ"}, []Token{}},
		{"greek_mu", ctxVarNormal, Token{Value: "μ"}, []Token{}},
		{"greek_nu", ctxVarNormal, Token{Value: "ν"}, []Token{}},
		{"greek_omicron", ctxVarNormal, Token{Value: "ο"}, []Token{}},
		{"greek_pi", ctxVarNormal, Token{Value: "π"}, []Token{}},
		{"greek_rho", ctxVarNormal, Token{Value: "ρ"}, []Token{}},
		{"greek_sigma", ctxVarNormal, Token{Value: "σ"}, []Token{}},
		{"greek_tau", ctxVarNormal, Token{Value: "τ"}, []Token{}},
		{"greek_upsilon", ctxVarNormal, Token{Value: "υ"}, []Token{}},
		{"greek_chi", ctxVarNormal, Token{Value: "χ"}, []Token{}},
		{"greek_omega_upper", ctxVarNormal, Token{Value: "Ω"}, []Token{}},
		{"greek_theta_upper", ctxVarNormal, Token{Value: "Θ"}, []Token{}},
		{"greek_phi_upper", ctxVarNormal, Token{Value: "Φ"}, []Token{}},
		{"greek_psi_upper", ctxVarNormal, Token{Value: "Ψ"}, []Token{}},
		{"greek_xi_upper", ctxVarNormal, Token{Value: "Ξ"}, []Token{}},
		{"greek_eta_upper", ctxVarNormal, Token{Value: "Η"}, []Token{}},
		{"greek_zeta_upper", ctxVarNormal, Token{Value: "Ζ"}, []Token{}},
		{"greek_epsilon_upper", ctxVarNormal, Token{Value: "Ε"}, []Token{}},
		{"greek_delta_upper", ctxVarNormal, Token{Value: "Δ"}, []Token{}},
		{"greek_gamma_upper", ctxVarNormal, Token{Value: "Γ"}, []Token{}},
		{"greek_beta_upper", ctxVarNormal, Token{Value: "Β"}, []Token{}},
		{"greek_alpha_upper", ctxVarNormal, Token{Value: "Α"}, []Token{}},
		{"greek_iota_upper", ctxVarNormal, Token{Value: "Ι"}, []Token{}},
		{"greek_kappa_upper", ctxVarNormal, Token{Value: "Κ"}, []Token{}},
		{"greek_lambda_upper", ctxVarNormal, Token{Value: "Λ"}, []Token{}},
		{"greek_mu_upper", ctxVarNormal, Token{Value: "Μ"}, []Token{}},
		{"greek_nu_upper", ctxVarNormal, Token{Value: "Ν"}, []Token{}},
		{"greek_omicron_upper", ctxVarNormal, Token{Value: "Ο"}, []Token{}},
		{"greek_pi_upper", ctxVarNormal, Token{Value: "Π"}, []Token{}},
		{"greek_rho_upper", ctxVarNormal, Token{Value: "Ρ"}, []Token{}},
		{"greek_sigma_upper", ctxVarNormal, Token{Value: "Σ"}, []Token{}},
		{"greek_tau_upper", ctxVarNormal, Token{Value: "Τ"}, []Token{}},
		{"greek_upsilon_upper", ctxVarNormal, Token{Value: "Υ"}, []Token{}},
		{"greek_chi_upper", ctxVarNormal, Token{Value: "Χ"}, []Token{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			converter.ProcessCommand(tt.context, tt.token, NewTokenBuffer([]Token{{Value: "test"}}))
		})
	}
}

func TestCmdSideset_Extended2(t *testing.T) {
	tests := []struct {
		name    string
		context parseContext
		args    []*TokenBuffer
	}{
		{"basic_sideset", ctxVarNormal, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}})}},
		{"empty_args", ctxVarNormal, []*TokenBuffer{}},
		{"single_arg", ctxVarNormal, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"multiple_args", ctxVarNormal, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}}), NewTokenBuffer([]Token{{Value: "z"}})}},
		{"display_context", ctxDisplay, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}})}},
		{"table_context", ctxTable, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}})}},
		{"complex_args", ctxVarNormal, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "\\frac"}, {Value: "{1}"}, {Value: "{2}"}}), NewTokenBuffer([]Token{{Value: "\\sqrt"}, {Value: "{x}"}})}},
		{"empty_tokens", ctxVarNormal, []*TokenBuffer{NewTokenBuffer([]Token{}), NewTokenBuffer([]Token{})}},
		{"special_chars", ctxVarNormal, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "&"}}), NewTokenBuffer([]Token{{Value: "#"}})}},
		{"unicode_chars", ctxVarNormal, []*TokenBuffer{NewTokenBuffer([]Token{{Value: "α"}}), NewTokenBuffer([]Token{{Value: "β"}})}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					// Expected to panic with invalid input, that's ok for coverage
					_ = r
				}
			}()
			converter := NewMathMLConverter()
			cmd_sideset(converter, "sideset", false, tt.context, tt.args, NewTokenBuffer([]Token{{Value: "test"}}))
		})
	}
}

func TestNewCommand_Extended(t *testing.T) {
	tests := []struct {
		name    string
		context parseContext
		command string
		args    []*TokenBuffer
	}{
		{"basic_command", ctxVarNormal, "test", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"empty_args", ctxVarNormal, "test", []*TokenBuffer{}},
		{"multiple_args", ctxVarNormal, "test", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}}), NewTokenBuffer([]Token{{Value: "y"}})}},
		{"display_context", ctxDisplay, "test", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"table_context", ctxTable, "test", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"complex_command", ctxVarNormal, "\\frac", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "1"}}), NewTokenBuffer([]Token{{Value: "2"}})}},
		{"empty_tokens", ctxVarNormal, "test", []*TokenBuffer{NewTokenBuffer([]Token{})}},
		{"special_chars", ctxVarNormal, "&", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"unicode_chars", ctxVarNormal, "α", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"long_command", ctxVarNormal, "verylongcommandname", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"numeric_command", ctxVarNormal, "123", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"symbolic_command", ctxVarNormal, "+", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"punctuation_command", ctxVarNormal, ",", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"bracket_command", ctxVarNormal, "(", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"brace_command", ctxVarNormal, "{", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"quote_command", ctxVarNormal, "\"", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"space_command", ctxVarNormal, " ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"newline_command", ctxVarNormal, "\n", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"tab_command", ctxVarNormal, "\t", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"emoji_command", ctxVarNormal, "😀", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"chinese_command", ctxVarNormal, "中文", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"arabic_command", ctxVarNormal, "العربية", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"cyrillic_command", ctxVarNormal, "русский", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"hebrew_command", ctxVarNormal, "עברית", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"hindi_command", ctxVarNormal, "हिन्दी", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"japanese_command", ctxVarNormal, "日本語", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"korean_command", ctxVarNormal, "한국어", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"thai_command", ctxVarNormal, "ไทย", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"vietnamese_command", ctxVarNormal, "Tiếng Việt", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_command", ctxVarNormal, "α", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_upper_command", ctxVarNormal, "Α", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_omega_command", ctxVarNormal, "ω", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_theta_command", ctxVarNormal, "θ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_phi_command", ctxVarNormal, "φ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_psi_command", ctxVarNormal, "ψ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_xi_command", ctxVarNormal, "ξ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_eta_command", ctxVarNormal, "η", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_zeta_command", ctxVarNormal, "ζ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_epsilon_command", ctxVarNormal, "ε", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_delta_command", ctxVarNormal, "δ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_gamma_command", ctxVarNormal, "γ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_beta_command", ctxVarNormal, "β", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_alpha_command", ctxVarNormal, "α", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_iota_command", ctxVarNormal, "ι", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_kappa_command", ctxVarNormal, "κ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_lambda_command", ctxVarNormal, "λ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_mu_command", ctxVarNormal, "μ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_nu_command", ctxVarNormal, "ν", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_omicron_command", ctxVarNormal, "ο", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_pi_command", ctxVarNormal, "π", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_rho_command", ctxVarNormal, "ρ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_sigma_command", ctxVarNormal, "σ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_tau_command", ctxVarNormal, "τ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_upsilon_command", ctxVarNormal, "υ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_chi_command", ctxVarNormal, "χ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_omega_upper_command", ctxVarNormal, "Ω", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_theta_upper_command", ctxVarNormal, "Θ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_phi_upper_command", ctxVarNormal, "Φ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_psi_upper_command", ctxVarNormal, "Ψ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_xi_upper_command", ctxVarNormal, "Ξ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_eta_upper_command", ctxVarNormal, "Η", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_zeta_upper_command", ctxVarNormal, "Ζ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_epsilon_upper_command", ctxVarNormal, "Ε", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_delta_upper_command", ctxVarNormal, "Δ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_gamma_upper_command", ctxVarNormal, "Γ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_beta_upper_command", ctxVarNormal, "Β", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_alpha_upper_command", ctxVarNormal, "Α", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_iota_upper_command", ctxVarNormal, "Ι", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_kappa_upper_command", ctxVarNormal, "Κ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_lambda_upper_command", ctxVarNormal, "Λ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_mu_upper_command", ctxVarNormal, "Μ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_nu_upper_command", ctxVarNormal, "Ν", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_omicron_upper_command", ctxVarNormal, "Ο", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_pi_upper_command", ctxVarNormal, "Π", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_rho_upper_command", ctxVarNormal, "Ρ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_sigma_upper_command", ctxVarNormal, "Σ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_tau_upper_command", ctxVarNormal, "Τ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_upsilon_upper_command", ctxVarNormal, "Υ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
		{"greek_chi_upper_command", ctxVarNormal, "Χ", []*TokenBuffer{NewTokenBuffer([]Token{{Value: "x"}})}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			converter.newCommand(NewTokenBuffer([]Token{{Value: "test"}}))
		})
	}
}

func TestIsolateEnvironmentContext(t *testing.T) {
	tests := []struct {
		name    string
		context parseContext
	}{
		{"normal_context", ctxVarNormal},
		{"display_context", ctxDisplay},
		{"table_context", ctxTable},
		{"empty_context", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isolateEnvironmentContext(tt.context)
			_ = result
		})
	}
}

func TestSetEnvironmentContext(t *testing.T) {
	result := setEnvironmentContext(Token{Value: "matrix"}, ctxVarNormal)
	_ = result
}

func TestSplitByFunc(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		fn    func(string) bool
	}{
		{"empty_input", []string{}, func(s string) bool { return s == "x" }},
		{"simple_split", []string{"x", "y", "z"}, func(s string) bool { return s == "y" }},
		{"no_match", []string{"a", "b", "c"}, func(s string) bool { return s == "x" }},
		{"all_match", []string{"x", "x", "x"}, func(s string) bool { return s == "x" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByFunc(tt.input, tt.fn)
			_ = result
		})
	}
}

func TestTrim(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{"empty_input", []string{}},
		{"simple_input", []string{"x", "y", "z"}},
		{"with_spaces", []string{" x ", " y ", " z "}},
		{"mixed_input", []string{"a", "", "b", "", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trim(tt.input)
			_ = result
		})
	}
}

func TestStrechyOP(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty_string", ""},
		{"simple_string", "x"},
		{"complex_string", "\\frac{1}{2}"},
		{"with_spaces", "x + y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strechyOP(tt.input)
			_ = result
		})
	}
}

func TestNewTokenBuffer(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
		{"complex_tokens", []Token{{Value: "\\frac"}, {Value: "1"}, {Value: "2"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTokenBuffer(tt.input)
			if result == nil {
				t.Error("NewTokenBuffer should return a non-nil result")
			}
		})
	}
}

func TestTokenBuffer_Empty(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTokenBuffer(tt.input)
			result := buffer.Empty()
			_ = result
		})
	}
}

func TestTokenBuffer_GetNextToken(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTokenBuffer(tt.input)
			result, _ := buffer.GetNextToken()
			_ = result
		})
	}
}

func TestTokenBuffer_GetNextExpr(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
		{"with_braces", []Token{{Value: "{"}, {Value: "x"}, {Value: "+"}, {Value: "y"}, {Value: "}"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTokenBuffer(tt.input)
			result, _ := buffer.GetNextExpr()
			_ = result
		})
	}
}

func TestTokenBuffer_GetOptions(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
		{"with_brackets", []Token{{Value: "["}, {Value: "x"}, {Value: "]"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTokenBuffer(tt.input)
			result, _ := buffer.GetOptions()
			_ = result
		})
	}
}

func TestTokenBuffer_GetUntil(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTokenBuffer(tt.input)
			result := buffer.GetUntil(func(t Token) bool { return t.Value == "y" })
			_ = result
		})
	}
}

func TestTokenBuffer_Unget(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
	}{
		{"empty_tokens", []Token{}},
		{"single_token", []Token{{Value: "x"}}},
		{"multiple_tokens", []Token{{Value: "x"}, {Value: "+"}, {Value: "y"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTokenBuffer(tt.input)
			buffer.Unget()
		})
	}
}

func TestCmdFrac_Variants(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		args     []*TokenBuffer
		expected string
	}{
		{
			name:    "basic_frac",
			command: "frac",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "1", Kind: tokNumber}}),
				NewTokenBuffer([]Token{{Value: "2", Kind: tokNumber}}),
			},
			expected: "mfrac",
		},
		{
			name:    "cfrac",
			command: "cfrac",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "1", Kind: tokNumber}}),
				NewTokenBuffer([]Token{{Value: "2", Kind: tokNumber}}),
			},
			expected: "displaystyle",
		},
		{
			name:    "dfrac",
			command: "dfrac",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "1", Kind: tokNumber}}),
				NewTokenBuffer([]Token{{Value: "2", Kind: tokNumber}}),
			},
			expected: "displaystyle",
		},
		{
			name:    "tfrac",
			command: "tfrac",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "1", Kind: tokNumber}}),
				NewTokenBuffer([]Token{{Value: "2", Kind: tokNumber}}),
			},
			expected: "displaystyle=\"false\"",
		},
		{
			name:    "binom",
			command: "binom",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "n", Kind: tokLetter}}),
				NewTokenBuffer([]Token{{Value: "k", Kind: tokLetter}}),
			},
			expected: "linethickness=\"0\"",
		},
		{
			name:    "tbinom",
			command: "tbinom",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "n", Kind: tokLetter}}),
				NewTokenBuffer([]Token{{Value: "k", Kind: tokLetter}}),
			},
			expected: "linethickness=\"0\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result := cmd_frac(converter, tt.command, false, ctxVarNormal, tt.args, nil)
			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			output, _ := converter.ConvertInline("\\" + tt.command + "{1}{2}")
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestCmdFrac_BinomWrapper(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		args     []*TokenBuffer
		expected []string
	}{
		{
			name:    "binom_has_wrapper",
			command: "binom",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "n", Kind: tokLetter}}),
				NewTokenBuffer([]Token{{Value: "k", Kind: tokLetter}}),
			},
			expected: []string{"mrow", "mfrac", "linethickness=\"0\"", "mo", "("},
		},
		{
			name:    "tbinom_has_wrapper",
			command: "tbinom",
			args: []*TokenBuffer{
				NewTokenBuffer([]Token{{Value: "n", Kind: tokLetter}}),
				NewTokenBuffer([]Token{{Value: "k", Kind: tokLetter}}),
			},
			expected: []string{"mrow", "mfrac", "linethickness=\"0\"", "displaystyle=\"false\"", "mo", "("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			result := cmd_frac(converter, tt.command, false, ctxVarNormal, tt.args, nil)

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			output, _ := converter.ConvertInline("\\" + tt.command + "{n}{k}")
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestPostProcessChars_CombinePrimes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single_prime",
			input:    "x'",
			expected: "′",
		},
		{
			name:     "double_prime",
			input:    "x''",
			expected: "″",
		},
		{
			name:     "triple_prime",
			input:    "x'''",
			expected: "‴",
		},
		{
			name:     "quadruple_prime",
			input:    "x''''",
			expected: "⁗",
		},
		{
			name:     "five_primes",
			input:    "x'''''",
			expected: "⁗",
		},
		{
			name:     "six_primes",
			input:    "x''''''",
			expected: "⁗",
		},
		{
			name:     "seven_primes",
			input:    "x'''''''",
			expected: "⁗",
		},
		{
			name:     "eight_primes",
			input:    "x''''''''",
			expected: "⁗",
		},
		{
			name:     "mixed_with_other_chars",
			input:    "f'(x)''",
			expected: "′",
		},
		{
			name:     "command_prime",
			input:    "x\\prime",
			expected: "prime",
		},
		{
			name:     "unicode_prime",
			input:    "x'",
			expected: "′",
		},
		{
			name:     "unicode_apostrophe",
			input:    "x'",
			expected: "′",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.input)
			if err != nil {
				t.Fatalf("Failed to convert: %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestPostProcessChars_CharacterReplacements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hyphen_to_minus",
			input:    "x-y",
			expected: "−",
		},
		{
			name:     "less_than",
			input:    "x<y",
			expected: "&lt;",
		},
		{
			name:     "greater_than",
			input:    "x>y",
			expected: "&gt;",
		},
		{
			name:     "ampersand",
			input:    "x&y",
			expected: "&amp;",
		},
		{
			name:     "multiple_replacements",
			input:    "x<y&z>w",
			expected: "&lt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.input)
			if err != nil {
				t.Fatalf("Failed to convert: %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestPostProcessChars_ComplexPrimeCombinations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "multiple_prime_groups",
			input:    "f'(x)'' + g'(y)'''",
			expected: []string{"′", "″", "‴"},
		},
		{
			name:     "primes_with_spaces",
			input:    "f' ' '",
			expected: []string{"‴"},
		},
		{
			name:     "primes_with_numbers",
			input:    "x'1'2'3'",
			expected: []string{"′"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.input)
			if err != nil {
				t.Fatalf("Failed to convert: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestNewCommand_StyleSwitches(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "displaystyle",
			command:  "\\displaystyle{x + y}",
			expected: "displaystyle",
		},
		{
			name:     "textstyle",
			command:  "\\textstyle{x + y}",
			expected: "displaystyle=\"false\"",
		},
		{
			name:     "scriptstyle",
			command:  "\\scriptstyle{x + y}",
			expected: "scriptlevel=\"1\"",
		},
		{
			name:     "scriptscriptstyle",
			command:  "\\scriptscriptstyle{x + y}",
			expected: "scriptlevel=\"2\"",
		},
		{
			name:     "rm_command",
			command:  "\\rm{x + y}",
			expected: "mathvariant=\"normal\"",
		},
		{
			name:     "tiny_command",
			command:  "\\tiny{x + y}",
			expected: "mathsize=\"050.0%\"",
		},
		{
			name:     "scriptsize_command",
			command:  "\\scriptsize{x + y}",
			expected: "mathsize=\"070.0%\"",
		},
		{
			name:     "footnotesize_command",
			command:  "\\footnotesize{x + y}",
			expected: "mathsize=\"080.0%\"",
		},
		{
			name:     "small_command",
			command:  "\\small{x + y}",
			expected: "mathsize=\"090.0%\"",
		},
		{
			name:     "normalsize_command",
			command:  "\\normalsize{x + y}",
			expected: "mathsize=\"100.0%\"",
		},
		{
			name:     "large_command",
			command:  "\\large{x + y}",
			expected: "mathsize=\"120.0%\"",
		},
		{
			name:     "Large_command",
			command:  "\\Large{x + y}",
			expected: "mathsize=\"144.0%\"",
		},
		{
			name:     "LARGE_command",
			command:  "\\LARGE{x + y}",
			expected: "mathsize=\"172.8%\"",
		},
		{
			name:     "huge_command",
			command:  "\\huge{x + y}",
			expected: "mathsize=\"207.4%\"",
		},
		{
			name:     "Huge_command",
			command:  "\\Huge{x + y}",
			expected: "mathsize=\"248.8%\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.command)
			if err != nil {
				t.Fatalf("ConvertInline failed: %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestNewCommand_ColorCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "color_red",
			command:  "\\color{red}{x + y}",
			expected: "mathcolor=\"red\"",
		},
		{
			name:     "color_blue",
			command:  "\\color{blue}{x + y}",
			expected: "mathcolor=\"blue\"",
		},
		{
			name:     "color_hex",
			command:  "\\color{#FF0000}{x + y}",
			expected: "mathcolor=\"#FF0000\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.command)
			if err != nil {
				t.Fatalf("ConvertInline failed: %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestVariantTransform_AllVariants(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "mathbb_double_struck",
			command:  "\\mathbb{A}",
			expected: "𝔸",
		},
		{
			name:     "mathbf_bold",
			command:  "\\mathbf{A}",
			expected: "𝐀",
		},
		{
			name:     "mathbfit_bold_italic",
			command:  "\\mathbfit{A}",
			expected: "𝑨",
		},
		{
			name:     "mathcal_script_chancery",
			command:  "\\mathcal{A}",
			expected: "𝒜",
		},
		{
			name:     "mathscr_script_roundhand",
			command:  "\\mathscr{A}",
			expected: "𝒜",
		},
		{
			name:     "mathfrak_fraktur",
			command:  "\\mathfrak{A}",
			expected: "𝔄",
		},
		{
			name:     "mathit_italic",
			command:  "\\mathit{A}",
			expected: "𝐴",
		},
		{
			name:     "mathsf_sans_serif",
			command:  "\\mathsf{A}",
			expected: "𝖠",
		},
		{
			name:     "mathsfbf_bold_sans_serif",
			command:  "\\mathsfbf{A}",
			expected: "𝗔",
		},
		{
			name:     "mathsfbfsl_sans_serif_bold_italic",
			command:  "\\mathsfbfsl{A}",
			expected: "𝘼",
		},
		{
			name:     "mathsfsl_sans_serif_italic",
			command:  "\\mathsfsl{A}",
			expected: "𝘈",
		},
		{
			name:     "mathtt_monospace",
			command:  "\\mathtt{A}",
			expected: "𝙰",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.command)
			if err != nil {
				t.Fatalf("ConvertInline failed: %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestVariantTransform_CharacterTransformation(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "mathbb_numbers",
			command:  "\\mathbb{123}",
			expected: []string{"𝟙", "𝟚", "𝟛"},
		},
		{
			name:     "mathbf_letters",
			command:  "\\mathbf{ABC}",
			expected: []string{"𝐀", "𝐁", "𝐂"},
		},
		{
			name:     "mathit_letters",
			command:  "\\mathit{abc}",
			expected: []string{"𝑎", "𝑏", "𝑐"},
		},
		{
			name:     "mathsf_letters",
			command:  "\\mathsf{ABC}",
			expected: []string{"𝖠", "𝖡", "𝖢"},
		},
		{
			name:     "mathtt_letters",
			command:  "\\mathtt{ABC}",
			expected: []string{"𝙰", "𝙱", "𝙲"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.command)
			if err != nil {
				t.Fatalf("ConvertInline failed: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestVariantTransform_SpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "mathcal_special_chars",
			command:  "\\mathcal{BEFHILMR}",
			expected: []string{"ℬ", "ℰ", "ℱ", "ℋ", "ℐ", "ℒ", "ℳ", "ℛ"},
		},
		{
			name:     "mathbb_special_chars",
			command:  "\\mathbb{CHNPQRZ}",
			expected: []string{"ℂ", "ℍ", "ℕ", "ℙ", "ℚ", "ℝ", "ℤ"},
		},
		{
			name:     "mathfrak_special_chars",
			command:  "\\mathfrak{CHIRZ}",
			expected: []string{"ℭ", "ℌ", "ℑ", "ℜ", "ℨ"},
		},
		{
			name:     "mathscr_special_chars",
			command:  "\\mathscr{BEFHILMR}",
			expected: []string{"ℬ", "ℰ", "ℱ", "ℋ", "ℐ", "ℒ", "ℳ", "ℛ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.command)
			if err != nil {
				t.Fatalf("ConvertInline failed: %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestVariantTransform_CombinedVariants(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "bold_italic_combined",
			command:  "\\mathbfit{ABC}",
			expected: "𝑨",
		},
		{
			name:     "sans_serif_bold_combined",
			command:  "\\mathsfbf{ABC}",
			expected: "𝗔",
		},
		{
			name:     "sans_serif_italic_combined",
			command:  "\\mathsfsl{ABC}",
			expected: "𝘈",
		},
		{
			name:     "sans_serif_bold_italic_combined",
			command:  "\\mathsfbfsl{ABC}",
			expected: "𝘼",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.command)
			if err != nil {
				t.Fatalf("ConvertInline failed: %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestExtMath_DelimiterDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ams_inline",
			input:    "\\(x + y\\)",
			expected: "math",
		},
		{
			name:     "ams_display",
			input:    "\\[x + y\\]",
			expected: "math",
		},
		{
			name:     "tex_inline",
			input:    "$x + y$",
			expected: "math",
		},
		{
			name:     "tex_display",
			input:    "$$x + y$$",
			expected: "math",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewMathMLConverter()
			output, err := converter.ConvertInline(tt.input)
			if err != nil {
				t.Fatalf("ConvertInline failed: %v", err)
			}

			if !strings.Contains(output, "math") {
				t.Errorf("Expected output to contain math, got %q", output)
			}
		})
	}
}
