package plugin

type Plugin interface {
	Type() string
}

type Name string
