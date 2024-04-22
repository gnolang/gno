package grc721

// Base metadata interface
type IGRC721Metadata interface {
	Name() string
	Symbol() string
	TokenURI(tid TokenID) (string, error)
}

// see: https://docs.opensea.io/docs/metadata-standards
type IGRC721MetadataOnchain interface {
	IGRC721Metadata
	TokenMetadata(tid TokenID) (Metadata, error)
}

type Trait struct{
    DisplayType string
    TraitType string
    Value string
}

type Metadata struct {
	Image string
	ImageData string
	ExternalURL string
	Description string
	Name string
	Attributes []Trait
	BackgroundColor string
	AnimationURL string
	YoutubeURL string
}