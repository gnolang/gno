package std

import (
	"time"
)

// The standard object for all signing,
// including transactions and other documents.
type SignDoc struct {
	ChainID  string    `json:"chain_id"`
	Time     time.Time `json:"time"`
	Sequence uint64    `json:"sequence"`
	Fee      string    `json:"fee"`
	Memo     string    `json:"memo"`
	Msg      Msg       `json:"msg"`
}
