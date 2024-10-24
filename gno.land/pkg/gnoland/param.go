package gnoland

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
)

type Param struct {
	key   string
	kind  string
	value interface{}
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

	keyWithKind := parts[0]
	p.kind = keyWithKind[strings.LastIndex(keyWithKind, ".")+1:]
	p.key = strings.TrimSuffix(keyWithKind, "."+p.kind)
	raw := parts[1]
	switch p.kind {
	case ParamKindString:
		p.value = raw
	case ParamKindInt64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		p.value = v
	case ParamKindBool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		p.value = v
	case ParamKindUint64:
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		p.value = v
	case ParamKindBytes:
		v, err := hex.DecodeString(raw)
		if err != nil {
			return err
		}
		p.value = v
	default:
		return errors.New("unsupported param kind: " + p.kind)
	}

	return p.Verify()
}

func (p Param) String() string {
	switch p.kind {
	case ParamKindString:
		return fmt.Sprintf("%s=%s", p.key, p.value)
	case ParamKindInt64:
		return fmt.Sprintf("%s=%d", p.key, p.value)
	case ParamKindUint64:
		return fmt.Sprintf("%s=%d", p.key, p.value)
	case ParamKindBool:
		if p.value.(bool) {
			return fmt.Sprintf("%s=true", p.key)
		}
		return fmt.Sprintf("%s=false", p.key)
	case ParamKindBytes:
		return fmt.Sprintf("%s=%x", p.key, p.value)
	}
	panic("invalid param kind:" + p.kind)
}

func (p *Param) UnmarshalAmino(rep string) error {
	return p.Parse(rep)
}

func (p Param) MarshalAmino() (string, error) {
	return p.String(), nil
}

func (p Param) Register(ctx sdk.Context, prk params.ParamsKeeperI) {
	key := p.key + "." + p.kind
	switch p.kind {
	case ParamKindString:
		prk.SetString(ctx, key, p.value.(string))
	case ParamKindInt64:
		prk.SetInt64(ctx, key, p.value.(int64))
	case ParamKindUint64:
		prk.SetUint64(ctx, key, p.value.(uint64))
	case ParamKindBool:
		prk.SetBool(ctx, key, p.value.(bool))
	case ParamKindBytes:
		prk.SetBytes(ctx, key, p.value.([]byte))
	default:
		panic("invalid param kind: " + p.kind)
	}
}
