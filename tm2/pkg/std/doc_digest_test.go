package std

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// ordinaryDocs are sign docs no digest substitution should touch. The node's
// third verification arm is gated on the digest rendering coming out identical
// to the current one for exactly these, so the list doubles as the set of
// transactions the digest mode must cost nothing.
func ordinaryDocs() []struct {
	name string
	doc  SignDoc
} {
	return []struct {
		name string
		doc  SignDoc
	}{
		{"ordinary", SignDoc{
			ChainID: "dev", AccountNumber: 42, Sequence: 7,
			Fee: NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}), Memo: "hello",
		}},
		{"zero fee", SignDoc{ChainID: "dev"}},
		{"large numbers", SignDoc{
			ChainID: "dev", AccountNumber: 1 << 62, Sequence: 1 << 62,
			Fee: NewFee(1<<62, Coin{Denom: "ugnot", Amount: 1 << 62}),
		}},
		{"memo at the boundary", SignDoc{
			ChainID: "dev",
			Fee:     NewFee(1, Coin{Denom: "ugnot", Amount: 1}),
			Memo:    strings.Repeat("m", MaxSignDocStringLen),
		}},
		{"memo holding the digest key as text", SignDoc{
			ChainID: "dev",
			Fee:     NewFee(1, Coin{Denom: "ugnot", Amount: 1}),
			Memo:    `{"` + digestKey + `":"` + strings.Repeat("ab", 32) + `"}`,
		}},
	}
}

// THE PROPERTY THE NODE RELIES ON. VerifySignaturePayload skips its third curve
// operation when the digest rendering comes out equal to the current one, and
// the ante handler does the same inline. If this ever stops holding, every
// transaction pays for a verification it cannot need, and the cost lands on the
// node rather than the sender.
func TestDigestIsIdentityWithoutOversizedStrings(t *testing.T) {
	t.Parallel()

	for _, tc := range ordinaryDocs() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			full, err := GetSignaturePayload(tc.doc)
			if err != nil {
				t.Fatal(err)
			}
			digested, _, err := GetSignaturePayloadDigest(tc.doc)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(full, digested) {
				t.Errorf("the digest rendering changed a payload with no oversized string\n full:   %s\n digest: %s", full, digested)
			}

			_, fields, err := GetSignaturePayloadDigest(tc.doc)
			if err != nil {
				t.Fatal(err)
			}
			if len(fields) != 0 {
				t.Errorf("reported %d digested field(s) where none was oversized: %v", len(fields), fields)
			}
		})
	}
}

// The boundary is exclusive, and it is the DECODED length that counts. A signer
// reading MaxSignDocStringLen has to be able to predict which fields the device
// will show them.
func TestDigestBoundaryIsExclusive(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		length     int
		wantDigest bool
	}{
		{"one under", MaxSignDocStringLen - 1, false},
		{"exactly at", MaxSignDocStringLen, false},
		{"one over", MaxSignDocStringLen + 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := SignDoc{ChainID: "dev", Memo: strings.Repeat("m", tc.length)}
			_, fields, err := GetSignaturePayloadDigest(doc)
			if err != nil {
				t.Fatal(err)
			}

			if got := len(fields) == 1; got != tc.wantDigest {
				t.Fatalf("a %d byte string digested=%v, want %v", tc.length, got, tc.wantDigest)
			}
			if !tc.wantDigest {
				return
			}
			if fields[0].Path != "memo" {
				t.Errorf("reported path %q, want memo", fields[0].Path)
			}
			if fields[0].Length != tc.length {
				t.Errorf("reported length %d, want %d", fields[0].Length, tc.length)
			}
		})
	}
}

