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
	rawValue := parts[1]
	p.kind = keyWithKind[strings.LastIndex(keyWithKind, ".")+1:]
	p.key = strings.TrimSuffix(keyWithKind, "."+p.kind)
	switch p.kind {
	case ParamKindString:
		p.value = rawValue
	case ParamKindInt64:
		v, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return err
		}
		p.value = v
	case ParamKindBool:
		v, err := strconv.ParseBool(rawValue)
		if err != nil {
			return err
		}
		p.value = v
	case ParamKindUint64:
		v, err := strconv.ParseUint(rawValue, 10, 64)
		if err != nil {
			return err
		}
		p.value = v
	case ParamKindBytes:
		v, err := hex.DecodeString(rawValue)
		if err != nil {
			return err
		}
		p.value = v
	default:
		return errors.New("unsupported param kind: " + p.kind + " (" + entry + ")")
	}

	return p.Verify()
}

func (p Param) String() string {
	typedKey := p.key + "." + p.kind
	switch p.kind {
	case ParamKindString:
		return fmt.Sprintf("%s=%s", typedKey, p.value)
	case ParamKindInt64:
		return fmt.Sprintf("%s=%d", typedKey, p.value)
	case ParamKindUint64:
		return fmt.Sprintf("%s=%d", typedKey, p.value)
	case ParamKindBool:
		if p.value.(bool) {
			return fmt.Sprintf("%s=true", typedKey)
		}
		return fmt.Sprintf("%s=false", typedKey)
	case ParamKindBytes:
		return fmt.Sprintf("%s=%x", typedKey, p.value)
	}
	panic("invalid param kind:" + p.kind)
}

func (p *Param) UnmarshalAmino(rep string) error {
	return p.Parse(rep)
}

func (p Param) MarshalAmino() (string, error) {
	return p.String(), nil
}

func (p Param) register(ctx sdk.Context, prk params.Keeper) error {
	var (
		key = p.key + "." + p.kind

		err error
	)

	switch p.kind {
	case ParamKindString:
		err = prk.SetString(ctx, key, p.value.(string))
	case ParamKindInt64:
		err = prk.SetInt64(ctx, key, p.value.(int64))
	case ParamKindUint64:
		err = prk.SetUint64(ctx, key, p.value.(uint64))
	case ParamKindBool:
		err = prk.SetBool(ctx, key, p.value.(bool))
	case ParamKindBytes:
		err = prk.SetBytes(ctx, key, p.value.([]byte))
	default:
		return fmt.Errorf("unsupported param kind: %s", p.kind)
	}

	if err != nil {
		return fmt.Errorf("could not set %s=%q,  %w", key, p.value, err)
	}

	return nil
}
