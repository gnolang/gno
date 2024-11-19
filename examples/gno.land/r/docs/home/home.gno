package home

func Render(_ string) string {
	return `# Gno Examples Documentation

Welcome to the Gno examples documentation index.
Explore various examples to learn more about Gno.


## Examples

- [Hello World](/r/docs/hello_world) - A simple introductory example.</li>
<!-- Add more links to other examples here -->

## Other resources

- [Official documentation](https://github.com/gnolang/gno/tree/master/docs) <!-- should be /docs with gnoweb embedding the docs/ folder. -->
`
}
