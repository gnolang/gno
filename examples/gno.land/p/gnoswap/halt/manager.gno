package halt

import (
	"gno.land/p/demo/ufmt"
)

type Manager struct {
	currentLevel HaltLevel
	levels       map[LevelID]HaltLevel
	operations   map[OpType]Operation
}

// ManagerOption is a function type for configuring the `Manager`
type ManagerOption func(*Manager)

// WithOperations sets the initial Operations
func WithOperations(ops []Operation) ManagerOption {
	return func(m *Manager) {
		for _, op := range ops {
			m.RegisterOperation(op)
		}
	}
}

// WithLevels sets the initial Levels
func WithLevels(lvs []HaltLevel) ManagerOption {
	return func(m *Manager) {
		for _, lv := range lvs {
			m.RegisterHaltLevel(lv)
		}
	}
}

// WithInitialLevel sets the initial Level
func WithInitialLevel(lv LevelID) ManagerOption {
	return func(m *Manager) {
		if err := m.SetCurrentLevel(lv); err != nil {
			panic(err)
		}
	}
}

// NewManager creates a new `Manager` instance.
// especially, it should used to add custom halt-related operations and levels in a realm.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		levels:     make(map[LevelID]HaltLevel),
		operations: make(map[OpType]Operation),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// DefaultManager creates a new default halt manager.
// Call this function in a realm-level contract to create global state.
func DefaultManager() *Manager {
	m := &Manager{
		levels:     make(map[LevelID]HaltLevel),
		operations: make(map[OpType]Operation),
	}

	// register default operations
	for _, op := range defaultOperations {
		m.RegisterOperation(NewOperation(op.opType, op.name, op.desc))
	}

	// register default halt levels
	for _, level := range defaultLevels {
		m.RegisterHaltLevel(NewHaltLevel(level.id, level.name, level.desc, level.allowed))
	}

	if err := m.SetCurrentLevel(LvNoHalt); err != nil {
		panic(err)
	}

	return m
}

// SetCurrentLevel adds a new halt level to the manager.
func (m *Manager) SetCurrentLevel(lv LevelID) error {
	newLevel, ok := m.levels[lv]
	if !ok {
		return ufmt.Errorf("invalid halt level: %d", lv)
	}
	m.currentLevel = newLevel
	return nil
}

// Enable or disable a specific operation type at the current halt level
// Related issue: https://github.com/gnoswap-labs/gnoswap/issues/517
func (m *Manager) SetOperationStatus(opType OpType, allowed bool) error {
	// Get current level
	level, ok := m.levels[m.currentLevel.Level()]
	if !ok {
		return ufmt.Errorf("invalid current halt level")
	}

	// Update operation status
	if basicLevel, ok := level.(BasicHaltLevel); ok {
		basicLevel.allowedOps[opType] = allowed
		m.levels[m.currentLevel.Level()] = basicLevel
		return nil
	}

	return ufmt.Errorf("cannot modify composite halt levels directly")
}

// RegisterOperation adds a new operation to the manager.
func (m *Manager) RegisterOperation(op Operation) {
	m.operations[op.Type()] = op
}

// RegisterHaltLevel adds a new halt level to the manager.
func (m *Manager) RegisterHaltLevel(lv HaltLevel) {
	m.levels[lv.Level()] = lv
}

// Status returns the current halt level name.
func (m *Manager) Status(o OpType) string {
	return m.currentLevel.Name()
}

// Level returns the current halt level.
func (m *Manager) Level() HaltLevel {
	return m.currentLevel
}
