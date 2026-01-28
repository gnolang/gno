package run

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gnolang/gno/contribs/tessera/pkg/cluster"
	"github.com/gnolang/gno/contribs/tessera/pkg/common"
	"github.com/gnolang/gno/contribs/tessera/pkg/recipe"
	"github.com/gnolang/gno/contribs/tessera/pkg/scenario"
	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/zap/zapcore"
)

const (
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

var (
	errInvalidOutputFormat = errors.New("invalid report output format")
	errInvalidTimeout      = errors.New("invalid timeout duration")
	errInvalidRecipesDir   = errors.New("invalid recipes directory path")
)

type runCfg struct {
	runInBand    bool
	outputPath   string
	outputFormat string
	runTimeout   time.Duration

	runAll     bool
	tags       commands.StringArr
	recipesDir string

	gnoRoot string // TODO revise
}

// NewRunCmd creates the tessera run subcommand
func NewRunCmd(io commands.IO) *commands.Command {
	cfg := &runCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] [<recipe-name>]",
			ShortHelp:  "runs the specific recipe suite",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execRun(ctx, args, cfg, io)
		},
	)
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.runInBand,
		"in-band",
		false,
		"run recipes sequentially (not in parallel)",
	)

	fs.StringVar(
		&c.outputFormat,
		"output",
		"",
		"the output report file for the tessera run. Leave empty for only CLI output",
	)

	fs.StringVar(
		&c.outputFormat,
		"output-format",
		"text",
		"the output report format. Possible [json, text]",
	)

	fs.DurationVar(
		&c.runTimeout,
		"timeout",
		time.Second*60,
		"the global run timeout (for all recipes)",
	)

	fs.BoolVar(
		&c.runAll,
		"all",
		false,
		"flag indicating if all recipes should be run in the directory",
	)

	fs.Var(
		&c.tags,
		"tag",
		"runs specific recipe paths",
	)

	fs.StringVar(
		&c.recipesDir,
		"recipes-dir",
		"",
		"the top-level path containing recipes",
	)

	fs.StringVar(
		&c.gnoRoot,
		"gno-root",
		"",
		"the root of the gno repository",
	)
}

func (c *runCfg) validateFlags(args []string) error {
	var (
		specificRecipe = len(args) > 0
		runAll         = c.runAll
		specificTag    = len(c.tags) > 0
	)

	// Make sure a specific recipe is not bundled with all runs
	if runAll && specificRecipe {
		return errors.New("invalid flag combination, cannot run all with a specified recipe")
	}

	// Make sure a specific recipe is not bundled with all runs
	if specificTag && specificRecipe {
		return errors.New("invalid flag combination, cannot run tags with a specified recipe")
	}

	if runAll && specificTag {
		return errors.New("invalid flag combination, cannot run all with a specified tag")
	}

	// Make sure the recipe is specified
	if !specificRecipe && !runAll && !specificTag {
		return errors.New("no run target specified") // TODO extract
	}

	// Make sure the recipes dir is set
	if c.recipesDir == "" {
		return errInvalidRecipesDir
	}

	// Make sure the timeout is valid
	if c.runTimeout.Seconds() <= 0 {
		return fmt.Errorf("%w: %.2fs", errInvalidTimeout, c.runTimeout.Seconds())
	}

	// Make sure the output format is valid
	if c.outputPath != "" &&
		(c.outputFormat != "text" && c.outputFormat != "json") {
		return fmt.Errorf("%w: %q", errInvalidOutputFormat, c.outputFormat)
	}

	// Make sure the gno root is set
	if c.gnoRoot == "" {
		return errors.New("gno root is not set")
	}

	return nil
}

func execRun(
	ctx context.Context,
	args []string,
	cfg *runCfg,
	io commands.IO,
) error {
	// Validate the flags
	if err := cfg.validateFlags(args); err != nil {
		return fmt.Errorf("unable to validate flags: %w", err)
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.runTimeout)
	defer cancel()

	// Collect recipe paths (single, all, or tagged)
	recipePaths, err := collectRecipePaths(cfg, args)
	if err != nil {
		return fmt.Errorf("unable to resolve recipe paths: %w", err)
	}

	if len(recipePaths) == 0 {
		return errors.New("no recipes found matching criteria") // TODO extract
	}

	// TODO set up a reporter (for output)

	// TODO cleanup
	gnoRoot, err := filepath.Abs(cfg.gnoRoot)
	if err != nil {
		return fmt.Errorf("unable to get abs path for gnoroot: %w", err)
	}

	cfg.gnoRoot = gnoRoot

	// Run the recipes
	io.Printf("Running %d recipes...\n\n", len(recipePaths))

	if cfg.runInBand {
		// Run sequentially
		for _, path := range recipePaths {
			if err := runSingleRecipe(timeoutCtx, path, io, cfg.gnoRoot); err != nil {
				io.ErrPrintfln("Error running recipe %s: %v\n", path, err)
			}
		}
	} else {
		// Run in parallel
		var (
			wg      sync.WaitGroup
			errorCh = make(chan error, len(recipePaths))
		)

		for _, path := range recipePaths {
			wg.Add(1)

			go func(recipePath string) {
				defer wg.Done()

				if e := runSingleRecipe(timeoutCtx, recipePath, io, cfg.gnoRoot); e != nil {
					errorCh <- fmt.Errorf("error running recipe %s: %w", recipePath, e)
				}
			}(path)
		}

		wg.Wait()
		close(errorCh)

		for e := range errorCh {
			io.ErrPrintln(e.Error())
		}
	}

	io.Println()
	io.Printfln("Recipe execution completed")

	return nil
}

