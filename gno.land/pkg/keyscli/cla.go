package keyscli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

// These values should follow future evolution of the CLA realm.
const (
	claErrorSubstring  = "has not signed the required CLA"
	sysCLARealmDefault = "gno.land/r/sys/cla"
	sysCLAParamPath    = "params/vm:p:syscla_pkgpath"
	sysCLAHashExpr     = "requiredHash"
	sysCLAURLExpr      = "claURL"
)

// isCLAError checks whether the error indicates a CLA signing failure.
// Gates on UnauthorizedUserError, then matches the CLA log message via %#v
func isCLAError(err error) bool {
	if !errors.Is(err, vm.UnauthorizedUserError{}) {
		return false
	}
	return strings.Contains(fmt.Sprintf("%#v", err), claErrorSubstring)
}

// wrapCLAError wraps the original error with a user-friendly CLA signing helper.
// Query failure warnings are appended below the helper message.
func wrapCLAError(err error, remote, chainID, nameOrBech32 string) error {
	var warnings []string

	claRealm, queryErr := queryCLARealmPath(remote)
	if queryErr != nil {
		warnings = append(warnings, fmt.Sprintf("warning: %v, using default %s", queryErr, sysCLARealmDefault))
	}

	hash, claURL, queryErr := queryCLAInfo(remote, claRealm)
	if queryErr != nil {
		warnings = append(warnings, fmt.Sprintf("warning: %v", queryErr))
	}

	helper := formatCLAHelper(hash, claURL, claRealm, chainID, remote, nameOrBech32)
	if len(warnings) > 0 {
		helper += "\n" + strings.Join(warnings, "\n")
	}

	return fmt.Errorf("%w\n%s", err, helper)
}

// queryCLARealmPath returns the CLA realm path from chain params, or the default on failure.
func queryCLARealmPath(remote string) (string, error) {
	cfg := &client.QueryCfg{
		RootCfg: &client.BaseCfg{BaseOptions: client.BaseOptions{Remote: remote}},
		Path:    sysCLAParamPath,
	}
	res, err := client.QueryHandler(cfg)
	if err != nil {
		return sysCLARealmDefault, fmt.Errorf("querying CLA realm path: %w", err)
	}
	if res.Response.Error != nil {
		return sysCLARealmDefault, fmt.Errorf("querying CLA realm path: %s", res.Response.Error.Error())
	}
	if len(res.Response.Data) == 0 {
		return sysCLARealmDefault, fmt.Errorf("querying CLA realm path: empty response")
	}
	path := string(res.Response.Data)
	// Params store returns quoted strings, e.g. "gno.land/r/sys/cla"
	if unquoted, err := strconv.Unquote(path); err == nil {
		path = unquoted
	}
	if path == "" {
		return sysCLARealmDefault, nil
	}
	return path, nil
}

// queryCLAInfo returns the required hash and URL from the CLA realm.
func queryCLAInfo(remote, claRealmPath string) (hash, url string, err error) {
	var errs []string

	hash, hashErr := queryEvalString(remote, claRealmPath, sysCLAHashExpr)
	if hashErr != nil {
		errs = append(errs, hashErr.Error())
	}

	url, urlErr := queryEvalString(remote, claRealmPath, sysCLAURLExpr)
	if urlErr != nil {
		errs = append(errs, urlErr.Error())
	}

	if len(errs) > 0 {
		err = fmt.Errorf("querying CLA info: %s", strings.Join(errs, "; "))
	}
	return
}

// queryEvalString evaluates an expression via vm/qeval and extracts the string result.
func queryEvalString(remote, pkgPath, expr string) (string, error) {
	cfg := &client.QueryCfg{
		RootCfg: &client.BaseCfg{BaseOptions: client.BaseOptions{Remote: remote}},
		Path:    "vm/qeval",
		Data:    pkgPath + "." + expr,
	}
	res, err := client.QueryHandler(cfg)
	if err != nil {
		return "", fmt.Errorf("evaluating %s.%s: %w", pkgPath, expr, err)
	}
	if res.Response.Error != nil {
		return "", fmt.Errorf("evaluating %s.%s: %s", pkgPath, expr, res.Response.Error.Error())
	}
	result, ok := parseQEvalString(string(res.Response.Data))
	if !ok {
		return "", fmt.Errorf("evaluating %s.%s: unexpected response format: %s", pkgPath, expr, string(res.Response.Data))
	}
	return result, nil
}

// parseQEvalString extracts the string value from a '("value" string)' qeval response.
// Returns ok=false when the input does not match the expected shape; ok=true with
// an empty value is a valid result (e.g. '("" string)').
func parseQEvalString(data string) (string, bool) {
	data = strings.TrimSpace(data)
	if len(data) < 2 || data[0] != '(' || data[len(data)-1] != ')' {
		return "", false
	}
	inner := strings.TrimSpace(data[1 : len(data)-1])
	// Split at last space to separate value from type: "<value> string"
	lastSpace := strings.LastIndex(inner, " ")
	if lastSpace < 0 || inner[lastSpace+1:] != "string" {
		return "", false
	}
	val := inner[:lastSpace]
	if unquoted, err := strconv.Unquote(val); err == nil {
		return unquoted, true
	}
	return val, true
}

// formatCLAHelper builds a user-friendly CLA signing helper.
// Missing values are replaced with placeholders (e.g. <CLA_HASH>).
func formatCLAHelper(hash, url, claRealmPath, chainID, remote, nameOrBech32 string) string {
	if hash == "" {
		hash = "<CLA_HASH>"
	}
	if claRealmPath == "" {
		claRealmPath = "<CLA_PKGPATH>"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("A Contributor License Agreement (CLA) must be signed before deploying packages.\n")
	b.WriteString("It clarifies the terms under which your contributions are licensed to the project.\n")
	b.WriteString("The CLA document is defined through a GovDAO governance proposal.\n")
	if url != "" {
		fmt.Fprintf(&b, "\nCLA document: %s\n", url)
	}
	fmt.Fprintf(&b, "\nTo sign the CLA, run:\n\n")
	fmt.Fprintf(&b, "gnokey maketx call -pkgpath %s -func Sign -args %s", claRealmPath, hash)
	fmt.Fprintf(&b, " -gas-fee 100000ugnot -gas-wanted 2000000")
	if remote != "" {
		fmt.Fprintf(&b, " -remote %s", remote)
	}
	if chainID != "" {
		fmt.Fprintf(&b, " -chainid %s", chainID)
	}
	fmt.Fprintf(&b, " %s\n", nameOrBech32)
	return b.String()
}
