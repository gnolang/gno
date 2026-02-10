package commands

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode"
)

type FieldsGenerator struct {
	// Optional callback to edit metadata
	MetaUpdate func(meta *Metadata, inputType string)
	// Tag to select for name, if empty will use the field Name
	TagNameSelector string
	// Will display description with tree representation
	TreeDisplay bool
	// Depth specifies the starting depth for generating commands from the fields.
	// If set, command generation will begin at this depth.
	Depth int
}

type parsedField struct {
	Metadata

	depth int
	names []string
}

// GenerateFrom generates a list of CLI subcommands for each field in the struct `s`,
// using field tags and metadata for help texts and names. Each generated command
// includes its full path in the struct as a dot-separated string (e.g., "foo.bar.baz").
// This path can be later used with `GetFieldByPath` to retrieve or manipulate
// the actual struct field at runtime, ensuring a direct mapping between the generated
// CLI command and its underlying field.
func (g *FieldsGenerator) GenerateFrom(s any, exec ExecMethod) []*Command {
	rv := reflect.ValueOf(s)
	fields := g.generateFields(rv, []string{}, 0)

	cmds := make([]*Command, 0, len(fields))
	for _, meta := range fields {
		if meta.depth < g.Depth {
			continue
		}

		exec := func(ctx context.Context, args []string) error {
			args = append([]string{meta.Name}, args...)
			return exec(ctx, args)
		}

		cmds = append(cmds, NewCommand(meta.Metadata, nil, exec))
	}

	return slices.Clip(cmds)
}

func (g *FieldsGenerator) generateFields(rv reflect.Value, parents []string, depth int) []parsedField {
	// Unwrap pointer if needed
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			// Create a new non-nil instance of the original type that was nil
			rv = reflect.New(rv.Type().Elem())
		}
		rv = rv.Elem() // Dereference to struct value
	}

	metas := []parsedField{}
	if rv.Kind() != reflect.Struct {
		return metas
	}

	rt := rv.Type()
	for i := range rv.NumField() {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldValue := rv.Field(i)
		name := field.Name
		// Get JSON tag name
		if g.TagNameSelector != "" {
			name, _, _ = strings.Cut(field.Tag.Get(g.TagNameSelector), ",")
			if name == "" || name == "-" {
				continue
			}
		}

		// Generate metadata
		meta := parsedField{
			depth: depth,
			names: append(parents, name),
		}
		meta.depth = depth

		// Name
		meta.Name = strings.Join(meta.names, ".")

		// Create a tree-like display to see nested fields
		if g.TreeDisplay && depth > 0 {
			meta.ShortHelp += strings.Repeat(" ", depth*2)
			if i == rv.NumField()-1 {
				meta.ShortHelp += "└─"
			} else {
				meta.ShortHelp += "├─"
			}
		}
		meta.ShortHelp += fmt.Sprintf("<%s>", field.Type)

		// Get Short/Long Help Message from comment tag
		comment := field.Tag.Get("comment")
		comment = strings.TrimFunc(comment, func(r rune) bool {
			return unicode.IsSpace(r) || r == '#'
		})

		if comment != "" {
			// Use the first line as short help
			meta.ShortHelp += " "
			meta.ShortHelp += strings.Split(comment, "\n")[0]

			// Display full comment as Long Help
			meta.LongHelp = comment
		} else {
			// If the comment is empty, it mostly means that there is no help.
			// Use a blank space to avoid falling back on short help.
			meta.LongHelp = " "
		}

		if g.MetaUpdate != nil {
			g.MetaUpdate(&meta.Metadata, field.Type.String())
		}

		// Recursive call for nested struct
		var childs []parsedField
		if k := fieldValue.Kind(); k == reflect.Ptr || k == reflect.Struct {
			childs = g.generateFields(fieldValue, meta.names, depth+1)
		}

		metas = append(metas, meta)
		metas = append(metas, childs...)
	}

	return metas
}

// GetFieldByPath traverses a struct (starting from `currentValue`) following
// the provided path (a slice of field names, e.g., ["foo", "bar", "baz"]) and returns
// a pointer to the reflect.Value of the target field
//
// This function is designed to work in tandem with `GenerateFrom`: the path provided to
// `GetFieldByPath` should match the path generated for each command in `GenerateFrom`.
func GetFieldByPath(currentValue reflect.Value, selTag string, path []string) (*reflect.Value, error) {
	// Examine the current section to determine if it's part of the current struct
	field := fieldByName(currentValue, selTag, path[0])
	if !field.IsValid() || !field.CanSet() {
		return nil, newInvalidFieldError(path[0], selTag, currentValue)
	}

	// Dereference the field if needed
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	// Check if this is not the end of the path
	// e.g., x.y.field
	if len(path) > 1 {
		// Recursively try to traverse the path and return the specified field
		return GetFieldByPath(field, selTag, path[1:])
	}

	return &field, nil
}

func fieldByName(value reflect.Value, selTag, name string) reflect.Value {
	var dst reflect.Value
	eachField(value, selTag, func(val reflect.Value, fname string) bool {
		if name == fname {
			dst = val
			return true
		}
		return false
	})
	return dst
}

// eachField iterates over each field in value (assumed to be a struct).
// For every field within the struct, iterationCallback is called with the value
// of the field and the associated name.
// (through the `<tagName>:"..."` struct field, or just the field's name as fallback).
// If iterationCallback returns true, the function returns immediately with
// true. If it always returns false, or no fields were processed, eachField
// will return false.
func eachField(value reflect.Value, selTag string, iterationCallback func(val reflect.Value, name string) bool) bool {
	// For reference:
	// https://github.com/pelletier/go-toml/blob/7dad87762adb203e30b96a46026d1428ef2491a2/unmarshaler.go#L1251-L1270

	currentType := value.Type()
	nf := currentType.NumField()
	for i := range nf {
		fld := currentType.Field(i)
		name := fld.Tag.Get(selTag)

		// Ignore `toml:"-"`, strip away any "omitempty" or other options.
		if name == "-" || !fld.IsExported() {
			continue
		}
		if pos := strings.IndexByte(name, ','); pos != -1 {
			name = name[:pos]
		}

		// Handle anonymous (embedded) fields.
		// Anonymous fields will be treated regularly if they have a tag.
		if fld.Anonymous && name == "" {
			anon := fld.Type
			if anon.Kind() == reflect.Ptr {
				anon = anon.Elem()
			}

			if anon.Kind() == reflect.Struct {
				// NOTE: naive, if there is a conflict the embedder should take
				// precedence over the embedded; but the TOML parser seems to
				// ignore this, too, and this "unmarshaler" isn't fit for general
				// purpose.
				if eachField(value.Field(i), selTag, iterationCallback) {
					return true
				}
				continue
			}
			// If it's not a struct, or *struct, it should be treated regularly.
		}

		// general case, simple struct field.
		if name == "" {
			name = fld.Name
		}
		if iterationCallback(value.Field(i), name) {
			return true
		}
	}
	return false
}

// newInvalidFieldError creates an error for non-existent struct fields
// being passed as arguments to [getFieldAtPath]
func newInvalidFieldError(field, selTag string, value reflect.Value) error {
	fields := make([]string, 0, value.NumField())

	eachField(value, selTag, func(val reflect.Value, name string) bool {
		fields = append(fields, name)
		return false
	})

	return fmt.Errorf(
		"field %q, is not a valid configuration key, available keys: %s",
		field,
		fields,
	)
}
