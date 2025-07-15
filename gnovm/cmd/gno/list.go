package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"text/template"

	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type listCfg struct {
	json   bool
	deps   bool
	test   bool
	format string
}

func newListCmd(io commands.IO) *commands.Command {
	cfg := &listCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "gno list [flags] <pattern> [patterns...]",
			ShortHelp:  "lists the named packages",
			LongHelp:   "List lists the named packages, one per line.",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execList(cfg, args, io)
		})
}

func (c *listCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.json, "json", false, "Output in JSON")
	fs.BoolVar(&c.deps, "deps", false, "Load dependencies")
	fs.BoolVar(&c.test, "test", false, "Load tests")
	fs.StringVar(&c.format, "f", "", "Output template in go-template format")
}

func execList(cfg *listCfg, args []string, io commands.IO) error {
	if len(args) == 0 {
		args = []string{"."}
	}

	var lw listWriter
	if cfg.json {
		if cfg.format != "" {
			return errors.New("gno list -f cannot be used with -json")
		}
		lw = newJsonListWriter(io.Out())
	} else {
		if cfg.format == "" {
			cfg.format = "{{.ImportPath}}"
		}
		var err error
		lw, err = newTemplateListWriter(io.Out(), cfg.format)
		if err != nil {
			return err
		}
	}

	loadCfg := packages.LoadConfig{
		Fetcher: testPackageFetcher,
		Deps:    cfg.deps,
		Test:    cfg.test,
		Out:     io.Err(),
	}
	pkgs, err := packages.Load(loadCfg, args...)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		if err := lw.write(pkg); err != nil {
			return err
		}
	}

	return nil
}

type listWriter interface {
	write(pkg *packages.Package) error
}

type jsonListWriter struct {
	encoder *json.Encoder
}

func newJsonListWriter(out io.Writer) listWriter {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "\t")
	return &jsonListWriter{
		encoder: encoder,
	}
}

func (lw *jsonListWriter) write(pkg *packages.Package) error {
	return lw.encoder.Encode(pkg)
}

type templateListWriter struct {
	tmpl *template.Template
	out  io.Writer
}

func newTemplateListWriter(out io.Writer, format string) (listWriter, error) {
	tmpl, err := template.New("list-format").Parse(format)
	if err != nil {
		return nil, fmt.Errorf("parse format %q: %w", format, err)
	}
	return &templateListWriter{
		tmpl: tmpl,
		out:  out,
	}, nil
}

func (lw *templateListWriter) write(pkg *packages.Package) error {
	if err := lw.tmpl.Execute(lw.out, pkg); err != nil {
		return err
	}
	_, err := lw.out.Write([]byte{'\n'})
	return err
}
