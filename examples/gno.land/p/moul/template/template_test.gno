package template

import (
	"testing"
)

func TestTrimMarkers(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "no_trim",
			template: "{{23}} < {{45}}",
			expected: "23 < 45",
		},
		{
			name:     "trim_both",
			template: "{{23 -}} < {{- 45}}",
			expected: "23<45",
		},
		{
			name:     "trim_left_only",
			template: "{{23}} < {{- 45}}",
			expected: "23 <45",
		},
		{
			name:     "trim_right_only",
			template: "{{23 -}} < {{45}}",
			expected: "23< 45",
		},
		{
			name:     "multiple_spaces",
			template: "{{23 -}}   <   {{- 45}}",
			expected: "23<45",
		},
		{
			name:     "newlines",
			template: "{{23 -}}\n<\n{{- 45}}",
			expected: "23<45",
		},
		{
			name:     "helper_with_trim",
			template: "{{Bold \"hello\" -}} {{- Bold \"world\"}}",
			expected: "**hello****world**",
		},
		{
			name:     "variable_with_trim",
			template: "{{ .First -}} {{- .Second}}",
			data: map[string]interface{}{
				"First":  "Hello",
				"Second": "World",
			},
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer()
			got := r.Render(tt.template, tt.data)
			if got != tt.expected {
				t.Errorf("\nTemplate: %s\nExpected: %q\nGot: %q", tt.template, tt.expected, got)
			}
		})
	}
}

func TestNestedPlaceholders(t *testing.T) {
	template := `# Welcome, {{ .Name }}!
[Your Profile]({{ .ProfileURL }})`

	data := map[string]interface{}{
		"Name":       "Alice",
		"ProfileURL": "https://example.com/profile/Alice",
	}

	expected := `# Welcome, Alice!
[Your Profile](https://example.com/profile/Alice)`

	r := NewRenderer()
	got := r.Render(template, data)
	if got != expected {
		t.Errorf("\nExpected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestHelpers(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "bold_helper",
			template: `{{ Bold "hello" }}`,
			expected: "**hello**",
		},
		{
			name: "multiple_helpers",
			template: `{{ H1 "Section" }}
{{ Bold "Important" }}
{{ BulletItem "First" }}
{{ BulletItem "Second" }}`,
			expected: "# Section\n**Important**\n- First\n\n- Second\n",
		},
		{
			name:     "helper_with_quotes",
			template: `{{ Bold "\"quoted text\"" }}`,
			expected: `**\"quoted text\"**`,
		},
		{
			name:     "helper_with_spaces",
			template: `{{ Bold "  spaced  " }}`,
			expected: "**spaced**",
		},
		{
			name:     "concat_two_strings",
			template: `{{ concat "Hello" " World" }}`,
			expected: "Hello World",
		},
		{
			name:     "concat_multiple_strings",
			template: `{{ concat "Hello" " " "wonderful" " " "World" "!" }}`,
			expected: "Hello wonderful World!",
		},
		{
			name:     "concat_with_vars",
			template: `{{ concat .Greeting " " .Name "!" }}`,
			data: map[string]interface{}{
				"Greeting": "Hello",
				"Name":     "Alice",
			},
			expected: "Hello Alice!",
		},
		{
			name:     "concat_empty",
			template: `{{ concat }}`,
			expected: "",
		},
		{
			name:     "printf_basic",
			template: `{{ printf "Hello %s!" "World" }}`,
			expected: "Hello World!",
		},
		{
			name:     "printf_multiple_args",
			template: `{{ printf "%s has %s items" "Cart" "5" }}`,
			expected: "Cart has 5 items",
		},
		{
			name:     "string_helper",
			template: `{{ string "test" }}`,
			expected: "test",
		},
		{
			name:     "trim_helper",
			template: `{{ trim "  spaced  " }}`,
			expected: "spaced",
		},
		{
			name:     "index_numeric",
			template: `{{ index "array" "0" }}`,
			expected: "[0]",
		},
		{
			name:     "index_string",
			template: `{{ index "map" "key" }}`,
			expected: "[key]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer()
			got := r.Render(tt.template, tt.data)
			if got != tt.expected {
				t.Errorf("\nTemplate: %s\nExpected: %q\nGot: %q", tt.template, tt.expected, got)
			}
		})
	}
}

