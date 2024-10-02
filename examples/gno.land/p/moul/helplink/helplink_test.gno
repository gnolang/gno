package helplink

import (
	"testing"
)

func TestLink(t *testing.T) {
	tests := []struct {
		title string
		fn    string
		args  []string
		want  string
	}{
		{"Example", "foo", []string{"bar", "1", "baz", "2"}, "[Example](?help&__func=foo&bar=1&baz=2)"},
		{"Single Arg", "testFunc", []string{"key", "value"}, "[Single Arg](?help&__func=testFunc&key=value)"},
		{"No Args", "noArgsFunc", []string{}, "[No Args](?help&__func=noArgsFunc)"},
		{"Odd Args", "oddArgsFunc", []string{"key"}, "[Odd Args](?help&__func=oddArgsFunc)"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := Link(tt.title, tt.fn, tt.args...)
			if got != tt.want {
				t.Errorf("Link(%q, %q, %v) = %v; want %v", tt.title, tt.fn, tt.args, got, tt.want)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		fn   string
		args []string
		want string
	}{
		{"foo", []string{"bar", "1", "baz", "2"}, "?help&__func=foo&bar=1&baz=2"},
		{"testFunc", []string{"key", "value"}, "?help&__func=testFunc&key=value"},
		{"noArgsFunc", []string{}, "?help&__func=noArgsFunc"},
		{"oddArgsFunc", []string{"key"}, "?help&__func=oddArgsFunc"},
	}

	for _, tt := range tests {
		t.Run(tt.fn, func(t *testing.T) {
			got := URL(tt.fn, tt.args...)
			if got != tt.want {
				t.Errorf("URL(%q, %v) = %v; want %v", tt.fn, tt.args, got, tt.want)
			}
		})
	}
}
