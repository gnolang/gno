package markdown

import (
	"testing"

	"github.com/yuin/goldmark"
)

// TestExtColumn_Valid tests the valid cases for gno-columns markdown extension.
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
			Name: "maximum level heading",
			Input: `
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
		{
			Name: "inline tags in title",
			Input: `
<gno-columns>
## [title 1](#)
content 1
## [title 2](#)
content 2
## **title 3**
content 3
</gno-columns>
`,
		},
		{
			Name: "image starter",
			Input: `
<gno-columns>
##

![img](http://abc.yz/image)

content 1

##

![img](http://abc.yz/image)

content 2

</gno-columns>
`,
		},
		{
			Name: "mix of tags",
			Input: `
:::
## title 1

content 1

## title 2

content 2

</gno-columns>
`,
		},
	}

	m := goldmark.New()
	Column.Extend(m)

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			testGoldamarkGoldenOuput(t, m, tc.Input)
		})
	}
}

// TestExtColumn_Invalid tests the invalid cases for gno-columns markdown extension.
// Invalid format should still have a predictable output.
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
			Name: "unstarted column",
			Input: `
## Title 1
content 1
## Title 2
content 2
</gno-columns>
`,
		},
		{
			Name: "top level intermediary content",
			Input: `
<gno-columns>

content 1
content 2

## Title 1
content 3
## Title 2
content 4
</gno-columns>
`,
		},
		{
			Name: "beyond maximum level heading",
			Input: `
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
		{
			Name: "heading in list",
			Input: `
<gno-columns>
- ## title 1
- ## title 2
- ## title 3
</gno-columns>
`,
		},
		{
			Name: "scopping columns",
			Input: `
<gno-columns>
## title 1

content 1

## title 2

<gno-columns>
## sub-title 1

content

## sub-title 2

content 2
</gno-columns>

## title 3

content 3
</gno-columns>
`,
		},
		{
			Name: "mix of scopped tags",
			Input: `
<gno-columns>
## title 1

content 1

## title 2

:::
## sub-title 1

content

## sub-title 2

content 2
:::

## title 3

content 3
</gno-columns>
`,
		},
	}

	m := goldmark.New()
	Column.Extend(m)

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			testGoldamarkGoldenOuput(t, m, tc.Input)
		})
	}
}