func TestHelperErrors(t *testing.T) {
	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "unknown_helper",
			template: `{{ Unknown "test" }}`,
		},
		{
			name:     "missing_argument",
			template: `{{ Bold }}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer()
			got := r.Render(tt.template, nil)
			// Should not panic and should return some result
			if got == "" {
				t.Error("Expected non-empty result even with errors")
			}
		})
	}
}

func TestControlFlow(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple_range",
			template: `{{ range .Items }}{{ . }}{{ end }}`,
			data: map[string]interface{}{
				"Items": []string{"a", "b", "c"},
			},
			expected: "abc",
		},
		{
			name:     "range_with_index",
			template: `{{ range .Items }}[{{ . }}]{{ end }}`,
			data: map[string]interface{}{
				"Items": []int{1, 2, 3},
			},
			expected: "[1][2][3]",
		},
		{
			name:     "range_with_field",
			template: "{{ range .Objects }}[{{ .Name }}]{{ end }}",
			data: map[string]interface{}{
				"Objects": []map[string]interface{}{
					{"Name": "obj1"},
					{"Name": "obj2"},
				},
			},
			expected: "[obj1][obj2]",
		},
		{
			name:     "range_with_helpers",
			template: `{{ range .Items }}{{ Bold . }}{{ end }}`,
			data: map[string]interface{}{
				"Items": []string{"a", "b", "c"},
			},
			expected: "**a****b****c**",
		},
		{
			name:     "Nested Range",
			template: "{{ range .Outer }}{{ range .Inner }}{{ . }}{{ end }}{{ end }}",
			data: map[string]interface{}{
				"Outer": []interface{}{
					map[string]interface{}{"Inner": []interface{}{"a", "b"}},
					map[string]interface{}{"Inner": []interface{}{"c", "d"}},
				},
			},
			expected: "abcd",
		},
		{
			name:     "Range in If",
			template: "{{ if .Show }}{{ range .Items }}{{ . }}{{ end }}{{ end }}",
			data: map[string]interface{}{
				"Show":  true,
				"Items": []interface{}{"1", "2", "3"},
			},
			expected: "123",
		},
		{
			name:     "If in Range",
			template: "{{ range .Items }}{{ if .Show }}{{ .Value }}{{ end }}{{ end }}",
			data: map[string]interface{}{
				"Items": []interface{}{
					map[string]interface{}{"Show": true, "Value": "x"},
					map[string]interface{}{"Show": false, "Value": "y"},
					map[string]interface{}{"Show": true, "Value": "z"},
				},
			},
			expected: "xz",
		},
		{
			name:     "If in If",
			template: "{{ if .Outer }}{{ if .Inner }}Yes{{ else }}No{{ end }}{{ else }}No{{ end }}",
			data: map[string]interface{}{
				"Outer": true,
				"Inner": true,
			},
			expected: "Yes",
		},
		{
			name:     "If in If in If",
			template: "{{ if .First }}{{ if .Second }}{{ if .Third }}All True{{ else }}Third False{{ end }}{{ else }}Second False{{ end }}{{ else }}First False{{ end }}",
			data: map[string]interface{}{
				"First":  true,
				"Second": true,
				"Third":  false,
			},
			expected: "Third False",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer()
			got := r.Render(tt.template, tt.data)
			if got != tt.expected {
				t.Errorf("\nTemplate: %s\nData: %v\nExpected: %q\nGot: %q", tt.template, tt.data, tt.expected, got)
			}
		})
	}
}

