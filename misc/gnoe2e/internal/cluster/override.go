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

// The struct tag each family of keys is resolved by, so a scenario spells an
// option the way the CLI that owns the same target spells it. "toml" is what
// `gnoland config set` resolves, and the only selector that names the
// top-level config section's keys, because config.BaseConfig carries no json
// tags; "json" is what `gnogenesis params set` resolves.
const (
	nodeConfigSelector    = "toml"
	genesisParamsSelector = "json"
)

// applyOverride sets the field at o.Key on root to o.Value. The key is a path
// of struct tags named by selTag, so a scenario spells an option the way the
// CLI that owns the same target spells it: "toml" for the node config, which is
// what `gnoland config set` resolves, and "json" for the genesis params, which
// is what `gnogenesis params set` resolves.
func applyOverride(root reflect.Value, selTag string, o Override) error {
	field, err := resolveField(root, selTag, o.Key)
	if err != nil {
		return fmt.Errorf("%s: %w", o.Key, err)
	}
	if err := setFromString(*field, o.Value); err != nil {
		return fmt.Errorf("%s: %w", o.Key, err)
	}
	return nil
}

// resolveField walks key one segment at a time, refusing to descend into
// anything but a struct.
//
// GetFieldByPath cannot be handed the whole path: past a scalar leaf it asks a
// non-struct for its fields and panics ("reflect: NumField of non-struct type
// int"), and a scenario file is where the extra segment comes from. One segment
// at a time keeps the tag matching upstream's and makes the overrun an error.
func resolveField(root reflect.Value, selTag, key string) (*reflect.Value, error) {
	segments := strings.Split(key, ".")
	current := root
	for i, segment := range segments {
		// A nil pointer section dereferences to the zero Value, which cannot
		// even be asked for its type.
		if !current.IsValid() {
			return nil, fmt.Errorf("%s is not set, so it has no field %q",
				strings.Join(segments[:i], "."), segment)
		}
		if current.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%s has type %s, not a section, so it has no field %q",
				strings.Join(segments[:i], "."), current.Type(), segment)
		}
		field, err := commands.GetFieldByPath(current, selTag, []string{segment})
		if err != nil {
			return nil, err
		}
		current = *field
	}
	return &current, nil
}

// fieldDefaults lists every path under root that a scenario can set, keyed by
// selTag, with the value the field currently holds. Declaration order, which
// is the order the same keys appear in config.toml.
func fieldDefaults(root reflect.Value, selTag string) []Override {
	return appendFieldDefaults(nil, root, selTag, "")
}

func appendFieldDefaults(dst []Override, value reflect.Value, selTag, prefix string) []Override {
	structType := value.Type()
	for i := range structType.NumField() {
		field := structType.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get(selTag), ",")
		if !field.IsExported() || name == "-" {
			continue
		}

		// An embedded section with no tag of its own contributes its fields
		// under the enclosing path, the way the tag readers treat it.
		if field.Anonymous && name == "" {
			if section, ok := sectionValue(value.Field(i)); ok {
				dst = appendFieldDefaults(dst, section, selTag, prefix)
				continue
			}
		}
		if name == "" {
			name = field.Name
		}

		if section, ok := sectionValue(value.Field(i)); ok {
			dst = appendFieldDefaults(dst, section, selTag, prefix+name+".")
			continue
		}
		if value.Field(i).Kind() == reflect.Pointer {
			// A nil section has no fields to list and no field to set: the
			// path through it is an error rather than a default.
			continue
		}
		dst = append(dst, Override{Key: prefix + name, Value: fieldText(value.Field(i))})
	}
	return dst
}

// sectionValue reports whether value is a section to descend into rather than
// a field to print. std.GasPrice is a struct and is still stated as one string
// ("1ugnot/1000gas"), so it is a field.
func sectionValue(value reflect.Value) (reflect.Value, bool) {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct || value.Type() == reflect.TypeFor[std.GasPrice]() {
		return reflect.Value{}, false
	}
	return value, true
}

// fieldText renders a field the way a scenario states it, so a listed default
// can be pasted back into a "-- cluster --" section.
func fieldText(value reflect.Value) string {
	switch typed := value.Interface().(type) {
	case []string:
		return strings.Join(typed, ",")
	case []crypto.Address:
		items := make([]string, len(typed))
		for i, addr := range typed {
			items[i] = addr.String()
		}
		return strings.Join(items, ",")
	default:
		return fmt.Sprint(typed)
	}
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
