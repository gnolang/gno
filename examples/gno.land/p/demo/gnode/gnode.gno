package gnode

// XXX what about Gnodes signing on behalf of others?
// XXX like a multi-sig of Gnodes?

type Name string

type Gnode interface {
	//----------------------------------------
	// Basic properties
	GetName() Name

	//----------------------------------------
	// Affiliate Gnodes
	NumAffiliates() int
	GetAffiliates(Name) Affiliate
	AddAffiliate(Affiliate) error // must be affiliated
	RemAffiliate(Name) error      // must have become unaffiliated

	//----------------------------------------
	// Signing
	NumSignedDocuments() int
	GetSignedDocument(idx int) Document
	SignDocument(doc Document) (int, error) // index relative to signer

	//----------------------------------------
	// Rendering
	RenderLines() []string
}

type Affiliate struct {
	Type  string
	Gnode Gnode
	Tags  []string
}

type MyGnode struct {
	Name
	// Owners     // voting set, something that gives authority of action.
	// Treasury   //
	// Affiliates //
	// Board      // discussions
	// Data       // XXX ?
}

type Affiliates []*Affiliate

// Documents are equal if they compare equal.
// NOTE: requires all fields to be comparable.
type Document struct {
	Authors string
	// Timestamp
	// Body
	// Attachments
}

// ACTIONS

// * Lend tokens
// * Pay tokens
// * Administrate transferrable and non-transferrable tokens
// * Sum tokens
// * Passthrough dependencies
// * Code
// * ...
