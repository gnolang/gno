package cluster

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Override is one dotted key a scenario sets on the node config or on the
// genesis params, kept as text until it is applied to the typed target.
type Override struct{ Key, Value string }

// applyOverride sets the field at o.Key on root to o.Value. The key is a path
// of json tags, which is the vocabulary `gnoland config set` and `gnogenesis
// params set` already take, so a scenario spells an option the way the CLI
// does.
func applyOverride(root reflect.Value, o Override) error {
	field, err := commands.GetFieldByPath(root, "json", strings.Split(o.Key, "."))
	if err != nil {
		return fmt.Errorf("%s: %w", o.Key, err)
	}
	if err := setFromString(*field, o.Value); err != nil {
		return fmt.Errorf("%s: %w", o.Key, err)
	}
	return nil
}

// setFromString converts value to the destination field's type and stores it.
// A scenario states every option as text, the way a command line does, so
// something has to carry it to the typed field.
//
// ceiling: the third copy of this conversion in the repo, after
// gno.land/cmd/gnoland's config set and contribs/gnogenesis's params set.
// Upgrade path is to export one from tm2/pkg/commands next to GetFieldByPath
// and delete all three.
func setFromString(dst reflect.Value, value string) error {
	// Kind, not concrete type, so a named string type such as
	// vm.CodeSubmissionPolicy is set rather than falling through to
	// "unsupported": `case string` misses it.
	if dst.Kind() == reflect.String {
		dst.SetString(value)
		return nil
	}

	badValue := func(err error) error {
		return fmt.Errorf("%q is not a valid %s: %w", value, dst.Type(), err)
	}

	switch dst.Interface().(type) {
	case []string:
		dst.Set(reflect.ValueOf(splitList(value)))
	case time.Duration:
		d, err := time.ParseDuration(value)
		if err != nil {
			return badValue(err)
		}
		dst.Set(reflect.ValueOf(d))
	case crypto.Address:
		addr, err := crypto.AddressFromBech32(value)
		if err != nil {
			return badValue(err)
		}
		dst.Set(reflect.ValueOf(addr))
	case []crypto.Address:
		items := splitList(value)
		addrs := make([]crypto.Address, 0, len(items))
		for _, item := range items {
			addr, err := crypto.AddressFromBech32(item)
			if err != nil {
				return badValue(err)
			}
			addrs = append(addrs, addr)
		}
		dst.Set(reflect.ValueOf(addrs))
	case std.GasPrice:
		price, err := std.ParseGasPrice(value)
		if err != nil {
			return badValue(err)
		}
		dst.Set(reflect.ValueOf(price))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		if err := json.Unmarshal([]byte(value), dst.Addr().Interface()); err != nil {
			return badValue(err)
		}
	default:
		return fmt.Errorf("no conversion from text to %s", dst.Type())
	}
	return nil
}

// splitList reads a list stated as one comma-separated value, the way the
// genesis and config CLIs read theirs. Empty items are dropped, so a trailing
// comma is not a blank entry.
func splitList(value string) []string {
	items := make([]string, 0, strings.Count(value, ",")+1)
	for item := range strings.SplitSeq(value, ",") {
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
