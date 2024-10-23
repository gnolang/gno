package gnoland

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
)

type Param struct {
	key string

	string_val string
	int64_val  int64
	uint64_val uint64
	bool_val   bool
	bytes_val  []byte
}

func (p Param) Verify() error {
	// XXX: validate
	return nil
}

func (p *Param) Parse(entry string) error {
	parts := strings.SplitN(strings.TrimSpace(entry), "=", 2) // <key>.<kind>=<value>
	if len(parts) != 2 {
		return fmt.Errorf("malformed entry: %q", entry)
	}

	p.key = parts[0]
	kind := p.Kind()
	raw := parts[1]
	switch kind {
	case "string":
		p.string_val = raw
	case "bool":
		panic("not implemented")
	case "int64":
		panic("not implemented")
	case "uint64":
		panic("not implemented")
	case "bytes":
		panic("not implemented")
	default:
		return errors.New("unsupported param kind: " + kind)
	}

	return p.Verify()
}

func (p Param) Kind() string {
	return p.key[strings.LastIndex(p.key, ".")+1:]
}

func (p Param) String() string {
	kind := p.Kind()
	switch kind {
	case "string":
		return fmt.Sprintf("%s=%s", p.key, p.string_val)
	case "int64":
		return fmt.Sprintf("%s=%d", p.key, p.int64_val)
	case "uint64":
		return fmt.Sprintf("%s=%d", p.key, p.uint64_val)
	case "bool":
		if p.bool_val {
			return fmt.Sprintf("%s=true", p.key)
		}
		return fmt.Sprintf("%s=false", p.key)
	case "bytes":
		return fmt.Sprintf("%s=%x", p.key, p.bytes_val)
	}
	panic("invalid param kind:" + kind)
}

func (p *Param) UnmarshalAmino(rep string) error {
	return p.Parse(rep)
}

func (p Param) MarshalAmino() (string, error) {
	return p.String(), nil
}

func (p Param) Register(ctx sdk.Context, prk params.ParamsKeeperI) {
	kind := p.Kind()
	switch kind {
	case "string":
		println("@@@@", p.key, p.string_val)
		prk.SetString(ctx, p.key, p.string_val)
	case "int64":
		prk.SetInt64(ctx, p.key, p.int64_val)
	case "uint64":
		prk.SetUint64(ctx, p.key, p.uint64_val)
	case "bool":
		prk.SetBool(ctx, p.key, p.bool_val)
	case "bytes":
		prk.SetBytes(ctx, p.key, p.bytes_val)
	default:
		panic("invalid param kind: " + kind)
	}
}
