package std

import (
	"reflect"
	"testing"
)

func TestTxSignDoc(t *testing.T) {
	t.Parallel()

	tx := Tx{
		Msgs: nil,
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
