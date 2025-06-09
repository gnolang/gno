package gnopages

func init() {
	path := "testnets"
	title := "gno.land Testnet List"
	body := `
- [Staging](https://docs.gno.land/concepts/portal-loop) - a rolling testnet
- _[test5.gno.land](https://test5.gno.land) (latest)_

For a list of RPC endpoints, see the [reference documentation](https://docs.gno.land/reference/rpc-endpoints).

## Local development

See the "Getting started" section in the [official documentation](https://docs.gno.land/getting-started/local-setup).
`
	_ = b.NewPost("", path, title, body, "2022-05-20T13:17:29Z", nil, nil)
}
