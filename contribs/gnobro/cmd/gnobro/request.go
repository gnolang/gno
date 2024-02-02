package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
)

const remoteAddr = "http://localhost:36657"
const qFileStr = "vm/qfile"

var (
	ErrInternalError  = errors.New("internal error")
	ErrRenderNotFound = errors.New("render not found")
)

type Request struct {
	RemoteAddr string
	QPath      string
	Data       []byte
}

func makeRequest(log log.Logger, req Request) (res *abci.ResponseQuery, err error) {
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	remote := req.RemoteAddr
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		req.QPath, req.Data, opts2)
	if err != nil {
		log.Error("request error", "path", req.QPath, "error", err)
		return nil, fmt.Errorf("unable to query path %q: %w", req.QPath, err)
	}
	if qres.Response.Error != nil {
		log.Error("response error", "path", req.QPath, "log", qres.Response.Log)
		return nil, fmt.Errorf("response error: %s\n\n%s\n", qres.Response.Error, qres.Response.Log)
	}
	return &qres.Response, nil
}

func makeRender(logger log.Logger, rlmpath string) ([]byte, error) {
	var req Request
	req.RemoteAddr = remoteAddr

	req.Data = []byte(fmt.Sprintf("%s\n%s", rlmpath, ""))
	req.QPath = "vm/qrender"
	res, err := makeRequest(logger, req)
	if err != nil {
		if strings.Contains(err.Error(), "Render not declared") {
			return nil, ErrRenderNotFound
		}

		return nil, fmt.Errorf("unable to make request on %q: %w", rlmpath, err)
	}

	return res.Data, nil
}

func makeFuncs(logger log.Logger, rlmpath string) (vm.FunctionSignatures, error) {
	var req Request
	req.RemoteAddr = remoteAddr

	req.QPath = "vm/qfuncs"
	req.Data = []byte(rlmpath)
	res, err := makeRequest(logger, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	var fsigs vm.FunctionSignatures
	amino.MustUnmarshalJSON(res.Data, &fsigs)
	// Fill fsigs with query parameters.
	// for i := range fsigs {
	// 	fsig := &(fsigs[i])
	// 	for j := range fsig.Params {
	// 		param := &(fsig.Params[j])
	// 		value := query.Get(param.Name)
	// 		param.Value = value
	// 	}
	// }
	return fsigs, nil
}

type makeCallCfg struct {
	kb           keys.Keybase
	pass         string
	rlmpath      string
	eval         string
	nameOrBech32 string
}

// gnokey maketx call -pkgpath "gno.land/r/dev/hello" -func "Inc" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "tendermint-test" -remote "http://127.0.0.1:36657" g1jg8mtut
func makeCall(logger log.Logger, makecfg makeCallCfg) ([]byte, error) {
	var cfg callCfg

	var err error
	cfg.funcName, cfg.args, err = parseMethodToArgs(makecfg.eval)
	if err != nil {
		return nil, err
	}

	if len(cfg.args) == 0 {
		cfg.args = nil
	}

	cfg.gasFee = "1000000ugnot"
	cfg.gasWanted = 2000000
	cfg.send = ""
	cfg.broadcast = true
	cfg.chainID = "tendermint_test"
	cfg.remote = remoteAddr
	cfg.pkgPath = makecfg.rlmpath
	cfg.kb = makecfg.kb

	res, err := execCall(makecfg.nameOrBech32, makecfg.pass, cfg)
	if err != nil {
		return nil, err
	}

	return res.DeliverTx.Data, nil
}

var reMethod = regexp.MustCompile(`([^(]+)\(([^)]*)\)`)

func parseMethodToArgs(call string) (method string, args []string, err error) {
	matches := reMethod.FindStringSubmatch(call)
	if len(matches) == 0 {
		err = fmt.Errorf("invalid call: %w", err)
		return
	}

	method = matches[1]
	sargs := matches[2]

	// Splitting arguments by comma
	args = strings.Split(sargs, ",")
	for i, arg := range args {
		args[i] = strings.Trim(strings.TrimSpace(arg), "\"")
	}
	return
}
