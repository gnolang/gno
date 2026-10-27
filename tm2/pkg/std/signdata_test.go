package std

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The test1 key (gnoland's default test mnemonic).
const signDataTestSigner = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

// Signatures made by @cosmjs/amino from the test1 mnemonic at m/44'/118'/0'/0/0
// over the payloads below. They come from another implementation on purpose:
// agreeing with ourselves would prove nothing.
var signDataVectors = []struct {
	name    string
	data    string
	payload string
	sig     string // base64, 64 bytes R||S
}{
	{
		name: "ascii",
		data: "hello",
		payload: `{"account_number":"0","chain_id":"gnoland-1","fee":{"amount":[],"gas":"0"},"memo":"",` +
			`"msgs":[{"type":"sign/MsgSignData","value":{"data":"hello","origin":"https://gno.land",` +
			`"purpose":"auth/SignIn","signer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}}],"sequence":"0"}`,
		sig: "/MurcSWJpSzxdBDx+lwL4Ldo/n2J7NQ7vywYD1/qj7MSZ0Ms5QJ1E/xami0mU/EETYdC5BQDGJCUf/EhHJSAdw==",
	},
	{
		name: "non-ascii",
		data: "Sign in to Gnoland · 서명 · 🔑",
		payload: `{"account_number":"0","chain_id":"gnoland-1","fee":{"amount":[],"gas":"0"},"memo":"",` +
			`"msgs":[{"type":"sign/MsgSignData","value":{"data":"Sign in to Gnoland · 서명 · 🔑",` +
			`"origin":"https://gno.land","purpose":"auth/SignIn","signer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}}],"sequence":"0"}`,
		sig: "qU9OIY0id/n22g621wUo3iyrrBtrgk28AGFii7jQrUU+/YwPWcajzX12ZuyL2CUmY62606/1cT7Gs4K+tnx/Ug==",
	},
	{
		// Every character an encoder escapes in a text the data may hold.
		name: "escapes",
		data: "Sign in to <gno.land> & \"Gnoland\"\npath: C:\\gno\nnonce: 9UQVko4o",
		payload: `{"account_number":"0","chain_id":"gnoland-1","fee":{"amount":[],"gas":"0"},"memo":"",` +
			`"msgs":[{"type":"sign/MsgSignData","value":{"data":"Sign in to \u003cgno.land\u003e \u0026 \"Gnoland\"\npath: C:\\gno\nnonce: 9UQVko4o",` +
			`"origin":"https://gno.land","purpose":"auth/SignIn","signer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}}],"sequence":"0"}`,
		sig: "hcquZkFbTSN9W0whGiNUA5ZRBVbpkYa+BmEUN08dpvdo7atfbCagVaNCZvZ9595xmtuM3fCntIwg+aAGBvTXFw==",
	},
}

// signDataVectorDoc returns what the vector with data says.
func signDataVectorDoc(data string) SignedData {
	return SignedData{
		ChainID: "gnoland-1",
		Origin:  "https://gno.land",
		Purpose: "auth/SignIn",
		Signer:  crypto.MustAddressFromString(signDataTestSigner),
		Data:    data,
	}
}

// The compressed public key of signDataTestSigner, base64.
const signDataTestPubKey = "A+FhNtsXHjLfSJk1lB8FbiL4mGPjc50Kt81J7EKDnJ2y"

func signDataTestKey(t *testing.T) secp256k1.PubKeySecp256k1 {
	t.Helper()

	raw, err := base64.StdEncoding.DecodeString(signDataTestPubKey)
	require.NoError(t, err)

	var pub secp256k1.PubKeySecp256k1
	copy(pub[:], raw)

	return pub
}

func TestSignDataPayload_MatchesOtherImplementations(t *testing.T) {
	t.Parallel()

	for _, v := range signDataVectors {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			payload, err := SignDataPayload(signDataVectorDoc(v.data))
			require.NoError(t, err)
			assert.Equal(t, v.payload, string(payload))
		})
	}
}

// encodeCharacterSet (tm2-js-client) and serializeSignDoc (cosmjs) both escape
// <, > and & in the payload; a verifier rebuilding the bytes must do the same.
func TestSignDataPayload_EscapesHTMLCharacters(t *testing.T) {
	t.Parallel()

	d := signDataVectorDoc("hello")
	d.ChainID = "a<b>c&d"
	payload, err := SignDataPayload(d)
	require.NoError(t, err)

	// From cosmjs serializeSignDoc.
	assert.Equal(t,
		`{"account_number":"0","chain_id":"a\u003cb\u003ec\u0026d","fee":{"amount":[],"gas":"0"},"memo":"",`+
			`"msgs":[{"type":"sign/MsgSignData","value":{"data":"hello","origin":"https://gno.land",`+
			`"purpose":"auth/SignIn","signer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}}],"sequence":"0"}`,
		string(payload),
	)
}

