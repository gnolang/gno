package std

import (
	"fmt"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// SignDoc is the standard object for transactions.
// AccountNumber is a replay-prevention field for the whole account
// (eg. nonce) to prevent the replay of txs after an account has been deleted
// (due to zero balance). Sequence is a replay-prevention field for each transaction
// given a nonce
type SignDoc struct {
	ChainID       string `json:"chain_id" yaml:"chain_id"`
	AccountNumber uint64 `json:"account_number" yaml:"account_number"`
	Sequence      uint64 `json:"sequence" yaml:"sequence"`
	Fee           Fee    `json:"fee" yaml:"fee"`
	Msgs          []Msg  `json:"msgs" yaml:"msgs"`
	Memo          string `json:"memo" yaml:"memo"`
}

// signDocFee is the fee as it appears in the signature payload.
//
// It is deliberately NOT std.Fee. The Ledger Cosmos app -- which gno has no
// app of its own and therefore borrows -- validates the amino sign doc against
// a fixed allowlist of keys, and refuses to sign anything containing a key it
// does not know. Inside "fee" it permits only amount, gas, granter and payer
// (ledger-cosmos, app/src/tx_validate.c).
//
// std.Fee renders as {"gas_wanted":...,"gas_fee":"1000000ugnot"}, so every gno
// transaction -- a send as much as an addpkg -- is refused with "Unexpected
// field" before the device will display anything. Three things differ from what
// the app expects: the key names, the cardinality (one Coin against an array),
// and the numeric type (amino JSON renders these as strings).
type signDocFee struct {
	Amount []signDocCoin `json:"amount"`
	Gas    string        `json:"gas"`
}

// signDocCoin spells out a Coin, because amino renders std.Coin as the single
// string "1000000ugnot" and the allowlist wants an object.
type signDocCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// signDocPayload mirrors SignDoc with the fee in the form above. Every other
// field already matches what the app allows, so only the fee is restated.
// TestSignDocPayloadMirrorsSignDoc keeps the two field sets in step.
type signDocPayload struct {
	ChainID       string     `json:"chain_id"`
	AccountNumber uint64     `json:"account_number"`
	Sequence      uint64     `json:"sequence"`
	Fee           signDocFee `json:"fee"`
	Msgs          []Msg      `json:"msgs"`
	Memo          string     `json:"memo"`
}

// GetSignaturePayloadLegacy returns the signature payload with the fee in its
// gas_wanted/gas_fee rendering: amino JSON over SignDoc itself, sorted by key.
//
// Verification accepts this rendering alongside the one GetSignaturePayload
// produces, so that clients building the payload themselves -- wallets, the
// genesis tooling, anything holding a key -- and signatures already written into
// genesis files or chain history keep verifying. See VerifySignaturePayload.
func GetSignaturePayloadLegacy(s SignDoc) ([]byte, error) {
	return signaturePayload(s)
}

// feeAmount renders the fee coin as Cosmos renders a coin list.
//
// A ZERO FEE IS AN EMPTY LIST, not a list holding an empty coin. Cosmos's Coins
// carries no zero entries, and the device DISPLAYS each coin it is given, so
// {"denom":"","amount":"0"} is both wrong and likely to be rejected -- which
// would reintroduce, for zero-fee transactions, exactly the failure this file
// exists to fix. Zero fees are a supported case, not a hypothetical: the ante
// handler branches on GasFee.IsZero(), and every genesis transaction is signed
// with GetSignBytes(chainID, 0, 0).
//
// The slice MUST be non-nil. Amino renders a nil slice as null and an empty one
// as [], and those are different signed bytes; TestSignaturePayloadZeroFeeIsEmptyList
// pins the latter.
func feeAmount(c Coin) []signDocCoin {
	if c.IsZero() {
		return []signDocCoin{}
	}
	return []signDocCoin{{
		Denom:  c.Denom,
		Amount: strconv.FormatInt(c.Amount, 10),
	}}
}

// GetSignaturePayload returns the signature payload for the SignDoc: amino JSON
// of signDocPayload, sorted by key. Every field of s passes through as is except
// the fee, which is restated as the coin list the Ledger Cosmos app parses (see
// signDocFee). The formula for signing is therefore
// sign(sortJSON(aminoJSON(signDocPayload(SignDoc)))).
func GetSignaturePayload(s SignDoc) ([]byte, error) {
	return signaturePayload(signDocPayload{
		ChainID:       s.ChainID,
		AccountNumber: s.AccountNumber,
		Sequence:      s.Sequence,
		Fee: signDocFee{
			Amount: feeAmount(s.Fee.GasFee),
			Gas:    strconv.FormatInt(s.Fee.GasWanted, 10),
		},
		Msgs: s.Msgs,
		Memo: s.Memo,
	})
}

// signaturePayload renders v as amino JSON with its keys sorted, the form every
// signature payload takes: sign(sortJSON(aminoJSON(v))).
func signaturePayload(v any) ([]byte, error) {
	data, err := amino.MarshalJSON(v)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal sign doc, %w", err)
	}

	sortedData, err := sortJSON(data)
	if err != nil {
		return nil, fmt.Errorf("unable to sort payload JSON, %w", err)
	}

	return sortedData, nil
}

