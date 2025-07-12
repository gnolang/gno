package coverage

import (
	"go/ast"
)

// CrossIdentifierDetector implements detection for 'cross' identifier
type CrossIdentifierDetector struct {
	markers []string
	cache   map[ast.Node]bool
}

func NewCrossIdentifierDetector() *CrossIdentifierDetector {
	return &CrossIdentifierDetector{
		markers: []string{"cross"},
		cache:   make(map[ast.Node]bool),
	}
}

func (d *CrossIdentifierDetector) IsExternallyInstrumented(node ast.Node) bool {
	if cached, ok := d.cache[node]; ok {
		return cached
	}

	var containsExternal bool
	ast.Inspect(node, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			for _, marker := range d.markers {
				if ident.Name == marker {
					containsExternal = true
					return false
				}
			}
		}
		return true
	})

	d.cache[node] = containsExternal
	return containsExternal
}

func (d *CrossIdentifierDetector) GetExternalMarkers() []string {
	return d.markers
}
