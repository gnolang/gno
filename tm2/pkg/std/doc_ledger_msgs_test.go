package std_test

import (
	"encoding/json"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// ledgerSignDoc carries a real chain message. The pins in doc_ledger_test.go
// use a nil Msgs, so they say nothing about the field that actually carries
// the authorisation; this file is external so it can import bank.
func ledgerSignDoc() std.SignDoc {
	return std.SignDoc{
		ChainID:       "dev",
		AccountNumber: 42,
		Sequence:      7,
		Fee:           std.NewFee(200000, std.Coin{Denom: "ugnot", Amount: 1000000}),
		Msgs: []std.Msg{bank.MsgSend{
			FromAddress: crypto.AddressFromPreimage([]byte("from")),
			ToAddress:   crypto.AddressFromPreimage([]byte("to")),
			Amount:      std.NewCoins(std.NewCoin("ugnot", 10)),
		}},
		Memo: "hello",
	}
}

// The msgs render through the payload struct exactly as they would through
// SignDoc itself, @type and all. Pinned byte for byte because any change here
// is a consensus change.
func TestSignaturePayloadPinsMsgs(t *testing.T) {
	t.Parallel()

	payload, err := std.GetSignaturePayload(ledgerSignDoc())
	if err != nil {
		t.Fatal(err)
	}

	const want = `{"account_number":"42","chain_id":"dev","fee":{"amount":[{"amount":"1000000","denom":"ugnot"}],"gas":"200000"},"memo":"hello","msgs":[{"@type":"/bank.MsgSend","amount":"10ugnot","from_address":"g1wkzh53vfnxzmunzdjs0fpd4njmtvj2jvej5y7z","to_address":"g1vcl2r0llu5pc70cv7enlznzz2lhl2tthd3rhvp"}],"sequence":"7"}`
	if got := string(payload); got != want {
		t.Errorf("signed bytes changed -- this is a consensus change\n got: %s\nwant: %s", got, want)
	}
}

// Only the fee differs between the two renderings. Every other top-level
// field, msgs included, is byte-identical, which is what lets a client that
// signs the legacy rendering be understood as authorising the same transaction.
func TestSignaturePayloadRenderingsDifferOnlyInFee(t *testing.T) {
	t.Parallel()

	doc := ledgerSignDoc()
	current, err := std.GetSignaturePayload(doc)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := std.GetSignaturePayloadLegacy(doc)
	if err != nil {
		t.Fatal(err)
	}

	var currentFields, legacyFields map[string]json.RawMessage
	if err := json.Unmarshal(current, &currentFields); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(legacy, &legacyFields); err != nil {
		t.Fatal(err)
	}
	if string(currentFields["fee"]) == string(legacyFields["fee"]) {
		t.Fatal("the two renderings agree on the fee, so this test proves nothing")
	}
	delete(currentFields, "fee")
	delete(legacyFields, "fee")

	if len(currentFields) != len(legacyFields) {
		t.Fatalf("field count differs outside the fee: current %d, legacy %d",
			len(currentFields), len(legacyFields))
	}
	for key, legacyValue := range legacyFields {
		if got, want := string(currentFields[key]), string(legacyValue); got != want {
			t.Errorf("%s differs between renderings\n current: %s\n legacy:  %s", key, got, want)
		}
	}
}
