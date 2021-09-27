package client

type BaseOptions struct {
	Home   string `flag:"home" help:"home directory"`
	Remote string `flag:"remote" help:"remote node URL"`
}
