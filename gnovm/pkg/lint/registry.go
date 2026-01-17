package lint

import "fmt"

type Registry struct {
	rules map[string]Rule
}

var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[string]Rule),
	}
}

func (r *Registry) Register(rule Rule) error {
	id := rule.Info().ID
	if _, exists := r.rules[id]; exists {
		return fmt.Errorf("rule %q already registered", id)
	}
	r.rules[id] = rule
	return nil
}

func (r *Registry) MustRegister(rule Rule) {
	if err := r.Register(rule); err != nil {
		panic(err)
	}
}

func (r *Registry) Get(id string) (Rule, bool) {
	rule, ok := r.rules[id]
	return rule, ok
}

func (r *Registry) All() []Rule {
	rules := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

func Register(rule Rule) error {
	return DefaultRegistry.Register(rule)
}

func MustRegister(rule Rule) {
	DefaultRegistry.MustRegister(rule)
}
