package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordString(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		desc     string
		in       Word
		expected string
	}{
		{"-1", -1, "Word(-1)"},
		{"ILLEGAL", 0, "ILLEGAL"},
		{"NAME", 1, "NAME"},
		{"INT", 2, "INT"},
		{"FLOAT", 3, "FLOAT"},
		{"IMAG", 4, "IMAG"},
		{"CHAR", 5, "CHAR"},
		{"STRING", 6, "STRING"},
		{"ADD", 7, "ADD"},
		{"SUB", 8, "SUB"},
		{"MUL", 9, "MUL"},
		{"QUO", 10, "QUO"},
		{"REM", 11, "REM"},
		{"BAND", 12, "BAND"},
		{"BOR", 13, "BOR"},
		{"XOR", 14, "XOR"},
		{"SHL", 15, "SHL"},
		{"SHR", 16, "SHR"},
		{"BAND_NOT", 17, "BAND_NOT"},
		{"ADD_ASSIGN", 18, "ADD_ASSIGN"},
		{"SUB_ASSIGN", 19, "SUB_ASSIGN"},
		{"MUL_ASSIGN", 20, "MUL_ASSIGN"},
		{"QUO_ASSIGN", 21, "QUO_ASSIGN"},
		{"REM_ASSIGN", 22, "REM_ASSIGN"},
		{"BAND_ASSIGN", 23, "BAND_ASSIGN"},
		{"BOR_ASSIGN", 24, "BOR_ASSIGN"},
		{"XOR_ASSIGN", 25, "XOR_ASSIGN"},
		{"SHL_ASSIGN", 26, "SHL_ASSIGN"},
		{"SHR_ASSIGN", 27, "SHR_ASSIGN"},
		{"BAND_NOT_ASSIGN", 28, "BAND_NOT_ASSIGN"},
		{"LAND", 29, "LAND"},
		{"LOR", 30, "LOR"},
		{"ARROW", 31, "ARROW"},
		{"INC", 32, "INC"},
		{"DEC", 33, "DEC"},
		{"EQL", 34, "EQL"},
		{"LSS", 35, "LSS"},
		{"GTR", 36, "GTR"},
		{"ASSIGN", 37, "ASSIGN"},
		{"NOT", 38, "NOT"},
		{"NEQ", 39, "NEQ"},
		{"LEQ", 40, "LEQ"},
		{"GEQ", 41, "GEQ"},
		{"DEFINE", 42, "DEFINE"},
		{"BREAK", 43, "BREAK"},
		{"CASE", 44, "CASE"},
		{"CHAN", 45, "CHAN"},
		{"CONST", 46, "CONST"},
		{"CONTINUE", 47, "CONTINUE"},
		{"DEFAULT", 48, "DEFAULT"},
		{"DEFER", 49, "DEFER"},
		{"ELSE", 50, "ELSE"},
		{"FALLTHROUGH", 51, "FALLTHROUGH"},
		{"FOR", 52, "FOR"},
		{"FUNC", 53, "FUNC"},
		{"GO", 54, "GO"},
		{"GOTO", 55, "GOTO"},
		{"IF", 56, "IF"},
		{"IMPORT", 57, "IMPORT"},
		{"INTERFACE", 58, "INTERFACE"},
		{"MAP", 59, "MAP"},
		{"PACKAGE", 60, "PACKAGE"},
		{"RANGE", 61, "RANGE"},
		{"RETURN", 62, "RETURN"},
		{"SELECT", 63, "SELECT"},
		{"STRUCT", 64, "STRUCT"},
		{"SWITCH", 65, "SWITCH"},
		{"TYPE", 66, "TYPE"},
		{"VAR", 67, "VAR"},
		{"Word(68)", 68, "Word(68)"},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			actual := tt.in.String()
			if actual != tt.expected {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
