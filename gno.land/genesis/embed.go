package genesis

import _ "embed"

//go:embed genesis_balances.txt
var DefaultGenesisBalances []byte

//go:embed genesis_txs.jsonl
var DefaultGenesisTxs []byte
