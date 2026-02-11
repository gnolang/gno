package lint

import "github.com/gnolang/gno/gnovm/pkg/gnolang"

type Engine struct {
	config   *Config
	registry *Registry
	reporter Reporter
}

// TODO: handle verbose mode for linting engine

func NewEngine(config *Config, registry *Registry, reporter Reporter) *Engine {
	return &Engine{
		config:   config,
		registry: registry,
		reporter: reporter,
	}
}

func (e *Engine) Run(fset *gnolang.FileSet, sources map[string]string) int {
	issueCount := 0

	rules := e.getEnabledRules()
	if len(rules) == 0 {
		return 0
	}

	for _, fn := range fset.Files {
		source := sources[fn.FileName]
		nolint := NewNolintParser(source)

		issueCount += e.runOnFile(fn, source, nolint, rules)
	}

	return issueCount
}

func (e *Engine) getEnabledRules() []Rule {
	all := e.registry.All()
	enabled := make([]Rule, 0, len(all))

	for _, rule := range all {
		if e.config.IsRuleEnabled(rule.Info().ID) {
			enabled = append(enabled, rule)
		}
	}

	return enabled
}

func (e *Engine) runOnFile(fn *gnolang.FileNode, source string, nolint *NolintParser, rules []Rule) int {
	issueCount := 0

	gnolang.Transcribe(fn, func(ns []gnolang.Node, ftype gnolang.TransField, index int, n gnolang.Node, stage gnolang.TransStage) (gnolang.Node, gnolang.TransCtrl) {
		if stage != gnolang.TRANS_ENTER {
			return n, gnolang.TRANS_CONTINUE
		}

		ctx := &RuleContext{
			File:    fn,
			Source:  source,
			Parents: ns,
		}

		for _, rule := range rules {
			issues := rule.Check(ctx, n)
			for _, issue := range issues {
				if nolint.IsSuppressed(issue.Line, issue.RuleID) {
					continue
				}

				issue.Severity = e.config.EffectiveSeverity(issue.Severity)

				e.reporter.Report(issue)
				issueCount++
			}
		}

		return n, gnolang.TRANS_CONTINUE
	})

	return issueCount
}

func (e *Engine) Flush() error {
	return e.reporter.Flush()
}

func (e *Engine) Summary() (info, warnings, errors int) {
	return e.reporter.Summary()
}
