package dynreplacer

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		pairs []Pair
	}{
		{
			name:  "empty constructor",
			pairs: []Pair{},
		},
		{
			name: "single pair",
			pairs: []Pair{
				{":name:", func() string { return "World" }},
			},
		},
		{
			name: "multiple pairs",
			pairs: []Pair{
				{":greeting:", func() string { return "Hello" }},
				{":name:", func() string { return "World" }},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.pairs...)
			uassert.True(t, r.callbacks != nil, "callbacks map should be initialized")
			uassert.True(t, r.cachedValues != nil, "cachedValues map should be initialized")

			// Verify all callbacks were registered
			for _, pair := range tt.pairs {
				_, exists := r.callbacks[pair.Placeholder]
				uassert.True(t, exists, "callback should be registered for "+pair.Placeholder)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		name     string
		layout   string
		setup    func(*Replacer)
		expected string
	}{
		{
			name:     "empty layout",
			layout:   "",
			setup:    func(r *Replacer) {},
			expected: "",
		},
		{
			name:   "single replacement",
			layout: "Hello :name:!",
			setup: func(r *Replacer) {
				r.RegisterCallback(":name:", func() string { return "World" })
			},
			expected: "Hello World!",
		},
		{
			name:   "multiple replacements",
			layout: ":greeting: :name:!",
			setup: func(r *Replacer) {
				r.RegisterCallback(":greeting:", func() string { return "Hello" })
				r.RegisterCallback(":name:", func() string { return "World" })
			},
			expected: "Hello World!",
		},
		{
			name:   "no recursive replacement",
			layout: ":outer:",
			setup: func(r *Replacer) {
				r.RegisterCallback(":outer:", func() string { return ":inner:" })
				r.RegisterCallback(":inner:", func() string { return "content" })
			},
			expected: ":inner:",
		},
		{
			name:   "unused callbacks",
			layout: "Hello :name:!",
			setup: func(r *Replacer) {
				r.RegisterCallback(":name:", func() string { return "World" })
				r.RegisterCallback(":unused:", func() string { return "Never Called" })
			},
			expected: "Hello World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			tt.setup(r)
			result := r.Replace(tt.layout)
			uassert.Equal(t, tt.expected, result)
		})
	}
}

func TestCaching(t *testing.T) {
	r := New()
	callCount := 0
	r.RegisterCallback(":expensive:", func() string {
		callCount++
		return "computed"
	})

	layout := "Value: :expensive:"

	// First call should compute
	result1 := r.Replace(layout)
	uassert.Equal(t, "Value: computed", result1)
	uassert.Equal(t, 1, callCount)

	// Second call should use cache
	result2 := r.Replace(layout)
	uassert.Equal(t, "Value: computed", result2)
	uassert.Equal(t, 1, callCount)

	// After clearing cache, should recompute
	r.ClearCache()
	result3 := r.Replace(layout)
	uassert.Equal(t, "Value: computed", result3)
	uassert.Equal(t, 2, callCount)
}

func TestComplexExample(t *testing.T) {
	layout := `
		# Welcome to gno.land

		## Blog
		:latest-blogposts:

		## Events
		:next-events:

		## Awesome Gno
		:awesome-gno:
	`

	r := New(
		Pair{":latest-blogposts:", func() string { return "Latest blog posts content here" }},
		Pair{":next-events:", func() string { return "Upcoming events listed here" }},
		Pair{":awesome-gno:", func() string { return ":latest-blogposts: (This should NOT be replaced again)" }},
	)

	result := r.Replace(layout)

	// Check that original placeholders are replaced
	uassert.True(t, !strings.Contains(result, ":latest-blogposts:\n"), "':latest-blogposts:' placeholder should be replaced")
	uassert.True(t, !strings.Contains(result, ":next-events:\n"), "':next-events:' placeholder should be replaced")
	uassert.True(t, !strings.Contains(result, ":awesome-gno:\n"), "':awesome-gno:' placeholder should be replaced")

	// Check that the replacement content is present
	uassert.True(t, strings.Contains(result, "Latest blog posts content here"), "Blog posts content should be present")
	uassert.True(t, strings.Contains(result, "Upcoming events listed here"), "Events content should be present")
	uassert.True(t, strings.Contains(result, ":latest-blogposts: (This should NOT be replaced again)"),
		"Nested placeholder should not be replaced")
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		layout   string
		setup    func(*Replacer)
		expected string
	}{
		{
			name:   "empty string placeholder",
			layout: "Hello :",
			setup: func(r *Replacer) {
				r.RegisterCallback("", func() string { return "World" })
			},
			expected: "WorldHWorldeWorldlWorldlWorldoWorld World:World",
		},
		{
			name:   "overlapping placeholders",
			layout: "Hello :name::greeting:",
			setup: func(r *Replacer) {
				r.RegisterCallback(":name:", func() string { return "World" })
				r.RegisterCallback(":greeting:", func() string { return "Hi" })
				r.RegisterCallback(":name::greeting:", func() string { return "Should not match" })
			},
			expected: "Hello WorldHi",
		},
		{
			name:   "replacement order",
			layout: ":a::b::c:",
			setup: func(r *Replacer) {
				r.RegisterCallback(":c:", func() string { return "3" })
				r.RegisterCallback(":b:", func() string { return "2" })
				r.RegisterCallback(":a:", func() string { return "1" })
			},
			expected: "123",
		},
		{
			name:   "special characters in placeholders",
			layout: "Hello :$name#123:!",
			setup: func(r *Replacer) {
				r.RegisterCallback(":$name#123:", func() string { return "World" })
			},
			expected: "Hello World!",
		},
		{
			name:   "multiple occurrences of same placeholder",
			layout: ":name: and :name: again",
			setup: func(r *Replacer) {
				callCount := 0
				r.RegisterCallback(":name:", func() string {
					callCount++
					return "World"
				})
			},
			expected: "World and World again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			tt.setup(r)
			result := r.Replace(tt.layout)
			uassert.Equal(t, tt.expected, result)
		})
	}
}
