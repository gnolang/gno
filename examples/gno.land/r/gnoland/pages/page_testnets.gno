package pages

func init() {
	path := "testnets"
	title := "Gno.land Testnet List"
	body := `
- [Staging](https://docs.gno.land/resources/gnoland-networks/#staging-environments-portal-loops) - a rolling testnet

For a list of RPC endpoints, see the [reference documentation](https://docs.gno.land/reference/rpc-endpoints).

## Local development

See the "Build on Gno.land" section in the [Official Documentation](https://docs.gno.land/).
`
	if err := b.NewPost("", path, title, body, "2022-05-20T13:17:27Z", nil, nil); err != nil {
		panic(err)
	}
}
