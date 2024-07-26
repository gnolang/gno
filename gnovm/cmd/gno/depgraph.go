package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type depGraphCfg struct {
	verbose        bool
	rootDir        string
	output         string
	multipleGraphs bool
}

func newDepGraphCmd(io commands.IO) *commands.Command {
	cfg := &depGraphCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "depgraph",
			ShortUsage: "depgraph [flags] <package> [<package>...]",
			ShortHelp:  "generates dependency graphs for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execDepGraph(cfg, args, io)
		},
	)
}

func (c *depGraphCfg) RegisterFlags(fs *flag.FlagSet) {
	rootDir := gnoenv.RootDir()
	fs.BoolVar(&c.verbose, "v", false, "verbose output when lintning")
	fs.StringVar(&c.rootDir, "root-dir", rootDir, "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.StringVar(&c.output, "o", "depgraph", "output (file if single graph, dir if multiple graphs)")
	fs.BoolVar(&c.multipleGraphs, "m", false, "make a separate graph for each package")
}

func execDepGraph(cfg *depGraphCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	var (
		verbose        = cfg.verbose
		rootDir        = cfg.rootDir
		output         = cfg.output
		multipleGraphs = cfg.multipleGraphs
		allPkgs        gnomod.PkgList
	)

	for _, arg := range args {
		pkgs, err := gnomod.ListPkgs(arg)
		if err != nil {
			return fmt.Errorf("error in parsing gno.mod: %w", err)
		}

		allPkgs = append(allPkgs, pkgs...)
	}

	//make one big graph (eg. for the entire examples/ dir)
	if !multipleGraphs {
		nodeData := ""  //nodes
		graphData := "" //edges

		//subgraph for .../p/...
		nodeData += "subgraph {\nrank=same\n"
		for _, pkg := range allPkgs {
			if strings.Contains(pkg.Name, "gno.land/p") {
				nodeData += "\"" + pkg.Name + "\" [color=\"blue\"]\n"
			}
		}

		//subgraph for .../r/...
		nodeData += "}\nsubgraph {\nrank=same\n"
		for _, pkg := range allPkgs {
			if strings.Contains(pkg.Name, "gno.land/r") {
				nodeData += "\"" + pkg.Name + "\" [color=\"red\"]\n"
			}
		}
		nodeData += "}"

		for _, pkg := range allPkgs {

			err := buildGraphData(pkg, allPkgs, make(map[string]bool), make(map[string]bool), &graphData)

			if err != nil {
				return fmt.Errorf("error in building graph: %w", err)
			}
		}

		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("couldn't open output file: %w", err)
		}
		graphFileData := fmt.Sprintf("Digraph G {\nrankdir=\"LR\"\nranksep=20\n%s\n%s\n}", nodeData, graphData)
		file.Write([]byte(graphFileData))
		file.Close()
	} else { //useful for testing - makes a separate graph for each found package
		if !osm.DirExists(output) {
			err := os.MkdirAll(output, os.ModePerm)
			if err != nil {
				return fmt.Errorf("couldn't make output dir: %w", err)
			}
		}

		for _, pkg := range allPkgs {
			pkgPath, err := filepath.Abs(pkg.Dir)
			if err != nil {
				return fmt.Errorf("error in getting path of pkg: %w", err)
			}

			pkgPath = strings.TrimPrefix(pkgPath, rootDir)
			pkgPath = strings.TrimSuffix(pkgPath, string([]rune{os.PathSeparator}))

			if verbose {
				fmt.Fprintf(io.Err(), "Generating graph for %q...\n", pkgPath)
			}

			graphData := ""
			graphPath := filepath.Join(output, pkgPath)
			graphPath = graphPath + ".dot"
			basePath := path.Dir(graphPath)
			err = os.MkdirAll(basePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error in making dir for graph: %w", err)
			}

			file, err := os.Create(graphPath)
			if err != nil {
				return fmt.Errorf("couldn't create output file: %w", err)
			}

			err = buildGraphData(pkg, allPkgs, make(map[string]bool), make(map[string]bool), &graphData)

			if err != nil {
				return fmt.Errorf("error in building graph: %w", err)
			}

			graphFileData := fmt.Sprintf("Digraph G {%s}\n", graphData)
			file.Write([]byte(graphFileData))
			file.Close()
		}
	}

	return nil
}

func buildGraphData(pkg gnomod.Pkg, allPkgs []gnomod.Pkg, visited map[string]bool, onStack map[string]bool, graphData *string) error {
	if onStack[pkg.Name] {
		return fmt.Errorf("cycle detected: %s", pkg.Name)
	}
	if visited[pkg.Name] {
		return nil
	}

	visited[pkg.Name] = true
	onStack[pkg.Name] = true

	for _, req := range pkg.Requires {
		found := false

		for _, candidate := range allPkgs {
			if candidate.Name != req {
				continue
			}
			if err := buildGraphData(candidate, allPkgs, visited, onStack, graphData); err != nil {
				return err
			}
			found = true
			//this check is wildly inefficient. should change graph data to map, then convert to string at end
			if !strings.Contains(*graphData, "\""+pkg.Name+"\" -> \""+req+"\"\n") {
				*graphData += "\"" + pkg.Name + "\" -> \"" + req + "\"\n"
			}
		}
		if !found {
			return fmt.Errorf("couldn't find dependency %q for package %q", req, pkg.Name)
		}
	}

	onStack[pkg.Name] = false

	return nil
}
