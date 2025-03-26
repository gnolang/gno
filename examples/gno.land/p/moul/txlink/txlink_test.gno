package txlink

import (
	"std"
	"testing"

	"gno.land/p/demo/urequire"
)

func TestCall(t *testing.T) {
	cd := std.ChainDomain()

	tests := []struct {
		fn    string
		args  []string
		want  string
		realm Realm
	}{
		{"foo", []string{"bar", "1", "baz", "2"}, "$help&func=foo&bar=1&baz=2", ""},
		{"testFunc", []string{"key", "value"}, "$help&func=testFunc&key=value", ""},
		{"noArgsFunc", []string{}, "$help&func=noArgsFunc", ""},
		{"oddArgsFunc", []string{"key"}, "$help&func=oddArgsFunc&error=odd+number+of+arguments", ""},
		{"foo", []string{"bar", "1", "baz", "2"}, "/r/lorem/ipsum$help&func=foo&bar=1&baz=2", Realm(cd + "/r/lorem/ipsum")},
		{"testFunc", []string{"key", "value"}, "/r/lorem/ipsum$help&func=testFunc&key=value", Realm(cd + "/r/lorem/ipsum")},
		{"noArgsFunc", []string{}, "/r/lorem/ipsum$help&func=noArgsFunc", Realm(cd + "/r/lorem/ipsum")},
		{"oddArgsFunc", []string{"key"}, "/r/lorem/ipsum$help&func=oddArgsFunc&error=odd+number+of+arguments", Realm(cd + "/r/lorem/ipsum")},
		{"foo", []string{"bar", "1", "baz", "2"}, "https://gno.world/r/lorem/ipsum$help&func=foo&bar=1&baz=2", "gno.world/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "https://gno.world/r/lorem/ipsum$help&func=testFunc&key=value", "gno.world/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "https://gno.world/r/lorem/ipsum$help&func=noArgsFunc", "gno.world/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "https://gno.world/r/lorem/ipsum$help&func=oddArgsFunc&error=odd+number+of+arguments", "gno.world/r/lorem/ipsum"},
		{"test", []string{"key", "hello world"}, "$help&func=test&key=hello+world", ""},
		{"test", []string{"key", "a&b=c"}, "$help&func=test&key=a%26b%3Dc", ""},
		{"test", []string{"key", ""}, "$help&func=test&key=", ""},
	}

	for _, tt := range tests {
		title := string(tt.realm) + "_" + tt.fn
		t.Run(title, func(t *testing.T) {
			got := tt.realm.Call(tt.fn, tt.args...)
			urequire.Equal(t, tt.want, got)
		})
	}
}