func TestSignDataPayload_Rejects(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name   string
		change func(*SignedData)
	}{
		// Without a chain id a proof made for one network is a proof on all.
		{"empty chain id", func(d *SignedData) { d.ChainID = "" }},
		// Without an origin or a purpose, a proof made for one service or one
		// use passes at another.
		{"empty origin", func(d *SignedData) { d.Origin = "" }},
		{"empty purpose", func(d *SignedData) { d.Purpose = "" }},
		{"zero signer", func(d *SignedData) { d.Signer = crypto.Address{} }},
		{"empty data", func(d *SignedData) { d.Data = "" }},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			d := signDataVectorDoc("hello")
			testCase.change(&d)
			_, err := SignDataPayload(d)
			assert.Error(t, err)
		})
	}
}

// One service has one origin, so a verifier compares it as a string: only the
// form a browser writes for a page's origin is signed.
func TestSignDataPayload_Origin(t *testing.T) {
	t.Parallel()

	valid := []string{
		"https://gno.land",
		"https://xn--bcher-kva.example",
		"http://localhost:8888",
		"https://127.0.0.1:8443",
		"http://[::1]:26657",
		"https://[2001:db8::1]",
		"com.example.game://callback",
	}
	invalid := []string{
		"gno.land",
		"https://",
		"//gno.land",
		"HTTPS://gno.land",
		"https://Gno.land",
		"https://gno.land/",
		"https://gno.land/r/demo",
		"https://gno.land?x=1",
		"https://gno.land#x",
		"https://user@gno.land",
		"https://gno.land.",
		"https://gno..land",
		"https://-gno.land",
		"https://bücher.example",
		"https://gno_land.example",
		"https://gno.land:443",
		"http://gno.land:80",
		"https://gno.land:",
		"https://gno.land:0",
		"https://gno.land:08443",
		"https://gno.land:65536",
		"https://127.1",
		"https://[::ffff:127.0.0.1]",
		"https://[2001:DB8::1]",
		"https://[2001:db8:0:0:0:0:0:1]",
		"https://[fe80::1%25en0]",
		"https://[::1",
		"https://[::1]x",
		"1https://gno.land",
		"https://" + strings.Repeat("a.", 130) + "land",
	}

	testSignDataField(t, "origin", func(d *SignedData, v string) { d.Origin = v }, valid, invalid)
}

// The data is a text a user reads before signing, and every JSON encoder writes
// it the same way.
func TestSignDataPayload_Data(t *testing.T) {
	t.Parallel()

	valid := []string{
		"hello",
		"line one\nline two",
		`a "quote" and a \ backslash`,
		"<b>&amp;</b>",
		"서명 · 🔑",
		"\u202e", // a wallet shows it visibly
	}
	invalid := []string{
		"\xff\xfe",
		"a\tb",
		"line one\r\nline two",
		"a\x00b",
		"\x1b[31mred",
		"a\x7fb",
		"a\u0085b",
		"a\u2028b",
		"a\u2029b",
	}

	testSignDataField(t, "sign-data data", func(d *SignedData, v string) { d.Data = v }, valid, invalid)
}

// A wallet shows the purpose as it is, so it holds nothing that could reshape
// the approval screen.
func TestSignDataPayload_Purpose(t *testing.T) {
	t.Parallel()

	valid := []string{"auth/SignIn", "username/Reserve", "v1.link-account_2", strings.Repeat("a", 64)}
	invalid := []string{
		"sign in",
		"auth/SignIn\nApprove: safe",
		"auth\tSignIn",
		"auth/SignIn\u202e",
		"аuth/SignIn", // a Cyrillic а
		"auth:SignIn",
		"<b>auth</b>",
		"\xff\xfe",
		strings.Repeat("a", 65),
	}

	testSignDataField(t, "purpose", func(d *SignedData, v string) { d.Purpose = v }, valid, invalid)
}

// testSignDataField checks that SignDataPayload signs the vector with each of
// valid set into it, and refuses it, naming field, with each of invalid.
func testSignDataField(t *testing.T, field string, set func(*SignedData, string), valid, invalid []string) {
	t.Helper()

	run := func(v string, ok bool) {
		t.Run(v, func(t *testing.T) {
			t.Parallel()

			d := signDataVectorDoc("hello")
			set(&d, v)
			_, err := SignDataPayload(d)
			if ok {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, field)
			}
		})
	}
	for _, v := range valid {
		run(v, true)
	}
	for _, v := range invalid {
		run(v, false)
	}
}

