package markdown

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestMeta(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Meta,
		),
	)
	source := `---
Title: goldmark-meta
Summary: Add YAML metadata to the document
Tags:
    - markdown
    - goldmark
---

# Hello goldmark-meta
`

	var buf bytes.Buffer
	context := parser.NewContext()
	if err := markdown.Convert([]byte(source), &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	metaData := Get(context)
	title := metaData["Title"]
	s, ok := title.(string)
	if !ok {
		t.Error("Title not found in meta data or is not a string")
	}
	if s != "goldmark-meta" {
		t.Errorf("Title must be %s, but got %v", "goldmark-meta", s)
	}
	if buf.String() != "<h1>Hello goldmark-meta</h1>\n" {
		t.Errorf("should render '<h1>Hello goldmark-meta</h1>', but '%s'", buf.String())
	}
	tags, ok := metaData["Tags"].([]interface{})
	if !ok {
		t.Error("Tags not found in meta data or is not a slice")
	}
	if len(tags) != 2 {
		t.Error("Tags must be a slice that has 2 elements")
	}
	if tags[0] != "markdown" {
		t.Errorf("Tag#1 must be 'markdown', but got %s", tags[0])
	}
	if tags[1] != "goldmark" {
		t.Errorf("Tag#2 must be 'goldmark', but got %s", tags[1])
	}
}

func TestMetaError(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Meta,
		),
	)
	source := `---
Title: goldmark-meta
Summary: Add YAML metadata to the document
Tags:
  - : {
  }
    - markdown
    - goldmark
---

# Hello goldmark-meta
`

	var buf bytes.Buffer
	context := parser.NewContext()
	if err := markdown.Convert([]byte(source), &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	if buf.String() != `Title: goldmark-meta
Summary: Add YAML metadata to the document
Tags:
  - : {
  }
    - markdown
    - goldmark
<!-- yaml: line 3: did not find expected key -->
<h1>Hello goldmark-meta</h1>
` {
		t.Error("invalid error output")
	}

	v, err := TryGet(context)
	if err == nil {
		t.Error("error should not be nil")
	}
	if v != nil {
		t.Error("data should be nil when there are errors")
	}
}

func TestMetaStoreInDocument(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			New(
				WithStoresInDocument(),
			),
		),
	)
	source := `---
Title: goldmark-meta
Summary: Add YAML metadata to the document
Tags:
    - markdown
    - goldmark
---
`

	document := markdown.Parser().Parse(text.NewReader([]byte(source)))
	metaData := document.OwnerDocument().Meta()
	title := metaData["Title"]
	s, ok := title.(string)
	if !ok {
		t.Error("Title not found in meta data or is not a string")
	}
	if s != "goldmark-meta" {
		t.Errorf("Title must be %s, but got %v", "goldmark-meta", s)
	}
	tags, ok := metaData["Tags"].([]interface{})
	if !ok {
		t.Error("Tags not found in meta data or is not a slice")
	}
	if len(tags) != 2 {
		t.Error("Tags must be a slice that has 2 elements")
	}
	if tags[0] != "markdown" {
		t.Errorf("Tag#1 must be 'markdown', but got %s", tags[0])
	}
	if tags[1] != "goldmark" {
		t.Errorf("Tag#2 must be 'goldmark', but got %s", tags[1])
	}
}
