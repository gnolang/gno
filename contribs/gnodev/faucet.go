package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"mime"
	"net/http"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// faucetCooldown is how long an address waits between two claims. The GNOT on
// a staging chain is worthless and the deployer's mnemonic is public, so this
// is not a security boundary: it only stops one script from emptying the
// faucet for everyone else.
const faucetCooldown = time.Minute

// faucetSimulateGas only bounds the simulation that measures a claim's real
// gas. A send measures ~1.2M; the tx that is broadcast asks for 1.5x what the
// simulation reports.
const faucetSimulateGas = 10_000_000

// faucetSender sends the faucet amount to an address and returns the tx hash.
type faucetSender func(ctx context.Context, to crypto.Address) (string, error)

// faucet serves GET /faucet (a form) and POST /faucet (a claim, as a form
// field or a JSON body {"address": "g1..."}). Claims are serialized: they
// all spend from one account, and two in flight would race on its sequence.
type faucet struct {
	logger  *slog.Logger
	amount  std.Coins
	chainID string
	send    faucetSender
	now     func() time.Time

	mu   sync.Mutex
	last map[crypto.Address]time.Time
}

func newFaucet(logger *slog.Logger, chainID string, amount std.Coins, send faucetSender) *faucet {
	return &faucet{
		logger:  logger,
		amount:  amount,
		chainID: chainID,
		send:    send,
		now:     time.Now,
		last:    map[crypto.Address]time.Time{},
	}
}

type faucetResult struct {
	Address string `json:"address,omitempty"`
	Amount  string `json:"amount,omitempty"`
	TxHash  string `json:"tx_hash,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (f *faucet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		f.render(w, r, http.StatusOK, nil)
	case http.MethodPost:
		f.claim(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (f *faucet) claim(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)

	var raw string
	if isJSON(r) {
		var req struct {
			Address string `json:"address"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			f.render(w, r, http.StatusBadRequest, &faucetResult{Error: "invalid JSON body"})
			return
		}
		raw = req.Address
	} else {
		if err := r.ParseForm(); err != nil {
			f.render(w, r, http.StatusBadRequest, &faucetResult{Error: "invalid form"})
			return
		}
		raw = r.PostForm.Get("address")
	}

	to, err := crypto.AddressFromBech32(raw)
	if err != nil {
		f.render(w, r, http.StatusBadRequest, &faucetResult{Address: raw, Error: "not a valid g1 address"})
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if wait := faucetCooldown - f.now().Sub(f.last[to]); wait > 0 {
		f.render(w, r, http.StatusTooManyRequests, &faucetResult{
			Address: to.String(),
			Error:   fmt.Sprintf("already funded, try again in %s", wait.Round(time.Second)),
		})
		return
	}

	hash, err := f.send(r.Context(), to)
	if err != nil {
		f.logger.Error("faucet send failed", "to", to.String(), "err", err)
		f.render(w, r, http.StatusInternalServerError, &faucetResult{Address: to.String(), Error: "send failed"})
		return
	}

	f.last[to] = f.now()
	f.logger.Info("faucet sent", "to", to.String(), "amount", f.amount.String(), "hash", hash)
	f.render(w, r, http.StatusOK, &faucetResult{Address: to.String(), Amount: f.amount.String(), TxHash: hash})
}

