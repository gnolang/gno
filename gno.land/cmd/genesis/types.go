package main

import "github.com/gnolang/gno/tm2/pkg/std"

type txMap map[string]std.Tx

func (m txMap) leftMerge(b txMap) {
	for key, bVal := range b {
		if _, present := (m)[key]; !present {
			(m)[key] = bVal
		}
	}
}

func (m txMap) toList() []std.Tx {
	txs := make([]std.Tx, 0, len(m))

	for _, tx := range m {
		txs = append(txs, tx)
	}

	return txs
}
