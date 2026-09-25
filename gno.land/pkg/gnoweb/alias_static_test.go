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
