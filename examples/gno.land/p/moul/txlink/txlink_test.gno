package txlink

import (
	"testing"

	"gno.land/p/demo/urequire"
)

func TestURL(t *testing.T) {
	tests := []struct {
		fn    string
		args  []string
		want  string
		realm Realm
	}{
		{"foo", []string{"bar", "1", "baz", "2"}, "?help&__func=foo&bar=1&baz=2", ""},
		{"testFunc", []string{"key", "value"}, "?help&__func=testFunc&key=value", ""},
		{"noArgsFunc", []string{}, "?help&__func=noArgsFunc", ""},
		{"oddArgsFunc", []string{"key"}, "?help&__func=oddArgsFunc", ""},
		{"foo", []string{"bar", "1", "baz", "2"}, "/r/lorem/ipsum?help&__func=foo&bar=1&baz=2", "gno.land/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "/r/lorem/ipsum?help&__func=testFunc&key=value", "gno.land/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "/r/lorem/ipsum?help&__func=noArgsFunc", "gno.land/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "/r/lorem/ipsum?help&__func=oddArgsFunc", "gno.land/r/lorem/ipsum"},
		{"foo", []string{"bar", "1", "baz", "2"}, "https://gno.world/r/lorem/ipsum?help&__func=foo&bar=1&baz=2", "gno.world/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "https://gno.world/r/lorem/ipsum?help&__func=testFunc&key=value", "gno.world/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "https://gno.world/r/lorem/ipsum?help&__func=noArgsFunc", "gno.world/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "https://gno.world/r/lorem/ipsum?help&__func=oddArgsFunc", "gno.world/r/lorem/ipsum"},
	}

	for _, tt := range tests {
		title := tt.fn
		t.Run(title, func(t *testing.T) {
			got := tt.realm.URL(tt.fn, tt.args...)
			urequire.Equal(t, tt.want, got)
		})
	}
}
