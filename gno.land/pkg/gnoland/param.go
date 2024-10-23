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

const (
	ParamKindString = "string"
	ParamKindInt64  = "int64"
	ParamKindUint64 = "uint64"
	ParamKindBool   = "bool"
	ParamKindBytes  = "bytes"
)

func (p *Param) Parse(entry string) error {
	parts := strings.SplitN(strings.TrimSpace(entry), "=", 2) // <key>.<kind>=<value>
	if len(parts) != 2 {
		return fmt.Errorf("malformed entry: %q", entry)
	}

	p.key = parts[0]
	kind := p.Kind()
	raw := parts[1]
	switch kind {
	case ParamKindString:
		p.string_val = raw
	case ParamKindBool:
		panic("not implemented")
	case ParamKindInt64:
		panic("not implemented")
	case ParamKindUint64:
		panic("not implemented")
	case ParamKindBytes:
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
	case ParamKindString:
		return fmt.Sprintf("%s=%s", p.key, p.string_val)
	case ParamKindInt64:
		return fmt.Sprintf("%s=%d", p.key, p.int64_val)
	case ParamKindUint64:
		return fmt.Sprintf("%s=%d", p.key, p.uint64_val)
	case ParamKindBool:
		if p.bool_val {
			return fmt.Sprintf("%s=true", p.key)
		}
		return fmt.Sprintf("%s=false", p.key)
	case ParamKindBytes:
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
	case ParamKindString:
		prk.SetString(ctx, p.key, p.string_val)
	case ParamKindInt64:
		prk.SetInt64(ctx, p.key, p.int64_val)
	case ParamKindUint64:
		prk.SetUint64(ctx, p.key, p.uint64_val)
	case ParamKindBool:
		prk.SetBool(ctx, p.key, p.bool_val)
	case ParamKindBytes:
		prk.SetBytes(ctx, p.key, p.bytes_val)
	default:
		panic("invalid param kind: " + kind)
	}
}
