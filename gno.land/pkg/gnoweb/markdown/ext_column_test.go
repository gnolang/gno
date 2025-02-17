package markdown

import (
	"testing"

	"github.com/yuin/goldmark"
)

func TestExtColumn_Valid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name  string
		Input string
	}{
		{
			Name: "basic",
			Input: `
<gno-columns>
## Title 1

content 1

## Title 2

content 2

## Title 3

content 3
</gno-columns>

`,
		},

		{
			Name: "empty heading",
			Input: `
<gno-columns>
## Title 1

content 1

##

content 2

## Title 3

content 3
</gno-columns>

`,
		},

		{
			Name: "shortcut separator",
			Input: `
:::
## Title 1

content 1

## Title 2

content 2

## Title 3

content 3
:::
`,
		},

		{
			Name: "sticky header",
			Input: `
<gno-columns>
## Title 1
content 1
## Title 2
content 2
## Title 3
content 3
</gno-columns>
`,
		},

		{
			Name: "multi level",
			Input: `
<gno-columns>
# Title 1
content 1
## Title 2
content 2
### Title 3
content 3
#### Title 4
content 4
</gno-columns>
`,
		},

		{
			Name: "multi level 2",
			Input: `
<gno-columns>
## Title 1
content 1
# Title 2
content 2
## Title 3
content 3
# Title 4
content 4
</gno-columns>
`,
		},

		{
			"maximum level heading", `
<gno-columns>
###### Title 1
content 1
## Title 2
content 2
###### Title 3
content 3
## Title 4
content 4
</gno-columns>
`,
		},

		{
			Name: "no column",
			Input: `
<gno-columns>
</gno-columns>
`,
		},
	}

	m := goldmark.New()
	Column.Extend(m)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			testGoldamarkGoldenOuput(t, m, tc.Input)
		})
	}
}

// Invalid format should still have a predictable output
func TestExtColumn_Invalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name  string
		Input string
	}{
		{
			Name:  "inline tag",
			Input: `<gno-columns></gno-columns>`,
		},

		{
			Name:  "inline shortcut tag",
			Input: `::: :::`,
		},

		{
			Name: "unfinished column",
			Input: `
<gno-columns>
## Title 1
content 1
## Title 2
content 2
`,
		},

		{
			"unstarted column", `
## Title 1
content 1
## Title 2
content 2
</gno-columns>
`,
		},

		{
			"beyond maximum level heading", `
<gno-columns>
####### Title 1
content 1
## Title 2
content 2
####### Title 3
content 3
## Title 4
content 4
</gno-columns>
`,
		},
	}

	m := goldmark.New()
	Column.Extend(m)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			testGoldamarkGoldenOuput(t, m, tc.Input)
		})
	}
}
