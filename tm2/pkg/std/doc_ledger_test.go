package std

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
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

	// The app also REQUIRES these six (tx_validate.c answers
	// parser_json_missing_<key> when one is absent), so a payload that dropped
	// a key -- an omitempty on the memo, say -- would pass the allowlist and
	// still be refused.
	requiredSignDocKeys = []string{
		"account_number",
		"chain_id",
		"fee",
		"memo",
		"msgs",
		"sequence",
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
	for _, key := range requiredSignDocKeys {
		if _, ok := doc[key]; !ok {
			t.Errorf("sign doc key %q is missing; the Ledger app requires it "+
				"even when its value is empty", key)
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
//
// The bytes are pinned rather than parsed: json.Unmarshal reads both `[]` and
// `null` into a zero-length slice, so a structural check cannot tell the empty
// list apart from the `null` a nil slice renders as. The two are different
// signed bytes, and only one of them is the Cosmos shape.
func TestSignaturePayloadZeroFeeIsEmptyList(t *testing.T) {
	t.Parallel()

	const want = `{"account_number":"0","chain_id":"dev","fee":{"amount":[],"gas":"0"},"memo":"","msgs":null,"sequence":"0"}`

	for _, tc := range []struct {
		name string
		fee  Fee
	}{
		{"wholly zero", Fee{}},
		{"zero amount with a denom", NewFee(0, Coin{Denom: "ugnot", Amount: 0})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payload, err := GetSignaturePayload(SignDoc{ChainID: "dev", Fee: tc.fee})
			if err != nil {
				t.Fatal(err)
			}
			if got := string(payload); got != want {
				t.Errorf("zero-fee payload\n got: %s\nwant: %s", got, want)
			}
		})
	}
}

// signDocPayload is a hand-written copy of SignDoc, and nothing in the compiler
// ties the two together. A field added to SignDoc and forgotten here would be
// part of the transaction yet absent from the signed bytes, with no compile
// error and no failing test. The fee is the one field allowed to differ.
func TestSignDocPayloadMirrorsSignDoc(t *testing.T) {
	t.Parallel()

	doc, payload := reflect.TypeFor[SignDoc](), reflect.TypeFor[signDocPayload]()
	if doc.NumField() != payload.NumField() {
		t.Fatalf("SignDoc has %d fields, signDocPayload has %d; a field is "+
			"missing from the signed bytes", doc.NumField(), payload.NumField())
	}

	for i := range doc.NumField() {
		df := doc.Field(i)
		pf, ok := payload.FieldByName(df.Name)
		if !ok {
			t.Errorf("SignDoc.%s has no counterpart in signDocPayload", df.Name)
			continue
		}
		if got, want := pf.Tag.Get("json"), df.Tag.Get("json"); got != want {
			t.Errorf("%s: json tag %q, want %q", df.Name, got, want)
		}
		if df.Name != "Fee" && df.Type != pf.Type {
			t.Errorf("%s: type %s, want %s", df.Name, pf.Type, df.Type)
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

// THE LEGACY RENDERING IS THE PRE-CHANGE ONE, byte for byte. Its whole purpose
// is to reproduce what clients signed before the fee moved, so a literal is the
// only honest way to pin it: anything derived from the current code would follow
// the current code wherever it went, which is exactly the failure this guards.
//
// The expectation below was taken from master before the change (amino over
// SignDoc itself, then sortJSON), and utils.go's sortJSON is untouched.
func TestSignaturePayloadLegacyIsTheOldRendering(t *testing.T) {
	t.Parallel()

	payload, err := GetSignaturePayloadLegacy(SignDoc{
		ChainID:       "dev",
		AccountNumber: 42,
		Sequence:      7,
		Fee:           NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}),
		Msgs:          nil,
		Memo:          "hello",
	})
	if err != nil {
		t.Fatalf("legacy payload: %v", err)
	}

	const want = `{"account_number":"42","chain_id":"dev","fee":{"gas_fee":"1000000ugnot","gas_wanted":"200000"},"memo":"hello","msgs":null,"sequence":"7"}`
	if got := string(payload); got != want {
		t.Errorf("the legacy rendering moved -- signatures made before the fee\n"+
			"change no longer verify, which breaks written genesis files\n got: %s\nwant: %s", got, want)
	}
}

// WHY A NODE MAY ACCEPT BOTH RENDERINGS AT ONCE. ante.go verifies against the
// current payload and, on failure, against the legacy one, so that clients which
// have not shipped the change -- and signatures already written into genesis
// files -- keep working. That is only sound if no legacy rendering of one
// transaction can equal the current rendering of a DIFFERENT one. If it could, a
// signature authorising T1 would also authorise T2, which is forgery wearing
// compatibility's coat.
//
// The separation is structural, not lucky: the legacy fee object carries
// gas_wanted and gas_fee, the current one amount and gas, and those key sets have
// no member in common. Two JSON objects that parse to different key sets are not
// the same bytes -- whatever else differs between the documents.
//
// THIS IS THE TEST TO CONSULT BEFORE CHANGING EITHER RENDERING. If a future fee
// shape reuses a legacy key name, dual verification stops being safe and this
// fails rather than letting it through quietly.
func TestSignaturePayloadEncodingsAreDisjoint(t *testing.T) {
	t.Parallel()

	/* The adversarial rows matter more than the ordinary one. An attacker
	   choosing the document gets to pick the memo, the chain id and the msgs,
	   so the separation has to hold when those carry the other encoding's own
	   text -- the fee is the only field they cannot reach. */
	docs := []struct {
		name string
		doc  SignDoc
	}{
		{"ordinary", SignDoc{
			ChainID: "dev", AccountNumber: 42, Sequence: 7,
			Fee: NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}), Memo: "hello",
		}},
		{"zero fee", SignDoc{ChainID: "dev"}},
		{"memo spelling the other fee shape", SignDoc{
			ChainID: "dev", AccountNumber: 1, Sequence: 1,
			Fee:  NewFee(1, Coin{Denom: "ugnot", Amount: 1}),
			Memo: `"fee":{"gas_wanted":"200000","gas_fee":"1000000ugnot"}`,
		}},
		{"chain id spelling it too", SignDoc{
			ChainID: `dev","fee":{"amount":[],"gas":"0`,
			Fee:     NewFee(5, Coin{Denom: "ugnot", Amount: 5}),
		}},
		{"large fee", SignDoc{
			ChainID: "dev", AccountNumber: 1 << 62, Sequence: 1 << 62,
			Fee: NewFee(1<<62, Coin{Denom: "ugnot", Amount: 1 << 62}),
		}},
	}

	feeKeys := func(t *testing.T, payload []byte) map[string]bool {
		t.Helper()
		var doc struct {
			Fee map[string]json.RawMessage `json:"fee"`
		}
		if err := json.Unmarshal(payload, &doc); err != nil {
			t.Fatalf("payload does not parse: %v (%s)", err, payload)
		}
		if len(doc.Fee) == 0 {
			t.Fatalf("payload carries no fee object at all: %s", payload)
		}
		keys := make(map[string]bool, len(doc.Fee))
		for k := range doc.Fee {
			keys[k] = true
		}
		return keys
	}

	legacyAll := make([][]byte, len(docs))
	currentAll := make([][]byte, len(docs))

	for i, tc := range docs {
		legacy, err := GetSignaturePayloadLegacy(tc.doc)
		if err != nil {
			t.Fatalf("%s: legacy payload: %v", tc.name, err)
		}
		current, err := GetSignaturePayload(tc.doc)
		if err != nil {
			t.Fatalf("%s: current payload: %v", tc.name, err)
		}
		legacyAll[i], currentAll[i] = legacy, current

		currentFeeKeys := feeKeys(t, current)
		for k := range feeKeys(t, legacy) {
			if currentFeeKeys[k] {
				t.Errorf("%s: fee key %q appears in BOTH renderings; dual "+
					"verification is no longer safe", tc.name, k)
			}
		}
	}

	/* The property that actually matters, stated over every pair: one
	   signature cannot span two different transactions. */
	for i := range docs {
		for j := range docs {
			if bytes.Equal(legacyAll[i], currentAll[j]) {
				t.Errorf("legacy rendering of %q equals the current rendering of %q:\n%s",
					docs[i].name, docs[j].name, legacyAll[i])
			}
		}
	}
}

// The helper the tools lean on: it must take both renderings, say which one
// matched, and still refuse a signature over neither.
func TestVerifySignaturePayloadTakesEitherRendering(t *testing.T) {
	t.Parallel()

	priv := ed25519.GenPrivKey()
	doc := SignDoc{
		ChainID: "dev", AccountNumber: 3, Sequence: 4,
		Fee: NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}), Memo: "m",
	}
	verify := func(t *testing.T, doc SignDoc, sig []byte) PayloadRendering {
		t.Helper()
		rendering, err := VerifySignaturePayload(priv.PubKey(), doc, sig)
		if err != nil {
			t.Fatal(err)
		}
		return rendering
	}

	for _, tc := range []struct {
		name   string
		render func(SignDoc) ([]byte, error)
		want   PayloadRendering
	}{
		{"current", GetSignaturePayload, PayloadRenderingCurrent},
		{"legacy", GetSignaturePayloadLegacy, PayloadRenderingLegacy},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payload, err := tc.render(doc)
			if err != nil {
				t.Fatal(err)
			}
			sig, err := priv.Sign(payload)
			if err != nil {
				t.Fatal(err)
			}

			if got := verify(t, doc, sig); got != tc.want {
				t.Errorf("a signature over the %s rendering verified as %d, want %d", tc.name, got, tc.want)
			}

			// AND IT STILL BINDS THE DOCUMENT. Accepting two renderings must
			// not mean accepting a signature made over a different
			// transaction: the same bytes against a changed sequence has to
			// fail, or the fallback has become a bypass.
			moved := doc
			moved.Sequence = doc.Sequence + 1
			if got := verify(t, moved, sig); got != PayloadRenderingNone {
				t.Errorf("%s: a signature verified against a DIFFERENT sign doc as %d", tc.name, got)
			}
		})
	}

	if got := verify(t, doc, []byte("not a signature")); got != PayloadRenderingNone {
		t.Errorf("garbage verified as a signature, rendering %d", got)
	}
}

// unregisteredMsg is a Msg amino cannot encode: no package registers it.
type unregisteredMsg struct{}

func (unregisteredMsg) Route() string                { return "" }
func (unregisteredMsg) Type() string                 { return "" }
func (unregisteredMsg) ValidateBasic() error         { return nil }
func (unregisteredMsg) GetSignBytes() []byte         { return nil }
func (unregisteredMsg) GetSigners() []crypto.Address { return nil }

// A payload that cannot be built is not a bad signature. Callers report the two
// differently -- one is a malformed transaction, the other a forgery -- so the
// helper has to keep them apart instead of folding both into "false".
func TestVerifySignaturePayloadReportsUnencodableMsgs(t *testing.T) {
	t.Parallel()

	priv := ed25519.GenPrivKey()
	doc := SignDoc{ChainID: "dev", Msgs: []Msg{unregisteredMsg{}}}

	rendering, err := VerifySignaturePayload(priv.PubKey(), doc, []byte("irrelevant"))
	if err == nil {
		t.Fatal("an unencodable sign doc verified without error")
	}
	if rendering != PayloadRenderingNone {
		t.Errorf("an unencodable sign doc verified as rendering %d", rendering)
	}
}
