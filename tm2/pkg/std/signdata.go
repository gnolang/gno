package std

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// SignDataMsgType is the one message of a sign-data document: ADR-036's message
// for off-chain data. A node rebuilds a transaction's sign bytes from the
// decoded transaction, and amino writes every message there as {"@type": ...}:
// a document whose message is {"type", "value"} is no transaction's sign
// document, so a signature over it proves control of an address and can never
// be broadcast. TestSignDataPayloadIsNoTransactionPayload pins this.
const SignDataMsgType = "sign/MsgSignData"

// A sign-data document is a transaction sign document with every field fixed
// but the chain id and the message.
// The chain id stays (ADR-036 wants it empty): a proof made on one network is
// then no proof on another.
type signDataPayload struct {
	ChainID       string        `json:"chain_id"`
	AccountNumber uint64        `json:"account_number"`
	Sequence      uint64        `json:"sequence"`
	Fee           signDocFee    `json:"fee"`
	Msgs          []signDataMsg `json:"msgs"`
	Memo          string        `json:"memo"`
}

// signDataMsg is written in amino's type/value form, which ADR-036 and the
// Ledger Cosmos app use, not as a std.Msg (whose JSON carries "@type").
type signDataMsg struct {
	Type  string        `json:"type"`
	Value signDataValue `json:"value"`
}

// signDataValue is ADR-036's {signer, data}, plus the url and the purpose that
// bind a proof to one service and one use.
type signDataValue struct {
	Signer  string `json:"signer"`
	Data    string `json:"data"` // base64 of the UTF-8 text
	URL     string `json:"url"`
	Purpose string `json:"purpose"`
}

// SignedData is what a sign-data document says: on ChainID, Signer answers the
// service at URL, for Purpose, with Data.
type SignedData struct {
	ChainID string
	// URL names the service the proof is for: a website's origin, which a wallet
	// takes from the requesting page, never from the request.
	URL string
	// Purpose is what the proof is for, as the service names it, such as
	// "username/Reserve".
	Purpose string
	Signer  crypto.Address
	// Data is a text the user reads before signing, typically the service's
	// challenge.
	Data string
}

// SignDataPayload returns the bytes a key signs to state d: the sign document
// sorted by key, its fee an empty coin list (see signDocFee).
func SignDataPayload(d SignedData) ([]byte, error) {
	switch {
	case d.ChainID == "":
		return nil, errors.New("a sign-data document needs a chain id")
	case d.URL == "":
		return nil, errors.New("a sign-data document needs a url")
	case d.Purpose == "":
		return nil, errors.New("a sign-data document needs a purpose")
	case d.Signer.IsZero():
		return nil, errors.New("a sign-data document needs a signer")
	case d.Data == "":
		return nil, errors.New("a sign-data document needs data")
	case !utf8.ValidString(d.URL), !utf8.ValidString(d.Purpose), !utf8.ValidString(d.Data):
		return nil, errors.New("sign-data url, purpose and data must be UTF-8 text")
	}

	return signaturePayload(signDataPayload{
		ChainID: d.ChainID,
		Fee:     signDocFee{Amount: []signDocCoin{}, Gas: "0"},
		Msgs: []signDataMsg{{
			Type: SignDataMsgType,
			Value: signDataValue{
				Signer:  d.Signer.String(),
				Data:    base64.StdEncoding.EncodeToString([]byte(d.Data)),
				URL:     d.URL,
				Purpose: d.Purpose,
			},
		}},
	})
}

// signDataDoc reads the values of a sign-data document; VerifySignData checks
// the rest of it against the canonical bytes.
type signDataDoc struct {
	ChainID string        `json:"chain_id"`
	Msgs    []signDataMsg `json:"msgs"`
}

