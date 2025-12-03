package rpc

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

// TODO update response types, since now we can properly encode them

// VMEval evaluates a call to an exported function without using gas, in read-only mode
func (s *Server) VMEval(_ *rpctypes.Context, height int64, realm, expr string) (string, error) {
	ctx, err := s.app.NewQueryContext(height)
	if err != nil {
		return "", fmt.Errorf("unable to create query context: %w", err)
	}

	result, err := s.app.VMKeeper().QueryEval(ctx, realm, expr)
	if err != nil {
		return "", fmt.Errorf("unable to evaluate expression: %w", err)
	}

	return result, nil
}

// VMRender evaluates the "Render" function call
func (s *Server) VMRender(_ *rpctypes.Context, height int64, pkgPath, path string) (string, error) {
	ctx, err := s.app.NewQueryContext(height)
	if err != nil {
		return "", fmt.Errorf("unable to create query context: %w", err)
	}

	expr := fmt.Sprintf("Render(%q)", path)
	result, err := s.app.VMKeeper().QueryEval(ctx, pkgPath, expr)
	if err != nil {
		if strings.Contains(err.Error(), "Render not declared") {
			err = vm.NoRenderDeclError{}
		}

		return "", fmt.Errorf("unable to call Render: %w", err)
	}

	return result, nil
}

// VMFuncs returns the exported functions for the given package path
func (s *Server) VMFuncs(_ *rpctypes.Context, height int64, pkgPath string) (string, error) {
	ctx, err := s.app.NewQueryContext(height)
	if err != nil {
		return "", fmt.Errorf("unable to create query context: %w", err)
	}

	funcSigs, err := s.app.VMKeeper().QueryFuncs(ctx, pkgPath)
	if err != nil {
		return "", err
	}

	return funcSigs.JSON(), nil
}

// VMPaths lists all existing package paths prefixed with the specified target string, paginated
func (s *Server) VMPaths(_ *rpctypes.Context, height int64, target string, limit int) (string, error) {
	const (
		defaultLimit = 1_000
		maxLimit     = 10_000
	)

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	ctx, err := s.app.NewQueryContext(height)
	if err != nil {
		return "", fmt.Errorf("unable to create query context: %w", err)
	}

	paths, err := s.app.VMKeeper().QueryPaths(ctx, target, limit)
	if err != nil {
		return "", err
	}

	return strings.Join(paths, "\n"), nil
}

// VMFile returns package contents for a given package path
func (s *Server) VMFile(_ *rpctypes.Context, height int64, filepath string) (string, error) {
	ctx, err := s.app.NewQueryContext(height)
	if err != nil {
		return "", fmt.Errorf("unable to create query context: %w", err)
	}

	result, err := s.app.VMKeeper().QueryFile(ctx, filepath)
	if err != nil {
		return "", err
	}

	return result, nil
}

// VMDoc returns the JSON of the doc for a given package path, suitable for printing
func (s *Server) VMDoc(_ *rpctypes.Context, height int64, pkgPath string) (string, error) {
	ctx, err := s.app.NewQueryContext(height)
	if err != nil {
		return "", fmt.Errorf("unable to create query context: %w", err)
	}

	jsonDoc, err := s.app.VMKeeper().QueryDoc(ctx, pkgPath)
	if err != nil {
		return "", err
	}

	return jsonDoc.JSON(), nil
}

// VMStorage returns storage usage and deposit locked in a realm
func (s *Server) VMStorage(_ *rpctypes.Context, height int64, pkgPath string) (string, error) {
	ctx, err := s.app.NewQueryContext(height)
	if err != nil {
		return "", fmt.Errorf("unable to create query context: %w", err)
	}

	result, err := s.app.VMKeeper().QueryStorage(ctx, pkgPath)
	if err != nil {
		return "", err
	}

	return result, nil
}

// TODO add simulate
