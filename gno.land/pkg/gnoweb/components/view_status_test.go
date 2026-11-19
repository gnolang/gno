package components

import (
	"bytes"
	"strings"
	"testing"
)

// TestStatusPendingApprovalEscapesTheReason pins that the reason text is
// escaped when rendered.
//
// The reason arrives over the wire from a node, so gnoweb does not author it.
// It is escaped today only because StatusData.Body is a plain string and the
// views are parsed with html/template -- change Body to template.HTML for some
// other view's benefit and this becomes an injection point, silently.
func TestStatusPendingApprovalEscapesTheReason(t *testing.T) {
	const hostile = `<script>alert(1)</script>`

	var buf bytes.Buffer
	if err := StatusPendingApprovalComponent(hostile).Render(&buf); err != nil {
		t.Fatalf("render: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, hostile) {
		t.Fatalf("reason rendered as live markup: %s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Fatalf("reason not html-escaped: %s", out)
	}
}
