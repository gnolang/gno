package std

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		prefix  string
		key     string
		kind    string
		wantErr bool
	}{
		// Valid cases
		{"module", "valid_key.string", "string", false},
		{"prefix1", "validKey.int64", "int64", false},
		{"p_", "_valid123.bool", "bool", false},

		// Invalid key cases
		{"module", "invalidKey", "string", true},         // Missing ".kind" suffix
		{"module", "", "string", true},                   // Empty key
		{"module", "1invalid.string", "string", true},    // Starts with a number
		{"module", "-invalid.string", "string", true},    // Starts with an invalid character
		{"module", "invalid-123.string", "string", true}, // Contains invalid character (-)
		{"module", "valid/path.key.bool", "bool", true},  // Contains invalid character (/)

		// Invalid prefix cases
		{"1prefix", "valid_key.string", "string", true},           // Prefix starts with a number
		{"-prefix", "valid_key.string", "string", true},           // Prefix starts with an invalid character
		{"prefix!", "valid_key.string", "string", true},           // Prefix contains invalid character (!)
		{"module/submodule", "valid/path.key.bool", "bool", true}, // Prefix contains invalid character (/)
	}

	for _, tt := range tests {
		err := validate(tt.prefix, tt.key, tt.kind)
		if (err != nil) != tt.wantErr {
			t.Errorf("validate(%q, %q, %q) = %v, wantErr %v", tt.prefix, tt.key, tt.kind, err, tt.wantErr)
		}
	}
}
