package gnopages

func init() {
	path := "testnets"
	title := "Gno.land Testnet List"
	body := `
- [Portal Loop](https://docs.gno.land/concepts/portal-loop) - a rolling testnet 
- [staging.gno.land](https://staging.gno.land) - wiped every commit to monorepo master
- test4.gno.land (upcoming)
- _[test3.gno.land](https://test3.gno.land) (latest)_
- _[test2.gno.land](https://test2.gno.land) (archive)_
- _[test1.gno.land](https://test1.gno.land) (archive)_

For a list of RPC endpoints, see the [reference documentation](https://docs.gno.land/reference/rpc-endpoints).

## Local development

See the "Getting started" section in the [official documentation](https://docs.gno.land/getting-started/local-setup).
`
	_ = b.NewPost("", path, title, body, "2022-05-20T13:17:29Z", nil, nil)
}
