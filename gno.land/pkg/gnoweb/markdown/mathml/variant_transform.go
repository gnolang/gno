package mathml

func (n *MMLNode) transformByVariant(variant string) {
	rules, ok := transforms[variant]
	if !ok {
		return
	}
	chars := []rune(n.Text)
	for idx, char := range chars {
		if xform, ok := orphans[variant][char]; ok {
			chars[idx] = xform
		}
		for _, r := range rules {
			if char >= r.begin && char <= r.end {
				if xform, ok := r.exceptions[char]; ok {
					chars[idx] = xform
				} else {
					delta := r.delta
					chars[idx] += delta
				}
			}
		}
	}
	n.Text = string(chars)
}

func (n *MMLNode) set_variants_from_context(context parseContext) {
	var variant string
	switch isolateMathVariant(context) {
	case ctxVarNormal:
		n.Attrib["mathvariant"] = "normal"
		return
	case ctxVarBb:
		variant = "double-struck"
	case ctxVarBold:
		variant = "bold"
	case ctxVarBold | ctxVarItalic:
		variant = "bold-italic"
	case ctxVarScriptChancery, ctxVarScriptRoundhand:
		variant = "script"
	case ctxVarFrak:
		variant = "fraktur"
	case ctxVarItalic:
		variant = "italic"
	case ctxVarSans:
		variant = "sans-serif"
	case ctxVarSans | ctxVarBold:
		variant = "bold-sans-serif"
	case ctxVarSans | ctxVarBold | ctxVarItalic:
		variant = "sans-serif-bold-italic"
	case ctxVarSans | ctxVarItalic:
		variant = "sans-serif-italic"
	case ctxVarMono:
		variant = "monospace"
	case 0:
		return
	}
	n.transformByVariant(variant)
	var variationselector rune
	switch isolateMathVariant(context) {
	case ctxVarScriptChancery:
		variationselector = 0xfe00
		n.Attrib["class"] = "mathcal"
	case ctxVarScriptRoundhand:
		variationselector = 0xfe01
		n.Attrib["class"] = "mathscr"
	}
	if variationselector > 0 {
		temp := make([]rune, 0)
		for _, r := range n.Text {
			temp = append(temp, r, variationselector)
		}
		n.Text = string(temp)
	}
}

