package helplink

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/urequire/v0"
)

func TestFunc(t *testing.T) {
	cd := runtime.ChainDomain()
	tests := []struct {
		title     string
		fn        string
		args      []string
		want      string
		realm_XXX Realm
	}{
		{"Example", "foo", []string{"bar", "1", "baz", "2"}, "[Example](/p/moul/helplink$help&func=foo&bar=1&baz=2)", ""},
		{"Realm Example", "foo", []string{"bar", "1", "baz", "2"}, "[Realm Example](/r/lorem/ipsum$help&func=foo&bar=1&baz=2)", Realm(cd + "/r/lorem/ipsum")},
		{"Single Arg", "testFunc", []string{"key", "value"}, "[Single Arg](/p/moul/helplink$help&func=testFunc&key=value)", ""},
		{"No Args", "noArgsFunc", []string{}, "[No Args](/p/moul/helplink$help&func=noArgsFunc)", ""},
		{"Odd Args", "oddArgsFunc", []string{"key"}, "[Odd Args](/p/moul/helplink$help&func=oddArgsFunc&error=odd+number+of+arguments)", ""},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if tt.fn == "oddArgsFunc" {
				defer func() {
					if r := recover(); r != nil {
						if r != "odd number of arguments" {
							t.Errorf("expected panic with message 'odd number of arguments', got: %v", r)
						}
					} else {
						t.Error("expected panic for odd number of arguments, but did not panic")
					}
				}()
			}
			got := tt.realm_XXX.Func(tt.title, tt.fn, tt.args...)
			urequire.Equal(t, tt.want, got)
		})
	}
}

func TestFuncURL(t *testing.T) {
	cd := runtime.ChainDomain()
	tests := []struct {
		fn        string
		args      []string
		want      string
		realm_XXX Realm
	}{
		{"foo", []string{"bar", "1", "baz", "2"}, "/p/moul/helplink$help&func=foo&bar=1&baz=2", ""},
		{"testFunc", []string{"key", "value"}, "/p/moul/helplink$help&func=testFunc&key=value", ""},
		{"noArgsFunc", []string{}, "/p/moul/helplink$help&func=noArgsFunc", ""},
		{"oddArgsFunc", []string{"key"}, "/p/moul/helplink$help&func=oddArgsFunc&error=odd+number+of+arguments", ""},
		{"foo", []string{"bar", "1", "baz", "2"}, "/r/lorem/ipsum$help&func=foo&bar=1&baz=2", Realm(cd + "/r/lorem/ipsum")},
		{"testFunc", []string{"key", "value"}, "/r/lorem/ipsum$help&func=testFunc&key=value", Realm(cd + "/r/lorem/ipsum")},
		{"noArgsFunc", []string{}, "/r/lorem/ipsum$help&func=noArgsFunc", Realm(cd + "/r/lorem/ipsum")},
		{"oddArgsFunc", []string{"key"}, "/r/lorem/ipsum$help&func=oddArgsFunc&error=odd+number+of+arguments", Realm(cd + "/r/lorem/ipsum")},
		{"foo", []string{"bar", "1", "baz", "2"}, "https://gno.world/r/lorem/ipsum$help&func=foo&bar=1&baz=2", "gno.world/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "https://gno.world/r/lorem/ipsum$help&func=testFunc&key=value", "gno.world/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "https://gno.world/r/lorem/ipsum$help&func=noArgsFunc", "gno.world/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "https://gno.world/r/lorem/ipsum$help&func=oddArgsFunc&error=odd+number+of+arguments", "gno.world/r/lorem/ipsum"},
	}

	for _, tt := range tests {
		title := tt.fn
		t.Run(title, func(t *testing.T) {
			if tt.fn == "oddArgsFunc" {
				defer func() {
					if r := recover(); r != nil {
						if r != "odd number of arguments" {
							t.Errorf("expected panic with message 'odd number of arguments', got: %v", r)
						}
					} else {
						t.Error("expected panic for odd number of arguments, but did not panic")
					}
				}()
			}
			got := tt.realm_XXX.FuncURL(tt.fn, tt.args...)
			urequire.Equal(t, tt.want, got)
		})
	}
}

func TestHome(t *testing.T) {
	cd := runtime.ChainDomain()
	tests := []struct {
		realm_XXX Realm
		want      string
	}{
		{"", "$help"},
		{Realm(cd + "/r/lorem/ipsum"), "/r/lorem/ipsum$help"},
		{"gno.world/r/lorem/ipsum", "https://gno.world/r/lorem/ipsum$help"},
	}

	for _, tt := range tests {
		t.Run(string(tt.realm_XXX), func(t *testing.T) {
			got := tt.realm_XXX.Home()
			urequire.Equal(t, tt.want, got)
		})
	}
}
