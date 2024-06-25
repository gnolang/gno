package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type metadataHelperGenerator struct {
	// Optional callback to edit metadata
	MetaUpdate func(*commands.Metadata)
	// Tag to select for name, if empty will use the field Name
	TagNameSelector string
	// Will display description with tree representation
	TreeDisplay bool
}

func generateSubCommandHelper(gen metadataHelperGenerator, s any, exec commands.ExecMethod) []*commands.Command {
	rv := reflect.ValueOf(s)
	metas := gen.generateFields(rv, "", 0)

	cmds := make([]*commands.Command, len(metas))
	for i := 0; i < len(metas); i++ {
		meta := metas[i]
		exec := func(ctx context.Context, args []string) error {
			args = append([]string{meta.Name}, args...)
			return exec(ctx, args)
		}
		cmds[i] = commands.NewCommand(meta, nil, exec)
	}

	return cmds
}

func (g *metadataHelperGenerator) generateFields(rv reflect.Value, parent string, depth int) []commands.Metadata {
	if parent != "" {
		parent += "."
	}

	// Unwrap pointer if needed
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			// Create a new non-nil instance of the original type that was nil
			rv = reflect.New(rv.Type().Elem())
		}
		rv = rv.Elem() // Dereference to struct value
	}

	metas := []commands.Metadata{}
	if rv.Kind() != reflect.Struct {
		return metas
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
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

		// Recursive call for nested struct
		var childs []commands.Metadata
		if k := fieldValue.Kind(); k == reflect.Ptr || k == reflect.Struct {
			childs = g.generateFields(fieldValue, name, depth+1)
		}

		// Generate metadata
		var meta commands.Metadata

		// Name
		meta.Name = parent + name

		// Create a tree-like display to see nested field
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
		}

		if g.MetaUpdate != nil {
			g.MetaUpdate(&meta)
		}

		metas = append(metas, meta)
		metas = append(metas, childs...)

	}

	return metas
}