// "orphans" do not belong to any of the character ranges in the transforms table.
var orphans = map[string]map[rune]rune{
	"bold":                   {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': '𝛝', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '𝛛', '∇': '𝛁'},
	"bold-fraktur":           {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"bold-italic":            {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': '𝝑', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '𝝏', '∇': '𝜵'},
	"bold-sans-serif":        {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': '𝞋', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '𝞉', '∇': '𝝯'},
	"bold-script":            {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"double-struck":          {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': '𞺩', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"fraktur":                {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"initial":                {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': '𞸩', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"italic":                 {'ı': '𝚤', 'ȷ': '𝚥', 'ϑ': '𝜗', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '𝜕', '∇': '𝛻'},
	"looped":                 {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': '𞺉', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"monospace":              {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"sans-serif":             {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"sans-serif-bold-italic": {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': '𝟅', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '𝟃', '∇': '𝞩'},
	"sans-serif-italic":      {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"script":                 {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': 'ي', 'ڡ': 'ڡ', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"stretched":              {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': '𞹩', 'ڡ': '𞹾', 'ں': 'ں', '∂': '∂', '∇': '∇'},
	"tailed":                 {'ı': 'ı', 'ȷ': 'ȷ', 'ϑ': 'ϑ', 'ي': '𞹉', 'ڡ': 'ڡ', 'ں': '𞹝', '∂': '∂', '∇': '∇'},
}

type transformRule struct {
	begin      rune
	delta      rune
	end        rune
	exceptions map[rune]rune
}

var transforms = map[string][]transformRule{
	"bold": {
		{
			begin:      '0',
			delta:      120734,
			end:        '9',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'A',
			delta:      119743,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      119737,
			end:        'z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'Α',
			delta:      119575,
			end:        'Ω',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'α',
			delta:      119569,
			end:        'ω',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ϕ',
			end:   'ϖ',
			exceptions: map[rune]rune{
				'ϕ': '𝛟',
				'ϖ': '𝛡'},
		},
		{
			begin:      'Ϝ',
			delta:      119790,
			end:        'ϝ',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ϰ',
			end:   'ϱ',
			exceptions: map[rune]rune{
				'ϰ': '𝛞',
				'ϱ': '𝛠'},
		},
		{
			begin: 'ϴ',
			end:   'ϵ',
			exceptions: map[rune]rune{
				'ϴ': '𝚹',
				'ϵ': '𝛜',
			},
		},
	},
	"bold-fraktur": {
		{
			begin:      'A',
			delta:      120107,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      120101,
			end:        'z',
			exceptions: map[rune]rune{}},
	},
	"bold-italic": {
		{
			begin:      'A',
			delta:      119847,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      119841,
			end:        'z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'Α',
			delta:      119691,
			end:        'Ω',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'α',
			delta:      119685,
			end:        'ω',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ϕ',
			end:   'ϖ',
			exceptions: map[rune]rune{
				'ϕ': '𝝓',
				'ϖ': '𝝕'},
		},
		{
			begin: 'ϰ',
			end:   'ϱ',
			exceptions: map[rune]rune{
				'ϰ': '𝝒',
				'ϱ': '𝝔'},
		},
		{
			begin: 'ϴ',
			end:   'ϵ',
			exceptions: map[rune]rune{
				'ϴ': '𝜭',
				'ϵ': '𝝐',
			},
		},
	},
	"bold-sans-serif": {
		{
			begin:      '0',
			delta:      120764,
			end:        '9',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'A',
			delta:      120211,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      120205,
			end:        'z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'Α',
			delta:      119749,
			end:        'Ω',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'α',
			delta:      119743,
			end:        'ω',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ϕ',
			end:   'ϖ',
			exceptions: map[rune]rune{
				'ϕ': '𝞍',
				'ϖ': '𝞏'},
		},
		{
			begin: 'ϰ',
			end:   'ϱ',
			exceptions: map[rune]rune{
				'ϰ': '𝞌',
				'ϱ': '𝞎'},
		},
		{
			begin: 'ϴ',
			end:   'ϵ',
			exceptions: map[rune]rune{
				'ϴ': '𝝧',
				'ϵ': '𝞊',
			},
		},
	},
	"bold-script": {
		{
			begin:      'A',
			delta:      119951,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      119945,
			end:        'z',
			exceptions: map[rune]rune{}},
	},
	"double-struck": {
		{
			begin:      '0',
			delta:      120744,
			end:        '9',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'A',
			delta: 120055,
			end:   'Z',
			exceptions: map[rune]rune{
				'C': 'ℂ',
				'H': 'ℍ',
				'N': 'ℕ',
				'P': 'ℙ',
				'Q': 'ℚ',
				'R': 'ℝ',
				'Z': 'ℤ'},
		},
		{
			begin:      'a',
			delta:      120049,
			end:        'z',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ا',
			end:   'ب',
			exceptions: map[rune]rune{
				'ا': 'ا',
				'ب': '𞺡'},
		},
		{
			begin: 'ت',
			delta: 125058,
			end:   'غ',
			exceptions: map[rune]rune{
				'ت': '𞺵',
				'ث': '𞺶',
				'ج': '𞺢',
				'ح': '𞺧',
				'خ': '𞺷',
				'د': '𞺣',
				'ذ': '𞺸',
				'ز': '𞺦',
				'س': '𞺮',
				'ش': '𞺴',
				'ص': '𞺱',
				'ض': '𞺹',
				'ط': '𞺨',
				'ع': '𞺯',
				'غ': '𞺻'},
		},
		{
			begin: 'ف',
			delta: 125031,
			end:   'و',
			exceptions: map[rune]rune{
				'ف': '𞺰',
				'ق': '𞺲',
				'ك': 'ك',
				'ه': 'ه',
				'و': '𞺥',
			},
		},
	},
	"fraktur": {
		{
			begin: 'A',
			delta: 120003,
			end:   'Z',
			exceptions: map[rune]rune{
				'C': 'ℭ',
				'H': 'ℌ',
				'I': 'ℑ',
				'R': 'ℜ',
				'Z': 'ℨ'},
		},
		{
			begin:      'a',
			delta:      119997,
			end:        'z',
			exceptions: map[rune]rune{}},
	},
	"initial": {
		{
			begin: 'ا',
			end:   'ب',
			exceptions: map[rune]rune{
				'ا': 'ا',
				'ب': '𞸡'},
		},
		{
			begin: 'ت',
			delta: 0,
			end:   'غ',
			exceptions: map[rune]rune{
				'ت': '𞸵',
				'ث': '𞸶',
				'ج': '𞸢',
				'ح': '𞸧',
				'خ': '𞸷',
				'س': '𞸮',
				'ش': '𞸴',
				'ص': '𞸱',
				'ض': '𞸹',
				'ع': '𞸯',
				'غ': '𞸻'},
		},
		{
			begin: 'ف',
			delta: 124903,
			end:   'و',
			exceptions: map[rune]rune{
				'ف': '𞸰',
				'ق': '𞸲',
				'ه': '𞸤',
				'و': 'و',
			},
		},
	},
	"italic": {
		{
			begin:      'A',
			delta:      119795,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'a',
			delta: 119789,
			end:   'z',
			exceptions: map[rune]rune{
				'h': 'ℎ'},
		},
		{
			begin:      'Α',
			delta:      119633,
			end:        'Ω',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'α',
			delta:      119627,
			end:        'ω',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ϕ',
			end:   'ϖ',
			exceptions: map[rune]rune{
				'ϕ': '𝜙',
				'ϖ': '𝜛'},
		},
		{
			begin: 'ϰ',
			end:   'ϱ',
			exceptions: map[rune]rune{
				'ϰ': '𝜘',
				'ϱ': '𝜚'},
		},
		{
			begin: 'ϴ',
			end:   'ϵ',
			exceptions: map[rune]rune{
				'ϴ': '𝛳',
				'ϵ': '𝜖',
			},
		},
	},
	"looped": {
		{
			begin:      'ا',
			delta:      125017,
			end:        'ب',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ت',
			delta: 125026,
			end:   'غ',
			exceptions: map[rune]rune{
				'ت': '𞺕',
				'ث': '𞺖',
				'ج': '𞺂',
				'ح': '𞺇',
				'خ': '𞺗',
				'د': '𞺃',
				'ذ': '𞺘',
				'ز': '𞺆',
				'س': '𞺎',
				'ش': '𞺔',
				'ص': '𞺑',
				'ض': '𞺙',
				'ط': '𞺈',
				'ع': '𞺏',
				'غ': '𞺛'},
		},
		{
			begin: 'ف',
			delta: 124999,
			end:   'و',
			exceptions: map[rune]rune{
				'ف': '𞺐',
				'ق': '𞺒',
				'ك': 'ك',
				'ه': '𞺄',
				'و': '𞺅',
			},
		},
	},
	"monospace": {
		{
			begin:      '0',
			delta:      120774,
			end:        '9',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'A',
			delta:      120367,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      120361,
			end:        'z',
			exceptions: map[rune]rune{}},
	},
	"sans-serif": {
		{
			begin:      '0',
			delta:      120754,
			end:        '9',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'A',
			delta:      120159,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      120153,
			end:        'z',
			exceptions: map[rune]rune{}},
	},
	"sans-serif-bold-italic": {
		{
			begin:      'A',
			delta:      120315,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      120309,
			end:        'z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'Α',
			delta:      119807,
			end:        'Ω',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'α',
			delta:      119801,
			end:        'ω',
			exceptions: map[rune]rune{},
		},
		{
			begin: 'ϕ',
			end:   'ϖ',
			exceptions: map[rune]rune{
				'ϕ': '𝟇',
				'ϖ': '𝟉'},
		},
		{
			begin: 'ϰ',
			end:   'ϱ',
			exceptions: map[rune]rune{
				'ϰ': '𝟆',
				'ϱ': '𝟈'},
		},
		{
			begin: 'ϴ',
			end:   'ϵ',
			exceptions: map[rune]rune{
				'ϴ': '𝞡',
				'ϵ': '𝟄',
			},
		},
	},
	"sans-serif-italic": {
		{
			begin:      'A',
			delta:      120263,
			end:        'Z',
			exceptions: map[rune]rune{},
		},
		{
			begin:      'a',
			delta:      120257,
			end:        'z',
			exceptions: map[rune]rune{}},
	},
	"script": {
		{
			begin: 'A',
			delta: 119899,
			end:   'Z',
			exceptions: map[rune]rune{
				'B': 'ℬ',
				'E': 'ℰ',
				'F': 'ℱ',
				'H': 'ℋ',
				'I': 'ℐ',
				'L': 'ℒ',
				'M': 'ℳ',
				'R': 'ℛ'},
		},
		{
			begin: 'a',
			delta: 119893,
			end:   'z',
			exceptions: map[rune]rune{
				'e': 'ℯ',
				'g': 'ℊ',
				'o': 'ℴ',
			},
		},
	},
	"stretched": {
		{
			begin: 'ا',
			end:   'ب',
			exceptions: map[rune]rune{
				'ا': 'ا',
				'ب': '𞹡'},
		},
		{
			begin: 'ت',
			delta: 0,
			end:   'غ',
			exceptions: map[rune]rune{
				'ت': '𞹵',
				'ث': '𞹶',
				'ج': '𞹢',
				'ح': '𞹧',
				'خ': '𞹷',
				'س': '𞹮',
				'ش': '𞹴',
				'ص': '𞹱',
				'ض': '𞹹',
				'ط': '𞹨',
				'ظ': '𞹺',
				'ع': '𞹯',
				'غ': '𞹻'},
		},
		{
			begin: 'ف',
			delta: 124967,
			end:   'و',
			exceptions: map[rune]rune{
				'ف': '𞹰',
				'ق': '𞹲',
				'ل': 'ل',
				'ه': '𞹤',
				'و': 'و'},
		},
		{
			begin: 'ٮ',
			end:   'ٯ',
			exceptions: map[rune]rune{
				'ٮ': '𞹼',
				'ٯ': 'ٯ',
			},
		},
	},
	"tailed": {
		{
			begin: 'ت',
			delta: 0,
			end:   'غ',
			exceptions: map[rune]rune{
				'ج': '𞹂',
				'ح': '𞹇',
				'خ': '𞹗',
				'س': '𞹎',
				'ش': '𞹔',
				'ص': '𞹑',
				'ض': '𞹙',
				'ع': '𞹏',
				'غ': '𞹛'},
		},
		{
			begin: 'ف',
			delta: 0,
			end:   'و',
			exceptions: map[rune]rune{
				'ق': '𞹒',
				'ل': '𞹋',
				'ن': '𞹍'},
		},
		{
			begin: 'ٮ',
			end:   'ٯ',
			exceptions: map[rune]rune{
				'ٮ': 'ٮ',
				'ٯ': '𞹟',
			},
		},
	},
}
