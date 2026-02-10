package main

import (
	"github.com/gnolang/gno/contribs/gnodev/pkg/rawterm"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var noopRestore = func() error { return nil }

func setupRawTerm(cfg *AppConfig, io commands.IO) (*rawterm.RawTerm, func() error, error) {
	rt := rawterm.NewRawTerm()
	restore := noopRestore
	if cfg.interactive {
		var err error
		restore, err = rt.Init()
		if err != nil {
			return nil, nil, err
		}
	}

	// correctly format output for terminal
	io.SetOut(commands.WriteNopCloser(rt))
	return rt, restore, nil
}
