package home

import (
    "std"
    "testing"
)

func TestOwnership(t *testing.T) {
    owner := GetOwner()
    if owner.String() != "g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj" {
        t.Errorf("Expected owner to be g16vfw3r7zuz43fhky3xfsuc2hdv9tnhvlkyn0nj, got %s", owner.String())
    }

    std.TestSetOrigCaller(owner)
    if !o.CallerIsOwner() {
        t.Error("Owner address should return true for CallerIsOwner check")
    }

    nonOwner := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
    std.TestSetOrigCaller(nonOwner)
    if o.CallerIsOwner() {
        t.Error("Non-owner address should return false for CallerIsOwner check")
    }
}

func TestConfig(t *testing.T) {
    cfg := GetConfig()
    
    if cfg.title != "JJOptimist's Home Realm 🏠" {
        t.Errorf("Expected title to be 'JJOptimist's Home Realm 🏠', got %s", cfg.title)
    }
    if cfg.description != "Exploring Gno and building on-chain" {
        t.Errorf("Expected description to be 'Exploring Gno and building on-chain', got %s", cfg.description)
    }
    if cfg.github != "jjoptimist" {
        t.Errorf("Expected github to be 'jjoptimist', got %s", cfg.github)
    }
    if cfg.css != styles {
        t.Errorf("Expected css to match styles constant")
    }
}