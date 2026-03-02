package scenario

// TODO revise this sketchy moment
var globalRegistry = NewRegistry()

func RegisterScenario(name string, scenario Scenario) {
	globalRegistry.Register(name, scenario)
}

// Registry is a lookup table for registered scenarios
type Registry struct {
	scenarios map[string]Scenario
}

// NewRegistry creates a fresh scenario registry
func NewRegistry() *Registry {
	return &Registry{
		scenarios: make(map[string]Scenario),
	}
}

// Register registers a new scenario under a typed name (ex. test_name)
func (r *Registry) Register(name string, scenario Scenario) {
	r.scenarios[name] = scenario
}

// Get fetches the scenario using the typed name, if any
func (r *Registry) Get(name string) (Scenario, bool) {
	scenario, found := r.scenarios[name]

	return scenario, found
}
