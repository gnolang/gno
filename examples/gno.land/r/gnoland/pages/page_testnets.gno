package gnopages

func init() {
	path := "testnets"
	title := "Gno.land Testnets"
	// XXX: description := """
	body := `## Other testnets

- **[staging.gno.land](https://staging.gno.land) (wiped every commit to master)**
- _[test3.gno.land](https://test3.gno.land) (latest)_
- _[test2.gno.land](https://test2.gno.land) (archive)_
- _[test1.gno.land](https://test1.gno.land) (archive)_

## Local devnet

See CONTRIBUTING.md on GitHub.
`
	_ = b.NewPost("", path, title, body, nil)
}
