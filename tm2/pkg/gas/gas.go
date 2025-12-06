// Package gas provides gas metering functionality for the SDK.
//
// This package centralizes:
//   - The list of all operations that cost gas
//   - The categorization of operations
//   - The default configuration for operation costs
//
// Performance considerations:
// Gas metering occurs during critical VM operations (CPU, memory, storage access, etc.),
// so the code is optimized for write performance. To achieve maximum speed, it uses
// the simplest possible types: a fixed-size array of structs containing two int64 values.
// Category sorting and aggregation only happens when explicitly requested via
// CategoryDetails(), which occurs at far less performance-critical moments (typically
// during result display or logging).
//
// This performance optimization requires manually keeping different values in sync
// across operation.go, particularly the operation list, category definitions, and
// their relationships.
package gas

import "fmt"

// Gas measured by the SDK.
type Gas = int64

// Detail tracks both the count of operations and the total gas consumed.
type Detail struct {
	OperationCount int64
	GasConsumed    Gas
}

// Add increments the operation count and adds the gas to the total.
func (d *Detail) Add(gas Gas) {
	d.OperationCount++
	d.GasConsumed += gas
}

// String returns a string representation of the Detail.
func (d Detail) String() string {
	return fmt.Sprintf("Operation count: %d, Gas consumed: %d", d.OperationCount, d.GasConsumed)
}

// GasDetail contains detailed gas consumption information.
type GasDetail struct {
	// Total gas consumed globally.
	Total Detail

	// Gas consumption detail per operation.
	Operations [OperationListMaxSize]Detail
}

// CategoryDetail contains gas consumption details for a specific category.
type CategoryDetail struct {
	// Total gas consumed in this category.
	Total Detail

	// Operation-wise gas consumption details within this category.
	Operations map[Operation]Detail
}

// CategoryDetails returns a map of CategoryDetail indexed by category name.
// NOTE: This mapping is constructed on access rather than during gas consumption
// to keep the gas metering code as fast as possible. Gas metering is called for
// every VM CPU, memory, and store operation and must be optimized for speed,
// while the access time for category details is not as critical.
func (gd GasDetail) CategoryDetails() map[string]CategoryDetail {
	categoryDetails := make(map[string]CategoryDetail, len(Categories()))

	// Iterate over all defined categories.
	for name, category := range Categories() {
		categoryDetail := CategoryDetail{Operations: make(map[Operation]Detail, category.Size())}

		// Iterate over all operations measured in the gas detail.
		for op, opDetail := range gd.Operations {
			operation := Operation(op)
			// If the operation belongs to the current category, accumulate its details.
			if category.Contains(operation) {
				categoryDetail.Total.OperationCount += opDetail.OperationCount
				categoryDetail.Total.GasConsumed += opDetail.GasConsumed
				categoryDetail.Operations[operation] = opDetail
			}
		}

		categoryDetails[name] = categoryDetail
	}

	return categoryDetails
}
