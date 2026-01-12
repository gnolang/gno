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
				Funcs: []*doc.JSONFunc{{
					Name:    "Render",
					Params:  []*doc.JSONField{{Type: "string"}},
					Results: []*doc.JSONField{{Type: "string"}},
				}},
			},
			expected: true,
		},
		{name: "nil doc", jdoc: nil, expected: false},
		{name: "empty funcs", jdoc: &doc.JSONDocumentation{}, expected: false},
		{
			name: "wrong param type",
			jdoc: &doc.JSONDocumentation{
				Funcs: []*doc.JSONFunc{{
					Name:    "Render",
					Params:  []*doc.JSONField{{Type: "int"}},
					Results: []*doc.JSONField{{Type: "string"}},
				}},
			},
			expected: false,
		},
		{
			name: "wrong result type",
			jdoc: &doc.JSONDocumentation{
				Funcs: []*doc.JSONFunc{{
					Name:    "Render",
					Params:  []*doc.JSONField{{Type: "string"}},
					Results: []*doc.JSONField{{Type: "int"}},
				}},
			},
			expected: false,
		},
		{
			name: "too many params",
			jdoc: &doc.JSONDocumentation{
				Funcs: []*doc.JSONFunc{{
					Name:    "Render",
					Params:  []*doc.JSONField{{Type: "string"}, {Type: "string"}},
					Results: []*doc.JSONField{{Type: "string"}},
				}},
			},
			expected: false,
		},
		{
			name: "no params",
			jdoc: &doc.JSONDocumentation{
				Funcs: []*doc.JSONFunc{{
					Name:    "Render",
					Params:  []*doc.JSONField{},
					Results: []*doc.JSONField{{Type: "string"}},
				}},
			},
			expected: false,
		},
		{
			name: "wrong function name",
			jdoc: &doc.JSONDocumentation{
				Funcs: []*doc.JSONFunc{{
					Name:    "NotRender",
					Params:  []*doc.JSONField{{Type: "string"}},
					Results: []*doc.JSONField{{Type: "string"}},
				}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, HasRenderFunction(tt.jdoc))
		})
	}
}
