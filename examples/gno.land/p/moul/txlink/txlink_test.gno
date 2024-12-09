package txlink

import (
	"testing"

	"gno.land/p/demo/urequire"
)

func TestCall(t *testing.T) {
	tests := []struct {
		fn    string
		args  []string
		want  string
		realm Realm
	}{
		// relative
		{"foo", []string{"bar", "1", "baz", "2"}, "$help&func=foo&bar=1&baz=2", ""},
		{"testFunc", []string{"key", "value"}, "$help&func=testFunc&key=value", ""},
		{"noArgsFunc", []string{}, "$help&func=noArgsFunc", ""},
		{"oddArgsFunc", []string{"key"}, "$help&func=oddArgsFunc", ""},
		// gno.land
		{"foo", []string{"bar", "1", "baz", "2"}, "https://gno.land/r/lorem/ipsum$help&func=foo&bar=1&baz=2", "gno.land/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "https://gno.land/r/lorem/ipsum$help&func=testFunc&key=value", "gno.land/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "https://gno.land/r/lorem/ipsum$help&func=noArgsFunc", "gno.land/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "https://gno.land/r/lorem/ipsum$help&func=oddArgsFunc", "gno.land/r/lorem/ipsum"},
		// gno.land
		{"foo", []string{"bar", "1", "baz", "2"}, "/r/lorem/ipsum$help&func=foo&bar=1&baz=2", "tests.gno.land/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "/r/lorem/ipsum$help&func=testFunc&key=value", "tests.gno.land/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "/r/lorem/ipsum$help&func=noArgsFunc", "tests.gno.land/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "/r/lorem/ipsum$help&func=oddArgsFunc", "tests.gno.land/r/lorem/ipsum"},
		// gno.world
		{"foo", []string{"bar", "1", "baz", "2"}, "https://gno.world/r/lorem/ipsum$help&func=foo&bar=1&baz=2", "gno.world/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "https://gno.world/r/lorem/ipsum$help&func=testFunc&key=value", "gno.world/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "https://gno.world/r/lorem/ipsum$help&func=noArgsFunc", "gno.world/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "https://gno.world/r/lorem/ipsum$help&func=oddArgsFunc", "gno.world/r/lorem/ipsum"},
	}

	for _, tt := range tests {
		title := string(tt.realm) + "-" + tt.fn
		t.Run(title, func(t *testing.T) {
			got := tt.realm.Call(tt.fn, tt.args...)
			urequire.Equal(t, tt.want, got)
		})
	}
}
