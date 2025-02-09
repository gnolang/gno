// Package dynreplacer provides a simple template engine for handling dynamic
// content replacement. It is similar to strings.Replacer but with lazy
// execution of replacements, making it more optimization-friendly in several
// cases. While strings.Replacer requires all replacement values to be computed
// upfront, dynreplacer only executes the callback functions for placeholders
// that actually exist in the template, avoiding unnecessary computations. 
//
// The package ensures efficient, non-recursive replacement of placeholders in a
// single pass. This lazy evaluation approach is particularly beneficial when:
// - Some replacement values are expensive to compute
// - Not all placeholders are guaranteed to be present in the template
// - Templates are reused with different content
//
// Example usage:
//
//	r := New(":name:", func() string { return "World" })
//	result := r.Replace("Hello :name:!") // Returns "Hello World!"
//
// The replacer caches computed values, so subsequent calls with the same
// placeholder will reuse the cached value instead of executing the callback
// again: 
//
//	r := New(":expensive:", func() string { return "computed" })
//	r.Replace("Value1: :expensive:") // Computes the value
//	r.Replace("Value2: :expensive:") // Uses cached value
//	r.ClearCache()                   // Force re-computation on next use
package dynreplacer

import (
	"strings"
)

// Replacer manages dynamic placeholders, their associated functions, and cached
// values.
type Replacer struct {
	callbacks    map[string]func() string
	cachedValues map[string]string
}

// New creates a new Replacer instance with optional initial replacements.
// It accepts pairs of arguments where each pair consists of a placeholder
// string and its corresponding callback function.
// 
// Example:
//
//        New(
//            ":name:", func() string { return "World" },
//            ":greeting:", func() string { return "Hello" },
//        )
func New(pairs ...interface{}) *Replacer {
	if len(pairs)%2 != 0 {
		panic("dynreplacer.New: odd argument count")
	}

	r := &Replacer{
		callbacks:    make(map[string]func() string),
		cachedValues: make(map[string]string),
	}

	for i := 0; i < len(pairs); i += 2 {
		placeholder, ok := pairs[i].(string)
		if !ok {
			panic("dynreplacer.New: placeholder must be a string")
		}

		callback, ok := pairs[i+1].(func() string)
		if !ok {
			panic("dynreplacer.New: callback must be a func() string")
		}

		r.RegisterCallback(placeholder, callback)
	}

	return r
}

// RegisterCallback associates a placeholder with a function to generate its
// content.
func (r *Replacer) RegisterCallback(placeholder string, callback func() string) {
	r.callbacks[placeholder] = callback
}

// Replace processes the given layout, replacing placeholders with cached or
// newly computed values.
func (r *Replacer) Replace(layout string) string {
	replacements := []string{}

	// Check for placeholders and compute/retrieve values
	hasReplacements := false
	for placeholder, callback := range r.callbacks {
		if strings.Contains(layout, placeholder) {
			value, exists := r.cachedValues[placeholder]
			if !exists {
				value = callback()
				r.cachedValues[placeholder] = value
			}
			replacements = append(replacements, placeholder, value)
			hasReplacements = true
		}
	}

	// If no replacements were found, return the original layout
	if !hasReplacements {
		return layout
	}

	// Create a strings.Replacer with all computed replacements
	replacer := strings.NewReplacer(replacements...)
	return replacer.Replace(layout)
}

// ClearCache clears all cached values, forcing re-computation on next Replace.
func (r *Replacer) ClearCache() {
	r.cachedValues = make(map[string]string)
}