// The digest is over the decoded string, and it is the one a signer compares
// against the device. A client that showed a digest of anything else -- the
// escaped form, say -- would have them compare two unrelated numbers and
// conclude the device agreed with them.
func TestDigestIsSHA256OfTheDecodedString(t *testing.T) {
	t.Parallel()

	// Newlines matter: they are one byte decoded and two escaped, so this
	// fails if the hash is taken over the JSON form.
	body := "package foo\n" + strings.Repeat("// a line of source\n", 100)
	doc := SignDoc{ChainID: "dev", Memo: body}

	sum := sha256.Sum256([]byte(body))
	want := hex.EncodeToString(sum[:])

	_, fields, err := GetSignaturePayloadDigest(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 {
		t.Fatalf("got %d digested fields, want 1", len(fields))
	}
	if fields[0].Sum != want {
		t.Errorf("reported digest %s, want %s", fields[0].Sum, want)
	}

	payload, _, err := GetSignaturePayloadDigest(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(payload, []byte(`"memo":{"`+digestKey+`":"`+want+`"}`)) {
		t.Errorf("the payload does not carry the digest object for the memo: %s", payload)
	}
	if bytes.Contains(payload, []byte("a line of source")) {
		t.Errorf("the payload still carries the string it was supposed to replace: %s", payload)
	}
}

// The walk has to reach a field at any depth, since the fields that overflow in
// practice are nested: msgs[0].package.files[1].body. Exercised against the
// internal function so the test does not need a registered message type with a
// deep string field.
func TestDigestWalksNestedValues(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", MaxSignDocStringLen+1)
	sum := sha256.Sum256([]byte(long))
	hexSum := hex.EncodeToString(sum[:])

	payload := fmt.Sprintf(
		`{"msgs":[{"package":{"files":[{"body":"short","name":"a.gno"},{"body":%q,"name":"b.gno"}]}}],"memo":%q}`,
		long, long,
	)

	var subs []Digested

	reduced, substituted, err := digestPayload([]byte(payload), &subs)
	if err != nil {
		t.Fatal(err)
	}
	if !substituted {
		t.Fatal("the walk reported no substitution")
	}

	wantPaths := []string{"memo", "msgs[0].package.files[1].body"}
	gotPaths := make([]string, len(subs))
	for i, sub := range subs {
		gotPaths[i] = sub.Path
		if sub.Sum != hexSum {
			t.Errorf("%s: digest %s, want %s", sub.Path, sub.Sum, hexSum)
		}
	}
	if fmt.Sprint(gotPaths) != fmt.Sprint(wantPaths) {
		t.Errorf("reported paths %v, want %v", gotPaths, wantPaths)
	}

	// Keys stay sorted and the short body is left alone, so what the device
	// displays is still the canonical payload with holes in it.
	want := fmt.Sprintf(
		`{"memo":{"%[1]s":"%[2]s"},"msgs":[{"package":{"files":[{"body":"short","name":"a.gno"},{"body":{"%[1]s":"%[2]s"},"name":"b.gno"}]}}]}`,
		digestKey, hexSum,
	)
	if string(reduced) != want {
		t.Errorf("digest payload\n got:  %s\n want: %s", reduced, want)
	}
}

// digestRendering adapts GetSignaturePayloadDigest to the shape the rendering
// tables below use. The substitutions are what the client wants; a test
// comparing renderings wants only the bytes.
func digestRendering(s SignDoc) ([]byte, error) {
	payload, _, err := GetSignaturePayloadDigest(s)

	return payload, err
}

// THE SCAN MUST NEVER MISS AN OVERSIZED STRING. payloadMayHoldLongString decides
// whether the payload is parsed at all, so a false negative would make a node
// read a digest-signed transaction as unsigned and reject what its peers accept.
// It measures escaped lengths, which are never shorter than decoded ones, so the
// direction holds; the newline case is the one that would break a scan comparing
// decoded lengths instead.
func TestPayloadScanHasNoFalseNegatives(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		memo string
	}{
		{"plain", strings.Repeat("m", MaxSignDocStringLen+1)},
		{"newlines, which escape to two bytes each", strings.Repeat("\n", MaxSignDocStringLen+1)},
		{"quotes, which escape", strings.Repeat(`"`, MaxSignDocStringLen+1)},
		{"runes, which escape to six bytes each", strings.Repeat("\x01", MaxSignDocStringLen+1)},
		{"a string ending in a backslash", strings.Repeat("m", MaxSignDocStringLen) + `\`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := SignDoc{ChainID: "dev", Memo: tc.memo}
			payload, err := GetSignaturePayload(doc)
			if err != nil {
				t.Fatal(err)
			}
			if !payloadMayHoldLongString(payload) {
				t.Fatalf("the scan missed a %d byte string, so its payload is never parsed", len(tc.memo))
			}

			// And the rendering it gates does substitute.
			_, fields, err := GetSignaturePayloadDigest(doc)
			if err != nil {
				t.Fatal(err)
			}
			if len(fields) != 1 {
				t.Fatalf("got %d substitutions, want 1", len(fields))
			}
		})
	}
}

// A false POSITIVE is allowed and must stay harmless: the payload gets parsed,
// nothing is substituted, and the rendering has to come back as the input
// itself, or the node would verify a second identical payload and the client
// would report fields nobody hid.
func TestPayloadScanFalsePositiveStaysIdentity(t *testing.T) {
	t.Parallel()

	// 400 newlines decode to 400 bytes and escape to 800, so the scan says
	// "maybe" and the walk says no.
	doc := SignDoc{ChainID: "dev", Memo: strings.Repeat("\n", 400)}

	payload, err := GetSignaturePayload(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !payloadMayHoldLongString(payload) {
		t.Skip("escaping no longer inflates this memo past the threshold; the case is gone")
	}

	digested, fields, err := GetSignaturePayloadDigest(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 0 {
		t.Errorf("reported %d substitutions for a string under the threshold: %v", len(fields), fields)
	}
	if !bytes.Equal(digested, payload) {
		t.Errorf("the rendering changed although nothing was substituted\n got:  %s\n want: %s", digested, payload)
	}
}

// findDigestObjects reports the paths at which tree holds an object keyed by
// digestKey.
func findDigestObjects(tree any, path string, found *[]string) {
	switch t := tree.(type) {
	case map[string]any:
		if _, ok := t[digestKey]; ok {
			*found = append(*found, path)
		}
		for key, value := range t {
			findDigestObjects(value, joinDigestPath(path, key), found)
		}
	case []any:
		for i, item := range t {
			findDigestObjects(item, fmt.Sprintf("%s[%d]", path, i), found)
		}
	}
}

// THE DISJOINTNESS ARGUMENT, as a test. The digest rendering differs from the
// other two only where it holds an object keyed digestKey, so it can be
// mistaken for a full rendering of a DIFFERENT transaction only if a full
// rendering can hold such an object. Amino keys come from Go struct tags and
// none begins with '$', so none can -- not even when an attacker writes the key
// into a field they control, because a string stays a string.
func TestDigestKeyIsUnreachableFromAmino(t *testing.T) {
	t.Parallel()

	if !strings.HasPrefix(digestKey, "$") {
		t.Fatalf("digestKey %q no longer begins with '$', which is what kept it out of reach of amino", digestKey)
	}

	hostile := []SignDoc{
		{ChainID: "dev", Memo: `{"` + digestKey + `":"` + strings.Repeat("ab", 32) + `"}`},
		{ChainID: `dev","memo":{"` + digestKey + `":"x"}`, Memo: "x"},
		{ChainID: "dev", Memo: strings.Repeat("m", MaxSignDocStringLen+1)},
	}

	for i, doc := range hostile {
		for _, render := range []struct {
			name string
			fn   func(SignDoc) ([]byte, error)
		}{
			{"current", GetSignaturePayload},
			{"legacy", GetSignaturePayloadLegacy},
		} {
			payload, err := render.fn(doc)
			if err != nil {
				t.Fatal(err)
			}

			var tree any
			if err := json.Unmarshal(payload, &tree); err != nil {
				t.Fatal(err)
			}
			var found []string
			findDigestObjects(tree, "", &found)
			if len(found) != 0 {
				t.Errorf("doc %d: the %s rendering holds a %s object at %v; it can now be "+
					"confused with the digest rendering of a different transaction",
					i, render.name, digestKey, found)
			}
		}
	}
}

// The concrete pair the argument above is about: one document whose field is
// digested, another whose field spells out the digest object as text.
func TestDigestRenderingIsDisjoint(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", MaxSignDocStringLen+1)
	sum := sha256.Sum256([]byte(long))
	hexSum := hex.EncodeToString(sum[:])

	digestedDoc := SignDoc{ChainID: "dev", Memo: long}
	impostors := []SignDoc{
		{ChainID: "dev", Memo: `{"` + digestKey + `":"` + hexSum + `"}`},
		{ChainID: "dev", Memo: hexSum},
		{ChainID: "dev", Memo: digestKey + ":" + hexSum},
	}

	digested, _, err := GetSignaturePayloadDigest(digestedDoc)
	if err != nil {
		t.Fatal(err)
	}

	for i, doc := range impostors {
		for _, render := range []struct {
			name string
			fn   func(SignDoc) ([]byte, error)
		}{
			{"current", GetSignaturePayload},
			{"legacy", GetSignaturePayloadLegacy},
			{"digest", digestRendering},
		} {
			payload, err := render.fn(doc)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(payload, digested) {
				t.Errorf("impostor %d: its %s rendering equals the digest rendering of a "+
					"different document, so one signature authorises both: %s",
					i, render.name, payload)
			}
		}
	}
}

// A digest-mode signature has to verify, report itself as such, and still bind
// one transaction: the point of hashing the oversized field is that changing it
// changes the payload.
func TestVerifySignaturePayloadTakesDigestRendering(t *testing.T) {
	t.Parallel()

	priv := ed25519.GenPrivKey()
	doc := SignDoc{
		ChainID: "dev", AccountNumber: 3, Sequence: 4,
		Fee:  NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}),
		Memo: strings.Repeat("a", MaxSignDocStringLen+1),
	}

	payload, _, err := GetSignaturePayloadDigest(doc)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := priv.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	rendering, err := VerifySignaturePayload(priv.PubKey(), doc, sig)
	if err != nil {
		t.Fatal(err)
	}
	if rendering != PayloadRenderingDigest {
		t.Errorf("a digest signature verified as rendering %d, want %d", rendering, PayloadRenderingDigest)
	}

	// IT STILL BINDS THE DIGESTED CONTENT. The signer could not read the field
	// on the device, so the only thing standing between them and a swapped
	// file body is that a different body is a different payload.
	swapped := doc
	swapped.Memo = strings.Repeat("b", MaxSignDocStringLen+1)
	if rendering, err := VerifySignaturePayload(priv.PubKey(), swapped, sig); err != nil {
		t.Fatal(err)
	} else if rendering != PayloadRenderingNone {
		t.Errorf("the signature verified against a document with a DIFFERENT digested field, as %d", rendering)
	}

	// And it still binds everything the device did display.
	moved := doc
	moved.Sequence = doc.Sequence + 1
	if rendering, err := VerifySignaturePayload(priv.PubKey(), moved, sig); err != nil {
		t.Fatal(err)
	} else if rendering != PayloadRenderingNone {
		t.Errorf("the signature verified against a different sequence, as %d", rendering)
	}
}

// A signature over the full rendering keeps reporting itself as the current
// rendering even when the document holds an oversized field, so the order of
// the arms cannot silently relabel ordinary signatures.
func TestFullRenderingStillWinsOnOversizedDocs(t *testing.T) {
	t.Parallel()

	priv := ed25519.GenPrivKey()
	doc := SignDoc{ChainID: "dev", Memo: strings.Repeat("a", MaxSignDocStringLen+1)}

	payload, err := GetSignaturePayload(doc)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := priv.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	rendering, err := VerifySignaturePayload(priv.PubKey(), doc, sig)
	if err != nil {
		t.Fatal(err)
	}
	if rendering != PayloadRenderingCurrent {
		t.Errorf("a full-rendering signature verified as %d, want %d", rendering, PayloadRenderingCurrent)
	}
}
