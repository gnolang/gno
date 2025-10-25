package txlink

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/urequire"
)

func TestCall(t *testing.T) {
	cd := runtime.ChainDomain()

	tests := []struct {
		fn        string
		args      []string
		want      string
		realm_XXX Realm
	}{
		{"foo", []string{"bar", "1", "baz", "2"}, "/p/moul/txlink$help&func=foo&bar=1&baz=2", ""},
		{"testFunc", []string{"key", "value"}, "/p/moul/txlink$help&func=testFunc&key=value", ""},
		{"noArgsFunc", []string{}, "/p/moul/txlink$help&func=noArgsFunc", ""},
		{"oddArgsFunc", []string{"key"}, "/p/moul/txlink$help&func=oddArgsFunc&error=odd+number+of+arguments", ""},
		{"foo", []string{"bar", "1", "baz", "2"}, "/r/lorem/ipsum$help&func=foo&bar=1&baz=2", Realm(cd + "/r/lorem/ipsum")},
		{"testFunc", []string{"key", "value"}, "/r/lorem/ipsum$help&func=testFunc&key=value", Realm(cd + "/r/lorem/ipsum")},
		{"noArgsFunc", []string{}, "/r/lorem/ipsum$help&func=noArgsFunc", Realm(cd + "/r/lorem/ipsum")},
		{"oddArgsFunc", []string{"key"}, "/r/lorem/ipsum$help&func=oddArgsFunc&error=odd+number+of+arguments", Realm(cd + "/r/lorem/ipsum")},
		{"foo", []string{"bar", "1", "baz", "2"}, "https://gno.world/r/lorem/ipsum$help&func=foo&bar=1&baz=2", "gno.world/r/lorem/ipsum"},
		{"testFunc", []string{"key", "value"}, "https://gno.world/r/lorem/ipsum$help&func=testFunc&key=value", "gno.world/r/lorem/ipsum"},
		{"noArgsFunc", []string{}, "https://gno.world/r/lorem/ipsum$help&func=noArgsFunc", "gno.world/r/lorem/ipsum"},
		{"oddArgsFunc", []string{"key"}, "https://gno.world/r/lorem/ipsum$help&func=oddArgsFunc&error=odd+number+of+arguments", "gno.world/r/lorem/ipsum"},
		{"test", []string{"key", "hello world"}, "/p/moul/txlink$help&func=test&key=hello+world", ""},
		{"test", []string{"key", "a&b=c"}, "/p/moul/txlink$help&func=test&key=a%26b%3Dc", ""},
		{"test", []string{"key", ""}, "/p/moul/txlink$help&func=test&key=", ""},
		{"testSend", []string{"key", "hello world", ".send", "1000000ugnot"}, "/p/moul/txlink$help&func=testSend&.send=1000000ugnot&key=hello+world", ""},
	}

	for _, tt := range tests {
		title := string(tt.realm_XXX) + "_" + tt.fn
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
			got := tt.realm_XXX.Call(tt.fn, tt.args...)
			urequire.Equal(t, tt.want, got)
		})
	}
}

func TestBuilder(t *testing.T) {
	cases := []struct {
		name     string
		build    func() string
		expected string
	}{
		// Basic functionality tests
		{
			name: "empty_function",
			build: func() string {
				return NewLink("").URL()
			},
			expected: "",
		},
		{
			name: "function_without_args",
			build: func() string {
				return NewLink("MyFunc").URL()
			},
			expected: "/p/moul/txlink$help&func=MyFunc",
		},

		// Realm tests
		{
			name: "gnoland_realm",
			build: func() string {
				return Realm("gno.land/r/demo").
					NewLink("MyFunc").
					AddArgs("key", "value").
					URL()
			},
			expected: "/r/demo$help&func=MyFunc&key=value",
		},
		{
			name: "external_realm",
			build: func() string {
				return Realm("gno.world/r/demo").
					NewLink("MyFunc").
					AddArgs("key", "value").
					URL()
			},
			expected: "https://gno.world/r/demo$help&func=MyFunc&key=value",
		},
		{
			name: "empty_realm",
			build: func() string {
				return Realm("").
					NewLink("func").
					AddArgs("key", "value").
					URL()
			},
			expected: "/p/moul/txlink$help&func=func&key=value",
		},

		// URL encoding tests
		{
			name: "url_encoding_with_spaces",
			build: func() string {
				return NewLink("test").
					AddArgs("key", "hello world").
					URL()
			},
			expected: "/p/moul/txlink$help&func=test&key=hello+world",
		},
		{
			name: "url_encoding_with_special_chars",
			build: func() string {
				return NewLink("test").
					AddArgs("key", "a&b=c").
					URL()
			},
			expected: "/p/moul/txlink$help&func=test&key=a%26b%3Dc",
		},
		{
			name: "url_encoding_with_unicode",
			build: func() string {
				return NewLink("func").
					AddArgs("key", "🌟").
					URL()
			},
			expected: "/p/moul/txlink$help&func=func&key=%F0%9F%8C%9F",
		},
		{
			name: "url_encoding_with_special_chars_in_key",
			build: func() string {
				return NewLink("func").
					AddArgs("my/key", "value").
					URL()
			},
			expected: "/p/moul/txlink$help&func=func&my%2Fkey=value",
		},

		// AddArgs tests
		{
			name: "addargs_with_multiple_pairs",
			build: func() string {
				return NewLink("MyFunc").
					AddArgs("key1", "value1", "key2", "value2").
					URL()
			},
			expected: "/p/moul/txlink$help&func=MyFunc&key1=value1&key2=value2",
		},
		{
			name: "addargs_with_odd_number_of_args",
			build: func() string {
				defer func() {
					if r := recover(); r != nil {
						if r != "odd number of arguments" {
							t.Errorf("expected panic with message 'odd number of arguments', got: %v", r)
						}
					} else {
						t.Error("expected panic for odd number of arguments, but did not panic")
					}
				}()
				return NewLink("MyFunc").
					AddArgs("key1", "value1", "orphan").
					URL()
			},
			expected: "",
		},

		// Empty values tests
		{
			name: "empty_key_should_be_ignored",
			build: func() string {
				return NewLink("func").
					AddArgs("", "value").
					URL()
			},
			expected: "/p/moul/txlink$help&func=func",
		},
		{
			name: "empty_value_should_be_kept",
			build: func() string {
				return NewLink("func").
					AddArgs("key", "").
					URL()
			},
			expected: "/p/moul/txlink$help&func=func&key=",
		},

		// Send tests
		{
			name: "send_via_addsend_method",
			build: func() string {
				return NewLink("MyFunc").
					AddArgs("key", "value").
					SetSend("1000000ugnot").
					URL()
			},
			expected: "/p/moul/txlink$help&func=MyFunc&.send=1000000ugnot&key=value",
		},
		{
			name: "send_via_addarg_method_panic",
			build: func() string {
				defer func() {
					if r := recover(); r != nil {
						if r != "invalid key" {
							t.Errorf("expected panic with message 'invalid key', got: %v", r)
						}
					} else {
						t.Errorf("expected panic for .send key, but did not panic")
					}
				}()
				NewLink("MyFunc").AddArgs(".send", "1000000ugnot")
				return "no panic occurred"
			},
			expected: "",
		},
		{
			name: "addsend_should_override_previous_addsend",
			build: func() string {
				return NewLink("MyFunc").
					SetSend("1000000ugnot").
					SetSend("2000000ugnot").
					URL()
			},
			expected: "/p/moul/txlink$help&func=MyFunc&.send=2000000ugnot",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.build()
			urequire.Equal(t, tc.expected, got)
		})
	}
}