func (f *faucet) render(w http.ResponseWriter, r *http.Request, status int, res *faucetResult) {
	if isJSON(r) || wantsJSON(r) {
		if res == nil {
			res = &faucetResult{Amount: f.amount.String()}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	faucetPage.Execute(w, struct {
		ChainID string
		Amount  string
		Result  *faucetResult
	}{f.chainID, f.amount.String(), res})
}

func isJSON(r *http.Request) bool {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mt == "application/json"
}

func wantsJSON(r *http.Request) bool {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	return err == nil && mt == "application/json"
}

var faucetPage = template.Must(template.New("faucet").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>faucet · {{.ChainID}}</title>
<style>body{font:16px/1.5 system-ui,sans-serif;max-width:36rem;margin:3rem auto;padding:0 1rem}
input,button{font:inherit;padding:.5rem}input{width:100%;box-sizing:border-box}
.ok{color:#176f2c}.err{color:#b3261e}code{word-break:break-all}</style></head>
<body><h1>faucet</h1>
<p>Sends <strong>{{.Amount}}</strong> on <code>{{.ChainID}}</code>, once a minute per address.
This GNOT is worthless: it exists on this preview chain only.</p>
<form method="post" action="/faucet">
<p><input name="address" placeholder="g1..." autocomplete="off" required{{with .Result}} value="{{.Address}}"{{end}}></p>
<p><button type="submit">send</button></p></form>
{{with .Result}}{{if .Error}}<p class="err">{{.Error}}</p>{{else if .TxHash}}<p class="ok">sent {{.Amount}} to <code>{{.Address}}</code><br>tx <code>{{.TxHash}}</code></p>{{end}}{{end}}
<p><a href="/">back</a></p></body></html>
`))

// newFaucetSender spends from the default deployer, which every gnodev
// funds at genesis. client is called on every send, since a reload replaces the
// node's client.
//
// Gas is not guessed: the tx is simulated first, and the fee is sized from
// the chain's own gas price.
func newFaucetSender(chainID string, amount std.Coins, client func() rpcclient.Client) (faucetSender, error) {
	signer, err := gnoclient.SignerFromBip39(DefaultDeployerSeed, chainID, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("faucet signer: %w", err)
	}
	info, err := signer.Info()
	if err != nil {
		return nil, fmt.Errorf("faucet signer: %w", err)
	}
	from := info.GetAddress()

	return func(_ context.Context, to crypto.Address) (string, error) {
		c := &gnoclient.Client{Signer: signer, RPCClient: client()}
		msg := bank.MsgSend{FromAddress: from, ToAddress: to, Amount: amount}

		acc, _, err := c.QueryAccount(from)
		if err != nil {
			return "", fmt.Errorf("query faucet account: %w", err)
		}

		// Simulate under a ceiling, priced like a real tx so the ante
		// handler accepts it. The node needs a signed tx to know the
		// signer's pubkey, and does not charge for a simulation.
		ceilingFee, err := faucetFee(c, faucetSimulateGas)
		if err != nil {
			return "", err
		}
		tx, err := gnoclient.NewSendTx(gnoclient.BaseTxCfg{GasFee: ceilingFee.String(), GasWanted: faucetSimulateGas}, msg)
		if err != nil {
			return "", err
		}
		signed, err := c.SignTx(*tx, acc.AccountNumber, acc.Sequence)
		if err != nil {
			return "", err
		}
		used, err := c.EstimateGas(signed)
		if err != nil {
			return "", fmt.Errorf("simulate: %w", err)
		}

		wanted := used * 3 / 2
		fee, err := faucetFee(c, wanted)
		if err != nil {
			return "", err
		}

		tx.Fee = std.NewFee(wanted, fee)
		signed, err = c.SignTx(*tx, acc.AccountNumber, acc.Sequence)
		if err != nil {
			return "", err
		}
		res, err := c.BroadcastTxCommit(signed)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%X", res.Hash), nil
	}, nil
}

// faucetFee prices wanted gas at the chain's last gas price, doubled so a
// price that rises before the block does not reject the tx.
func faucetFee(c *gnoclient.Client, wanted int64) (std.Coin, error) {
	res, err := c.Query(gnoclient.QueryCfg{Path: "auth/gasprice"})
	if err != nil {
		return std.Coin{}, fmt.Errorf("query gas price: %w", err)
	}
	var gp std.GasPrice
	if err := amino.UnmarshalJSON(res.Response.Data, &gp); err != nil {
		return std.Coin{}, fmt.Errorf("decode gas price: %w", err)
	}
	if gp.Gas <= 0 || gp.Price.Denom == "" {
		return std.Coin{}, errors.New("chain reports no gas price")
	}
	amount := (wanted*gp.Price.Amount + gp.Gas - 1) / gp.Gas * 2
	return std.NewCoin(gp.Price.Denom, max(amount, 1)), nil
}
