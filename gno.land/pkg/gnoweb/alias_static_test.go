package gnoweb_test

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/stretchr/testify/assert"
)

func TestNewStaticAlias(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		content     string
		title       string
		description string
		body        string
	}{
		{
			name:        "front matter is lifted off the page",
			content:     "---\ntitle: About\ndescription: What gno.land is.\n---\n\n# About\n\nBody.\n",
			title:       "About",
			description: "What gno.land is.",
			body:        "# About\n\nBody.\n",
		},
		{
			name:    "a page without front matter is untouched",
			content: "# About\n\nBody.\n",
			body:    "# About\n\nBody.\n",
		},
		{
			name:    "an unterminated block stays part of the page",
			content: "---\ntitle: About\n\n# About\n",
			body:    "---\ntitle: About\n\n# About\n",
		},
		{
			name:    "a value keeps its colons and loses its quotes",
			content: "---\ntitle: \"Gno: a language\"\n---\nBody.\n",
			title:   "Gno: a language",
			body:    "Body.\n",
		},
		{
			name:    "an unknown key is ignored",
			content: "---\nauthor: someone\n---\nBody.\n",
			body:    "Body.\n",
		},
		{
			name:        "an overlong description is capped like a derived one",
			content:     "---\ndescription: " + strings.Repeat("word ", 200) + "\n---\nBody.\n",
			description: strings.TrimSpace(strings.Repeat("word ", 32)) + "…",
			body:        "Body.\n",
		},
		{
			name:    "an overlong title is capped like a derived one",
			content: "---\ntitle: " + strings.Repeat("word ", 50) + "\n---\nBody.\n",
			title:   strings.TrimSpace(strings.Repeat("word ", 12)) + "…",
			body:    "Body.\n",
		},
		{
			// A page may open on a rule and draw another further down; the
			// prose between them is content, not a header to strip.
			name:    "prose between two thematic breaks is not front matter",
			content: "---\n\nIntro paragraph.\n\n---\n\nMore.\n",
			body:    "---\n\nIntro paragraph.\n\n---\n\nMore.\n",
		},
		{
			name:    "comments and continuation lines stay front matter",
			content: "---\n# shipped with the node\ntitle: About\nnotes: >\n  folded text\n---\nBody.\n",
			title:   "About",
			body:    "Body.\n",
		},
		{
			// Only a lowercase key makes a line YAML; prose, a URL, a heading
			// or indented code between two rules stays on the page.
			name:    "colon prose between two thematic breaks is not front matter",
			content: "---\n\nNote: read this.\n\n---\n\nMore.\n",
			body:    "---\n\nNote: read this.\n\n---\n\nMore.\n",
		},
		{
			name:    "a URL between two thematic breaks is not front matter",
			content: "---\n\nSee https://gno.land for more.\n\n---\n\nMore.\n",
			body:    "---\n\nSee https://gno.land for more.\n\n---\n\nMore.\n",
		},
		{
			name:    "a line opening on a URL between two thematic breaks is not front matter",
			content: "---\n\nhttps://gno.land/r/gnoland/blog\n\n---\n\nMore.\n",
			body:    "---\n\nhttps://gno.land/r/gnoland/blog\n\n---\n\nMore.\n",
		},
		{
			name:    "a line opening on a mailto link between two thematic breaks is not front matter",
			content: "---\n\nmailto:hello@gno.land\n\n---\n\nMore.\n",
			body:    "---\n\nmailto:hello@gno.land\n\n---\n\nMore.\n",
		},
		{
			name:    "a key padded before its colon still sets its field",
			content: "---\ntitle : About\n---\nBody.\n",
			title:   "About",
			body:    "Body.\n",
		},
		{
			name:    "a key with an empty value ends the line on its colon",
			content: "---\ntitle: About\ndescription:\n---\nBody.\n",
			title:   "About",
			body:    "Body.\n",
		},
		{
			name:    "a key followed by a tab is a key",
			content: "---\ntitle:\tAbout\n---\nBody.\n",
			title:   "About",
			body:    "Body.\n",
		},
		{
			name:    "a heading between two thematic breaks is not front matter",
			content: "---\n\n## Section\n\n    indented code\n\n---\n\nMore.\n",
			body:    "---\n\n## Section\n\n    indented code\n\n---\n\nMore.\n",
		},
		{
			name:    "a top-level list stays front matter",
			content: "---\ntitle: About\ntags:\n- gno\n---\nBody.\n",
			title:   "About",
			body:    "Body.\n",
		},
		{
			name:    "a title is not taken from a block that is not front matter",
			content: "---\ntitle: About\nIntro.\n---\nBody.\n",
			body:    "---\ntitle: About\nIntro.\n---\nBody.\n",
		},
		{
			name:        "a folded description joins its lines with spaces",
			content:     "---\ntitle: About\ndescription: >\n  What gno.land is\n  and why it exists.\n---\nBody.\n",
			title:       "About",
			description: "What gno.land is and why it exists.",
			body:        "Body.\n",
		},
		{
			name:        "a folded description with a chomping indicator",
			content:     "---\ndescription: >-\n  What gno.land is\n\n  and why it exists.\ntitle: About\n---\nBody.\n",
			title:       "About",
			description: "What gno.land is and why it exists.",
			body:        "Body.\n",
		},
		{
			// A head is one line, so a literal block's newlines fold too.
			name:        "a literal description is put on one line",
			content:     "---\ndescription: |\n  What gno.land is.\n  Why it exists.\n---\nBody.\n",
			description: "What gno.land is. Why it exists.",
			body:        "Body.\n",
		},
		{
			name:        "a wrapped plain description keeps every line",
			content:     "---\ndescription: What gno.land is\n  and why it exists.\ntitle: About\n---\nBody.\n",
			title:       "About",
			description: "What gno.land is and why it exists.",
			body:        "Body.\n",
		},
		{
			name:    "a wrapped quoted title loses its quotes",
			content: "---\ntitle: \"Gno: a\n  language\"\n---\nBody.\n",
			title:   "Gno: a language",
			body:    "Body.\n",
		},
		{
			name:        "front matter with CRLF line endings is lifted off the page",
			content:     "---\r\ntitle: About\r\ndescription: What gno.land is.\r\n---\r\n\r\n# About\r\n",
			title:       "About",
			description: "What gno.land is.",
			body:        "# About\n",
		},
		{
			name:    "front matter after a byte order mark is lifted off the page",
			content: "\uFEFF---\ntitle: About\n---\nBody.\n",
			title:   "About",
			body:    "Body.\n",
		},
		{
			name:    "quotes inside a plain title are kept",
			content: "---\ntitle: Learn \"Gno\"\n---\nBody.\n",
			title:   "Learn \"Gno\"",
			body:    "Body.\n",
		},
		{
			name:    "a quoted title loses only its wrapping pair",
			content: "---\ntitle: \"Learn \"Gno\"\"\n---\nBody.\n",
			title:   "Learn \"Gno\"",
			body:    "Body.\n",
		},
		{
			name:    "a title that only ends on a quote keeps it",
			content: "---\ntitle: 5' tall\n---\nBody.\n",
			title:   "5' tall",
			body:    "Body.\n",
		},
		{
			name:    "a thematic break is not front matter",
			content: "---\n\nBody.\n",
			body:    "---\n\nBody.\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := gnoweb.NewStaticAlias(tc.content)
			assert.Equal(t, gnoweb.StaticMarkdown, target.Kind)
			assert.Equal(t, tc.title, target.Title)
			assert.Equal(t, tc.description, target.Description)
			assert.Equal(t, tc.body, target.Value)
		})
	}
}
