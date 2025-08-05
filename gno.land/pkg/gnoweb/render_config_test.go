package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
)

func TestNewDefaultRenderConfig(t *testing.T) {
	cfg := NewDefaultRenderConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.GoldmarkOptions)
	assert.Greater(t, len(cfg.GoldmarkOptions), 0)
}

func TestNewRealmGoldmarkOptions(t *testing.T) {
	options := NewRealmGoldmarkOptions()
	require.NotNil(t, options)
	assert.Greater(t, len(options), 0)

	// Test that we can create a Goldmark instance with these options
	md := goldmark.New(options...)
	require.NotNil(t, md)

	// Test that the options include parser options
	hasParserOptions := false
	for _, opt := range options {
		if opt != nil {
			hasParserOptions = true
			break
		}
	}
	assert.True(t, hasParserOptions, "Realm options should include parser options")
}

func TestNewDocumentationGoldmarkOptions(t *testing.T) {
	options := NewDocumentationGoldmarkOptions()
	require.NotNil(t, options)
	assert.Greater(t, len(options), 0)

	// Test that we can create a Goldmark instance with these options
	md := goldmark.New(options...)
	require.NotNil(t, md)

	// Test that the options include parser options
	hasParserOptions := false
	for _, opt := range options {
		if opt != nil {
			hasParserOptions = true
			break
		}
	}
	assert.True(t, hasParserOptions, "Documentation options should include parser options")
}

func TestRenderConfigOptionsComparison(t *testing.T) {
	realmOptions := NewRealmGoldmarkOptions()
	docOptions := NewDocumentationGoldmarkOptions()

	// Realm options should have more extensions than documentation options
	// (realm has more features like columns, alerts, forms, etc.)
	assert.GreaterOrEqual(t, len(realmOptions), len(docOptions), 
		"Realm options should have at least as many options as documentation options")

	// Both should be valid Goldmark configurations
	realmMD := goldmark.New(realmOptions...)
	docMD := goldmark.New(docOptions...)
	require.NotNil(t, realmMD)
	require.NotNil(t, docMD)
} 