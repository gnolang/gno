package home

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	cfg := GetConfig()

	if cfg.Title != "JJOptimist's Home Realm 🏠" {
		t.Errorf("Expected title to be 'JJOptimist's Home Realm 🏠', got %s", cfg.Title)
	}
	if cfg.Description != "Exploring Gno and building on-chain" {
		t.Errorf("Expected description to be 'Exploring Gno and building on-chain', got %s", cfg.Description)
	}
	if cfg.Github != "jjoptimist" {
		t.Errorf("Expected github to be 'jjoptimist', got %s", cfg.Github)
	}
}

func TestRender(t *testing.T) {
	output := Render("")

	// Test that required sections are present
	if !strings.Contains(output, "# "+config.Title) {
		t.Error("Rendered output missing title")
	}
	if !strings.Contains(output, "## About Me") {
		t.Error("Rendered output missing About Me section")
	}
	if !strings.Contains(output, "## Contact") {
		t.Error("Rendered output missing Contact section")
	}
	if !strings.Contains(output, config.Description) {
		t.Error("Rendered output missing description")
	}
	if !strings.Contains(output, config.Github) {
		t.Error("Rendered output missing github link")
	}
}

func TestGetGnomeArt(t *testing.T) {
	tests := []struct {
		height   int64
		expected string
	}{
		{7, gnomeArt4}, // height divisible by 7
		{5, gnomeArt3}, // height divisible by 5
		{3, gnomeArt2}, // height divisible by 3
		{2, gnomeArt1}, // default case
	}

	for _, tt := range tests {
		art := getGnomeArt(tt.height)
		if !strings.Contains(art, tt.expected) {
			t.Errorf("For height %d, expected art containing %s, got %s", tt.height, tt.expected, art)
		}
	}
}
