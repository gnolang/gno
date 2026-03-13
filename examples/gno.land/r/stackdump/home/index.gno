package home

import (
	"strings"

	mm "gno.land/p/labs/metamodel"
)

var (
	renderOpts = map[string]interface{}{
		"dev": map[string]interface{}{
			"Title":       "Stackdump Home",
			"Description": "Metamodel Registry",
			"Version":     "0.0.1",
			"BasePath":    "/r/stackdump/home",
		},
	}

	// consider any path that ends with /metamodel/
	authorizedPaths = []string{
		"gno.land/r/stackdump/home",
		"", // TODO: this opens perms for self, does it open for all EOAs? (i.e. non-realm accounts)
	}

	registry = mm.NewRegistry(authorizedPaths, renderOpts)
)

func register(key string, m *mm.Model, kw ...string) {
	registry.Register(key, m, kw...)
}

func renderIndex(path string) string {
	var sb strings.Builder
	sb.WriteString("# Metamodel000 Index\n\n")
	sb.WriteString("This is a public registry of metamodels.\n\n")
	sb.WriteString("### What is a Metamodel?\n\n")
	sb.WriteString("A _metamodel_ is a framework for modeling dynamic systems using places, transitions, and arrows, inspired by Petri nets.\n\n")
	sb.WriteString("Models are defined by a set of _$objects_, each with a **unique identity** (e.g. $eth, $btc, $coffee_beans )\n\n")
	sb.WriteString("Each _$object_ is defined by its behavior **in all contexts**.\n\n")
	sb.WriteString("Read the [Metamodel Codex](/p/labs/metamodel$source) for more background.\n\n")

	return sb.String()
}

func Render(path string) string {
	return registry.Render(path, renderIndex)
}