// VerifySignData checks that sig, by pubKey, signs the sign-data document doc,
// and that pubKey is the key of the document's signer. Only a secp256k1 key
// signs a proof. It returns what the document says; the caller still checks the
// chain id, url, purpose and data it expects.
//
// doc may come in any JSON rendering (key order, spacing, string escapes), but
// must otherwise be the canonical document: re-encoded sorted by key and
// compacted, it must equal the bytes rebuilt from its values, and no object may
// repeat a key. Missing or extra fields, numbers for strings, a null for the
// empty fee, non-canonical base64 in the data and a duplicate key are refused,
// so no other document passes for one, whatever parser reads it.
func VerifySignData(doc []byte, pubKey secp256k1.PubKeySecp256k1, sig []byte) (SignedData, error) {
	var d signDataDoc
	if err := unmarshalStrict(doc, &d); err != nil {
		return SignedData{}, fmt.Errorf("unable to read sign-data document, %w", err)
	}
	if len(d.Msgs) != 1 {
		return SignedData{}, errors.New("not a sign-data document")
	}

	signer, err := crypto.AddressFromBech32(d.Msgs[0].Value.Signer)
	if err != nil {
		return SignedData{}, fmt.Errorf("invalid sign-data signer, %w", err)
	}
	data, err := base64.StdEncoding.DecodeString(d.Msgs[0].Value.Data)
	if err != nil {
		return SignedData{}, fmt.Errorf("invalid sign-data data, %w", err)
	}

	signed := SignedData{
		ChainID: d.ChainID,
		URL:     d.Msgs[0].Value.URL,
		Purpose: d.Msgs[0].Value.Purpose,
		Signer:  signer,
		Data:    string(data),
	}
	payload, err := SignDataPayload(signed)
	if err != nil {
		return SignedData{}, err
	}

	// doc already parsed, so sortJSON cannot fail.
	if canonical, _ := sortJSON(doc); !bytes.Equal(canonical, payload) {
		return SignedData{}, errors.New("not a sign-data document")
	}

	if pubKey.Address() != signer {
		return SignedData{}, fmt.Errorf("the key belongs to %s, not to the signer %s", pubKey.Address(), signer)
	}
	if !pubKey.VerifyBytes(payload, sig) {
		return SignedData{}, errors.New("invalid sign-data signature")
	}

	return signed, nil
}

// SignDataPubKeyType is the type URL of a sign-data proof's public key.
const SignDataPubKeyType = "/tm.PubKeySecp256k1"

// SignDataProof is a sign-data proof: the signed document, the public key and
// the signature.
type SignDataProof struct {
	Signed    json.RawMessage   `json:"signed"`
	Signature SignDataSignature `json:"signature"`
}

type SignDataSignature struct {
	PubKey    SignDataPubKey `json:"pubKey"`
	Signature string         `json:"signature"` // base64, 64 bytes R||S
}

type SignDataPubKey struct {
	TypeURL string `json:"typeUrl"`
	Value   string `json:"value"` // base64 of the compressed key itself
}

// NewSignDataProof returns the proof that sig, by pubKey, signs payload.
func NewSignDataProof(payload []byte, pubKey secp256k1.PubKeySecp256k1, sig []byte) SignDataProof {
	return SignDataProof{
		Signed: payload,
		Signature: SignDataSignature{
			PubKey: SignDataPubKey{
				TypeURL: SignDataPubKeyType,
				Value:   base64.StdEncoding.EncodeToString(pubKey[:]),
			},
			Signature: base64.StdEncoding.EncodeToString(sig),
		},
	}
}

// VerifySignDataProof checks a proof in SignDataProof's JSON with
// VerifySignData; the caller still checks the chain id, url, purpose and data it
// expects.
// Its cost grows with the proof's size, so a service bounds what it accepts
// (a proof of a short challenge is under 1 KiB).
//
// A proof stays valid forever, and the same proof can be written in many byte
// forms (the document's rendering, the envelope's base64 and key spelling). A
// service therefore refuses a replay by its challenge, never by the proof's
// bytes: it hands out a fresh, unpredictable challenge naming the session,
// gives it an expiry, and accepts it once. The url and the purpose keep a proof
// made for one service, or one use, from passing at another.
func VerifySignDataProof(proof []byte) (SignedData, error) {
	var p SignDataProof
	if err := unmarshalStrict(proof, &p); err != nil {
		return SignedData{}, fmt.Errorf("unable to read sign-data proof, %w", err)
	}

	if p.Signature.PubKey.TypeURL != SignDataPubKeyType {
		return SignedData{}, fmt.Errorf("the proof's key is a %q, not a %s", p.Signature.PubKey.TypeURL, SignDataPubKeyType)
	}

	key, err := base64.StdEncoding.DecodeString(p.Signature.PubKey.Value)
	if err != nil || len(key) != secp256k1.PubKeySecp256k1Size {
		return SignedData{}, errors.New("the proof's public key is not a base64 compressed secp256k1 key")
	}

	var pubKey secp256k1.PubKeySecp256k1
	copy(pubKey[:], key)

	sig, err := base64.StdEncoding.DecodeString(p.Signature.Signature)
	if err != nil {
		return SignedData{}, fmt.Errorf("the proof's signature is not base64, %w", err)
	}

	return VerifySignData(p.Signed, pubKey, sig)
}
