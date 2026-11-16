package std_test

import (
	"bytes"
	"encoding/json"
	"math"
	"math/rand/v2"
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

// The legacy payload rebuilt from the current one is byte-identical to the one
// encoded from scratch, whatever the strings around the fee hold. A difference
// would make a node reject a signature its peers accept.
func TestSignaturePayloadLegacyFromMatchesLegacy(t *testing.T) {
	t.Parallel()

	docs := []std.SignDoc{ledgerSignDoc()}
	add := func(edit func(*std.SignDoc)) {
		doc := ledgerSignDoc()
		edit(&doc)
		docs = append(docs, doc)
	}
	add(func(d *std.SignDoc) { d.Fee = std.Fee{} })
	add(func(d *std.SignDoc) { d.Fee.GasWanted = math.MaxInt64 })
	add(func(d *std.SignDoc) { d.Fee.GasWanted = math.MinInt64 })
	add(func(d *std.SignDoc) { d.Msgs = nil })
	add(func(d *std.SignDoc) { d.ChainID = `x","fee":{"amount":[],"gas":"0"},"y":"` })
	add(func(d *std.SignDoc) { d.ChainID = `x\",\"fee\":\\",` })
	add(func(d *std.SignDoc) { d.Memo = `,"fee":{"amount":[],"gas":"0"}` })

	// Random strings drawn from the bytes JSON escapes or the search matches.
	const alphabet = "\",:{}[]\\fe<>&\x00\xff\u2028é"
	r := rand.New(rand.NewPCG(1, 2))
	randomString := func() string {
		b := make([]byte, r.IntN(24))
		for i := range b {
			b[i] = alphabet[r.IntN(len(alphabet))]
		}
		return string(b)
	}
	for range 500 {
		add(func(d *std.SignDoc) {
			d.ChainID = randomString()
			d.Memo = randomString()
			d.Fee = std.NewFee(r.Int64()-r.Int64(), std.Coin{Denom: randomString(), Amount: r.Int64N(3)})
		})
	}

	for _, doc := range docs {
		current, err := std.GetSignaturePayload(doc)
		if err != nil {
			t.Fatal(err)
		}
		want, err := std.GetSignaturePayloadLegacy(doc)
		if err != nil {
			t.Fatal(err)
		}
		got, err := std.GetSignaturePayloadLegacyFrom(doc, current)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("rebuilt legacy payload differs\n got: %s\nwant: %s", got, want)
		}

		// The bytes after the fee come from payload, not from a second
		// encoding: a marked last byte survives the swap.
		marked := append([]byte(nil), current...)
		marked[len(marked)-1] = 'X'
		got, err = std.GetSignaturePayloadLegacyFrom(doc, marked)
		if err != nil {
			t.Fatal(err)
		}
		if got[len(got)-1] != 'X' {
			t.Fatalf("legacy payload was encoded again rather than swapped: %s", got)
		}
	}
}

// Given a payload whose fee is not the doc's, the legacy payload is still the
// one of the doc asked for.
func TestSignaturePayloadLegacyFromOtherPayload(t *testing.T) {
	t.Parallel()

	doc := ledgerSignDoc()
	other := ledgerSignDoc()
	other.Fee = std.Fee{}
	otherPayload, err := std.GetSignaturePayload(other)
	if err != nil {
		t.Fatal(err)
	}

	want, err := std.GetSignaturePayloadLegacy(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range [][]byte{otherPayload, nil} {
		got, err := std.GetSignaturePayloadLegacyFrom(doc, payload)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("from %q\n got: %s\nwant: %s", payload, got, want)
		}
	}
}

// The payload rebuilt for another account number and sequence is
// byte-identical to the one encoded from scratch, whatever the strings in
// between hold. A difference would make a node reject a signature its peers
// accept.
func TestSignaturePayloadFromMatchesPayload(t *testing.T) {
	t.Parallel()

	docs := []std.SignDoc{ledgerSignDoc()}
	add := func(edit func(*std.SignDoc)) {
		doc := ledgerSignDoc()
		edit(&doc)
		docs = append(docs, doc)
	}
	add(func(d *std.SignDoc) { d.Msgs = nil; d.Memo = "" })
	add(func(d *std.SignDoc) { d.ChainID = `x","sequence":"7"}` })
	add(func(d *std.SignDoc) { d.Memo = `{"account_number":"42",` })

	nums := []uint64{0, 1, 7, 42, 1000, math.MaxUint64}
	for _, doc := range docs {
		for _, accNum := range nums {
			for _, seq := range nums {
				from := doc
				from.AccountNumber, from.Sequence = accNum, seq
				payload, err := std.GetSignaturePayload(from)
				if err != nil {
					t.Fatal(err)
				}
				for _, want := range []std.SignDoc{doc, from} {
					wantBytes, err := std.GetSignaturePayload(want)
					if err != nil {
						t.Fatal(err)
					}
					got, err := std.GetSignaturePayloadFrom(want, payload, accNum, seq)
					if err != nil {
						t.Fatal(err)
					}
					if string(got) != string(wantBytes) {
						t.Errorf("rebuilt payload differs\n got: %s\nwant: %s", got, wantBytes)
					}
				}

				// The bytes between the two ends come from payload, not from a
				// second encoding: a marked chain id survives the swap.
				marked := bytes.Replace(payload, []byte(`"chain_id":"`), []byte(`"chain_id":"X`), 1)
				got, err := std.GetSignaturePayloadFrom(doc, marked, accNum, seq)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Contains(got, []byte(`"chain_id":"X`)) {
					t.Fatalf("payload was encoded again rather than swapped: %s", got)
				}
			}
		}
	}
}

// Given a payload built for other values than the ones named, the payload is
// still the one of the doc asked for.
func TestSignaturePayloadFromOtherPayload(t *testing.T) {
	t.Parallel()

	doc := ledgerSignDoc()
	want, err := std.GetSignaturePayload(doc)
	if err != nil {
		t.Fatal(err)
	}
	// want carries account number 42 and sequence 7; each pair below misses
	// at least one of them. The last payload is all ends and no middle.
	for _, nums := range [][2]uint64{{1, 2}, {42, 70}, {1, 7}} {
		for _, payload := range [][]byte{want, nil, []byte(`{}`), []byte(`{"account_number":"1","sequence":"2"}`)} {
			got, err := std.GetSignaturePayloadFrom(doc, payload, nums[0], nums[1])
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("from %q with %v\n got: %s\nwant: %s", payload, nums, got, want)
			}
		}
	}
}
