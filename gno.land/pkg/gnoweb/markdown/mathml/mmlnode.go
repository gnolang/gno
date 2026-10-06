package mathml

import (
	"slices"
	"strings"
)

// writeEscaped writes s to w, escaping <, >, " and &. This makes the output
// safe both as element content and inside a double-quoted attribute value.
// Node text always holds literal characters, never character references, so
// an entity typed by the author (\text{&lt;b&gt;}) displays as typed.
func writeEscaped(w *strings.Builder, s string) {
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '<':
			w.WriteString("&lt;")
		case '>':
			w.WriteString("&gt;")
		case '"':
			w.WriteString("&#34;")
		case '&':
			w.WriteString("&amp;")
		default:
			w.WriteByte(c)
		}
	}
}

// isAttrName reports whether s is a safe XML attribute name. Attribute names
// are not user-controlled today; this is defense in depth.
func isAttrName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-' || c == '_' || c == ':' || i > 0 && c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// An MMLNode is the representation of a MathML tag or tree.
type MMLNode struct {
	Tok        Token             // the token from which this node was created
	Text       string            // the <tag>text</tag> enclosed in the Tag.
	Tag        string            // the value of the MathML tag, e.g. <mrow>, <msqrt>, <mo>....
	Option     string            // container for any options that may be passed and processed for a tex command
	Properties NodeProperties    // bitfield of NodeProperties
	Attrib     map[string]string // key value pairs of XML attributes; nil until one is set
	Children   []*MMLNode        // ordered list of child MathML elements
}

func makeMMLError() *MMLNode {
	mml := NewMMLNode("math")
	e := NewMMLNode("merror")
	t := NewMMLNode("mtext")
	t.Text = "invalid math input"
	e.Children = append(e.Children, t)
	mml.Children = append(mml.Children, e)
	return mml
}

// NewMMLNode allocates a new MathML node.
// The first optional argument sets the value of Tag.
// The second optional argument sets the value of Text.
func NewMMLNode(opt ...string) *MMLNode {
	tagText := make([]string, 2)
	for i, o := range opt {
		if i > 2 {
			break
		}
		tagText[i] = o
	}
	return &MMLNode{
		Tag:      tagText[0],
		Text:     tagText[1],
		Children: make([]*MMLNode, 0),
	}
}

// set the attribute name to "true"
func (n *MMLNode) SetTrue(name string) *MMLNode {
	return n.SetAttr(name, "true")
}

// set the attribute name to "false"
func (n *MMLNode) SetFalse(name string) *MMLNode {
	return n.SetAttr(name, "false")
}

// remove the attribute entirely
func (n *MMLNode) UnsetAttr(name string) *MMLNode {
	delete(n.Attrib, name)
	return n
}

// SetAttr sets the attribute name to "value" and returns the same MMLNode.
func (n *MMLNode) SetAttr(name, value string) *MMLNode {
	if n.Attrib == nil {
		// Most nodes have no attribute: the map is allocated on first use.
		n.Attrib = make(map[string]string, 2)
	}
	n.Attrib[name] = value
	return n
}

// AddClass adds class to the classes of n and returns n.
func (n *MMLNode) AddClass(class string) *MMLNode {
	if c := n.Attrib["class"]; c != "" {
		class = c + " " + class
	}
	return n.SetAttr("class", class)
}

func (n *MMLNode) SetProps(p NodeProperties) *MMLNode {
	n.Properties = p
	return n
}

func (n *MMLNode) AddProps(p NodeProperties) *MMLNode {
	n.Properties |= p
	return n
}

// If a property corresponds to an attribute in the final XML representation, set it here.
func (n *MMLNode) setAttribsFromProperties() {
	if n.Properties&propLargeop > 0 {
		n.SetTrue("largeop")
	}
	if n.Properties&propMovablelimits > 0 {
		n.SetTrue("movablelimits")
	}
	if n.Properties&propStretchy > 0 {
		n.SetTrue("stretchy")
	}
}

// AppendChild appends the child (or children) provided to the children of n.
func (n *MMLNode) AppendChild(child ...*MMLNode) *MMLNode {
	n.Children = append(n.Children, child...)
	return n
}

// AppendNew creates a new MMLNode and appends it to the children of n. The newly created MMLNode is returned.
func (n *MMLNode) AppendNew(opt ...string) *MMLNode {
	newnode := NewMMLNode(opt...)
	n.Children = append(n.Children, newnode)
	return newnode
}

// Write the MMLNode to the strings.Builder w, on one line.
func (n *MMLNode) Write(w *strings.Builder) {
	if n == nil {
		return
	}
	if n.Properties&propNonprint > 0 {
		return
	}
	var tag string
	if len(n.Tag) > 0 {
		tag = n.Tag
	} else {
		return
	}
	w.WriteRune('<')
	w.WriteString(tag)

	// Sort attributes for deterministic output. Nodes have few attributes:
	// sorting them in a stack buffer saves an allocation per node.
	var buf [8]string
	keys := buf[:0]
	for key := range n.Attrib {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if !isAttrName(key) {
			continue
		}
		w.WriteRune(' ')
		w.WriteString(key)
		w.WriteString(`="`)
		writeEscaped(w, n.Attrib[key])
		w.WriteRune('"')
	}
	w.WriteRune('>')
	if !self_closing_tags[tag] {
		if len(n.Children) == 0 && tag == "merror" && n.Text != "" {
			// merror lays out its children like mrow: browsers draw no
			// text placed directly inside it, so the text goes in an mtext.
			w.WriteString("<mtext>")
			writeEscaped(w, n.Text)
			w.WriteString("</mtext>")
		} else if len(n.Children) == 0 {
			writeEscaped(w, n.Text)
		} else {
			for _, child := range n.Children {
				child.Write(w)
			}
		}
	}
	w.WriteString("</")
	w.WriteString(tag)
	w.WriteRune('>')
}
