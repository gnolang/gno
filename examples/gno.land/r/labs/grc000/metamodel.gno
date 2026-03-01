package grc000

import (
	"strings"

	mm "gno.land/p/labs/metamodel"
)

type Record struct {
	Id          string
	Name        string
	Description string
	Model       *mm.Model
}

const (
	Token = "$token"
)

var (
	basePath = "/r/labs/grc000"

	modelOpts = map[string]any{
		"balance":           100,
		"recipient_balance": 50,
		"transfer_amount":   10,
		"minter_initial":    1,
		"mint_amount":       5,
		"burn_amount":       2,
		"objects":           []string{Token},
	}

	registry = []*Record{
		{
			Id:          "wallet",
			Name:        "GRC000_Wallet",
			Description: "A simple wallet model that holds tokens.",
			Model:       GRC000_Wallet(modelOpts),
		},
		{
			Id:          "transfer",
			Name:        "GRC000_Transfer",
			Description: "A model for transferring tokens between wallets.",
			Model:       mm.New(GRC000_Wallet(modelOpts), GRC000_Transfer(modelOpts), map[string]any{"version": "v0"}),
		},
		{
			Id:          "mintburn",
			Name:        "GRC000_MintBurn",
			Description: "A model for minting into and burning from a wallet (with burned-accounting sink).",
			Model:       mm.New(GRC000_Wallet(modelOpts), GRC000_MintBurn(modelOpts), map[string]any{"version": "v0"}),
		},
		{
			Id:          "token",
			Name:        "GRC000_Token",
			Description: "A comprehensive model that composes wallet, transfer, and mint/burn.",
			Model:       GRC000_Token(modelOpts),
		},
	}

	registryMap = make(map[string]*Record)
)

// ------- utils -------

func getIntOpt(opts map[string]any, key string, def int) int {
	if v, ok := opts[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getStringSliceOpt(opts map[string]any, key string, def []string) []string {
	if v, ok := opts[key]; ok {
		if sl, ok := v.([]string); ok && len(sl) > 0 {
			return sl
		}
	}
	return def
}

func objectsFrom(opts map[string]any) []string {
	return getStringSliceOpt(opts, "objects", []string{Token})
}

// ------- models -------

func GRC000_Wallet(opts map[string]any) *mm.Model {
	initial := mm.T(getIntOpt(opts, "balance", 1))
	m := mm.New(
		objectsFrom(opts),
		map[string]mm.Place{
			"$wallet": {Initial: initial, Capacity: mm.T(0), X: 160, Y: 180},
		},
		map[string]any{"version": "v0"},
	)
	return m
}

func GRC000_Transfer(opts map[string]any) *mm.Model {
	recipientInitial := mm.T(getIntOpt(opts, "recipient_balance", 0))
	transferWeight := int64(getIntOpt(opts, "transfer_amount", 1))
	objectModel := mm.New(objectsFrom(opts))

	m := mm.New(
		objectModel.Objects,
		map[string]mm.Place{
			"$recipient": {Initial: recipientInitial, X: 360, Y: 180},
		},
		map[string]mm.Transition{
			"transfer": {X: 260, Y: 180},
		},
		[]mm.Arrow{
			objectModel.Arrow(mm.Arc{Source: "$wallet", Target: "transfer", Weight: transferWeight, Object: Token}),
			objectModel.Arrow(mm.Arc{Source: "transfer", Target: "$recipient", Weight: transferWeight, Object: Token}),
		},
		map[string]any{"version": "v0"},
	)
	return m
}

func GRC000_MintBurn(opts map[string]any) *mm.Model {
	minterInitial := mm.T(getIntOpt(opts, "minter_initial", 1))
	mintWeight := int64(getIntOpt(opts, "mint_amount", 1))
	burnWeight := int64(getIntOpt(opts, "burn_amount", 1))
	objectModel := mm.New(objectsFrom(opts))

	m := mm.New(
		objectModel.Objects,
		map[string]mm.Place{
			"$minter": {Initial: minterInitial, X: 40, Y: 80},
			"$burned": {Initial: mm.T(0), X: 280, Y: 280}, // accounting sink
		},
		map[string]mm.Transition{
			"mint": {X: 160, Y: 80},
			"burn": {X: 160, Y: 280},
		},
		[]mm.Arrow{
			// mint: $minter -> mint -> $wallet
			objectModel.Arrow(mm.Arc{Source: "$minter", Target: "mint", Weight: mintWeight, Object: Token}),
			objectModel.Arrow(mm.Arc{Source: "mint", Target: "$wallet", Weight: mintWeight, Object: Token}),

			// burn: $wallet -> burn -> $burned (accounting)
			objectModel.Arrow(mm.Arc{Source: "$wallet", Target: "burn", Weight: burnWeight, Object: Token}),
			objectModel.Arrow(mm.Arc{Source: "burn", Target: "$burned", Weight: burnWeight, Object: Token}),
		},
		map[string]any{"version": "v0"},
	)
	return m
}

func GRC000_Token(opts map[string]any) *mm.Model {
	return mm.New(
		objectsFrom(opts),
		GRC000_Wallet(opts),
		GRC000_Transfer(opts),
		GRC000_MintBurn(opts),
		map[string]any{"version": "v0"},
	)
}

// ------- rendering -------

func init() {
	for _, rec := range registry {
		registryMap[rec.Id] = rec
		rec.Model.Binding = func(_ string) string {
			var sb strings.Builder
			sb.WriteString("# " + rec.Name + "\n")
			sb.WriteString(rec.Description + "\n\n")
			sb.WriteString("## Model\n")
			sb.WriteString(rec.Model.Markdown() + "\n")
			return sb.String()
		}
	}
}

func indexPage(_ string) string {
	content := `# GRC000 Metamodel
![grc000_Token combines all models](` + GRC000_Token(modelOpts).ThumbnailDataUrl() + `)
This is a collection of GRC000 metamodels that demonstrate token-related functionalities such as wallets, transfers, and minting/burning tokens.

Each model is designed to illustrate specific aspects of token management and can be used as a foundation for building more complex token systems.

All of these models compose to form a comprehensive token model called **GRC000_Token** (pictured above).

### What is a Metamodel?

A _metamodel_ is a framework for modeling dynamic systems using places, transitions, and arrows, inspired by Petri nets.

Models are defined by a set of _$objects_, each with a **unique identity** (e.g. $eth, $btc, $coffee_beans)

Each _$object_ is defined by its behavior **in all contexts**.

See the [Metamodel Token Standard](/r/labs/grc000$source) for more details.

## Models
`
	for _, rec := range registryMap {
		content += "- [" + rec.Name + "](" + basePath + ":" + rec.Id + "): " + rec.Description + "\n"
	}
	content += `## How can this be used?
- Extend with staking, governance, multi-token systems.
- Use as a foundation for dApps that need token management.
- Simulate and analyze token flows off-chain for insight.
`
	return content
}

func Render(path string) string {
	if rec, ok := registryMap[path]; ok {
		if content, ok := rec.Model.Binding.(func(string) string); ok {
			return content(path)
		}
		return rec.Model.Markdown()
	}
	return indexPage(path)
}
