package std

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
)

var maxGasWanted = int64((1 << 60) - 1) // something smaller than math.MaxInt64

// Tx is a standard way to wrap a Msg with Fee and Signatures.
// NOTE: the first signature is the fee payer (Signatures must not be nil).
type Tx struct {
	Msgs       []Msg       `json:"msg" yaml:"msg"`
	Fee        Fee         `json:"fee" yaml:"fee"`
	Signatures []Signature `json:"signatures" yaml:"signatures"`
	Memo       string      `json:"memo" yaml:"memo"`
}

func NewTx(msgs []Msg, fee Fee, sigs []Signature, memo string) Tx {
	return Tx{
		Msgs:       msgs,
		Fee:        fee,
		Signatures: sigs,
		Memo:       memo,
	}
}

// GetMsgs returns the all the transaction's messages.
func (tx Tx) GetMsgs() []Msg { return tx.Msgs }

// ValidateBasic does a simple and lightweight validation check that doesn't
// require access to any other information.
func (tx Tx) ValidateBasic() error {
	stdSigs := tx.GetSignatures()

	if tx.Fee.GasWanted > maxGasWanted {
		return ErrGasOverflow(fmt.Sprintf("invalid gas supplied; %d > %d", tx.Fee.GasWanted, maxGasWanted))
	}
	if !tx.Fee.GasFee.IsValid() {
		return ErrInsufficientFee(fmt.Sprintf("invalid fee %s amount provided", tx.Fee.GasFee))
	}
	if len(stdSigs) == 0 {
		return ErrNoSignatures("no signers")
	}
	if len(stdSigs) != len(tx.GetSigners()) {
		return ErrUnauthorized("wrong number of signers")
	}

	return nil
}

// CountSubKeys counts the total number of keys for a multi-sig public key.
func CountSubKeys(pub crypto.PubKey) int {
	v, ok := pub.(multisig.PubKeyMultisigThreshold)
	if !ok {
		return 1
	}

	numKeys := 0
	for _, subkey := range v.PubKeys {
		numKeys += CountSubKeys(subkey)
	}

	return numKeys
}

// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a deterministic order.
// They are accumulated from the GetSigners method for each Msg
// in the order they appear in tx.GetMsgs().
// Duplicate addresses will be omitted.
func (tx Tx) GetSigners() []crypto.Address {
	seen := map[string]bool{}
	var signers []crypto.Address
	for _, msg := range tx.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}
	return signers
}

// GetMemo returns the memo
func (tx Tx) GetMemo() string { return tx.Memo }

// GetSignatures returns the signature of signers who signed the Msg.
// GetSignatures returns the signature of signers who signed the Msg.
// CONTRACT: Length returned is same as length of
// pubkeys returned from MsgKeySigners, and the order
// matches.
// CONTRACT: If the signature is missing (ie the Msg is
// invalid), then the corresponding signature is
// .Empty().
func (tx Tx) GetSignatures() []Signature { return tx.Signatures }

func (tx Tx) GetSignBytes(chainID string, accountNumber uint64, sequence uint64) ([]byte, error) {
	return GetSignaturePayload(SignDoc{
		ChainID:       chainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
		Fee:           tx.Fee,
		Msgs:          tx.Msgs,
		Memo:          tx.Memo,
	})
}

// __________________________________________________________

// Fee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type Fee struct {
	GasWanted int64 `json:"gas_wanted" yaml:"gas_wanted"`
	GasFee    Coin  `json:"gas_fee" yaml:"gas_fee"`
}

// NewFee returns a new instance of Fee
func NewFee(gasWanted int64, gasFee Coin) Fee {
	return Fee{
		GasWanted: gasWanted,
		GasFee:    gasFee,
	}
}

// Bytes for signing later
func (fee Fee) Bytes() []byte {
	bz, err := amino.MarshalJSON(fee) // TODO
	if err != nil {
		panic(err)
	}
	return bz
}
