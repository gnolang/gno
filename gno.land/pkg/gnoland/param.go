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
	key string

	stringVal string
	int64Val  int64
	uint64Val uint64
	boolVal   bool
	bytesVal  []byte
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
		p.stringVal = raw
	case ParamKindInt64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		p.int64Val = v
	case ParamKindBool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		p.boolVal = v
	case ParamKindUint64:
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		p.uint64Val = v
	case ParamKindBytes:
		v, err := hex.DecodeString(raw)
		if err != nil {
			return err
		}
		p.bytesVal = v
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
		return fmt.Sprintf("%s=%s", p.key, p.stringVal)
	case ParamKindInt64:
		return fmt.Sprintf("%s=%d", p.key, p.int64Val)
	case ParamKindUint64:
		return fmt.Sprintf("%s=%d", p.key, p.uint64Val)
	case ParamKindBool:
		if p.boolVal {
			return fmt.Sprintf("%s=true", p.key)
		}
		return fmt.Sprintf("%s=false", p.key)
	case ParamKindBytes:
		return fmt.Sprintf("%s=%x", p.key, p.bytesVal)
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
		prk.SetString(ctx, p.key, p.stringVal)
	case ParamKindInt64:
		prk.SetInt64(ctx, p.key, p.int64Val)
	case ParamKindUint64:
		prk.SetUint64(ctx, p.key, p.uint64Val)
	case ParamKindBool:
		prk.SetBool(ctx, p.key, p.boolVal)
	case ParamKindBytes:
		prk.SetBytes(ctx, p.key, p.bytesVal)
	default:
		panic("invalid param kind: " + kind)
	}
}