// VerifySignaturePayload reports whether sig is a valid signature by pubKey over
// s in either of the two accepted renderings: the amount/gas fee shape the
// Ledger Cosmos app parses (GetSignaturePayload), and the gas_wanted/gas_fee
// shape (GetSignaturePayloadLegacy). An error means no payload could be built
// to check against, which is a malformed sign doc rather than a bad signature.
//
// WHY BOTH ARE ACCEPTED. Clients build the signature payload themselves, so the
// rendering cannot change on one side only. A node that took only the amount/gas
// bytes would reject every wallet still producing the other shape, and every
// signature already made over it -- including the ones sitting in written
// genesis files, which cannot be re-signed. A node that took only the
// gas_wanted/gas_fee bytes leaves every Ledger unable to sign at all.
//
// WHY ACCEPTING BOTH IS SAFE, and not merely convenient. The two renderings
// cannot be confused for one another: one fee object carries gas_wanted and
// gas_fee, the other amount and gas, and those key sets have no member in
// common. Two JSON objects that parse to different key sets are not the same
// bytes, so no legacy rendering of one transaction can equal the current
// rendering of a DIFFERENT one. A signature therefore still authorises exactly
// one transaction; what widens is the set of acceptable proofs of it, not the set
// of transactions behind a proof. TestSignaturePayloadEncodingsAreDisjoint pins
// this, and it is the property to re-check before either rendering is touched.
//
// The legacy rendering is computed only when the current one does not verify, so
// the ordinary path pays for one encoding and one curve operation. A signature
// that matches neither pays for two of each, and nothing meters that: the cost of
// rejecting an invalid signature is borne by the node, not the sender.
func VerifySignaturePayload(pubKey crypto.PubKey, s SignDoc, sig []byte) (bool, error) {
	payload, err := GetSignaturePayload(s)
	if err != nil {
		return false, err
	}
	if pubKey.VerifyBytes(payload, sig) {
		return true, nil
	}

	legacy, err := GetSignaturePayloadLegacy(s)
	if err != nil {
		return false, err
	}

	return pubKey.VerifyBytes(legacy, sig), nil
}

// Signature represents a wrapped signature of a transaction
type Signature struct {
	PubKey    crypto.PubKey `json:"pub_key" yaml:"pub_key"` // optional
	Signature []byte        `json:"signature" yaml:"signature"`
	// SessionAddr identifies a session account for delegated signing.
	// Zero-value means a master-key signature. When non-zero, the AnteHandler
	// loads the session account at /a/<signer>/s/<SessionAddr> for verification.
	// Amino binary and JSON both skip the field when the address is zero
	// ([20]byte{}), so master-signed txs pay no wire-size overhead.
	SessionAddr crypto.Address `json:"session_addr,omitempty" yaml:"session_addr,omitempty"`
}
