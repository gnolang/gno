package main

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type callCfg struct {
	home string

	gasWanted int64
	gasFee    string
	memo      string
	kb        keys.Keybase
	remote    string

	broadcast bool
	chainID   string

	send     string
	pkgPath  string
	funcName string
	args     commands.StringArr
}

func execCall(nameOrBech32 string, pass string, cfg callCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	if cfg.pkgPath == "" {
		return nil, errors.New("pkgpath not specified")
	}
	if cfg.funcName == "" {
		return nil, errors.New("func not specified")
	}

	if cfg.gasWanted == 0 {
		return nil, errors.New("gas-wanted not specified")
	}
	if cfg.gasFee == "" {
		return nil, errors.New("gas-fee not specified")
	}

	// read statement.
	fnc := cfg.funcName

	// read account pubkey.
	info, err := cfg.kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return nil, err
	}
	caller := info.GetAddress()
	// info.GetPubKey()

	// Parse send amount.
	send, err := std.ParseCoins(cfg.send)
	if err != nil {
		return nil, fmt.Errorf("parsing send coins: %w", err)
	}

	// parse gas wanted & fee.
	gaswanted := cfg.gasWanted
	gasfee, err := std.ParseCoin(cfg.gasFee)
	if err != nil {
		return nil, fmt.Errorf("parsing gas fee coin: %w", err)
	}

	// construct msg & tx and marshal.
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    send,
		PkgPath: cfg.pkgPath,
		Func:    fnc,
		Args:    cfg.args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.memo,
	}

	res, err := signAndBroadcast(cfg.remote, nameOrBech32, cfg.chainID, pass, cfg.args, tx)
	if err != nil {
		return nil, fmt.Errorf("unable to sign and broadcast: %w", err)
	}

	return res, nil
}

// func loadKeyBaseName(nameOrBech32 string) error {
// 	home := gnoenv.HomeDir()
// 	kb, err := keys.NewKeyBaseFromDir(home)
// 	if err != nil {
// 		return err
// 	}
// 	info, err := kb.GetByNameOrAddress(nameOrBech32)
// 	if err != nil {
// 		return err
// 	}

// 	caller := info.GetAddress()
// 	return nil
// }

type queryCfg struct {
	data   string
	height int64
	prove  bool

	// internal
	path string
}

func queryHandler(remote string, cfg *queryCfg) (*ctypes.ResultABCIQuery, error) {
	if remote == "" || remote == "y" {
		return nil, errors.New("missing remote url")
	}

	data := []byte(cfg.data)
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		cfg.path, data, opts2)
	if err != nil {
		return nil, errors.Wrap(err, "querying")
	}

	return qres, nil
}

type signCfg struct {
	txPath        string
	chainID       string
	accountNumber uint64
	sequence      uint64
	showSignBytes bool
	home          string

	// internal flags, when called programmatically
	nameOrBech32 string
	txJSON       []byte
	pass         string
}

func signAndBroadcast(remote, nameOrBech32, chainid, pass string, args []string, tx std.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	home := gnoenv.HomeDir()

	// query account
	kb, err := keys.NewKeyBaseFromDir(home)
	if err != nil {
		return nil, err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return nil, err
	}
	accountAddr := info.GetAddress()

	qopts := &queryCfg{
		path: fmt.Sprintf("auth/accounts/%s", accountAddr),
	}
	qres, err := queryHandler(remote, qopts)
	if err != nil {
		return nil, fmt.Errorf("query account: %w", err)
	}
	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return nil, fmt.Errorf("unmarshall query response: %w", err)
	}

	// sign tx
	accountNumber := qret.BaseAccount.AccountNumber
	sequence := qret.BaseAccount.Sequence
	sopts := &signCfg{
		home:          home,
		sequence:      sequence,
		accountNumber: accountNumber,
		chainID:       chainid,
		nameOrBech32:  nameOrBech32,
		txJSON:        amino.MustMarshalJSON(tx),
	}

	signedTx, err := SignHandler(sopts)
	if err != nil {
		return nil, fmt.Errorf("sign tx: %w", err)
	}

	// broadcast signed tx
	bopts := &broadcastCfg{
		tx: signedTx,
	}

	bres, err := broadcastHandler(remote, bopts)
	if err != nil {
		return nil, fmt.Errorf("broadcast: %w", err)
	}

	if bres.CheckTx.IsErr() {
		req := fmt.Sprintf("remote:%s name:%s chainid:%s pass:%s", remote, nameOrBech32, chainid, pass)
		return nil, fmt.Errorf("check trasaction: %w \nreq: %s\nLog:%s", bres.CheckTx.Error, req, bres.CheckTx.Log)
	}

	if bres.DeliverTx.IsErr() {
		return nil, fmt.Errorf("check trasaction: %w \nLog:%s", bres.DeliverTx.Error, bres.DeliverTx.Log)
	}

	return bres, nil
}

func SignHandler(cfg *signCfg) (*std.Tx, error) {
	var err error
	var tx std.Tx

	if cfg.txJSON == nil {
		return nil, errors.New("invalid tx content")
	}

	kb, err := keys.NewKeyBaseFromDir(cfg.home)
	if err != nil {
		return nil, err
	}

	err = amino.UnmarshalJSON(cfg.txJSON, &tx)
	if err != nil {
		return nil, err
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

	// validate document to sign.
	err = tx.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// derive sign doc bytes.
	chainID := cfg.chainID
	accountNumber := cfg.accountNumber
	sequence := cfg.sequence
	signbz := tx.GetSignBytes(chainID, accountNumber, sequence)
	if cfg.showSignBytes {
		return nil, fmt.Errorf("sign bytes: %X\n", signbz)
	}

	sig, pub, err := kb.Sign(cfg.nameOrBech32, cfg.pass, signbz)
	if err != nil {
		return nil, err
	}
	addr := pub.Address()
	found := false
	for i := range tx.Signatures {
		// override signature for matching slot.
		if signers[i] == addr {
			found = true
			tx.Signatures[i] = std.Signature{
				PubKey:    pub,
				Signature: sig,
			}
		}
	}
	if !found {
		return nil, errors.New(
			fmt.Sprintf("addr %v (%s) not in signer set", addr, cfg.nameOrBech32),
		)
	}

	return &tx, nil
}

type broadcastCfg struct {
	// internal
	tx *std.Tx
}

func broadcastHandler(remote string, cfg *broadcastCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	if cfg.tx == nil {
		return nil, errors.New("invalid tx")
	}

	if remote == "" || remote == "y" {
		return nil, errors.New("missing remote url")
	}

	bz, err := amino.Marshal(cfg.tx)
	if err != nil {
		return nil, errors.Wrap(err, "remarshaling tx binary bytes")
	}

	cli := client.NewHTTP(remote, "/websocket")
	bres, err := cli.BroadcastTxCommit(bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	return bres, nil
}