func TestVerifySignData_AcceptsOtherImplementations(t *testing.T) {
	t.Parallel()

	pub := signDataTestKey(t)

	for _, v := range signDataVectors {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			sig, err := base64.StdEncoding.DecodeString(v.sig)
			require.NoError(t, err)

			signed, err := VerifySignData([]byte(v.payload), pub, sig)
			require.NoError(t, err)
			assert.Equal(t, signDataVectorDoc(v.data), signed)
		})
	}
}

// A verifier gets the document back as an object (the proof's `signed`) and may
// serialize it in any key order, spacing or escaping; the signature is over the
// canonical bytes, which the verifier rebuilds.
func TestVerifySignData_AcceptsAnyRenderingOfTheDocument(t *testing.T) {
	t.Parallel()

	v := signDataVectors[0]
	sig, err := base64.StdEncoding.DecodeString(v.sig)
	require.NoError(t, err)

	// The document as a JavaScript client would stringify it, in its own key order
	// and with an escape of its own.
	doc := `{
	  "chain_id": "gnoland-1", "account_number": "0", "sequence": "0",
	  "fee": {"amount": [], "gas": "0"},
	  "msgs": [{"type": "sign/MsgSignData", "value": {"signer": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
	    "data": "h\u0065llo", "origin": "https://gno.land", "purpose": "auth/SignIn"}}],
	  "memo": ""
	}`

	signed, err := VerifySignData([]byte(doc), signDataTestKey(t), sig)
	require.NoError(t, err)
	assert.Equal(t, "hello", signed.Data)
}

func TestVerifySignData_RoundTrip(t *testing.T) {
	t.Parallel()

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey().(secp256k1.PubKeySecp256k1)
	d := signDataVectorDoc("prove it")
	d.Signer = pub.Address()

	payload, err := SignDataPayload(d)
	require.NoError(t, err)

	sig, err := priv.Sign(payload)
	require.NoError(t, err)

	signed, err := VerifySignData(payload, pub, sig)
	require.NoError(t, err)
	assert.Equal(t, d, signed)
}

func TestVerifySignData_Rejects(t *testing.T) {
	t.Parallel()

	v := signDataVectors[0]
	sig, err := base64.StdEncoding.DecodeString(v.sig)
	require.NoError(t, err)

	pub := signDataTestKey(t)
	other := secp256k1.GenPrivKey().PubKey().(secp256k1.PubKeySecp256k1)
	// msg is the vector's one message.
	msg := v.payload[strings.Index(v.payload, `{"type"`) : strings.Index(v.payload, `}}]`)+2]

	// replace returns the pinned payload with one substring changed.
	replace := func(from, to string) string {
		require.Contains(t, v.payload, from)
		return strings.Replace(v.payload, from, to, 1)
	}

	testTable := []struct {
		name string
		doc  string
		pub  secp256k1.PubKeySecp256k1
		sig  []byte
	}{
		// The claim a verifier relies on: the key it is handed is the signer's.
		{"a key that is not the signer's", v.payload, other, sig},
		{"tampered data", replace(`"data":"hello"`, `"data":"hellp"`), pub, sig},
		{"another chain", replace(`"gnoland-1"`, `"dev"`), pub, sig},
		{"another origin", replace(`"https://gno.land"`, `"https://evil.example"`), pub, sig},
		{"another purpose", replace(`"auth/SignIn"`, `"auth/Other"`), pub, sig},
		{"a corrupted signature", v.payload, pub, append([]byte{sig[0] ^ 1}, sig[1:]...)},
		// The shape is fixed, so that nothing but this document can pass for it.
		{"an account number", replace(`"account_number":"0"`, `"account_number":"1"`), pub, sig},
		{"a sequence", replace(`"sequence":"0"`, `"sequence":"1"`), pub, sig},
		{"gas", replace(`"gas":"0"`, `"gas":"1"`), pub, sig},
		{"a fee coin", replace(`"amount":[]`, `"amount":[{"amount":"1","denom":"ugnot"}]`), pub, sig},
		{"a memo", replace(`"memo":""`, `"memo":"x"`), pub, sig},
		{"another message type", replace(`"sign/MsgSignData"`, `"bank/MsgSend"`), pub, sig},
		{"two messages", replace(`"msgs":[`, `"msgs":[`+msg+`,`), pub, sig},
		{"an unknown field", replace(`"memo":""`, `"memo":"","extra":1`), pub, sig},
		{"data with a line separator", replace(`"data":"hello"`, `"data":"hel\u2028lo"`), pub, sig},
		{"a signer that is not an address", replace(signDataTestSigner, "g1nope"), pub, sig},
		{"not JSON", "hello", pub, sig},
		// Beyond its JSON rendering, a document must be the canonical one,
		// even where Go would read the same values from it.
		{"a missing memo", replace(`"memo":"",`, ``), pub, sig},
		{"a missing origin", replace(`"origin":"https://gno.land",`, ``), pub, sig},
		// ADR-036's {signer, data}, which binds a proof to no service or use.
		{"no origin and no purpose", replace(`"origin":"https://gno.land","purpose":"auth/SignIn",`, ``), pub, sig},
		{"a null fee amount", replace(`"amount":[]`, `"amount":null`), pub, sig},
		{"a number for a string", replace(`"account_number":"0"`, `"account_number":0`), pub, sig},
		{"a key in another case", replace(`"chain_id"`, `"Chain_ID"`), pub, sig},
		{"data after the document", v.payload + "{}", pub, sig},
		// Parsers disagree on which value a repeated key holds.
		{"a repeated chain id", replace(`"chain_id":"gnoland-1"`, `"chain_id":"evil","chain_id":"gnoland-1"`), pub, sig},
		{"repeated data", replace(`"data":"hello"`, `"data":"evil","data":"hello"`), pub, sig},
		// Nested past Unmarshal's depth limit, so the duplicate-key walk never
		// recurses that deep.
		{"deep nesting", strings.Repeat("[", 3_000_000) + strings.Repeat("]", 3_000_000), pub, sig},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := VerifySignData([]byte(testCase.doc), testCase.pub, testCase.sig)
			assert.Error(t, err)
		})
	}
}

