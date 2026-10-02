package std

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

func TestTxSignDoc(t *testing.T) {
	t.Parallel()

	tx := Tx{
		Fee:  NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}),
		Memo: "hello",
	}

	got := tx.SignDoc("dev", 42, 7)
	want := SignDoc{
		ChainID:       "dev",
		AccountNumber: 42,
		Sequence:      7,
		Fee:           tx.Fee,
		Msgs:          tx.Msgs,
		Memo:          "hello",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tx.SignDoc() = %+v, want %+v", got, want)
	}
}

// feeTestMsg is a Msg with one signer, so a tx carrying it and one signature
// passes ValidateBasic's signature checks and only the fee decides.
type feeTestMsg struct{}

func (feeTestMsg) Route() string                { return "test" }
func (feeTestMsg) Type() string                 { return "test" }
func (feeTestMsg) ValidateBasic() error         { return nil }
func (feeTestMsg) GetSignBytes() []byte         { return nil }
func (feeTestMsg) GetSigners() []crypto.Address { return []crypto.Address{{1}} }

// ValidateBasic accepts an empty fee, which a sponsored (0-fee) tx carries; the
// ante decides whether a 0-fee tx is allowed. A malformed fee is still refused.
func TestTxValidateBasicFee(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		fee   Coin
		valid bool
	}{
		{"empty zero coin", Coin{}, true},
		{"valid coin", Coin{Denom: "ugnot", Amount: 1}, true},
		{"zero amount, valid denom", Coin{Denom: "ugnot", Amount: 0}, true},
		{"zero amount, bad denom", Coin{Denom: "BAD!", Amount: 0}, false},
		{"empty denom, non-zero amount", Coin{Amount: 5}, false},
		{"negative amount", Coin{Denom: "ugnot", Amount: -1}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tx := Tx{
				Msgs:       []Msg{feeTestMsg{}},
				Fee:        NewFee(100_000, tc.fee),
				Signatures: []Signature{{}},
			}
			err := tx.ValidateBasic()
			if tc.valid {
				require.NoError(t, err)
				return
			}
			require.IsType(t, InsufficientFeeError{}, errors.Cause(err))
		})
	}
}