// runSingleRecipe // TODO drop gnoroot
func runSingleRecipe(ctx context.Context, path string, io commands.IO, gnoRoot string) error {
	io.Printf("Running recipe: %s\n", path)

	// Set up logger
	// TODO move this out to the top-level ctx and use it
	zLogger := log.GetZapLoggerFn(log.ConsoleFormat)(io.Out(), zapcore.DebugLevel)
	logger := log.ZapLoggerToSlog(zLogger)

	// Load the recipe config
	r, err := common.LoadYAML[recipe.Config](path)
	if err != nil {
		return fmt.Errorf("failed to load recipe: %w", err)
	}

	// Create and set up cluster
	c, err := cluster.New(ctx, logger, r.Cluster, gnoRoot)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	defer c.Shutdown(ctx)

	// Run scenarios
	for i, scenarioCfg := range r.Scenarios {
		io.Printf("[%d/%d] Running scenario: %s\n", i+1, len(r.Scenarios), scenarioCfg.Name)

		start := time.Now()

		loadedScenario, err := scenario.Load(scenarioCfg.Name, scenarioCfg.Params)
		if err != nil {
			io.ErrPrintfln(
				"[R: %s] Scenario %s loading failed: %s [%dms]",
				r.Name, scenarioCfg.Name, err.Error(), time.Since(start).Milliseconds(),
			)

			return fmt.Errorf("failed to create test case: %w", err)
		}

		if err := loadedScenario.Execute(ctx, c); err != nil {
			io.ErrPrintfln(
				"[R: %s] Scenario %s execution failed: %s [%dms]",
				r.Name, scenarioCfg.Name, err.Error(), time.Since(start).Milliseconds(),
			)

			return fmt.Errorf("scenario execution failed: %w", err)
		}

		if err := loadedScenario.Verify(ctx, c); err != nil {
			io.ErrPrintfln(
				"[R: %s] Scenario %s verification failed: %s [%dms]",
				r.Name, scenarioCfg.Name, err.Error(), time.Since(start).Milliseconds(),
			)

			return fmt.Errorf("scenario verification failed: %w", err)
		}

		io.Printfln(
			"[R: %s] Scenario %s completed successfully [%dms]",
			r.Name, scenarioCfg.Name, time.Since(start).Milliseconds(),
		)
	}

	// Report success
	io.Printfln("Recipe %s completed successfully", r.Name)

	return nil
}

// collectRecipePaths collects the recipe paths to run, based
// on the run configuration
func collectRecipePaths(cfg *runCfg, args []string) ([]string, error) {
	switch {
	case len(args) > 0: // specific recipe
		return collectSingleRecipesPath(cfg.recipesDir, args)
	case len(cfg.tags) > 0: // specific recipe subset (tag)
		return collectTaggedRecipes(cfg.recipesDir, cfg.tags)
	default: // all recipes
		return collectAllRecipes(cfg.recipesDir)
	}
}

// collectSingleRecipesPath handles the single recipe paths
func collectSingleRecipesPath(recipesDir string, recipeArgs []string) ([]string, error) {
	recipePaths := make([]string, 0, len(recipeArgs))

	for _, recipeArg := range recipeArgs {
		// Ensure .yaml extension
		ext := filepath.Ext(recipeArg)
		if ext != yamlExt && ext != ymlExt {
			recipeArg += yamlExt
		}

		fullPath := filepath.Join(recipesDir, recipeArg)
		if _, err := os.Stat(fullPath); err != nil {
			return nil, fmt.Errorf("recipe not found: %s", fullPath)
		}

		recipePaths = append(recipePaths, fullPath)
	}

	return recipePaths, nil
}

// collectAllRecipes walks the recipes directory, returning all .yaml / .yml files (recipes)
func collectAllRecipes(recipesDir string) ([]string, error) {
	var recipePaths []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := filepath.Ext(path)

			if ext == yamlExt || ext == ymlExt {
				recipePaths = append(recipePaths, path)
			}
		}

		return nil
	}

	if err := filepath.Walk(recipesDir, walkFn); err != nil {
		return nil, fmt.Errorf("unable to walk recipes directory: %w", err)
	}

	return recipePaths, nil
}

// collectTaggedRecipes collects the recipe paths that match the specific directory tag
func collectTaggedRecipes(recipesDir string, tags []string) ([]string, error) {
	// Get the list of all recipes (paths), so they can be filtered out
	allRecipes, err := collectAllRecipes(recipesDir)
	if err != nil {
		return nil, err
	}

	var taggedRecipes []string

	for _, path := range allRecipes {
		pathTag := getPathTag(recipesDir, path)

		// Check for tag matches (relative path matches),
		// for example `-tag basic/subcat`, "basic/subcat" will
		// be the (directory) tag to match
		for _, tag := range tags {
			if tag != pathTag {
				continue
			}

			taggedRecipes = append(taggedRecipes, path)
		}
	}

	return taggedRecipes, nil
}

// getPathTag returns the directory portion of a recipe's path relative
// to the base recipes directory
func getPathTag(baseDir, fullPath string) string {
	relPath, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		// If something went wrong or baseDir == fullPath, fall back to the directory
		return filepath.Dir(fullPath)
	}

	return filepath.Dir(relPath)
}
