package gnoweb

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
)

func TestHasRenderFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		jdoc     *doc.JSONDocumentation
		expected bool
	}{
		{
			name: "valid Render function",
			jdoc: &doc.JSONDocumentation{
				Funcs: []*doc.JSONFunc{
					{
						Name:    "Render",
						Params:  []*doc.JSONField{{Name: "path", Type: "string"}},
						Results: []*doc.JSONField{{Name: "", Type: "string"}},
					},
				},
			},
			expected: true,
		},
		{
			name:     "nil doc",
			jdoc:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := HasRenderFunction(tt.jdoc)
			assert.Equal(t, tt.expected, result)
		})
	}
}
