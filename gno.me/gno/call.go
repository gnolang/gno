package gno

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gno.me/state"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

type MsgCall struct {
	AppName string   `json:"app_name"`
	Func    string   `json:"func"`
	Args    []string `json:"args"`
}

// TODO: make this also return a slice of events for the cases where a call
// may result in multiple events for non-deterministic operations.
func (v *VMKeeper) Call(
	ctx context.Context,
	appName string,
	isPackage bool,
	functionName string,
	args ...string,
) (string, []*state.Event, error) {
	v.Lock()
	defer v.Unlock()
	defer v.store.Commit()

	fmt.Println("Call", appName, isPackage, functionName, args)

	prefix := AppPrefix
	if isPackage {
		prefix = PkgPrefix
	}

	pkgPath := prefix + appName
	msg := vm.MsgCall{
		PkgPath: pkgPath,
		Func:    functionName,
		Args:    append([]string{}, args...),
	}

	result, err := v.instance.Call(sdk.Context{}.WithContext(ctx), msg)
	if err != nil {
		return "", nil, err
	}

	if isPackage || functionName == "Render" {
		return result, nil, nil
	}

	msg = vm.MsgCall{
		PkgPath: prefix + "events",
		Func:    "NextSequence",
		Args:    []string{pkgPath},
	}

	seqResult, err := v.instance.Call(sdk.Context{}.WithContext(ctx), msg)
	if err != nil {
		panic("unable to obtain sequence: cannot proceed without persisting event: " + err.Error())
	}

	sequence := seqResult[1:strings.Index(seqResult, " ")]
	msg = vm.MsgCall{
		PkgPath: prefix + "events",
		Func:    "Store",
		Args: []string{
			pkgPath,
			sequence,
			functionName,
			encodeArgs(args),
		},
	}

	_, err = v.instance.Call(sdk.Context{}.WithContext(ctx), msg)
	if err != nil {
		panic("unable to store event: cannot proceed without persisting event: " + err.Error())
	}

	event := state.Event{
		Sequence: sequence,
		AppName:  appName,
		Func:     functionName,
		Args:     args,
	}

	return result, []*state.Event{&event}, nil
}

func encodeArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, arg := range args {
		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(url.QueryEscape(arg))
	}

	return sb.String()
}

func decodeArgs(args string) ([]string, error) {
	if args == "" {
		return nil, nil
	}

	var decoded []string
	for _, arg := range strings.Split(args, ",") {
		unescaped, err := url.QueryUnescape(arg)
		if err != nil {
			return nil, fmt.Errorf("error decoding args -- data corrupt: %w", err)
		}

		decoded = append(decoded, unescaped)
	}

	return decoded, nil
}
