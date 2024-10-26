package gnoland

import (
	"errors"

	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrBalanceEmptyAddress = errors.New("balance address is empty")
	ErrBalanceEmptyAmount  = errors.New("balance amount is empty")
)

type GnoAccount struct {
	std.BaseAccount
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

// GnoGenesis defines the gno genesis API,
// adopted by differing genesis state implementations
type GnoGenesis interface {
	// GenesisBalances returns the genesis balances associated
	// with the Gno genesis state
	GenesisBalances() []Balance

	// GenesisTxs returns the genesis transactions associated
	// with the Gno genesis state
	GenesisTxs() []GenesisTx
}

type GenesisTx interface {
	// Tx returns the standard TM2 transaction
	Tx() std.Tx

	// Metadata returns the metadata tied
	// to the tx, if any
	Metadata() *GnoTxMetadata
}

type GnoGenesisState struct {
	Balances []Balance `json:"balances"`
	Txs      []std.Tx  `json:"txs"`
}

func (g GnoGenesisState) GenesisBalances() []Balance {
	return g.Balances
}

func (g GnoGenesisState) GenesisTxs() []GenesisTx {
	genesisTxs := make([]GenesisTx, len(g.Txs))

	for i, tx := range g.Txs {
		genesisTxs[i] = gnoGenesisTx{
			tx: tx,
		}
	}

	return genesisTxs
}

type gnoGenesisTx struct {
	tx std.Tx
}

func (g gnoGenesisTx) Tx() std.Tx {
	return g.tx
}

func (g gnoGenesisTx) Metadata() *GnoTxMetadata {
	return nil
}

type MetadataGenesisState struct {
	Balances []Balance    `json:"balances"`
	Txs      []MetadataTx `json:"txs"`
}

type MetadataTx struct {
	GenesisTx  std.Tx        `json:"tx"`
	TxMetadata GnoTxMetadata `json:"metadata"`
}

func (m MetadataTx) Tx() std.Tx {
	return m.GenesisTx
}

func (m MetadataTx) Metadata() *GnoTxMetadata {
	return &m.TxMetadata
}

type GnoTxMetadata struct {
	Timestamp int64 `json:"timestamp"`
}

func (m MetadataGenesisState) GenesisBalances() []Balance {
	return m.Balances
}

func (m MetadataGenesisState) GenesisTxs() []GenesisTx {
	genesisTxs := make([]GenesisTx, len(m.Txs))

	for i, tx := range m.Txs {
		genesisTxs[i] = tx
	}

	return genesisTxs
}
