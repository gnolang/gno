package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	tm2Client "github.com/gnolang/faucet/client/http"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func getAccountBalanceMiddleware(tm2Client *tm2Client.Client, maxBalance int64) func(next http.Handler) http.Handler {
	type request struct {
		To string `json:"to"`
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				var data request
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				err = json.Unmarshal(body, &data)
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				balance, err := checkAccountBalance(tm2Client, data.To)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if balance >= maxBalance {
					http.Error(w, "accounts is already topped up", http.StatusBadRequest)
					return
				}
				next.ServeHTTP(w, r)
			},
		)
	}
}

var checkAccountBalance = func(tm2Client *tm2Client.Client, walletAddress string) (int64, error) {
	address, err := crypto.AddressFromString(walletAddress)
	if err != nil {
		return 0, err
	}
	acc, err := tm2Client.GetAccount(address)
	if err != nil {
		return 0, err
	}
	return acc.GetCoins().AmountOf("ugnot"), nil
}
