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
// for off-chain data. Nothing registers it with amino, so no node can decode a
// transaction carrying it: a signature over the document proves control of an
// address and can never be broadcast as a transaction.
const SignDataMsgType = "sign/MsgSignData"

// A sign-data document is a transaction sign document with every field fixed
// but the chain id, the signer and the data, as Adena's SignArbitrary builds it.
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

type signDataValue struct {
	Signer string `json:"signer"`
	Data   string `json:"data"` // base64 of the UTF-8 text
}

// SignedData is what a verified sign-data document says.
type SignedData struct {
	ChainID string
	Signer  crypto.Address
	Data    string
}

// SignDataPayload returns the bytes a key signs to prove, on chainID, that it
// controls signer, with data (a text the user reads before signing, typically a
// verifier's challenge). The bytes are those of Adena's SignArbitrary: the sign
// document sorted by key, its fee an empty coin list (see signDocFee).
func SignDataPayload(chainID string, signer crypto.Address, data string) ([]byte, error) {
	if chainID == "" {
		return nil, errors.New("a sign-data document needs a chain id")
	}
	if signer.IsZero() {
		return nil, errors.New("a sign-data document needs a signer")
	}
	if data == "" {
		return nil, errors.New("a sign-data document needs data")
	}
	if !utf8.ValidString(data) {
		return nil, errors.New("sign-data data must be UTF-8 text")
	}

	return signaturePayload(signDataPayload{
		ChainID: chainID,
		Fee:     signDocFee{Amount: []signDocCoin{}, Gas: "0"},
		Msgs: []signDataMsg{{
			Type: SignDataMsgType,
			Value: signDataValue{
				Signer: signer.String(),
				Data:   base64.StdEncoding.EncodeToString([]byte(data)),
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
// and that pubKey is the key of the document's signer. It returns what the
// document says; the caller still checks the chain id and the data it expects.
//
// doc may come in any key order or spacing, but must otherwise be the canonical
// document: sorted by key and compacted, it must equal the bytes rebuilt from
// its values, and no object may repeat a key. Missing or extra fields, numbers
// for strings, a null for the empty fee, non-canonical base64 and a duplicate
// key are refused, so no other document passes for one, whatever parser reads
// it.
func VerifySignData(doc []byte, pubKey crypto.PubKey, sig []byte) (SignedData, error) {
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

	payload, err := SignDataPayload(d.ChainID, signer, string(data))
	if err != nil {
		return SignedData{}, err
	}

	// doc already parsed, so sortJSON cannot fail.
	if canonical, _ := sortJSON(doc); !bytes.Equal(canonical, payload) {
		return SignedData{}, errors.New("not a sign-data document")
	}

	if pubKey == nil {
		return SignedData{}, errors.New("no public key to verify the sign-data document with")
	}
	if pubKey.Address() != signer {
		return SignedData{}, fmt.Errorf("the key belongs to %s, not to the signer %s", pubKey.Address(), signer)
	}
	if !pubKey.VerifyBytes(payload, sig) {
		return SignedData{}, errors.New("invalid sign-data signature")
	}

	return SignedData{ChainID: d.ChainID, Signer: signer, Data: string(data)}, nil
}

// SignDataPubKeyType is the type URL of a sign-data proof's public key.
const SignDataPubKeyType = "/tm.PubKeySecp256k1"

// SignDataProof is a sign-data proof as Adena's SignArbitrary answers it: the
// signed document, the public key and the signature.
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
// VerifySignData; the caller still checks the chain id and the data it expects.
// Its cost grows with the proof's size, so a service bounds what it accepts
// (a proof of a short challenge is under 1 KiB).
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