// signDataTestProof is the proof for the first vector.
var signDataTestProof = `{
  "signed": ` + signDataVectors[0].payload + `,
  "signature": {
    "pubKey": {"typeUrl": "/tm.PubKeySecp256k1", "value": "` + signDataTestPubKey + `"},
    "signature": "` + signDataVectors[0].sig + `"
  }
}`

func TestVerifySignDataProof_AcceptsOtherImplementations(t *testing.T) {
	t.Parallel()

	signed, err := VerifySignDataProof([]byte(signDataTestProof))
	require.NoError(t, err)
	assert.Equal(t, signDataVectorDoc("hello"), signed)
}

func TestVerifySignDataProof_RoundTrip(t *testing.T) {
	t.Parallel()

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey().(secp256k1.PubKeySecp256k1)

	d := signDataVectorDoc("prove it")
	d.Signer = pub.Address()
	payload, err := SignDataPayload(d)
	require.NoError(t, err)
	sig, err := priv.Sign(payload)
	require.NoError(t, err)

	proof, err := json.Marshal(NewSignDataProof(payload, pub, sig))
	require.NoError(t, err)

	signed, err := VerifySignDataProof(proof)
	require.NoError(t, err)
	assert.Equal(t, d, signed)
}

func TestVerifySignDataProof_Rejects(t *testing.T) {
	t.Parallel()

	// replace returns the proof with one substring changed.
	replace := func(from, to string) string {
		require.Contains(t, signDataTestProof, from)
		return strings.Replace(signDataTestProof, from, to, 1)
	}

	testTable := []struct {
		name    string
		proof   string
		wantErr string
	}{
		{"another key type", replace(SignDataPubKeyType, "/tm.PubKeyEd25519"), `"/tm.PubKeyEd25519"`},
		{"a key that is not base64", replace(signDataTestPubKey, "not base64!"), "public key"},
		{"a truncated key", replace(signDataTestPubKey, signDataTestPubKey[:40]), "public key"},
		{"a signature that is not base64", replace(signDataVectors[0].sig, "not base64!"), "signature is not base64"},
		{"a corrupted signature", replace(signDataVectors[0].sig, "9"+signDataVectors[0].sig[1:]), "invalid sign-data signature"},
		{"tampered data", replace(`"data":"hello"`, `"data":"hellp"`), "invalid sign-data signature"},
		{"not JSON", "hello", "unable to read sign-data proof"},
		{"a repeated signed document", replace(`"signed": `, `"signed": {}, "signed": `), `duplicate key "signed"`},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := VerifySignDataProof([]byte(testCase.proof))
			assert.ErrorContains(t, err, testCase.wantErr)
		})
	}
}
