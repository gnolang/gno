package docs

func Render(_ string) string {
	return `# Gno Examples Documentation

Welcome to the Gno examples documentation index.
Explore various examples to learn more about Gno functionality and usage.

## Examples

- [Hello World](/r/docs/hello) - A simple introductory example.
- [Adder](/r/docs/adder) - An interactive example to update a number with transactions.
- [Source](/r/docs/source) - View realm source code.
- [Buttons](/r/docs/buttons) - Add buttons to your realm's render.
- [AVL Pager](/r/docs/avl_pager) - Paginate through AVL tree items.  
- ...

## Other resources

- [Official documentation](https://github.com/gnolang/gno/tree/master/docs) <!-- should be /docs with gnoweb embedding the docs/ folder. -->
`
}
