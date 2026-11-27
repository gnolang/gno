package std

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

type utf8TestInner struct {
	Note string `json:"note,omitempty"`
}

type utf8TestMsg struct {
	Args     []string       `json:"args"`
	Data     []byte         `json:"data"`
	Key      [4]byte        `json:"key"`
	Inner    *utf8TestInner `json:"inner"`
	Any      any            `json:"any"`
	Untagged string
	Skipped  string `json:"-"`
	private  string
}

func (utf8TestMsg) Route() string                { return "test" }
func (utf8TestMsg) Type() string                 { return "test" }
func (utf8TestMsg) ValidateBasic() error         { return nil }
func (utf8TestMsg) GetSignBytes() []byte         { return nil }
func (utf8TestMsg) GetSigners() []crypto.Address { return nil }

// utf8TestCode is a number amino writes as a one-byte string.
type utf8TestCode uint16

func (c utf8TestCode) MarshalAmino() (string, error) { return string([]byte{byte(c)}), nil }

// utf8TestByte has a string representation too, but amino never calls it on
// a list element: it writes a list of byte-kind values as bytes.
type utf8TestByte uint8

func (b utf8TestByte) MarshalAmino() (string, error) { return string([]byte{byte(b)}), nil }

// utf8TestFails has a representation that cannot be built, so nothing can be
// checked.
type utf8TestFails struct{}

func (utf8TestFails) MarshalAmino() (string, error) { return "", errors.New("no representation") }

func TestValidateUTF8(t *testing.T) {
	t.Parallel()

	tx := func(memo, denom string, msg utf8TestMsg) Tx {
		return Tx{
			Msgs: []Msg{utf8TestMsg{}, msg},
			Fee:  NewFee(1, Coin{Denom: denom, Amount: 1}),
			Memo: memo,
		}
	}
	valid := utf8TestMsg{
		Args:  []string{"plain", "a\uFFFDb", "héllo"},
		Data:  []byte{0xff, 0xfe},     // binary, not text
		Key:   [4]byte{0xff, 0, 0, 0}, // likewise
		Inner: &utf8TestInner{Note: "ok"},
		Any:   "ok",
	}

	tests := []struct {
		name    string
		v       any
		wantErr string // empty when valid
	}{
		{"valid tx", tx("m\uFFFD", "ugnot", valid), ""},
		{"zero tx", Tx{}, ""},
		{"nil pointer and interface", tx("", "", utf8TestMsg{}), ""},
		{"unexported field", tx("", "", utf8TestMsg{private: "\xff"}), ""},
		{"bare string", "a\xffb", "invalid UTF-8 string"},
		{"memo", tx("a\xffb", "ugnot", valid), "invalid UTF-8 in memo"},
		{"fee coin, through its amino string", tx("", "ugnot\xff", valid), "invalid UTF-8 in fee.gas_fee"},
		{"arg", tx("", "", utf8TestMsg{Args: []string{"ok", "\x80"}}), "invalid UTF-8 in msg[1].args[1]"},
		{"through pointer", tx("", "", utf8TestMsg{Inner: &utf8TestInner{Note: "\xc3"}}), "invalid UTF-8 in msg[1].inner.note"},
		{"through interface", tx("", "", utf8TestMsg{Any: []string{"\xfe"}}), "invalid UTF-8 in msg[1].any[0]"},
		{"untagged field", tx("", "", utf8TestMsg{Untagged: "\xff"}), "invalid UTF-8 in msg[1].Untagged"},
		{"json dash field, which amino skips", tx("", "", utf8TestMsg{Skipped: "\xff"}), ""},
		{"numbers", []int64{1, 2}, ""},
		{"numbers written as strings", []utf8TestCode{'a', 0xff}, "invalid UTF-8 in [1]"},
		{"bytes with a string representation", []utf8TestByte{0xff}, ""},
		{"representation fails", utf8TestFails{}, "cannot check UTF-8: no representation"},
		{"representation fails in a field", tx("", "", utf8TestMsg{Any: utf8TestFails{}}), "cannot check UTF-8 in msg[1].any: no representation"},
		{"truncated sequence", tx("", "", utf8TestMsg{Args: []string{"\xe2\x82"}}), "invalid UTF-8 in msg[1].args[0]"},
		{"surrogate half", tx("", "", utf8TestMsg{Args: []string{"\xed\xa0\x80"}}), "invalid UTF-8 in msg[1].args[0]"},
		{"first of two reported", tx("\xff", "", utf8TestMsg{Args: []string{"\xff"}}), "invalid UTF-8 in msg[1].args[0]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateUTF8(tt.v)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

// A signer gets an error rather than a payload that would also verify the
// same tx with its invalid bytes swapped for U+FFFD.
func TestSignaturePayloadRefusesInvalidUTF8(t *testing.T) {
	t.Parallel()

	tx := Tx{Fee: NewFee(1, Coin{Denom: "ugnot", Amount: 1}), Memo: "a\xffb"}
	_, err := tx.GetSignBytes("dev", 1, 2)
	require.ErrorContains(t, err, "invalid UTF-8 in memo")
	_, err = tx.GetSignBytesLegacy("dev", 1, 2)
	require.ErrorContains(t, err, "invalid UTF-8 in memo")

	tx.Memo = "a\uFFFDb"
	_, err = tx.GetSignBytes("dev", 1, 2)
	require.NoError(t, err)
	_, err = tx.GetSignBytesLegacy("dev", 1, 2)
	require.NoError(t, err)
}
