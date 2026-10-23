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

// The test1 key (gnoland's default test mnemonic), which Adena's
// sign-arbitrary spec also uses.
const signDataTestSigner = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

// Signatures made by @cosmjs/crypto, the stack Adena signs with, from the test1
// mnemonic at m/44'/118'/0'/0/0 over the payloads below. They come from another
// implementation on purpose: agreeing with ourselves would prove nothing.
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
			`"msgs":[{"type":"sign/MsgSignData","value":{"data":"aGVsbG8=","purpose":"auth/SignIn",` +
			`"signer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5","url":"https://gno.land"}}],"sequence":"0"}`,
		sig: "gquasFDi6ZQesFU2xD1xrKKuABlpDf4sylnYBMAbgpEGQQoQ2OgkpHxyvOkX5qZJrdpApTg4Ga+1SdHqeV2YPQ==",
	},
	{
		name: "non-ascii",
		data: "Sign in to Gnoland · 서명 · 🔑",
		payload: `{"account_number":"0","chain_id":"gnoland-1","fee":{"amount":[],"gas":"0"},"memo":"",` +
			`"msgs":[{"type":"sign/MsgSignData","value":{"data":"U2lnbiBpbiB0byBHbm9sYW5kIMK3IOyEnOuqhSDCtyDwn5SR",` +
			`"purpose":"auth/SignIn","signer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5","url":"https://gno.land"}}],"sequence":"0"}`,
		sig: "yu7MT7Cdjw7hvbHhFW5zWIncExD+uQ80dMraWBSOvAkONgfvOPSCXENCGuldYm9J5+sIfA5RBg/mYcTIbZ5LoQ==",
	},
}

// signDataVectorDoc returns what the vector with data says.
func signDataVectorDoc(data string) SignedData {
	return SignedData{
		ChainID: "gnoland-1",
		URL:     "https://gno.land",
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
	d.URL = "https://gno.land/?a=<b>&c"
	payload, err := SignDataPayload(d)
	require.NoError(t, err)

	// From cosmjs serializeSignDoc.
	assert.Equal(t,
		`{"account_number":"0","chain_id":"a\u003cb\u003ec\u0026d","fee":{"amount":[],"gas":"0"},"memo":"",`+
			`"msgs":[{"type":"sign/MsgSignData","value":{"data":"aGVsbG8=","purpose":"auth/SignIn",`+
			`"signer":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5","url":"https://gno.land/?a=\u003cb\u003e\u0026c"}}],"sequence":"0"}`,
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
		// Without a url or a purpose, a proof made for one service or one use
		// passes at another.
		{"empty url", func(d *SignedData) { d.URL = "" }},
		{"empty purpose", func(d *SignedData) { d.Purpose = "" }},
		{"zero signer", func(d *SignedData) { d.Signer = crypto.Address{} }},
		{"empty data", func(d *SignedData) { d.Data = "" }},
		// They are texts a user reads before signing, not bytes.
		{"invalid UTF-8 data", func(d *SignedData) { d.Data = "\xff\xfe" }},
		{"invalid UTF-8 url", func(d *SignedData) { d.URL = "\xff\xfe" }},
		{"invalid UTF-8 purpose", func(d *SignedData) { d.Purpose = "\xff\xfe" }},
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

// A verifier gets the document back as an object (Adena's `signed`) and may
// serialize it in any key order, spacing or escaping; the signature is over the canonical
// bytes, which the verifier rebuilds.
func TestVerifySignData_AcceptsAnyRenderingOfTheDocument(t *testing.T) {
	t.Parallel()

	v := signDataVectors[0]
	sig, err := base64.StdEncoding.DecodeString(v.sig)
	require.NoError(t, err)

	// JSON.stringify of Adena's makeArbitraryMessageDoc, in its own key order.
	doc := `{
	  "chain_id": "gnoland-1", "account_number": "0", "sequence": "0",
	  "fee": {"amount": [], "gas": "0"},
	  "msgs": [{"type": "sign/MsgSignData", "value": {"signer": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
	    "data": "aGVsbG8=", "url": "https://gno.land", "purpose": "auth/SignIn"}}],
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
		{"tampered data", replace(`"aGVsbG8="`, `"aGVsbG8h"`), pub, sig},
		{"another chain", replace(`"gnoland-1"`, `"dev"`), pub, sig},
		{"another url", replace(`"https://gno.land"`, `"https://evil.example"`), pub, sig},
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
		{"data that is not base64", replace(`"aGVsbG8="`, `"not base64!"`), pub, sig},
		{"a signer that is not an address", replace(signDataTestSigner, "g1nope"), pub, sig},
		{"not JSON", "hello", pub, sig},
		// Beyond its JSON rendering, a document must be the canonical one,
		// even where Go would read the same values from it.
		{"a missing memo", replace(`"memo":"",`, ``), pub, sig},
		{"a missing url", replace(`,"url":"https://gno.land"`, ``), pub, sig},
		{"a null fee amount", replace(`"amount":[]`, `"amount":null`), pub, sig},
		{"a number for a string", replace(`"account_number":"0"`, `"account_number":0`), pub, sig},
		{"a key in another case", replace(`"chain_id"`, `"Chain_ID"`), pub, sig},
		{"non-canonical base64", replace(`"aGVsbG8="`, `"aGVsbG9="`), pub, sig},
		{"data after the document", v.payload + "{}", pub, sig},
		// Parsers disagree on which value a repeated key holds.
		{"a repeated chain id", replace(`"chain_id":"gnoland-1"`, `"chain_id":"evil","chain_id":"gnoland-1"`), pub, sig},
		{"repeated data", replace(`"data":"aGVsbG8="`, `"data":"ZXZpbA==","data":"aGVsbG8="`), pub, sig},
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

// signDataTestProof is Adena's SignArbitrary answer for the first vector.
var signDataTestProof = `{
  "signed": ` + signDataVectors[0].payload + `,
  "signature": {
    "pubKey": {"typeUrl": "/tm.PubKeySecp256k1", "value": "` + signDataTestPubKey + `"},
    "signature": "` + signDataVectors[0].sig + `"
  }
}`

func TestVerifySignDataProof_AcceptsAdenasAnswer(t *testing.T) {
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
		{"tampered data", replace(`"aGVsbG8="`, `"aGVsbG8h"`), "invalid sign-data signature"},
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
