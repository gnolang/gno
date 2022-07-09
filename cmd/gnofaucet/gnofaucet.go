package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/amino"
	rpcclient "github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/command"
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

type (
	AppItem = command.AppItem
	AppList = command.AppList
)

var mainApps AppList = []AppItem{
	{serveApp, "serve", "serve faucet", DefaultServeOptions},
}

func runMain(cmd *command.Command, exec string, args []string) error {
	// show help message.
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		cmd.Println("available subcommands:")
		for _, appItem := range mainApps {
			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)
		}
		return nil
	}

	// switch on first argument.
	for _, appItem := range mainApps {
		if appItem.Name == args[0] {
			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			return err // done
		}
	}

	// unknown app command!
	return errors.New("unknown command " + args[0])
}

func main() {
	cmd := command.NewStdCommand()
	exec := os.Args[0]
	args := os.Args[1:]
	err := runMain(cmd, exec, args)
	if err != nil {
		cmd.ErrPrintfln("%s", err.Error())
		cmd.ErrPrintfln("%#v", err)
		return // exit
	}
}

//----------------------------------------
// serveApp

type serveOptions struct {
	client.BaseOptions        // home, ...
	ChainID            string `flag:"chain-id" help:"chain id"`
	GasWanted          int64  `flag:"gas-wanted" help:"gas requested for tx"`
	GasFee             string `flag:"gas-fee" help:"gas payment fee"`
	Memo               string `flag:"memo" help:"any descriptive text"`
	TestTo             string `flag:"test-to" help:"test addr (optional)"`
	Send               string `flag:"send" help:"send coins"`
	CaptchaSecret      string `flag:"captcha-secret" help:"recaptcha secret key (if empty, captcha are disabled)"`
	IsBehindProxy      bool   `flag:"is-behind-proxy" help:"use X-Forwarded-For IP for throttling."`
}

var DefaultServeOptions = serveOptions{
	ChainID:       "", // must override
	GasWanted:     50000,
	GasFee:        "1ugnot",
	Memo:          "",
	TestTo:        "",
	Send:          "1ugnot",
	CaptchaSecret: "",
	IsBehindProxy: false,
}

func serveApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(serveOptions)
	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: serve <keyname>")
		return errors.New("invalid args")
	}
	if opts.ChainID == "" {
		return errors.New("chain-id not specified")
	}
	if opts.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if opts.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	remote := opts.Remote
	if remote == "" || remote == "y" {
		return errors.New("missing remote url")
	}
	cli := rpcclient.NewHTTP(remote, "/websocket")

	// XXX XXX
	// Read supply account pubkey.
	name := args[0]
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByName(name)
	if err != nil {
		return err
	}
	fromAddr := info.GetAddress()
	// pub := info.GetPubKey()

	// query for initial number and sequence.
	path := fmt.Sprintf("auth/accounts/%s", fromAddr.String())
	data := []byte(nil)
	opts2 := rpcclient.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
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
	if opts.Quiet {
		pass, err = cmd.GetPassword("")
	} else {
		pass, err = cmd.GetPassword("Enter password.")
	}
	if err != nil {
		return err
	}
	_, _, err = kb.Sign(name, pass, []byte(dummy))
	if err != nil {
		return err
	}

	// Parse send amount.
	send, err := std.ParseCoins(opts.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// Parse test-to address. If present, send and quit.
	if opts.TestTo != "" {
		testToAddr, err := crypto.AddressFromBech32(opts.TestTo)
		if err != nil {
			return err
		}
		err = sendAmountTo(cmd, cli, name, pass, testToAddr, accountNumber, sequence, send, opts)
		return err
	}

	// Start throttled faucet.
	st := NewSubnetThrottler()
	st.Start()

	// handle route using handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := ""
		if !opts.IsBehindProxy {
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

		if !st.Request(ip) {
			fmt.Println("wrong ip format")
			w.Write([]byte("wrong ip format"))
			return
		}

		r.ParseForm()

		// only when command line argument 'captcha-secret' has entered > captcha are enabled.
		// veryify captcha
		if opts.CaptchaSecret != "" {
			passedMsg := r.Form["g-recaptcha-response"]
			if passedMsg == nil {
				fmt.Println("no 'captcha' request")
				w.Write([]byte("check captcha request"))
				return
			}

			capMsg := strings.TrimSpace(passedMsg[0])

			if err := checkRecaptcha(opts.CaptchaSecret, capMsg); err != nil {
				fmt.Printf("recaptcha failed; %v\n", err)
				w.Write([]byte("Unauthorized"))
				return
			}

		}

		passedAddr := r.Form["toaddr"]
		if passedAddr == nil {
			fmt.Println("input your address")
			w.Write([]byte("no address found"))
			return
		}

		toAddrStr := strings.TrimSpace(passedAddr[0])

		// OK.
		toAddr, err := crypto.AddressFromBech32(toAddrStr)
		if err != nil {
			fmt.Println("error:", err)
			w.Write([]byte("invalid address format"))
			return
		}
		err = sendAmountTo(cmd, cli, name, pass, toAddr, accountNumber, sequence, send, opts)

		if err != nil {
			fmt.Println("error:", err)
			w.Write([]byte("faucet fail"))
			return
		} else {
			sequence += 1
			w.Write([]byte("faucet success"))
		}
	})

	// listen to port
	http.ListenAndServe(":5050", nil)

	return nil
}

func sendAmountTo(cmd *command.Command, cli rpcclient.Client, name, pass string, toAddr crypto.Address, accountNumber, sequence uint64, send std.Coins, opts serveOptions) error {
	// Read supply account pubkey.
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
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
	gaswanted := opts.GasWanted
	gasfee, err := std.ParseCoin(opts.GasFee)
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
		Memo:       opts.Memo,
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
	chainID := opts.ChainID
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
		cmd.Println(string(bres.DeliverTx.Data))
		cmd.Println("OK!")
		cmd.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
		cmd.Println("GAS USED:  ", bres.DeliverTx.GasUsed)
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
