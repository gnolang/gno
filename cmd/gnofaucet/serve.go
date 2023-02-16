package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/amino"
	rpcclient "github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/std"
)

// url & struct for verify captcha
const siteVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

type config struct {
	client.BaseOptions // home, ...

	ChainID               string
	GasWanted             int64
	GasFee                string
	Memo                  string
	TestTo                string
	Send                  string
	CaptchaSecret         string
	IsBehindProxy         bool
	InsecurePasswordStdin bool
}

func newServeCmd() *commands.Command {
	cfg := &config{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "serve",
			ShortUsage: "serve [flags] <key>",
			LongHelp:   "Serves the gno.land faucet to users",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execServe(cfg, args, bufio.NewReader(os.Stdin))
		},
	)
}

func (c *config) RegisterFlags(fs *flag.FlagSet) {
	// Base config options
	fs.StringVar(
		&c.BaseOptions.Home,
		"home",
		client.DefaultBaseOptions.Home,
		"home directory",
	)

	fs.StringVar(
		&c.BaseOptions.Remote,
		"remote",
		client.DefaultBaseOptions.Remote,
		"remote node URL",
	)

	fs.BoolVar(
		&c.BaseOptions.Quiet,
		"quiet",
		client.DefaultBaseOptions.Quiet,
		"for parsing output",
	)

	// Command options
	fs.StringVar(
		&c.ChainID,
		"chain-id",
		"",
		"the ID of the chain",
	)

	fs.Int64Var(
		&c.GasWanted,
		"gas-wanted",
		50000,
		"gas requested for the tx",
	)

	fs.StringVar(
		&c.GasFee,
		"gas-fee",
		"1000000ugnot",
		"gas payment fee",
	)

	fs.StringVar(
		&c.Memo,
		"memo",
		"",
		"any descriptive text",
	)

	fs.StringVar(
		&c.TestTo,
		"test-to",
		"",
		"test address (optional)",
	)

	fs.StringVar(
		&c.Send,
		"send",
		"1000000ugnot",
		"send coins",
	)

	fs.StringVar(
		&c.CaptchaSecret,
		"captcha-secret",
		"",
		"recaptcha secret key (if empty, captcha are disabled)",
	)

	fs.BoolVar(
		&c.IsBehindProxy,
		"is-behind-proxy",
		false,
		"use X-Forwarded-For IP for throttling",
	)

	fs.BoolVar(
		&c.InsecurePasswordStdin,
		"insecure-password-stdin",
		false,
		"WARNING! take password from stdin",
	)
}

func execServe(cfg *config, args []string, input *bufio.Reader) error {
	if len(args) != 1 {
		return errors.New("invalid args")
	}

	if cfg.ChainID == "" {
		return errors.New("chain-id not specified")
	}

	if cfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}

	if cfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	remote := cfg.Remote
	if remote == "" || remote == "y" {
		return errors.New("missing remote url")
	}
	cli := rpcclient.NewHTTP(remote, "/websocket")

	// XXX XXX
	// Read supply account pubkey.
	name := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByName(name)
	if err != nil {
		return err
	}
	fromAddr := info.GetAddress()

	// query for initial number and sequence.
	path := fmt.Sprintf("auth/accounts/%s", fromAddr.String())
	data := []byte(nil)
	opts2 := rpcclient.ABCIQueryOptions{}
	qres, err := cli.ABCIQueryWithOptions(
		path, data, opts2)
	if err != nil {
		return errors.Wrap(err, "querying")
	}
	if qres.Response.Error != nil {
		fmt.Printf("Log: %s\n",
			qres.Response.Log)
		return qres.Response.Error
	}
	resdata := qres.Response.Data
	var acc gnoland.GnoAccount
	amino.MustUnmarshalJSON(resdata, &acc)
	accountNumber := acc.BaseAccount.AccountNumber
	sequence := acc.BaseAccount.Sequence

	// Get password for supply account.
	// Test by signing a dummy message;
	const dummy = "test"
	var pass string
	if cfg.Quiet {
		pass, err = commands.GetPassword("", cfg.InsecurePasswordStdin, input)
	} else {
		pass, err = commands.GetPassword("Enter password", cfg.InsecurePasswordStdin, input)
	}

	if err != nil {
		return err
	}

	_, _, err = kb.Sign(name, pass, []byte(dummy))
	if err != nil {
		return err
	}

	// Parse send amount.
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// Parse test-to address. If present, send and quit.
	if cfg.TestTo != "" {
		testToAddr, err := crypto.AddressFromBech32(cfg.TestTo)
		if err != nil {
			return err
		}
		err = sendAmountTo(cfg, cli, name, pass, testToAddr, accountNumber, sequence, send)
		return err
	}

	// Start throttled faucet.
	st := NewSubnetThrottler()
	st.Start()

	// handle route using handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := ""
		if !cfg.IsBehindProxy {
			addr := r.RemoteAddr
			host_, _, err := net.SplitHostPort(addr)
			if err != nil {
				return
			}
			host = host_
		} else if xff, found := r.Header["X-Forwarded-For"]; found && len(xff) > 0 {
			host = xff[0]
		}

		// if can't identify the IP, everyone is in the same pool.
		// if host using ipv6 loopback addr, make it ipv4
		if host == "" || host == "::1" || host == "0:0:0:0:0:0:0:1" {
			host = "127.0.0.1"
		}
		ip := net.ParseIP(host)
		if ip == nil {
			fmt.Println("no ip found")
			w.Write([]byte("no ip found"))
			return
		}

		allowed, reason := st.Request(ip)
		if !allowed {
			msg := fmt.Sprintf("abuse protection system (%s)", reason)
			fmt.Println(ip, msg)
			w.Write([]byte(msg))
			return
		}

		r.ParseForm()

		// only when command line argument 'captcha-secret' has entered > captcha are enabled.
		// veryify captcha
		if cfg.CaptchaSecret != "" {
			passedMsg := r.Form["g-recaptcha-response"]
			if passedMsg == nil {
				fmt.Println(ip, "no 'captcha' request")
				w.Write([]byte("check captcha request"))
				return
			}

			capMsg := strings.TrimSpace(passedMsg[0])

			if err := checkRecaptcha(cfg.CaptchaSecret, capMsg); err != nil {
				fmt.Printf("%s recaptcha failed; %v\n", ip, err)
				w.Write([]byte("Unauthorized"))
				return
			}
		}

		passedAddr := r.Form["toaddr"]
		if passedAddr == nil {
			fmt.Println(ip, "no address found")
			w.Write([]byte("no address found"))
			return
		}

		toAddrStr := strings.TrimSpace(passedAddr[0])

		// OK.
		toAddr, err := crypto.AddressFromBech32(toAddrStr)
		if err != nil {
			fmt.Println(ip, "invalid address format", err)
			w.Write([]byte("invalid address format"))
			return
		}
		err = sendAmountTo(cfg, cli, name, pass, toAddr, accountNumber, sequence, send)
		if err != nil {
			fmt.Println(ip, "faucet failed", err)
			w.Write([]byte("faucet failed"))
			return
		} else {
			sequence += 1
			fmt.Println(ip, "faucet success")
			w.Write([]byte("faucet success"))
		}
	})

	// listen to port
	fmt.Println("Starting server at port 5050")

	server := &http.Server{
		Addr:              ":5050",
		ReadHeaderTimeout: 60 * time.Second,
	}
	server.ListenAndServe()

	return nil
}

