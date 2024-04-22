
package grc721

type IGRC721Metadata interface {	
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