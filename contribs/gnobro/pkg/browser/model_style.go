package browser

import "github.com/charmbracelet/glamour/ansi"

const (
	defaultListIndent      = 2
	defaultListLevelIndent = 4
	defaultMargin          = 2
)

// Catpuccin style: https://github.com/catppuccin/catppuccin
// XXX: update this with `gno` colors scheme
var CatppuccinStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#cad3f5"),
		},
		Margin: uintPtr(defaultMargin),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#cad3f5"),
			Italic: boolPtr(true),
		},
		Indent: uintPtr(1),
	},
	List: ansi.StyleList{
		LevelIndent: defaultListIndent,
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#cad3f5"),
			},
		},
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#cad3f5"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          " ",
			Suffix:          " ",
			BackgroundColor: stringPtr("#f0c6c6"),
			Color:           stringPtr("#181926"),
			Bold:            boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "● ",
			Color:  stringPtr("#f5a97f"),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "◉  ",
			Color:  stringPtr("#eed49f"),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "○   ",
			Color:  stringPtr("#a6da95"),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "◌    ",
			Color:  stringPtr("#7dc4e4"),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "‣    ",
			Color:  stringPtr("#b7bdf8"),
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Color:  stringPtr("#cad3f5"),
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Bold:  boolPtr(true),
		Color: stringPtr("#cad3f5"),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#6e738d"),
		Format: "\n--------\n",
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
		Color:       stringPtr("#cad3f5"),
	},
	Task: ansi.StyleTask{
		StylePrimitive: ansi.StylePrimitive{},
		Ticked:         "[✓] ",
		Unticked:       "[ ] ",
	},
	Link: ansi.StylePrimitive{
		Color:     stringPtr("#8aadf4"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#b7bdf8"),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#8aadf4"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#b7bdf8"),
		Format: "Image: {{.text}} →",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#ee99a0"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#1e2030"),
			},
			Margin: uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#cad3f5"),
			},
			Error: ansi.StylePrimitive{
				Color:           stringPtr("#cad3f5"),
				BackgroundColor: stringPtr("#ed8796"),
			},
			Comment: ansi.StylePrimitive{
				Color: stringPtr("#6e738d"),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#8aadf4"),
			},
			Keyword: ansi.StylePrimitive{
				Color: stringPtr("#c6a0f6"),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#c6a0f6"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#eed49f"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#eed49f"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#91d7e3"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#939ab7"),
			},
			Name: ansi.StylePrimitive{
				Color: stringPtr("#b7bdf8"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#f5a97f"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#c6a0f6"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#eed49f"),
			},
			NameClass: ansi.StylePrimitive{
				Color: stringPtr("#eed49f"),
			},
			NameConstant: ansi.StylePrimitive{
				Color: stringPtr("#eed49f"),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#f5bde6"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#8aadf4"),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color: stringPtr("#f5a97f"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#a6da95"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#f5bde6"),
			},
			GenericDeleted: ansi.StylePrimitive{
				Color: stringPtr("#ed8796"),
			},
			GenericEmph: ansi.StylePrimitive{
				Color:  stringPtr("#cad3f5"),
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				Color: stringPtr("#a6da95"),
			},
			GenericStrong: ansi.StylePrimitive{
				Color: stringPtr("#cad3f5"),
				Bold:  boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#91d7e3"),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#1e2030"),
			},
		},
	},
	Table: ansi.StyleTable{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{},
		},
		CenterSeparator: stringPtr("┼"),
		ColumnSeparator: stringPtr("│"),
		RowSeparator:    stringPtr("─"),
	},
	DefinitionDescription: ansi.StylePrimitive{
		BlockPrefix: "\n🠶 ",
	},
}

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }
