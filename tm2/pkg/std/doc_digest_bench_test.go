package std

import (
	"strings"
	"testing"
)

// The cost claims in tm2/adr/prxxxx_digest_sign_mode.md are made here, because
// a third rendering on the consensus path is unmetered work and "it is cheap"
// is not a reviewable statement. Run with:
//
//	go test ./tm2/pkg/std/ -run XXX -bench Digest -benchmem
//
// Two fixtures, both about 1MB, which is the shape an attacker sends with a bad
// signature attached:
//
//   - digestable: twenty 50KB bodies, every one replaced.
//   - identity: 7,650 hundred-byte bodies, none replaceable. This is the
//     cheapest 1MB a node can be made to reject, and the case the scan exists
//     for.
func benchDoc(files, size int) SignDoc {
	body := strings.Repeat("x", size)
	msgs := make([]Msg, 0, 1)
	bodies := make([]string, files)
	for i := range bodies {
		bodies[i] = body
	}

	// A memo carries the payload rather than a message type, so that this
	// benchmark does not need one gno.land defines.
	msgs = msgs[:0]

	return SignDoc{
		ChainID: "dev",
		Fee:     NewFee(200000, Coin{Denom: "ugnot", Amount: 1000000}),
		Msgs:    msgs,
		Memo:    strings.Join(bodies, "\x20"),
	}
}

func digestableDoc() SignDoc { return benchDoc(20, 50_000) }

// identityPayload is a payload of many short strings: nothing in it can be
// digested, so the whole cost of the digest arm is waste.
func identityPayload(tb testing.TB) []byte {
	tb.Helper()

	// Many short strings in one object, which is what a package of small files
	// renders to. Built directly, since no std.Msg here has that shape.
	var out strings.Builder
	out.WriteString(`{"account_number":"0","chain_id":"dev","fee":{"amount":[],"gas":"0"},"memo":"","msgs":[{"files":[`)
	for i := range 7650 {
		if i > 0 {
			out.WriteByte(',')
		}
		out.WriteString(`{"body":"`)
		out.WriteString(strings.Repeat("y", 100))
		out.WriteString(`","name":"f.gno"}`)
	}
	out.WriteString(`]}],"sequence":"0"}`)

	return []byte(out.String())
}

// BenchmarkDigestFullRender is the cost of building a rendering from the
// transaction, which is what the ante handler used to pay a third time before
// VerifySignaturePayloadFrom took the payload it already had.
func BenchmarkDigestFullRender(b *testing.B) {
	doc := digestableDoc()

	for b.Loop() {
		if _, err := GetSignaturePayload(doc); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDigestTransformInHand is what the digest arm costs now: the
// substitution applied to the payload the caller already holds.
func BenchmarkDigestTransformInHand(b *testing.B) {
	payload, err := GetSignaturePayload(digestableDoc())
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, _, err := digestPayload(payload, nil); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDigestIdentityScan is the digest arm on a transaction that cannot
// have been signed in digest mode: the scan answers no and nothing is parsed.
func BenchmarkDigestIdentityScan(b *testing.B) {
	payload := identityPayload(b)

	for b.Loop() {
		if _, _, err := digestPayload(payload, nil); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDigestIdentityWalk is the same case with the scan bypassed, which is
// what the first draft of this change paid. The gap between this and
// BenchmarkDigestIdentityScan is the reason the scan is hand-written.
func BenchmarkDigestIdentityWalk(b *testing.B) {
	payload := identityPayload(b)
	if !payloadMayHoldLongString(payload) {
		// Confirms the fixture is the identity case, and that the walk below
		// is measuring wasted work rather than real substitution.
		b.Log("fixture holds no oversized string, as intended")
	}

	for b.Loop() {
		if _, _, err := digestWalk(payload, nil); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDigestReportedWalk is the walk with the path and key-ordering work
// the client needs and the consensus path does not, against
// BenchmarkDigestIdentityWalk which asks for no report.
func BenchmarkDigestReportedWalk(b *testing.B) {
	payload := identityPayload(b)

	for b.Loop() {
		var subs []Digested
		if _, _, err := digestWalk(payload, &subs); err != nil {
			b.Fatal(err)
		}
	}
}
