package gnopkgfetch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

type examplesMockClient struct {
	examplesRoot string
}

func (m *examplesMockClient) SendRequest(ctx context.Context, request types.RPCRequest) (*types.RPCResponse, error) {
	params := struct {
		Path string `json:"path"`
		Data []byte `json:"data"`
	}{}
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}
	path := params.Path
	if path != "vm/qfile" {
		return nil, fmt.Errorf("unexpected call to %q", path)
	}
	data := string(params.Data)

	target := filepath.Join(m.examplesRoot, data)

	res := ctypes.ResultABCIQuery{}

	finfo, err := os.Stat(target)
	if os.IsNotExist(err) {
		res.Response = sdk.ABCIResponseQueryFromError(fmt.Errorf("package %q is not available", data))
		return &types.RPCResponse{
			Result: amino.MustMarshalJSON(res),
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat %q: %w", data, err)
	}

	if finfo.IsDir() {
		entries, err := os.ReadDir(target)
		if err != nil {
			return nil, fmt.Errorf("failed to get package %q: %w", data, err)
		}
		files := []string{}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}
		res.Response.Data = []byte(strings.Join(files, "\n"))
	} else {
		content, err := os.ReadFile(target)
		if err != nil {
			return nil, fmt.Errorf("failed to get file %q: %w", data, err)
		}
		res.Response.Data = content
	}

	return &types.RPCResponse{
		Result: amino.MustMarshalJSON(res),
	}, nil
}

func (m *examplesMockClient) SendBatch(ctx context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
	return nil, errors.New("not implemented")
}

func (m *examplesMockClient) Close() error {
	return nil
}
