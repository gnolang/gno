package markdown

import (
	"testing"

	"github.com/yuin/goldmark"
)

// The realm gm is built from NewGnoExtension; the emphasis guard must be wired
// there.
func TestEmphasisGuard_GnoExtensionInstance(t *testing.T) {
	assertEmphasisGuardActive(t, goldmark.New(goldmark.WithExtensions(NewGnoExtension())))
}

// The inner <gno-foreign> instance renders attacker-controlled body bytes.
func TestEmphasisGuard_InnerForeignInstance(t *testing.T) {
	assertEmphasisGuardActive(t, buildInnerForeignMarkdown(nil))
}
