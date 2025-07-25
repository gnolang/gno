package browser

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

var (
	ErrInternalError  = errors.New("internal error")
	ErrRenderNotFound = errors.New("render not found")
)

type NodeClient struct {
	base   gnoclient.BaseTxCfg
	client *gnoclient.Client
	logger *slog.Logger
}

func NewNodeClient(logger *slog.Logger, base gnoclient.BaseTxCfg, client *gnoclient.Client) *NodeClient {
	return &NodeClient{
		base:   base,
		client: client,
		logger: logger,
	}
}

func (ncl *NodeClient) Call(path, call string) ([]byte, error) {
	method, args, err := parseMethodToArgs(call)
	if err != nil {
		return nil, fmt.Errorf("unable to parse method/args: %w", err)
	}

	if len(args) == 0 {
		args = nil
	}

	infos, err := ncl.client.Signer.Info()
	if err != nil {
		return nil, fmt.Errorf("unable to get signer infos: %w", err)
	}

	cm, err := ncl.client.Call(ncl.base, vm.MsgCall{
		Caller:  infos.GetAddress(),
		PkgPath: path,
		Func:    method,
		Args:    args,
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

func (ncl *NodeClient) Funcs(path string) (vm.FunctionSignatures, error) {
	res, err := ncl.client.Query(gnoclient.QueryCfg{
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
	if err := amino.UnmarshalJSON(res.Response.Data, &fsigs); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response: %w", err)
	}

	return fsigs, nil
}

func (ncl *NodeClient) Render(path, args string) ([]byte, error) {
	data, res, err := ncl.client.Render(path, args)
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
		return "", nil, fmt.Errorf("invalid call: %w", err)
	}

	method = matches[1]
	sargs := matches[2]
	if sargs == "" {
		return method, args, err
	}

	// Splitting arguments by comma
	args = strings.Split(sargs, ",")
	for i, arg := range args {
		args[i] = strings.Trim(strings.TrimSpace(arg), "\"")
	}

	return method, args, err
}