func sendAmountTo(
	cfg *config,
	cli rpcclient.Client,
	name,
	pass string,
	toAddr crypto.Address,
	accountNumber,
	sequence uint64,
	send std.Coins,
) error {
	// Read supply account pubkey.
	kb, err := keys.NewKeyBaseFromDir(cfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByName(name)
	if err != nil {
		return err
	}
	fromAddr := info.GetAddress()
	pub := info.GetPubKey()

	// parse gas wanted & fee.
	gaswanted := cfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := bank.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      send,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.Memo,
	}
	// fill tx signatures.
	signers := tx.GetSigners()
	if tx.Signatures == nil {
		for range signers {
			tx.Signatures = append(tx.Signatures, std.Signature{
				PubKey:    nil, // zero signature
				Signature: nil, // zero signature
			})
		}
	}
	err = tx.ValidateBasic()
	if err != nil {
		return err
	}
	// fmt.Println("will sign:", string(amino.MustMarshalJSON(tx)))

	// get sign-bytes and make signature.
	chainID := cfg.ChainID
	signbz := tx.GetSignBytes(chainID, accountNumber, sequence)
	sig, _, err := kb.Sign(name, pass, signbz)
	if err != nil {
		return err
	}

	found := false
	for i := range tx.Signatures {
		// override signature for matching slot.
		if signers[i] == fromAddr {
			found = true
			tx.Signatures[i] = std.Signature{
				PubKey:    pub,
				Signature: sig,
			}
		}
	}
	if !found {
		return errors.New("addr %v (%s) not in signer set",
			fromAddr, name)
	}
	fmt.Println("will deliver:", string(amino.MustMarshalJSON(tx)))

	// construct tx serialized bytes.
	txbz := amino.MustMarshal(tx)

	// broadcast tx bytes.
	bres, err := cli.BroadcastTxCommit(txbz)
	if err != nil {
		return errors.Wrap(err, "broadcasting bytes")
	}
	if bres.CheckTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", bres, bres.CheckTx.Log)
	} else if bres.DeliverTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", bres, bres.DeliverTx.Log)
	} else {
		fmt.Println(string(bres.DeliverTx.Data))
		fmt.Println("OK!")
		fmt.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
		fmt.Println("GAS USED:  ", bres.DeliverTx.GasUsed)
	}
	return nil
}

func checkRecaptcha(secret, response string) error {
	req, err := http.NewRequest(http.MethodPost, siteVerifyURL, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("secret", secret)
	q.Add("response", response)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req) // 200 OK
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body SiteVerifyResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return errors.New("fail, decode response")
	}

	if !body.Success {
		return errors.New("unsuccessful recaptcha verify request")
	}

	return nil
}
