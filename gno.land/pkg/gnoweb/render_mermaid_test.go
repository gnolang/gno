package gnoweb

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTMLRendererMermaid(t *testing.T) {
	t.Parallel()

	render := func(enabled bool) string {
		cfg := NewDefaultRenderConfig()
		cfg.Mermaid = enabled
		r := NewHTMLRenderer(log.NewTestingLogger(t), cfg, nil)
		var out bytes.Buffer
		_, err := r.RenderRealm(&out, &weburl.GnoURL{Path: "/r/test"}, []byte("```mermaid\nsequenceDiagram\nA->>B: hi\n```\n"), RealmRenderContext{})
		require.NoError(t, err)
		return out.String()
	}

	assert.Contains(t, render(true), `<div class="gno-mermaid"><div class="gno-mermaid__scale" data-copy-exclude><svg`)
	assert.NotContains(t, render(false), "<svg")
}
