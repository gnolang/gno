package std

import (
	"time"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/crypto"
)

//----------------------------------------
// SignDoc

// The standard object for all signing, including transactions and other
// documents. AccountNumber is a replay-prevention field for the whole account
// (eg. nonce) to prevent the replay of txs after an account has been deleted
// (due to zero balance). Time can also be used in the future instead of
// AccountNumber . Sequence is a replay-prevention field for each transaction
// given a nonce.
type SignDoc struct {
	ChainID       string    `json:"chain_id" yaml:"chain_id"`
	AccountNumber uint64    `json:"account_number" yaml:"account_number"`
	Time          time.Time `json:"time" yaml:"time"`
	Sequence      uint64    `json:"sequence" yaml:"sequence"`
	Fee           Fee       `json:"fee" yaml:"fee"`
	Msgs          []Msg     `json:"msgs" yaml:"msgs"`
	Memo          string    `json:"memo" yaml:"memo"`
}

// SignBytes returns the bytes to sign for a transaction.
func SignBytes(chainID string, accountNumber uint64, sequence uint64, fee Fee, msgs []Msg, memo string) []byte {
	bz, err := amino.MarshalJSON(SignDoc{
		ChainID:       chainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
		Fee:           fee,
		Msgs:          msgs,
		Memo:          memo,
	})
	if err != nil {
		panic(err)
	}
	return MustSortJSON(bz)
}

//----------------------------------------
// Signature

// Signature represents a signature of a SignDoc.
type Signature struct {
	PubKey    crypto.PubKey `json:"pub_key" yaml:"pub_key"` // optional
	Signature []byte        `json:"signature" yaml:"signature"`
}
