package extendednft

import (
	"std"

	"gno.land/p/demo/grc/grc721remake"
)

// RoyaltyInfo represents royalty information for a token
type RoyaltyInfo struct {
	PaymentAddress std.Address // Address where royalty payment should be sent
	Percentage     uint64      // Royalty percentage (e.g., 10 means 10%)
}

// Trait represents a metadata attribute
type Trait struct {
	DisplayType string
	TraitType   string
	Value       string
}

// Metadata follows the OpenSea metadata standard
type Metadata struct {
	Image           string  // URL to the image
	ImageData       string  // Raw SVG image data
	ExternalURL     string  // URL that will appear below the asset
	Description     string  // Human-readable description
	Name            string  // Name of the item
	Attributes      []Trait // Attributes for the item
	BackgroundColor string  // Background color (hex without #)
	AnimationURL    string  // URL to multimedia attachment
	YoutubeURL      string  // URL to a YouTube video
}

type TokenID = grc721remake.TokenID
