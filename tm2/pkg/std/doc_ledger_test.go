package std

import (
	"bytes"
	"encoding/json"
	"testing"
)

// The Ledger Cosmos app validates the amino sign doc against a fixed allowlist
// of keys and refuses to sign anything carrying a key it does not know
// (ledger-cosmos, app/src/tx_validate.c). gno has no Ledger app of its own and
// borrows that one, so the sign doc's shape is not an internal detail: any key
// outside these sets makes every gno transaction unsignable on a Ledger, with
// the device reporting only "Unexpected field".
//
// These are the app's own lists, transcribed.
var (
	allowedSignDocKeys = map[string]bool{
		"account_number": true,
		"chain_id":       true,
		"fee":            true,
		"memo":           true,
		"msgs":           true,
		"sequence":       true,
		"tip":            true,
		"timeout_height": true,
	}
	allowedFeeKeys = map[string]bool{
		"amount":  true,
		"gas":     true,
		"granter": true,
		"payer":   true,
	}
)

func TestSignaturePayloadKeysAreLedgerCompatible(t *testing.T) {
	t.Parallel()

	payload, err := GetSignaturePayload(SignDoc{
		ChainID:       "dev",
		AccountNumber: 42,
		Sequence:      7,
		Fee:           NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}),
		Msgs:          nil,
		Memo:          "",
	})
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatalf("payload is not an object: %v", err)
	}
	for key := range doc {
		if !allowedSignDocKeys[key] {
			t.Errorf("sign doc key %q is outside the Ledger allowlist; "+
				"every gno transaction becomes unsignable on a Ledger", key)
		}
	}

	var fee map[string]json.RawMessage
	if err := json.Unmarshal(doc["fee"], &fee); err != nil {
		t.Fatalf("fee is not an object: %v", err)
	}
	for key := range fee {
		if !allowedFeeKeys[key] {
			t.Errorf("fee key %q is outside the Ledger allowlist", key)
		}
	}
}

// The app expects the Cosmos shapes as well as the Cosmos names: an ARRAY of
// coins under "amount", and every number rendered as a string. A payload that
// used the right key names with the wrong shapes would pass the test above and
// still be refused by the device.
func TestSignaturePayloadFeeShape(t *testing.T) {
	t.Parallel()

	payload, err := GetSignaturePayload(SignDoc{
		ChainID:       "dev",
		AccountNumber: 42,
		Sequence:      7,
		Fee:           NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}),
	})
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}

	var doc struct {
		Fee struct {
			Amount []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"amount"`
			Gas string `json:"gas"`
		} `json:"fee"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatalf("fee does not have the expected shape: %v", err)
	}

	if len(doc.Fee.Amount) != 1 {
		t.Fatalf("fee.amount must be an array of one coin, got %d", len(doc.Fee.Amount))
	}
	if got, want := doc.Fee.Amount[0].Denom, "ugnot"; got != want {
		t.Errorf("fee.amount[0].denom = %q, want %q", got, want)
	}
	if got, want := doc.Fee.Amount[0].Amount, "1000000"; got != want {
		t.Errorf("fee.amount[0].amount = %q, want %q (a string, not a number)", got, want)
	}
	if got, want := doc.Fee.Gas, "200000"; got != want {
		t.Errorf("fee.gas = %q, want %q (a string, not a number)", got, want)
	}
}

// A ZERO FEE IS AN EMPTY LIST. Genesis transactions are signed with
// GetSignBytes(chainID, 0, 0) and the ante handler branches on GasFee.IsZero(),
// so this is a live path, not a curiosity. Rendering it as a list holding
// {"denom":"","amount":"0"} would put an empty denom in front of a device that
// displays every coin it is given.
func TestSignaturePayloadZeroFeeIsEmptyList(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		fee  Fee
	}{
		{"wholly zero", Fee{}},
		{"zero amount with a denom", NewFee(0, Coin{Denom: "ugnot", Amount: 0})},
	} {
		payload, err := GetSignaturePayload(SignDoc{ChainID: "dev", Fee: tc.fee})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		var doc struct {
			Fee struct {
				Amount []json.RawMessage `json:"amount"`
			} `json:"fee"`
		}
		if err := json.Unmarshal(payload, &doc); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if len(doc.Fee.Amount) != 0 {
			t.Errorf("%s: fee.amount = %v, want an empty list", tc.name, doc.Fee.Amount)
		}
		if bytes.Contains(payload, []byte(`"denom":""`)) {
			t.Errorf("%s: payload carries a coin with an empty denom: %s", tc.name, payload)
		}
	}
}

// THE WHOLE PAYLOAD, PINNED. The two tests above check the properties that
// matter to the device; this one fails on ANY change to the signed bytes, which
// is a consensus change whether or not it was intended as one. Nothing in the
// tree pinned these bytes before, which is how the Ledger incompatibility
// arrived unnoticed.
func TestSignaturePayloadIsStable(t *testing.T) {
	t.Parallel()

	payload, err := GetSignaturePayload(SignDoc{
		ChainID:       "dev",
		AccountNumber: 42,
		Sequence:      7,
		Fee:           NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}),
		Msgs:          nil,
		Memo:          "hello",
	})
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}

	const want = `{"account_number":"42","chain_id":"dev","fee":{"amount":[{"amount":"1000000","denom":"ugnot"}],"gas":"200000"},"memo":"hello","msgs":null,"sequence":"7"}`
	if got := string(payload); got != want {
		t.Errorf("signed bytes changed -- this is a consensus change\n got: %s\nwant: %s", got, want)
	}
}
