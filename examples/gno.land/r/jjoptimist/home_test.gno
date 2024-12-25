package home

import (
    "std"
    "testing"
    "time"
)

func TestOwnership(t *testing.T) {
    
    owner := GetOwner()
    if owner.String() != "g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj" {
        t.Errorf("Expected owner to be g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj, got %s", owner.String())
    }

    
    if !IsOwner(owner) {
        t.Error("Owner address should return true for IsOwner check")
    }

    
    nonOwner := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
    if IsOwner(nonOwner) {
        t.Error("Non-owner address should return false for IsOwner check")
    }
}

func TestConfig(t *testing.T) {
    cfg := GetConfig()
    
    // Test initial config values
    if cfg.title != "JJOptimist's Home Realm 🏠" {
        t.Errorf("Expected title to be 'JJOptimist's Home Realm 🏠', got %s", cfg.title)
    }
    if cfg.description != "Exploring Gno and building on-chain" {
        t.Errorf("Expected description to be 'Exploring Gno and building on-chain', got %s", cfg.description)
    }
    if cfg.github != "jjoptimist" {
        t.Errorf("Expected github to be 'jjoptimist', got %s", cfg.github)
    }
}

func TestRender(t *testing.T) {
    output := Render("")
    
    // Random test for section
    expectedSections := []string{
        "# JJOptimist's Home Realm",
        "## About Me",
        "## Contact",
        "GitHub: [jjoptimist]",
        "_Owner: g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj_",
    }
    
    for _, section := range expectedSections {
        if !contains(output, section) {
            t.Errorf("Expected output to contain '%s'", section)
        }
    }
}


func contains(s, substr string) bool {
    for i := 0; i < len(s)-len(substr)+1; i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}