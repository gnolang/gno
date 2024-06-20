package main

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

const remoteAddr = "http://localhost:36657"
const qFileStr = "vm/qfile"

var (
	ErrInternalError  = errors.New("internal error")
	ErrRenderNotFound = errors.New("render not found")
)

type BroClient struct {
	base   gnoclient.BaseTxCfg
	client *gnoclient.Client
	logger *slog.Logger
}

func NewBroClient(logger *slog.Logger, base gnoclient.BaseTxCfg, client *gnoclient.Client) *BroClient {
	return &BroClient{
		base:   base,
		client: client,
		logger: logger,
	}
}

// gnokey maketx call -pkgpath "gno.land/r/dev/hello" -func "Inc" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "tendermint-test" -remote "http://127.0.0.1:36657" g1jg8mtut
func (bl *BroClient) Call(path, call string) ([]byte, error) {
	method, args, err := parseMethodToArgs(call)
	if err != nil {
		return nil, fmt.Errorf("unable to parse method/args: %w", err)
	}

	if len(args) == 0 {
		args = nil
	}

	cm, err := bl.client.Call(bl.base, gnoclient.MsgCall{
		PkgPath:  path,
		FuncName: method,
		Args:     args,
	})

	if err != nil {
		return nil, err
	}

	if cm.CheckTx.Error != nil {
		return nil, fmt.Errorf("check error: %w", err)
	}

	if cm.DeliverTx.Error != nil {
		return nil, fmt.Errorf("delivry error: %w", err)
	}

	return cm.DeliverTx.Data, nil
}

func (bl *BroClient) Funcs(path string) (vm.FunctionSignatures, error) {
	res, err := bl.client.Query(gnoclient.QueryCfg{
		Path: "vm/qfuncs",
		Data: []byte(path),
	})
	if err != nil {
		return nil, err
	}

	if err := res.Response.Error; err != nil {
		return nil, err
	}

	var fsigs vm.FunctionSignatures
	amino.MustUnmarshalJSON(res.Response.Data, &fsigs)
	return fsigs, nil
}

func (bl *BroClient) Render(path, args string) ([]byte, error) {
	data, res, err := bl.client.Render(path, args)
	if err != nil {
		return nil, err
	}
	if err := res.Response.Error; err != nil {
		return nil, err
	}

	return []byte(data), nil
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
	if sargs == "" {
		return
	}

	// Splitting arguments by comma
	args = strings.Split(sargs, ",")
	for i, arg := range args {
		args[i] = strings.Trim(strings.TrimSpace(arg), "\"")
	}
	return
}
